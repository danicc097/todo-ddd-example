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

type OnboardWorkspaceHandler struct {
	repo        domain.WorkspaceRepository
	userGateway UserGateway
}

var _ application.RequestHandler[OnboardWorkspaceCommand, domain.WorkspaceID] = (*OnboardWorkspaceHandler)(nil)

func NewOnboardWorkspaceHandler(repo domain.WorkspaceRepository, ug UserGateway) *OnboardWorkspaceHandler {
	return &OnboardWorkspaceHandler{
		repo:        repo,
		userGateway: ug,
	}
}

func (h *OnboardWorkspaceHandler) Handle(ctx context.Context, cmd OnboardWorkspaceCommand) (domain.WorkspaceID, error) {
	meta := causation.FromContext(ctx)

	ownerID := cmd.OwnerID
	if ownerID.UUID() == uuid.Nil && meta.IsUser() {
		// fallback to letting the authn user be the owner
		// if the system were to create a workspace on behalg of a user we would still know
		// the system created it via the causation pkg.
		ownerID = userDomain.UserID(meta.UserID)
	}

	ws := domain.NewWorkspace(cmd.Name, cmd.Description, ownerID)

	for memberID, config := range cmd.Members {
		if memberID == ownerID {
			continue
		}

		if err := ws.AddMember(memberID, config.Role); err != nil {
			return domain.WorkspaceID{}, fmt.Errorf("failed to add member %s: %w", memberID, err)
		}
	}

	if err := h.repo.Save(ctx, ws); err != nil {
		return domain.WorkspaceID{}, err
	}

	return ws.ID(), nil
}
