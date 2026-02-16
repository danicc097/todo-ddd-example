package mfa

import (
	"context"
	"log/slog"

	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
)

type MockSMSService struct{}

func NewMockSMSService() *MockSMSService {
	return &MockSMSService{}
}

func (s *MockSMSService) Initiate(ctx context.Context, auth *domain.UserAuth) (domain.MFAChallenge, error) {
	slog.InfoContext(ctx, "MOCK_SMS_SENT", slog.String("user_id", auth.UserID().String()), slog.String("code", "123456"))
	return domain.MFAChallenge{Sent: true}, nil
}

func (s *MockSMSService) Verify(ctx context.Context, auth *domain.UserAuth, code string) error {
	if code == "123456" {
		return nil
	}
	return domain.ErrInvalidOTP
}
