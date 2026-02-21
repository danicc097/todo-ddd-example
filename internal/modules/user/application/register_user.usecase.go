package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type RegisterUserCommand struct {
	Email string
	Name  string
}

type RegisterUserResponse struct {
	ID domain.UserID
}

type RegisterUserUseCase struct {
	repo domain.UserRepository
}

func NewRegisterUserUseCase(repo domain.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, cmd RegisterUserCommand) (RegisterUserResponse, error) {
	email, err := domain.NewUserEmail(cmd.Email)
	if err != nil {
		return RegisterUserResponse{}, err
	}

	name, err := domain.NewUserName(cmd.Name)
	if err != nil {
		return RegisterUserResponse{}, err
	}

	user := domain.CreateUser(email, name)
	if err := uc.repo.Save(ctx, user); err != nil {
		return RegisterUserResponse{}, err
	}

	return RegisterUserResponse{ID: user.ID()}, nil
}
