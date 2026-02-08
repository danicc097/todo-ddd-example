package domain

import (
	"errors"
	"strings"
)

var (
	ErrTitleTooLong = errors.New("title is too long")
	ErrEmptyTitle   = errors.New("title cannot be emprty")
)

type TodoTitle struct {
	value string
}

func NewTodoTitle(val string) (TodoTitle, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return TodoTitle{}, ErrEmptyTitle
	}
	if len(val) > 100 {
		return TodoTitle{}, ErrTitleTooLong
	}
	return TodoTitle{value: val}, nil
}

func (t TodoTitle) String() string {
	return t.value
}
