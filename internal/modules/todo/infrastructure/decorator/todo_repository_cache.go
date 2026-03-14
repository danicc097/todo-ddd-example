package decorator

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type todoRepositoryCache struct {
	base  domain.TodoRepository
	store cache.Store
	ttl   time.Duration
	codec cache.Codec[*domain.Todo]
}

func NewTodoRepositoryCache(
	base domain.TodoRepository,
	store cache.Store,
	ttl time.Duration,
	codec cache.Codec[*domain.Todo],
) domain.TodoRepository {
	return &todoRepositoryCache{
		base:  base,
		store: store,
		ttl:   ttl,
		codec: codec,
	}
}

func (r *todoRepositoryCache) Save(ctx context.Context, todo *domain.Todo) error {
	if err := r.base.Save(ctx, todo); err != nil {
		return err
	}

	db.AfterCommit(ctx, func(ctx context.Context) {
		_ = r.store.Delete(ctx, cache.Keys.TodoAggregate(todo.ID()), cache.Keys.TodoReadModel(todo.ID()))
		_ = r.store.Invalidate(ctx, cache.Keys.WorkspaceTag(todo.WorkspaceID()))
		_, _ = r.store.Incr(ctx, cache.Keys.WorkspaceRevision(todo.WorkspaceID()))
	})

	return nil
}

func (r *todoRepositoryCache) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
	key := cache.Keys.TodoAggregate(id)

	return cache.GetOrFetch(ctx, r.store, key, r.ttl, r.codec, func(ctx context.Context) (*domain.Todo, error) {
		return r.base.FindByID(ctx, id)
	})
}

func (r *todoRepositoryCache) FindAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID) ([]*domain.Todo, error) {
	revisionBytes, _ := r.store.Get(ctx, cache.Keys.WorkspaceRevision(wsID))

	revision := string(revisionBytes)
	if revision == "" {
		revision = "0"
	}

	key := cache.Keys.TodoWorkspaceCollection(wsID, revision)
	tag := cache.Keys.WorkspaceTag(wsID)

	return cache.GetOrFetch(ctx, r.store, key, r.ttl, cache.NewCollectionCodec[*domain.Todo](), func(ctx context.Context) ([]*domain.Todo, error) {
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
		_ = r.store.Delete(ctx, cache.Keys.TodoAggregate(id), cache.Keys.TodoReadModel(id))
		_ = r.store.Invalidate(ctx, cache.Keys.WorkspaceTag(todo.WorkspaceID()))
		_, _ = r.store.Incr(ctx, cache.Keys.WorkspaceRevision(todo.WorkspaceID()))
	})

	return nil
}
