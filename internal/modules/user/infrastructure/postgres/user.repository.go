package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type UserRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *UserMapper
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		q:      db.New(),
		pool:   pool,
		mapper: &UserMapper{},
	}
}

func (r *UserRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *UserRepo) Save(ctx context.Context, u *domain.User) error {
	dbtx := r.getDB(ctx)
	p := r.mapper.ToPersistence(u)

	_, err := r.q.CreateUser(ctx, dbtx, db.CreateUserParams(p))
	if err != nil {
		return fmt.Errorf("failed to save user %s: %w", u.ID(), sharedPg.ParseDBError(err))
	}

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, u)
}

func (r *UserRepo) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, r.getDB(ctx), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user %s: %w", id, sharedPg.ParseDBError(err))
	}

	return r.mapper.ToDomain(row), nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email domain.UserEmail) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, r.getDB(ctx), email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email %s: %w", email, sharedPg.ParseDBError(err))
	}

	return r.mapper.ToDomain(row), nil
}

func (r *UserRepo) Delete(ctx context.Context, id domain.UserID) error {
	dbtx := r.getDB(ctx)

	u, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.q.DeleteUser(ctx, dbtx, id); err != nil {
		return fmt.Errorf("failed to delete user %s: %w", id, sharedPg.ParseDBError(err))
	}

	u.Delete()

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, u)
}
