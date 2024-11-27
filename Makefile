.PHONY: build-server
build-server:
	go build -o ./cmd/server/server ./cmd/server

.PHONY: build-agent
build-agent:
	go build -o ./cmd/agent/agent ./cmd/agent

.PHONY: build
build: build-server build-agent

.PHONY: test
test:
	go test ./...