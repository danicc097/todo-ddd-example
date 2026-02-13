package domain

import (
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type AggregateRoot struct {
	events []shared.DomainEvent
}

func (a *AggregateRoot) RecordEvent(e shared.DomainEvent) {
	a.events = append(a.events, e)
}

func (a *AggregateRoot) Events() []shared.DomainEvent {
	return a.events
}

func (a *AggregateRoot) ClearEvents() {
	a.events = nil
}
