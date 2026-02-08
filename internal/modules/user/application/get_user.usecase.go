package application

import (
	"context"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/google/uuid"
)

type GetUserUseCase struct {
	repo domain.UserRepository
}

func NewGetUserUseCase(repo domain.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{repo: repo}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.repo.FindByID(ctx, id)
}
