# Setup

```bash
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc
go generate ./...
# assume migrated db
DATABASE_URL=postgresql://postgres:postgres@localhost:5656/postgres go run cmd/api/main.go
```

# Example

```bash
curl -X POST http://localhost:8082/api/v1/todos -d '{"title": "New todo"}'
>>> {"id":"c9e34c82-5b43-4e7e-a650-bca484057943"}
curl -X PATCH http://localhost:8082/api/v1/todos/c9e34c82-5b43-4e7e-a650-bca484057943/complete
```
