package application

import (
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type FocusSessionReadModel struct {
	ID        uuid.UUID
	StartTime time.Time
	EndTime   *time.Time
}

type TodoReadModel struct {
	ID                 domain.TodoID
	WorkspaceID        wsDomain.WorkspaceID
	Title              string
	Status             string
	CreatedAt          time.Time
	DueDate            *time.Time
	RecurrenceInterval *string
	RecurrenceAmount   *int
	FocusSessions      []FocusSessionReadModel
}
