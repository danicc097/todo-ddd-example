package application

import "github.com/danicc097/todo-ddd-example/internal/shared/domain"

type DecoratorBuilder[C any, R any] struct {
	handler RequestHandler[C, R]
}

func NewDecoratorBuilder[C any, R any](base RequestHandler[C, R]) *DecoratorBuilder[C, R] {
	return &DecoratorBuilder[C, R]{handler: base}
}

func (b *DecoratorBuilder[C, R]) WithLogging(name string) *DecoratorBuilder[C, R] {
	b.handler = WithLogging(b.handler, name)
	return b
}

func (b *DecoratorBuilder[C, R]) WithValidation() *DecoratorBuilder[C, R] {
	b.handler = WithValidation(b.handler)
	return b
}

func (b *DecoratorBuilder[C, R]) WithUoW(uow UnitOfWork) *DecoratorBuilder[C, R] {
	b.handler = WithUoW(b.handler, uow)
	return b
}

func (b *DecoratorBuilder[C, R]) WithRetry(maxRetries int) *DecoratorBuilder[C, R] {
	b.handler = WithRetry(b.handler, maxRetries)
	return b
}

// WithRetryOnConflict is a more specific retry for database concurrency errors.
func (b *DecoratorBuilder[C, R]) WithRetryOnConflict(maxRetries int) *DecoratorBuilder[C, R] {
	b.handler = WithRetryOn(b.handler, maxRetries, domain.ErrConcurrentUpdate)
	return b
}

func (b *DecoratorBuilder[C, R]) Build() RequestHandler[C, R] {
	return b.handler
}

func BuildCommand[C any, R any](h RequestHandler[C, R], uow UnitOfWork, name string) RequestHandler[C, R] {
	return NewCommandDecoratorBuilder(h, uow, name).Build()
}

func BuildQuery[C any, R any](h RequestHandler[C, R], name string) RequestHandler[C, R] {
	return NewQueryDecoratorBuilder(h, name).Build()
}

type Decorator[T any] func(T) T

// Apply is a generic helper to stack decorators.
func Apply[T any](base T, decorators ...Decorator[T]) T {
	for _, d := range decorators {
		base = d(base)
	}

	return base
}

// NewCommandDecoratorBuilder provides a convenient way to create a builder with default decorators for command handlers.
func NewCommandDecoratorBuilder[C any, R any](base RequestHandler[C, R], uow UnitOfWork, tracerName string) *DecoratorBuilder[C, R] {
	return NewDecoratorBuilder(base).
		WithLogging(tracerName).
		WithValidation().
		WithUoW(uow)
}

// NewQueryDecoratorBuilder provides a convenient way to create a builder with default decorators for query handlers.
func NewQueryDecoratorBuilder[C any, R any](base RequestHandler[C, R], tracerName string) *DecoratorBuilder[C, R] {
	return NewDecoratorBuilder(base).
		WithLogging(tracerName).
		WithValidation()
}
