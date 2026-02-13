package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type TodoHandler struct {
	createHandler    sharedApp.RequestHandler[application.CreateTodoCommand, domain.TodoID]
	completeHandler  sharedApp.RequestHandler[application.CompleteTodoCommand, sharedApp.Void]
	createTagHandler sharedApp.RequestHandler[application.CreateTagCommand, domain.TagID]

	// keeping queries nontransactional
	queryService application.TodoQueryService

	hub *ws.TodoHub
}

func NewTodoHandler(
	c sharedApp.RequestHandler[application.CreateTodoCommand, domain.TodoID],
	comp sharedApp.RequestHandler[application.CompleteTodoCommand, sharedApp.Void],
	ct sharedApp.RequestHandler[application.CreateTagCommand, domain.TagID],
	qs application.TodoQueryService,
	hub *ws.TodoHub,
) *TodoHandler {
	return &TodoHandler{
		createHandler:    c,
		completeHandler:  comp,
		createTagHandler: ct,
		queryService:     qs,
		hub:              hub,
	}
}

func (h *TodoHandler) WS(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}

func (h *TodoHandler) CreateTodo(c *gin.Context, params api.CreateTodoParams) {
	var req api.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest))
		return
	}

	id, err := h.createHandler.Handle(c.Request.Context(), application.CreateTodoCommand{Title: req.Title})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.UUID})
}

func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	todos, err := h.queryService.GetAll(c.Request.Context())
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

func (h *TodoHandler) CreateTag(c *gin.Context, params api.CreateTagParams) {
	var req api.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest))
		return
	}

	id, err := h.createTagHandler.Handle(c.Request.Context(), application.CreateTagCommand{Name: req.Name})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.UUID})
}
