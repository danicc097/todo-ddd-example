package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter

	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func ETag() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		writer := &bodyWriter{body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		hash := sha256.Sum256(writer.body.Bytes())
		etag := `"` + hex.EncodeToString(hash[:]) + `"`

		if c.Request.Header.Get("If-None-Match") == etag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		c.Header("ETag", etag)
		writer.ResponseWriter.Write(writer.body.Bytes())
	}
}
