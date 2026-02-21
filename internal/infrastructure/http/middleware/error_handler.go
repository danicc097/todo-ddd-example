package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

func errorCodeToHTTP(code apperrors.ErrorCode) int {
	//exhaustive:enforce
	switch code {
	case apperrors.NotFound:
		return http.StatusNotFound
	case apperrors.InvalidInput:
		return http.StatusBadRequest
	case apperrors.Conflict:
		return http.StatusConflict
	case apperrors.Unprocessable:
		return http.StatusUnprocessableEntity
	case apperrors.Unauthorized:
		return http.StatusUnauthorized
	case apperrors.MFARequired:
		return http.StatusForbidden
	case apperrors.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperrors.AppError

		if errors.As(err, &appErr) {
			if !errors.Is(err, appErr) {
				// if we wrapped it (repos), override inner string to keep the repo's context
				// else just use apperror's message, likely from handler
				appErr.Message = err.Error()
			}
		} else {
			var domainErr sharedDomain.DomainError
			if errors.As(err, &domainErr) {
				appErr = apperrors.New(domainErr.Code(), err.Error())
			} else {
				// don't leak internal issues
				appErr = apperrors.New(apperrors.Internal, "Internal Server Error")
			}
		}

		httpStatus := errorCodeToHTTP(appErr.Code)
		span := trace.SpanFromContext(c.Request.Context())

		errPayload := gin.H{
			"code":     appErr.Code,
			"message":  appErr.Message,
			"trace_id": span.SpanContext().TraceID().String(),
		}

		if appErr.Validation != nil {
			errPayload["validation"] = appErr.Validation
		}

		c.JSON(httpStatus, gin.H{"error": errPayload})
	}
}
