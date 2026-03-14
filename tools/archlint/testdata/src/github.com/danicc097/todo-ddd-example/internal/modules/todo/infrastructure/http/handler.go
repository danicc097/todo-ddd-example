package http

import "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"

type TodoHandler struct {
	repo domain.TodoRepository
}

func (h *TodoHandler) Create() {
	h.repo.Save() // want "Arch violation: TodoHandler calls TodoRepository.Save directly. Handlers must route through Application UseCases."
}
