package http

import (
	"net/http"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TodoHandler struct {
	createUC   *application.CreateTodoUseCase
	completeUC *application.CompleteTodoUseCase
	getAllUC   *application.GetAllTodosUseCase
}

func NewTodoHandler(
	create *application.CreateTodoUseCase,
	complete *application.CompleteTodoUseCase,
	getAll *application.GetAllTodosUseCase,
) *TodoHandler {
	return &TodoHandler{
		createUC:   create,
		completeUC: complete,
		getAllUC:   getAll,
	}
}

type createRequest struct {
	Title string `json:"title" binding:"required"`
}

func (h *TodoHandler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := application.CreateTodoCommand{Title: req.Title}

	id, err := h.createUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		if err == domain.ErrEmptyTitle {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *TodoHandler) Complete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.completeUC.Execute(c.Request.Context(), id); err != nil {
		if err == domain.ErrTodoNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *TodoHandler) GetAll(c *gin.Context) {
	todos, err := h.getAllUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}
