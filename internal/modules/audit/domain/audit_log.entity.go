package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type AuditLog struct {
	id            uuid.UUID
	correlationID string
	causationID   string
	actorID       *uuid.UUID
	actorIP       string
	userAgentHash string
	aggregateType shared.AggregateType
	aggregateID   uuid.UUID
	operation     AuditOperation
	changes       map[string]any
	occurredAt    time.Time
}

func NewAuditLog(
	correlationID string,
	causationID string,
	actorID *uuid.UUID,
	actorIP string,
	userAgentRaw string,
	aggType shared.AggregateType,
	aggID uuid.UUID,
	op AuditOperation,
	changes map[string]any,
) (*AuditLog, error) {
	if err := op.IsValid(); err != nil {
		return nil, err
	}

	if correlationID == "" {
		return nil, errors.New("audit log requires correlation_id")
	}

	return &AuditLog{
		id:            uuid.New(),
		correlationID: correlationID,
		causationID:   causationID,
		actorID:       actorID,
		actorIP:       actorIP,                  // PII
		userAgentHash: hashString(userAgentRaw), // only care whether it changed
		aggregateType: aggType,
		aggregateID:   aggID,
		operation:     op,
		changes:       changes,
		occurredAt:    time.Now(),
	}, nil
}

func hashString(s string) string {
	if s == "" {
		return ""
	}

	h := sha256.Sum256([]byte(s))

	return hex.EncodeToString(h[:])
}

func (a *AuditLog) ID() uuid.UUID                       { return a.id }
func (a *AuditLog) CorrelationID() string               { return a.correlationID }
func (a *AuditLog) CausationID() string                 { return a.causationID }
func (a *AuditLog) ActorID() *uuid.UUID                 { return a.actorID }
func (a *AuditLog) ActorIP() string                     { return a.actorIP }
func (a *AuditLog) UserAgentHash() string               { return a.userAgentHash }
func (a *AuditLog) AggregateType() shared.AggregateType { return shared.AggregateType(a.aggregateType) }
func (a *AuditLog) AggregateID() uuid.UUID              { return a.aggregateID }
func (a *AuditLog) Operation() string                   { return a.operation.String() }
func (a *AuditLog) Changes() map[string]any             { return a.changes }
func (a *AuditLog) OccurredAt() time.Time               { return a.occurredAt }
