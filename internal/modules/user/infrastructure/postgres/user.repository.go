package postgres

import (
	"context"
	"errors"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *UserMapper
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: db.New(), pool: pool, mapper: &UserMapper{}}
}

func (r *UserRepo) Save(ctx context.Context, u *domain.User) error {
	p := r.mapper.ToPersistence(u)
	_, err := r.q.CreateUser(ctx, r.pool, db.CreateUserParams(p))
	return err
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, r.pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return r.mapper.ToDomain(row), nil
}
