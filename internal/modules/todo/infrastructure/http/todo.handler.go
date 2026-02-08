package http

import (
	"net/http"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TodoHandler struct {
	createUC   *application.CreateTodoUseCase
	completeUC *application.CompleteTodoUseCase
	getAllUC   *application.GetAllTodosUseCase
	getTodoUC  *application.GetTodoUseCase
	hub        *ws.TodoHub
	mapper     *TodoRestMapper
}

var _ api.ServerInterface = (*TodoHandler)(nil)

func NewTodoHandler(c *application.CreateTodoUseCase, comp *application.CompleteTodoUseCase, g *application.GetAllTodosUseCase, gt *application.GetTodoUseCase, hub *ws.TodoHub) *TodoHandler {
	return &TodoHandler{
		createUC:   c,
		completeUC: comp,
		getAllUC:   g,
		getTodoUC:  gt,
		hub:        hub,
		mapper:     &TodoRestMapper{},
	}
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req api.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.createUC.Execute(c.Request.Context(), application.CreateTodoCommand{Title: req.Title})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	todos, err := h.getAllUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.mapper.ToResponseList(todos))
}

func (h *TodoHandler) GetTodoByID(c *gin.Context, id uuid.UUID) {
	todo, err := h.getTodoUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.mapper.ToResponse(todo))
}

func (h *TodoHandler) CompleteTodo(c *gin.Context, id uuid.UUID) {
	if err := h.completeUC.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *TodoHandler) WS(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}
