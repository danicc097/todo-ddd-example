package db

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
)

type RepositoryProvider interface {
	Todo() todoDomain.TodoRepository
	Tag() todoDomain.TagRepository
	User() userDomain.UserRepository
}

type pgxRepositoryProvider struct {
	tx pgx.Tx
}

func (p *pgxRepositoryProvider) Todo() todoDomain.TodoRepository {
	return todoPg.NewTodoRepoFromTx(p.tx)
}

func (p *pgxRepositoryProvider) Tag() todoDomain.TagRepository {
	return todoPg.NewTagRepoFromTx(p.tx)
}

func (p *pgxRepositoryProvider) User() userDomain.UserRepository {
	return userPg.NewUserRepoFromTx(p.tx)
}

type TransactionManager interface {
	Exec(ctx context.Context, fn func(p RepositoryProvider) error) error
}

type pgxTransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) TransactionManager {
	return &pgxTransactionManager{pool: pool}
}

func (tm *pgxTransactionManager) Exec(ctx context.Context, fn func(p RepositoryProvider) error) error {
	maxRetries := 5

	const (
		baseDelay = 10 * time.Millisecond
		maxDelay  = 200 * time.Millisecond
	)

	var err error

	for i := range maxRetries {
		err = tm.execOnce(ctx, fn)
		if err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == pgerrcode.SerializationFailure || pgErr.Code == pgerrcode.DeadlockDetected) {
			backoff := float64(baseDelay) * float64((int(1) << i))

			jitter := (rand.Float64() * 0.2) + 0.9
			sleepDuration := min(time.Duration(backoff*jitter), maxDelay)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleepDuration):
				continue
			}
		}

		return err
	}

	return err
}

func (tm *pgxTransactionManager) execOnce(ctx context.Context, fn func(p RepositoryProvider) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(&pgxRepositoryProvider{tx: tx}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
