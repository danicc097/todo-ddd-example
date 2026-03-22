package postgres

import (
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type AuthMapper struct{}

func (m *AuthMapper) MapEvent(e shared.DomainEvent) (shared.EventType, any, error) {
	return "", nil, nil
}
