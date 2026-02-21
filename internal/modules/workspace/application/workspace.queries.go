package application

import (
	"context"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

//go:generate go tool gowrap gen -g -i WorkspaceQueryService -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/workspace_query_service_tracing.gen.go
//counterfeiter:generate . WorkspaceQueryService
type WorkspaceQueryService interface {
	List(ctx context.Context, limit, offset int32) ([]WorkspaceReadModel, error)
	ListByUserID(ctx context.Context, userID userDomain.UserID) ([]WorkspaceReadModel, error)
	ListTagsByWorkspaceID(ctx context.Context, workspaceID domain.WorkspaceID) ([]TagReadModel, error)
}
