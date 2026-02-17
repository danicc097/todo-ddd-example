package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	container testcontainers.Container
	pool      *pgxpool.Pool
	mu        sync.Mutex
}

var (
	globalPgOnce      sync.Once
	globalPgContainer *PostgreSQLContainer
)

const schemaSQLFile = "../../sql/schema.sql"

func GetGlobalPostgresPool(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	globalPgOnce.Do(func() {
		// possibly no file lock needed, testcontainers handles Reuse concurrency,
		// and pg_advisory_lock handles schema concurrency
		globalPgContainer = NewPostgreSQLContainer(ctx, t)
		globalPgContainer.Connect(ctx, t)
	})

	return globalPgContainer.Pool()
}

func CloseGlobalPostgresPool() {
	if globalPgContainer != nil {
		globalPgContainer.Close(context.Background(), nil)
	}
}

func NewPostgreSQLContainer(ctx context.Context, t *testing.T) *PostgreSQLContainer {
	_, thisFile, _, _ := runtime.Caller(0)
	absSchemaPath, _ := filepath.Abs(filepath.Join(filepath.Dir(thisFile), schemaSQLFile))

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Name:  "todo-ddd-test-pg",
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "testpass",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		if t != nil {
			t.Fatalf("failed to start postgres container: %v", err)
		}

		panic(err)
	}

	pg := &PostgreSQLContainer{container: container}

	if !pg.isSchemaApplied(ctx) {
		pg.applySchema(ctx, t, absSchemaPath)
	}

	return pg
}

func (p *PostgreSQLContainer) isSchemaApplied(ctx context.Context) bool {
	connStr := p.ConnectionString(ctx, "sslmode=disable")

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return false
	}
	defer db.Close()

	var exists bool

	err = db.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'outbox')").Scan(&exists)

	return err == nil && exists
}

func (p *PostgreSQLContainer) applySchema(ctx context.Context, t *testing.T, schemaSQLFile string) {
	schemaBytes, err := os.ReadFile(schemaSQLFile)
	if err != nil {
		if t != nil {
			t.Fatalf("failed to read schema file: %v", err)
		}

		panic(err)
	}

	schemaSQL := string(schemaBytes)

	connStr := p.ConnectionString(ctx, "sslmode=disable")

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		if t != nil {
			t.Fatalf("failed to connect for schema apply: %v", err)
		}

		panic(err)
	}
	defer db.Close()

	// coordinates schema application across processes
	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_lock(123456789)"); err != nil {
		if t != nil {
			t.Fatalf("failed to acquire advisory lock: %v", err)
		}

		panic(err)
	}
	defer db.ExecContext(ctx, "SELECT pg_advisory_unlock(123456789)")

	var exists bool

	_ = db.QueryRowContext(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'outbox')").Scan(&exists)
	if exists {
		return
	}

	schemaSQL = strings.ReplaceAll(schemaSQL, "CREATE SCHEMA public;", "CREATE SCHEMA IF NOT EXISTS public;")

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		if t != nil {
			t.Fatalf("failed to apply schema from %s: %v", schemaSQLFile, err)
		}

		panic(err)
	}
}

func (p *PostgreSQLContainer) Connect(ctx context.Context, t *testing.T) *pgxpool.Pool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		return p.pool
	}

	connStr := p.ConnectionString(ctx, "sslmode=disable")

	var (
		pool *pgxpool.Pool
		err  error
	)

	for range 10 {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			if err := pool.Ping(ctx); err == nil {
				p.pool = pool
				return pool
			}

			pool.Close()
		}

		time.Sleep(500 * time.Millisecond)
	}

	if t != nil {
		t.Fatalf("failed to connect to postgres after retries: %v", err)
	} else {
		panic(fmt.Sprintf("failed to connect to postgres after retries: %v", err))
	}

	return nil
}

func (p *PostgreSQLContainer) Close(ctx context.Context, t *testing.T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		p.pool.Close()
	}

	if err := p.container.Terminate(ctx); err != nil {
		if t != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}
}

func (p *PostgreSQLContainer) ConnectionString(ctx context.Context, opts ...string) string {
	if pg, ok := p.container.(*postgres.PostgresContainer); ok {
		connStr, _ := pg.ConnectionString(ctx, opts...)
		return connStr
	}

	endpoint, _ := p.container.Endpoint(ctx, "")

	return fmt.Sprintf("postgresql://postgres:testpass@%s/testdb?sslmode=disable", endpoint)
}

func (p *PostgreSQLContainer) Pool() *pgxpool.Pool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.pool
}
