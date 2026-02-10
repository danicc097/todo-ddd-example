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
	db     db.DBTX
	mapper *TodoMapper
}

func NewTodoRepo(pool *pgxpool.Pool) *TodoRepo {
	return &TodoRepo{
		q:      db.New(),
		db:     pool,
		mapper: &TodoMapper{},
	}
}

func NewTodoRepoFromTx(tx pgx.Tx) *TodoRepo {
	return &TodoRepo{
		q:      db.New(),
		db:     tx,
		mapper: &TodoMapper{},
	}
}

func (r *TodoRepo) Save(ctx context.Context, t *domain.Todo) error {
	p := r.mapper.ToPersistence(t)
	_, err := r.q.CreateTodo(ctx, r.db, db.CreateTodoParams(p))
	if err != nil {
		return err
	}

	for _, tagID := range t.Tags() {
		err := r.q.AddTagToTodo(ctx, r.db, db.AddTagToTodoParams{
			TodoID: t.ID(),
			TagID:  tagID,
		})
		if err != nil {
			return err
		}
	}

	return r.saveDomainEvents(ctx, t)
}

func (r *TodoRepo) Update(ctx context.Context, t *domain.Todo) error {
	p := r.mapper.ToPersistence(t)
	err := r.q.UpdateTodo(ctx, r.db, db.UpdateTodoParams{
		ID:     p.ID,
		Title:  p.Title,
		Status: p.Status,
	})
	if err != nil {
		return err
	}

	return r.saveDomainEvents(ctx, t)
}

func (r *TodoRepo) saveDomainEvents(ctx context.Context, t *domain.Todo) error {
	for _, e := range t.Events() {
		eventName, payload, err := r.mapper.MapEvent(e)
		if err != nil {
			return err
		}
		if payload == nil {
			continue
		}

		if err := r.q.SaveOutboxEvent(ctx, r.db, db.SaveOutboxEventParams{
			ID:        uuid.New(),
			EventType: eventName,
			Payload:   payload,
		}); err != nil {
			return err
		}
	}
	t.ClearEvents()
	return nil
}

func (r *TodoRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Todo, error) {
	row, err := r.q.GetTodoByID(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}
		return nil, err
	}
	return r.mapper.ToDomain(row), nil
}

func (r *TodoRepo) FindAll(ctx context.Context) ([]*domain.Todo, error) {
	rows, err := r.q.ListTodos(ctx, r.db)
	if err != nil {
		return nil, err
	}
	todos := make([]*domain.Todo, len(rows))
	for i, row := range rows {
		todos[i] = r.mapper.ListRowToDomain(row)
	}
	return todos, nil
}
