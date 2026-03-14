package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	sharedHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestRateLimiter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rdb := testutils.GetGlobalRedis(t).Connect(ctx, t)

	opID := "limitedOp-" + uuid.New().String()

	specRaw := []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /limited:
    get:
      operationId: ` + opID + `
      x-rate-limit:
        limit: 2
        window: 1s
      responses:
        '200':
          description: OK
`)

	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(specRaw)
	require.NoError(t, err)

	spec.Servers = append(spec.Servers, &openapi3.Server{URL: "/"})

	router, err := gorillamux.NewRouter(spec)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.RateLimiter(rdb, router, internal.AppEnvDev))
	r.GET("/limited", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("Allow within limit", func(t *testing.T) {
		rdb.Del(ctx, "ratelimit:"+opID+":127.0.0.1")

		for range 2 {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/limited", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "2", w.Header().Get("Ratelimit-Limit"))
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/limited", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "Rate limit exceeded")
	})

	t.Run("Skip limit with header in non-prod", func(t *testing.T) {
		rdb.Del(ctx, "ratelimit:"+opID+":127.0.0.1")

		for range 2 {
			req, _ := http.NewRequest(http.MethodGet, "/limited", nil)
			r.ServeHTTP(httptest.NewRecorder(), req)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/limited", nil)
		req.Header.Set(sharedHttp.SkipRateLimitHeader, "1")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
