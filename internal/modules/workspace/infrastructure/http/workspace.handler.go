package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedApp "github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type WorkspaceHandler struct {
	onboardHandler      sharedApp.RequestHandler[application.OnboardWorkspaceCommand, application.OnboardWorkspaceResponse]
	addMemberHandler    sharedApp.RequestHandler[application.AddWorkspaceMemberCommand, sharedApp.Void]
	removeMemberHandler sharedApp.RequestHandler[application.RemoveWorkspaceMemberCommand, sharedApp.Void]
	deleteHandler       sharedApp.RequestHandler[application.DeleteWorkspaceCommand, sharedApp.Void]

	queryService application.WorkspaceQueryService
}

func NewWorkspaceHandler(
	onboardHandler sharedApp.RequestHandler[application.OnboardWorkspaceCommand, application.OnboardWorkspaceResponse],
	addMemberHandler sharedApp.RequestHandler[application.AddWorkspaceMemberCommand, sharedApp.Void],
	removeMemberHandler sharedApp.RequestHandler[application.RemoveWorkspaceMemberCommand, sharedApp.Void],
	qs application.WorkspaceQueryService,
	deleteHandler sharedApp.RequestHandler[application.DeleteWorkspaceCommand, sharedApp.Void],
) *WorkspaceHandler {
	return &WorkspaceHandler{
		onboardHandler:      onboardHandler,
		addMemberHandler:    addMemberHandler,
		removeMemberHandler: removeMemberHandler,
		queryService:        qs,
		deleteHandler:       deleteHandler,
	}
}

func (h *WorkspaceHandler) OnboardWorkspace(c *gin.Context, params api.OnboardWorkspaceParams) {
	var req api.OnboardWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	members := make(map[userDomain.UserID]application.MemberInitialState)

	if req.Members != nil {
		for uid, role := range *req.Members {
			members[userDomain.UserID(uid)] = application.MemberInitialState{Role: role}
		}
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	cmd := application.OnboardWorkspaceCommand{
		Name:        req.Name,
		Description: description,
		Members:     members,
		OwnerID:     userDomain.UserID{},
	}

	resp, err := h.onboardHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": resp.ID.UUID()})
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context, params api.ListWorkspacesParams) {
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	list, err := h.queryService.List(c.Request.Context(), int32(limit), int32(offset))
	if err != nil {
		c.Error(err)
		return
	}

	apiWorkspaces := make([]api.Workspace, len(list))
	for i, w := range list {
		apiWorkspaces[i] = api.Workspace{
			Id:          w.ID,
			Name:        w.Name,
			Description: w.Description,
		}
	}

	c.JSON(http.StatusOK, apiWorkspaces)
}

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context, id domain.WorkspaceID) {
	cmd := application.DeleteWorkspaceCommand{
		ID: id,
	}

	if _, err := h.deleteHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) AddWorkspaceMember(c *gin.Context, id domain.WorkspaceID, params api.AddWorkspaceMemberParams) {
	var req struct {
		UserID userDomain.UserID    `json:"userId"`
		Role   domain.WorkspaceRole `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(apperrors.InvalidInput, err.Error()))
		return
	}

	if _, err := h.addMemberHandler.Handle(c.Request.Context(), application.AddWorkspaceMemberCommand{
		WorkspaceID: id,
		UserID:      req.UserID,
		Role:        req.Role,
	}); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) RemoveWorkspaceMember(c *gin.Context, id domain.WorkspaceID, userID userDomain.UserID) {
	cmd := application.RemoveWorkspaceMemberCommand{
		WorkspaceID: id,
		MemberID:    userID,
	}

	if _, err := h.removeMemberHandler.Handle(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) GetWorkspaceTags(c *gin.Context, id domain.WorkspaceID) {
	tags, err := h.queryService.ListTagsByWorkspaceID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	apiTags := make([]api.Tag, len(tags))
	for i, t := range tags {
		apiTags[i] = api.Tag{
			Id:   t.ID,
			Name: t.Name,
		}
	}

	c.JSON(http.StatusOK, apiTags)
}
