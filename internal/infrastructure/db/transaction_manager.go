package db

import (
	"context"

	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	userPg "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//counterfeiter:generate . RepositoryProvider
type RepositoryProvider interface {
	Todo() todoDomain.TodoRepository
	User() userDomain.UserRepository
}

type pgxRepositoryProvider struct {
	tx pgx.Tx
}

func (p *pgxRepositoryProvider) Todo() todoDomain.TodoRepository {
	return todoPg.NewTodoRepoFromTx(p.tx)
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
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	provider := &pgxRepositoryProvider{tx: tx}

	if err := fn(provider); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
