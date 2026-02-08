package postgres

import (
	"context"
	"errors"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres/generated/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewTodoRepo(pool *pgxpool.Pool) *TodoRepo {
	return &TodoRepo{
		q:    db.New(),
		pool: pool,
	}
}

func (r *TodoRepo) Save(ctx context.Context, t *domain.Todo) (uuid.UUID, error) {
	_, err := r.q.CreateTodo(ctx, r.pool, db.CreateTodoParams{
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed,
		CreatedAt: t.CreatedAt,
	})
	return t.ID, err
}

func (r *TodoRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Todo, error) {
	row, err := r.q.GetTodoByID(ctx, r.pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}
		return nil, err
	}

	return &domain.Todo{
		ID:        row.ID,
		Title:     row.Title,
		Completed: row.Completed,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *TodoRepo) Update(ctx context.Context, t *domain.Todo) error {
	return r.q.UpdateTodo(ctx, r.pool, db.UpdateTodoParams{
		ID:        t.ID,
		Title:     t.Title,
		Completed: t.Completed,
	})
}
