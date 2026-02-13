package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type UserRepo struct {
	q      *db.Queries
	db     db.DBTX
	mapper *UserMapper
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		q:      db.New(),
		db:     pool,
		mapper: &UserMapper{},
	}
}

func NewUserRepoFromTx(tx pgx.Tx) *UserRepo {
	return &UserRepo{
		q:      db.New(),
		db:     tx,
		mapper: &UserMapper{},
	}
}

func (r *UserRepo) Save(ctx context.Context, u *domain.User) error {
	p := r.mapper.ToPersistence(u)
	_, err := r.q.CreateUser(ctx, r.db, db.CreateUserParams(p))

	return err
}

func (r *UserRepo) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return r.mapper.ToDomain(row), nil
}
