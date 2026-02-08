# Setup

```bash
./deploy.sh
# ...docker swarm cluster will be listening on port 8090
# poor man's migration:
docker compose exec -T db psql -U postgres -d postgres < sql/schema.sql
```

## Commands

- psql: `docker compose exec -ti db psql -U postgres -d postgres`

# Example

```bash
# listen to todo_updated:
wscat -c ws://localhost:8090/ws
>>> Connected (press CTRL+C to quit)
>>> < {"..."} # (will get notified regardless of node)
# create:
curl -X POST http://localhost:8090/api/v1/todos -d '{"title": "New todo"}'
>>> {"id":"c9e34c82-5b43-4e7e-a650-bca484057943"}
# complete:
curl -X PATCH http://localhost:8090/api/v1/todos/c9e34c82-5b43-4e7e-a650-bca484057943/complete
# ...will notify all todo_updated listeners
# list:
curl -X GET http://localhost:8090/api/v1/todos
>>> [{"ID":"c9e34c82-5b43-4e7e-a650-bca484057943","Title":"New todo","Completed":true,"CreatedAt":"2026-02-08T16:07:47.573757+01:00"}]
```
