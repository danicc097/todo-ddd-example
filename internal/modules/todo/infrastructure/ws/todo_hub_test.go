package ws_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/danicc097/todo-ddd-example/internal/infrastructure/cache"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/infrastructure/ws"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/application/applicationfakes"
	wsdomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

type wsReceiver struct {
	mu           sync.Mutex
	workspaceIDs []uuid.UUID
}

func (r *wsReceiver) has(id uuid.UUID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Contains(r.workspaceIDs, id)
}

func TestTodoHub_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	redisClient := testutils.GetGlobalRedis(t).Connect(ctx, t)

	wsID1 := uuid.New()
	wsID2 := uuid.New()
	userID := uuid.New()

	mockWS := &applicationfakes.FakeWorkspaceQueryService{
		ListByUserIDStub: func(ctx context.Context, ui userDomain.UserID) ([]wsApp.WorkspaceReadModel, error) {
			return []wsApp.WorkspaceReadModel{
				{ID: wsdomain.WorkspaceID(wsID1)},
			}, nil
		},
	}

	setupHub := func() (*ws.Hub, *httptest.Server) {
		hub := ws.NewTodoHub(redisClient, mockWS)
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			meta := causation.Metadata{UserID: userID}
			r = r.WithContext(causation.WithMetadata(r.Context(), meta))
			hub.HandleWebSocket(w, r)
		}))

		return hub, s
	}

	dialAndRecord := func(t *testing.T, serverURL string) (*websocket.Conn, *wsReceiver) {
		t.Helper()

		wsURL := "ws" + strings.TrimPrefix(serverURL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		rec := &wsReceiver{}

		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}

				var envelope struct {
					Data struct {
						WorkspaceID uuid.UUID `json:"workspace_id"`
					} `json:"data"`
				}

				if err := json.Unmarshal(message, &envelope); err == nil {
					rec.mu.Lock()
					rec.workspaceIDs = append(rec.workspaceIDs, envelope.Data.WorkspaceID)
					rec.mu.Unlock()
				}
			}
		}()

		return conn, rec
	}

	t.Run("receives message for allowed workspace", func(t *testing.T) {
		hub, s := setupHub()
		defer hub.Shutdown()
		defer s.Close()

		conn, rec := dialAndRecord(t, s.URL)
		defer conn.Close()

		data, _ := json.Marshal(map[string]any{
			"event":     string(sharedDomain.TodoCreated),
			"timestamp": time.Now(),
			"data": map[string]any{
				"workspace_id": wsID1,
				"id":           uuid.New(),
			},
		})

		require.Eventually(t, func() bool {
			_ = redisClient.Publish(ctx, cache.Keys.TodoAPIUpdatesChannel(), data).Err()
			return rec.has(wsID1)
		}, 5*time.Second, 100*time.Millisecond)
	})

	t.Run("does not receive message for forbidden workspace", func(t *testing.T) {
		hub, s := setupHub()
		defer hub.Shutdown()
		defer s.Close()

		conn, rec := dialAndRecord(t, s.URL)
		defer conn.Close()

		forbiddenData, _ := json.Marshal(map[string]any{
			"event":     string(sharedDomain.TodoCreated),
			"timestamp": time.Now(),
			"data": map[string]any{
				"workspace_id": wsID2,
				"id":           uuid.New(),
			},
		})

		allowedData, _ := json.Marshal(map[string]any{
			"event":     string(sharedDomain.TodoCreated),
			"timestamp": time.Now(),
			"data": map[string]any{
				"workspace_id": wsID1,
				"id":           uuid.New(),
			},
		})

		require.Eventually(t, func() bool {
			// processed sequentially, ensure forbidden sent first
			_ = redisClient.Publish(ctx, cache.Keys.TodoAPIUpdatesChannel(), forbiddenData).Err()
			_ = redisClient.Publish(ctx, cache.Keys.TodoAPIUpdatesChannel(), allowedData).Err()

			return rec.has(wsID1)
		}, 5*time.Second, 100*time.Millisecond)

		assert.False(t, rec.has(wsID2))
	})

	t.Run("unauthorized connection rejected", func(t *testing.T) {
		hub, sOrig := setupHub()
		defer hub.Shutdown()
		defer sOrig.Close()

		sNoAuth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub.HandleWebSocket(w, r)
		}))
		defer sNoAuth.Close()

		urlAuth := "ws" + strings.TrimPrefix(sNoAuth.URL, "http")
		_, resp, err := websocket.DefaultDialer.Dial(urlAuth, nil)

		require.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
