package application

import (
	"context"

	"github.com/negrel/secrecy"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type RegisterCommand struct {
	Email    string
	Name     string
	Password secrecy.Secret[string]
}

type RegisterUserResponse struct {
	ID userDomain.UserID
}

type RegisterHandler struct {
	userRepo userDomain.UserRepository
	authRepo domain.AuthRepository
	hasher   domain.PasswordHasher
	uow      sharedApp.UnitOfWork
}

func NewRegisterHandler(
	userRepo userDomain.UserRepository,
	authRepo domain.AuthRepository,
	hasher domain.PasswordHasher,
	uow sharedApp.UnitOfWork,
) *RegisterHandler {
	return &RegisterHandler{
		userRepo: userRepo,
		authRepo: authRepo,
		hasher:   hasher,
		uow:      uow,
	}
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCommand) (RegisterUserResponse, error) {
	email, err := userDomain.NewUserEmail(cmd.Email)
	if err != nil {
		return RegisterUserResponse{}, err
	}

	name, err := userDomain.NewUserName(cmd.Name)
	if err != nil {
		return RegisterUserResponse{}, err
	}

	// non-owasp: should also check passwords against a compromised list and password strength.
	hash, err := h.hasher.Hash(cmd.Password.ExposeSecret())
	if err != nil {
		return RegisterUserResponse{}, err
	}

	user := userDomain.NewUser(email, name)
	auth := domain.NewUserAuth(user.ID(), hash)

	err = h.uow.Execute(ctx, func(ctx context.Context) error {
		if err := h.userRepo.Save(ctx, user); err != nil {
			return err
		}

		return h.authRepo.Save(ctx, auth)
	})
	if err != nil {
		return RegisterUserResponse{}, err
	}

	return RegisterUserResponse{ID: user.ID()}, nil
}
