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
		Name:        w.Name(),
		Description: w.Description(),
		CreatedAt:   w.CreatedAt(),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert workspace: %w", sharedPg.ParseDBError(err))
	}

	// diffing not worth it
	if err := r.q.DeleteWorkspaceMembers(ctx, dbtx, w.ID()); err != nil {
		return fmt.Errorf("failed to clear members: %w", sharedPg.ParseDBError(err))
	}

	for userID, role := range w.Members() {
		err := r.q.AddWorkspaceMember(ctx, dbtx, db.AddWorkspaceMemberParams{
			WorkspaceID: w.ID(),
			UserID:      userID,
			Role:        string(role),
		})
		if err != nil {
			return fmt.Errorf("failed to add member %s: %w", userID, sharedPg.ParseDBError(err))
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
		domainMemberMap[userDomain.UserID{UUID: uid}] = role
	}

	return domain.ReconstituteWorkspace(
		w.ID,
		w.Name,
		w.Description,
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

		memberMap[m.UserID.UUID] = role
	}

	return memberMap, nil
}

func (r *WorkspaceRepo) FindAll(ctx context.Context) ([]*domain.Workspace, error) {
	dbtx := r.getDB(ctx)

	rows, err := r.q.ListWorkspaces(ctx, dbtx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", sharedPg.ParseDBError(err))
	}

	workspaces := make([]*domain.Workspace, 0, len(rows))

	for _, row := range rows {
		// N+1 for now, should batch
		members, err := r.q.GetWorkspaceMembers(ctx, dbtx, row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch members for workspace %s: %w", row.ID, sharedPg.ParseDBError(err))
		}

		memberMap, _ := toMemberMap(members)

		domainMemberMap := make(map[userDomain.UserID]domain.WorkspaceRole, len(memberMap))
		for uid, role := range memberMap {
			domainMemberMap[userDomain.UserID{UUID: uid}] = role
		}

		workspaces = append(workspaces, domain.ReconstituteWorkspace(
			row.ID,
			row.Name,
			row.Description,
			row.CreatedAt,
			domainMemberMap,
		))
	}

	return workspaces, nil
}

func (r *WorkspaceRepo) Delete(ctx context.Context, id domain.WorkspaceID) error {
	dbtx := r.getDB(ctx)

	if err := r.q.DeleteWorkspace(ctx, dbtx, id); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", sharedPg.ParseDBError(err))
	}

	return nil
}
