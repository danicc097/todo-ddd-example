package causation

import (
	"context"

	"github.com/google/uuid"
)

type key int

const (
	metadataKey key = 1
	systemID        = "system"
)

// Metadata carries traceability info.
type Metadata struct {
	CorrelationID string    // original request ID
	CausationID   string    // ID of the event that triggered this
	UserID        uuid.UUID // who
	UserIP        string    // where
	UserAgent     string    // how
	MFAVerified   bool
}

func (m Metadata) IsUser() bool {
	return m.UserID != uuid.Nil
}

func (m Metadata) IsSystem() bool {
	return m.UserIP == systemID
}

func FromContext(ctx context.Context) Metadata {
	if md, ok := ctx.Value(metadataKey).(Metadata); ok {
		return md
	}

	// background jobs, etc. fallback
	return Metadata{
		CorrelationID: uuid.NewString(),
		UserIP:        systemID,
	}
}

func WithMetadata(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, metadataKey, md)
}

// NewNextStep creates metadata for the next step in the chain.
func NewNextStep(ctx context.Context, currentStepID string) Metadata {
	prev := FromContext(ctx)

	return Metadata{
		CorrelationID: prev.CorrelationID,
		CausationID:   currentStepID, // event -> command
		UserID:        prev.UserID,
		UserIP:        prev.UserIP,
		UserAgent:     prev.UserAgent,
	}
}
