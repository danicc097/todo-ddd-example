package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type GetUserUseCase struct {
	repo domain.UserRepository
}

func NewGetUserUseCase(repo domain.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{repo: repo}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, id domain.UserID) (UserReadModel, error) {
	u, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return UserReadModel{}, err
	}

	return UserReadModel{
		ID:    u.ID(),
		Email: u.Email().String(),
		Name:  u.Name().String(),
	}, nil
}
