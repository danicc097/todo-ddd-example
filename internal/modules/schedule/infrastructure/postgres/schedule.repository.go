package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type ScheduleRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *ScheduleMapper
	uow    application.UnitOfWork
}

func NewScheduleRepo(pool *pgxpool.Pool, uow application.UnitOfWork) *ScheduleRepo {
	return &ScheduleRepo{
		q:      db.New(),
		pool:   pool,
		mapper: &ScheduleMapper{},
		uow:    uow,
	}
}

func (r *ScheduleRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *ScheduleRepo) Save(ctx context.Context, sched *domain.DailySchedule) error {
	dbtx := r.getDB(ctx)
	p := r.mapper.ToPersistence(sched)

	_, err := r.q.UpsertDailySchedule(ctx, dbtx, db.UpsertDailyScheduleParams{
		UserID:         p.UserID,
		Date:           p.Date,
		MaxCapacity:    p.MaxCapacity,
		CurrentVersion: int32(sched.Version()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrConcurrentUpdate
		}

		return fmt.Errorf("failed to upsert schedule: %w", sharedPg.ParseDBError(err))
	}

	todoIDs := make([]uuid.UUID, 0, len(sched.CommittedTasks()))
	userIDs := make([]uuid.UUID, 0, len(sched.CommittedTasks()))
	dates := make([]time.Time, 0, len(sched.CommittedTasks()))
	costs := make([]int32, 0, len(sched.CommittedTasks()))

	for id, cost := range sched.CommittedTasks() {
		todoIDs = append(todoIDs, id.UUID())
		userIDs = append(userIDs, p.UserID)
		dates = append(dates, p.Date)
		costs = append(costs, int32(cost))
	}

	err = r.q.RemoveMissingTasksFromSchedule(ctx, dbtx, db.RemoveMissingTasksFromScheduleParams{
		UserID:  p.UserID,
		Date:    p.Date,
		TodoIds: todoIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to sync schedule tasks: %w", sharedPg.ParseDBError(err))
	}

	if len(todoIDs) > 0 {
		err = r.q.BulkUpsertScheduleTasks(ctx, dbtx, db.BulkUpsertScheduleTasksParams{
			UserIds:     userIDs,
			Dates:       dates,
			TodoIds:     todoIDs,
			EnergyCosts: costs,
		})
		if err != nil {
			return fmt.Errorf("failed to bulk upsert schedule tasks: %w", sharedPg.ParseDBError(err))
		}
	}

	r.uow.Collect(ctx, r.mapper, sched)

	return nil
}

func (r *ScheduleRepo) FindByUserAndDate(ctx context.Context, userID userDomain.UserID, date domain.ScheduleDate) (*domain.DailySchedule, error) {
	dbtx := r.getDB(ctx)

	t, err := time.Parse(time.DateOnly, string(date))
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	row, err := r.q.GetDailySchedule(ctx, dbtx, db.GetDailyScheduleParams{
		UserID: userID.UUID(),
		Date:   t,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleNotFound
		}

		return nil, fmt.Errorf("failed to get schedule: %w", sharedPg.ParseDBError(err))
	}

	tasks, err := r.q.GetScheduleTasks(ctx, dbtx, db.GetScheduleTasksParams{
		UserID: userID.UUID(),
		Date:   t,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule tasks: %w", sharedPg.ParseDBError(err))
	}

	return r.mapper.ToDomain(row, tasks), nil
}

func (r *ScheduleRepo) FindSchedulesByTodoID(ctx context.Context, todoID todoDomain.TodoID) ([]*domain.DailySchedule, error) {
	dbtx := r.getDB(ctx)

	rows, err := r.q.GetSchedulesByTodoID(ctx, dbtx, todoID.UUID())
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules for todo %s: %w", todoID, sharedPg.ParseDBError(err))
	}

	schedules := make([]*domain.DailySchedule, 0, len(rows))
	for _, row := range rows {
		dateStr := row.Date.Format(time.DateOnly)

		s, err := r.FindByUserAndDate(ctx, userDomain.UserID(row.UserID), domain.ScheduleDate(dateStr))
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, s)
	}

	return schedules, nil
}
