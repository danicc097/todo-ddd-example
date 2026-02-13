package domain

import (
	"github.com/google/uuid"
)

// ID is a generic identifier type to distinguish entities.
type ID[T any] struct {
	uuid.UUID
}

// NewID creates a new generated ID.
func NewID[T any]() ID[T] {
	return ID[T]{UUID: uuid.New()}
}

// ParseID safely parses a string into a typed ID.
func ParseID[T any](s string) (ID[T], error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID[T]{}, err
	}

	return ID[T]{UUID: u}, nil
}
