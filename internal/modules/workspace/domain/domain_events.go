package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceCreatedEvent struct {
	ID       uuid.UUID
	Name     string
	Occurred time.Time
}

func (e WorkspaceCreatedEvent) EventName() string      { return "workspace.created" }
func (e WorkspaceCreatedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e WorkspaceCreatedEvent) AggregateID() uuid.UUID { return e.ID }
