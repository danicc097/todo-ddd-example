package postgres

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type WorkspaceMapper struct{}

func (m *WorkspaceMapper) ToDomain(w db.Workspaces, members []db.WorkspaceMembers) (*domain.Workspace, error) {
	memberMap, err := toMemberMap(members)
	if err != nil {
		return nil, fmt.Errorf("failed to map workspace members: %w", err)
	}

	domainMemberMap := make(map[userDomain.UserID]domain.WorkspaceRole, len(memberMap))
	for uid, role := range memberMap {
		domainMemberMap[userDomain.UserID(uid)] = role
	}

	return domain.ReconstituteWorkspace(
		w.ID,
		w.Name,
		w.Description,
		w.CreatedAt,
		domainMemberMap,
	), nil
}

func (m *WorkspaceMapper) ToPersistence(w *domain.Workspace) db.Workspaces {
	return db.Workspaces{
		ID:          w.ID(),
		Name:        w.Name(),
		Description: w.Description(),
		CreatedAt:   w.CreatedAt(),
	}
}

// Outbox DTOs

type workspaceCreatedDTO struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	OwnerID  uuid.UUID `json:"owner_id"`
	Occurred time.Time `json:"occurred_at"`
}

type workspaceDeletedDTO struct {
	ID       uuid.UUID `json:"id"`
	Occurred time.Time `json:"occurred_at"`
}

type memberAddedDTO struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Occurred    time.Time `json:"occurred_at"`
}

type memberRemovedDTO struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	Occurred    time.Time `json:"occurred_at"`
}

func (m *WorkspaceMapper) MapEvent(event shared.DomainEvent) (shared.EventType, []byte, error) {
	var payload any

	switch evt := event.(type) {
	case domain.WorkspaceCreatedEvent:
		payload = workspaceCreatedDTO{
			ID:       evt.ID.UUID(),
			Name:     evt.Name,
			OwnerID:  evt.OwnerID.UUID(),
			Occurred: evt.Occurred,
		}
	case domain.MemberAddedEvent:
		payload = memberAddedDTO{
			WorkspaceID: evt.WorkspaceID.UUID(),
			UserID:      evt.UserID.UUID(),
			Role:        evt.Role,
			Occurred:    evt.Occurred,
		}
	case domain.MemberRemovedEvent:
		payload = memberRemovedDTO{
			WorkspaceID: evt.WorkspaceID.UUID(),
			UserID:      evt.UserID.UUID(),
			Occurred:    evt.Occurred,
		}
	case domain.WorkspaceDeletedEvent:
		payload = workspaceDeletedDTO{
			ID:       evt.ID.UUID(),
			Occurred: evt.Occurred,
		}
	default:
		slog.Warn("received unmapped event type, skipping outbox persistence", slog.Any("event_type", evt.EventName()))
		return "", nil, nil
	}

	b, err := json.Marshal(payload)

	return event.EventName(), b, err
}
