ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

.SILENT:

KNOWN_TARGETS := test test-race test-e2e lint clean deps lint dev gen gen-sqlc gen-cli gen-schema db-init migrate-up gen-oapi deploy psql logs run-gen-schema debug-swarm ws-listen rabbitmq-messages rabbitmq-queues rabbitmq-exchanges rabbitmq-bindings rabbitmq-watch



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

lint:
	go build ./... >/dev/null
	go test -c ./tests/e2e/... -tags e2e -o /dev/null
	$(GOLINT) run ./... --allow-parallel-runners --fix --config=.golangci.yml --issues-exit-code=0 >/dev/null

dev:
	$(AIR) -c .air.toml

test:
	$(MAKE) gen-schema
	go test ./... -count=1

test-race:
	$(MAKE) gen-schema
	go test ./... -race -shuffle=on -count=5 -v -timeout 15m

test-e2e:
	go test -tags e2e -v ./tests/e2e/...

clean:
	rm -f $(SERVICE)

gen:
	$(MAKE) gen-sqlc
	$(MAKE) gen-oapi
	go generate ./...

CACHE_DIR := .cache
$(shell mkdir -p $(CACHE_DIR))

CHKSM := $(shell command -v sha1sum >/dev/null && echo "sha1sum" || echo "md5 -q")


MIG_DEPS    := $(MIGRATIONS_DIR)
SQLC_DEPS   := internal/sqlc.yaml $(SCHEMA_OUT) ./sql/queries
OAPI_DEPS := internal/oapi-codegen-client.yaml internal/oapi-codegen.yaml openapi.yaml

STRIP := awk '{print $$1}'

# caching allows for go test cache to work properly, else on regen of the same e.g. schema.sql its invalidated.
get_hash = find $(1) -type f -not -path '*/.*' | sort | xargs $(CHKSM) | $(CHKSM) | $(STRIP)
is_changed = [ "$$($(call get_hash,$(1)))" != "$$(cat $(CACHE_DIR)/$(2) 2>/dev/null)" ]
update_cache = $(call get_hash,$(1)) > $(CACHE_DIR)/$(2)


gen-schema:
	if $(call is_changed,$(MIG_DEPS),mig_hash); then \
		echo "Migrations changed. Re-generating..."; \
		$(MAKE) run-gen-schema || { rm -f $(CACHE_DIR)/mig_hash; exit 1; }; \
		$(call update_cache,$(MIG_DEPS),mig_hash); \
	fi

gen-sqlc: gen-schema
	if $(call is_changed,$(SQLC_DEPS),sqlc_hash); then \
		echo "SQL changed. Updating SQLC..."; \
		$(SQLC) generate -f internal/sqlc.yaml || { rm -f $(CACHE_DIR)/sqlc_hash; exit 1; }; \
		$(call update_cache,$(SQLC_DEPS),sqlc_hash); \
	fi

gen-oapi:
	if $(call is_changed,$(OAPI_DEPS),oapi_hash); then \
		echo "OpenAPI spec/config changed. Updating..."; \
		go tool oapi-codegen -config internal/oapi-codegen.yaml openapi.yaml || { rm -f $(CACHE_DIR)/oapi_hash; exit 1; }; \
		go tool oapi-codegen -config internal/oapi-codegen-client.yaml openapi.yaml || { rm -f $(CACHE_DIR)/oapi_hash; exit 1; }; \
		$(call update_cache,$(OAPI_DEPS),oapi_hash); \
	fi

gen-cli:
	go generate ./cmd/cli
	go build -o todo-cli ./cmd/cli

run-gen-schema:
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

ws-listen:
	WS_URL=$$(echo "$$API_URL" | sed 's/^http:/ws:/' | sed 's/^https:/wss:/'); \
	if [ -z "$$API_TOKEN" ]; then \
		echo "Error: API_TOKEN is not set."; \
		exit 1; \
	fi; \
	wscat -H "Authorization: Bearer $$API_TOKEN" -c "$${WS_URL}/ws"

N ?= 5
QUEUE ?= todo_events

rabbitmq-messages:
	docker exec myapp-rabbitmq rabbitmqadmin -V / get queue="$(QUEUE)" count="$(N)" -f pretty_json | jq 'reverse'

rabbitmq-watch:
	docker exec myapp-rabbitmq rabbitmqadmin delete queue name=debug_tap > /dev/null 2>&1 || true
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
