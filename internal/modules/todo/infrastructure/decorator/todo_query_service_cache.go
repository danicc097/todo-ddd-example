package decorator

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type todoQueryServiceCache struct {
	base  application.TodoQueryService
	rdb   *redis.Client
	ttl   time.Duration
	codec cache.Codec[*application.TodoReadModel]
}

func NewTodoQueryServiceCache(
	base application.TodoQueryService,
	rdb *redis.Client,
	ttl time.Duration,
	codec cache.Codec[*application.TodoReadModel],
) application.TodoQueryService {
	return &todoQueryServiceCache{
		base:  base,
		rdb:   rdb,
		ttl:   ttl,
		codec: codec,
	}
}

func (s *todoQueryServiceCache) GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID, limit, offset int32) ([]application.TodoReadModel, error) {
	key := fmt.Sprintf("%s:limit:%d:offset:%d", cache.Keys.TodoWorkspaceCollection(wsID), limit, offset)
	tag := cache.Keys.WorkspaceTag(wsID)

	return cache.GetOrFetch(ctx, s.rdb, key, s.ttl, cache.NewCollectionCodec[application.TodoReadModel](), func(ctx context.Context) ([]application.TodoReadModel, error) {
		return s.base.GetAllByWorkspace(ctx, wsID, limit, offset)
	}, tag)
}

func (s *todoQueryServiceCache) GetByID(ctx context.Context, id domain.TodoID) (*application.TodoReadModel, error) {
	key := cache.Keys.Todo(id)

	return cache.GetOrFetch(ctx, s.rdb, key, s.ttl, s.codec, func(ctx context.Context) (*application.TodoReadModel, error) {
		return s.base.GetByID(ctx, id)
	})
}
