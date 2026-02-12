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
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type WorkspaceRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewWorkspaceRepo(pool *pgxpool.Pool) *WorkspaceRepo {
	return &WorkspaceRepo{
		q:    db.New(),
		pool: pool,
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

	// Note: In a real production system, 'Save' usually handles Upsert (Insert or Update).
	// Since we are using specific SQL queries, we assume this is a Create operation
	// or that the ID generation strategy prevents collisions.
	_, err := r.q.CreateWorkspace(ctx, dbtx, db.CreateWorkspaceParams{
		ID:          w.ID(),
		Name:        w.Name(),
		Description: w.Description(),
		CreatedAt:   w.CreatedAt(),
	})
	if err != nil {
		return fmt.Errorf("failed to save workspace: %w", err)
	}

	// Save Members
	// In a real generic Save, we would need to handle diffs (detect removed members).
	// For this implementation, we insert the current state.
	for userID, role := range w.Members() {
		err := r.q.AddWorkspaceMember(ctx, dbtx, db.AddWorkspaceMemberParams{
			WorkspaceID: w.ID(),
			UserID:      userID,
			Role:        string(role),
		})
		if err != nil {
			return fmt.Errorf("failed to save workspace member %s: %w", userID, err)
		}
	}

	return nil
}

func (r *WorkspaceRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	dbtx := r.getDB(ctx)

	w, err := r.q.GetWorkspaceByID(ctx, dbtx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	members, err := r.q.GetWorkspaceMembers(ctx, dbtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace members: %w", err)
	}

	memberMap := make(map[uuid.UUID]domain.WorkspaceRole, len(members))
	for _, m := range members {
		role, err := domain.NewWorkspaceRole(m.Role)
		if err != nil {
			return nil, err
		}
		memberMap[m.UserID] = role
	}

	return domain.ReconstituteWorkspace(
		w.ID,
		w.Name,
		w.Description,
		w.CreatedAt,
		memberMap,
	), nil
}

func (r *WorkspaceRepo) FindAll(ctx context.Context) ([]*domain.Workspace, error) {
	dbtx := r.getDB(ctx)

	rows, err := r.q.ListWorkspaces(ctx, dbtx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	workspaces := make([]*domain.Workspace, 0, len(rows))

	for _, row := range rows {
		// N+1 Query: In a high-throughput scenario, fetch all members in one batch query
		// using 'WHERE workspace_id IN (...)' and map them in memory.
		// For this implementation, we query per workspace to ensure Aggregate consistency.
		members, err := r.q.GetWorkspaceMembers(ctx, dbtx, row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch members for workspace %s: %w", row.ID, err)
		}

		memberMap := make(map[uuid.UUID]domain.WorkspaceRole, len(members))
		for _, m := range members {
			role, _ := domain.NewWorkspaceRole(m.Role)
			memberMap[m.UserID] = role
		}

		workspaces = append(workspaces, domain.ReconstituteWorkspace(
			row.ID,
			row.Name,
			row.Description,
			row.CreatedAt,
			memberMap,
		))
	}

	return workspaces, nil
}

func (r *WorkspaceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	dbtx := r.getDB(ctx)

	// Cascading delete in DB handles members automatically
	if err := r.q.DeleteWorkspace(ctx, dbtx, id); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	return nil
}
