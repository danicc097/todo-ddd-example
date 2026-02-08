package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	id        uuid.UUID
	email     UserEmail
	name      string
	createdAt time.Time
}

func NewUser(id uuid.UUID, email UserEmail, name string, createdAt time.Time) *User {
	return &User{id: id, email: email, name: name, createdAt: createdAt}
}

func CreateUser(email UserEmail, name string) *User {
	return &User{id: uuid.New(), email: email, name: name, createdAt: time.Now()}
}

func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) Email() UserEmail     { return u.email }
func (u *User) Name() string         { return u.name }
func (u *User) CreatedAt() time.Time { return u.createdAt }
