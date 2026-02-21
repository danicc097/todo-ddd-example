package postgres

import (
	"errors"
	"regexp"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var errorUniqueViolationRegex = regexp.MustCompile(`\((.*)\)=\((.*)\)`)

// ParseDBError adapts postgres-specific errors to domain or application errors.
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		// fallback if repo doesn't map to specific entity not found
		return sharedDomain.NewDomainError(apperrors.NotFound, "resource not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return sharedDomain.WrapDomainError(err, apperrors.InvalidInput)

		case pgerrcode.UniqueViolation:
			return sharedDomain.WrapDomainError(err, apperrors.Conflict)
		}

		return apperrors.Wrap(err, apperrors.Internal, "database error occurred")
	}

	return err
}
