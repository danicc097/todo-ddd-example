package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	sharedHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type responseWrapper struct {
	gin.ResponseWriter

	body *bytes.Buffer
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// DBIdempotency implements database-backed idempotency with atomic locking.
func DBIdempotency(pool *pgxpool.Pool) gin.HandlerFunc {
	q := db.New()

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

		affected, err := q.TryLockIdempotencyKey(ctx, pool, key)
		if err != nil {
			c.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		if affected == 0 {
			existing, err := q.GetIdempotencyKey(ctx, pool, key)
			if err != nil {
				c.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)

				return
			}

			if existing.ResponseStatus == 0 {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Request is currently being processed"})

				return
			}

			var headers map[string]string
			if json.Unmarshal(existing.ResponseHeaders, &headers) == nil {
				for k, v := range headers {
					c.Writer.Header().Set(k, v)
				}

				c.Data(int(existing.ResponseStatus), c.Writer.Header().Get("Content-Type"), existing.ResponseBody)
				c.Abort()

				return
			}
		}

		// lock acquired. handlers should manage their own internal txs
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

			headersJSON, _ := json.Marshal(headers)

			_ = q.UpdateIdempotencyKey(ctx, pool, db.UpdateIdempotencyKeyParams{
				ID:              key,
				ResponseStatus:  int32(w.Status()),
				ResponseHeaders: headersJSON,
				ResponseBody:    w.body.Bytes(),
			})
		} else {
			_ = q.DeleteIdempotencyKey(ctx, pool, key) // delete so we can retry
		}
	}
}
