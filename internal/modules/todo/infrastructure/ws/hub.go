/**
 * See https://github.com/gorilla/websocket/blob/main/examples/chat
 */

package ws

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/danicc097/todo-ddd-example/internal/shared/causation"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline  = []byte{'\n'}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// PermissionProvider fetches authorized rooms for a user.
type PermissionProvider func(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)

// MessageFilter extracts a room ID from a message.
type MessageFilter func(message []byte) (uuid.UUID, bool)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	rooms  map[uuid.UUID]bool
	userID uuid.UUID
}

func (c *Client) readPump() {
	defer func() {
		select {
		case c.hub.unregister <- c:
		case <-c.hub.stop:
		}

		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// must read regardless of broadcast-only or not
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(message)

			for range len(c.send) {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.hub.stop:
			return
		}
	}
}

// Hub maintains the set of active clients and broadcasts messages aware of rooms/permissions.
type Hub struct {
	rooms        map[uuid.UUID]map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	broadcast    chan []byte
	redis        *redis.Client
	channelName  string
	permProvider PermissionProvider
	msgFilter    MessageFilter
	stop         chan struct{}
	wg           sync.WaitGroup
}

func NewHub(r *redis.Client, channel string, pp PermissionProvider, mf MessageFilter) *Hub {
	h := &Hub{
		rooms:        make(map[uuid.UUID]map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan []byte),
		redis:        r,
		channelName:  channel,
		permProvider: pp,
		msgFilter:    mf,
		stop:         make(chan struct{}),
	}

	h.wg.Add(2)

	go h.run()
	go h.listenRedis()

	return h
}

func (h *Hub) Shutdown() {
	close(h.stop)
	h.wg.Wait()
}

func (h *Hub) run() {
	defer h.wg.Done()

	for {
		select {
		case <-h.stop:
			return
		case client := <-h.register:
			for roomID := range client.rooms {
				if h.rooms[roomID] == nil {
					h.rooms[roomID] = make(map[*Client]bool)
				}

				h.rooms[roomID][client] = true
			}
		case client := <-h.unregister:
			for roomID := range client.rooms {
				if clients, ok := h.rooms[roomID]; ok {
					delete(clients, client)

					if len(clients) == 0 {
						delete(h.rooms, roomID)
					}
				}
			}

			close(client.send)
		case message := <-h.broadcast:
			roomID, ok := h.msgFilter(message)
			if !ok {
				continue
			}

			if clients, ok := h.rooms[roomID]; ok {
				for client := range clients {
					select {
					case client.send <- message: // happy path
					default: // queue is full
						select {
						case <-client.send: // drop oldest message if buffer is full
						default:
						}

						select {
						case client.send <- message: // try again
						default: // still full
							client.conn.Close()
						}
					}
				}
			}
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	meta := causation.FromContext(r.Context())
	if !meta.IsUser() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// restrict authorized rooms per user
	rooms, err := h.permProvider(r.Context(), meta.UserID)
	if err != nil {
		conn.Close()
		return
	}

	roomMap := make(map[uuid.UUID]bool)
	for _, id := range rooms {
		roomMap[id] = true
	}

	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 1024),
		rooms:  roomMap,
		userID: meta.UserID,
	}

	select {
	case h.register <- client:
		go client.writePump()
		go client.readPump()
	case <-h.stop:
		conn.Close()
	}
}

func (h *Hub) listenRedis() {
	defer h.wg.Done()

	// we control termination via h.stop not ctx
	pubsub := h.redis.Subscribe(context.Background(), h.channelName)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-h.stop:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			select {
			case h.broadcast <- []byte(msg.Payload):
			case <-h.stop:
				return
			}
		}
	}
}
