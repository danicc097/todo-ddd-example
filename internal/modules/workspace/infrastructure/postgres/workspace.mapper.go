package postgres

import (
	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type WorkspaceMapper struct{}

func (m *WorkspaceMapper) ToDomain(w db.Workspaces, members []db.WorkspaceMembers) (*domain.Workspace, error) {
	// TODO: ReconstituteWorkspace
	domWs := domain.NewWorkspace(w.Name, w.Description, members[0].UserID)

	return domWs, nil
}

func (m *WorkspaceMapper) ToPersistence(w *domain.Workspace) db.Workspaces {
	return db.Workspaces{
		ID:          w.ID(),
		Name:        w.Name(),
		Description: "",            // Add Description getter to entity if missing
		CreatedAt:   w.CreatedAt(), // Add CreatedAt getter
	}
}
