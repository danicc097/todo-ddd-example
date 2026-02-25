package ws_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestHub_Behavioral(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rdb := testutils.GetGlobalRedis(t).Connect(ctx, t)

	roomID := uuid.New()
	userID := uuid.New()

	permProvider := func(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error) {
		if uid == userID {
			return []uuid.UUID{roomID}, nil
		}

		return nil, nil
	}

	hub := ws.NewHub(rdb, permProvider, ws.Config{
		GlobalChannel:          "test_global",
		WorkspaceChannelPrefix: "test_ws:",
	})
	defer hub.Shutdown()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meta := causation.Metadata{UserID: userID}
		r = r.WithContext(causation.WithMetadata(r.Context(), meta))
		hub.HandleWebSocket(w, r)
	}))
	defer server.Close()

	t.Run("client receives message for authorized room", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		defer conn.Close()

		time.Sleep(300 * time.Millisecond) // TODO: everntually

		message := map[string]string{"msg": "hello authorized room"}
		payload, _ := json.Marshal(message)

		require.Eventually(t, func() bool {
			_ = rdb.Publish(ctx, "test_ws:"+roomID.String(), payload).Err()

			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			_, msg, err := conn.ReadMessage()
			if err != nil {
				return false
			}

			return strings.Contains(string(msg), "hello authorized room")
		}, 5*time.Second, 200*time.Millisecond)
	})

	t.Run("client does not receive message for unauthorized room", func(t *testing.T) {
		otherRoomID := uuid.New()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		defer conn.Close()

		payload := []byte(`{"msg": "secret"}`)

		for range 5 {
			_ = rdb.Publish(ctx, "test_ws:"+otherRoomID.String(), payload).Err()
		}

		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, _, err = conn.ReadMessage()

		require.Error(t, err, "expected timeout as no message should arrive")
		assert.True(t, websocket.IsUnexpectedCloseError(err) || strings.Contains(err.Error(), "timeout"))
	})
}
