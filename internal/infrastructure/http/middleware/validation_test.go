package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/http/middleware"
)

var testSchema = []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /resource:
    get:
      parameters:
        - name: id
          in: query
          schema:
            type: integer
            minimum: 10
      responses:
        '200':
          description: OK
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
              required:
                - name
      responses:
        '204':
          description: No Content
  /protected:
    get:
      security:
        - myAuth: [someScope]
      responses:
        '204':
          description: No Content
components:
  securitySchemes:
    myAuth:
      type: http
      scheme: bearer
`)

func TestOapiRequestValidator(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(testSchema)
	require.NoError(t, err)

	doc.Servers = append(doc.Servers, &openapi3.Server{URL: "/"}) // required

	mockAuthFunc := func(c context.Context, input *openapi3filter.AuthenticationInput) error {
		if input.SecuritySchemeName != "myAuth" {
			return errors.New("forbidden")
		}

		if len(input.Scopes) > 0 && input.Scopes[0] == "someScope" {
			return nil
		}

		return errors.New("forbidden")
	}

	validator := middleware.NewOpenapiMiddleware(doc)
	opts := &middleware.OAValidatorOptions{
		Options: openapi3filter.Options{
			AuthenticationFunc: mockAuthFunc,
		},
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			c.String(statusCode, "custom: "+message)
		},
	}

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(validator.RequestValidatorWithOptions(opts))

	r.GET("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/resource", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	t.Run("Valid GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource?id=15", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource?id=5", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "custom: parameter \"id\"")
	})

	t.Run("Invalid POST body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(`{"name": 123}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "request body has an error")
	})

	t.Run("Valid POST body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource", strings.NewReader(`{"name": "test"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Auth success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// auhtFunc will hit the 'someScope' check.
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
