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
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/extra/redisotel/v9"
	rdb "github.com/redis/go-redis/v9"
	"github.com/wagslane/go-rabbitmq"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gopkg.in/yaml.v3"

	"github.com/danicc097/todo-ddd-example/internal"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/logger"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	auditMem "github.com/danicc097/todo-ddd-example/internal/modules/audit/infrastructure/memory"
	authApp "github.com/danicc097/todo-ddd-example/internal/modules/auth/application"
	authHttp "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/http"
	authPg "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/postgres"
	authRedis "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/redis"
	todoApp "github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	todoDecorator "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/decorator"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoMsg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/messaging"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	todoRabbit "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/rabbitmq"
	todoRedis "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/redis"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	userApp "github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userAdapters "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/adapters"
	userHttp "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/http"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	wsApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	wsDecorator "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/decorator"
	wsHttp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/http"
	wsPg "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/postgres"
	sharedMiddleware "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/middleware"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

type CompositeHandler struct {
	*todoHttp.TodoHandler
	*userHttp.UserHandler
	*wsHttp.WorkspaceHandler
	*authHttp.AuthHandler
}

func (h *CompositeHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
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

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff") // prevent script injection
		c.Header("X-Frame-Options", "DENY")           // prevent clickjacking

		isDocs := c.Request.URL.Path == "/api/v1/docs"
		isProd := internal.Config.Env == internal.AppEnvProd

		if isProd || !isDocs {
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}

		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains") // force https for the next year
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")  // ensure possibly sensitive data is never cached
		c.Header("Pragma", "no-cache")                                               // for legacy http clients

		c.Next()
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
	cachedTodoRepo := todoDecorator.NewTodoRepositoryWithCache(
		baseTodoRepo,
		redisClient,
		5*time.Minute,
		todoCodec,
	)
	todoRepo := todoPg.NewTodoRepositoryWithTracing(cachedTodoRepo, "todo-ddd-api")

	baseTagRepo := todoPg.NewTagRepo(pool)
	tagCodec := todoRedis.NewTagCacheCodec()
	cachedTagRepo := todoDecorator.NewTagRepositoryWithCache(
		baseTagRepo,
		redisClient,
		60*time.Minute,
		tagCodec,
	)
	tagRepo := todoPg.NewTagRepositoryWithTracing(cachedTagRepo, "todo-ddd-api")

	baseUserRepo := userPg.NewUserRepo(pool)
	userRepo := userPg.NewUserRepositoryWithTracing(baseUserRepo, "todo-ddd-api")

	auditRepo := auditMem.NewAuditRepository()
	baseWsRepo := wsPg.NewWorkspaceRepo(pool)
	auditedWsRepo := wsDecorator.NewWorkspaceAuditWrapper(baseWsRepo, auditRepo)
	wsRepo := wsPg.NewWorkspaceRepositoryWithTracing(auditedWsRepo, "todo-ddd-api")

	if len(internal.Config.MFAMasterKey) != 32 {
		slog.Error("MFA_MASTER_KEY must be exactly 32 bytes")
		os.Exit(1)
	}

	privKeyBytes, err := os.ReadFile("private.pem")
	if err != nil {
		slog.Error("failed to read private key", slog.String("error", err.Error()))
		os.Exit(1)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		slog.Error("failed to parse private key", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pubKeyBytes, err := os.ReadFile("public.pem")
	if err != nil {
		slog.Error("failed to read public key", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		slog.Error("failed to parse public key", slog.String("error", err.Error()))
		os.Exit(1)
	}

	tokenIssuer := crypto.NewTokenIssuer(privKey, "todo-ddd-api")
	tokenVerifier := crypto.NewTokenVerifier(pubKey)

	baseAuthRepo := authPg.NewAuthRepo(pool)
	authRepo := authPg.NewAuthRepositoryWithTracing(baseAuthRepo, "todo-ddd-api")

	masterKey := []byte(internal.Config.MFAMasterKey)
	totpGuard := authRedis.NewTOTPGuard(redisClient)

	mqConn, err := rabbitmq.NewConn(
		internal.Config.RabbitMQ.URL,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
	)
	if err != nil {
		slog.Error("failed to connect to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() { _ = mqConn.Close() }()

	todoRabbitPub, err := todoRabbit.NewPublisher(mqConn, "todo_events")
	if err != nil {
		slog.Error("failed to create Todo rabbitmq publisher", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer todoRabbitPub.Close()

	redisPub := todoRedis.NewRedisPublisher(redisClient)
	todoPublisher := todoMsg.NewMultiPublisher(todoRabbitPub, redisPub)

	getUserUC := userApp.NewGetUserUseCase(userRepo)

	wsUserGateway := userAdapters.NewWorkspaceUserGateway(userRepo)

	onboardWsBase := wsApp.NewOnboardWorkspaceHandler(wsRepo, wsUserGateway)
	onboardWsHandler := sharedMiddleware.Transactional(pool, onboardWsBase)

	addWsMemberBase := wsApp.NewAddWorkspaceMemberHandler(wsRepo)
	addWsMemberHandler := sharedMiddleware.Transactional(pool, addWsMemberBase)

	removeWsMemberBase := wsApp.NewRemoveWorkspaceMemberHandler(wsRepo)
	removeWsMemberHandler := sharedMiddleware.Transactional(pool, removeWsMemberBase)

	deleteWsBase := wsApp.NewDeleteWorkspaceHandler(wsRepo)
	deleteWsHandler := sharedMiddleware.Transactional(pool, deleteWsBase)

	baseWorkspaceQueryService := wsPg.NewWorkspaceQueryService(pool)
	workspaceQueryService := wsPg.NewWorkspaceQueryServiceWithTracing(baseWorkspaceQueryService, "todo-ddd-api")

	hub := ws.NewTodoHub(redisClient, workspaceQueryService)

	createTodoBase := todoApp.NewCreateTodoHandler(todoRepo)
	createTodoHandler := sharedMiddleware.Transactional(pool, createTodoBase)

	completeTodoBase := todoApp.NewCompleteTodoHandler(todoRepo, wsRepo)
	completeTodoHandler := sharedMiddleware.Transactional(pool, completeTodoBase)

	createTagBase := todoApp.NewCreateTagHandler(tagRepo)
	createTagHandler := sharedMiddleware.Transactional(pool, createTagBase)

	assignTagToTodoBase := todoApp.NewAssignTagToTodoHandler(todoRepo)
	assignTagToTodoHandler := sharedMiddleware.Transactional(pool, assignTagToTodoBase)

	// queries bypass tx
	baseTodoQueryService := todoPg.NewTodoQueryService(pool)
	todoQueryService := todoPg.NewTodoQueryServiceWithTracing(baseTodoQueryService, "todo-ddd-api")

	loginHandler := authApp.NewLoginHandler(userRepo, authRepo, tokenIssuer)
	registerHandler := authApp.NewRegisterHandler(userRepo, authRepo)
	registerHandlerTx := sharedMiddleware.Transactional(pool, registerHandler)

	initiateTOTPHandler := authApp.NewInitiateTOTPHandler(authRepo, masterKey)
	verifyTOTPHandler := authApp.NewVerifyTOTPHandler(authRepo, totpGuard, tokenIssuer, masterKey)

	th := todoHttp.NewTodoHandler(
		createTodoHandler,
		completeTodoHandler,
		createTagHandler,
		assignTagToTodoHandler,
		todoQueryService,
		hub,
	)

	uh := userHttp.NewUserHandler(
		getUserUC,
		workspaceQueryService,
	)

	wh := wsHttp.NewWorkspaceHandler(
		onboardWsHandler,
		addWsMemberHandler,
		removeWsMemberHandler,
		workspaceQueryService,
		deleteWsHandler,
	)

	ah := authHttp.NewAuthHandler(loginHandler, registerHandlerTx, initiateTOTPHandler, verifyTOTPHandler)

	relay := outbox.NewRelay(pool)
	relay.Register("todo.created", todoRabbit.MakeCreatedHandler(todoPublisher))
	relay.Register("todo.completed", todoRabbit.MakeUpdatedHandler(todoPublisher))
	relay.Register("todo.tagadded", todoRabbit.MakeTagAddedHandler(todoPublisher))

	go relay.Start(ctx)

	r := gin.New()
	r.Use(otelgin.Middleware("todo-ddd-api"))
	r.Use(SecurityHeaders())
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.Idempotency(redisClient))
	r.Use(middleware.IdentityAndMFAResolver(tokenVerifier))
	r.Use(middleware.ErrorHandler())

	// non-owasp: missing cors based on consumers

	// load openapi spec explicitly to share the router with validation and rate limiting
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromFile("./openapi.yaml")
	if err != nil {
		slog.Error("failed to load openapi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	openapiRouter, err := gorillamux.NewRouter(doc)
	if err != nil {
		slog.Error("failed to create openapi router", slog.String("error", err.Error()))
		os.Exit(1)
	}

	r.Use(middleware.RateLimiter(redisClient, openapiRouter))

	validator := middleware.NewOpenapiMiddleware(doc).RequestValidatorWithOptions(&middleware.OAValidatorOptions{
		ValidateResponse: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	})

	r.Use(func(c *gin.Context) {
		p := c.Request.URL.Path
		if p == "/ws" ||
			p == "/api/v1/ping" ||
			p == "/api/v1/docs" ||
			p == "/favicon.ico" ||
			p == "/openapi.yaml" {
			c.Next()
			return
		}

		validator(c)
	})

	explodedSpec, err := getExplodedSpec("./openapi.yaml")
	if err != nil {
		slog.Error("failed to explode openapi spec", slog.String("error", err.Error()))
		os.Exit(1)
	}

	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/x-yaml", explodedSpec)
	})
	r.GET("/api/v1/docs", swaggerUIHandler("/openapi.yaml"))

	handler := &CompositeHandler{
		TodoHandler:      th,
		UserHandler:      uh,
		WorkspaceHandler: wh,
		AuthHandler:      ah,
	}
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

func getExplodedSpec(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var spec any

	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&spec); err != nil {
		return nil, err
	}

	return yaml.Marshal(spec)
}
