package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/utils/mapper"
)

type WorkspaceHandler struct {
	createUC application.CreateWorkspaceUseCase
	listUC   application.ListWorkspacesUseCase
	deleteUC application.DeleteWorkspaceUseCase
	mapper   *WorkspaceRestMapper
}

func NewWorkspaceHandler(
	createUC application.CreateWorkspaceUseCase,
	listUC application.ListWorkspacesUseCase,
	deleteUC application.DeleteWorkspaceUseCase,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		createUC: createUC,
		listUC:   listUC,
		deleteUC: deleteUC,
		mapper:   &WorkspaceRestMapper{},
	}
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context, params api.CreateWorkspaceParams) {
	var req api.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.ErrCodeInvalidInput, err.Error(), http.StatusBadRequest))
		return
	}

	cmd := application.CreateWorkspaceCommand{
		Name:        req.Name,
		Description: *req.Description,
	}

	id, err := h.createUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	list, err := h.listUC.Execute(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, mapper.MapList(list, h.mapper.ToResponse))
}

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context, id uuid.UUID) {
	if err := h.deleteUC.Execute(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
