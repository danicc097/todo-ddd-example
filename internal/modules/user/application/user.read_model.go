package application

import (
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type UserReadModel struct {
	ID    domain.UserID
	Email string
	Name  string
}
