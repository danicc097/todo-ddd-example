package http

import (
	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

// BindJSON attempts to bind the request body and automatically handles the error response.
func BindJSON[T any](c *gin.Context) (T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return req, false
	}

	return req, true
}

// Execute executes a Use Case command and automatically handles the error response.
func Execute[C any, R any](c *gin.Context, handler application.RequestHandler[C, R], cmd C) (R, bool) {
	resp, err := handler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.Error(err)
		return resp, false
	}

	return resp, true
}
