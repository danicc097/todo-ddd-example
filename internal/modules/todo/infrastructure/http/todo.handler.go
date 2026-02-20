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
	createHandler    sharedApp.RequestHandler[application.CreateTodoCommand, domain.TodoID]
	completeHandler  sharedApp.RequestHandler[application.CompleteTodoCommand, sharedApp.Void]
	createTagHandler sharedApp.RequestHandler[application.CreateTagCommand, domain.TagID]
	assignTagHandler sharedApp.RequestHandler[application.AssignTagToTodoCommand, sharedApp.Void]

	// keeping queries nontransactional
	queryService application.TodoQueryService

	hub   *ws.Hub
	redis *redis.Client
}

func NewTodoHandler(
	c sharedApp.RequestHandler[application.CreateTodoCommand, domain.TodoID],
	comp sharedApp.RequestHandler[application.CompleteTodoCommand, sharedApp.Void],
	ct sharedApp.RequestHandler[application.CreateTagCommand, domain.TagID],
	at sharedApp.RequestHandler[application.AssignTagToTodoCommand, sharedApp.Void],
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

	todoID, err := h.createHandler.Handle(c.Request.Context(), application.CreateTodoCommand{
		Title:       req.Title,
		WorkspaceID: id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": todoID.UUID()})
}

func (h *TodoHandler) GetWorkspaceTodos(c *gin.Context, id wsDomain.WorkspaceID) {
	revision, err := h.redis.Get(c.Request.Context(), cache.Keys.WorkspaceRevision(id)).Result()
	if err == nil {
		etag := fmt.Sprintf(`"W/%s"`, revision)

		if c.Request.Header.Get("If-None-Match") == etag {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		c.Header("ETag", etag)
	}

	todos, err := h.queryService.GetAllByWorkspace(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, todos)
}

func (h *TodoHandler) GetTodoByID(c *gin.Context, id domain.TodoID) {
	todo, err := h.queryService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, todo)
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

	tagID, err := h.createTagHandler.Handle(c.Request.Context(), application.CreateTagCommand{
		Name:        req.Name,
		WorkspaceID: id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": tagID.UUID()})
}
