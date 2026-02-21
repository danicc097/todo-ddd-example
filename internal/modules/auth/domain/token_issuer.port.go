package domain

import (
	"time"

	"github.com/google/uuid"
)

// TokenIssuer defineshow authentication tokens are issued.
type TokenIssuer interface {
	Issue(userID uuid.UUID, mfaVerified bool, duration time.Duration) (string, error)
}
