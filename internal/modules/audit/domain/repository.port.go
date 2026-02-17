package domain

import "context"

// AuditRepository allows saving audit logs.
type AuditRepository interface {
	Save(ctx context.Context, log *AuditLog) error
}
