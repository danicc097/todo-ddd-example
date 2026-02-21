package decorator

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type tagRepositoryCache struct {
	base  domain.TagRepository
	rdb   *redis.Client
	ttl   time.Duration
	codec cache.Codec[*domain.Tag]
}

func NewTagRepositoryCache(
	base domain.TagRepository,
	rdb *redis.Client,
	ttl time.Duration,
	codec cache.Codec[*domain.Tag],
) domain.TagRepository {
	return &tagRepositoryCache{
		base:  base,
		rdb:   rdb,
		ttl:   ttl,
		codec: codec,
	}
}

func (r *tagRepositoryCache) Save(ctx context.Context, tag *domain.Tag) error {
	if err := r.base.Save(ctx, tag); err != nil {
		return err
	}

	db.AfterCommit(ctx, func(ctx context.Context) {
		r.rdb.Del(ctx, cache.Keys.Tag(tag.ID()))
		_ = cache.InvalidateTag(ctx, r.rdb, cache.Keys.WorkspaceTag(tag.WorkspaceID()))
	})

	return nil
}

func (r *tagRepositoryCache) FindByID(ctx context.Context, id domain.TagID) (*domain.Tag, error) {
	key := cache.Keys.Tag(id)

	return cache.GetOrFetch(ctx, r.rdb, key, r.ttl, r.codec, func(ctx context.Context) (*domain.Tag, error) {
		return r.base.FindByID(ctx, id)
	})
}

func (r *tagRepositoryCache) FindByName(ctx context.Context, workspaceID wsDomain.WorkspaceID, name string) (*domain.Tag, error) {
	return r.base.FindByName(ctx, workspaceID, name)
}
