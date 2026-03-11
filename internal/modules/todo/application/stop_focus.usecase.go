package application

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type StopFocusCommand struct {
	ID domain.TodoID
}

type StopFocusResponse struct{}

type StopFocusHandler struct {
	repo   domain.TodoRepository
	wsProv WorkspaceProvider
}

var _ application.RequestHandler[StopFocusCommand, StopFocusResponse] = (*StopFocusHandler)(nil)

func NewStopFocusHandler(repo domain.TodoRepository, wsProv WorkspaceProvider) *StopFocusHandler {
	return &StopFocusHandler{repo: repo, wsProv: wsProv}
}

func (h *StopFocusHandler) Handle(ctx context.Context, cmd StopFocusCommand) (StopFocusResponse, error) {
	meta := causation.FromContext(ctx)

	todo, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return StopFocusResponse{}, err
	}

	isMember, err := h.wsProv.IsMember(ctx, todo.WorkspaceID(), userDomain.UserID(meta.UserID))
	if err != nil {
		return StopFocusResponse{}, err
	}

	if !isMember && !meta.IsSystem() {
		return StopFocusResponse{}, wsDomain.ErrNotOwner
	}

	if err := todo.StopFocus(time.Now()); err != nil {
		return StopFocusResponse{}, err
	}

	return StopFocusResponse{}, h.repo.Save(ctx, todo)
}
