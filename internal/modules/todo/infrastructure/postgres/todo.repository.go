package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
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

func (r *TodoRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *TodoRepo) Save(ctx context.Context, t *domain.Todo) error {
	dbtx := r.getDB(ctx) // Dynamic resolution
	p := r.mapper.ToPersistence(t)

	_, err := r.q.CreateTodo(ctx, dbtx, db.CreateTodoParams(p))
	if err != nil {
		return err
	}

	for _, tagID := range t.Tags() {
		err := r.q.AddTagToTodo(ctx, dbtx, db.AddTagToTodoParams{
			TodoID: t.ID(),
			TagID:  tagID,
		})
		if err != nil {
			return err
		}
	}

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, t)
}

func (r *TodoRepo) Update(ctx context.Context, t *domain.Todo) error {
	dbtx := r.getDB(ctx)
	p := r.mapper.ToPersistence(t)

	err := r.q.UpdateTodo(ctx, dbtx, db.UpdateTodoParams{
		ID:     p.ID,
		Title:  p.Title,
		Status: p.Status,
	})
	if err != nil {
		return err
	}

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, t)
}

func (r *TodoRepo) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
	row, err := r.q.GetTodoByID(ctx, r.getDB(ctx), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}

		return nil, err
	}

	return r.mapper.ToDomain(row), nil
}

func (r *TodoRepo) FindAll(ctx context.Context) ([]*domain.Todo, error) {
	rows, err := r.q.ListTodos(ctx, r.getDB(ctx))
	if err != nil {
		return nil, err
	}

	todos := make([]*domain.Todo, len(rows))
	for i, row := range rows {
		todos[i] = r.mapper.ListRowToDomain(row)
	}

	return todos, nil
}
