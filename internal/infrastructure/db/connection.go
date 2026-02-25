package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

func NewConnectionPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var (
		pool *pgxpool.Pool
		err  error
	)

	for i := range 15 {
		pool, err = pgxpool.New(ctx, pgURL)
		if err == nil {
			err = pool.Ping(ctx)
		}

		if err == nil {
			return pool, nil
		}

		slog.WarnContext(ctx, "Database not ready, retrying...", slog.Int("attempt", i+1))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return nil, fmt.Errorf("unable to connect to database after retries: %w", err)
}
