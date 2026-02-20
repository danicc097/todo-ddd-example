package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/getkin/kin-openapi/routers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
)

type rateLimitExt struct {
	Limit  int64  `json:"limit"`
	Window string `json:"window"`
}

func RateLimiter(rdb *redis.Client, router routers.Router) gin.HandlerFunc {
	return func(c *gin.Context) {
		route, _, err := router.FindRoute(c.Request)
		if err != nil {
			c.Next()
			return
		}

		extRaw, ok := route.Operation.Extensions["x-rate-limit"]
		if !ok {
			c.Next()
			return
		}

		extBytes, err := json.Marshal(extRaw)
		if err != nil {
			c.Next()
			return
		}

		var rlConfig rateLimitExt
		if err := json.Unmarshal(extBytes, &rlConfig); err != nil {
			c.Next()
			return
		}

		windowDuration, err := time.ParseDuration(rlConfig.Window)
		if err != nil {
			windowDuration = 60 * time.Second
		}

		ip := c.ClientIP()
		key := cache.Keys.RateLimit(route.Operation.OperationID, ip)
		now := time.Now().UnixMicro()
		windowStart := now - windowDuration.Microseconds()
		ctx := c.Request.Context()

		pipe := rdb.Pipeline()
		// removes all elements in the sorted set stored at key with a score between min and max (inclusive)
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
		countCmd := pipe.ZCard(ctx, key) // number of elements in the set

		if _, err := pipe.Exec(ctx); err != nil {
			c.Next()
			return
		}

		currentCount := countCmd.Val()
		remaining := rlConfig.Limit - currentCount

		c.Header("RateLimit-Limit", strconv.FormatInt(rlConfig.Limit, 10))
		c.Header("RateLimit-Remaining", strconv.FormatInt(max(0, remaining-1), 10))
		c.Header("RateLimit-Reset", strconv.FormatInt(int64(windowDuration.Seconds()), 10))

		if currentCount >= rlConfig.Limit {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()

			return
		}

		recordPipe := rdb.Pipeline()
		recordPipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
		recordPipe.Expire(ctx, key, windowDuration)
		_, _ = recordPipe.Exec(ctx)

		c.Next()
	}
}
