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
	uow  application.UnitOfWork
}

var _ application.RequestHandler[CreateTagCommand, CreateTagResponse] = (*CreateTagHandler)(nil)

func NewCreateTagHandler(repo domain.TagRepository, uow application.UnitOfWork) *CreateTagHandler {
	return &CreateTagHandler{repo: repo, uow: uow}
}

func (h *CreateTagHandler) Handle(ctx context.Context, cmd CreateTagCommand) (CreateTagResponse, error) {
	tn, err := domain.NewTagName(cmd.Name)
	if err != nil {
		return CreateTagResponse{}, err
	}

	var res CreateTagResponse

	err = h.uow.Execute(ctx, func(ctx context.Context) error {
		tag := domain.NewTag(tn, cmd.WorkspaceID)

		if err := h.repo.Save(ctx, tag); err != nil {
			return err
		}

		res = CreateTagResponse{ID: tag.ID()}

		return nil
	})

	return res, err
}
