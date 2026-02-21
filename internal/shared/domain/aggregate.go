package domain

type AggregateRoot struct {
	events []DomainEvent
}

func (a *AggregateRoot) RecordEvent(e DomainEvent) {
	a.events = append(a.events, e)
}

func (a *AggregateRoot) Events() []DomainEvent {
	return a.events
}

func (a *AggregateRoot) ClearEvents() {
	a.events = nil
}
