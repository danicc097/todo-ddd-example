package http

import (
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

func (h *TodoHandler) Create(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	id, err := h.createUC.Execute(c.Request.Context(), application.CreateTodoCommand{Title: req.Title})
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"id": id})
}

func (h *TodoHandler) Complete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.completeUC.Execute(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(200)
}

func (h *TodoHandler) WS(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}
