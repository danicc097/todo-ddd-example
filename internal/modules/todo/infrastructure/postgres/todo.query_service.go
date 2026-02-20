package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
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

func (s *todoQueryService) GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]api.Todo, error) {
	rows, err := s.q.ListTodosByWorkspaceID(ctx, s.pool, wsID)
	if err != nil {
		return nil, err
	}

	todos := make([]api.Todo, len(rows))
	for i, r := range rows {
		todos[i] = api.Todo{
			Id:          r.ID,
			WorkspaceId: r.WorkspaceID,
			Title:       r.Title,
			Status:      api.TodoStatus(r.Status),
			CreatedAt:   r.CreatedAt,
		}
	}

	return todos, nil
}

func (s *todoQueryService) GetByID(ctx context.Context, id domain.TodoID) (*api.Todo, error) {
	row, err := s.q.GetTodoByID(ctx, s.pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}

		return nil, err
	}

	return &api.Todo{
		Id:        row.ID,
		Title:     row.Title,
		Status:    api.TodoStatus(row.Status),
		CreatedAt: row.CreatedAt,
	}, nil
}
