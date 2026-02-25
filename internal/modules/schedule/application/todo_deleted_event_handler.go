package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TodoDeletedEventPayload struct {
	ID uuid.UUID `json:"id"`
}

type TodoDeletedEventHandler struct {
	repo domain.ScheduleRepository
}

func NewTodoDeletedEventHandler(repo domain.ScheduleRepository) *TodoDeletedEventHandler {
	return &TodoDeletedEventHandler{repo: repo}
}

func (h *TodoDeletedEventHandler) Handle(ctx context.Context, data []byte) error {
	var envelope struct {
		Data TodoDeletedEventPayload `json:"data"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		if err := json.Unmarshal(data, &envelope.Data); err != nil {
			return fmt.Errorf("failed to unmarshal TodoDeleted event: %w", err)
		}
	}

	todoID := todoDomain.TodoID(envelope.Data.ID)

	schedules, err := h.repo.FindSchedulesByTodoID(ctx, todoID)
	if err != nil {
		return err
	}

	for _, s := range schedules {
		s.RemoveTask(todoID)

		if err := h.repo.Save(ctx, s); err != nil {
			return err
		}
	}

	return nil
}
