package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/auth/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type AuthRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{q: db.New(), pool: pool}
}

func (r *AuthRepo) FindByUserID(ctx context.Context, userID userDomain.UserID) (*domain.UserAuth, error) {
	row, err := r.q.GetUserAuth(ctx, r.pool, userID.UUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAuthNotFound
		}

		return nil, err
	}

	passhash := ""
	if row.PasswordHash != nil {
		passhash = *row.PasswordHash
	}

	return domain.ReconstituteUserAuth(userID,
		row.TotpStatus,
		row.TotpSecretCipher,
		row.TotpSecretNonce,
		passhash,
	), nil
}

func (r *AuthRepo) Save(ctx context.Context, auth *domain.UserAuth) error {
	cipher, nonce := auth.TOTPCredentials()
	pass := auth.PasswordHash()

	return r.q.UpsertUserAuth(ctx, r.pool, db.UpsertUserAuthParams{
		UserID:           auth.UserID().UUID(),
		TotpStatus:       auth.TOTPStatus(),
		TotpSecretCipher: cipher,
		TotpSecretNonce:  nonce,
		PasswordHash:     &pass,
	})
}
