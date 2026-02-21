package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type DeleteUserCommand struct {
	ID domain.UserID
}

type DeleteUserResponse struct{}

type DeleteUserHandler struct {
	repo domain.UserRepository
}

var _ application.RequestHandler[DeleteUserCommand, DeleteUserResponse] = (*DeleteUserHandler)(nil)

func NewDeleteUserHandler(repo domain.UserRepository) *DeleteUserHandler {
	return &DeleteUserHandler{repo: repo}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUserCommand) (DeleteUserResponse, error) {
	_ = causation.FromContext(ctx)

	// TODO: should check authz
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		return DeleteUserResponse{}, err
	}

	return DeleteUserResponse{}, nil
}
