package application

import (
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type ScheduleUseCases struct {
	CommitTask application.RequestHandler[CommitTaskCommand, CommitTaskResponse]
}
