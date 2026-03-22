package postgres

import "context"

type EventMapper interface {
	MapEvent() error
}

type EventsAggregate interface {
	Events() []any
}

type UnitOfWork interface {
	Collect(ctx context.Context, mapper EventMapper, agg EventsAggregate)
}

type Repo struct {
	uow UnitOfWork
}

type MyAggregate struct{}

func (a *MyAggregate) Events() []any { return nil }

type MyMapper struct{}

func (m *MyMapper) MapEvent() error { return nil }

func (r *Repo) Save(ctx context.Context) {
	var agg MyAggregate
	r.uow.Collect(ctx, nil, &agg) // want "Arch violation: uow.Collect\\(\\) called with nil EventMapper. Domain events will be silently lost. Provide a concrete EventMapper implementation."
}

func (r *Repo) SaveCorrectly(ctx context.Context) {
	var agg MyAggregate
	r.uow.Collect(ctx, &MyMapper{}, &agg) // OK - non-nil mapper
}
