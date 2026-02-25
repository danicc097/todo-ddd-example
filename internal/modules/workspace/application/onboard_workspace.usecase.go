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
	uow          application.UnitOfWork
}

var _ application.RequestHandler[OnboardWorkspaceCommand, OnboardWorkspaceResponse] = (*OnboardWorkspaceHandler)(nil)

func NewOnboardWorkspaceHandler(repo domain.WorkspaceRepository, up UserProvider, uow application.UnitOfWork) *OnboardWorkspaceHandler {
	return &OnboardWorkspaceHandler{
		repo:         repo,
		userProvider: up,
		uow:          uow,
	}
}

func (h *OnboardWorkspaceHandler) Handle(ctx context.Context, cmd OnboardWorkspaceCommand) (OnboardWorkspaceResponse, error) {
	meta := causation.FromContext(ctx)

	var res OnboardWorkspaceResponse

	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		ownerID := cmd.OwnerID
		if ownerID.UUID() == uuid.Nil && meta.IsUser() {
			ownerID = userDomain.UserID(meta.UserID)
		}

		name, err := domain.NewWorkspaceName(cmd.Name)
		if err != nil {
			return err
		}

		desc, err := domain.NewWorkspaceDescription(cmd.Description)
		if err != nil {
			return err
		}

		ws := domain.NewWorkspace(name, desc, ownerID)

		for memberID, config := range cmd.Members {
			if memberID == ownerID {
				continue
			}

			if err := ws.AddMember(memberID, config.Role); err != nil {
				return fmt.Errorf("failed to add member %s: %w", memberID, err)
			}
		}

		if err := h.repo.Save(ctx, ws); err != nil {
			return err
		}

		res = OnboardWorkspaceResponse{ID: ws.ID()}

		return nil
	})

	return res, err
}
