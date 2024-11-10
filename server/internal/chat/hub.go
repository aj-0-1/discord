package chat

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

type Hub struct {
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	redis      *redis.Client
	log        *zerolog.Logger
	mu         sync.RWMutex
}

type Client struct {
	hub    *Hub
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
}

func NewHub(redis *redis.Client, log *zerolog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      redis,
		log:        log,
	}
}

func (h *Hub) Run() {
	pubsub := h.redis.PSubscribe(context.Background(), "user:*:messages")
	defer pubsub.Close()

	// Handle Redis messages
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var message Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				h.log.Error().Err(err).
					Str("payload", msg.Payload).
					Msg("failed to unmarshal message")
				continue
			}

			h.mu.RLock()
			if clients, ok := h.clients[message.ToID.String()]; ok {
				// Marshal the message for WebSocket delivery
				payload, err := json.Marshal(message)
				if err != nil {
					h.log.Error().Err(err).Msg("failed to marshal message for websocket")
					h.mu.RUnlock()
					continue
				}

				for client := range clients {
					select {
					case client.send <- payload:
						h.log.Debug().
							Str("toId", message.ToID.String()).
							Msg("message sent to websocket client")
					default:
						h.mu.RUnlock()
						h.unregister <- client
						h.mu.RLock()
					}
				}
			}
			h.mu.RUnlock()
		}
	}()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				c.hub.log.Error().Err(err).Msg("websocket read error")
			}
			break
		}
		// We only use server-to-client communication
		// Client-to-server goes through regular HTTP endpoints
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
