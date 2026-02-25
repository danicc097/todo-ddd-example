package postgres

import (
	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type TagMapper struct{}

func (m *TagMapper) ToDomain(row db.Tags) *domain.Tag {
	name, _ := domain.NewTagName(row.Name)

	return domain.ReconstituteTag(domain.ReconstituteTagArgs{
		ID:          row.ID,
		Name:        name,
		WorkspaceID: row.WorkspaceID,
	})
}

// ToPersistence maps Domain to the primary table struct.
func (m *TagMapper) ToPersistence(t *domain.Tag) db.Tags {
	return db.Tags{
		ID:          t.ID(),
		Name:        t.Name().String(),
		WorkspaceID: t.WorkspaceID(),
	}
}

type TagCreatedDTO struct {
	ID           domain.TagID         `json:"id"`
	Name         string               `json:"name"`
	WorkspaceID  wsDomain.WorkspaceID `json:"workspace_id"`
	EventVersion int                  `json:"event_version"`
}

func (m *TagMapper) MapEvent(e shared.DomainEvent) (shared.EventType, any, error) {
	switch evt := e.(type) {
	case domain.TagCreatedEvent:
		return shared.TodoTagCreated, TagCreatedDTO{
			ID:           evt.ID,
			Name:         evt.Name.String(),
			WorkspaceID:  evt.WsID,
			EventVersion: 1,
		}, nil
	}

	return "", nil, nil
}
