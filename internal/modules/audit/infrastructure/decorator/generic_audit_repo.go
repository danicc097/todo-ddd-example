package decorator

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	auditDomain "github.com/danicc097/todo-ddd-example/internal/modules/audit/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

type Identifiable[ID any] interface {
	ID() ID
}

type ChangeExtractorFunc[T any] func(entity T) map[string]any

type AuditRepoDecorator[T Identifiable[ID], ID any] struct {
	auditRepo     auditDomain.AuditRepository
	aggregateType auditDomain.AuditAggregateType
	extractor     ChangeExtractorFunc[T]
	idToUUID      func(ID) uuid.UUID
}

func NewAuditRepoDecorator[T Identifiable[ID], ID any](
	auditRepo auditDomain.AuditRepository,
	aggType auditDomain.AuditAggregateType,
	extractor ChangeExtractorFunc[T],
	idToUUID func(ID) uuid.UUID,
) *AuditRepoDecorator[T, ID] {
	return &AuditRepoDecorator[T, ID]{
		auditRepo:     auditRepo,
		aggregateType: aggType,
		extractor:     extractor,
		idToUUID:      idToUUID,
	}
}

// AuditSave wraps a save/update operation.
func (d *AuditRepoDecorator[T, ID]) AuditSave(
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
func (d *AuditRepoDecorator[T, ID]) AuditDelete(
	ctx context.Context,
	id ID,
	fetchFn func(context.Context, ID) (T, error),
	deleteFn func(context.Context, ID) error,
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

func (d *AuditRepoDecorator[T, ID]) log(ctx context.Context, entity T, op auditDomain.AuditOperation) error {
	meta := causation.FromContext(ctx)
	changes := d.extractor(entity)

	auditLog, err := auditDomain.NewAuditLog(
		meta.CorrelationID,
		meta.CausationID,
		&meta.UserID,
		meta.UserIP,
		meta.UserAgent,
		d.aggregateType,
		d.idToUUID(entity.ID()),
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
