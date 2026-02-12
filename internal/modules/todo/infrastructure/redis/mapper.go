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
	return TodoCacheDTO{
		ID:        t.ID(),
		Title:     t.Title().String(),
		Status:    t.Status().String(),
		CreatedAt: t.CreatedAt(),
		Tags:      t.Tags(),
	}
}

func FromTodoCacheDTO(dto TodoCacheDTO) *domain.Todo {
	title, _ := domain.NewTodoTitle(dto.Title)

	return domain.ReconstituteTodo(
		dto.ID,
		title,
		domain.TodoStatus(dto.Status),
		dto.CreatedAt,
		dto.Tags,
	)
}

type TagCacheDTO struct {
	ID   uuid.UUID `msgpack:"id"`
	Name string    `msgpack:"name"`
}

func ToTagCacheDTO(t *domain.Tag) TagCacheDTO {
	return TagCacheDTO{
		ID:   t.ID(),
		Name: t.Name().String(),
	}
}

func FromTagCacheDTO(dto TagCacheDTO) *domain.Tag {
	name, _ := domain.NewTagName(dto.Name)
	return domain.ReconstituteTag(dto.ID, name)
}
