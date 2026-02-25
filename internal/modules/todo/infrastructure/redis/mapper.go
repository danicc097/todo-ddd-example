package redis

import (
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type TodoCacheDTO struct {
	ID                 uuid.UUID              `json:"id"`
	WorkspaceID        uuid.UUID              `json:"workspace_id"`
	Title              string                 `json:"title"`
	Status             string                 `json:"status"`
	CreatedAt          time.Time              `json:"created_at"`
	Tags               []uuid.UUID            `json:"tags"`
	DueDate            *time.Time             `json:"due_date"`
	RecurrenceInterval *string                `json:"recurrence_interval"`
	RecurrenceAmount   *int                   `json:"recurrence_amount"`
	LastCompletedAt    *time.Time             `json:"last_completed_at"`
	Sessions           []FocusSessionCacheDTO `json:"sessions"`
}

type FocusSessionCacheDTO struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

func ToTodoCacheDTO(t *domain.Todo) TodoCacheDTO {
	tagUUIDs := make([]uuid.UUID, len(t.Tags()))
	for i, id := range t.Tags() {
		tagUUIDs[i] = id.UUID()
	}

	var (
		rInterval *string
		rAmount   *int
	)

	if t.Recurrence() != nil {
		interval := t.Recurrence().Interval()
		amount := t.Recurrence().Amount()
		rInterval = &interval
		rAmount = &amount
	}

	sessions := make([]FocusSessionCacheDTO, len(t.Sessions()))
	for i, s := range t.Sessions() {
		sessions[i] = FocusSessionCacheDTO{
			ID:        s.ID().UUID(),
			UserID:    s.UserID().UUID(),
			StartTime: s.StartTime(),
			EndTime:   s.EndTime(),
		}
	}

	return TodoCacheDTO{
		ID:                 t.ID().UUID(),
		WorkspaceID:        t.WorkspaceID().UUID(),
		Title:              t.Title().String(),
		Status:             t.Status().String(),
		CreatedAt:          t.CreatedAt(),
		Tags:               tagUUIDs,
		DueDate:            t.DueDate(),
		RecurrenceInterval: rInterval,
		RecurrenceAmount:   rAmount,
		LastCompletedAt:    t.LastCompletedAt(),
		Sessions:           sessions,
	}
}

func FromTodoCacheDTO(dto TodoCacheDTO) *domain.Todo {
	title, _ := domain.NewTodoTitle(dto.Title)

	tagIDs := make([]domain.TagID, len(dto.Tags))
	for i, id := range dto.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	var recurrence *domain.RecurrenceRule

	if dto.RecurrenceInterval != nil && dto.RecurrenceAmount != nil {
		r, _ := domain.NewRecurrenceRule(*dto.RecurrenceInterval, *dto.RecurrenceAmount)
		recurrence = &r
	}

	sessions := make([]domain.FocusSession, len(dto.Sessions))
	for i, s := range dto.Sessions {
		sessions[i] = domain.ReconstituteFocusSession(domain.ReconstituteFocusSessionArgs{
			ID:        domain.FocusSessionID(s.ID),
			UserID:    userDomain.UserID(s.UserID),
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
		})
	}

	return domain.ReconstituteTodo(domain.ReconstituteTodoArgs{
		ID:              domain.TodoID(dto.ID),
		Title:           title,
		Status:          domain.TodoStatus(dto.Status),
		CreatedAt:       dto.CreatedAt,
		Tags:            tagIDs,
		WorkspaceID:     wsDomain.WorkspaceID(dto.WorkspaceID),
		DueDate:         dto.DueDate,
		Recurrence:      recurrence,
		LastCompletedAt: dto.LastCompletedAt,
		Sessions:        sessions,
	})
}

type TagCacheDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

func ToTagCacheDTO(t *domain.Tag) TagCacheDTO {
	return TagCacheDTO{
		ID:          t.ID().UUID(),
		Name:        t.Name().String(),
		WorkspaceID: t.WorkspaceID().UUID(),
	}
}

func FromTagCacheDTO(dto TagCacheDTO) *domain.Tag {
	name, _ := domain.NewTagName(dto.Name)

	return domain.ReconstituteTag(domain.ReconstituteTagArgs{
		ID:          domain.TagID(dto.ID),
		Name:        name,
		WorkspaceID: wsDomain.WorkspaceID(dto.WorkspaceID),
	})
}
