package application

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (c *CommitTaskCommand) Validate() error {
	if _, err := domain.NewEnergyCost(c.Cost); err != nil {
		return err
	}

	if _, err := time.Parse(time.DateOnly, c.Date); err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	return nil
}

type CommitTaskResponse struct{}

type CommitTaskHandler struct {
	repo     domain.ScheduleRepository
	todoRepo todoDomain.TodoRepository
}

var _ application.RequestHandler[CommitTaskCommand, CommitTaskResponse] = (*CommitTaskHandler)(nil)

func NewCommitTaskHandler(
	repo domain.ScheduleRepository,
	todoRepo todoDomain.TodoRepository,
) *CommitTaskHandler {
	return &CommitTaskHandler{
		repo:     repo,
		todoRepo: todoRepo,
	}
}

func (h *CommitTaskHandler) Handle(ctx context.Context, cmd CommitTaskCommand) (CommitTaskResponse, error) {
	meta := causation.FromContext(ctx)
	userID := userDomain.UserID(meta.UserID)

	todoID := todoDomain.TodoID(cmd.TodoID)
	date := domain.ScheduleDate(cmd.Date)

	cost, _ := domain.NewEnergyCost(cmd.Cost)

	_, err := h.todoRepo.FindByID(ctx, todoID)
	if err != nil {
		return CommitTaskResponse{}, err
	}

	s, err := h.repo.FindByUserAndDate(ctx, userID, date)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			s, _ = domain.NewDailySchedule(userID, date, 10)
		} else {
			return CommitTaskResponse{}, err
		}
	}

	if err := s.CommitTask(todoID, cost); err != nil {
		return CommitTaskResponse{}, err
	}

	return CommitTaskResponse{}, h.repo.Save(ctx, s)
}
