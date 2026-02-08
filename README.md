# Setup

```bash
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc
go generate ./...
# assume migrated db
DATABASE_URL=postgresql://postgres:postgres@localhost:5656/postgres go run cmd/api/main.go
```

# Example

```bash
# listen to todo updates:
wscat -c http://localhost:PORT/ws
>>> Connected (press CTRL+C to quit)
>>> < {"..."}
# create
curl -X POST http://localhost:PORT/api/v1/todos -d '{"title": "New todo"}'
>>> {"id":"c9e34c82-5b43-4e7e-a650-bca484057943"}
# complete
curl -X PATCH http://localhost:PORT/api/v1/todos/c9e34c82-5b43-4e7e-a650-bca484057943/complete
# list
curl -X GET http://localhost:PORT/api/v1/todos
>>> [{"ID":"c9e34c82-5b43-4e7e-a650-bca484057943","Title":"New todo","Completed":true,"CreatedAt":"2026-02-08T16:07:47.573757+01:00"}]
```
