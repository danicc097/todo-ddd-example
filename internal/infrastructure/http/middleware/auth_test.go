package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	infraHttp "github.com/danicc097/todo-ddd-example/internal/infrastructure/http"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/utils/crypto"
)

func TestIdentityAndMFAResolver_Integration(t *testing.T) {
	t.Parallel()

	privKeyBytes, _ := os.ReadFile("../../../../private.pem")
	privKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privKeyBytes)
	pubKeyBytes, _ := os.ReadFile("../../../../public.pem")
	pubKey, _ := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)

	tokenIssuer := crypto.NewTokenIssuer(privKey, "test")
	tokenVerifier := crypto.NewTokenVerifier(pubKey)

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.IdentityAndMFAResolver(tokenVerifier))
	r.GET("/test", func(c *gin.Context) {
		meta := causation.FromContext(c.Request.Context())
		c.JSON(http.StatusOK, meta)
	})

	t.Run("valid JWT", func(t *testing.T) {
		uid := uuid.New()
		token, err := tokenIssuer.Issue(uid, true, time.Hour)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(infraHttp.AuthorizationHeader, fmt.Sprintf("%s %s", infraHttp.BearerScheme, token))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"MFAVerified":true`)
		assert.Contains(t, w.Body.String(), uid.String())
	})

	t.Run("invalid JWT", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(infraHttp.AuthorizationHeader, fmt.Sprintf("%s %s", infraHttp.BearerScheme, "invalid-token"))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"UserID":"00000000-0000-0000-0000-000000000000"`)
	})
}
