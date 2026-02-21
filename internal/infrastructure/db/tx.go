package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type txState struct {
	tx               pgx.Tx
	afterCommitHooks []func(ctx context.Context)
}

// InjectTx puts a transaction into the context.
func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, &txState{tx: tx})
}

// ExtractTx retrieves a transaction from the context, if any.
func ExtractTx(ctx context.Context) pgx.Tx {
	if state, ok := ctx.Value(txKey{}).(*txState); ok {
		return state.tx
	}

	return nil
}

// AfterCommit registers a hook to be executed after the transaction successfully commits.
func AfterCommit(ctx context.Context, fn func(ctx context.Context)) {
	if state, ok := ctx.Value(txKey{}).(*txState); ok {
		state.afterCommitHooks = append(state.afterCommitHooks, fn)
	} else {
		fn(ctx)
	}
}

// RunInTx executes a function within a database transaction.
func RunInTx[T any](ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) (T, error)) (result T, err error) {
	// nested transaction support could go here
	if ExtractTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to begin transaction: %w", err)
	}

	state := &txState{tx: tx}
	txCtx := context.WithValue(ctx, txKey{}, state)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)

			panic(p) // rethrow
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			} else {
				for _, hook := range state.afterCommitHooks {
					hook(ctx)
				}
			}
		}
	}()

	return fn(txCtx)
}
