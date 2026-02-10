package domain

import (
	"context"

	"github.com/google/uuid"
)

//counterfeiter:generate . TodoRepository
//go:generate go tool gowrap gen -g -i TodoRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/todo_repository_tracing.gen.go
type TodoRepository interface {
	Save(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	FindAll(ctx context.Context) ([]*Todo, error)
	Update(ctx context.Context, todo *Todo) error
}

//counterfeiter:generate . TagRepository
//go:generate go tool gowrap gen -g -i TagRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/tag_repository_tracing.gen.go
type TagRepository interface {
	Save(ctx context.Context, tag *Tag) error
	FindByID(ctx context.Context, id uuid.UUID) (*Tag, error)
	FindByName(ctx context.Context, name string) (*Tag, error)
}
