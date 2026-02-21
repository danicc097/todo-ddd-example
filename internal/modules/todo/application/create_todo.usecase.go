package application

import (
	"context"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type CreateTodoCommand struct {
	Title       string
	WorkspaceID wsDomain.WorkspaceID
	TagIDs      []domain.TagID
}

type CreateTodoResponse struct {
	ID domain.TodoID
}

type CreateTodoHandler struct {
	repo   domain.TodoRepository
	wsProv WorkspaceProvider
}

var _ application.RequestHandler[CreateTodoCommand, CreateTodoResponse] = (*CreateTodoHandler)(nil)

func NewCreateTodoHandler(repo domain.TodoRepository, wsProv WorkspaceProvider) *CreateTodoHandler {
	return &CreateTodoHandler{repo: repo, wsProv: wsProv}
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (CreateTodoResponse, error) {
	meta := causation.FromContext(ctx)

	isMember, err := h.wsProv.IsMember(ctx, cmd.WorkspaceID, userDomain.UserID(meta.UserID))
	if err != nil {
		return CreateTodoResponse{}, err
	}

	if !isMember && !meta.IsSystem() {
		return CreateTodoResponse{}, wsDomain.ErrNotOwner
	}

	title, err := domain.NewTodoTitle(cmd.Title)
	if err != nil {
		return CreateTodoResponse{}, err
	}

	todo := domain.NewTodo(title, cmd.WorkspaceID)
	for _, tagID := range cmd.TagIDs {
		todo.AddTag(tagID)
	}

	if err := h.repo.Save(ctx, todo); err != nil {
		return CreateTodoResponse{}, err
	}

	return CreateTodoResponse{ID: todo.ID()}, nil
}
