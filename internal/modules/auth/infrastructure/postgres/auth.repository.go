package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/application"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type AuthRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
	uow  application.UnitOfWork
}

func NewAuthRepo(pool *pgxpool.Pool, uow application.UnitOfWork) *AuthRepo {
	return &AuthRepo{q: db.New(), pool: pool, uow: uow}
}

func (r *AuthRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *AuthRepo) FindByUserID(ctx context.Context, userID userDomain.UserID) (*domain.UserAuth, error) {
	row, err := r.q.GetUserAuth(ctx, r.getDB(ctx), userID.UUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAuthNotFound
		}

		return nil, fmt.Errorf("failed to get auth for user %s: %w", userID, sharedPg.ParseDBError(err))
	}

	passhash := ""
	if row.PasswordHash != nil {
		passhash = *row.PasswordHash
	}

	return domain.ReconstituteUserAuth(domain.ReconstituteUserAuthArgs{
		ID:           userID,
		Status:       row.TotpStatus,
		Cipher:       row.TotpSecretCipher,
		Nonce:        row.TotpSecretNonce,
		PasswordHash: passhash,
	}), nil
}

func (r *AuthRepo) Save(ctx context.Context, auth *domain.UserAuth) error {
	cipher, nonce := auth.TOTPCredentials()
	pass := auth.PasswordHash()

	err := r.q.UpsertUserAuth(ctx, r.getDB(ctx), db.UpsertUserAuthParams{
		UserID:           auth.UserID().UUID(),
		TotpStatus:       auth.TOTPStatus(),
		TotpSecretCipher: cipher,
		TotpSecretNonce:  nonce,
		PasswordHash:     &pass,
	})
	if err != nil {
		return fmt.Errorf("failed to save auth for user %s: %w", auth.UserID(), sharedPg.ParseDBError(err))
	}

	r.uow.Collect(ctx, nil, auth)

	return nil
}
