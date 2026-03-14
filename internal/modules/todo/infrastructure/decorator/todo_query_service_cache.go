package decorator

import (
	"context"
	"time"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type todoQueryServiceCache struct {
	base  application.TodoQueryService
	store cache.Store
	ttl   time.Duration
	codec cache.Codec[*application.TodoReadModel]
}

func NewTodoQueryServiceCache(
	base application.TodoQueryService,
	store cache.Store,
	ttl time.Duration,
	codec cache.Codec[*application.TodoReadModel],
) application.TodoQueryService {
	return &todoQueryServiceCache{
		base:  base,
		store: store,
		ttl:   ttl,
		codec: codec,
	}
}

func (s *todoQueryServiceCache) GetAllByWorkspace(ctx context.Context, wsID wsDomain.WorkspaceID, limit, offset int32) ([]application.TodoReadModel, error) {
	revisionBytes, _ := s.store.Get(ctx, cache.Keys.WorkspaceRevision(wsID))

	revision := string(revisionBytes)
	if revision == "" {
		revision = "0"
	}

	key := cache.Keys.TodoWorkspaceCollectionPaginated(wsID, limit, offset, revision)
	tag := cache.Keys.WorkspaceTag(wsID)

	return cache.GetOrFetch(ctx, s.store, key, s.ttl, cache.NewCollectionCodec[application.TodoReadModel](), func(ctx context.Context) ([]application.TodoReadModel, error) {
		return s.base.GetAllByWorkspace(ctx, wsID, limit, offset)
	}, tag)
}

func (s *todoQueryServiceCache) GetByID(ctx context.Context, id domain.TodoID) (*application.TodoReadModel, error) {
	key := cache.Keys.TodoReadModel(id)

	return cache.GetOrFetch(ctx, s.store, key, s.ttl, s.codec, func(ctx context.Context) (*application.TodoReadModel, error) {
		return s.base.GetByID(ctx, id)
	})
}
