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

type todoRepositoryCache struct {
	base  domain.TodoRepository
	rdb   *redis.Client
	ttl   time.Duration
	codec cache.Codec[*domain.Todo]
}

func NewTodoRepositoryCache(
	base domain.TodoRepository,
	rdb *redis.Client,
	ttl time.Duration,
	codec cache.Codec[*domain.Todo],
) domain.TodoRepository {
	return &todoRepositoryCache{
		base:  base,
		rdb:   rdb,
		ttl:   ttl,
		codec: codec,
	}
}

func (r *todoRepositoryCache) Save(ctx context.Context, todo *domain.Todo) error {
	if err := r.base.Save(ctx, todo); err != nil {
		return err
	}

	db.AfterCommit(ctx, func(ctx context.Context) {
		r.rdb.Del(ctx, cache.Keys.TodoAggregate(todo.ID()), cache.Keys.TodoReadModel(todo.ID()))
		_ = cache.InvalidateTag(ctx, r.rdb, cache.Keys.WorkspaceTag(todo.WorkspaceID()))
		r.rdb.Incr(ctx, cache.Keys.WorkspaceRevision(todo.WorkspaceID()))
	})

	return nil
}

func (r *todoRepositoryCache) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
	key := cache.Keys.TodoAggregate(id)

	return cache.GetOrFetch(ctx, r.rdb, key, r.ttl, r.codec, func(ctx context.Context) (*domain.Todo, error) {
		return r.base.FindByID(ctx, id)
	})
}

func (r *todoRepositoryCache) FindAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]*domain.Todo, error) {
	revision, _ := r.rdb.Get(ctx, cache.Keys.WorkspaceRevision(wsID)).Result()
	if revision == "" {
		revision = "0"
	}

	key := cache.Keys.TodoWorkspaceCollection(wsID, revision)
	tag := cache.Keys.WorkspaceTag(wsID)

	return cache.GetOrFetch(ctx, r.rdb, key, r.ttl, cache.NewCollectionCodec[*domain.Todo](), func(ctx context.Context) ([]*domain.Todo, error) {
		return r.base.FindAllByWorkspace(ctx, wsID)
	}, tag)
}

func (r *todoRepositoryCache) Delete(ctx context.Context, id domain.TodoID) error {
	todo, err := r.base.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.base.Delete(ctx, id); err != nil {
		return err
	}

	db.AfterCommit(ctx, func(ctx context.Context) {
		r.rdb.Del(ctx, cache.Keys.TodoAggregate(id), cache.Keys.TodoReadModel(id))
		_ = cache.InvalidateTag(ctx, r.rdb, cache.Keys.WorkspaceTag(todo.WorkspaceID()))
		r.rdb.Incr(ctx, cache.Keys.WorkspaceRevision(todo.WorkspaceID()))
	})

	return nil
}
