package http

import (
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	repo domain.TodoRepository
}

func (h *TodoHandler) Create(c *gin.Context) {
	h.repo.Save() // want "Arch violation: HTTP handler Create calls TodoRepository.Save directly. Handlers must route through Application UseCases."
}
