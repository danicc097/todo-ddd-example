ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

.SILENT:

SQLC   := go tool sqlc
PGROLL := go tool pgroll

DB_USER ?= postgres
DB_PASS ?= postgres
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5732
DB_NAME ?= postgres
SERVICE ?= myapp

PG_URL     := postgresql://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
GEN_DB     := $(SERVICE)_gen
GEN_PG_URL := postgresql://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(GEN_DB)?sslmode=disable

MIGRATIONS_DIR := ./migrations
SCHEMA_OUT     := ./sql/schema.sql

DB_CONTAINER_ID = $(shell docker ps -q -f name=$(SERVICE)_db)
DOCKER_PSQL = docker exec -i $(DB_CONTAINER_ID) psql -U $(DB_USER)

.PHONY: all test clean deps
all: test build

deps:
	go mod download
	$(SQLC) version
	$(PGROLL) --version

test:
	go test -v ./...

clean:
	rm -f $(SERVICE)

.PHONY: gen-sqlc
gen-sqlc:
	@if [ -z "$(DB_CONTAINER_ID)" ]; then echo "DB container not found. Is the stack running?"; exit 1; fi

	$(DOCKER_PSQL) -d postgres -c "DROP DATABASE IF EXISTS $(GEN_DB);" &>/dev/null
	$(DOCKER_PSQL) -d postgres -c "CREATE DATABASE $(GEN_DB);" &>/dev/null

	$(PGROLL) --postgres-url "$(GEN_PG_URL)" init

	find $(MIGRATIONS_DIR) -name "*.json" | sort | xargs -I % $(PGROLL) --postgres-url "$(GEN_PG_URL)" start --complete %

	docker exec -i $(DB_CONTAINER_ID) pg_dump -s -x -n public -U $(DB_USER) -d $(GEN_DB) \
		| grep -v '^\\' \
		| grep -v '^--' \
		| sed '/^$$/d' \
		> $(SCHEMA_OUT)

	$(DOCKER_PSQL) -d postgres -c "DROP DATABASE IF EXISTS $(GEN_DB);" &>/dev/null

	$(SQLC) generate -f internal/sqlc.yaml

.PHONY: db-init migrate-up
db-init:
	$(PGROLL) --postgres-url "$(PG_URL)" init

migrate-up:
	echo "Running migrations against $(PG_URL)..."
	$(PGROLL) --postgres-url "$(PG_URL)" init || true
	for file in $$(find $(MIGRATIONS_DIR) -name "*.json" | sort); do \
		echo "Applying $$file..."; \
		$(PGROLL) --postgres-url "$(PG_URL)" start --complete $$file; \
	done

.PHONY: deploy psql logs debug-swarm
deploy:
	./deploy.sh

psql:
	docker exec -it $(DB_CONTAINER_ID) psql -U $(DB_USER) -d $(DB_NAME)

logs:
	docker service logs -f $(SERVICE)_go-app

debug-swarm:
	docker service ps --no-trunc $(SERVICE)_go-app

API_URL := http://127.0.0.1:8090

.PHONY: req-create req-list req-complete ws-listen

req-create:
	curl -s -X POST $(API_URL)/api/v1/todos -d '{"title": "New todo $(shell date +%s)"}' | jq .

req-list:
	curl -s -X GET $(API_URL)/api/v1/todos | jq .

req-complete:
ifndef ID
	$(error ID is undefined. Usage: make req-complete ID=...)
endif
	curl -s -X PATCH $(API_URL)/api/v1/todos/$(ID)/complete

ws-listen:
	wscat -c ws://127.0.0.1:8090/ws
