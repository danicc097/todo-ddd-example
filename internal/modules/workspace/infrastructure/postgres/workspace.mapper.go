package postgres

import (
	"fmt"
	"log/slog"
	"time"

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

	name, _ := domain.NewWorkspaceName(w.Name)
	desc, _ := domain.NewWorkspaceDescription(w.Description)

	return domain.ReconstituteWorkspace(domain.ReconstituteWorkspaceArgs{
		ID:          w.ID,
		Name:        name,
		Description: desc,
		CreatedAt:   w.CreatedAt,
		Members:     domainMemberMap,
	}), nil
}

func (m *WorkspaceMapper) ToPersistence(w *domain.Workspace) db.Workspaces {
	return db.Workspaces{
		ID:          w.ID(),
		Name:        w.Name().String(),
		Description: w.Description().String(),
		CreatedAt:   w.CreatedAt(),
	}
}

// Outbox DTOs

type WorkspaceCreatedDTO struct {
	ID           domain.WorkspaceID `json:"id"`
	Name         string             `json:"name"`
	OwnerID      userDomain.UserID  `json:"owner_id"`
	Occurred     time.Time          `json:"occurred_at"`
	EventVersion int                `json:"event_version"`
}

type WorkspaceDeletedDTO struct {
	ID           domain.WorkspaceID `json:"id"`
	Occurred     time.Time          `json:"occurred_at"`
	EventVersion int                `json:"event_version"`
}

type MemberAddedDTO struct {
	WorkspaceID  domain.WorkspaceID `json:"workspace_id"`
	UserID       userDomain.UserID  `json:"user_id"`
	Role         string             `json:"role"`
	Occurred     time.Time          `json:"occurred_at"`
	EventVersion int                `json:"event_version"`
}

type MemberRemovedDTO struct {
	WorkspaceID  domain.WorkspaceID `json:"workspace_id"`
	UserID       userDomain.UserID  `json:"user_id"`
	Occurred     time.Time          `json:"occurred_at"`
	EventVersion int                `json:"event_version"`
}

func (m *WorkspaceMapper) MapEvent(event shared.DomainEvent) (shared.EventType, any, error) {
	var payload any

	switch evt := event.(type) {
	case domain.WorkspaceCreatedEvent:
		payload = WorkspaceCreatedDTO{
			ID:           evt.ID,
			Name:         evt.Name.String(),
			OwnerID:      evt.OwnerID,
			Occurred:     evt.Occurred,
			EventVersion: 1,
		}
	case domain.MemberAddedEvent:
		payload = MemberAddedDTO{
			WorkspaceID:  evt.WsID,
			UserID:       evt.UserID,
			Role:         evt.Role.String(),
			Occurred:     evt.Occurred,
			EventVersion: 1,
		}
	case domain.MemberRemovedEvent:
		payload = MemberRemovedDTO{
			WorkspaceID:  evt.WsID,
			UserID:       evt.UserID,
			Occurred:     evt.Occurred,
			EventVersion: 1,
		}
	case domain.WorkspaceDeletedEvent:
		payload = WorkspaceDeletedDTO{
			ID:           evt.ID,
			Occurred:     evt.Occurred,
			EventVersion: 1,
		}
	default:
		slog.Warn("received unmapped event type, skipping outbox persistence", slog.Any("event_type", event.EventName()))
		return "", nil, nil
	}

	return event.EventName(), payload, nil
}
