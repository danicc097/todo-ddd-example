package db

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	todoPg "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager interface {
	Exec(ctx context.Context, fn func(todoRepo domain.TodoRepository) error) error
}

type pgxTransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) TransactionManager {
	return &pgxTransactionManager{pool: pool}
}

func (tm *pgxTransactionManager) Exec(ctx context.Context, fn func(todoRepo domain.TodoRepository) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	repo := todoPg.NewTodoRepoFromTx(tx)

	if err := fn(repo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
