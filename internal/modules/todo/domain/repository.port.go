package domain

import (
	"context"

	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

//counterfeiter:generate . TodoRepository
//go:generate go tool gowrap gen -g -i TodoRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/todo_repository_tracing.gen.go
//go:generate go tool gowrap gen -g -i TodoRepository -t ../../../../templates/cache.gotmpl -o ../infrastructure/decorator/todo_repository_cache.gen.go -v "KeyPrefix=todo" -v "EntityType=*_sourceDomain.Todo"
type TodoRepository interface {
	Save(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id TodoID) (*Todo, error)
	FindAll(ctx context.Context) ([]*Todo, error)
	Update(ctx context.Context, todo *Todo) error
}

//counterfeiter:generate . TagRepository
//go:generate go tool gowrap gen -g -i TagRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/tag_repository_tracing.gen.go
//go:generate go tool gowrap gen -g -i TagRepository -t ../../../../templates/cache.gotmpl -o ../infrastructure/decorator/tag_repository_cache.gen.go -v "KeyPrefix=tag" -v "EntityType=*_sourceDomain.Tag"
type TagRepository interface {
	Save(ctx context.Context, tag *Tag) error
	FindByID(ctx context.Context, id TagID) (*Tag, error)
	FindByName(ctx context.Context, workspaceID wsDomain.WorkspaceID, name string) (*Tag, error)
}
