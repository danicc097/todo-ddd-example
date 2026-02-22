package http

const (
	APIV1Prefix = "/api/v1"

	RouteWS          = "/ws"
	RoutePing        = APIV1Prefix + "/ping"
	RouteDocs        = APIV1Prefix + "/docs"
	RouteOpenAPISpec = "/openapi.yaml"
	RouteFavicon     = "/favicon.ico"
)
