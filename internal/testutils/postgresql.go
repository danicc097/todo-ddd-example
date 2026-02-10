package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQLContainer struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
}

// generated via make gen-schema
const schemaSQLFile = "../../../../../sql/schema.sql"

func NewPostgreSQLContainer(ctx context.Context, t *testing.T) *PostgreSQLContainer {
	t.Helper()

	absSchemaPath, err := filepath.Abs(schemaSQLFile)
	if err != nil {
		t.Fatalf("failed to get absolute path for schema file: %v", err)
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"), // use role from generated schema
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	pg := &PostgreSQLContainer{container: container}

	pg.applySchema(ctx, t, absSchemaPath)

	return pg
}

func (p *PostgreSQLContainer) applySchema(ctx context.Context, t *testing.T, schemaSQLFile string) {
	t.Helper()

	schemaSQL, err := os.ReadFile(schemaSQLFile)
	if err != nil {
		t.Fatalf("failed to read schema file %s: %v", schemaSQLFile, err)
	}

	connStr, err := p.container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string for schema apply: %v", err)
	}

	var db *sql.DB
	for range 10 {
		db, err = sql.Open("pgx", connStr)
		if err == nil {
			err = db.PingContext(ctx)
		}
		if err == nil {
			break
		}
		if db != nil {
			db.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("failed to connect to database for schema apply after retries: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, "DROP SCHEMA IF EXISTS public CASCADE"); err != nil {
		t.Fatalf("failed to drop public schema: %v", err)
	}

	if _, err := db.ExecContext(ctx, string(schemaSQL)); err != nil {
		t.Fatalf("failed to apply schema from %s: %v", schemaSQLFile, err)
	}
}

func (p *PostgreSQLContainer) Connect(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()

	var pool *pgxpool.Pool
	var err error

	for range 10 {
		connStr, err := p.container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			t.Fatalf("failed to get connection string: %v", err)
		}

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

	t.Fatalf("failed to connect to postgres after retries: %v", err)
	return nil
}

func (p *PostgreSQLContainer) Close(ctx context.Context, t *testing.T) {
	t.Helper()

	if p.pool != nil {
		p.pool.Close()
	}

	if err := p.container.Terminate(ctx); err != nil {
		t.Logf("failed to terminate postgres container: %v", err)
	}
}

func (p *PostgreSQLContainer) ConnectionString(ctx context.Context, opts ...string) string {
	connStr, err := p.container.ConnectionString(ctx, opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}
	return connStr
}

func (p *PostgreSQLContainer) Pool() *pgxpool.Pool {
	return p.pool
}
