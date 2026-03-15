package domain

import "time"

type AggregateRoot struct {
	events []any
}

func (a *AggregateRoot) RecordEvent(e any) {
	a.events = append(a.events, e)
}

type TodoCreatedEvent struct {
	ID         string
	OccurredAt time.Time
}

type Todo struct {
	AggregateRoot

	id string
}

// CreateWithoutRecording creates an event variable but forgets to record it -- should be flagged.
func (t *Todo) CreateWithoutRecording() {
	evt := TodoCreatedEvent{ID: t.id} // want "Arch violation: Domain event variable evt is created but not passed to RecordEvent"
	_ = evt
}

func (t *Todo) CreateAndRecord() {
	evt := TodoCreatedEvent{ID: t.id}
	t.RecordEvent(evt)
}

func (t *Todo) CreateAndRecordDirect() {
	t.RecordEvent(TodoCreatedEvent{ID: t.id})
}
