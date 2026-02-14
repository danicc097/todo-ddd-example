package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
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

func (s *workspaceQueryService) ListByUserID(ctx context.Context, userID userDomain.UserID) ([]api.Workspace, error) {
	rows, err := s.q.ListWorkspacesByUserID(ctx, s.pool, userID)
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

func (s *workspaceQueryService) ListTagsByWorkspaceID(ctx context.Context, workspaceID domain.WorkspaceID) ([]api.Tag, error) {
	rows, err := s.q.ListTagsByWorkspaceID(ctx, s.pool, workspaceID)
	if err != nil {
		return nil, err
	}

	tags := make([]api.Tag, len(rows))
	for i, r := range rows {
		tags[i] = api.Tag{
			Id:   r.ID,
			Name: r.Name,
		}
	}

	return tags, nil
}
