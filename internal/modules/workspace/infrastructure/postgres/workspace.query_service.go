package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

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

func (s *workspaceQueryService) List(ctx context.Context, limit, offset int32) ([]application.WorkspaceReadModel, error) {
	rows, err := s.q.ListWorkspaces(ctx, s.pool, db.ListWorkspacesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	workspaces := make([]application.WorkspaceReadModel, len(rows))
	for i, r := range rows {
		workspaces[i] = application.WorkspaceReadModel{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}

	return workspaces, nil
}

func (s *workspaceQueryService) ListByUserID(ctx context.Context, userID userDomain.UserID) ([]application.WorkspaceReadModel, error) {
	rows, err := s.q.ListWorkspacesByUserID(ctx, s.pool, userID)
	if err != nil {
		return nil, err
	}

	workspaces := make([]application.WorkspaceReadModel, len(rows))
	for i, r := range rows {
		workspaces[i] = application.WorkspaceReadModel{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}

	return workspaces, nil
}

func (s *workspaceQueryService) ListTagsByWorkspaceID(ctx context.Context, workspaceID domain.WorkspaceID) ([]application.TagReadModel, error) {
	rows, err := s.q.ListTagsByWorkspaceID(ctx, s.pool, workspaceID)
	if err != nil {
		return nil, err
	}

	tags := make([]application.TagReadModel, len(rows))
	for i, r := range rows {
		tags[i] = application.TagReadModel{
			ID:   r.ID,
			Name: r.Name,
		}
	}

	return tags, nil
}
