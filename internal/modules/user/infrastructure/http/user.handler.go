package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/modules/user/application"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	workspaceApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
)

type UserHandler struct {
	getUserUC             *application.GetUserUseCase
	workspaceQueryService workspaceApp.WorkspaceQueryService
	mapper                *UserRestMapper
}

func NewUserHandler(g *application.GetUserUseCase, wqs workspaceApp.WorkspaceQueryService) *UserHandler {
	return &UserHandler{
		getUserUC:             g,
		workspaceQueryService: wqs,
		mapper:                &UserRestMapper{},
	}
}

func (h *UserHandler) GetUserByID(c *gin.Context, id userDomain.UserID) {
	user, err := h.getUserUC.Execute(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, h.mapper.ToResponse(user))
}

func (h *UserHandler) GetUserWorkspaces(c *gin.Context, id userDomain.UserID) {
	workspaces, err := h.workspaceQueryService.ListByUserID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, workspaces)
}
