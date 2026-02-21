package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppConfig(t *testing.T) {
	t.Run("correct env", func(t *testing.T) {
		t.Setenv("DB_USER", "user")
		t.Setenv("DB_PASS", "pass")
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("DB_NAME", "db")
		t.Setenv("REDIS_ADDR", "localhost:6379")
		t.Setenv("RABBITMQ_URL", "amqp://localhost")
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
		t.Setenv("LOG_LEVEL", "DEBUG")
		t.Setenv("ENV", "ci")
		t.Setenv("PORT", "8080")
		t.Setenv("MFA_MASTER_KEY", "masterkey")

		cfg, err := LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, "user", cfg.Postgres.User)
		assert.Equal(t, "pass", cfg.Postgres.Password)
		assert.Equal(t, "localhost", cfg.Postgres.Host)
		assert.Equal(t, "5432", cfg.Postgres.Port)
		assert.Equal(t, "db", cfg.Postgres.DBName)
		assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
		assert.Equal(t, "amqp://localhost", cfg.RabbitMQ.URL)
		assert.Equal(t, "localhost:4317", cfg.OTEL.Endpoint)
		assert.Equal(t, "DEBUG", cfg.LogLevel)
		assert.Equal(t, AppEnvCI, cfg.Env)
		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "masterkey", cfg.MFAMasterKey)
	})

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("DB_USER", "")
		t.Setenv("LOG_LEVEL", "")
		t.Setenv("ENV", "")

		cfg, err := LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, "INFO", cfg.LogLevel)
		assert.Equal(t, AppEnvDev, cfg.Env)
	})
}
