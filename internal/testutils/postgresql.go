package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	container testcontainers.Container
	baseURI   string
	adminPool *pgxpool.Pool
}

var (
	globalPgOnce      sync.Once
	globalPgContainer *PostgreSQLContainer
	globalPgErr       error
)

const schemaSQLFile = "../../sql/schema.sql"

func GetGlobalPostgresPool(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	globalPgOnce.Do(func() {
		globalPgContainer, globalPgErr = newPostgreSQLContainer(ctx)
	})

	if globalPgErr != nil {
		t.Fatalf("Failed to initialize global postgres container: %v", globalPgErr)
	}

	return globalPgContainer.CreateTestDatabase(ctx, t)
}

func newPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {
	// prevents from killing db when the first package finishes testing
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	_ = os.Setenv("TESTCONTAINERS_REUSE_ENABLE", "true")

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Name:  "todo-ddd-test-pg",
		Env: map[string]string{
			"POSTGRES_DB":       "postgres",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "testpass",
		},
		Labels: map[string]string{
			"todo-ddd-test": "true", // cleanup watchdog
		},
		Cmd: []string{
			"-c", "max_connections=500",
			"-c", "fsync=off",
			"-c", "synchronous_commit=off",
			"-c", "full_page_writes=off",
			"-c", "shared_buffers=128MB",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.ShmSize = 256 * 1024 * 1024
		},
	}

	containerInstance, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, _ := containerInstance.Host(ctx)
	port, _ := containerInstance.MappedPort(ctx, "5432")
	baseURI := fmt.Sprintf("postgresql://postgres:testpass@%s:%s", host, port.Port())

	// for lock free CREATE/DROP db
	adminPool, err := pgxpool.New(ctx, baseURI+"/postgres?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to open admin pool: %w", err)
	}

	pg := &PostgreSQLContainer{
		container: containerInstance,
		baseURI:   baseURI,
		adminPool: adminPool,
	}

	if err := pg.cleanupOrphanedDatabases(ctx); err != nil {
		return nil, fmt.Errorf("failed to clean orphaned databases: %w", err)
	}

	if err := pg.prepareTemplate(ctx); err != nil {
		return nil, err
	}

	return pg, nil
}

func (p *PostgreSQLContainer) cleanupOrphanedDatabases(ctx context.Context) error {
	if _, err := p.adminPool.Exec(ctx, "SELECT pg_advisory_lock(999999999)"); err != nil {
		return err
	}
	defer p.adminPool.Exec(ctx, "SELECT pg_advisory_unlock(999999999)")

	rows, err := p.adminPool.Query(ctx, "SELECT datname FROM pg_database WHERE datname LIKE 'test_%'")
	if err != nil {
		return err
	}

	var orphanDBs []string

	now := time.Now()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			parts := strings.Split(name, "_") // test_<timestamp>_<uuid>
			if len(parts) >= 2 {
				var ts int64
				if _, err := fmt.Sscanf(parts[1], "%d", &ts); err == nil {
					if time.Unix(ts, 0).Add(5 * time.Minute).Before(now) {
						orphanDBs = append(orphanDBs, name)
					}
				}
			}
		}
	}

	rows.Close()

	for _, name := range orphanDBs {
		_, _ = p.adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", name))
	}

	return nil
}

func (p *PostgreSQLContainer) prepareTemplate(ctx context.Context) error {
	_, err := p.adminPool.Exec(ctx, "SELECT pg_advisory_lock(123456789)")
	if err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	defer p.adminPool.Exec(ctx, "SELECT pg_advisory_unlock(123456789)")

	var exists bool

	_ = p.adminPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_database WHERE datname = 'template_db')").Scan(&exists)
	if exists {
		return nil
	}

	if _, err := p.adminPool.Exec(ctx, "CREATE DATABASE template_db;"); err != nil {
		return fmt.Errorf("failed to create template_db: %w", err)
	}

	templatePool, err := pgxpool.New(ctx, p.baseURI+"/template_db?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to open template db pool: %w", err)
	}
	defer templatePool.Close()

	_, thisFile, _, _ := runtime.Caller(0)
	absSchemaPath, _ := filepath.Abs(filepath.Join(filepath.Dir(thisFile), schemaSQLFile))

	schemaBytes, err := os.ReadFile(absSchemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	schemaSQL := strings.ReplaceAll(string(schemaBytes), "CREATE SCHEMA public;", "CREATE SCHEMA IF NOT EXISTS public;")

	if _, err := templatePool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("failed to apply schema to template_db: %w", err)
	}

	return nil
}

func (p *PostgreSQLContainer) CreateTestDatabase(ctx context.Context, t *testing.T) *pgxpool.Pool {
	// so cleanup can distinguish new active dbs
	testDBName := fmt.Sprintf("test_%d_%s", time.Now().Unix(), strings.ReplaceAll(uuid.New().String(), "-", "")[:12])

	// 100% Lock-Free execution for concurrent tests using the shared admin pool
	if _, err := p.adminPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE template_db;", testDBName)); err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	testConnStr := p.baseURI + "/" + testDBName + "?sslmode=disable"

	var (
		pool *pgxpool.Pool
		err  error
	)

	for range 50 {
		pool, err = pgxpool.New(ctx, testConnStr)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				break
			} else {
				err = pingErr

				pool.Close()
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("failed to connect to test db after retries: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()

		_, _ = p.adminPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", testDBName))
	})

	return pool
}
