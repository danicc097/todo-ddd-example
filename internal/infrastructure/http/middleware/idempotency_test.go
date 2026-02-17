package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedisContainer(t *testing.T) *redis.Client {
	ctx := context.Background()

	redisContainer, err := tcRedis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %s", err)
		}
	})

	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)
	opt, err := redis.ParseURL(uri)
	require.NoError(t, err)

	client := redis.NewClient(opt)
	require.NoError(t, client.Ping(ctx).Err())

	return client
}

func TestIdempotencyMiddleware(t *testing.T) {
	t.Parallel()

	rdb := setupRedisContainer(t)

	gin.SetMode(gin.TestMode)

	t.Run("First request processes normally and caches response", func(t *testing.T) {
		key := "req-1"

		r := gin.New()
		r.Use(Idempotency(rdb))
		r.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "123"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set(idempotencyHeaderKey, key)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t, `{"id": "123"}`, w.Body.String())

		ctx := context.Background()
		val, err := rdb.Get(ctx, newIdempotencyRedisKey(key)).Result()
		assert.NoError(t, err)
		assert.NotEmpty(t, val)
		assert.NotEqual(t, idempotencyStatusProcessing, val)
	})

	t.Run("Second request returns cached response", func(t *testing.T) {
		key := "req-2"

		r := gin.New()
		r.Use(Idempotency(rdb))
		r.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"data": "a"})
		})

		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req1.Header.Set(idempotencyHeaderKey, key)
		r.ServeHTTP(w1, req1)

		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req2.Header.Set(idempotencyHeaderKey, key)
		r.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusCreated, w2.Code)
	})

	t.Run("Concurrent returns conflict", func(t *testing.T) {
		key := "req-3"

		ctx := context.Background()
		err := rdb.Set(ctx, newIdempotencyRedisKey(key), idempotencyStatusProcessing, 10*time.Second).Err()
		require.NoError(t, err)

		r := gin.New()
		r.Use(Idempotency(rdb))
		r.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set(idempotencyHeaderKey, key)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), `currently being processed`)
	})

	t.Run("Request without header ignored", func(t *testing.T) {
		r := gin.New()
		r.Use(Idempotency(rdb))
		r.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Failed request is not cached", func(t *testing.T) {
		key := "req-5"

		r := gin.New()
		r.Use(Idempotency(rdb))
		r.POST("/test", func(c *gin.Context) {
			c.Status(http.StatusInternalServerError)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set(idempotencyHeaderKey, key)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		ctx := context.Background()
		_, err := rdb.Get(ctx, newIdempotencyRedisKey(key)).Result()
		assert.ErrorIs(t, err, redis.Nil) // deleted
	})
}
