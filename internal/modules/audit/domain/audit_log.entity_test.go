package domain_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/audit/domain"
)

func TestAuditLog_Instantiation(t *testing.T) {
	t.Parallel()

	t.Run("valid instantiation", func(t *testing.T) {
		corrID := uuid.NewString()
		causID := uuid.NewString()
		actorID := uuid.New()
		ip := "127.0.0.1"
		ua := "Mozilla/5.0"
		aggID := uuid.New()
		changes := map[string]any{"field": "new"}

		log, err := domain.NewAuditLog(
			corrID,
			causID,
			&actorID,
			ip,
			ua,
			domain.AggTodo,
			aggID,
			domain.OpCreate,
			changes,
		)

		require.NoError(t, err)
		assert.Equal(t, corrID, log.CorrelationID())
		assert.Equal(t, causID, log.CausationID())
		assert.Equal(t, &actorID, log.ActorID())
		assert.Equal(t, ip, log.ActorIP())

		expectedUAHash := sha256.Sum256([]byte(ua))
		assert.Equal(t, hex.EncodeToString(expectedUAHash[:]), log.UserAgentHash())

		assert.Equal(t, domain.AggTodo.String(), log.AggregateType())
		assert.Equal(t, aggID, log.AggregateID())
		assert.Equal(t, domain.OpCreate.String(), log.Operation())
		assert.Equal(t, changes, log.Changes())
		assert.NotZero(t, log.OccurredAt())
	})

	t.Run("invalid aggregate type", func(t *testing.T) {
		_, err := domain.ParseAuditAggregateType("INVALID")
		assert.Error(t, err)
	})

	t.Run("missing correlation id", func(t *testing.T) {
		_, err := domain.NewAuditLog(
			"",
			"causation",
			nil,
			"ip",
			"ua",
			domain.AggTodo,
			uuid.New(),
			domain.OpCreate,
			nil,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires correlation_id")
	})
}
