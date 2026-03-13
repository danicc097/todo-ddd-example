package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	infraHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

type WorkspaceHandler struct {
	uc           application.WorkspaceUseCases
	queryService application.WorkspaceQueryService
}

func NewWorkspaceHandler(uc application.WorkspaceUseCases, qs application.WorkspaceQueryService) *WorkspaceHandler {
	return &WorkspaceHandler{
		uc:           uc,
		queryService: qs,
	}
}

func (h *WorkspaceHandler) OnboardWorkspace(c *gin.Context, params api.OnboardWorkspaceParams) {
	req, ok := infraHttp.BindJSON[api.OnboardWorkspaceRequest](c)
	if !ok {
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

	resp, ok := infraHttp.Execute(c, h.uc.Onboard, application.OnboardWorkspaceCommand{
		Name:        req.Name,
		Description: description,
		Members:     members,
		OwnerID:     userDomain.UserID{},
	})
	if ok {
		c.JSON(http.StatusCreated, api.IdResponse{Id: resp.ID.UUID()})
	}
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context, params api.ListWorkspacesParams) {
	limit := infraHttp.DefaultPaginationLimit
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
	if _, ok := infraHttp.Execute(c, h.uc.Delete, application.DeleteWorkspaceCommand{ID: id}); ok {
		c.Status(http.StatusNoContent)
	}
}

func (h *WorkspaceHandler) AddWorkspaceMember(c *gin.Context, id domain.WorkspaceID, params api.AddWorkspaceMemberParams) {
	req, ok := infraHttp.BindJSON[api.AddWorkspaceMemberRequest](c)
	if !ok {
		return
	}

	if _, ok := infraHttp.Execute(c, h.uc.AddMember, application.AddWorkspaceMemberCommand{
		WorkspaceID: id,
		UserID:      userDomain.UserID(req.UserId),
		Role:        req.Role,
	}); ok {
		c.Status(http.StatusNoContent)
	}
}

func (h *WorkspaceHandler) RemoveWorkspaceMember(c *gin.Context, id domain.WorkspaceID, userID userDomain.UserID) {
	if _, ok := infraHttp.Execute(c, h.uc.RemoveMember, application.RemoveWorkspaceMemberCommand{
		WorkspaceID: id,
		MemberID:    userID,
	}); ok {
		c.Status(http.StatusNoContent)
	}
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
