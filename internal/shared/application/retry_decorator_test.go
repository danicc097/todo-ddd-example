package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/application/mocks"
)

func TestRetryDecorator(t *testing.T) {
	t.Parallel()

	t.Run("Retries on error until success", func(t *testing.T) {
		h := new(mocks.FakeRequestHandler[string, string])

		errs := make(chan error, 1)
		errs <- errors.New("fail")

		close(errs)

		h.HandleCalls(func(context.Context, string) (string, error) {
			return "res", <-errs
		})

		res, err := application.WithRetry(h, 2).Handle(context.Background(), "cmd")

		require.NoError(t, err)
		assert.Equal(t, "res", res)
		assert.Equal(t, 2, h.HandleCallCount())
	})

	t.Run("Fails after max retries", func(t *testing.T) {
		expectedErr := errors.New("permanent fail")
		h := new(mocks.FakeRequestHandler[string, string])
		h.HandleReturns("res", expectedErr)

		const maxRetries = 3

		_, err := application.WithRetry(h, maxRetries).Handle(context.Background(), "cmd")

		require.ErrorIs(t, err, expectedErr)
		assert.Equal(t, maxRetries+1, h.HandleCallCount())
	})

	t.Run("Retries only on specific error", func(t *testing.T) {
		targetErr, otherErr := errors.New("retry me"), errors.New("don't retry me")
		h := new(mocks.FakeRequestHandler[string, string])
		h.HandleReturns("res", otherErr)

		_, err := application.WithRetryOn(h, 3, targetErr).Handle(context.Background(), "cmd")

		require.ErrorIs(t, err, otherErr)
		assert.Equal(t, 1, h.HandleCallCount())
	})
}
