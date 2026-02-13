package middleware

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type transactionalHandler[C any, R any] struct {
	pool *pgxpool.Pool
	next application.RequestHandler[C, R]
}

// Transactional wraps a RequestHandler with a database transaction.
func Transactional[C any, R any](
	pool *pgxpool.Pool,
	next application.RequestHandler[C, R],
) application.RequestHandler[C, R] {
	return &transactionalHandler[C, R]{
		pool: pool,
		next: next,
	}
}

func (h *transactionalHandler[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	return db.RunInTx(ctx, h.pool, func(ctx context.Context) (R, error) {
		return h.next.Handle(ctx, cmd)
	})
}
