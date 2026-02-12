package http

import (
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type WorkspaceRestMapper struct{}

func (m *WorkspaceRestMapper) ToResponse(w *domain.Workspace) api.Workspace {
	return api.Workspace{
		Id:          w.ID(),
		Name:        w.Name(),
		Description: w.Description(),
	}
}
