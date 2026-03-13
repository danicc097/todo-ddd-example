package infrastructure

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authHttp "github.com/danicc097/todo-ddd-example/internal/modules/auth/infrastructure/http"
	scheduleHttp "github.com/danicc097/todo-ddd-example/internal/modules/schedule/infrastructure/http"
	todoHttp "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/http"
	todoWS "github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
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

func NewHandlers(s *Services, c *Container) *CompositeHandler {
	hub := todoWS.NewTodoHub(c.Redis, s.WorkspaceQuery)

	return &CompositeHandler{
		TodoHandler:      todoHttp.NewTodoHandler(s.Todo, s.TodoQuery, hub, c.Redis),
		UserHandler:      userHttp.NewUserHandler(s.UserQuery, s.WorkspaceQuery),
		WorkspaceHandler: wsHttp.NewWorkspaceHandler(s.Workspace, s.WorkspaceQuery),
		AuthHandler:      authHttp.NewAuthHandler(s.Auth),
		ScheduleHandler:  scheduleHttp.NewScheduleHandler(s.Schedule),
	}
}

func (h *CompositeHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (h *CompositeHandler) WS(c *gin.Context) {
	h.TodoHandler.WS(c)
}
