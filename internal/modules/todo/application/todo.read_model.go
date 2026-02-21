package application

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type TodoReadModel struct {
	ID          domain.TodoID
	WorkspaceID wsDomain.WorkspaceID
	Title       string
	Status      string
	CreatedAt   time.Time
}
