package workflows

import (
	"context"
	"fmt"
	"time"

	"github.com/inngest/inngestgo"
	"github.com/inngest/inngestgo/step"
)

const TaskCreatedEvent = "app/task.created"

type TaskData struct {
	ID string `json:"id"`
}

func Register(client inngestgo.Client) error {
	_, err := inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{ID: "process-task"},
		inngestgo.EventTrigger(TaskCreatedEvent, nil),
		ProcessTask,
	)
	return err
}

func ProcessTask(ctx context.Context, input inngestgo.Input[TaskData]) (any, error) {
	taskID := input.Event.Data.ID

	result, err := step.Run(ctx, "handle-task", func(ctx context.Context) (map[string]any, error) {
		return map[string]any{"processed": true, "id": taskID}, nil
	})
	if err != nil {
		return nil, err
	}

	step.Sleep(ctx, "pause", time.Second)

	return map[string]any{
		"message": fmt.Sprintf("Task %s complete", taskID),
		"result":  result,
	}, nil
}
