package domain

import (
	"context"
)

//go:generate go tool gowrap gen -g -i WorkspaceRepository -t ../../../../templates/opentelemetry.gotmpl -o ../infrastructure/postgres/workspace_repository_tracing.gen.go
type WorkspaceRepository interface {
	Save(ctx context.Context, w *Workspace) error
	FindByID(ctx context.Context, id WorkspaceID) (*Workspace, error)
	Delete(ctx context.Context, id WorkspaceID) error
}
