# todo-ddd-example

[![Go Report Card](https://goreportcard.com/badge/github.com/danicc097/todo-ddd-example)](https://goreportcard.com/report/github.com/danicc097/todo-ddd-example)
[![tests](https://github.com/danicc097/todo-ddd-example/actions/workflows/tests.yaml/badge.svg)](https://github.com/danicc097/todo-ddd-example/actions/workflows/tests.yaml)

## Stack

- **Architecture**: Follows/Inspired by **Clean Architecture**.
- **Database:** **PostgreSQL** with **pgroll** allows for zero-downtime schema
  migrations. **sqlc** for compile-time checked queries.
- **API:** Contract-first with **OpenAPI 3.0** with **oapi-codegen**.
- **Observability:** **OpenTelemetry** with **Jaeger** and **Prometheus**.
- **Messaging:** **RabbitMQ** for events and **Redis PubSub** for cross-node WebSocket synchronization. Transactional outbox pattern and dead letter queue implementations.
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

```bash
$ make ws-listen
>>> Connected (press CTRL+C to quit)
# < {"event":"todo.created","id":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8","status":"PENDING","title":"New todo 1770748039"}
# < {"event":"todo.updated","id":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8","status":"COMPLETED","title":"New todo 1770748039"}
$ make req-create
{
  "id": "ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8"
}
...
$ make req-complete ID=ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8
...
$ make req-list
>>> [{"ID":"ae1e2ddc-5880-4f9a-8c3f-1d1fae16fbd8",...}]
```
