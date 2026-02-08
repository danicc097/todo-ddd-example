package main

import (
	"context"
	"log"
	"os"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
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
		redisAddr = "redis:6379"
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
	getTodoUC := application.NewGetTodoUseCase(repo)

	th := http.NewTodoHandler(createUC, completeUC, getAllUC, getTodoUC, hub)

	r := gin.Default()

	api.RegisterHandlers(r.Group("/api/v1"), th)

	r.GET("/ws", th.WS)

	if err := r.Run(":8090"); err != nil {
		log.Fatal(err)
	}
}
