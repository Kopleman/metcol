current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

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

.PHONY: prepare-lint
prepare-lint:
	rm -rf golangci-lint/*

.PHONY: raw-lint
raw-lint:
	docker run --rm \
        -v $(PWD):/app \
        -v $(PWD)/golangci-lint/.cache/golangci-lint/v1.57.2:/root/.cache \
        -w /app \
        golangci/golangci-lint:v1.57.2 \
            golangci-lint run \
                -c .golangci.yml \
            > ./golangci-lint/report-unformatted.json

.PHONY: format-lint-report
format-lint-report:
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json
	rm ./golangci-lint/report-unformatted.json

.PHONY: lint
lint: prepare-lint raw-lint format-lint-report
