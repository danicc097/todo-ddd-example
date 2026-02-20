package decorator

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type todoQueryServiceCache struct {
	base  application.TodoQueryService
	rdb   *redis.Client
	ttl   time.Duration
	codec cache.Codec[*api.Todo]
}

func NewTodoQueryServiceCache(
	base application.TodoQueryService,
	rdb *redis.Client,
	ttl time.Duration,
	codec cache.Codec[*api.Todo],
) application.TodoQueryService {
	return &todoQueryServiceCache{
		base:  base,
		rdb:   rdb,
		ttl:   ttl,
		codec: codec,
	}
}

func (s *todoQueryServiceCache) GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]api.Todo, error) {
	key := cache.Keys.TodoWorkspaceCollection(wsID)
	tag := cache.Keys.WorkspaceTag(wsID)

	return cache.GetOrFetch(ctx, s.rdb, key, s.ttl, cache.NewCollectionCodec[api.Todo](), func(ctx context.Context) ([]api.Todo, error) {
		return s.base.GetAllByWorkspace(ctx, wsID)
	}, tag)
}

func (s *todoQueryServiceCache) GetByID(ctx context.Context, id domain.TodoID) (*api.Todo, error) {
	key := cache.Keys.Todo(id)

	return cache.GetOrFetch(ctx, s.rdb, key, s.ttl, s.codec, func(ctx context.Context) (*api.Todo, error) {
		return s.base.GetByID(ctx, id)
	})
}
