package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

// pgUnitOfWork implements the UnitOfWork interface.
type pgUnitOfWork struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

var _ application.UnitOfWork = (*pgUnitOfWork)(nil)

func NewUnitOfWork(pool *pgxpool.Pool) application.UnitOfWork {
	return &pgUnitOfWork{
		pool: pool,
		q:    db.New(),
	}
}

type uowKey struct{}

type collectedAggregate struct {
	mapper domain.EventMapper
	agg    domain.EventsAggregate
}

type uowState struct {
	aggregates []collectedAggregate
}

func (u *pgUnitOfWork) Collect(ctx context.Context, mapper domain.EventMapper, agg domain.EventsAggregate) {
	if state, ok := ctx.Value(uowKey{}).(*uowState); ok {
		state.aggregates = append(state.aggregates, collectedAggregate{mapper: mapper, agg: agg})
	}
}

func (u *pgUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	state := &uowState{}
	uowCtx := context.WithValue(ctx, uowKey{}, state)

	_, err := infraDB.RunInTx(uowCtx, u.pool, func(txCtx context.Context) (any, error) {
		if err := fn(txCtx); err != nil {
			return nil, err
		}

		dbtx := infraDB.ExtractTx(txCtx)
		if dbtx == nil {
			return nil, errors.New("transaction not found in context")
		}

		for _, item := range state.aggregates {
			if item.mapper == nil {
				continue
			}

			if err := SaveDomainEvents(txCtx, u.q, dbtx, item.mapper, item.agg); err != nil {
				return nil, fmt.Errorf("failed to save domain events: %w", err)
			}
		}

		//nolint: nilnil
		return nil, nil
	})

	return err
}
