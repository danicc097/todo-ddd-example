package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type CreateTagCommand struct {
	Name        string
	WorkspaceID wsDomain.WorkspaceID
}

type CreateTagResponse struct {
	ID domain.TagID
}

type CreateTagHandler struct {
	repo domain.TagRepository
}

var _ application.RequestHandler[CreateTagCommand, CreateTagResponse] = (*CreateTagHandler)(nil)

func NewCreateTagHandler(repo domain.TagRepository) *CreateTagHandler {
	return &CreateTagHandler{repo: repo}
}

func (h *CreateTagHandler) Handle(ctx context.Context, cmd CreateTagCommand) (CreateTagResponse, error) {
	tn, err := domain.NewTagName(cmd.Name)
	if err != nil {
		return CreateTagResponse{}, err
	}

	tag := domain.NewTag(tn, cmd.WorkspaceID)

	if err := h.repo.Save(ctx, tag); err != nil {
		return CreateTagResponse{}, err
	}

	return CreateTagResponse{ID: tag.ID()}, nil
}
