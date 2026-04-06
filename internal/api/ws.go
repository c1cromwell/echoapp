package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage represents a message sent over WebSocket.
type WSMessage struct {
	Type           string          `json:"type"`                      // "text", "control"
	From           string          `json:"from,omitempty"`            // sender user ID
	To             string          `json:"to,omitempty"`              // recipient user ID (empty = broadcast)
	ConversationID string          `json:"conversation_id,omitempty"` // optional conversation scope
	Payload        json.RawMessage `json:"payload"`                   // message content
	Timestamp      string          `json:"timestamp"`
}

// WSControlMessage represents a control action (ping, subscribe, etc.).
type WSControlMessage struct {
	Action string            `json:"action"`
	Data   map[string]string `json:"data,omitempty"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// Hub manages all active WebSocket connections and routes messages.
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client // userID -> client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop. Call this in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full — drop and disconnect
					close(client.send)
					delete(h.clients, client.userID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser delivers a message to a specific user if connected.
// Returns true if the user was found and the message was queued.
func (h *Hub) SendToUser(userID string, data []byte) bool {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case client.send <- data:
		return true
	default:
		return false
	}
}

// ConnectedUsers returns the list of currently connected user IDs.
func (h *Hub) ConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]string, 0, len(h.clients))
	for id := range h.clients {
		users = append(users, id)
	}
	return users
}

// --- WebSocket HTTP handler ---

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for development/testing
	},
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 65536
)

// ServeWS handles WebSocket upgrade requests.
// The user ID is extracted from the "Authorization: Bearer <token>" header
// using the same logic as the REST auth middleware.
func ServeWS(hub *Hub, userIDExtractor func(token string) string, w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth header
	token := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := splitBearer(authHeader)
		if parts != "" {
			token = parts
		}
	}

	// Also accept token as query param for easier client usage
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	if token == "" {
		http.Error(w, "authorization required", http.StatusUnauthorized)
		return
	}

	userID := userIDExtractor(token)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func splitBearer(header string) string {
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse incoming message
		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			// Treat as plain text
			msg = WSMessage{
				Type:      "text",
				From:      c.userID,
				Payload:   raw,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
		} else {
			msg.From = c.userID
			if msg.Timestamp == "" {
				msg.Timestamp = time.Now().UTC().Format(time.RFC3339)
			}
		}

		// Handle control messages
		if msg.Type == "control" {
			c.handleControl(msg)
			continue
		}

		outBytes, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		// Route: if To is set, send to specific user; otherwise broadcast
		if msg.To != "" {
			c.hub.SendToUser(msg.To, outBytes)
		} else {
			c.hub.broadcast <- outBytes
		}
	}
}

// handleControl processes control messages (ping/pong, subscribe, etc.).
func (c *Client) handleControl(msg WSMessage) {
	var ctrl WSControlMessage
	if err := json.Unmarshal(msg.Payload, &ctrl); err != nil {
		return
	}

	switch ctrl.Action {
	case "ping":
		// Respond with pong
		pong := WSMessage{
			Type:      "control",
			From:      "server",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		pongPayload, _ := json.Marshal(WSControlMessage{Action: "pong"})
		pong.Payload = pongPayload
		data, _ := json.Marshal(pong)
		select {
		case c.send <- data:
		default:
		}
	}
}

// writePump writes messages from the send channel to the WebSocket connection.
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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
