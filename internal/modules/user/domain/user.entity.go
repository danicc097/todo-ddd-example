package domain

import (
	"errors"
	"time"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserID = shared.ID[User]

type User struct {
	id        UserID
	email     UserEmail
	name      string
	createdAt time.Time
}

func NewUser(id UserID, email UserEmail, name string, createdAt time.Time) *User {
	return &User{id: id, email: email, name: name, createdAt: createdAt}
}

func CreateUser(email UserEmail, name string) *User {
	return &User{id: shared.NewID[User](), email: email, name: name, createdAt: time.Now()}
}

func (u *User) ID() UserID          { return u.id }
func (u *User) Email() UserEmail     { return u.email }
func (u *User) Name() string         { return u.name }
func (u *User) CreatedAt() time.Time { return u.createdAt }
