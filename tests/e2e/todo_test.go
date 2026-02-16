//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/negrel/secrecy"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/generated/client"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

func TestE2E_TodoLifecycle(t *testing.T) {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://127.0.0.1:8090"
	}

	if !strings.Contains(apiURL, "/api/v1") {
		apiURL = strings.TrimSuffix(apiURL, "/") + "/api/v1"
	}

	c, err := client.NewClientWithResponses(apiURL)
	require.NoError(t, err)

	ctx := context.Background()

	pass := secrecy.NewSecret("Password123!")

	email := fmt.Sprintf("e2e-%d@example.com", time.Now().UnixNano())
	regResp, err := c.RegisterWithResponse(ctx, &client.RegisterParams{}, client.RegisterUserRequestBody{
		Email:    openapi_types.Email(email),
		Name:     "E2E User",
		Password: *pass,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, regResp.StatusCode())

	loginResp, err := c.LoginWithResponse(ctx, client.LoginRequestBody{
		Email:    openapi_types.Email(email),
		Password: *pass,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginResp.StatusCode())
	token := loginResp.JSON200.AccessToken

	c, err = client.NewClientWithResponses(apiURL, client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}))
	require.NoError(t, err)

	wsName := "E2E Workspace"
	onboardResp, err := c.OnboardWorkspaceWithResponse(ctx, &client.OnboardWorkspaceParams{}, client.OnboardWorkspaceRequest{
		Name: wsName,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, onboardResp.StatusCode())
	wsID := wsDomain.WorkspaceID{UUID: onboardResp.JSON201.Id}

	title := "E2E Test Todo"
	createResp, err := c.CreateTodoWithResponse(ctx,
		wsID,
		&client.CreateTodoParams{},
		client.CreateTodoRequest{Title: title},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode())
	require.NotNil(t, createResp.JSON201)

	todoID := domain.TodoID{UUID: createResp.JSON201.Id}

	listResp, err := c.GetWorkspaceTodosWithResponse(ctx, wsID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode())

	found := false

	for _, todo := range *listResp.JSON200 {
		if todo.Id == (todoID) {
			assert.Equal(t, title, todo.Title)
			assert.Equal(t, client.PENDING, todo.Status)
			assert.Equal(t, wsID, todo.WorkspaceId)

			found = true

			break
		}
	}

	assert.True(t, found)

	completeResp, err := c.CompleteTodoWithResponse(ctx, todoID, &client.CompleteTodoParams{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, completeResp.StatusCode())
}
