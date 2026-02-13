package rabbitmq

import (
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TodoEventDTO struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func ToTodoEventDTO(t *domain.Todo) TodoEventDTO {
	return TodoEventDTO{
		ID:        t.ID().UUID,
		Title:     t.Title().String(),
		Status:    t.Status().String(),
		CreatedAt: t.CreatedAt(),
	}
}

type TagAddedEventDTO struct {
	TodoID uuid.UUID `json:"todo_id"`
	TagID  uuid.UUID `json:"tag_id"`
}
