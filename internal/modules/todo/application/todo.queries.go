package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

//go:generate go tool gowrap gen -g -i TodoQueryService -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/todo_query_service_tracing.gen.go
type TodoQueryService interface {
	GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID, limit, offset int32) ([]TodoReadModel, error)
	GetByID(ctx context.Context, id domain.TodoID) (*TodoReadModel, error)
}
