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
  /resource/{path_id}:
    get:
      parameters:
        - name: path_id
          in: path
          required: true
          schema:
            type: integer
        - name: id
          in: query
          required: true
          schema:
            type: integer
            minimum: 10
        - name: x-header
          in: header
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                required: [id]
                properties:
                  id:
                    type: integer
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                age:
                  type: integer
                tags:
                  type: array
                  items:
                    type: string
              required:
                - name
                - age
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

	doc.Servers = append(doc.Servers, &openapi3.Server{URL: "/"})

	mockAuthFunc := func(c context.Context, input *openapi3filter.AuthenticationInput) error {
		if input.SecuritySchemeName != "myAuth" {
			return errors.New("forbidden")
		}

		if len(input.Scopes) > 0 && input.Scopes[0] == "someScope" {
			if input.RequestValidationInput.Request.Header.Get("Authorization") == "Bearer valid" {
				return nil
			}
		}

		return errors.New("unauthorized")
	}

	validator := middleware.NewOpenapiMiddleware(doc)
	opts := &middleware.OAValidatorOptions{
		ValidateResponse: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: mockAuthFunc,
		},
	}

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(validator.RequestValidatorWithOptions(opts))

	r.GET("/resource/:path_id", func(c *gin.Context) {
		id := c.Query("id")
		if id == "99" {
			c.JSON(http.StatusOK, gin.H{"wrong_field": 123})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": 15})
	})
	r.POST("/resource/:path_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	t.Run("valid get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource/1?id=15", nil)
		req.Header.Set("X-Header", "test-val")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid path param schema", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource/abc?id=15", nil)
		req.Header.Set("X-Header", "test-val")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["path","path_id"]`)
		assert.Contains(t, body, `invalid syntax`)
	})

	t.Run("missing required query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource/1", nil)
		req.Header.Set("X-Header", "test-val")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["query","id"]`)
		assert.Contains(t, body, `value is required but missing`)
	})

	t.Run("invalid query param schema", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource/1?id=5", nil)
		req.Header.Set("X-Header", "test-val")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["query","id"]`)
		assert.Contains(t, body, `number must be at least 10`)
	})

	t.Run("missing required header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource/1?id=15", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["header","x-header"]`)
		assert.Contains(t, body, `value is required but missing`)
	})

	t.Run("invalid post body type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource/1", strings.NewReader(`{"name": 123, "age": 20}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["body","name"]`)
	})

	t.Run("missing multiple properties", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource/1", strings.NewReader(`{"sname": "somestring"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()

		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["body","name"]`)
		assert.Contains(t, body, `"loc":["body","age"]`)
		assert.Contains(t, body, `"value":"{\"sname\":\"somestring\"}"`)
	})

	t.Run("invalid array item type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource/1", strings.NewReader(`{"name": "test", "age": 30, "tags": ["ok", 123]}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()

		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `"loc":["body","tags","1"]`)
		assert.Contains(t, body, `value must be a string`)
	})

	t.Run("malformed json body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource/1", strings.NewReader(`{"name": "test", "age": `))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, `failed to decode request body`)
	})

	t.Run("valid post", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resource/1", strings.NewReader(`{"name": "test", "age": 30}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Auth success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// auhtFunc will hit the 'someScope' check.
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("auth failure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INVALID_INPUT"`)
		assert.Contains(t, body, "Security requirements failed")
	})

	t.Run("response validation failure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resource/1?id=99", nil)
		req.Header.Set("X-Header", "test-val")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"INTERNAL_ERROR"`)
		assert.Contains(t, body, `"loc":["response","id"]`)
		assert.Contains(t, body, `property \"id\" is missing`)
	})
}
