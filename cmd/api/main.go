package main

import (
	"context"
	"log/slog"
	"os"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/logger"
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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type CompositeHandler struct {
	*todoHttp.TodoHandler
	*userHttp.UserHandler
}

func main() {
	ctx := context.Background()

	isProd := os.Getenv("ENV") == "production"
	shutdown, err := logger.Init(os.Getenv("LOG_LEVEL"), isProd)
	if err != nil {
		os.Stderr.WriteString("logger init failed: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer shutdown(ctx)

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		slog.ErrorContext(ctx, "DATABASE_URL is not set")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.ErrorContext(ctx, "Database unreachable", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" { redisAddr = "redis:6379" }

	redisClient := rdb.NewClient(&rdb.Options{Addr: redisAddr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		slog.ErrorContext(ctx, "Unable to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
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

	uh := userHttp.NewUserHandler(
		userApp.NewRegisterUserUseCase(userPg.NewUserRepo(pool)),
		userApp.NewGetUserUseCase(userPg.NewUserRepo(pool)),
	)

	relay := outbox.NewRelay(pool)
	relay.Register("todo.created", todoMsg.MakeCreatedHandler(publisher))
	relay.Register("todo.completed", todoMsg.MakeUpdatedHandler(publisher))
	go relay.Start(ctx)

	r := gin.New()
	r.Use(otelgin.Middleware("todo-ddd-api"))
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())

	handler := &CompositeHandler{TodoHandler: th, UserHandler: uh}
	api.RegisterHandlers(r.Group("/api/v1"), handler)
	r.GET("/ws", th.WS)

	port := os.Getenv("PORT")
	if port == "" { port = "8090" }

	slog.InfoContext(ctx, "Application server starting", slog.String("port", port))
	if err := r.Run(":" + port); err != nil {
		slog.ErrorContext(ctx, "Server exit", slog.String("error", err.Error()))
	}
}
