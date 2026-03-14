package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	sharedHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type IdempotencyData struct {
	ResponseStatus  int               `json:"response_status"`
	ResponseHeaders map[string]string `json:"response_headers"`
	ResponseBody    []byte            `json:"response_body"`
}

const (
	ProcessingState = "PROCESSING"
	LockDuration    = 60 // seconds
	RetentionPeriod = 24 * time.Hour
)

var idempotencyScript = redis.NewScript(`
local key = KEYS[1]
local processing_val = ARGV[1]
local lock_ttl = ARGV[2]

local current_val = redis.call("GET", key)
if not current_val then
	redis.call("SET", key, processing_val, "EX", lock_ttl)
	return nil
end

return current_val
`)

// RedisIdempotency implements Redis-backed idempotency.
// NOTE: requires at least redis.conf appendonly yes and appendfsync everysec for almost same guarantees as db equivalent mw.
func RedisIdempotency(rdb redis.UniversalClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		keyStr := c.GetHeader(sharedHttp.IdempotencyKeyHeader)
		if keyStr == "" {
			c.Next()
			return
		}

		key, err := uuid.Parse(keyStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid Idempotency-Key format"})
			return
		}

		ctx := c.Request.Context()
		redisKey := cache.Keys.IdempotencyKey(key)

		val, err := idempotencyScript.Run(ctx, rdb, []string{redisKey}, ProcessingState, LockDuration).Result()

		if err != nil && !errors.Is(err, redis.Nil) {
			c.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		if !errors.Is(err, redis.Nil) {
			existingState, _ := val.(string)

			if existingState == ProcessingState {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Request is currently being processed"})
				return
			}

			var existing IdempotencyData
			if err := json.Unmarshal([]byte(existingState), &existing); err == nil {
				for k, v := range existing.ResponseHeaders {
					c.Writer.Header().Set(k, v)
				}

				c.Data(existing.ResponseStatus, c.Writer.Header().Get("Content-Type"), existing.ResponseBody)
				c.Abort()

				return
			}
		}

		w := &responseWrapper{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		c.Next()

		if w.Status() >= 200 && w.Status() < 300 {
			headers := make(map[string]string)
			for k, v := range w.Header() {
				headers[k] = v[0]
			}

			data := IdempotencyData{
				ResponseStatus:  w.Status(),
				ResponseHeaders: headers,
				ResponseBody:    w.body.Bytes(),
			}

			dataJSON, _ := json.Marshal(data)

			rdb.Set(context.Background(), redisKey, dataJSON, RetentionPeriod)
		} else {
			rdb.Del(context.Background(), redisKey)
		}
	}
}
