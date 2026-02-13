package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
)

type workspaceQueryService struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewWorkspaceQueryService(pool *pgxpool.Pool) application.WorkspaceQueryService {
	return &workspaceQueryService{
		q:    db.New(),
		pool: pool,
	}
}

func (s *workspaceQueryService) List(ctx context.Context) ([]api.Workspace, error) {
	rows, err := s.q.ListWorkspaces(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	workspaces := make([]api.Workspace, len(rows))
	for i, r := range rows {
		workspaces[i] = api.Workspace{
			Id:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}

	return workspaces, nil
}
