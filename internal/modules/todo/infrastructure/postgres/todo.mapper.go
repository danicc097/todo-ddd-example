package postgres

import (
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/postgres/generated/db"
)

type TodoMapper struct{}

func (m *TodoMapper) ToDomain(row db.Todos) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)
	status, _ := domain.NewTodoStatus(row.Status)

	return domain.NewTodo(
		row.ID,
		title,
		status,
		row.CreatedAt,
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
