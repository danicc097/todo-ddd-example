package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
	sharedHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestIdempotency(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	redisClient := testutils.GetGlobalRedis(t).Connect(context.Background(), t)
	pool := testutils.GetGlobalPostgresPool(t)

	testCases := []struct {
		name       string
		middleware gin.HandlerFunc
	}{
		{
			name:       "redis",
			middleware: middleware.RedisIdempotency(redisClient),
		},
		{
			name:       "db",
			middleware: middleware.DBIdempotency(pool),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"/concurrent results in single execution", func(t *testing.T) {
			r := gin.New()
			r.Use(tc.middleware)

			var executionCount int32

			enteredHandler := make(chan struct{})
			releaseHandler := make(chan struct{})

			var once sync.Once

			r.POST("/concurrent", func(c *gin.Context) {
				atomic.AddInt32(&executionCount, 1)

				once.Do(func() { close(enteredHandler) })

				<-releaseHandler // block indefinitely in PROCESSING state

				c.JSON(http.StatusOK, gin.H{"status": "processed"})
			})

			idemKey := uuid.New().String()
			concurrentRequests := 10

			results := make(chan int, concurrentRequests)

			for range concurrentRequests {
				go func() {
					req, _ := http.NewRequest(http.MethodPost, "/concurrent", nil)
					req.Header.Set(sharedHttp.IdempotencyKeyHeader, idemKey)

					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)

					results <- w.Code
				}()
			}

			<-enteredHandler // block until we are certain 1st req acquired the lock

			var status409 int

			for range concurrentRequests - 1 {
				if c := <-results; c == http.StatusConflict {
					status409++
				}
			}

			close(releaseHandler) // releases the first request. the rest have already exited early

			assert.Equal(t, http.StatusOK, <-results)
			assert.Equal(t, 9, status409)
			assert.EqualValues(t, 1, atomic.LoadInt32(&executionCount)) // rest should skip handler
		})

		t.Run(tc.name+"/sequential execution returns cached result", func(t *testing.T) {
			r := gin.New()
			r.Use(tc.middleware)

			var executionCount int32

			r.POST("/sequential", func(c *gin.Context) {
				atomic.AddInt32(&executionCount, 1)
				c.Header("X-Custom-Header", "test-value")
				c.JSON(http.StatusCreated, gin.H{"message": "created"})
			})

			idemKey := uuid.New().String()

			for range 3 {
				req, _ := http.NewRequest(http.MethodPost, "/sequential", nil)
				req.Header.Set(sharedHttp.IdempotencyKeyHeader, idemKey)

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusCreated, w.Code)
				assert.Equal(t, "test-value", w.Header().Get("X-Custom-Header"))

				var res map[string]string

				err := json.Unmarshal(w.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, "created", res["message"])
			}

			assert.EqualValues(t, 1, atomic.LoadInt32(&executionCount)) // rest should skip handler
		})

		t.Run(tc.name+"/invalid key format", func(t *testing.T) {
			r := gin.New()
			r.Use(tc.middleware)

			r.POST("/bad-key", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodPost, "/bad-key", nil)
			req.Header.Set(sharedHttp.IdempotencyKeyHeader, "not-a-uuid")

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
