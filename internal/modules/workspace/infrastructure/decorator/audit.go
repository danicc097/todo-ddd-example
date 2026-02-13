package decorator

import (
	"context"

	"github.com/google/uuid"

	auditDomain "github.com/danicc097/todo-ddd-example/internal/modules/audit/domain"
	auditDecorator "github.com/danicc097/todo-ddd-example/internal/modules/audit/infrastructure/decorator"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

var _ wsDomain.WorkspaceRepository = (*WorkspaceAuditWrapper)(nil)

type WorkspaceAuditWrapper struct {
	base    wsDomain.WorkspaceRepository
	auditor *auditDecorator.AuditRepoDecorator[*wsDomain.Workspace, wsDomain.WorkspaceID]
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
		func(id wsDomain.WorkspaceID) uuid.UUID {
			return id.UUID
		},
	)

	return &WorkspaceAuditWrapper{
		base:    base,
		auditor: auditor,
	}
}

func (w *WorkspaceAuditWrapper) Save(ctx context.Context, entity *wsDomain.Workspace) error {
	return w.auditor.AuditSave(ctx, entity, auditDomain.OpUpsert, w.base.Save)
}

func (w *WorkspaceAuditWrapper) Delete(ctx context.Context, id wsDomain.WorkspaceID) error {
	return w.auditor.AuditDelete(ctx, id, w.base.FindByID, w.base.Delete)
}

/*
* Read methods bypass audit
 */

func (w *WorkspaceAuditWrapper) FindByID(ctx context.Context, id wsDomain.WorkspaceID) (*wsDomain.Workspace, error) {
	return w.base.FindByID(ctx, id)
}

func (w *WorkspaceAuditWrapper) FindAll(ctx context.Context) ([]*wsDomain.Workspace, error) {
	return w.base.FindAll(ctx)
}
