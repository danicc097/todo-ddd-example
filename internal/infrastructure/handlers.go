package infrastructure

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authHttp "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/http"
	scheduleHttp "github.com/danicc097/todo-ddd-example/internal/modules/schedule/infrastructure/http"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	userHttp "github.com/danicc097/todo-ddd-example/internal/modules/user/infrastructure/http"
	wsHttp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/infrastructure/http"
)

type CompositeHandler struct {
	*todoHttp.TodoHandler
	*userHttp.UserHandler
	*wsHttp.WorkspaceHandler
	*authHttp.AuthHandler
	*scheduleHttp.ScheduleHandler
}

func (h *CompositeHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (h *CompositeHandler) WS(c *gin.Context) {
	h.TodoHandler.WS(c)
}
