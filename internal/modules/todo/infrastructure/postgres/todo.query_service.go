package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type todoQueryService struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewTodoQueryService(pool *pgxpool.Pool) application.TodoQueryService {
	return &todoQueryService{
		q:    db.New(),
		pool: pool,
	}
}

func (s *todoQueryService) GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID, limit, offset int32) ([]application.TodoReadModel, error) {
	rows, err := s.q.ListTodosByWorkspaceID(ctx, s.pool, db.ListTodosByWorkspaceIDParams{
		WorkspaceID: wsID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, err
	}

	todos := make([]application.TodoReadModel, len(rows))
	for i, r := range rows {
		todos[i] = application.TodoReadModel{
			ID:          r.ID,
			WorkspaceID: r.WorkspaceID,
			Title:       r.Title,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		}
	}

	return todos, nil
}

func (s *todoQueryService) GetByID(ctx context.Context, id domain.TodoID) (*application.TodoReadModel, error) {
	row, err := s.q.GetTodoByID(ctx, s.pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}

		return nil, err
	}

	return &application.TodoReadModel{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		Title:       row.Title,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
	}, nil
}
