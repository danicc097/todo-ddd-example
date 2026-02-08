package postgres

import (
	"context"
	"encoding/json"
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

func (r *TodoRepo) Save(ctx context.Context, t *domain.Todo) (uuid.UUID, error) {
	p := r.mapper.ToPersistence(t)
	_, err := r.q.CreateTodo(ctx, r.db, db.CreateTodoParams(p))
	return p.ID, err
}

func (r *TodoRepo) AddTag(ctx context.Context, todoID uuid.UUID, tagID uuid.UUID) error {
	return r.q.AddTagToTodo(ctx, r.db, db.AddTagToTodoParams{
		TodoID: todoID,
		TagID:  tagID,
	})
}

func (r *TodoRepo) SaveEvent(ctx context.Context, eventType string, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.q.SaveOutboxEvent(ctx, r.db, db.SaveOutboxEventParams{
		ID:        uuid.New(),
		EventType: eventType,
		Payload:   bytes,
	})
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
		todos[i] = r.mapper.ToDomain(row)
	}
	return todos, nil
}

func (r *TodoRepo) Update(ctx context.Context, t *domain.Todo) error {
	p := r.mapper.ToPersistence(t)
	return r.q.UpdateTodo(ctx, r.db, db.UpdateTodoParams{
		ID:     p.ID,
		Title:  p.Title,
		Status: p.Status,
	})
}
