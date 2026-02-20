package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

/**
 * Todo
 */

type TodoMapper struct{}

func (m *TodoMapper) ToDomain(row db.GetTodoByIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)

	tagIDs := make([]domain.TagID, len(row.Tags))
	for i, id := range row.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	return domain.ReconstituteTodo(
		row.ID,
		title,
		domain.TodoStatus(row.Status),
		row.CreatedAt,
		tagIDs,
		row.WorkspaceID,
	)
}

func (m *TodoMapper) ListRowToDomain(row db.ListTodosByWorkspaceIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)

	tagIDs := make([]domain.TagID, len(row.Tags))
	for i, id := range row.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	return domain.ReconstituteTodo(
		row.ID,
		title,
		domain.TodoStatus(row.Status),
		row.CreatedAt,
		tagIDs,
		row.WorkspaceID,
	)
}

// ToPersistence maps Domain to the primary table struct.
func (m *TodoMapper) ToPersistence(t *domain.Todo) db.Todos {
	return db.Todos{
		ID:          t.ID(),
		Title:       t.Title().String(),
		Status:      t.Status().String(),
		CreatedAt:   t.CreatedAt(),
		WorkspaceID: t.WorkspaceID(),
	}
}

/**
 * Tags
 */

type TagMapper struct{}

func (m *TagMapper) ToDomain(row db.Tags) *domain.Tag {
	name, _ := domain.NewTagName(row.Name)
	return domain.ReconstituteTag(row.ID, name, row.WorkspaceID)
}

// ToPersistence maps Domain to the primary table struct.
func (m *TagMapper) ToPersistence(t *domain.Tag) db.Tags {
	return db.Tags{
		ID:          t.ID(),
		Name:        t.Name().String(),
		WorkspaceID: t.WorkspaceID(),
	}
}

/**
 * Events
 */

type todoOutboxDTO struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type tagAddedOutboxDTO struct {
	TodoID      uuid.UUID `json:"todo_id"`
	TagID       uuid.UUID `json:"tag_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

func (m *TodoMapper) MapEvent(e shared.DomainEvent) (shared.EventType, []byte, error) {
	var payload any

	switch evt := e.(type) {
	case domain.TodoCreatedEvent:
		payload = todoOutboxDTO{
			ID:          evt.ID.UUID(),
			WorkspaceID: evt.WorkspaceID.UUID(),
			Title:       evt.Title,
			Status:      evt.Status,
			CreatedAt:   evt.CreatedAt,
		}
	case domain.TodoCompletedEvent:
		payload = todoOutboxDTO{
			ID:          evt.ID.UUID(),
			WorkspaceID: evt.WorkspaceID.UUID(),
			Title:       evt.Title,
			Status:      evt.Status,
			CreatedAt:   evt.CreatedAt,
		}
	case domain.TagAddedEvent:
		payload = tagAddedOutboxDTO{
			TodoID:      evt.TodoID.UUID(),
			TagID:       evt.TagID.UUID(),
			WorkspaceID: evt.WorkspaceID.UUID(),
		}
	default:
		return "", nil, nil
	}

	b, err := json.Marshal(payload)

	return e.EventName(), b, err
}
