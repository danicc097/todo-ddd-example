package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	rdb "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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

	ctx := context.Background()

	isProd := internal.Config.Env == "production"
	shutdown, err := logger.Init(internal.Config.LogLevel, isProd)
	if err != nil {
		os.Stderr.WriteString("logger init failed: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer shutdown(ctx)

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
	tm := db.NewTransactionManager(pool)

	todoRepo := todoPg.NewTodoRepositoryWithTracing(todoPg.NewTodoRepo(pool), "todo-ddd-api")
	tagRepo := todoPg.NewTagRepositoryWithTracing(todoPg.NewTagRepo(pool), "todo-ddd-api")
	userRepo := userPg.NewUserRepositoryWithTracing(userPg.NewUserRepo(pool), "todo-ddd-api")

	mqConn, err := amqp.Dial(internal.Config.RabbitMQ.URL)
	if err != nil {
		slog.Error("failed to connect to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}
	todoRabbitPub, err := todoMsg.NewRabbitMQPublisher(mqConn)
	if err != nil {
		slog.Error("failed to create Todo rabbitmq publisher", slog.String("error", err.Error()))
		os.Exit(1)
	}

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
	go relay.Start(ctx)

	r := gin.New()
	r.Use(otelgin.Middleware("todo-ddd-api"))
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.Idempotency(redisClient)) // let it cache <500 errors from ErrorHandler
	r.Use(middleware.ErrorHandler())

	r.StaticFile("/openapi.yaml", "./openapi.yaml")
	r.GET("/api/v1/docs", swaggerUIHandler("/openapi.yaml"))

	handler := &CompositeHandler{TodoHandler: th, UserHandler: uh}
	api.RegisterHandlers(r.Group("/api/v1"), handler)

	r.GET("/ws", th.WS)

	slog.InfoContext(ctx, "Application server starting", slog.String("port", internal.Config.Port))
	if err := r.Run(":" + internal.Config.Port); err != nil {
		slog.Error("Server exit", slog.String("error", err.Error()))
	}
}
