package application

import (
	"context"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/google/uuid"
)

type RegisterUserCommand struct {
	Email string
	Name  string
}

type RegisterUserUseCase struct {
	repo domain.UserRepository
}

func NewRegisterUserUseCase(repo domain.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, cmd RegisterUserCommand) (uuid.UUID, error) {
	email, err := domain.NewUserEmail(cmd.Email)
	if err != nil {
		return uuid.UUID{}, err
	}

	user := domain.CreateUser(email, cmd.Name)
	if err := uc.repo.Save(ctx, user); err != nil {
		return uuid.UUID{}, err
	}

	return user.ID(), nil
}
