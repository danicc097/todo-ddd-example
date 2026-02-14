package postgres

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

var errorUniqueViolationRegex = regexp.MustCompile(`\((.*)\)=\((.*)\)`)

// ParseDBError adapts postgres-specific errors to domain or application errors.
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		// fallback if repo doesn't map to specific entity not found
		return apperrors.New(apperrors.NotFound, "resource not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			switch pgErr.ConstraintName {
			case "fk_wm_user":
				return wsDomain.ErrMemberNotFound
			case "fk_wm_workspace":
				return wsDomain.ErrWorkspaceNotFound
			}

			return apperrors.New(apperrors.InvalidInput, "invalid reference for "+pgErr.TableName)

		case pgerrcode.UniqueViolation:
			switch pgErr.ConstraintName {
			case "workspace_members_pkey":
				return wsDomain.ErrUserAlreadyMember
			}

			matches := errorUniqueViolationRegex.FindStringSubmatch(pgErr.Detail)
			if len(matches) > 2 {
				msg := fmt.Sprintf("%s %q already exists", matches[1], matches[2]) // uses column name atm
				return apperrors.New(apperrors.Conflict, msg)
			}

			return apperrors.New(apperrors.Conflict, "resource already exists")
		}

		return apperrors.Wrap(err, apperrors.Internal, "database error occurred")
	}

	return err
}
