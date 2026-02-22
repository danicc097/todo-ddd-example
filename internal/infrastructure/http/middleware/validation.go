package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/danicc097/todo-ddd-example/internal"
	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	infraHttp "github.com/danicc097/todo-ddd-example/internal/infrastructure/http"
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

	return &openapiMiddleware{router: router}
}

func (m *openapiMiddleware) RequestValidatorWithOptions(options *OAValidatorOptions) gin.HandlerFunc {
	if options != nil {
		options.Options.MultiError = true
	}

	return func(c *gin.Context) {
		if err := validateRequest(c, m.router, options); err != nil {
			if options != nil && options.ErrorHandler != nil {
				options.ErrorHandler(c, err.Error(), http.StatusBadRequest)
			} else {
				valErrs := parseValidationErrors(err)
				appErr := apperrors.New(apperrors.InvalidInput, "OpenAPI request validation failed")
				appErr.Validation = any(&valErrs)

				c.Error(appErr)
			}

			c.Abort()

			return
		}

		if options == nil || !options.ValidateResponse || (internal.Config != nil && internal.Config.Env == internal.AppEnvProd) {
			c.Next()
			return
		}

		rbw := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = rbw

		c.Next()

		c.Writer = rbw.ResponseWriter

		if len(c.Errors) > 0 {
			return
		}

		if err := validateResponse(c, m.router, rbw, options); err != nil {
			slog.InfoContext(c.Request.Context(), "response validation failed")

			span := trace.SpanFromContext(c.Request.Context())
			span.RecordError(err)
			span.SetStatus(codes.Error, "response validation failed")
			span.SetAttributes(
				attribute.Bool("response.validation.failed", true),
				attribute.String("response.validation.error", err.Error()),
			)

			valErrs := parseValidationErrors(err)
			appErr := apperrors.New(apperrors.Internal, "OpenAPI response validation failed")
			appErr.Validation = any(&valErrs)

			c.Error(appErr)
			c.Abort()

			return
		}

		if rbw.status > 0 {
			c.Writer.WriteHeader(rbw.status)
		} else {
			c.Writer.WriteHeader(rbw.ResponseWriter.Status())
		}

		c.Writer.Write(rbw.body.Bytes())
	}
}

func parseValidationErrors(err error) api.HTTPValidationError {
	var (
		detail   []api.ValidationError
		messages []string
	)

	extractKinOpenApiError(err, nil, &detail, &messages)

	return api.HTTPValidationError{
		Detail:   &detail,
		Messages: messages,
	}
}

func extractKinOpenApiError(err error, baseLoc []string, detail *[]api.ValidationError, messages *[]string) {
	if err == nil {
		return
	}

	if baseLoc == nil {
		baseLoc = []string{}
	}

	var reqErr *openapi3filter.RequestError
	if errors.As(err, &reqErr) {
		loc := append([]string{}, baseLoc...)
		if reqErr.Parameter != nil {
			loc = append(loc, reqErr.Parameter.In, reqErr.Parameter.Name)
		} else {
			loc = append(loc, "body")
		}

		extractKinOpenApiError(reqErr.Err, loc, detail, messages)

		return
	}

	var respErr *openapi3filter.ResponseError
	if errors.As(err, &respErr) {
		loc := append([]string{}, baseLoc...)
		loc = append(loc, "response")
		extractKinOpenApiError(respErr.Err, loc, detail, messages)

		return
	}

	var me openapi3.MultiError
	if errors.As(err, &me) {
		for _, errItem := range me {
			extractKinOpenApiError(errItem, baseLoc, detail, messages)
		}

		return
	}

	var secErr *openapi3filter.SecurityRequirementsError
	if errors.As(err, &secErr) {
		*messages = append(*messages, "Security requirements failed")
		return
	}

	var parseErr *openapi3filter.ParseError
	if errors.As(err, &parseErr) {
		msg := parseErr.Error()
		if parseErr.Cause != nil {
			msg = parseErr.Cause.Error()
		}

		if strings.Contains(msg, "EOF") {
			msg = "failed to decode request body"
		}

		*detail = append(*detail, api.ValidationError{
			Loc: append([]string{}, baseLoc...),
			Msg: msg,
			Detail: api.ValidationErrorDetail{
				Value: fmt.Sprintf("%v", parseErr.Value),
			},
		})

		return
	}

	var schemaErr *openapi3.SchemaError
	if errors.As(err, &schemaErr) {
		loc := append([]string{}, baseLoc...)
		loc = append(loc, schemaErr.JSONPointer()...)

		var valStr string

		lastLoc := ""
		if len(loc) > 0 {
			lastLoc = loc[len(loc)-1]
		}

		if _, ok := infraHttp.SensitiveFields[lastLoc]; ok {
			valStr = "***REDACTED***"
		} else if b, jsonErr := json.Marshal(schemaErr.Value); jsonErr == nil {
			valStr = string(b)
		} else {
			valStr = fmt.Sprintf("%v", schemaErr.Value)
		}

		*detail = append(*detail, api.ValidationError{
			Loc: loc,
			Msg: schemaErr.Reason,
			Detail: api.ValidationErrorDetail{
				Value: valStr,
			},
		})

		return
	}

	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		extractKinOpenApiError(unwrapped, baseLoc, detail, messages)
		return
	}

	if len(baseLoc) > 0 {
		*detail = append(*detail, api.ValidationError{
			Loc:    append([]string{}, baseLoc...),
			Msg:    err.Error(),
			Detail: api.ValidationErrorDetail{Value: ""},
		})
	} else {
		*messages = append(*messages, err.Error())
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter

	body   *bytes.Buffer
	status int
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseBodyWriter) WriteString(s string) (int, error) {
	return r.body.WriteString(s)
}

func (r *responseBodyWriter) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *responseBodyWriter) Status() int {
	if r.status > 0 {
		return r.status
	}

	return r.ResponseWriter.Status()
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

	return openapi3filter.ValidateRequest(ctx, validationInput)
}

func validateResponse(c *gin.Context, router routers.Router, rbw *responseBodyWriter, options *OAValidatorOptions) error {
	route, pathParams, err := router.FindRoute(c.Request)
	if err != nil {
		return err
	}

	var opts *openapi3filter.Options
	if options != nil {
		opts = &options.Options
	}

	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options:    opts,
		},
		Status:  rbw.Status(),
		Header:  rbw.Header(),
		Options: opts,
	}

	input.SetBodyBytes(rbw.body.Bytes())

	return openapi3filter.ValidateResponse(c.Request.Context(), input)
}
