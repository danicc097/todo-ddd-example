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

		msgChan := make(chan string, 200)

		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return // conn closed or error
				}

				msgChan <- string(msg)
			}
		}()

		message := map[string]string{"msg": "hello authorized room"}
		payload, _ := json.Marshal(message)

		require.Eventually(t, func() bool {
			_ = rdb.Publish(ctx, "test_ws:"+roomID.String(), payload).Err()

			select {
			case msg := <-msgChan:
				return strings.Contains(msg, "hello authorized room")
			default:
				return false
			}
		}, 5*time.Second, 50*time.Millisecond, "failed to receive authorized message")
	})

	t.Run("client does not receive message for unauthorized room", func(t *testing.T) {
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		defer conn.Close()

		msgChan := make(chan string, 200)

		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}

				msgChan <- string(msg)
			}
		}()

		require.Eventually(t, func() bool {
			_ = rdb.Publish(ctx, "test_ws:"+roomID.String(), []byte(`{"msg": "sync"}`)).Err()

			select {
			case msg := <-msgChan:
				return strings.Contains(msg, "sync")
			default:
				return false
			}
		}, 5*time.Second, 50*time.Millisecond, "client never became ready")

		unauthRoomID := uuid.New()
		_ = rdb.Publish(ctx, "test_ws:"+unauthRoomID.String(), []byte(`{"msg": "secret"}`)).Err()
		// we use a single PubSub conn with Subscribe, thus publishes sequentially
		_ = rdb.Publish(ctx, "test_ws:"+roomID.String(), []byte(`{"msg": "marker"}`)).Err()

		for {
			select {
			case msg := <-msgChan:
				require.NotContains(t, msg, "secret")

				if strings.Contains(msg, "marker") {
					return // success (marker sent afterwards, so must have skipped secret)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timed out waiting for marker message")
			}
		}
	})
}
