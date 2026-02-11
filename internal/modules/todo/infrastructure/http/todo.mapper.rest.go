package http

import (
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TodoRestMapper struct{}

func (m *TodoRestMapper) ToResponse(t *domain.Todo) api.Todo {
	return api.Todo{
		Id:        t.ID(),
		Title:     t.Title().String(),
		Status:    api.TodoStatus(t.Status()),
		CreatedAt: t.CreatedAt(),
	}
}
