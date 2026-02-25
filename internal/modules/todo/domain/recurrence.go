package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrCannotCompleteFutureOccurrence = shared.NewDomainError(apperrors.Unprocessable, "cannot complete future occurrence")
	ErrInvalidRecurrenceRule          = shared.NewDomainError(apperrors.InvalidInput, "invalid recurrence rule")
)

type RecurrenceInterval string

const (
	Daily   RecurrenceInterval = "DAILY"
	Weekly  RecurrenceInterval = "WEEKLY"
	Monthly RecurrenceInterval = "MONTHLY"
)

type RecurrenceRule struct {
	interval RecurrenceInterval
	amount   int
}

func NewRecurrenceRule(interval string, amount int) (RecurrenceRule, error) {
	if amount <= 0 {
		return RecurrenceRule{}, ErrInvalidRecurrenceRule
	}

	rInterval := RecurrenceInterval(interval)
	switch rInterval {
	case Daily, Weekly, Monthly:
		return RecurrenceRule{interval: rInterval, amount: amount}, nil
	}

	return RecurrenceRule{}, ErrInvalidRecurrenceRule
}

func (r RecurrenceRule) CalculateNext(from time.Time) time.Time {
	switch r.interval {
	case Daily:
		return from.AddDate(0, 0, r.amount)
	case Weekly:
		return from.AddDate(0, 0, 7*r.amount)
	case Monthly:
		next := from.AddDate(0, r.amount, 0)
		if next.Day() != from.Day() {
			next = next.AddDate(0, 0, -next.Day())
		}

		return next
	}

	return from
}

func (r RecurrenceRule) Interval() string { return string(r.interval) }
func (r RecurrenceRule) Amount() int      { return r.amount }
