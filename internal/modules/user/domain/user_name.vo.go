package domain

import (
	"strings"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrUserNameEmpty   = shared.NewDomainError(apperrors.InvalidInput, "user name cannot be empty")
	ErrUserNameTooLong = shared.NewDomainError(apperrors.InvalidInput, "user name is too long")
)

const userNameMaxLen = 70

type UserName struct {
	value string
}

func NewUserName(val string) (UserName, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return UserName{}, ErrUserNameEmpty
	}

	if len(val) > userNameMaxLen {
		return UserName{}, ErrUserNameTooLong
	}

	return UserName{value: val}, nil
}

func (n UserName) String() string {
	return n.value
}
