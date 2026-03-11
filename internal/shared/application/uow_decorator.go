package application

import "context"

type uowDecorator[C any, R any] struct {
	base RequestHandler[C, R]
	uow  UnitOfWork
}

// WithUoW wraps a RequestHandler within a transactional Unit of Work.
func WithUoW[C any, R any](base RequestHandler[C, R], uow UnitOfWork) RequestHandler[C, R] {
	return &uowDecorator[C, R]{
		base: base,
		uow:  uow,
	}
}

func (d *uowDecorator[C, R]) Handle(ctx context.Context, cmd C) (R, error) {
	var res R

	err := d.uow.Execute(ctx, func(txCtx context.Context) error {
		var innerErr error

		res, innerErr = d.base.Handle(txCtx, cmd)

		return innerErr
	})

	return res, err
}
