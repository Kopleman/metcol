-include .env

MIGRATIONS_DIR   = ./sql/migrations/
POSTGRES_DSN	 = $(DATABASE_DSN)


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

.PHONY: create-report-dir
create-report-dir:
	mkdir -p ./golangci-lint

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
lint: create-report-dir prepare-lint raw-lint format-lint-report

.PHONY: field-alignment
field-alignment:
	fieldalignment -fix ./...

.PHONY: install-dev-tools
install-dev-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/oligot/go-mod-upgrade@latest
	go install github.com/tkcrm/pgxgen/cmd/pgxgen@latest

.PHONY: migrate
migrate:
	migrate -path "$(MIGRATIONS_DIR)" -database "$(POSTGRES_DSN)" $(filter-out $@,$(MAKECMDGOALS))

.PHONY: db-create-migration
db-create-migration:
	migrate create -ext sql -dir "$(MIGRATIONS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

.PHONY: gensql
gensql:
	sqlc generate