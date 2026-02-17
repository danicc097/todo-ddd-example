package domain

import (
	"context"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

//go:generate go tool gowrap gen -g -i AuthRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/auth_repository_tracing.gen.go
type AuthRepository interface {
	FindByUserID(ctx context.Context, userID userDomain.UserID) (*UserAuth, error)
	Save(ctx context.Context, auth *UserAuth) error
}
