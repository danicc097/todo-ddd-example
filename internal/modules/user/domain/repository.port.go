package domain

import (
	"context"
)

//go:generate go tool gowrap gen -g -i UserRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/user_repository_tracing.gen.go
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email UserEmail) (*User, error)
}
