package postgres

import (
	"encoding/json"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/google/uuid"
)

type TodoMapper struct{}

func (m *TodoMapper) ToDomain(row db.GetTodoByIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)
	return domain.ReconstituteTodo(
		row.ID,
		title,
		domain.TodoStatus(row.Status),
		row.CreatedAt,
		row.Tags,
	)
}

func (m *TodoMapper) ListRowToDomain(row db.ListTodosRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)
	return domain.ReconstituteTodo(
		row.ID,
		title,
		domain.TodoStatus(row.Status),
		row.CreatedAt,
		row.Tags,
	)
}

func (m *TodoMapper) ToPersistence(t *domain.Todo) db.Todos {
	return db.Todos{
		ID:        t.ID(),
		Title:     t.Title().String(),
		Status:    t.Status().String(),
		CreatedAt: t.CreatedAt(),
	}
}

type TodoEventDTO struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type TagAddedEventDTO struct {
	TodoID uuid.UUID `json:"todo_id"`
	TagID  uuid.UUID `json:"tag_id"`
}

func (m *TodoMapper) MapEvent(e domain.DomainEvent) (string, []byte, error) {
	var payload any

	switch evt := e.(type) {
	case domain.TodoCreatedEvent:
		payload = TodoEventDTO{
			ID:        evt.ID,
			Title:     evt.Title,
			Status:    evt.Status,
			CreatedAt: evt.CreatedAt,
		}
	case domain.TodoCompletedEvent:
		payload = TodoEventDTO{
			ID:        evt.ID,
			Title:     evt.Title,
			Status:    evt.Status,
			CreatedAt: evt.CreatedAt,
		}
	case domain.TagAddedEvent:
		payload = TagAddedEventDTO{
			TodoID: evt.TodoID,
			TagID:  evt.TagID,
		}
	default:
		return "", nil, nil
	}

	b, err := json.Marshal(payload)
	return e.EventName(), b, err
}
