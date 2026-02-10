package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type TodoHub struct {
	clients map[*websocket.Conn]bool
	mutex   sync.Mutex
	redis   *redis.Client
}

func NewTodoHub(r *redis.Client) *TodoHub {
	hub := &TodoHub{
		clients: make(map[*websocket.Conn]bool),
		redis:   r,
	}
	go hub.listenRedis()

	return hub
}

func (h *TodoHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrade error: %v\n", err)
		return
	}

	h.mutex.Lock()
	h.clients[conn] = true
	h.mutex.Unlock()
}

func (h *TodoHub) listenRedis() {
	pubsub := h.redis.Subscribe(context.Background(), "todo_updates")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		h.mutex.Lock()

		for client := range h.clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				client.Close()
				delete(h.clients, client)
			}
		}

		h.mutex.Unlock()
	}
}
