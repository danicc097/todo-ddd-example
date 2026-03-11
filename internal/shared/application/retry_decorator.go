package application

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"
)

// WithRetry is a generic retry decorator.
func WithRetry[C any, R any](next RequestHandler[C, R], maxRetries int) RequestHandler[C, R] {
	return WithRetryOn(next, maxRetries, nil) // retry on any error
}

// WithRetryOn wraps a RequestHandler with retry logic for a specific error.
func WithRetryOn[C any, R any](next RequestHandler[C, R], maxRetries int, targetErr error) RequestHandler[C, R] {
	return &retryDecorator[C, R]{
		next:        next,
		maxRetries:  maxRetries,
		targetError: targetErr,
	}
}

type retryDecorator[C any, R any] struct {
	next        RequestHandler[C, R]
	maxRetries  int
	targetError error
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

		res, err := h.next.Handle(ctx, cmd)
		if err == nil {
			return res, nil
		}

		if h.targetError != nil && !errors.Is(err, h.targetError) {
			return zero, err
		}

		if i < h.maxRetries {
			// exponential backoff with jitter
			backoff := time.Duration(1<<i)*10*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond
			slog.WarnContext(ctx, "use case failed, retrying...",
				slog.Int("attempt", i+1),
				slog.Duration("backoff", backoff),
				slog.String("error", err.Error()),
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
