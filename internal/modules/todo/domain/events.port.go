package domain

import "context"

// EventPublisher is an output port
//
//counterfeiter:generate . EventPublisher
type EventPublisher interface {
	PublishTodoCreated(ctx context.Context, todo *Todo) error
	PublishTodoUpdated(ctx context.Context, todo *Todo) error
}
