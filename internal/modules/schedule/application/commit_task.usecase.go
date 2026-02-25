package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type CommitTaskCommand struct {
	TodoID uuid.UUID
	Cost   int
	Date   string
}

type CommitTaskResponse struct{}

type CommitTaskHandler struct {
	repo     domain.ScheduleRepository
	todoRepo todoDomain.TodoRepository
	uow      application.UnitOfWork
}

var _ application.RequestHandler[CommitTaskCommand, CommitTaskResponse] = (*CommitTaskHandler)(nil)

func NewCommitTaskHandler(
	repo domain.ScheduleRepository,
	todoRepo todoDomain.TodoRepository,
	uow application.UnitOfWork,
) *CommitTaskHandler {
	return &CommitTaskHandler{
		repo:     repo,
		todoRepo: todoRepo,
		uow:      uow,
	}
}

func (h *CommitTaskHandler) Handle(ctx context.Context, cmd CommitTaskCommand) (CommitTaskResponse, error) {
	meta := causation.FromContext(ctx)
	userID := userDomain.UserID(meta.UserID)

	todoID := todoDomain.TodoID(cmd.TodoID)
	date := domain.ScheduleDate(cmd.Date)

	cost, err := domain.NewEnergyCost(cmd.Cost)
	if err != nil {
		return CommitTaskResponse{}, err
	}

	err = h.uow.Execute(ctx, func(ctx context.Context) error {
		_, err := h.todoRepo.FindByID(ctx, todoID)
		if err != nil {
			return err
		}

		s, err := h.repo.FindByUserAndDate(ctx, userID, date)
		if err != nil {
			if errors.Is(err, domain.ErrScheduleNotFound) {
				s, _ = domain.NewDailySchedule(userID, date, 10)
			} else {
				return err
			}
		}

		if err := s.CommitTask(todoID, cost); err != nil {
			return err
		}

		return h.repo.Save(ctx, s)
	})
	if err != nil {
		return CommitTaskResponse{}, err
	}

	return CommitTaskResponse{}, nil
}
