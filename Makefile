ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

.SILENT:

KNOWN_TARGETS := test test-e2e lint clean deps dev gen gen-sqlc gen-schema db-init migrate-up gen-client deploy psql logs debug-swarm req-create req-list req-complete ws-listen rabbitmq-messages rabbitmq-queues rabbitmq-exchanges rabbitmq-bindings rabbitmq-watch



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

test-e2e:
	go test -tags e2e -v ./tests/e2e/...

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

gen-client:
	go tool oapi-codegen --config=internal/oapi-codegen-client.yaml openapi.yaml

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
	docker service logs myapp_go-app --no-trunc --raw -f

API_URL ?= http://127.0.0.1:8090
FAIL_FAST ?= 0

ifeq ($(FAIL_FAST),1)
    CURL_FLAGS := -f
else
    CURL_FLAGS :=
endif

req-create:
	curl $(CURL_FLAGS) -sS -X POST $(API_URL)/api/v1/todos \
		-H "Content-Type: application/json" \
		-d '{"title": "New todo $(shell date +%s)"}' | \
	jq -e .id

req-list:
	curl $(CURL_FLAGS) -sS -X GET $(API_URL)/api/v1/todos \
		-H "Content-Type: application/json" | \
		jq -e .

req-complete:
ifndef ID
	$(error ID is undefined. Usage: make req-complete ID=...)
endif
	curl $(CURL_FLAGS) -sS -X PATCH $(API_URL)/api/v1/todos/$(ID)/complete \
		-H "Content-Type: application/json"


ws-listen:
	WS_URL=$$(echo "$$API_URL" | sed 's/^http:/ws:/' | sed 's/^https:/wss:/'); \
	wscat -c "$${WS_URL}/ws"

N ?= 5
QUEUE ?= todo_events

rabbitmq-messages:
	docker exec myapp-rabbitmq rabbitmqadmin -V / get queue="$(QUEUE)" count="$(N)" -f pretty_json | jq 'reverse'

rabbitmq-watch:
	docker exec myapp-rabbitmq rabbitmqadmin delete queue name=debug_tap > /dev/null 2>&1
	docker exec myapp-rabbitmq rabbitmqadmin declare queue name=debug_tap auto_delete=true > /dev/null
	docker exec myapp-rabbitmq rabbitmqadmin declare binding source=$(QUEUE) destination=debug_tap routing_key="#" > /dev/null

	@echo ">>> Tailing live events on '$(QUEUE)'..."
	@# ack_requeue_false: remove message after read.
	while true; do \
		docker exec myapp-rabbitmq rabbitmqadmin get queue=debug_tap count=10 ackmode=ack_requeue_false -f pretty_json \
		| jq -r '.[]?'; \
		sleep 0.5; \
	done

rabbitmq-queues:
	docker exec myapp-rabbitmq rabbitmqadmin -V / list queues name messages messages_ready consumers -f table

rabbitmq-exchanges:
	docker exec myapp-rabbitmq rabbitmqadmin -V / list exchanges name type -f table

rabbitmq-bindings:
	docker exec myapp-rabbitmq rabbitmqadmin -V / list bindings source destination routing_key -f table
