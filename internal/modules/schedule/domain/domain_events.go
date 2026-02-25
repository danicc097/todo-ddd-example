package domain

import (
	"time"

	"github.com/google/uuid"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type DailyScheduleCreatedEvent struct {
	UserID      userDomain.UserID
	Date        ScheduleDate
	MaxCapacity int
	Occurred    time.Time
}

func (e DailyScheduleCreatedEvent) EventName() shared.EventType         { return shared.ScheduleCreated }
func (e DailyScheduleCreatedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e DailyScheduleCreatedEvent) AggregateID() uuid.UUID              { return e.UserID.UUID() } // good enough for routing
func (e DailyScheduleCreatedEvent) AggregateType() shared.AggregateType { return shared.AggSchedule }

type TaskCommittedToScheduleEvent struct {
	UserID   userDomain.UserID
	Date     ScheduleDate
	TodoID   uuid.UUID
	Cost     EnergyCost
	Occurred time.Time
}

func (e TaskCommittedToScheduleEvent) EventName() shared.EventType         { return shared.TaskCommitted }
func (e TaskCommittedToScheduleEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TaskCommittedToScheduleEvent) AggregateID() uuid.UUID              { return e.UserID.UUID() } // good enough for routing
func (e TaskCommittedToScheduleEvent) AggregateType() shared.AggregateType { return shared.AggSchedule }
