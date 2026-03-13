package application

import (
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
)

type TodoUseCases struct {
	CreateTodo application.RequestHandler[CreateTodoCommand, CreateTodoResponse]
	Complete   application.RequestHandler[CompleteTodoCommand, CompleteTodoResponse]
	CreateTag  application.RequestHandler[CreateTagCommand, CreateTagResponse]
	AssignTag  application.RequestHandler[AssignTagToTodoCommand, AssignTagToTodoResponse]
	StartFocus application.RequestHandler[StartFocusCommand, StartFocusResponse]
	StopFocus  application.RequestHandler[StopFocusCommand, StopFocusResponse]
}
