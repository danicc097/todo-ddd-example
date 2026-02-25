package domain

import (
	"context"

	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

type ScheduleRepository interface {
	Save(ctx context.Context, s *DailySchedule) error
	FindByUserAndDate(ctx context.Context, userID userDomain.UserID, date ScheduleDate) (*DailySchedule, error)
	FindSchedulesByTodoID(ctx context.Context, todoID todoDomain.TodoID) ([]*DailySchedule, error)
}
