package http

import (
	"net/http"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TodoHandler struct {
	createUC   *application.CreateTodoUseCase
	completeUC *application.CompleteTodoUseCase
	getAllUC   *application.GetAllTodosUseCase
	hub        *ws.TodoHub
}

func NewTodoHandler(c *application.CreateTodoUseCase, comp *application.CompleteTodoUseCase, g *application.GetAllTodosUseCase, hub *ws.TodoHub) *TodoHandler {
	return &TodoHandler{createUC: c, completeUC: comp, getAllUC: g, hub: hub}
}

// todoResponse is the DTO for a Todo.
type todoResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"createdAt"`
}

func (h *TodoHandler) Create(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
	}
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

func (h *TodoHandler) Complete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}
	if err := h.completeUC.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *TodoHandler) WS(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}

func (h *TodoHandler) GetAll(c *gin.Context) {
	todos, err := h.getAllUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]todoResponse, len(todos))
	for i, t := range todos {
		response[i] = todoResponse{
			ID:        t.ID(),
			Title:     t.Title().String(),
			Status:    t.Status().String(),
			CreatedAt: t.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, response)
}
