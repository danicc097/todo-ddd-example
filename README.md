# Setup

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
# listen to todo_updated:
wscat -c ws://127.0.0.1:8090/ws
>>> Connected (press CTRL+C to quit)
>>> < {"..."} # (will get notified regardless of node)
# create:
curl -X POST http://127.0.0.1:8090/api/v1/todos -d '{"title": "New todo"}'
>>> {"id":"c9e34c82-5b43-4e7e-a650-bca484057943"}
# complete:
curl -X PATCH http://127.0.0.1:8090/api/v1/todos/c9e34c82-5b43-4e7e-a650-bca484057943/complete
# ...will notify all todo_updated listeners
# list:
curl -X GET http://127.0.0.1:8090/api/v1/todos
>>> [{"ID":"c9e34c82-5b43-4e7e-a650-bca484057943","Title":"New todo","Completed":true,"CreatedAt":"2026-02-08T16:07:47.573757+01:00"}]
```
