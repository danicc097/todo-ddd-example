package decorator

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type workspaceRepositoryCache struct {
	base  domain.WorkspaceRepository
	rdb   *redis.Client
	ttl   time.Duration
	codec cache.Codec[*domain.Workspace]
}

var _ domain.WorkspaceRepository = (*workspaceRepositoryCache)(nil)

func NewWorkspaceRepositoryCache(
	base domain.WorkspaceRepository,
	rdb *redis.Client,
	ttl time.Duration,
	codec cache.Codec[*domain.Workspace],
) domain.WorkspaceRepository {
	return &workspaceRepositoryCache{
		base:  base,
		rdb:   rdb,
		ttl:   ttl,
		codec: codec,
	}
}

func (r *workspaceRepositoryCache) Save(ctx context.Context, w *domain.Workspace) error {
	if err := r.base.Save(ctx, w); err != nil {
		return err
	}

	r.rdb.Del(ctx, cache.Keys.Workspace(w.ID()), cache.Keys.AllWorkspaces())
	r.rdb.Incr(ctx, cache.Keys.WorkspaceRevision(w.ID()))

	return nil
}

func (r *workspaceRepositoryCache) FindByID(ctx context.Context, id domain.WorkspaceID) (*domain.Workspace, error) {
	key := cache.Keys.Workspace(id)

	return cache.GetOrFetch(ctx, r.rdb, key, r.ttl, r.codec, func(ctx context.Context) (*domain.Workspace, error) {
		return r.base.FindByID(ctx, id)
	})
}

func (r *workspaceRepositoryCache) FindAll(ctx context.Context) ([]*domain.Workspace, error) {
	key := cache.Keys.AllWorkspaces()

	return cache.GetOrFetch(ctx, r.rdb, key, r.ttl, cache.NewCollectionCodec[*domain.Workspace](), func(ctx context.Context) ([]*domain.Workspace, error) {
		return r.base.FindAll(ctx)
	})
}

func (r *workspaceRepositoryCache) Delete(ctx context.Context, id domain.WorkspaceID) error {
	if err := r.base.Delete(ctx, id); err != nil {
		return err
	}

	r.rdb.Del(ctx, cache.Keys.Workspace(id), cache.Keys.AllWorkspaces())
	r.rdb.Incr(ctx, cache.Keys.WorkspaceRevision(id))

	return nil
}
