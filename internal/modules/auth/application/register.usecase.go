package application

import (
	"context"

	"github.com/negrel/secrecy"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

type RegisterCommand struct {
	Email    string
	Name     string
	Password secrecy.Secret[string]
}

type RegisterHandler struct {
	userRepo userDomain.UserRepository
	authRepo domain.AuthRepository
}

func NewRegisterHandler(userRepo userDomain.UserRepository, authRepo domain.AuthRepository) *RegisterHandler {
	return &RegisterHandler{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCommand) (userDomain.UserID, error) {
	email, err := userDomain.NewUserEmail(cmd.Email)
	if err != nil {
		return userDomain.UserID{}, err
	}

	user := userDomain.CreateUser(email, cmd.Name)

	hash, err := crypto.HashPassword(cmd.Password.ExposeSecret(), crypto.DefaultArgon2Params)
	if err != nil {
		return userDomain.UserID{}, err
	}

	auth := domain.NewUserAuth(user.ID(), hash)

	if err := h.userRepo.Save(ctx, user); err != nil {
		return userDomain.UserID{}, err
	}

	if err := h.authRepo.Save(ctx, auth); err != nil {
		return userDomain.UserID{}, err
	}

	return user.ID(), nil
}
