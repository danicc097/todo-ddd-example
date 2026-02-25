package http

const (
	IdempotencyKeyHeader   = "Idempotency-Key"
	AuthorizationHeader    = "Authorization"
	BearerScheme           = "Bearer"
	DefaultPaginationLimit = 20

	APIV1Prefix = "/api/v1"

	RouteWS          = "/ws"
	RoutePing        = APIV1Prefix + "/ping"
	RouteDocs        = APIV1Prefix + "/docs"
	RouteOpenAPISpec = "/openapi.yaml"
	RouteFavicon     = "/favicon.ico"
)

var SensitiveFields = map[string]struct{}{
	"password": {},
	"code":     {},
}
