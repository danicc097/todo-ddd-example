package types

import (
	todo "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	user "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	workspace "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type (
	TodoID      = todo.TodoID
	TagID       = todo.TagID
	UserID      = user.UserID
	WorkspaceID = workspace.WorkspaceID
)
