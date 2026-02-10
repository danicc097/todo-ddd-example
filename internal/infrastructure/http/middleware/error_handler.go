package middleware

import (
	"errors"
	"net/http"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		var appErr *apperrors.AppError

		switch {
		case errors.Is(err, domain.ErrTodoNotFound), errors.Is(err, domain.ErrTagNotFound), errors.Is(err, userDomain.ErrUserNotFound):
			appErr = apperrors.New(apperrors.ErrCodeNotFound, "Resource not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidStatus):
			appErr = apperrors.New(apperrors.ErrCodeUnprocessable, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrTitleEmpty), errors.Is(err, domain.ErrTitleTooLong), errors.Is(err, userDomain.ErrInvalidEmail):
			appErr = apperrors.New(apperrors.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
		default:
			if asAppErr, ok := err.(*apperrors.AppError); ok {
				appErr = asAppErr
			} else {
				appErr = apperrors.New(apperrors.ErrCodeInternal, "Internal Server Error", http.StatusInternalServerError)
			}
		}

		span := trace.SpanFromContext(c.Request.Context())
		traceID := span.SpanContext().TraceID().String()

		c.JSON(appErr.Status, gin.H{
			"error": gin.H{
				"code":     appErr.Code,
				"message":  appErr.Message,
				"trace_id": traceID,
			},
		})
	}
}
