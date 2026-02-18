package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsApp "github.com/danicc097/todo-ddd-example/internal/modules/workspace/application"
	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	hub        *TodoHub
	conn       *websocket.Conn
	send       chan []byte
	uid        userDomain.UserID
	workspaces map[uuid.UUID]bool
}

func (c *Client) writePump() {
	defer func() {
		c.hub.unregister <- c

		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

type TodoHub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	redis      *redis.Client
	wsQuery    wsApp.WorkspaceQueryService
}

func NewTodoHub(r *redis.Client, wsQuery wsApp.WorkspaceQueryService) *TodoHub {
	hub := &TodoHub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      r,
		wsQuery:    wsQuery,
	}
	go hub.run()
	go hub.listenRedis()

	return hub
}

func (h *TodoHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		}
	}
}

func (h *TodoHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	meta := causation.FromContext(r.Context())
	if !meta.IsUser() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Upgrade error", slog.String("error", err.Error()))
		return
	}

	workspaces, err := h.wsQuery.ListByUserID(r.Context(), userDomain.UserID{UUID: meta.UserID})
	if err != nil {
		http.Error(w, "Could not fetch permissions", http.StatusInternalServerError)
		return
	}

	wsMap := make(map[uuid.UUID]bool)
	for _, ws := range workspaces {
		wsMap[ws.Id.UUID] = true
	}

	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, 256),
		uid:        userDomain.UserID{UUID: meta.UserID},
		workspaces: wsMap,
	}
	h.register <- client

	go client.writePump()
}

func (h *TodoHub) listenRedis() {
	pubsub := h.redis.Subscribe(context.Background(), "todo_updates")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var payload struct {
			WorkspaceID uuid.UUID `json:"workspace_id"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}

		wsID := payload.WorkspaceID

		for client := range h.clients {
			if client.workspaces[wsID] {
				select {
				case client.send <- []byte(msg.Payload):
				default:
					// slow consumer, unregister
					h.unregister <- client
				}
			}
		}
	}
}
