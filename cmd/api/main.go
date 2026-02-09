package main

import (
	"context"
	"log"
	"os"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoMsg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/messaging"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	redisPub "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	userApp "github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userHttp "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/http"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	rdb "github.com/redis/go-redis/v9"
)

type CompositeHandler struct {
	*todoHttp.TodoHandler
	*userHttp.UserHandler
}

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

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	redisClient := rdb.NewClient(&rdb.Options{Addr: redisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to redis: %v", err)
	}

	todoRepo := todoPg.NewTodoRepo(pool)
	publisher := redisPub.NewRedisPublisher(redisClient)
	hub := ws.NewTodoHub(redisClient)
	tm := db.NewTransactionManager(pool)

	th := todoHttp.NewTodoHandler(
		todoApp.NewCreateTodoUseCase(tm),
		todoApp.NewCompleteTodoUseCase(tm),
		todoApp.NewGetAllTodosUseCase(todoRepo),
		todoApp.NewGetTodoUseCase(todoRepo),
		hub,
	)

	userRepo := userPg.NewUserRepo(pool)
	uh := userHttp.NewUserHandler(
		userApp.NewRegisterUserUseCase(userRepo),
		userApp.NewGetUserUseCase(userRepo),
	)

	relay := outbox.NewRelay(pool)
	relay.Register("todo.created", todoMsg.MakeCreatedHandler(publisher))
	relay.Register("todo.completed", todoMsg.MakeUpdatedHandler(publisher))
	go relay.Start(ctx)

	handler := &CompositeHandler{TodoHandler: th, UserHandler: uh}

	r := gin.Default()
	api.RegisterHandlers(r.Group("/api/v1"), handler)
	r.GET("/ws", th.WS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	log.Fatal(r.Run(":" + port))
}
