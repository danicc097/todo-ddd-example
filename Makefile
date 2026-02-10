ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

.SILENT:

KNOWN_TARGETS := test lint clean deps dev gen gen-sqlc gen-schema db-init migrate-up deploy psql logs debug-swarm req-create req-list req-complete ws-listen rabbitmq-messages rabbitmq-queues

ifeq ($(findstring p,$(MAKEFLAGS)),)
  ifneq ($(filter-out $(KNOWN_TARGETS),$(MAKECMDGOALS)),)
    $(error Unknown target(s): $(filter-out $(KNOWN_TARGETS),$(MAKECMDGOALS)). Valid targets: $(KNOWN_TARGETS))
  endif
endif

SQLC   := go tool sqlc
PGROLL := go tool pgroll
AIR    := go tool air
GOLINT    := go tool golangci-lint

DB_USER ?= postgres
DB_PASS ?= postgres
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5732
DB_NAME ?= postgres
SERVICE ?= myapp

# URLs
PG_URL     := postgresql://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
GEN_DB     := $(SERVICE)_gen
GEN_PG_URL := postgresql://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(GEN_DB)?sslmode=disable

MIGRATIONS_DIR := ./migrations
SCHEMA_OUT     := ./sql/schema.sql

DB_CONTAINER_NAME = myapp-db
DOCKER_PSQL = docker exec -i $(DB_CONTAINER_NAME) psql -U $(DB_USER)

.PHONY: $(KNOWN_TARGETS)

deps:
	go mod download
	$(SQLC) version
	$(PGROLL) --version

dev:
	$(AIR) -c .air.toml

test:
	make gen-schema
	go test ./...

clean:
	rm -f $(SERVICE)

lint:
	$(GOLINT) run --allow-parallel-runners --fix

gen:
	make gen-sqlc
	go generate ./...

gen-sqlc:
	make gen-schema

	$(SQLC) generate -f internal/sqlc.yaml

gen-schema:
	if ! docker ps --format '{{.Names}}' | grep -q "^$(DB_CONTAINER_NAME)$$"; then \
		echo "Error: Container $(DB_CONTAINER_NAME) not found. Run 'make deploy' first."; exit 1; \
	fi

	$(DOCKER_PSQL) -d postgres -c "DROP DATABASE IF EXISTS $(GEN_DB);" >/dev/null
	$(DOCKER_PSQL) -d postgres -c "CREATE DATABASE $(GEN_DB);" >/dev/null

	$(PGROLL) --postgres-url "$(GEN_PG_URL)" init

	find $(MIGRATIONS_DIR) -name "*.json" | sort | xargs -I % $(PGROLL) --postgres-url "$(GEN_PG_URL)" start --complete %

	docker exec -i $(DB_CONTAINER_NAME) pg_dump -s -x -n public -U $(DB_USER) -d $(GEN_DB) \
		| grep -v '^\\' \
		| grep -v '^--' \
		| sed '/^$$/d' \
		> $(SCHEMA_OUT)

	$(DOCKER_PSQL) -d postgres -c "DROP DATABASE IF EXISTS $(GEN_DB);" >/dev/null

db-init:
	$(PGROLL) --postgres-url "$(PG_URL)" init

# Idempotent migration target
migrate-up:
	echo "Running migrations against $(PG_URL)..."
	$(PGROLL) --postgres-url "$(PG_URL)" init 2>/dev/null || true

	for file in $$(find $(MIGRATIONS_DIR) -name "*.json" | sort); do \
		NAME=$$(basename $$file .json); \
		EXISTS=$$(docker exec $(DB_CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME) -tAc "SELECT EXISTS(SELECT 1 FROM pgroll.migrations WHERE name = '$$NAME')"); \
		if [ "$$EXISTS" = "t" ]; then \
			echo "Skipping $$NAME (already applied)"; \
		else \
			echo "Applying $$file..."; \
			$(PGROLL) --postgres-url "$(PG_URL)" start --complete $$file || exit 1; \
		fi \
	done

deploy:
	./deploy.sh

psql:
	docker exec -it $(DB_CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME)

logs:
	docker service logs -f $(SERVICE)_go-app

debug-swarm:
	docker service ps --no-trunc $(SERVICE)_go-app

API_URL ?= http://127.0.0.1:8090

req-create:
	curl -sSf -X POST $(API_URL)/api/v1/todos -d '{"title": "New todo $(shell date +%s)"}' | jq -e .

req-list:
	curl -sSf -X GET $(API_URL)/api/v1/todos | jq -e .

req-complete:
ifndef ID
	$(error ID is undefined. Usage: make req-complete ID=...)
endif
	curl -sSf -X PATCH $(API_URL)/api/v1/todos/$(ID)/complete

ws-listen:
	WS_URL=$$(echo "$$API_URL" | sed 's/^http:/ws:/' | sed 's/^https:/wss:/'); \
	wscat -c "$${WS_URL}/ws"

N ?= 5
QUEUE ?= todo_events

rabbitmq-messages:
	docker exec myapp-rabbitmq rabbitmqadmin -V / get queue="$(QUEUE)" count="$(N)" -f pretty_json

rabbitmq-queues:
	docker exec myapp-rabbitmq rabbitmqadmin -V / list queues name messages messages_ready consumers -f table
