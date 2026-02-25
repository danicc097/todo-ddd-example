package postgres

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/schedule/domain"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type ScheduleMapper struct{}

func (m *ScheduleMapper) ToDomain(s db.DailySchedules, tasks []db.ScheduleTasks) *domain.DailySchedule {
	committedTasks := make(map[todoDomain.TodoID]domain.EnergyCost)

	for _, t := range tasks {
		cost, _ := domain.NewEnergyCost(int(t.EnergyCost))
		committedTasks[todoDomain.TodoID(t.TodoID)] = cost
	}

	dateStr := s.Date.Format(time.DateOnly)

	return domain.ReconstituteDailySchedule(domain.ReconstituteDailyScheduleArgs{
		UserID:         userDomain.UserID(s.UserID),
		Date:           domain.ScheduleDate(dateStr),
		MaxCapacity:    int(s.MaxCapacity),
		Version:        int(s.Version),
		CommittedTasks: committedTasks,
	})
}

func (m *ScheduleMapper) ToPersistence(s *domain.DailySchedule) db.DailySchedules {
	t, _ := time.Parse(time.DateOnly, s.Date().String())

	return db.DailySchedules{
		UserID:      s.UserID().UUID(),
		Date:        t,
		MaxCapacity: int32(s.MaxCapacity()),
		Version:     int32(s.Version()),
	}
}

type ScheduleCreatedDTO struct {
	UserID      userDomain.UserID `json:"user_id"`
	Date        string            `json:"date"`
	MaxCapacity int               `json:"max_capacity"`
}

type TaskCommittedDTO struct {
	UserID userDomain.UserID `json:"user_id"`
	Date   string            `json:"date"`
	TodoID todoDomain.TodoID `json:"todo_id"`
	Cost   int               `json:"cost"`
}

func (m *ScheduleMapper) MapEvent(e shared.DomainEvent) (shared.EventType, any, error) {
	switch evt := e.(type) {
	case domain.DailyScheduleCreatedEvent:
		return shared.ScheduleCreated, ScheduleCreatedDTO{
			UserID:      evt.UserID,
			Date:        evt.Date.String(),
			MaxCapacity: evt.MaxCapacity,
		}, nil
	case domain.TaskCommittedToScheduleEvent:
		return shared.TaskCommitted, TaskCommittedDTO{
			UserID: evt.UserID,
			Date:   evt.Date.String(),
			TodoID: todoDomain.TodoID(evt.TodoID),
			Cost:   int(evt.Cost),
		}, nil
	}

	return "", nil, nil
}
