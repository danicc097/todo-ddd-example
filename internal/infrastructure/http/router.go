package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/danicc097/todo-ddd-example/internal"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/crypto"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/messaging"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type RouterConfig struct {
	Env           internal.AppEnv
	Pool          *pgxpool.Pool
	Redis         *redis.Client
	TokenVerifier *crypto.TokenVerifier
	Handler       api.ServerInterface
	WSHandler     gin.HandlerFunc
}

func NewRouter(cfg RouterConfig) (*gin.Engine, error) {
	r := gin.New()
	r.Use(otelgin.Middleware(messaging.Keys.ServiceName()))
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.SecurityHeaders(cfg.Env))
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())
	r.Use(middleware.DBIdempotency(cfg.Pool))
	r.Use(middleware.IdentityAndMFAResolver(cfg.TokenVerifier))

	// load openapi spec explicitly to share the router with validation and rate limiting
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromFile("./openapi.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load openapi spec: %w", err)
	}

	openapiRouter, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create openapi router: %w", err)
	}

	r.Use(middleware.RateLimiter(cfg.Redis, openapiRouter))

	validator := middleware.NewOpenapiMiddleware(doc).RequestValidatorWithOptions(&middleware.OAValidatorOptions{
		ValidateResponse: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
				if c := causation.FromContext(ctx); c.IsUser() || c.IsSystem() {
					return nil
				}

				return errors.New("unauthorized")
			},
		},
	})

	r.Use(func(c *gin.Context) {
		p := c.Request.URL.Path
		if p == sharedHttp.RouteWS ||
			p == sharedHttp.RoutePing ||
			p == sharedHttp.RouteDocs ||
			p == sharedHttp.RouteFavicon ||
			p == sharedHttp.RouteOpenAPISpec {
			c.Next()
			return
		}

		validator(c)
	})

	explodedSpec, err := GetExplodedSpec("./openapi.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to explode openapi spec: %w", err)
	}

	r.GET(sharedHttp.RouteOpenAPISpec, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/x-yaml", explodedSpec)
	})
	r.GET(sharedHttp.RouteDocs, SwaggerUIHandler(sharedHttp.RouteOpenAPISpec))

	api.RegisterHandlers(r.Group(sharedHttp.APIV1Prefix), cfg.Handler)

	r.GET(sharedHttp.RouteWS, cfg.WSHandler)

	return r, nil
}
