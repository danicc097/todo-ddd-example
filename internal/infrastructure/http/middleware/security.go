package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal"
)

func SecurityHeaders(env internal.AppEnv) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")

		isDocs := c.Request.URL.Path == "/api/v1/docs"
		isProd := env == internal.AppEnvProd

		if isProd || !isDocs {
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}

		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")

		c.Next()
	}
}
