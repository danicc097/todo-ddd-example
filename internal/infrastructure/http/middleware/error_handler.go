package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
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
		case errors.Is(err, todoDomain.ErrTodoNotFound), errors.Is(err, todoDomain.ErrTagNotFound), errors.Is(err, userDomain.ErrUserNotFound), errors.Is(err, wsDomain.ErrWorkspaceNotFound), errors.Is(err, wsDomain.ErrMemberNotFound):
			appErr = apperrors.New(apperrors.ErrCodeNotFound, "Resource not found", http.StatusNotFound)
		case errors.Is(err, todoDomain.ErrInvalidStatus):
			appErr = apperrors.New(apperrors.ErrCodeUnprocessable, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, todoDomain.ErrTitleEmpty), errors.Is(err, todoDomain.ErrTitleTooLong), errors.Is(err, userDomain.ErrInvalidEmail), errors.Is(err, wsDomain.ErrAtLeastOneOwner), errors.Is(err, wsDomain.ErrUserAlreadyMember):
			appErr = apperrors.New(apperrors.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest)
		default:
			asAppErr := &apperrors.AppError{}
			if errors.As(err, &asAppErr) {
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
