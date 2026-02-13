package redis

import (
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TodoCacheDTO struct {
	ID        uuid.UUID   `msgpack:"id"`
	Title     string      `msgpack:"title"`
	Status    string      `msgpack:"status"`
	CreatedAt time.Time   `msgpack:"created_at"`
	Tags      []uuid.UUID `msgpack:"tags"`
}

func ToTodoCacheDTO(t *domain.Todo) TodoCacheDTO {
	tagUUIDs := make([]uuid.UUID, len(t.Tags()))
	for i, id := range t.Tags() {
		tagUUIDs[i] = id.UUID
	}

	return TodoCacheDTO{
		ID:        t.ID().UUID,
		Title:     t.Title().String(),
		Status:    t.Status().String(),
		CreatedAt: t.CreatedAt(),
		Tags:      tagUUIDs,
	}
}

func FromTodoCacheDTO(dto TodoCacheDTO) *domain.Todo {
	title, _ := domain.NewTodoTitle(dto.Title)

	tagIDs := make([]domain.TagID, len(dto.Tags))
	for i, id := range dto.Tags {
		tagIDs[i] = domain.TagID{UUID: id}
	}

	return domain.ReconstituteTodo(
		domain.TodoID{UUID: dto.ID},
		title,
		domain.TodoStatus(dto.Status),
		dto.CreatedAt,
		tagIDs,
	)
}

type TagCacheDTO struct {
	ID   uuid.UUID `msgpack:"id"`
	Name string    `msgpack:"name"`
}

func ToTagCacheDTO(t *domain.Tag) TagCacheDTO {
	return TagCacheDTO{
		ID:   t.ID().UUID,
		Name: t.Name().String(),
	}
}

func FromTagCacheDTO(dto TagCacheDTO) *domain.Tag {
	name, _ := domain.NewTagName(dto.Name)
	return domain.ReconstituteTag(domain.TagID{UUID: dto.ID}, name)
}
