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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/extra/redisotel/v9"
	rdb "github.com/redis/go-redis/v9"
	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/danicc097/todo-ddd-example/internal"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/logger"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/decorator"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoMsg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/messaging"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	todoRabbit "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/rabbitmq"
	todoRedis "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
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

	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		slog.Error("failed to instrument redis", slog.String("error", err.Error()))
	}

	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		slog.Error("failed to instrument redis metrics", slog.String("error", err.Error()))
	}

	baseTodoRepo := todoPg.NewTodoRepo(pool)
	todoCodec := todoRedis.NewTodoCacheCodec()
	cachedTodoRepo := decorator.NewTodoRepositoryWithCache(
		baseTodoRepo,
		redisClient,
		5*time.Minute,
		todoCodec,
	)
	todoRepo := todoPg.NewTodoRepositoryWithTracing(cachedTodoRepo, "todo-ddd-api")

	baseTagRepo := todoPg.NewTagRepo(pool)
	tagCodec := todoRedis.NewTagCacheCodec()
	cachedTagRepo := decorator.NewTagRepositoryWithCache(
		baseTagRepo,
		redisClient,
		60*time.Minute,
		tagCodec,
	)
	tagRepo := todoPg.NewTagRepositoryWithTracing(cachedTagRepo, "todo-ddd-api")

	// User Repo (No cache)
	baseUserRepo := userPg.NewUserRepo(pool)
	userRepo := userPg.NewUserRepositoryWithTracing(baseUserRepo, "todo-ddd-api")

	mqConn, err := rabbitmq.NewConn(
		internal.Config.RabbitMQ.URL,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
	)
	if err != nil {
		slog.Error("failed to connect to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() { _ = mqConn.Close() }()

	todoRabbitPub, err := todoRabbit.NewPublisher(mqConn)
	if err != nil {
		slog.Error("failed to create Todo rabbitmq publisher", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer todoRabbitPub.Close()

	redisPub := todoRedis.NewRedisPublisher(redisClient)
	todoPublisher := todoMsg.NewMultiPublisher(todoRabbitPub, redisPub)

	hub := ws.NewTodoHub(redisClient)

	createTodoBase := todoApp.NewCreateTodoUseCase(todoRepo)
	createTodoUC := decorator.NewCreateTodoUseCaseWithTransaction(createTodoBase, pool)

	completeTodoBase := todoApp.NewCompleteTodoUseCase(todoRepo)
	completeTodoUC := decorator.NewCompleteTodoUseCaseWithTransaction(completeTodoBase, pool)

	createTagBase := todoApp.NewCreateTagUseCase(tagRepo)
	createTagUC := decorator.NewCreateTagUseCaseWithTransaction(createTagBase, pool)

	getAllTodosUC := todoApp.NewGetAllTodosUseCase(todoRepo)
	getTodoUC := todoApp.NewGetTodoUseCase(todoRepo)

	registerUserUC := userApp.NewRegisterUserUseCase(userRepo)
	getUserUC := userApp.NewGetUserUseCase(userRepo)

	th := todoHttp.NewTodoHandler(
		createTodoUC,
		completeTodoUC,
		getAllTodosUC,
		getTodoUC,
		createTagUC,
		hub,
	)

	uh := userHttp.NewUserHandler(
		registerUserUC,
		getUserUC,
	)

	relay := outbox.NewRelay(pool)
	relay.Register("todo.created", todoRabbit.MakeCreatedHandler(todoPublisher))
	relay.Register("todo.completed", todoRabbit.MakeUpdatedHandler(todoPublisher))
	relay.Register("todo.tagadded", todoRabbit.MakeTagAddedHandler(todoPublisher))

	go relay.Start(ctx)

	r := gin.New()
	r.Use(otelgin.Middleware("todo-ddd-api"))
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.Idempotency(redisClient))
	r.Use(middleware.ETag()) // HTTP Caching
	r.Use(middleware.ErrorHandler())

	if internal.Config.Env != "production" {
		validator := createOpenAPIValidatorMw()

		r.Use(func(c *gin.Context) {
			p := c.Request.URL.Path
			if p == "/ws" ||
				p == "/api/v1/docs" ||
				p == "/openapi.yaml" {
				c.Next()
				return
			}

			validator(c)
		})
	}

	r.StaticFile("/openapi.yaml", "./openapi.yaml")
	r.GET("/api/v1/docs", swaggerUIHandler("/openapi.yaml"))

	handler := &CompositeHandler{TodoHandler: th, UserHandler: uh}
	api.RegisterHandlers(r.Group("/api/v1"), handler)

	r.GET("/ws", th.WS)

	srv := &http.Server{
		Addr:    ":" + internal.Config.Port,
		Handler: r,
	}

	go func() {
		slog.InfoContext(ctx, "Application server starting", slog.String("port", internal.Config.Port))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server listen error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	cancel()

	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTimeout()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	slog.Info("Server exiting")
}

func createOpenAPIValidatorMw() gin.HandlerFunc {
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromFile("./openapi.yaml")
	if err != nil {
		slog.Error("failed to load openapi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	oaMiddleware := middleware.NewOpenapiMiddleware(doc)

	validatorOpts := &middleware.OAValidatorOptions{
		ValidateResponse: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	}

	return oaMiddleware.RequestValidatorWithOptions(validatorOpts)
}
