package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

// IdentityAndMFAResolver verifies the RS256 JWT and establishes identity.
func IdentityAndMFAResolver(verifier *crypto.TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		claims, err := verifier.Verify(tokenString)
		if err != nil {
			c.Next()
			return
		}

		meta := causation.Metadata{
			UserID:        claims.UserID,
			UserIP:        c.ClientIP(),
			CorrelationID: uuid.NewString(),
			MFAVerified:   claims.MFAVerified,
			UserAgent:     c.Request.UserAgent(),
		}

		ctx := causation.WithMetadata(c.Request.Context(), meta)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
