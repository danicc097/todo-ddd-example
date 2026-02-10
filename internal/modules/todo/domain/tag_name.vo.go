package domain

import (
	"errors"
	"strings"
)

var (
	ErrTagNameEmpty   = errors.New("tag name cannot be empty")
	ErrTagNameTooLong = errors.New("tag name is too long")
)

type TagName struct {
	value string
}

func NewTagName(val string) (TagName, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return TagName{}, ErrTagNameEmpty
	}

	if len(val) > 20 {
		return TagName{}, ErrTagNameTooLong
	}

	return TagName{value: val}, nil
}

func (t TagName) String() string {
	return t.value
}
