package domain

import (
	"errors"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email address")

type UserEmail struct {
	value string
}

func NewUserEmail(val string) (UserEmail, error) {
	val = strings.ToLower(strings.TrimSpace(val))
	if !strings.Contains(val, "@") {
		return UserEmail{}, ErrInvalidEmail
	}
	return UserEmail{value: val}, nil
}

func (e UserEmail) String() string {
	return e.value
}
