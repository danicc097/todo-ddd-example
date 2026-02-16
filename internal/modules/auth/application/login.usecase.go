package application

import (
	"context"
	"time"

	"github.com/negrel/secrecy"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

type LoginCommand struct {
	Email    string
	Password secrecy.Secret[string]
}

type LoginResponse struct {
	AccessToken string
}

type LoginHandler struct {
	userRepo userDomain.UserRepository
	authRepo domain.AuthRepository
	issuer   *crypto.TokenIssuer
}

func NewLoginHandler(userRepo userDomain.UserRepository, authRepo domain.AuthRepository, issuer *crypto.TokenIssuer) *LoginHandler {
	return &LoginHandler{
		userRepo: userRepo,
		authRepo: authRepo,
		issuer:   issuer,
	}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (LoginResponse, error) {
	email, err := userDomain.NewUserEmail(cmd.Email)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	user, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	auth, err := h.authRepo.FindByUserID(ctx, user.ID())
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	match, err := crypto.ComparePassword(cmd.Password.ExposeSecret(), auth.PasswordHash())
	if err != nil || !match {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	token, err := h.issuer.Issue(user.ID().UUID, false, 15*time.Minute)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{AccessToken: token}, nil
}
