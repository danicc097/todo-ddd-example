package application

import (
	"context"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
)

//go:generate go tool gowrap gen -g -i WorkspaceQueryService -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/workspace_query_service_tracing.gen.go
type WorkspaceQueryService interface {
	List(ctx context.Context) ([]api.Workspace, error)
}
