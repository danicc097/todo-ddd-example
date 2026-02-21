package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrUserNotFound = shared.NewDomainError(apperrors.NotFound, "user not found")

type UserID = shared.ID[User]

type User struct {
	shared.AggregateRoot

	id        UserID
	email     UserEmail
	name      UserName
	createdAt time.Time
}

func ReconstituteUser(id UserID, email UserEmail, name UserName, createdAt time.Time) *User {
	return &User{id: id, email: email, name: name, createdAt: createdAt}
}

func NewUser(email UserEmail, name UserName) *User {
	id := shared.NewID[User]()
	now := time.Now()
	u := &User{
		id:        id,
		email:     email,
		name:      name,
		createdAt: now,
	}
	u.RecordEvent(UserCreatedEvent{
		ID:       id,
		Email:    email,
		Name:     name,
		Occurred: now,
	})

	return u
}

func (u *User) ID() UserID           { return u.id }
func (u *User) Email() UserEmail     { return u.email }
func (u *User) Name() UserName       { return u.name }
func (u *User) CreatedAt() time.Time { return u.createdAt }

func (u *User) Delete() {
	u.RecordEvent(UserDeletedEvent{
		ID:       u.id,
		Occurred: time.Now(),
	})
}
