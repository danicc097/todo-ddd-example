package http

import (
	api "github.com/danicc097/todo-ddd-example/internal/generated/api"
	"github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type UserRestMapper struct{}

func (m *UserRestMapper) ToResponse(u *domain.User) api.User {
	return api.User{
		Id:    u.ID(),
		Email: u.Email().String(),
		Name:  u.Name().String(),
	}
}
