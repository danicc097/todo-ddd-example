package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
)

type ErrorH func(c *gin.Context, message string, statusCode int)

type OAValidatorOptions struct {
	ValidateResponse bool
	ErrorHandler     ErrorH
	Options          openapi3filter.Options
	ParamDecoder     openapi3filter.ContentParameterDecoder
	UserData         any
}

type openapiMiddleware struct {
	router routers.Router
}

func NewOpenapiMiddleware(spec *openapi3.T) *openapiMiddleware {
	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		panic(fmt.Sprintf("gorillamux.NewRouter: %v", err))
	}

	return &openapiMiddleware{
		router: router,
	}
}

func (m *openapiMiddleware) RequestValidatorWithOptions(options *OAValidatorOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		rbw := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rbw

		err := validateRequest(c, m.router, options)
		if err != nil {
			if options != nil && options.ErrorHandler != nil {
				options.ErrorHandler(c, err.Error(), http.StatusBadRequest)
			} else {
				c.Error(apperrors.New(apperrors.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest))
			}

			c.Abort()

			return
		}

		c.Next()

		if options == nil || !options.ValidateResponse {
			return
		}

		if err := validateResponse(c, m.router, rbw, options); err != nil {
			// In a real app, you might log this error but still return the response,
			// or fail completely in dev environments.
			c.Error(apperrors.New(apperrors.ErrCodeInternal, fmt.Sprintf("response validation failed: %v", err), http.StatusInternalServerError))
		}
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter

	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func validateRequest(c *gin.Context, router routers.Router, options *OAValidatorOptions) error {
	route, pathParams, err := router.FindRoute(c.Request)
	if err != nil {
		return fmt.Errorf("route not found: %w", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    c.Request,
		PathParams: pathParams,
		Route:      route,
	}

	ctx := context.WithValue(c.Request.Context(), "ginContext", c)

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		ctx = context.WithValue(ctx, "userData", options.UserData)
	}

	err = openapi3filter.ValidateRequest(ctx, validationInput)
	if err != nil {
		{
			var (
				e  *openapi3filter.RequestError
				e1 *openapi3filter.SecurityRequirementsError
			)

			switch {
			case errors.As(err, &e):
				return errors.New(strings.Split(e.Error(), "\n")[0])
			case errors.As(err, &e1):
				return fmt.Errorf("security requirements failed: %w", e1)
			default:
				return err
			}
		}
	}

	return nil
}

func validateResponse(c *gin.Context, router routers.Router, rbw *responseBodyWriter, options *OAValidatorOptions) error {
	route, pathParams, err := router.FindRoute(c.Request)
	if err != nil {
		return err
	}

	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options:    &options.Options,
		},
		Status:  rbw.Status(),
		Header:  rbw.Header(),
		Options: &options.Options,
	}

	input.SetBodyBytes(rbw.body.Bytes())

	return openapi3filter.ValidateResponse(c.Request.Context(), input)
}
