package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
	infraHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type TodoUseCases struct {
	CreateTodo sharedApp.RequestHandler[application.CreateTodoCommand, application.CreateTodoResponse]
	Complete   sharedApp.RequestHandler[application.CompleteTodoCommand, application.CompleteTodoResponse]
	CreateTag  sharedApp.RequestHandler[application.CreateTagCommand, application.CreateTagResponse]
	AssignTag  sharedApp.RequestHandler[application.AssignTagToTodoCommand, application.AssignTagToTodoResponse]
	StartFocus sharedApp.RequestHandler[application.StartFocusCommand, application.StartFocusResponse]
	StopFocus  sharedApp.RequestHandler[application.StopFocusCommand, application.StopFocusResponse]
}

type TodoHandler struct {
	uc           TodoUseCases
	queryService application.TodoQueryService
	hub          *ws.Hub
	redis        *redis.Client
}

func NewTodoHandler(
	uc TodoUseCases,
	qs application.TodoQueryService,
	hub *ws.Hub,
	redis *redis.Client,
) *TodoHandler {
	return &TodoHandler{
		uc:           uc,
		queryService: qs,
		hub:          hub,
		redis:        redis,
	}
}

func (h *TodoHandler) WS(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}

func (h *TodoHandler) CreateTodo(c *gin.Context, id wsDomain.WorkspaceID, params api.CreateTodoParams) {
	req, ok := infraHttp.BindJSON[api.CreateTodoRequest](c)
	if !ok {
		return
	}

	resp, ok := infraHttp.Execute(c, h.uc.CreateTodo, application.CreateTodoCommand{
		Title:              req.Title,
		WorkspaceID:        id,
		DueDate:            req.DueDate,
		RecurrenceInterval: (*string)(req.RecurrenceInterval),
		RecurrenceAmount:   req.RecurrenceAmount,
	})
	if ok {
		c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
	}
}

func (h *TodoHandler) GetWorkspaceTodos(c *gin.Context, id wsDomain.WorkspaceID, params api.GetWorkspaceTodosParams) {
	revision, err := h.redis.Get(c.Request.Context(), cache.Keys.WorkspaceRevision(id)).Result()
	if err == nil {
		etag := fmt.Sprintf(`"W/%s"`, revision)

		if c.Request.Header.Get("If-None-Match") == etag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		c.Header("ETag", etag)
	}

	limit := infraHttp.DefaultPaginationLimit
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	todos, err := h.queryService.GetAllByWorkspace(c.Request.Context(), id, int32(limit), int32(offset))
	if err != nil {
		c.Error(err)
		return
	}

	apiTodos := make([]api.Todo, len(todos))
	for i, t := range todos {
		apiTodos[i] = h.mapReadModelToAPI(t)
	}

	c.JSON(http.StatusOK, apiTodos)
}

func (h *TodoHandler) GetTodoByID(c *gin.Context, id domain.TodoID) {
	todo, err := h.queryService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, h.mapReadModelToAPI(*todo))
}

func (h *TodoHandler) mapReadModelToAPI(t application.TodoReadModel) api.Todo {
	sessions := make([]api.FocusSession, len(t.FocusSessions))
	for i, s := range t.FocusSessions {
		sessions[i] = api.FocusSession{
			Id:        s.ID,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
		}
	}

	return api.Todo{
		Id:                 t.ID,
		WorkspaceId:        t.WorkspaceID,
		Title:              t.Title,
		Status:             api.TodoStatus(t.Status),
		CreatedAt:          t.CreatedAt,
		DueDate:            t.DueDate,
		RecurrenceInterval: (*api.RecurrenceInterval)(t.RecurrenceInterval),
		RecurrenceAmount:   t.RecurrenceAmount,
		FocusSessions:      &sessions,
	}
}

func (h *TodoHandler) CompleteTodo(c *gin.Context, id domain.TodoID, params api.CompleteTodoParams) {
	if _, ok := infraHttp.Execute(c, h.uc.Complete, application.CompleteTodoCommand{ID: id}); ok {
		c.Status(http.StatusOK)
	}
}

func (h *TodoHandler) AssignTagToTodo(c *gin.Context, id domain.TodoID, params api.AssignTagToTodoParams) {
	req, ok := infraHttp.BindJSON[api.AssignTagToTodoRequest](c)
	if !ok {
		return
	}

	if _, ok := infraHttp.Execute(c, h.uc.AssignTag, application.AssignTagToTodoCommand{
		TodoID: id,
		TagID:  req.TagId,
	}); ok {
		c.Status(http.StatusNoContent)
	}
}

func (h *TodoHandler) StartFocus(c *gin.Context, id domain.TodoID) {
	if _, ok := infraHttp.Execute(c, h.uc.StartFocus, application.StartFocusCommand{ID: id}); ok {
		c.Status(http.StatusNoContent)
	}
}

func (h *TodoHandler) StopFocus(c *gin.Context, id domain.TodoID) {
	if _, ok := infraHttp.Execute(c, h.uc.StopFocus, application.StopFocusCommand{ID: id}); ok {
		c.Status(http.StatusNoContent)
	}
}

func (h *TodoHandler) CreateTag(c *gin.Context, id wsDomain.WorkspaceID, params api.CreateTagParams) {
	req, ok := infraHttp.BindJSON[api.CreateTagRequest](c)
	if !ok {
		return
	}

	resp, ok := infraHttp.Execute(c, h.uc.CreateTag, application.CreateTagCommand{
		Name:        req.Name,
		WorkspaceID: id,
	})
	if ok {
		c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
	}
}
