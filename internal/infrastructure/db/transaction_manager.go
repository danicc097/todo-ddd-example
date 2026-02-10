package db

import (
	"context"
	"errors"
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
	maxRetries := 3

	var err error

	for i := range maxRetries {
		err = tm.execOnce(ctx, fn)
		if err == nil {
			return nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.SerializationFailure {
			time.Sleep((time.Duration(i + 1)) * 10 * time.Millisecond)
			continue
		}

		break
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
