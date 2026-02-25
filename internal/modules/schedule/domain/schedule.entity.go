package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	todoDomain "github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrDailyCapacityExceeded = shared.NewDomainError(apperrors.Conflict, "daily energy capacity exceeded")
	ErrInvalidEnergyCost     = shared.NewDomainError(apperrors.InvalidInput, "energy cost must be between 1 and 5")
	ErrInvalidCapacity       = shared.NewDomainError(apperrors.InvalidInput, "capacity must be greater than 0")
	ErrScheduleNotFound      = shared.NewDomainError(apperrors.NotFound, "schedule not found")
	ErrConcurrentUpdate      = shared.ErrConcurrentUpdate
)

// EnergyCost represents the estimated effort required to complete a task.
// Higher values mean more effort is required.
type EnergyCost int

func NewEnergyCost(val int) (EnergyCost, error) {
	if val < 1 || val > 5 {
		return 0, ErrInvalidEnergyCost
	}

	return EnergyCost(val), nil
}

type ScheduleDate string

func NewScheduleDate(t time.Time) ScheduleDate {
	return ScheduleDate(t.Format(time.DateOnly))
}

func (d ScheduleDate) String() string { return string(d) }

type DailySchedule struct {
	shared.AggregateRoot

	userID         userDomain.UserID
	date           ScheduleDate
	maxCapacity    int
	version        int
	committedTasks map[todoDomain.TodoID]EnergyCost
}

func NewDailySchedule(userID userDomain.UserID, date ScheduleDate, maxCapacity int) (*DailySchedule, error) {
	if maxCapacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	s := &DailySchedule{
		userID:         userID,
		date:           date,
		maxCapacity:    maxCapacity,
		version:        0,
		committedTasks: make(map[todoDomain.TodoID]EnergyCost),
	}

	s.RecordEvent(DailyScheduleCreatedEvent{
		UserID:      s.userID,
		Date:        s.date,
		MaxCapacity: s.maxCapacity,
		Occurred:    time.Now(),
	})

	return s, nil
}

type ReconstituteDailyScheduleArgs struct {
	UserID         userDomain.UserID
	Date           ScheduleDate
	MaxCapacity    int
	Version        int
	CommittedTasks map[todoDomain.TodoID]EnergyCost
}

func ReconstituteDailySchedule(args ReconstituteDailyScheduleArgs) *DailySchedule {
	return &DailySchedule{
		userID:         args.UserID,
		date:           args.Date,
		maxCapacity:    args.MaxCapacity,
		version:        args.Version,
		committedTasks: args.CommittedTasks,
	}
}

func (s *DailySchedule) CommitTask(todoID todoDomain.TodoID, cost EnergyCost) error {
	currentLoad := 0
	for _, c := range s.committedTasks {
		currentLoad += int(c)
	}

	if currentLoad+int(cost) > s.maxCapacity {
		return ErrDailyCapacityExceeded
	}

	s.committedTasks[todoID] = cost
	s.RecordEvent(TaskCommittedToScheduleEvent{
		UserID:   s.userID,
		Date:     s.date,
		TodoID:   todoID.UUID(),
		Cost:     cost,
		Occurred: time.Now(),
	})

	return nil
}

func (s *DailySchedule) RemoveTask(todoID todoDomain.TodoID) {
	delete(s.committedTasks, todoID)
}

func (s *DailySchedule) UserID() userDomain.UserID                        { return s.userID }
func (s *DailySchedule) Date() ScheduleDate                               { return s.date }
func (s *DailySchedule) MaxCapacity() int                                 { return s.maxCapacity }
func (s *DailySchedule) Version() int                                     { return s.version }
func (s *DailySchedule) CommittedTasks() map[todoDomain.TodoID]EnergyCost { return s.committedTasks }
