//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func runPgroll(t *testing.T, pgURL string, command string, file string) {
	t.Helper()
	args := []string{"tool", "pgroll", "--postgres-url", pgURL, command}
	if file != "" {
		args = append(args, "--complete", file)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = "../../"
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "pgroll command failed: %s\nOutput: %s", command, string(out))
}

func getSortedMigrations(t *testing.T) []string {
	t.Helper()

	files, err := filepath.Glob("../../migrations/*.json")
	require.NoError(t, err, "failed to read migrations directory")

	sort.Strings(files)

	var normalized []string
	for _, f := range files {
		norm := strings.TrimPrefix(filepath.ToSlash(f), "../../")
		normalized = append(normalized, norm)
	}

	return normalized
}

type migrationHook struct {
	Before func(t *testing.T, ctx context.Context, db *sql.DB)
	After  func(t *testing.T, ctx context.Context, db *sql.DB)
}

func TestMigrations_E2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	pgURL, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", pgURL)
	require.NoError(t, err)
	defer db.Close()

	runPgroll(t, pgURL, "init", "")

	var (
		testWorkspaceID = uuid.New()
		testUserID      = uuid.New()
		testTagID       = uuid.New()
		testTodoID      = uuid.New()
	)

	hooks := map[string]migrationHook{
		"05_tag_workspace.json": { // tags now are assigned to workspaces

			Before: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, err := db.ExecContext(ctx, `INSERT INTO users (id, email, name) VALUES ($1, 'test@test.com', 'test')`, testUserID)
				require.NoError(t, err)

				_, err = db.ExecContext(ctx, `INSERT INTO workspaces (id, name, description) VALUES ($1, 'Test Workspace', 'Desc')`, testWorkspaceID)
				require.NoError(t, err)

				_, err = db.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES ($1, 'Urgent')`, testTagID)
				require.NoError(t, err)
			},
			After: func(t *testing.T, ctx context.Context, db *sql.DB) {
				var wID uuid.UUID
				err := db.QueryRowContext(ctx, `SELECT workspace_id FROM tags WHERE id = $1`, testTagID).Scan(&wID)
				require.NoError(t, err)

				assert.Equal(t, testWorkspaceID, wID)
			},
		},
		"07_todo_workspace.json": { // todos now belong to workspaces
			Before: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, err := db.ExecContext(ctx, `INSERT INTO todos (id, title, status) VALUES ($1, 'Legacy Todo', 'PENDING')`, testTodoID)
				require.NoError(t, err)
			},
			After: func(t *testing.T, ctx context.Context, db *sql.DB) {
				var wID uuid.UUID
				err := db.QueryRowContext(ctx, `SELECT workspace_id FROM todos WHERE id = $1`, testTodoID).Scan(&wID)
				require.NoError(t, err)

				assert.Equal(t, testWorkspaceID, wID)
			},
		},
	}

	allMigrations := getSortedMigrations(t)
	require.NotEmpty(t, allMigrations)

	for _, mig := range allMigrations {
		baseName := filepath.Base(mig)

		success := t.Run(baseName, func(t *testing.T) {
			hook, hasHook := hooks[baseName]

			if hasHook && hook.Before != nil {
				hook.Before(t, ctx, db)
			}

			runPgroll(t, pgURL, "start", mig)

			if hasHook && hook.After != nil {
				hook.After(t, ctx, db)
			}
		})

		if !success {
			t.Fatalf("migration %s failed", baseName)
		}
	}
}
