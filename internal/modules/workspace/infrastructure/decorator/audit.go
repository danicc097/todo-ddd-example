package decorator

import (
	"context"

	"github.com/google/uuid"

	auditDomain "github.com/danicc097/todo-ddd-example/internal/modules/audit/domain"
	auditDecorator "github.com/danicc097/todo-ddd-example/internal/modules/audit/infrastructure/decorator"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

var _ auditDecorator.Identifiable = (*wsDomain.Workspace)(nil)

type WorkspaceAuditWrapper struct {
	base    wsDomain.WorkspaceRepository
	auditor *auditDecorator.AuditRepoDecorator[*wsDomain.Workspace]
}

func NewWorkspaceAuditWrapper(
	base wsDomain.WorkspaceRepository,
	auditRepo auditDomain.AuditRepository,
) *WorkspaceAuditWrapper {
	extractor := func(w *wsDomain.Workspace) map[string]any {
		return map[string]any{
			"name":        w.Name(),
			"members_cnt": len(w.Members()),
		}
	}

	auditor := auditDecorator.NewAuditRepoDecorator(
		auditRepo,
		auditDomain.AggWorkspace,
		extractor,
	)

	return &WorkspaceAuditWrapper{
		base:    base,
		auditor: auditor,
	}
}

func (w *WorkspaceAuditWrapper) Save(ctx context.Context, entity *wsDomain.Workspace) error {
	// assume Create for now, could detect
	return w.auditor.AuditSave(ctx, entity, auditDomain.OpCreate, w.base.Save)
}

func (w *WorkspaceAuditWrapper) Delete(ctx context.Context, id uuid.UUID) error {
	// handles fetch-before-delete logic
	return w.auditor.AuditDelete(ctx, id, w.base.FindByID, w.base.Delete)
}

/*
* Read methods bypass audit
 */

func (w *WorkspaceAuditWrapper) FindByID(ctx context.Context, id uuid.UUID) (*wsDomain.Workspace, error) {
	return w.base.FindByID(ctx, id)
}

func (w *WorkspaceAuditWrapper) FindAll(ctx context.Context) ([]*wsDomain.Workspace, error) {
	return w.base.FindAll(ctx)
}
