package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type TodoRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *TodoMapper
	uow    application.UnitOfWork
}

func NewTodoRepo(pool *pgxpool.Pool, uow application.UnitOfWork) *TodoRepo {
	return &TodoRepo{
		q:      db.New(),
		pool:   pool,
		mapper: &TodoMapper{},
		uow:    uow,
	}
}

func (r *TodoRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *TodoRepo) Save(ctx context.Context, todo *domain.Todo) error {
	dbtx := r.getDB(ctx)
	p := r.mapper.ToPersistence(todo)

	_, err := r.q.UpsertTodo(ctx, dbtx, db.UpsertTodoParams{
		ID:                 p.ID,
		Title:              p.Title,
		Status:             p.Status,
		CreatedAt:          p.CreatedAt,
		WorkspaceID:        p.WorkspaceID,
		DueDate:            p.DueDate,
		RecurrenceInterval: p.RecurrenceInterval,
		RecurrenceAmount:   p.RecurrenceAmount,
		LastCompletedAt:    p.LastCompletedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert todo %s: %w", todo.ID(), sharedPg.ParseDBError(err))
	}

	tagUUIDs := make([]uuid.UUID, len(todo.Tags()))
	for i, tagID := range todo.Tags() {
		tagUUIDs[i] = tagID.UUID()
	}

	err = r.q.RemoveMissingTagsFromTodo(ctx, dbtx, db.RemoveMissingTagsFromTodoParams{
		TodoID: todo.ID(),
		Tags:   tagUUIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to sync tags for todo %s: %w", todo.ID(), sharedPg.ParseDBError(err))
	}

	if len(tagUUIDs) > 0 {
		todoIDs := make([]uuid.UUID, len(tagUUIDs))
		for i := range tagUUIDs {
			todoIDs[i] = todo.ID().UUID()
		}

		err := r.q.BulkAddTagsToTodo(ctx, dbtx, db.BulkAddTagsToTodoParams{
			TodoIds: todoIDs,
			TagIds:  tagUUIDs,
		})
		if err != nil {
			return fmt.Errorf("failed to bulk add tags to todo %s: %w", todo.ID(), sharedPg.ParseDBError(err))
		}
	}

	sessionIDs := make([]uuid.UUID, 0, len(todo.Sessions()))
	todoIDs := make([]uuid.UUID, 0, len(todo.Sessions()))
	userIDs := make([]uuid.UUID, 0, len(todo.Sessions()))
	startTimes := make([]time.Time, 0, len(todo.Sessions()))
	endTimes := make([]time.Time, 0, len(todo.Sessions()))

	for _, s := range todo.Sessions() {
		sessionIDs = append(sessionIDs, s.ID().UUID())
		todoIDs = append(todoIDs, todo.ID().UUID())
		userIDs = append(userIDs, s.UserID().UUID())

		startTimes = append(startTimes, s.StartTime())
		if s.EndTime() != nil {
			endTimes = append(endTimes, *s.EndTime())
		} else {
			endTimes = append(endTimes, time.Time{})
		}
	}

	err = r.q.RemoveMissingFocusSessionsFromTodo(ctx, dbtx, db.RemoveMissingFocusSessionsFromTodoParams{
		TodoID:     todo.ID().UUID(),
		SessionIds: sessionIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to sync focus sessions for todo %s: %w", todo.ID(), sharedPg.ParseDBError(err))
	}

	if len(sessionIDs) > 0 {
		err = r.q.BulkUpsertFocusSessions(ctx, dbtx, db.BulkUpsertFocusSessionsParams{
			Ids:        sessionIDs,
			TodoIds:    todoIDs,
			UserIds:    userIDs,
			StartTimes: startTimes,
			EndTimes:   endTimes,
		})
		if err != nil {
			return fmt.Errorf("failed to bulk upsert focus sessions for todo %s: %w", todo.ID(), sharedPg.ParseDBError(err))
		}
	}

	r.uow.Collect(ctx, r.mapper, todo)

	return nil
}

func (r *TodoRepo) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
	row, err := r.q.GetTodoAggregateByID(ctx, r.getDB(ctx), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}

		return nil, fmt.Errorf("failed to get todo %s: %w", id, sharedPg.ParseDBError(err))
	}

	return r.mapper.ToDomain(row), nil
}

func (r *TodoRepo) FindAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]*domain.Todo, error) {
	rows, err := r.q.ListTodosByWorkspaceID(ctx, r.getDB(ctx), db.ListTodosByWorkspaceIDParams{
		WorkspaceID: wsID,
		Limit:       math.MaxInt32,
		Offset:      0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list todos for workspace %s: %w", wsID, sharedPg.ParseDBError(err))
	}

	todos := make([]*domain.Todo, len(rows))
	for i, row := range rows {
		todos[i] = r.mapper.ListRowToDomain(row)
	}

	return todos, nil
}

func (r *TodoRepo) Delete(ctx context.Context, id domain.TodoID) error {
	dbtx := r.getDB(ctx)

	t, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.q.DeleteTodo(ctx, dbtx, id); err != nil {
		return fmt.Errorf("failed to delete todo %s: %w", id, sharedPg.ParseDBError(err))
	}

	t.Delete()

	r.uow.Collect(ctx, r.mapper, t)

	return nil
}
