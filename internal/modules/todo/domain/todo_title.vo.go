package domain

import (
	"errors"
	"strings"
)

var (
	ErrTitleEmpty   = errors.New("title cannot be empty")
	ErrTitleTooLong = errors.New("title is too long")
)

type TodoTitle struct {
	value string
}

func NewTodoTitle(val string) (TodoTitle, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return TodoTitle{}, ErrTitleEmpty
	}

	if len(val) > 100 {
		return TodoTitle{}, ErrTitleTooLong
	}

	return TodoTitle{value: val}, nil
}

func (t TodoTitle) String() string {
	return t.value
}
