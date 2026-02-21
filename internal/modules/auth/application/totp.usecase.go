package application

import (
	"context"
	"errors"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

type TOTPGuard interface {
	Consume(ctx context.Context, userID userDomain.UserID, code string) error
}

type InitiateTOTPHandler struct {
	repo      domain.AuthRepository
	masterKey []byte
}

func NewInitiateTOTPHandler(repo domain.AuthRepository, masterKey []byte) *InitiateTOTPHandler {
	return &InitiateTOTPHandler{repo: repo, masterKey: masterKey}
}

func (h *InitiateTOTPHandler) Handle(ctx context.Context, _ application.Void) (string, error) {
	meta := causation.FromContext(ctx)
	userID := userDomain.UserID(meta.UserID)

	auth, err := h.repo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Todo-DDD-App", AccountName: userID.String()})
	if err != nil {
		return "", err
	}

	cipher, nonce, err := crypto.Encrypt([]byte(key.Secret()), h.masterKey)
	if err != nil {
		return "", err
	}

	auth.InitiateTOTP(cipher, nonce)

	if err := h.repo.Save(ctx, auth); err != nil {
		return "", err
	}

	return key.URL(), nil
}

type VerifyTOTPCommand struct{ Code string }

type VerifyTOTPResponse struct {
	AccessToken string
}

type VerifyTOTPHandler struct {
	repo      domain.AuthRepository
	guard     TOTPGuard
	issuer    domain.TokenIssuer
	masterKey []byte
}

func NewVerifyTOTPHandler(repo domain.AuthRepository, guard TOTPGuard, issuer domain.TokenIssuer, masterKey []byte) *VerifyTOTPHandler {
	return &VerifyTOTPHandler{repo: repo, guard: guard, issuer: issuer, masterKey: masterKey}
}

func (h *VerifyTOTPHandler) Handle(ctx context.Context, cmd VerifyTOTPCommand) (VerifyTOTPResponse, error) {
	meta := causation.FromContext(ctx)
	userID := userDomain.UserID(meta.UserID)

	auth, err := h.repo.FindByUserID(ctx, userID)
	if err != nil {
		return VerifyTOTPResponse{}, err
	}

	if !auth.IsTOTPPending() && !auth.IsTOTPActive() {
		return VerifyTOTPResponse{}, errors.New("TOTP not set up")
	}

	cipher, nonce := auth.TOTPCredentials()

	secret, err := crypto.Decrypt(cipher, nonce, h.masterKey)
	if err != nil {
		return VerifyTOTPResponse{}, err
	}

	if !totp.Validate(cmd.Code, string(secret)) {
		return VerifyTOTPResponse{}, errors.New("invalid code")
	}

	if err := h.guard.Consume(ctx, userID, cmd.Code); err != nil {
		return VerifyTOTPResponse{}, err
	}

	if auth.IsTOTPPending() {
		auth.ActivateTOTP()

		if err := h.repo.Save(ctx, auth); err != nil {
			return VerifyTOTPResponse{}, err
		}
	}

	token, err := h.issuer.Issue(userID.UUID(), true, 15*time.Minute)
	if err != nil {
		return VerifyTOTPResponse{}, err
	}

	return VerifyTOTPResponse{AccessToken: token}, nil
}
