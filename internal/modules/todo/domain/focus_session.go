package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrFocusSessionAlreadyActive  = shared.NewDomainError(apperrors.Conflict, "an active focus session already exists")
	ErrNoActiveFocusSession       = shared.NewDomainError(apperrors.Unprocessable, "no active focus session found")
	ErrCannotFocusOnCompletedTask = shared.NewDomainError(apperrors.Unprocessable, "cannot focus on a completed task")
	ErrInvalidFocusStopTimeAfter  = shared.NewDomainError(apperrors.InvalidInput, "stop time must be after start time")
)

type FocusSessionID uuid.UUID

func (id FocusSessionID) UUID() uuid.UUID { return uuid.UUID(id) }

type FocusSession struct {
	id        FocusSessionID
	userID    userDomain.UserID
	startTime time.Time
	endTime   *time.Time
}

func NewFocusSession(id FocusSessionID, userID userDomain.UserID, start time.Time) FocusSession {
	return FocusSession{id: id, userID: userID, startTime: start}
}

type ReconstituteFocusSessionArgs struct {
	ID        FocusSessionID
	UserID    userDomain.UserID
	StartTime time.Time
	EndTime   *time.Time
}

func ReconstituteFocusSession(args ReconstituteFocusSessionArgs) FocusSession {
	return FocusSession{id: args.ID, userID: args.UserID, startTime: args.StartTime, endTime: args.EndTime}
}

func (s FocusSession) IsActive() bool            { return s.endTime == nil }
func (s FocusSession) ID() FocusSessionID        { return s.id }
func (s FocusSession) UserID() userDomain.UserID { return s.userID }
func (s FocusSession) StartTime() time.Time      { return s.startTime }
func (s FocusSession) EndTime() *time.Time       { return s.endTime }
