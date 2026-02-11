package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	rdb "github.com/redis/go-redis/v9"
	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/danicc097/todo-ddd-example/internal"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/logger"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoMsg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/messaging"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	userApp "github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userHttp "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/http"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
)

type CompositeHandler struct {
	*todoHttp.TodoHandler
	*userHttp.UserHandler
}

func swaggerUIHandler(url string) gin.HandlerFunc {
	return func(c *gin.Context) {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerJS" />
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.31.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.31.0/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '%s',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>`, url)

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}

func main() {
	var envPath string
	flag.StringVar(&envPath, "env", ".env", "Environment Variables filename")
	flag.Parse()

	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			slog.Warn("failed to load env file", slog.String("path", envPath), slog.String("error", err.Error()))
		}
	}

	if err := internal.NewAppConfig(); err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isProd := internal.Config.Env == "production"

	shutdownLogger, err := logger.Init(ctx, internal.Config.LogLevel, isProd)
	if err != nil {
		os.Stderr.WriteString("logger init failed: " + err.Error() + "\n")
		os.Exit(1)
	}

	defer func() {
		_ = shutdownLogger(context.Background())
	}()

	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		internal.Config.Postgres.User,
		internal.Config.Postgres.Password,
		internal.Config.Postgres.Host,
		internal.Config.Postgres.Port,
		internal.Config.Postgres.DBName,
	)

	var pool *pgxpool.Pool
	for i := range 15 {
		pool, err = pgxpool.New(ctx, pgUrl)
		if err == nil {
			err = pool.Ping(ctx)
		}

		if err == nil {
			break
		}

		slog.Warn("Database not ready, retrying...", slog.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		slog.ErrorContext(ctx, "Unable to connect to database after retries", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer pool.Close()

	redisClient := rdb.NewClient(&rdb.Options{Addr: internal.Config.Redis.Addr})
	defer redisClient.Close()

	tm := db.NewTransactionManager(pool)

	todoRepo := todoPg.NewTodoRepositoryWithTracing(todoPg.NewTodoRepo(pool), "todo-ddd-api")
	tagRepo := todoPg.NewTagRepositoryWithTracing(todoPg.NewTagRepo(pool), "todo-ddd-api")
	userRepo := userPg.NewUserRepositoryWithTracing(userPg.NewUserRepo(pool), "todo-ddd-api")

	mqConn, err := rabbitmq.NewConn(
		internal.Config.RabbitMQ.URL,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
	)
	if err != nil {
		slog.Error("failed to connect to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		_ = mqConn.Close()
	}()

	todoRabbitPub, err := todoMsg.NewRabbitMQPublisher(mqConn)
	if err != nil {
		slog.Error("failed to create Todo rabbitmq publisher", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer todoRabbitPub.Close()

	redisPub := redis.NewRedisPublisher(redisClient)
	todoPublisher := todoMsg.NewMultiPublisher(todoRabbitPub, redisPub)

	hub := ws.NewTodoHub(redisClient)

	_ = tagRepo
	th := todoHttp.NewTodoHandler(
		todoApp.NewCreateTodoUseCase(tm),
		todoApp.NewCompleteTodoUseCase(tm),
		todoApp.NewGetAllTodosUseCase(todoRepo),
		todoApp.NewGetTodoUseCase(todoRepo),
		todoApp.NewCreateTagUseCase(tm),
		hub,
	)

	uh := userHttp.NewUserHandler(
		userApp.NewRegisterUserUseCase(userRepo),
		userApp.NewGetUserUseCase(userRepo),
	)

	relay := outbox.NewRelay(pool)
	relay.Register("todo.created", todoMsg.MakeCreatedHandler(todoPublisher))
	relay.Register("todo.completed", todoMsg.MakeUpdatedHandler(todoPublisher))
	relay.Register("todo.tagadded", todoMsg.MakeUpdatedHandler(todoPublisher))

	// Run relay in a goroutine with the main context
	// When main context is cancelled, relay will shut down gracefully
	go relay.Start(ctx)

	r := gin.New()
	r.Use(otelgin.Middleware("todo-ddd-api"))
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.Idempotency(redisClient))
	r.Use(middleware.ErrorHandler())

	r.StaticFile("/openapi.yaml", "./openapi.yaml")
	r.GET("/api/v1/docs", swaggerUIHandler("/openapi.yaml"))

	handler := &CompositeHandler{TodoHandler: th, UserHandler: uh}
	api.RegisterHandlers(r.Group("/api/v1"), handler)

	r.GET("/ws", th.WS)

	srv := &http.Server{
		Addr:    ":" + internal.Config.Port,
		Handler: r,
	}

	// Run server in goroutine
	go func() {
		slog.InfoContext(ctx, "Application server starting", slog.String("port", internal.Config.Port))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server listen error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful Shutdown Logic
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// 1. Cancel context to stop Relay loop and other background tasks
	cancel()

	// 2. Shutdown HTTP server with timeout
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTimeout()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	slog.Info("Server exiting")
}
