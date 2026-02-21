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
		// fallback to letting the authn user be the owner
		// if the system were to create a workspace on behalg of a user we would still know
		// the system created it via the causation pkg.
		ownerID = userDomain.UserID(meta.UserID)
	}

	name, err := domain.NewWorkspaceName(cmd.Name)
	if err != nil {
		return OnboardWorkspaceResponse{}, err
	}

	desc, err := domain.NewWorkspaceDescription(cmd.Description)
	if err != nil {
		return OnboardWorkspaceResponse{}, err
	}

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
