# todo-ddd-example

[![Go Report Card](https://goreportcard.com/badge/github.com/danicc097/todo-ddd-example)](https://goreportcard.com/report/github.com/danicc097/todo-ddd-example)
[![tests](https://github.com/danicc097/todo-ddd-example/actions/workflows/tests.yaml/badge.svg)](https://github.com/danicc097/todo-ddd-example/actions/workflows/tests.yaml)

## Stack

- **Architecture**: Follows/Inspired by **Clean Architecture** with **DDD** and
  basic **CQRS**. Caching, tracing and auditing via decorators with `gowrap`.
- **Database:** **PostgreSQL** with **pgroll** for zero-downtime schema migrations. **sqlc** for type-safe queries.
- **API:** Contract-first via **OpenAPI 3.0** with `oapi-codegen`.
  - Automatic request/response validation.
  - Spec-defined rate limiting.
  - Idempotency keys to let clients safely handle non-idempotent request retries.
- **Observability:** **OpenTelemetry** with **Jaeger** and **Prometheus**.
- **Messaging:** **RabbitMQ** for events and **Redis PubSub** for cross-node
  WebSocket synchronization. Transactional outbox pattern with at-least-once delivery.
- **Tooling:** Custom generated CLI client from the OpenAPI spec with completion if using `direnv`.
- **Infra:** **Docker swarm** for multinode deployment with Caddy.
- **CI:** See `.github/workflows/tests.yaml`.

## Setup

```bash
make deploy
```

## Test

```bash
make test
```

## Web UIs

- API docs: http://127.0.0.1:8090/api/v1/docs
- RabbitMQ: http://127.0.0.1:15672/
- Prometheus: http://127.0.0.1:9090/
- Jaeger: http://127.0.0.1:16686/search

# Example API usage

Full flow with MFA via generated `todo-cli` client:

```bash
scripts/example-flow.sh # with local swarm
# or
API_URL="http://localhost:8099" scripts/example-flow.sh # with make dev
```

# Websockets and message queues

First register, login and create a workspace:

```bash
./todo-cli register -p '{"email": "user@example.com", "name": "User", "password": "Password123!"}'
export API_TOKEN=$(./todo-cli login -p '{"email": "user@example.com", "password": "Password123!"}' | jq -r .accessToken)

export WS_ID=$(./todo-cli onboard-workspace -p '{"name": "My Workspace"}' | jq -r .id)
```

## Websockets:

```bash
$ make ws-listen
>>> Connected (press CTRL+C to quit)
# < {"event":"todo.created","id":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8","status":"PENDING","title":"New todo 1770748039"}
# < {"event":"todo.updated","id":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8","status":"COMPLETED","title":"New todo 1770748039"}
$ ./todo-cli create-todo $WS_ID -p '{"title": "New todo"}'
{"id":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8"}
...
$ ./todo-cli complete-todo ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8
...
$ ./todo-cli get-workspace-todos $WS_ID
>>> [{"id":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8",...}]
```

## Rabbitmq messages:

```bash
$ make rabbitmq-watch
>>> Tailing live events on 'todo_events'...
...

$ ./todo-cli complete-todo $(./todo-cli create-todo $WS_ID -p '{"title": "New todo"}' | jq -r .id)
# ...will show "todo.created" and "todo.completed" messages in watcher
```
