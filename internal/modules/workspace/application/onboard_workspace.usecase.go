package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type MemberInitialState struct {
	Role domain.WorkspaceRole
}

type OnboardWorkspaceCommand struct {
	Name        string
	Description string
	OwnerID     userDomain.UserID
	Members     map[userDomain.UserID]MemberInitialState
}

func (c *OnboardWorkspaceCommand) Validate() error {
	if _, err := domain.NewWorkspaceName(c.Name); err != nil {
		return err
	}

	if _, err := domain.NewWorkspaceDescription(c.Description); err != nil {
		return err
	}

	return nil
}

type OnboardWorkspaceResponse struct {
	ID domain.WorkspaceID
}

type OnboardWorkspaceHandler struct {
	repo         domain.WorkspaceRepository
	userProvider UserProvider
}

var _ application.RequestHandler[OnboardWorkspaceCommand, OnboardWorkspaceResponse] = (*OnboardWorkspaceHandler)(nil)

func NewOnboardWorkspaceHandler(repo domain.WorkspaceRepository, up UserProvider) *OnboardWorkspaceHandler {
	return &OnboardWorkspaceHandler{
		repo:         repo,
		userProvider: up,
	}
}

func (h *OnboardWorkspaceHandler) Handle(ctx context.Context, cmd OnboardWorkspaceCommand) (OnboardWorkspaceResponse, error) {
	meta := causation.FromContext(ctx)

	ownerID := cmd.OwnerID
	if ownerID.UUID() == uuid.Nil && meta.IsUser() {
		ownerID = userDomain.UserID(meta.UserID)
	}

	name, _ := domain.NewWorkspaceName(cmd.Name)
	desc, _ := domain.NewWorkspaceDescription(cmd.Description)

	ws := domain.NewWorkspace(name, desc, ownerID)

	for memberID, config := range cmd.Members {
		if memberID == ownerID {
			continue
		}

		if err := ws.AddMember(memberID, config.Role); err != nil {
			return OnboardWorkspaceResponse{}, fmt.Errorf("failed to add member %s: %w", memberID, err)
		}
	}

	if err := h.repo.Save(ctx, ws); err != nil {
		return OnboardWorkspaceResponse{}, err
	}

	return OnboardWorkspaceResponse{ID: ws.ID()}, nil
}
