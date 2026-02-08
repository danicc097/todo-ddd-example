package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	rdb "github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	pool, _ := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	redisClient := rdb.NewClient(&rdb.Options{Addr: "localhost:6379"})

	repo := postgres.NewTodoRepo(pool)
	publisher := redis.NewRedisPublisher(redisClient)
	hub := ws.NewTodoHub(redisClient)

	createUC := application.NewCreateTodoUseCase(repo)
	completeUC := application.NewCompleteTodoUseCase(repo, publisher)
	getAllUC := application.NewGetAllTodosUseCase(repo)

	handler := http.NewTodoHandler(createUC, completeUC, getAllUC, hub)

	r := gin.Default()
	r.GET("/ws", handler.WS)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/todos", handler.Create)
		v1.PATCH("/todos/:id/complete", handler.Complete)
	}

	ln, _ := net.Listen("tcp", ":0")
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	fmt.Printf("Running in port %s\n", port)

	r.RunListener(ln)
}
