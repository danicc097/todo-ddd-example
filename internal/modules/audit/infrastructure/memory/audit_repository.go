package memory

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/danicc097/todo-ddd-example/internal/modules/audit/domain"
)

// InMemoryAuditRepo simulates a client to a remote audit storage service.
// For demonstration purposes only.
type InMemoryAuditRepo struct {
	mu   sync.RWMutex
	logs []*domain.AuditLog
}

func NewAuditRepository() *InMemoryAuditRepo {
	return &InMemoryAuditRepo{
		logs: make([]*domain.AuditLog, 0),
	}
}

// Save implements domain.AuditRepository.
func (r *InMemoryAuditRepo) Save(ctx context.Context, log *domain.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = append(r.logs, log)

	debugDump(ctx, log)

	return nil
}

func debugDump(ctx context.Context, log *domain.AuditLog) {
	// similar to what would be sent to SIEM/ElasticSearch/...
	dto := map[string]any{
		"event_id":          log.ID(),
		"trace_correlation": log.CorrelationID(),
		"trace_causation":   log.CausationID(),
		"actor_id":          log.ActorID(),
		"actor_ip":          log.ActorIP(),
		"ua_hash":           log.UserAgentHash(),
		"target_resource":   log.AggregateType().String() + ":" + log.AggregateID().String(),
		"action":            log.Operation(),
		"diff":              log.Changes(),
		"timestamp":         log.OccurredAt(),
	}

	b, _ := json.Marshal(dto)
	slog.InfoContext(ctx, "AUDIT_LOG_EMITTED", slog.String("payload", string(b)))
}

func (r *InMemoryAuditRepo) FindAll() []*domain.AuditLog {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dst := make([]*domain.AuditLog, len(r.logs))
	copy(dst, r.logs)

	return dst
}
