// Package relay implements the content-blind message relay service.
//
// The relay transports E2E encrypted blobs between clients. It cannot read,
// decrypt, or modify message content. It handles ciphertext blobs, offline
// queuing, push notifications, and Merkle commitment batching.
package relay

import (
	"errors"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

var (
	ErrRecipientNotFound = errors.New("recipient not found")
	ErrQueueFull         = errors.New("offline queue full for recipient")
	ErrMessageExpired    = errors.New("message has expired")
)

// RelayMessage represents an E2E encrypted blob in transit.
// The server sees metadata only; content is opaque ciphertext.
type RelayMessage struct {
	MessageID      string    `json:"messageId"`
	ConversationID string    `json:"conversationId"`
	SenderDID      string    `json:"senderDID"`
	RecipientDIDs  []string  `json:"recipientDIDs"`
	ContentType    string    `json:"contentType"`
	EncryptedBlob  []byte    `json:"encryptedBlob"`
	Commitment     []byte    `json:"commitment"`
	Signature      []byte    `json:"signature"`
	Timestamp      time.Time `json:"timestamp"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

// RelayResult indicates whether the message was delivered live or queued.
type RelayResult struct {
	MessageID  string            `json:"messageId"`
	Status     string            `json:"status"` // "relayed", "queued", or "partial"
	Recipients map[string]string `json:"recipients"` // DID -> "relayed" or "queued"
	Timestamp  time.Time         `json:"timestamp"`
}

// RelayService manages WebSocket connections and message transport.
type RelayService struct {
	connections *ConnectionManager
	queue       *OfflineQueue
	rateLimiter *infra.RateLimiter
	commitments *CommitmentBatch
}

// NewRelayService creates a new relay service.
func NewRelayService(rateLimiter *infra.RateLimiter) *RelayService {
	return &RelayService{
		connections: NewConnectionManager(),
		queue:       NewOfflineQueue(),
		rateLimiter: rateLimiter,
		commitments: NewCommitmentBatch(),
	}
}

// Relay processes an incoming message from a sender.
func (s *RelayService) Relay(msg RelayMessage) (*RelayResult, error) {
	// Check expiry
	if msg.ExpiresAt != nil && msg.ExpiresAt.Before(time.Now()) {
		return nil, ErrMessageExpired
	}

	// Rate limit check
	if err := s.rateLimiter.Check(msg.SenderDID, "message_send"); err != nil {
		return nil, infra.ErrRateLimitExceeded
	}

	result := &RelayResult{
		MessageID:  msg.MessageID,
		Recipients: make(map[string]string),
		Timestamp:  time.Now(),
	}

	hasRelayed := false
	hasQueued := false

	for _, recipientDID := range msg.RecipientDIDs {
		if s.connections.IsOnline(recipientDID) {
			if err := s.connections.Send(recipientDID, msg); err != nil {
				// Fallback to offline queue on send failure
				if qErr := s.queue.Enqueue(recipientDID, msg); qErr != nil {
					result.Recipients[recipientDID] = "failed"
					continue
				}
				result.Recipients[recipientDID] = "queued"
				hasQueued = true
			} else {
				result.Recipients[recipientDID] = "relayed"
				hasRelayed = true
			}
		} else {
			if err := s.queue.Enqueue(recipientDID, msg); err != nil {
				result.Recipients[recipientDID] = "failed"
				continue
			}
			result.Recipients[recipientDID] = "queued"
			hasQueued = true
		}
	}

	// Determine overall status
	if hasRelayed && hasQueued {
		result.Status = "partial"
	} else if hasRelayed {
		result.Status = "relayed"
	} else {
		result.Status = "queued"
	}

	// Add commitment to anchoring batch
	if len(msg.Commitment) > 0 {
		s.commitments.Add(msg.MessageID, msg.Commitment)
	}

	return result, nil
}

// DrainOfflineQueue sends all queued messages to a reconnecting client.
func (s *RelayService) DrainOfflineQueue(did string) ([]RelayMessage, error) {
	return s.queue.DequeueAll(did)
}

// Connect registers a user as online.
func (s *RelayService) Connect(did string) {
	s.connections.Register(did)
}

// Disconnect removes a user from online state.
func (s *RelayService) Disconnect(did string) {
	s.connections.Unregister(did)
}

// OnlineCount returns the number of currently connected users.
func (s *RelayService) OnlineCount() int {
	return s.connections.Count()
}

// QueueDepth returns the number of queued messages for a DID.
func (s *RelayService) QueueDepth(did string) int {
	return s.queue.Depth(did)
}

// PendingCommitments returns the number of uncommitted commitments.
func (s *RelayService) PendingCommitments() int {
	return s.commitments.Len()
}

// FlushCommitments returns and clears all pending commitments.
func (s *RelayService) FlushCommitments() []CommitmentEntry {
	return s.commitments.Flush()
}

// --- ConnectionManager ---

// ConnectionManager tracks which DIDs are currently online.
type ConnectionManager struct {
	mu       sync.RWMutex
	online   map[string]bool
	sendFunc func(did string, msg RelayMessage) error // injectable for testing
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		online: make(map[string]bool),
	}
}

// Register marks a DID as online.
func (cm *ConnectionManager) Register(did string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.online[did] = true
}

// Unregister marks a DID as offline.
func (cm *ConnectionManager) Unregister(did string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.online, did)
}

// IsOnline checks if a DID is currently connected.
func (cm *ConnectionManager) IsOnline(did string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.online[did]
}

// Send delivers a message to a connected user.
func (cm *ConnectionManager) Send(did string, msg RelayMessage) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if !cm.online[did] {
		return ErrRecipientNotFound
	}
	if cm.sendFunc != nil {
		return cm.sendFunc(did, msg)
	}
	return nil
}

// Count returns the number of online connections.
func (cm *ConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.online)
}

// --- OfflineQueue ---

const (
	MaxQueueDepth    = 1000
	DefaultRetention = 30 * 24 * time.Hour // 30 days for 1:1
	GroupRetention   = 7 * 24 * time.Hour  // 7 days for large groups
)

// OfflineQueue stores encrypted message blobs for offline recipients.
type OfflineQueue struct {
	mu     sync.Mutex
	queues map[string][]queueEntry
}

type queueEntry struct {
	Message   RelayMessage
	EnqueuedAt time.Time
	Retention  time.Duration
}

// NewOfflineQueue creates a new in-memory offline queue.
func NewOfflineQueue() *OfflineQueue {
	return &OfflineQueue{
		queues: make(map[string][]queueEntry),
	}
}

// Enqueue adds an encrypted blob to the recipient's offline queue.
func (q *OfflineQueue) Enqueue(recipientDID string, msg RelayMessage) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	queue := q.queues[recipientDID]

	// Evict oldest if at max depth
	if len(queue) >= MaxQueueDepth {
		queue = queue[1:]
	}

	retention := DefaultRetention
	if len(msg.RecipientDIDs) > 100 {
		retention = GroupRetention
	}

	queue = append(queue, queueEntry{
		Message:    msg,
		EnqueuedAt: time.Now(),
		Retention:  retention,
	})
	q.queues[recipientDID] = queue
	return nil
}

// DequeueAll retrieves and removes all queued messages for a recipient.
// Expired messages are silently dropped.
func (q *OfflineQueue) DequeueAll(recipientDID string) ([]RelayMessage, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	entries := q.queues[recipientDID]
	delete(q.queues, recipientDID)

	now := time.Now()
	messages := make([]RelayMessage, 0, len(entries))
	for _, e := range entries {
		if now.Sub(e.EnqueuedAt) < e.Retention {
			messages = append(messages, e.Message)
		}
	}
	return messages, nil
}

// Depth returns the current queue depth for a DID.
func (q *OfflineQueue) Depth(did string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queues[did])
}

// --- CommitmentBatch ---

// CommitmentEntry is a message commitment hash pending Merkle anchoring.
type CommitmentEntry struct {
	MessageID string
	Hash      []byte
	Timestamp time.Time
}

// CommitmentBatch collects commitments before Merkle tree construction.
type CommitmentBatch struct {
	mu          sync.Mutex
	commitments []CommitmentEntry
}

// NewCommitmentBatch creates a new commitment batch.
func NewCommitmentBatch() *CommitmentBatch {
	return &CommitmentBatch{}
}

// Add appends a commitment to the current batch.
func (cb *CommitmentBatch) Add(messageID string, hash []byte) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.commitments = append(cb.commitments, CommitmentEntry{
		MessageID: messageID,
		Hash:      hash,
		Timestamp: time.Now(),
	})
}

// Len returns the number of pending commitments.
func (cb *CommitmentBatch) Len() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return len(cb.commitments)
}

// Flush returns and clears all pending commitments.
func (cb *CommitmentBatch) Flush() []CommitmentEntry {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	batch := cb.commitments
	cb.commitments = nil
	return batch
}
