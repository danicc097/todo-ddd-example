package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ID is a generic identifier type to distinguish entities.
type ID[T any] uuid.UUID

// NewID creates a new generated ID.
func NewID[T any]() ID[T] {
	return ID[T](uuid.New())
}

func (id ID[T]) String() string {
	return uuid.UUID(id).String()
}

func (id ID[T]) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ID[T]) IsNil() bool {
	return uuid.UUID(id) == uuid.Nil
}

func (id ID[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *ID[T]) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	u, err := uuid.Parse(s)
	if err != nil {
		return err
	}

	*id = ID[T](u)

	return nil
}

func (id *ID[T]) UnmarshalText(data []byte) error {
	var u uuid.UUID
	if err := u.UnmarshalText(data); err != nil {
		return err
	}

	*id = ID[T](u)

	return nil
}

func (id ID[T]) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}
