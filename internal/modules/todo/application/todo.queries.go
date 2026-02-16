package application

import (
	"context"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

//go:generate go tool gowrap gen -g -i TodoQueryService -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/todo_query_service_tracing.gen.go
//go:generate go tool gowrap gen -g -i TodoQueryService -t ../../../../templates/cache.gotmpl -o ../infrastructure/decorator/todo_query_service_cache.gen.go -v "KeyPrefix=todo_query" -v "EntityType=*api.Todo"
type TodoQueryService interface {
	GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]api.Todo, error)
	GetByID(ctx context.Context, id domain.TodoID) (*api.Todo, error)
}
