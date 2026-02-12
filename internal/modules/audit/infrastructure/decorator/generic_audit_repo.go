package decorator

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	auditDomain "github.com/danicc097/todo-ddd-example/internal/modules/audit/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type Identifiable interface {
	ID() uuid.UUID
}

type ChangeExtractorFunc[T any] func(entity T) map[string]any

type AuditRepoDecorator[T Identifiable] struct {
	auditRepo     auditDomain.AuditRepository
	aggregateType auditDomain.AuditAggregateType
	extractor     ChangeExtractorFunc[T]
}

func NewAuditRepoDecorator[T Identifiable](
	auditRepo auditDomain.AuditRepository,
	aggType auditDomain.AuditAggregateType,
	extractor ChangeExtractorFunc[T],
) *AuditRepoDecorator[T] {
	return &AuditRepoDecorator[T]{
		auditRepo:     auditRepo,
		aggregateType: aggType,
		extractor:     extractor,
	}
}

// AuditSave wraps a save/update operation.
func (d *AuditRepoDecorator[T]) AuditSave(
	ctx context.Context,
	entity T,
	op auditDomain.AuditOperation,
	saveFn func(context.Context, T) error,
) error {
	if err := saveFn(ctx, entity); err != nil {
		return err
	}

	return d.log(ctx, entity, op)
}

// AuditDelete wraps a delete operation.
// It requires a fetchFn to snapshot the state before deletion.
func (d *AuditRepoDecorator[T]) AuditDelete(
	ctx context.Context,
	id uuid.UUID,
	fetchFn func(context.Context, uuid.UUID) (T, error),
	deleteFn func(context.Context, uuid.UUID) error,
) error {
	entity, err := fetchFn(ctx, id)
	if err != nil {
		return err // probably cant delete if cant find
	}

	if err := deleteFn(ctx, id); err != nil {
		return err
	}

	// log the snapshot
	return d.log(ctx, entity, auditDomain.OpDelete)
}

func (d *AuditRepoDecorator[T]) log(ctx context.Context, entity T, op auditDomain.AuditOperation) error {
	meta := causation.FromContext(ctx)
	changes := d.extractor(entity)

	auditLog, err := auditDomain.NewAuditLog(
		meta.CorrelationID,
		meta.CausationID,
		&meta.UserID,
		meta.UserIP,
		meta.UserAgent,
		d.aggregateType,
		entity.ID(),
		op,
		changes,
	)
	if err != nil {
		return fmt.Errorf("audit log creation failed: %w", err)
	}

	if err := d.auditRepo.Save(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}
