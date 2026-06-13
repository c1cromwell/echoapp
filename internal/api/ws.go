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

// ephemeralSignalTypes are real-time conversation signals that are relayed once
// to a specific recipient and never persisted. They MUST carry an explicit
// recipient (`to`) — they are never broadcast to all connected clients, since
// that would leak who is typing / reading to everyone online.
var ephemeralSignalTypes = map[string]bool{
	"typing":       true, // payload: TypingSignal
	"read_receipt": true, // payload: ReadReceiptSignal
	"reaction":     true, // payload: ReactionSignal (live update; durable truth is the reactions API)
}

// TypingSignal is the payload of a Type:"typing" WS message (ephemeral).
type TypingSignal struct {
	ConversationID string `json:"conversation_id"`
	State          string `json:"state"` // "start" | "stop"
}

// ReadReceiptSignal is the payload of a Type:"read_receipt" WS message: the
// sender is told which of their messages the recipient has now read.
type ReadReceiptSignal struct {
	ConversationID string   `json:"conversation_id"`
	MessageIDs     []string `json:"message_ids"`
	ReadAt         string   `json:"read_at"`
}

// ReactionSignal is the payload of a Type:"reaction" WS message — a live reaction
// update relayed to the conversation peer. Empty Emoji means the reaction was
// removed. The persisted reactions API is the source of truth.
type ReactionSignal struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Emoji          string `json:"emoji"` // empty = removed
}

// EditSignal is the payload of a Type:"edit" WS message — a message was edited.
// Ciphertext is opaque (JSON-encoded as base64); the peer decrypts and replaces.
type EditSignal struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Ciphertext     []byte `json:"ciphertext"`
	Version        int    `json:"version,omitempty"` // server-retained version, if any
}

// DeleteSignal is the payload of a Type:"delete" WS message — a synchronized
// delete (WO-84). The peer tombstones the message locally.
type DeleteSignal struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
}

// PinSignal is the payload of a Type:"pin" WS message — a message was pinned or
// unpinned in the conversation (WO-59).
type PinSignal struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Pinned         bool   `json:"pinned"`
}

// DisappearingSignal is the payload of a Type:"disappearing_config" WS message —
// the conversation's disappearing-message TTL changed; the peer applies the timer.
type DisappearingSignal struct {
	ConversationID string `json:"conversation_id"`
	TTLSeconds     int    `json:"ttl_seconds"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// OfflineNotifier is invoked when a directed message cannot be delivered live
// because the recipient is not connected. Implementations send a content-blind
// push (conversation id + sender only, never content) so the device wakes and
// fetches. Best-effort and must not block; the hub calls it asynchronously.
type OfflineNotifier interface {
	NotifyUndelivered(recipientID, senderID, conversationID string)
}

// Hub manages all active WebSocket connections and routes messages.
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client // userID -> client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	notifier   OfflineNotifier // optional; push for offline recipients (WO-57)
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

// SignalPublisher pushes an ephemeral conversation signal to a single user over
// their live WebSocket connection, if any. It is best-effort: a false return means
// the user was offline or their buffer was full. The durable REST API (reactions,
// receipts) remains the source of truth — the signal is only a live nudge.
//
// *Hub satisfies this interface; handlers depend on the interface so they can be
// tested without a real hub (and so a nil publisher is a safe no-op).
type SignalPublisher interface {
	PublishSignal(to string, msg WSMessage) bool
}

// PublishSignal marshals msg and delivers it to `to` if they are connected.
// It fills Timestamp when empty. Returns false if `to` is empty or offline.
func (h *Hub) PublishSignal(to string, msg WSMessage) bool {
	if h == nil || to == "" {
		return false
	}
	if msg.Timestamp == "" {
		msg.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return false
	}
	return h.SendToUser(to, data)
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

// SetOfflineNotifier configures the push notifier used when a directed message's
// recipient is offline (WO-57). Safe to call once at startup before serving.
func (h *Hub) SetOfflineNotifier(n OfflineNotifier) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.notifier = n
}

// notifyUndelivered fires a content-blind push asynchronously if a notifier is set.
func (h *Hub) notifyUndelivered(recipientID, senderID, conversationID string) {
	h.mu.RLock()
	n := h.notifier
	h.mu.RUnlock()
	if n == nil || recipientID == "" {
		return
	}
	go n.NotifyUndelivered(recipientID, senderID, conversationID)
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

		c.routeInbound(msg)
	}
}

// routeInbound relays a parsed inbound message. Ephemeral conversation signals
// (typing, read receipts) are delivered only to their explicit recipient and are
// dropped if none is set — they are never broadcast to all connected clients.
// Other messages preserve the existing behavior: direct when `to` is set, else
// broadcast.
func (c *Client) routeInbound(msg WSMessage) {
	outBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	if ephemeralSignalTypes[msg.Type] {
		if msg.To == "" || msg.To == c.userID {
			// No recipient (or self) — drop. Never broadcast an ephemeral signal.
			return
		}
		c.hub.SendToUser(msg.To, outBytes)
		return
	}

	if msg.To != "" {
		// Directed message: deliver live, or push so the offline device wakes (WO-57).
		if !c.hub.SendToUser(msg.To, outBytes) {
			c.hub.notifyUndelivered(msg.To, c.userID, msg.ConversationID)
		}
	} else {
		c.hub.broadcast <- outBytes
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
