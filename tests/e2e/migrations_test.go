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

func TestMigration_05_Backfill(t *testing.T) {
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

	targetMigration := "migrations/05_tag_workspace.json"
	allMigrations := getSortedMigrations(t)

	targetFound := false
	for _, mig := range allMigrations {
		if mig == targetMigration {
			targetFound = true
			break
		}
		runPgroll(t, pgURL, "start", mig)
	}
	require.True(t, targetFound, "Target migration %s not found in migrations directory", targetMigration)

	userID := uuid.New()
	workspaceID := uuid.New()
	tagID := uuid.New()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, email, name) VALUES ($1, 'test@test.com', 'test')`, userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO workspaces (id, name, description) VALUES ($1, 'Test Workspace', 'Desc')`, workspaceID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES ($1, 'Urgent')`, tagID)
	require.NoError(t, err)

	runPgroll(t, pgURL, "start", targetMigration)

	// note this backfilling is nonsensical (grabs ID of first workspace and assigns all tags),
	// but serves as a baseline for other backfilling tests.
	var wID uuid.UUID
	err = db.QueryRowContext(ctx, `SELECT workspace_id FROM tags WHERE id = $1`, tagID).Scan(&wID)
	require.NoError(t, err)

	assert.Equal(t, workspaceID, wID)
}
