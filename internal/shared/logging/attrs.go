package logging

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

// EventIDAttr returns a slog attribute for an event ID.
func EventIDAttr(id uuid.UUID) slog.Attr {
	return slog.String(causation.AttrEventID, id.String())
}
