package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type TagRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *TagMapper
}

func NewTagRepo(pool *pgxpool.Pool) *TagRepo {
	return &TagRepo{
		q:      db.New(),
		pool:   pool,
		mapper: &TagMapper{},
	}
}

func (r *TagRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *TagRepo) Save(ctx context.Context, t *domain.Tag) error {
	dbtx := r.getDB(ctx)
	p := r.mapper.ToPersistence(t)

	_, err := r.q.CreateTag(ctx, dbtx, db.CreateTagParams(p))
	if err != nil {
		return fmt.Errorf("failed to save tag %s: %w", t.ID(), sharedPg.ParseDBError(err))
	}

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, t)
}

func (r *TagRepo) FindByID(ctx context.Context, id domain.TagID) (*domain.Tag, error) {
	row, err := r.q.GetTagByID(ctx, r.getDB(ctx), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTagNotFound
		}

		return nil, fmt.Errorf("failed to get tag %s: %w", id, sharedPg.ParseDBError(err))
	}

	return r.mapper.ToDomain(row), nil
}

func (r *TagRepo) FindByName(ctx context.Context, workspaceID wsDomain.WorkspaceID, name string) (*domain.Tag, error) {
	row, err := r.q.GetTagByName(ctx, r.getDB(ctx), db.GetTagByNameParams{
		WorkspaceID: workspaceID,
		Name:        name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTagNotFound
		}

		return nil, fmt.Errorf("failed to get tag by name %s in workspace %s: %w", name, workspaceID, sharedPg.ParseDBError(err))
	}

	return r.mapper.ToDomain(row), nil
}

func (r *TagRepo) Delete(ctx context.Context, id domain.TagID) error {
	dbtx := r.getDB(ctx)

	if err := r.q.DeleteTag(ctx, dbtx, id); err != nil {
		return fmt.Errorf("failed to delete tag %s: %w", id, sharedPg.ParseDBError(err))
	}

	return nil
}
