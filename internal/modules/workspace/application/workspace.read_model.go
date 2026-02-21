package application

import (
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type WorkspaceReadModel struct {
	ID          domain.WorkspaceID
	Name        string
	Description string
}

type TagReadModel struct {
	ID   todoDomain.TagID
	Name string
}
