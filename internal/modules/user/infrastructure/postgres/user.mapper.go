package postgres

import (
	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type UserMapper struct{}

func (m *UserMapper) ToDomain(row db.Users) *domain.User {
	email, _ := domain.NewUserEmail(row.Email)
	name, _ := domain.NewUserName(row.Name)

	return domain.ReconstituteUser(row.ID, email, name, row.CreatedAt)
}

func (m *UserMapper) ToPersistence(u *domain.User) db.Users {
	return db.Users{
		ID:        u.ID(),
		Email:     u.Email().String(),
		Name:      u.Name().String(),
		CreatedAt: u.CreatedAt(),
	}
}

type UserCreatedDTO struct {
	ID           domain.UserID `json:"id"`
	Email        string        `json:"email"`
	Name         string        `json:"name"`
	EventVersion int           `json:"event_version"`
}

type UserDeletedDTO struct {
	ID           domain.UserID `json:"id"`
	EventVersion int           `json:"event_version"`
}

func (m *UserMapper) MapEvent(e shared.DomainEvent) (shared.EventType, any, error) {
	switch evt := e.(type) {
	case domain.UserCreatedEvent:
		return shared.UserCreated, UserCreatedDTO{
			ID:           evt.ID,
			Email:        evt.Email.String(),
			Name:         evt.Name.String(),
			EventVersion: 1,
		}, nil
	case domain.UserDeletedEvent:
		return shared.UserDeleted, UserDeletedDTO{
			ID:           evt.ID,
			EventVersion: 1,
		}, nil
	}

	return "", nil, nil
}
