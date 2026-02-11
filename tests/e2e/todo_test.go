//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/generated/client"
)

func TestE2E_TodoLifecycle(t *testing.T) {
	apiURL := "http://127.0.0.1:8090/api/v1"
	c, err := client.NewClientWithResponses(apiURL)
	require.NoError(t, err)

	ctx := context.Background()

	title := "E2E Test Todo"
	createResp, err := c.CreateTodoWithResponse(ctx,
		&client.CreateTodoParams{},
		client.CreateTodoRequest{Title: title},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode())
	require.NotNil(t, createResp.JSON201)

	todoID := createResp.JSON201.Id

	listResp, err := c.GetAllTodosWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode())

	found := false

	for _, todo := range *listResp.JSON200 {
		if todo.Id == *todoID {
			assert.Equal(t, title, todo.Title)
			assert.Equal(t, client.PENDING, todo.Status)

			found = true

			break
		}
	}

	assert.True(t, found)

	completeResp, err := c.CompleteTodoWithResponse(ctx, *todoID, &client.CompleteTodoParams{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, completeResp.StatusCode())
}
