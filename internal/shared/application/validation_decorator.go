package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	"github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type validationDecorator[C any, R any] struct {
	base RequestHandler[C, R]
}

// WithValidation wraps a RequestHandler with validation logic.
func WithValidation[C any, R any](base RequestHandler[C, R]) RequestHandler[C, R] {
	return &validationDecorator[C, R]{
		base: base,
	}
}

func (d *validationDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	var zero R

	validatable, ok := any(cmd).(domain.Validatable)
	if !ok {
		// if the Validate method has a pointer receiver.
		if ptr, ok := any(&cmd).(domain.Validatable); ok {
			validatable = ptr
		}
	}

	if validatable != nil {
		if err := validatable.Validate(); err != nil {
			return zero, apperrors.Wrap(err, apperrors.InvalidInput, err.Error())
		}
	}

	return d.base.Handle(ctx, cmd)
}
