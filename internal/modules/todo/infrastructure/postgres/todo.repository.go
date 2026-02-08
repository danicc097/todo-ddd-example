package postgres

import (
	"context"
	"errors"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *TodoMapper
}

func NewTodoRepo(pool *pgxpool.Pool) *TodoRepo {
	return &TodoRepo{
		q:      db.New(),
		pool:   pool,
		mapper: &TodoMapper{},
	}
}

func (r *TodoRepo) Save(ctx context.Context, t *domain.Todo) (uuid.UUID, error) {
	p := r.mapper.ToPersistence(t)
	_, err := r.q.CreateTodo(ctx, r.pool, db.CreateTodoParams{
		ID:        p.ID,
		Title:     p.Title,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
	})
	return p.ID, err
}

func (r *TodoRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Todo, error) {
	row, err := r.q.GetTodoByID(ctx, r.pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}
		return nil, err
	}
	return r.mapper.ToDomain(row), nil
}

func (r *TodoRepo) FindAll(ctx context.Context) ([]*domain.Todo, error) {
	rows, err := r.q.ListTodos(ctx, r.pool)
	if err != nil {
		return nil, err
	}

	todos := make([]*domain.Todo, len(rows))
	for i, row := range rows {
		todos[i] = r.mapper.ToDomain(row)
	}
	return todos, nil
}

func (r *TodoRepo) Update(ctx context.Context, t *domain.Todo) error {
	p := r.mapper.ToPersistence(t)
	return r.q.UpdateTodo(ctx, r.pool, db.UpdateTodoParams{
		ID:     p.ID,
		Title:  p.Title,
		Status: p.Status,
	})
}
