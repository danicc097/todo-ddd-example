package postgres

import (
	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type UserMapper struct{}

func (m *UserMapper) ToDomain(row db.Users) *domain.User {
	email, _ := domain.NewUserEmail(row.Email)
	name, _ := domain.NewUserName(row.Name)

	return domain.NewUser(row.ID, email, name, row.CreatedAt)
}

func (m *UserMapper) ToPersistence(u *domain.User) db.Users {
	return db.Users{
		ID:        u.ID(),
		Email:     u.Email().String(),
		Name:      u.Name().String(),
		CreatedAt: u.CreatedAt(),
	}
}
