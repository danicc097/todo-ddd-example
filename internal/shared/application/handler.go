package application

import "context"

//go:generate go tool counterfeiter -o mocks/request_handler.gen.go . RequestHandler

// RequestHandler is the generic interface for all Use Cases.
type RequestHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// Void is a return type for commands that don't return data.
type Void struct{}
