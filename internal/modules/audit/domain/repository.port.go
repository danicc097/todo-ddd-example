package domain

import "context"

// AuditRepository allows saving audit logs.
//
//counterfeiter:generate . AuditRepository
type AuditRepository interface {
	Save(ctx context.Context, log *AuditLog) error
}
