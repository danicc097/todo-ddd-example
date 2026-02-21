package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	infraDB "github.com/danicc097/todo-ddd-example/internal/infrastructure/db"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	sharedPg "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/postgres"
)

type WorkspaceRepo struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	mapper *WorkspaceMapper
}

func NewWorkspaceRepo(pool *pgxpool.Pool) *WorkspaceRepo {
	return &WorkspaceRepo{
		q:      db.New(),
		pool:   pool,
		mapper: &WorkspaceMapper{},
	}
}

func (r *WorkspaceRepo) getDB(ctx context.Context) db.DBTX {
	if tx := infraDB.ExtractTx(ctx); tx != nil {
		return tx
	}

	return r.pool
}

func (r *WorkspaceRepo) Save(ctx context.Context, w *domain.Workspace) error {
	dbtx := r.getDB(ctx)

	_, err := r.q.UpsertWorkspace(ctx, dbtx, db.UpsertWorkspaceParams{
		ID:          w.ID(),
		Name:        w.Name().String(),
		Description: w.Description().String(),
		CreatedAt:   w.CreatedAt(),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert workspace: %w", sharedPg.ParseDBError(err))
	}

	currentMembers, err := r.q.GetWorkspaceMembers(ctx, dbtx, w.ID())
	if err != nil {
		return fmt.Errorf("failed to get workspace members: %w", sharedPg.ParseDBError(err))
	}

	ccmm := make(map[uuid.UUID]string)
	for _, m := range currentMembers {
		ccmm[m.UserID.UUID()] = m.Role
	}

	for userID, role := range w.Members() {
		currentRole, exists := ccmm[userID.UUID()]
		if !exists || currentRole != string(role) {
			err := r.q.UpsertWorkspaceMember(ctx, dbtx, db.UpsertWorkspaceMemberParams{
				WorkspaceID: w.ID(),
				UserID:      userID,
				Role:        string(role),
			})
			if err != nil {
				return fmt.Errorf("failed to upsert member %s: %w", userID, sharedPg.ParseDBError(err))
			}
		}

		delete(ccmm, userID.UUID())
	}

	for removedUserID := range ccmm {
		err := r.q.RemoveWorkspaceMember(ctx, dbtx, db.RemoveWorkspaceMemberParams{
			WorkspaceID: w.ID(),
			UserID:      userDomain.UserID(removedUserID),
		})
		if err != nil {
			return fmt.Errorf("failed to remove member %s: %w", removedUserID, sharedPg.ParseDBError(err))
		}
	}

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, w)
}

func (r *WorkspaceRepo) FindByID(ctx context.Context, id domain.WorkspaceID) (*domain.Workspace, error) {
	dbtx := r.getDB(ctx)

	w, err := r.q.GetWorkspaceByID(ctx, dbtx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkspaceNotFound
		}

		return nil, fmt.Errorf("failed to get workspace: %w", sharedPg.ParseDBError(err))
	}

	members, err := r.q.GetWorkspaceMembers(ctx, dbtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", sharedPg.ParseDBError(err))
	}

	memberMap, err := toMemberMap(members)
	if err != nil {
		return nil, fmt.Errorf("failed to map workspace members: %w", err)
	}

	domainMemberMap := make(map[userDomain.UserID]domain.WorkspaceRole, len(memberMap))
	for uid, role := range memberMap {
		domainMemberMap[userDomain.UserID(uid)] = role
	}

	name, _ := domain.NewWorkspaceName(w.Name)
	description, _ := domain.NewWorkspaceDescription(w.Description)

	return domain.ReconstituteWorkspace(
		w.ID,
		name,
		description,
		w.CreatedAt,
		domainMemberMap,
	), nil
}

func toMemberMap(members []db.WorkspaceMembers) (map[uuid.UUID]domain.WorkspaceRole, error) {
	memberMap := make(map[uuid.UUID]domain.WorkspaceRole, len(members))
	for _, m := range members {
		role, err := domain.NewWorkspaceRole(m.Role)
		if err != nil {
			return nil, err
		}

		memberMap[m.UserID.UUID()] = role
	}

	return memberMap, nil
}

func (r *WorkspaceRepo) Delete(ctx context.Context, id domain.WorkspaceID) error {
	dbtx := r.getDB(ctx)

	ws, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.q.DeleteWorkspace(ctx, dbtx, id); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", sharedPg.ParseDBError(err))
	}

	ws.Delete()

	return sharedPg.SaveDomainEvents(ctx, r.q, dbtx, r.mapper, ws)
}
