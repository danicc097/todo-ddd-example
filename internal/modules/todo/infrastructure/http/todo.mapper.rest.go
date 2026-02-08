package http

import "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"

type TodoRestMapper struct{}

func (m *TodoRestMapper) ToResponse(t *domain.Todo) todoResponse {
	return todoResponse{
		ID:        t.ID(),
		Title:     t.Title().String(),
		Status:    t.Status().String(),
		CreatedAt: t.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (m *TodoRestMapper) ToResponseList(todos []*domain.Todo) []todoResponse {
	res := make([]todoResponse, len(todos))
	for i, t := range todos {
		res[i] = m.ToResponse(t)
	}
	return res
}
