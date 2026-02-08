package domain

import (
	"context"
	"github.com/google/uuid"
)

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}
