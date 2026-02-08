package main

import (
	"context"
	"log"
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

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379" // Default to docker service name
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	redisClient := rdb.NewClient(&rdb.Options{Addr: redisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to redis: %v", err)
	}

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
		v1.GET("/todos", handler.GetAll)
		v1.POST("/todos", handler.Create)
		v1.PATCH("/todos/:id/complete", handler.Complete)
	}

	if err := r.Run(":8090"); err != nil {
		log.Fatal(err)
	}
}
