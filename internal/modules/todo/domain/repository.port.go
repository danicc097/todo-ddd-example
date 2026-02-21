package domain

import (
	"context"

	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

//go:generate go tool gowrap gen -g -i TodoRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/todo_repository_tracing.gen.go
type TodoRepository interface {
	Save(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id TodoID) (*Todo, error)
	FindAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]*Todo, error)
	Delete(ctx context.Context, id TodoID) error
}

//go:generate go tool gowrap gen -g -i TagRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/tag_repository_tracing.gen.go
type TagRepository interface {
	Save(ctx context.Context, tag *Tag) error
	FindByID(ctx context.Context, id TagID) (*Tag, error)
	FindByName(ctx context.Context, workspaceID wsDomain.WorkspaceID, name string) (*Tag, error)
	Delete(ctx context.Context, id TagID) error
}
