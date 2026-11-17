SHELL := /usr/bin/env bash
.DEFAULT_GOAL := dev

-include .env
export

APP_PORT     ?= 8080
INNGEST_PORT ?= 8288
APP_URL      ?= http://localhost:$(APP_PORT)/api/inngest
BIN_DIR      := .bin
INNGEST      := $(BIN_DIR)/inngest
APP          := $(BIN_DIR)/shopmerge
GO_SOURCES   := $(shell find cmd internal -name '*.go' 2>/dev/null)

.PHONY: dev app inngest build run fmt vet test tidy clean

dev: $(APP) $(INNGEST)
	@echo "  app      http://localhost:$(APP_PORT)"
	@echo "  inngest  http://localhost:$(INNGEST_PORT)"
	@trap 'kill 0' EXIT INT TERM; \
		$(APP) & \
		$(INNGEST) dev --port $(INNGEST_PORT) -u $(APP_URL) & \
		wait

app: $(APP)
	@exec $(APP)

inngest: $(INNGEST)
	@exec $(INNGEST) dev --port $(INNGEST_PORT) -u $(APP_URL)

build: $(APP)

run: $(APP)
	@exec $(APP)

fmt:
	gofmt -l -w ./cmd ./internal

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

$(APP): $(GO_SOURCES) go.mod go.sum
	@mkdir -p $(BIN_DIR)
	go build -o $@ ./cmd

$(INNGEST):
	@set -euo pipefail; \
	mkdir -p $(BIN_DIR); \
	os=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	arch=$$(uname -m); \
	case $$arch in \
		x86_64|amd64) arch=amd64 ;; \
		aarch64|arm64) arch=arm64 ;; \
		*) echo "unsupported architecture: $$arch" >&2; exit 1 ;; \
	esac; \
	tag=$$(curl -sSfL https://api.github.com/repos/inngest/inngest/releases/latest | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1); \
	[ -n "$$tag" ] || { echo "could not resolve the latest inngest release" >&2; exit 1; }; \
	version=$${tag#v}; \
	asset=inngest_$${version}_$${os}_$${arch}.tar.gz; \
	url=https://github.com/inngest/inngest/releases/download/$$tag; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	curl -sSfL "$$url/$$asset" -o "$$tmp/$$asset"; \
	curl -sSfL "$$url/checksums.txt" -o "$$tmp/checksums.txt"; \
	(cd "$$tmp" && awk -v a="$$asset" '$$2 == a' checksums.txt | sha256sum -c -); \
	tar -xzf "$$tmp/$$asset" -C "$$tmp" inngest; \
	install "$$tmp/inngest" $(INNGEST); \
	echo "installed $(INNGEST) (inngest $$version)"
