package relay

import (
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

func newTestRelay() *RelayService {
	limits := map[string]infra.RateLimitConfig{
		"message_send": {MaxRequests: 60, Window: time.Minute},
	}
	return NewRelayService(infra.NewRateLimiter(limits))
}

func testMessage(id, sender string, recipients ...string) RelayMessage {
	return RelayMessage{
		MessageID:     id,
		ConversationID: "conv-1",
		SenderDID:     sender,
		RecipientDIDs: recipients,
		ContentType:   "text",
		EncryptedBlob: []byte("encrypted-blob-data"),
		Commitment:    []byte("commitment-hash"),
		Signature:     []byte("signature"),
		Timestamp:     time.Now(),
	}
}

// --- RelayService Tests ---

func TestRelay_OnlineRecipient(t *testing.T) {
	rs := newTestRelay()
	rs.Connect("did:dag:bob")

	msg := testMessage("msg-1", "did:dag:alice", "did:dag:bob")
	result, err := rs.Relay(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "relayed" {
		t.Errorf("expected relayed, got %s", result.Status)
	}
	if result.Recipients["did:dag:bob"] != "relayed" {
		t.Errorf("bob should be relayed, got %s", result.Recipients["did:dag:bob"])
	}
}

func TestRelay_OfflineRecipient(t *testing.T) {
	rs := newTestRelay()
	// bob is NOT connected

	msg := testMessage("msg-2", "did:dag:alice", "did:dag:bob")
	result, err := rs.Relay(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "queued" {
		t.Errorf("expected queued, got %s", result.Status)
	}
	if rs.QueueDepth("did:dag:bob") != 1 {
		t.Errorf("expected 1 queued message for bob, got %d", rs.QueueDepth("did:dag:bob"))
	}
}

func TestRelay_MixedRecipients(t *testing.T) {
	rs := newTestRelay()
	rs.Connect("did:dag:bob")
	// charlie is offline

	msg := testMessage("msg-3", "did:dag:alice", "did:dag:bob", "did:dag:charlie")
	result, err := rs.Relay(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "partial" {
		t.Errorf("expected partial, got %s", result.Status)
	}
	if result.Recipients["did:dag:bob"] != "relayed" {
		t.Error("bob should be relayed")
	}
	if result.Recipients["did:dag:charlie"] != "queued" {
		t.Error("charlie should be queued")
	}
}

func TestRelay_RateLimit(t *testing.T) {
	limits := map[string]infra.RateLimitConfig{
		"message_send": {MaxRequests: 3, Window: time.Minute},
	}
	rs := NewRelayService(infra.NewRateLimiter(limits))

	for i := 0; i < 3; i++ {
		msg := testMessage("msg-rl-"+string(rune('0'+i)), "did:dag:alice", "did:dag:bob")
		_, err := rs.Relay(msg)
		if err != nil {
			t.Fatalf("request %d should succeed: %v", i, err)
		}
	}

	msg := testMessage("msg-rl-blocked", "did:dag:alice", "did:dag:bob")
	_, err := rs.Relay(msg)
	if err != infra.ErrRateLimitExceeded {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestRelay_ExpiredMessage(t *testing.T) {
	rs := newTestRelay()
	past := time.Now().Add(-1 * time.Hour)
	msg := testMessage("msg-exp", "did:dag:alice", "did:dag:bob")
	msg.ExpiresAt = &past

	_, err := rs.Relay(msg)
	if err != ErrMessageExpired {
		t.Errorf("expected expired error, got: %v", err)
	}
}

func TestRelay_CommitmentBatching(t *testing.T) {
	rs := newTestRelay()

	for i := 0; i < 5; i++ {
		msg := testMessage("msg-c-"+string(rune('0'+i)), "did:dag:alice", "did:dag:bob")
		rs.Relay(msg)
	}

	if rs.PendingCommitments() != 5 {
		t.Errorf("expected 5 pending commitments, got %d", rs.PendingCommitments())
	}

	entries := rs.FlushCommitments()
	if len(entries) != 5 {
		t.Errorf("expected 5 flushed commitments, got %d", len(entries))
	}

	if rs.PendingCommitments() != 0 {
		t.Error("commitments should be empty after flush")
	}
}

// --- ConnectionManager Tests ---

func TestConnectionManager_RegisterAndOnline(t *testing.T) {
	cm := NewConnectionManager()
	cm.Register("did:dag:user1")

	if !cm.IsOnline("did:dag:user1") {
		t.Error("registered user should be online")
	}
	if cm.IsOnline("did:dag:user2") {
		t.Error("unregistered user should not be online")
	}
}

func TestConnectionManager_Unregister(t *testing.T) {
	cm := NewConnectionManager()
	cm.Register("did:dag:user1")
	cm.Unregister("did:dag:user1")

	if cm.IsOnline("did:dag:user1") {
		t.Error("unregistered user should not be online")
	}
}

func TestConnectionManager_Count(t *testing.T) {
	cm := NewConnectionManager()
	cm.Register("did:dag:a")
	cm.Register("did:dag:b")
	cm.Register("did:dag:c")

	if cm.Count() != 3 {
		t.Errorf("expected 3 connections, got %d", cm.Count())
	}

	cm.Unregister("did:dag:b")
	if cm.Count() != 2 {
		t.Errorf("expected 2 connections, got %d", cm.Count())
	}
}

// --- OfflineQueue Tests ---

func TestOfflineQueue_EnqueueAndDequeue(t *testing.T) {
	q := NewOfflineQueue()
	msg := testMessage("msg-q1", "did:dag:alice", "did:dag:bob")

	err := q.Enqueue("did:dag:bob", msg)
	if err != nil {
		t.Fatalf("enqueue error: %v", err)
	}

	if q.Depth("did:dag:bob") != 1 {
		t.Errorf("expected depth 1, got %d", q.Depth("did:dag:bob"))
	}

	messages, err := q.DequeueAll("did:dag:bob")
	if err != nil {
		t.Fatalf("dequeue error: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
	if messages[0].MessageID != "msg-q1" {
		t.Errorf("expected msg-q1, got %s", messages[0].MessageID)
	}

	// Queue should be empty after dequeue
	if q.Depth("did:dag:bob") != 0 {
		t.Error("queue should be empty after dequeue")
	}
}

func TestOfflineQueue_MaxDepthEviction(t *testing.T) {
	q := NewOfflineQueue()

	// Fill queue to max
	for i := 0; i < MaxQueueDepth+5; i++ {
		msg := testMessage("msg-"+string(rune(i)), "did:dag:alice", "did:dag:bob")
		q.Enqueue("did:dag:bob", msg)
	}

	if q.Depth("did:dag:bob") != MaxQueueDepth {
		t.Errorf("queue should be capped at %d, got %d", MaxQueueDepth, q.Depth("did:dag:bob"))
	}
}

func TestOfflineQueue_SeparateRecipients(t *testing.T) {
	q := NewOfflineQueue()

	q.Enqueue("did:dag:bob", testMessage("msg-1", "alice", "did:dag:bob"))
	q.Enqueue("did:dag:charlie", testMessage("msg-2", "alice", "did:dag:charlie"))

	if q.Depth("did:dag:bob") != 1 {
		t.Error("bob should have 1 message")
	}
	if q.Depth("did:dag:charlie") != 1 {
		t.Error("charlie should have 1 message")
	}

	// Dequeue bob doesn't affect charlie
	q.DequeueAll("did:dag:bob")
	if q.Depth("did:dag:charlie") != 1 {
		t.Error("charlie's queue should be unaffected")
	}
}

func TestOfflineQueue_EmptyDequeue(t *testing.T) {
	q := NewOfflineQueue()
	messages, err := q.DequeueAll("did:dag:nobody")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

// --- CommitmentBatch Tests ---

func TestCommitmentBatch_AddAndFlush(t *testing.T) {
	cb := NewCommitmentBatch()

	cb.Add("msg-1", []byte("hash1"))
	cb.Add("msg-2", []byte("hash2"))
	cb.Add("msg-3", []byte("hash3"))

	if cb.Len() != 3 {
		t.Errorf("expected 3 commitments, got %d", cb.Len())
	}

	entries := cb.Flush()
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	if cb.Len() != 0 {
		t.Error("should be empty after flush")
	}
}

func TestCommitmentBatch_FlushEmpty(t *testing.T) {
	cb := NewCommitmentBatch()
	entries := cb.Flush()
	if entries != nil {
		t.Error("flushing empty batch should return nil")
	}
}
