package domain

import "context"

// EventPublisher is an output port
//
//counterfeiter:generate . EventPublisher
type EventPublisher interface {
	PublishTodoUpdated(ctx context.Context, todo *Todo) error
}
