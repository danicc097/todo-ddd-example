package http

const (
	IdempotencyKeyHeader   = "Idempotency-Key"
	AuthorizationHeader    = "Authorization"
	SkipRateLimitHeader    = "x-skip-rate-limit"
	BearerScheme           = "Bearer"
	DefaultPaginationLimit = 20

	APIV1Prefix = "/api/v1"

	RouteWS          = "/ws"
	RouteHealthz     = APIV1Prefix + "/healthz"
	RouteDocs        = APIV1Prefix + "/docs"
	RouteOpenAPISpec = "/openapi.yaml"
	RouteFavicon     = "/favicon.ico"
)

var SensitiveFields = map[string]struct{}{
	"password": {},
	"code":     {},
}
