package http

import (
	"net/http"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	registerUC *application.RegisterUserUseCase
	getUserUC  *application.GetUserUseCase
	mapper     *UserRestMapper
}

func NewUserHandler(r *application.RegisterUserUseCase, g *application.GetUserUseCase) *UserHandler {
	return &UserHandler{registerUC: r, getUserUC: g, mapper: &UserRestMapper{}}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req api.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.registerUC.Execute(c.Request.Context(), application.RegisterUserCommand{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *UserHandler) GetUserByID(c *gin.Context, id uuid.UUID) {
	user, err := h.getUserUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.mapper.ToResponse(user))
}
