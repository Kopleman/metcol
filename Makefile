include scripts/*.mk
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