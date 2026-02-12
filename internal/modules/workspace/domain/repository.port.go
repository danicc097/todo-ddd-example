package domain

import (
	"context"

	"github.com/google/uuid"
)

//counterfeiter:generate . WorkspaceRepository
//go:generate go tool gowrap gen -g -i WorkspaceRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/workspace_repository_tracing.gen.go
//go:generate go tool gowrap gen -g -i WorkspaceRepository -t ../../../../templates/cache.gotmpl -o ../infrastructure/decorator/workspace_repository_cache.gen.go -v "KeyPrefix=ws" -v "EntityType=*_sourceDomain.Workspace"
type WorkspaceRepository interface {
	Save(ctx context.Context, w *Workspace) error
	FindByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	FindAll(ctx context.Context) ([]*Workspace, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
