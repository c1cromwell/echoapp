package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/services/bots"
	"github.com/thechadcromwell/echoapp/internal/services/messaging"
	"github.com/thechadcromwell/echoapp/internal/services/relay"
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

// ContactBlockChecker reports whether a directed relay should be dropped (WO-190).
type ContactBlockChecker interface {
	IsEitherBlocked(ctx context.Context, a, b string) (bool, error)
}

// WSMessage represents a message sent over WebSocket.
type WSMessage struct {
	Type           string          `json:"type"`                      // "text", "control"
	From           string          `json:"from,omitempty"`            // sender user ID
	To             string          `json:"to,omitempty"`              // recipient user ID (frames without one are dropped unless group-addressed)
	ConversationID string          `json:"conversation_id,omitempty"` // optional conversation scope
	Payload        json.RawMessage `json:"payload"`                   // message content
	Timestamp      string          `json:"timestamp"`
	Silent         bool            `json:"silent,omitempty"` // WO-56: suppress alert push for recipient
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
	"typing":           true, // payload: TypingSignal
	"read_receipt":     true, // payload: ReadReceiptSignal
	"reaction":         true, // payload: ReactionSignal (live update; durable truth is the reactions API)
	"group_key":        true, // payload: GroupKeySignal (E2E key package; content-blind opaque blob)
	"screenshot_alert": true, // payload: ScreenshotAlertSignal (M6)
	"ratchet_prekey":   true, // payload: ratchet pre-key + optional PQ hybrid bundle/ciphertext (WO-SX1/SX2)
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

// GroupKeySignal is the payload of a Type:"group_key" WS message — an admin
// distributes a per-member encrypted AES-256 group key package (WO-207 / M2).
type GroupKeySignal struct {
	GroupID       string `json:"group_id"`
	Version       int    `json:"version"`
	EncryptedKey  []byte `json:"encrypted_key"`
	DistributedBy string `json:"distributed_by"`
}

// PollSignal is the payload of a Type:"poll" WS message (WO-23 / M6).
// Options and votes are opaque client-encrypted blobs; relay routes only.
type PollSignal struct {
	ConversationID string `json:"conversation_id"`
	PollID         string `json:"poll_id"`
	Action         string `json:"action"` // create|vote|close
	OptionID       string `json:"option_id,omitempty"`
	Ciphertext     []byte `json:"ciphertext,omitempty"`
}

// ScreenshotAlertSignal notifies the peer that a screenshot may have been taken (M6).
type ScreenshotAlertSignal struct {
	ConversationID string `json:"conversation_id"`
	AlertedAt      string `json:"alerted_at"`
}

// CallSignal is the payload of Type:"call_signal" WS messages (M4 / WO-5).
// SDP and ICE payloads are opaque to the relay (content-blind).
type CallSignal struct {
	CallID   string          `json:"call_id"`
	Action   string          `json:"action"` // offer|answer|ice|hangup|reject|ring
	CallType string          `json:"call_type,omitempty"`
	SDP      string          `json:"sdp,omitempty"`
	Ice      json.RawMessage `json:"ice_candidate,omitempty"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// GroupMemberLister resolves group membership for content-blind group text fan-out (M2).
type GroupMemberLister interface {
	GroupMemberDIDs(groupID string) ([]string, error)
	IsGroupMember(groupID, memberID string) (bool, error)
}

// OfflineNotifier is invoked when a directed message cannot be delivered live
// because the recipient is not connected. Implementations send a content-blind
// push (conversation id + sender only, never content) so the device wakes and
// fetches. Best-effort and must not block; the hub calls it asynchronously.
type OfflineNotifier interface {
	NotifyUndelivered(recipientID, senderID, conversationID string, silent bool)
	// NotifyMissedCall fires when an offline recipient misses a call offer (M4c / WO-196).
	NotifyMissedCall(recipientID, senderID, callID string)
}

// Hub manages all active WebSocket connections and routes messages.
type Hub struct {
	mu             sync.RWMutex
	clients        map[string]*Client // userID -> client
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	notifier       OfflineNotifier                               // optional; push for offline recipients (WO-57)
	groupMembers   GroupMemberLister                             // optional; fan-out group text to members (M2)
	offlineQueue   *wsOfflineQueue                               // directed WS blobs for offline recipients (M2)
	durableQueue   database.DB                                   // optional Postgres/memory message_queue (survives restart)
	rateLimiter    *infra.RateLimiter                            // WO-44 per-DID WS message budget (optional)
	sealedTokens   *messaging.SealedTokenStore                   // WO-219 sealed-sender delivery tokens
	convNotifPrefs *messaging.ConversationNotificationPrefsStore // WO-56 mute → silent push
	commitments    *relay.CommitmentBatch                        // message integrity commitments (Data L1 prep)
	botWebhooks    *bots.WebhookRegistry                         // WO-11 inbound bot webhooks
	blockChecker   ContactBlockChecker                           // WO-190: drop relay when either party blocked
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[string]*Client),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		offlineQueue: newWSOfflineQueue(),
		commitments:  relay.NewCommitmentBatch(),
	}
}

// SetCommitmentBatch replaces the hub's commitment collector (optional).
func (h *Hub) SetCommitmentBatch(cb *relay.CommitmentBatch) {
	h.commitments = cb
}

// PendingCommitments returns unflushed commitment count.
func (h *Hub) PendingCommitments() int {
	if h.commitments == nil {
		return 0
	}
	return h.commitments.Len()
}

// FlushCommitments drains pending commitments for Merkle anchoring.
func (h *Hub) FlushCommitments() []relay.CommitmentEntry {
	if h.commitments == nil {
		return nil
	}
	return h.commitments.Flush()
}

// SetRateLimiter enables per-DID WebSocket send rate limits (WO-44).
func (h *Hub) SetRateLimiter(rl *infra.RateLimiter) {
	h.rateLimiter = rl
}

// Run starts the hub's event loop. Call this in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// A reconnect arrives before the old socket has noticed it is dead.
			// Retire the superseded socket here so it cannot linger.
			if old, ok := h.clients[client.userID]; ok && old != client {
				close(old.send)
			}
			h.clients[client.userID] = client
			h.mu.Unlock()
			go h.flushOffline(client)

		case client := <-h.unregister:
			h.mu.Lock()
			// Only remove the entry if it is still THIS socket. Matching on the
			// user ID alone let a stale socket's late unregister evict the
			// user's fresh connection, leaving them connected but silently
			// undeliverable until the next reconnect.
			if cur, ok := h.clients[client.userID]; ok && cur == client {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// Write lock: slow clients are removed from the map below.
			h.mu.Lock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full — drop and disconnect
					close(client.send)
					delete(h.clients, client.userID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// DisconnectUser closes the user's live socket, if any. Used when an account
// is deleted so an already-open connection cannot outlive it.
func (h *Hub) DisconnectUser(userID string) {
	if h == nil || userID == "" {
		return
	}
	h.mu.Lock()
	if c, ok := h.clients[userID]; ok {
		delete(h.clients, userID)
		close(c.send)
	}
	h.mu.Unlock()
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

// PublishConfirmation delivers a message-integrity confirmation to the sender (WO-15).
func (h *Hub) PublishConfirmation(to string, confirmation metagraph.AnchorConfirmation) bool {
	if h == nil || to == "" {
		return false
	}
	payload, err := json.Marshal(confirmation)
	if err != nil {
		return false
	}
	return h.PublishSignal(to, WSMessage{
		Type:    "confirmation",
		To:      to,
		Payload: payload,
	})
}

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
	if h.deliverOrQueue(msg.To, data, msg.From, msg.ConversationID, msg.Silent) {
		return true
	}
	return false
}

// SetConversationNotificationPrefs configures mute lookup for silent push (WO-56).
func (h *Hub) SetConversationNotificationPrefs(s *messaging.ConversationNotificationPrefsStore) {
	h.convNotifPrefs = s
}

// deliverOrQueue sends live or enqueues for reconnect replay (M2 offline group + signals).
func (h *Hub) deliverOrQueue(recipient string, data []byte, senderID, conversationID string, silent bool) bool {
	queueData := scrubRelayMetadataForQueue(data)
	if h.SendToUser(recipient, queueData) {
		return true
	}
	h.enqueueOffline(recipient, queueData, senderID, conversationID)
	h.notifyUndelivered(recipient, senderID, conversationID, silent)
	return false
}

func (h *Hub) enqueueOffline(recipient string, data []byte, senderID, conversationID string) {
	if h.durableQueue != nil {
		_ = h.durableQueue.Enqueue(context.Background(), &database.QueuedMessage{
			MessageID:      uuid.New().String(),
			ConversationID: conversationID,
			SenderDID:      senderID,
			RecipientDID:   recipient,
			Payload:        append([]byte(nil), data...),
			ExpiresAt:      time.Now().Add(wsOfflineRetention),
		})
		return
	}
	if h.offlineQueue != nil {
		h.offlineQueue.Enqueue(recipient, data, wsOfflineRetention)
	}
}

// scrubRelayMetadataForQueue strips sender `from` on sealed payloads before offline
// persistence (WO-SX3 metadata minimization).
func scrubRelayMetadataForQueue(data []byte) []byte {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return data
	}
	if msg.Type == "sealed_text" {
		msg.From = ""
		if out, err := json.Marshal(msg); err == nil {
			return out
		}
	}
	return data
}

// offlineFlushPage is how many durable-queue messages are claimed per round
// trip. It must stay well under the client send buffer so a claimed page always
// fits without dropping.
const offlineFlushPage = 64

// messageRequeuer is implemented by durable queues that can hand back messages
// which were claimed but never reached the socket.
type messageRequeuer interface {
	Requeue(ctx context.Context, messageIDs []string) error
}

// isCurrent reports whether c is still the registered socket for its user.
func (h *Hub) isCurrent(c *Client) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[c.userID] == c
}

// trySend offers data to c without blocking. It holds the read lock so the
// send cannot race the hub closing c.send (which happens under the write lock);
// a superseded or unregistered client is never written to.
func (h *Hub) trySend(c *Client, data []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.clients[c.userID] != c {
		return false
	}
	select {
	case c.send <- data:
		return true
	default:
		return false
	}
}

// flushOffline replays queued directed WS payloads when a client reconnects.
//
// The durable queue is drained a page at a time, and a page is only claimed
// once the send buffer has room for all of it. Claiming the whole backlog up
// front (it can be 1000 deep against a 256-slot buffer) marked everything
// delivered and then dropped whatever did not fit, so a user returning from a
// long absence silently lost messages.
func (h *Hub) flushOffline(c *Client) {
	if h.durableQueue != nil {
		ctx := context.Background()
		page := offlineFlushPage
		if cap(c.send) < page {
			page = cap(c.send)
		}
		if page == 0 {
			return
		}
		for {
			waitUntil := time.Now().Add(30 * time.Second)
			for cap(c.send)-len(c.send) < page {
				if !h.isCurrent(c) || time.Now().After(waitUntil) {
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
			if !h.isCurrent(c) {
				return
			}

			msgs, err := h.durableQueue.Dequeue(ctx, c.userID, page)
			if err != nil || len(msgs) == 0 {
				return
			}
			for i, msg := range msgs {
				if msg == nil || len(msg.Payload) == 0 {
					continue
				}
				if h.trySend(c, msg.Payload) {
					continue
				}
				// The socket went away or stalled mid-page: hand this message
				// and the rest of the page back for the next reconnect.
				var ids []string
				for _, rest := range msgs[i:] {
					if rest != nil {
						ids = append(ids, rest.MessageID)
					}
				}
				if rq, ok := h.durableQueue.(messageRequeuer); ok {
					_ = rq.Requeue(ctx, ids)
				}
				return
			}
			if len(msgs) < page {
				return
			}
		}
	}
	if h.offlineQueue == nil {
		return
	}
	drain := h.offlineQueue.DequeueAll(c.userID)
	for i, data := range drain.Blobs {
		if h.trySend(c, data) {
			continue
		}
		// Put back this blob and everything after it, not just the one that
		// failed to fit.
		for _, rest := range drain.Blobs[i:] {
			h.offlineQueue.Enqueue(c.userID, rest, wsOfflineRetention)
		}
		return
	}
	if len(drain.OverflowURIs) > 0 {
		payload, err := json.Marshal(map[string]interface{}{
			"storage_uris": drain.OverflowURIs,
		})
		if err != nil {
			return
		}
		manifest, err := json.Marshal(WSMessage{
			Type:      "overflow_manifest",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   payload,
		})
		if err != nil {
			return
		}
		h.trySend(c, manifest)
	}
}

// SendToUser delivers a message to a specific user if connected.
// Returns true if the user was found and the message was queued.
func (h *Hub) SendToUser(userID string, data []byte) bool {
	// The read lock is held across the non-blocking send so it cannot race the
	// hub closing client.send, which happens under the write lock.
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[userID]
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

// SetContactBlockChecker configures block-list enforcement on directed relay (WO-190).
func (h *Hub) SetContactBlockChecker(c ContactBlockChecker) {
	h.blockChecker = c
}

func (h *Hub) shouldDropBlocked(recipient, sender string) bool {
	if h == nil || h.blockChecker == nil || recipient == "" || sender == "" || recipient == sender {
		return false
	}
	blocked, err := h.blockChecker.IsEitherBlocked(context.Background(), recipient, sender)
	return err == nil && blocked
}

// SetOfflineNotifier configures the push notifier used when a directed message's
// recipient is offline (WO-57). Safe to call once at startup before serving.
func (h *Hub) SetOfflineNotifier(n OfflineNotifier) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.notifier = n
}

// SetBotWebhooks enables inbound user→bot webhook dispatch (WO-11).
func (h *Hub) SetBotWebhooks(registry *bots.WebhookRegistry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.botWebhooks = registry
}

// SetGroupMemberLister configures group membership lookup for group text fan-out (M2).
func (h *Hub) SetGroupMemberLister(l GroupMemberLister) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.groupMembers = l
}

// SetOfflineOverflowStorage configures encblob overflow for WO-237 queue depth > 1000.
func (h *Hub) SetOfflineOverflowStorage(s encblob.Storage) {
	if h.offlineQueue != nil {
		h.offlineQueue.SetOverflowStorage(s)
	}
}

// SetDurableOfflineQueue persists undelivered WS blobs in message_queue so they
// survive process restart. When set, it replaces the in-memory queue.
func (h *Hub) SetDurableOfflineQueue(db database.DB) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.durableQueue = db
}

// notifyUndelivered fires a content-blind push asynchronously if a notifier is set.
func (h *Hub) notifyUndelivered(recipientID, senderID, conversationID string, silent bool) {
	h.mu.RLock()
	n := h.notifier
	prefs := h.convNotifPrefs
	h.mu.RUnlock()
	if n == nil || recipientID == "" {
		return
	}
	if prefs != nil && prefs.IsMuted(recipientID, conversationID) {
		return
	}
	go n.NotifyUndelivered(recipientID, senderID, conversationID, silent)
}

// deliverOrQueueCall delivers call_signal live, queues for reconnect, and pushes missed-call on offer.
func (h *Hub) deliverOrQueueCall(recipient string, data []byte, senderID string, payload json.RawMessage) bool {
	if h.SendToUser(recipient, data) {
		return true
	}
	h.enqueueOffline(recipient, data, senderID, "")
	h.notifyCallUndelivered(recipient, senderID, payload)
	return false
}

func (h *Hub) notifyCallUndelivered(recipientID, senderID string, payload json.RawMessage) {
	h.mu.RLock()
	n := h.notifier
	h.mu.RUnlock()
	if n == nil || recipientID == "" {
		return
	}
	var sig CallSignal
	if err := json.Unmarshal(payload, &sig); err == nil && sig.Action == "offer" && sig.CallID != "" {
		go n.NotifyMissedCall(recipientID, senderID, sig.CallID)
		return
	}
	go n.NotifyUndelivered(recipientID, senderID, "", false)
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

	// The extractor validates the token and returns "" when it is missing,
	// malformed, expired, or revoked. Never upgrade such a connection: an
	// anonymous client would otherwise join the hub with an empty identity,
	// skip the per-DID rate limiter and block lists, and be able to push
	// frames at any user.
	userID := userIDExtractor(token)
	if userID == "" {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

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
			if msg.Type != "sealed_text" {
				msg.From = c.userID
			}
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
	if c.hub.rateLimiter != nil && c.userID != "" && msg.Type != "control" {
		action := "websocket_msg"
		if msg.Type == "text" || msg.Type == "sealed_text" {
			action = "message_send"
		}
		if err := c.hub.rateLimiter.Check(c.userID, action); err != nil {
			// Tell the sender instead of dropping silently: without this the
			// client believes the frame went out and never retries it.
			c.notifyRateLimited(msg, err)
			return
		}
	}
	if msg.From == "" && msg.Type != "sealed_text" {
		msg.From = c.userID
	}

	if c.hub.commitments != nil && (msg.Type == "text" || msg.Type == "sealed_text") {
		if messageID, hash, ok := commitmentFromWSMessage(msg); ok && len(hash) > 0 {
			if messageID == "" {
				messageID = msg.ConversationID + ":" + msg.Timestamp
			}
			sender := msg.From
			if sender == "" {
				sender = c.userID
			}
			c.hub.commitments.Add(messageID, sender, hash)
		}
	}

	outBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	if msg.Type == "sealed_text" {
		if msg.To != "" && c.hub.shouldDropBlocked(msg.To, c.userID) {
			return
		}
		c.routeSealedText(msg)
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

	// Poll updates (WO-23): directed only; queue offline for reconnect (S4 durability).
	if msg.Type == "poll" {
		if msg.To == "" || msg.To == c.userID {
			return
		}
		if c.hub.shouldDropBlocked(msg.To, c.userID) {
			return
		}
		c.hub.deliverOrQueue(msg.To, outBytes, c.userID, msg.ConversationID, msg.Silent)
		return
	}

	// WebRTC call signaling (M4): directed only; queue offline + content-blind push.
	if msg.Type == "call_signal" {
		if msg.To == "" || msg.To == c.userID {
			return
		}
		if c.hub.shouldDropBlocked(msg.To, c.userID) {
			return
		}
		c.hub.deliverOrQueueCall(msg.To, outBytes, c.userID, msg.Payload)
		return
	}

	// Group ciphertext fan-out: conversation_id "group:{id}" with no `to` routes to all members.
	if msg.Type == "text" && strings.HasPrefix(msg.ConversationID, "group:") && msg.To == "" {
		c.routeGroupText(msg)
		return
	}

	if msg.To != "" {
		if msg.Type == "text" && c.hub.botWebhooks != nil && bots.IsCatalogBot(msg.To) {
			c.hub.dispatchBotInbound(msg)
		}
		if c.hub.shouldDropBlocked(msg.To, c.userID) {
			return
		}
		// Directed message: deliver live, queue for reconnect, or push (WO-57).
		c.hub.deliverOrQueue(msg.To, outBytes, c.userID, msg.ConversationID, msg.Silent)
	}
	// A client frame with no recipient is dropped. It used to fan out to every
	// connected client, which let any single user spam the whole network and
	// exposed sender DIDs to strangers. Server-originated fan-out still goes
	// through Hub.Broadcast; group traffic goes through routeGroupText above.
}

// notifyRateLimited sends the client an "error" frame for a frame the per-DID
// budget rejected, identifying which one so it can be retried after the delay.
func (c *Client) notifyRateLimited(rejected WSMessage, limitErr error) {
	retryAfter := 1
	if rle, ok := limitErr.(*infra.RateLimitExceededError); ok {
		if secs := int(rle.RetryAfter.Seconds()) + 1; secs > retryAfter {
			retryAfter = secs
		}
	}
	payload, err := json.Marshal(map[string]interface{}{
		"code":                "RATE_LIMITED",
		"retry_after_seconds": retryAfter,
		"rejected_type":       rejected.Type,
		"to":                  rejected.To,
		"conversation_id":     rejected.ConversationID,
	})
	if err != nil {
		return
	}
	frame, err := json.Marshal(WSMessage{
		Type:      "error",
		From:      "server",
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return
	}
	c.hub.trySend(c, frame)
}

func (h *Hub) dispatchBotInbound(msg WSMessage) {
	if h == nil || h.botWebhooks == nil || msg.To == "" {
		return
	}
	var wire struct {
		Text       string `json:"text"`
		Ciphertext []byte `json:"ciphertext"`
	}
	_ = json.Unmarshal(msg.Payload, &wire)
	go bots.DispatchInbound(h.botWebhooks, msg.To, bots.InboundPayload{
		FromDID:        msg.From,
		ConversationID: msg.ConversationID,
		Text:           wire.Text,
		Ciphertext:     wire.Ciphertext,
		ReceivedAt:     time.Now().UTC().Format(time.RFC3339),
	})
}

// routeGroupText delivers an opaque group text blob to every group member except the sender.
func (c *Client) routeGroupText(msg WSMessage) {
	groupID := strings.TrimPrefix(msg.ConversationID, "group:")
	if groupID == "" {
		return
	}

	c.hub.mu.RLock()
	lister := c.hub.groupMembers
	c.hub.mu.RUnlock()
	if lister == nil {
		return
	}

	ok, err := lister.IsGroupMember(groupID, c.userID)
	if err != nil || !ok {
		return
	}

	members, err := lister.GroupMemberDIDs(groupID)
	if err != nil {
		return
	}

	for _, member := range members {
		if member == "" || member == c.userID {
			continue
		}
		delivered := msg
		delivered.To = member
		perMember, err := json.Marshal(delivered)
		if err != nil {
			continue
		}
		c.hub.deliverOrQueue(member, perMember, c.userID, msg.ConversationID, msg.Silent)
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
		c.hub.trySend(c, data)
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
