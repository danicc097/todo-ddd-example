ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

SHELL := /bin/bash
.SILENT:

KNOWN_TARGETS := test test-race test-e2e lint db-drop-dev clean deps lint dev gen gen-sqlc gen-cli gen-schema db-init migrate-up gen-oapi gen-k6 deploy psql logs run-gen-schema k8s-teardown k8s-validate ws-listen rabbitmq-messages rabbitmq-queues rabbitmq-exchanges rabbitmq-bindings rabbitmq-watch bench-seed bench bench-prometheus



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

NAMESPACE ?= myapp
KUBECTL_PSQL = kubectl exec -n $(NAMESPACE) sts/postgres -- env PGPASSWORD=$(DB_PASS) psql -U $(DB_USER)

.PHONY: $(KNOWN_TARGETS)

deps:
	go mod download
	$(SQLC) version
	$(PGROLL) --version

lint:
	go build ./... >/dev/null
	go test -c ./tests/e2e/... -tags e2e -o /dev/null
	@echo ">>> Running custom architectural analyzer..."
	go run ./tools/archlint/cmd/archlint ./...
	@echo ">>> Running critical linters..."
	$(GOLINT) run ./... --allow-parallel-runners --config=.golangci.yml --issues-exit-code=1 --enable-only depguard,exhaustruct,wrapcheck,contextcheck
	@echo ">>> Running linters and fix"
	$(GOLINT) run ./... --allow-parallel-runners --fix --config=.golangci.yml --issues-exit-code=0 >/dev/null || true

dev:
	$(AIR) -c .air.toml

test:
	$(MAKE) gen-schema
	go test ./... -count=1 $(ARGS)

test-race:
	$(MAKE) gen-schema
	go test ./... -race -shuffle=on -count=5 -timeout 15m $(ARGS)

test-e2e:
	go test -tags e2e -v ./tests/e2e/... $(ARGS)

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
K6_DEPS   := openapi.yaml

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

gen-k6:
	if $(call is_changed,$(K6_DEPS),k6_hash); then \
		echo "OpenAPI spec changed. Updating k6 client..."; \
		for dep in yq openapi-to-k6; do \
			command -v $$dep >/dev/null || { echo "Skipping k6 codegen: $$dep not found"; exit 0; }; \
		done; \
		api_path="$$(mktemp /tmp/todo-openapi.XXXXXX.yaml)"; \
		yq 'explode(.)' openapi.yaml > "$$api_path"; \
		openapi-to-k6 "$$api_path" scripts/k6 --verbose; \
		rm -f "$$api_path"; \
		$(call update_cache,$(K6_DEPS),k6_hash); \
	fi

gen-cli:
	go generate ./cmd/cli
	go build -o todo-cli ./cmd/cli

run-gen-schema:
	if ! kubectl get sts/postgres -n $(NAMESPACE) &>/dev/null; then \
		echo "Error: Postgres not found in namespace $(NAMESPACE). Run 'make deploy' first."; exit 1; \
	fi
	$(KUBECTL_PSQL) -d postgres -c "DROP DATABASE IF EXISTS $(GEN_DB);" >/dev/null
	$(KUBECTL_PSQL) -d postgres -c "CREATE DATABASE $(GEN_DB);" >/dev/null
	$(PGROLL) --postgres-url "$(GEN_PG_URL)" init
	find $(MIGRATIONS_DIR) -name "*.json" | sort | xargs -I % $(PGROLL) --postgres-url "$(GEN_PG_URL)" start --complete %
	kubectl exec -n $(NAMESPACE) sts/postgres -- env PGPASSWORD=$(DB_PASS) pg_dump -s -x -n public -U $(DB_USER) -d $(GEN_DB) \
		| grep -v '^\\' \
		| grep -v '^--' \
		| sed '/^$$/d' \
		> $(SCHEMA_OUT)

	$(KUBECTL_PSQL) -d postgres -c "DROP DATABASE IF EXISTS $(GEN_DB);" >/dev/null

db-init:
	$(PGROLL) --postgres-url "$(PG_URL)" init

db-drop-dev:
	read -p "Type 'DROP' to delete and recreate $(DB_NAME): " ans; \
	[ "$$ans" = "DROP" ] || (echo "Aborted." && exit 1)
	$(KUBECTL_PSQL) -d template1 -c "DROP DATABASE IF EXISTS $(DB_NAME) WITH (FORCE);"
	$(KUBECTL_PSQL) -d template1 -c "CREATE DATABASE $(DB_NAME);"
	echo "You can now run 'make migrate-up'."

# Idempotent migration target
migrate-up:
	echo "Running migrations against $(PG_URL)..."
	$(PGROLL) --postgres-url "$(PG_URL)" init 2>/dev/null || true

	for file in $$(find $(MIGRATIONS_DIR) -name "*.json" | sort); do \
		NAME=$$(basename $$file .json); \
		EXISTS=$$($(KUBECTL_PSQL) -d $(DB_NAME) -tAc "SELECT EXISTS(SELECT 1 FROM pgroll.migrations WHERE name = '$$NAME')"); \
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
	kubectl exec -it -n $(NAMESPACE) sts/postgres -- env PGPASSWORD=$(DB_PASS) psql -U $(DB_USER) -d $(DB_NAME)

logs:
	kubectl logs -f -n $(NAMESPACE) -l app.kubernetes.io/name=go-app --all-containers

k8s-teardown:
	kind delete cluster --name myapp

k8s-validate:
	./scripts/k8s-validate.sh

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
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin -V / get queue="$(QUEUE)" count="$(N)" -f pretty_json | jq 'reverse'

rabbitmq-watch:
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin delete queue name=debug_tap > /dev/null 2>&1 || true
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin declare queue name=debug_tap auto_delete=true > /dev/null
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin declare binding source=$(QUEUE) destination=debug_tap routing_key="#" > /dev/null

	@echo ">>> Tailing live events on '$(QUEUE)'..."
	@# ack_requeue_false: remove message after read.
	while true; do \
		kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin get queue=debug_tap count=10 ackmode=ack_requeue_false -f pretty_json \
		| jq -r '.[]?'; \
		sleep 0.5; \
	done

rabbitmq-queues:
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin -V / list queues name messages messages_ready consumers -f table

rabbitmq-exchanges:
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin -V / list exchanges name type -f table

rabbitmq-bindings:
	kubectl exec -n $(NAMESPACE) deploy/rabbitmq -- rabbitmqadmin -V / list bindings source destination routing_key -f table

bench-seed:
	RESEED=1 API_URL="$(API_URL)" SCENARIO=$(or $(SCENARIO),load) bash scripts/k6/run.sh

bench:
	API_URL="$(API_URL)" SCENARIO=$(or $(SCENARIO),load) bash scripts/k6/run.sh

bench-prometheus:
	API_URL="$(API_URL)" SCENARIO=$(or $(SCENARIO),load) K6_PUSH_PROMETHEUS=1 bash scripts/k6/run.sh
