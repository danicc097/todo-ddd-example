package application

import (
	"context"
	"time"

	"github.com/negrel/secrecy"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
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
	issuer   domain.TokenIssuer
	hasher   domain.PasswordHasher
}

func NewLoginHandler(userRepo userDomain.UserRepository, authRepo domain.AuthRepository, issuer domain.TokenIssuer, hasher domain.PasswordHasher) *LoginHandler {
	return &LoginHandler{
		userRepo: userRepo,
		authRepo: authRepo,
		issuer:   issuer,
		hasher:   hasher,
	}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (LoginResponse, error) {
	email, err := userDomain.NewUserEmail(cmd.Email)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	// non-owasp: prioritize availability over timing attack protection
	user, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	// non-owasp: we rate limit by ip address. should implement timed lockout or progressive delays
	auth, err := h.authRepo.FindByUserID(ctx, user.ID())
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	match, err := h.hasher.Compare(cmd.Password.ExposeSecret(), auth.PasswordHash())
	if err != nil || !match {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	token, err := h.issuer.Issue(user.ID().UUID(), false, 15*time.Minute)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{AccessToken: token}, nil
}
