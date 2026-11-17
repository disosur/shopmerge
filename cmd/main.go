package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/disosur/shopmerge/internal/workflows"
	"github.com/inngest/inngestgo"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	client, err := inngestgo.NewClient(inngestgo.ClientOpts{AppID: envOr("INNGEST_APP_ID", "shopmerge")})
	if err != nil {
		log.Fatal(err)
	}

	if err := workflows.Register(client); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/inngest", client.Serve())
	mux.HandleFunc("POST /api/create-task", createTask(client))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	addr := ":" + envOr("PORT", "8080")
	log.Printf("listening on %s (dev server: %v)", addr, os.Getenv("INNGEST_DEV") != "")
	log.Fatal(http.ListenAndServe(addr, mux))
}

type createTaskRequest struct {
	ID string `json:"id"`
}

func createTask(client inngestgo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTaskRequest
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if req.ID == "" {
			req.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
		}

		eventID, err := client.Send(r.Context(), inngestgo.Event{
			Name: workflows.TaskCreatedEvent,
			Data: map[string]any{"id": req.ID},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]string{
			"eventId": eventID,
			"taskId":  req.ID,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
