package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type TodoHandler struct {
	createHandler    sharedApp.RequestHandler[application.CreateTodoCommand, application.CreateTodoResponse]
	completeHandler  sharedApp.RequestHandler[application.CompleteTodoCommand, application.CompleteTodoResponse]
	createTagHandler sharedApp.RequestHandler[application.CreateTagCommand, application.CreateTagResponse]
	assignTagHandler sharedApp.RequestHandler[application.AssignTagToTodoCommand, application.AssignTagToTodoResponse]

	// keeping queries nontransactional
	queryService application.TodoQueryService

	hub   *ws.Hub
	redis *redis.Client
}

func NewTodoHandler(
	c sharedApp.RequestHandler[application.CreateTodoCommand, application.CreateTodoResponse],
	comp sharedApp.RequestHandler[application.CompleteTodoCommand, application.CompleteTodoResponse],
	ct sharedApp.RequestHandler[application.CreateTagCommand, application.CreateTagResponse],
	at sharedApp.RequestHandler[application.AssignTagToTodoCommand, application.AssignTagToTodoResponse],
	qs application.TodoQueryService,
	hub *ws.Hub,
	redis *redis.Client,
) *TodoHandler {
	return &TodoHandler{
		createHandler:    c,
		completeHandler:  comp,
		createTagHandler: ct,
		assignTagHandler: at,
		queryService:     qs,
		hub:              hub,
		redis:            redis,
	}
}

func (h *TodoHandler) WS(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}

func (h *TodoHandler) CreateTodo(c *gin.Context, id wsDomain.WorkspaceID, params api.CreateTodoParams) {
	var req api.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.createHandler.Handle(c.Request.Context(), application.CreateTodoCommand{
		Title:       req.Title,
		WorkspaceID: id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
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

	limit := 20
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
		apiTodos[i] = api.Todo{
			Id:          t.ID,
			WorkspaceId: t.WorkspaceID,
			Title:       t.Title,
			Status:      api.TodoStatus(t.Status),
			CreatedAt:   t.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, apiTodos)
}

func (h *TodoHandler) GetTodoByID(c *gin.Context, id domain.TodoID) {
	todo, err := h.queryService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.Todo{
		Id:          todo.ID,
		WorkspaceId: todo.WorkspaceID,
		Title:       todo.Title,
		Status:      api.TodoStatus(todo.Status),
		CreatedAt:   todo.CreatedAt,
	})
}

func (h *TodoHandler) CompleteTodo(c *gin.Context, id domain.TodoID, params api.CompleteTodoParams) {
	if _, err := h.completeHandler.Handle(c.Request.Context(), application.CompleteTodoCommand{ID: id}); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *TodoHandler) AssignTagToTodo(c *gin.Context, id domain.TodoID, params api.AssignTagToTodoParams) {
	var req api.AssignTagToTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	if _, err := h.assignTagHandler.Handle(c.Request.Context(), application.AssignTagToTodoCommand{
		TodoID: id,
		TagID:  req.TagId,
	}); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TodoHandler) CreateTag(c *gin.Context, id wsDomain.WorkspaceID, params api.CreateTagParams) {
	var req api.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	resp, err := h.createTagHandler.Handle(c.Request.Context(), application.CreateTagCommand{
		Name:        req.Name,
		WorkspaceID: id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
}
