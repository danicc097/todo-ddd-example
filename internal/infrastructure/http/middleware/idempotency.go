package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	idempotencyHeaderKey        = "Idempotency-Key"
	idempotencyTTL              = 1 * time.Hour
	idempotencyStatusProcessing = "PROCESSING"
)

type responseWrapper struct {
	gin.ResponseWriter

	body *bytes.Buffer
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type storedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// Idempotency implements the cache aside pattern by deduplicating requests.
// This ensures that retrying a non-idempotent operation does not result in duplicate resources.
func Idempotency(client *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		key := c.GetHeader(idempotencyHeaderKey)
		if key == "" {
			c.Next()
			return
		}

		redisKey := newIdempotencyRedisKey(key)
		ctx := c.Request.Context()

		acquired, err := client.SetNX(ctx, redisKey, idempotencyStatusProcessing, 10*time.Second).Result()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !acquired { // exists
			val, err := client.Get(ctx, redisKey).Bytes()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			if string(val) == idempotencyStatusProcessing { // concurrent
				c.JSON(http.StatusConflict, gin.H{"error": "Request is currently being processed"})
				c.Abort()

				return
			}

			var resp storedResponse
			if json.Unmarshal(val, &resp) == nil {
				for k, v := range resp.Headers {
					c.Writer.Header().Set(k, v)
				}

				c.Data(resp.Status, c.Writer.Header().Get("Content-Type"), resp.Body)
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

		if w.Status() >= 400 {
			client.Del(ctx, redisKey)
			return
		}

		headers := make(map[string]string)
		for k, v := range w.Header() {
			headers[k] = v[0]
		}

		resp := storedResponse{
			Status:  w.Status(),
			Headers: headers,
			Body:    w.body.Bytes(),
		}

		data, _ := json.Marshal(resp)
		client.Set(ctx, redisKey, data, idempotencyTTL)
	}
}

func newIdempotencyRedisKey(key string) string {
	return "idempotency:" + key
}
