package application

import (
	"context"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type UserProvider interface {
	Exists(ctx context.Context, userID userDomain.UserID) (bool, error)
}
