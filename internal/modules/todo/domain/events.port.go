package domain

import "context"

// EventPublisher is an output port
type EventPublisher interface {
	PublishTodoUpdated(ctx context.Context, todo *Todo) error
}
