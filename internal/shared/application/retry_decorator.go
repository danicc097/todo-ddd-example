package application

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type retryDecorator[C any, R any] struct {
	next       RequestHandler[C, R]
	maxRetries int
}

// Retry wraps a RequestHandler with retry logic for concurrent updates.
func Retry[C any, R any](next RequestHandler[C, R], maxRetries int) RequestHandler[C, R] {
	return &retryDecorator[C, R]{
		next:       next,
		maxRetries: maxRetries,
	}
}

func (h *retryDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	var (
		zero R
		err  error
	)

	for i := 0; i <= h.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		var res R

		res, err = h.next.Handle(ctx, cmd)
		if err == nil {
			return res, nil
		}

		if !errors.Is(err, domain.ErrConcurrentUpdate) {
			return zero, err
		}

		if i < h.maxRetries {
			// exponential backoff with jitter
			backoff := time.Duration(1<<i)*10*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond

			slog.WarnContext(ctx, "concurrent update detected, retrying...",
				slog.Int("attempt", i+1),
				slog.Duration("backoff", backoff),
			)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
	}

	return zero, err
}
