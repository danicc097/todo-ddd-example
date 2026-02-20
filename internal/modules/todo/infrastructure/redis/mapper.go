package redis

import (
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type TodoCacheDTO struct {
	ID          uuid.UUID   `json:"id"`
	WorkspaceID uuid.UUID   `json:"workspace_id"`
	Title       string      `json:"title"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	Tags        []uuid.UUID `json:"tags"`
}

func ToTodoCacheDTO(t *domain.Todo) TodoCacheDTO {
	tagUUIDs := make([]uuid.UUID, len(t.Tags()))
	for i, id := range t.Tags() {
		tagUUIDs[i] = id.UUID()
	}

	return TodoCacheDTO{
		ID:          t.ID().UUID(),
		WorkspaceID: t.WorkspaceID().UUID(),
		Title:       t.Title().String(),
		Status:      t.Status().String(),
		CreatedAt:   t.CreatedAt(),
		Tags:        tagUUIDs,
	}
}

func FromTodoCacheDTO(dto TodoCacheDTO) *domain.Todo {
	title, _ := domain.NewTodoTitle(dto.Title)

	tagIDs := make([]domain.TagID, len(dto.Tags))
	for i, id := range dto.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	return domain.ReconstituteTodo(
		domain.TodoID(dto.ID),
		title,
		domain.TodoStatus(dto.Status),
		dto.CreatedAt,
		tagIDs,
		wsDomain.WorkspaceID(dto.WorkspaceID),
	)
}

type TagCacheDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

func ToTagCacheDTO(t *domain.Tag) TagCacheDTO {
	return TagCacheDTO{
		ID:          t.ID().UUID(),
		Name:        t.Name().String(),
		WorkspaceID: t.WorkspaceID().UUID(),
	}
}

func FromTagCacheDTO(dto TagCacheDTO) *domain.Tag {
	name, _ := domain.NewTagName(dto.Name)
	return domain.ReconstituteTag(domain.TagID(dto.ID), name, wsDomain.WorkspaceID(dto.WorkspaceID))
}
