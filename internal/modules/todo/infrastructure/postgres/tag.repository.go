package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
)

type TagRepo struct {
	q      *db.Queries
	db     db.DBTX
	mapper *TagMapper
}

func NewTagRepo(pool *pgxpool.Pool) *TagRepo {
	return &TagRepo{
		q:      db.New(),
		db:     pool,
		mapper: &TagMapper{},
	}
}

func NewTagRepoFromTx(tx pgx.Tx) *TagRepo {
	return &TagRepo{
		q:      db.New(),
		db:     tx,
		mapper: &TagMapper{},
	}
}

func (r *TagRepo) Save(ctx context.Context, t *domain.Tag) error {
	p := r.mapper.ToPersistence(t)

	_, err := r.q.CreateTag(ctx, r.db, db.CreateTagParams(p))

	return err
}

func (r *TagRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	row, err := r.q.GetTagByID(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTagNotFound
		}

		return nil, err
	}

	return r.mapper.ToDomain(row), nil
}

func (r *TagRepo) FindByName(ctx context.Context, name string) (*domain.Tag, error) {
	row, err := r.q.GetTagByName(ctx, r.db, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTagNotFound
		}

		return nil, err
	}

	return r.mapper.ToDomain(row), nil
}
