package application

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/utils/pointers"
)

type CreateTodoCommand struct {
	Title              string
	WorkspaceID        wsDomain.WorkspaceID
	TagIDs             []domain.TagID
	DueDate            *time.Time
	RecurrenceInterval *string
	RecurrenceAmount   *int
}

type CreateTodoResponse struct {
	ID domain.TodoID
}

type CreateTodoHandler struct {
	repo   domain.TodoRepository
	wsProv WorkspaceProvider
	uow    application.UnitOfWork
}

var _ application.RequestHandler[CreateTodoCommand, CreateTodoResponse] = (*CreateTodoHandler)(nil)

func NewCreateTodoHandler(repo domain.TodoRepository, wsProv WorkspaceProvider, uow application.UnitOfWork) *CreateTodoHandler {
	return &CreateTodoHandler{repo: repo, wsProv: wsProv, uow: uow}
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) (CreateTodoResponse, error) {
	meta := causation.FromContext(ctx)

	var res CreateTodoResponse

	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		isMember, err := h.wsProv.IsMember(ctx, cmd.WorkspaceID, userDomain.UserID(meta.UserID))
		if err != nil {
			return err
		}

		if !isMember && !meta.IsSystem() {
			return wsDomain.ErrNotOwner
		}

		title, err := domain.NewTodoTitle(cmd.Title)
		if err != nil {
			return err
		}

		todo := domain.NewTodo(title, cmd.WorkspaceID)
		for _, tagID := range cmd.TagIDs {
			todo.AddTag(tagID)
		}

		if cmd.DueDate != nil {
			todo.SetDueDate(cmd.DueDate)
		}

		if cmd.RecurrenceInterval != nil && cmd.RecurrenceAmount != nil {
			r, err := domain.NewRecurrenceRule(*cmd.RecurrenceInterval, *cmd.RecurrenceAmount)
			if err != nil {
				return err
			}

			todo.SetRecurrence(pointers.New(r))
		}

		if err := h.repo.Save(ctx, todo); err != nil {
			return err
		}

		res = CreateTodoResponse{ID: todo.ID()}

		return nil
	})

	return res, err
}
