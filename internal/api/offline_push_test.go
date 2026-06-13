package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// fakeNotifier records content-blind offline pushes (WO-57).
type fakeNotifier struct {
	mu   sync.Mutex
	sent []pushRecord
}

type pushRecord struct {
	recipient, sender, conversation string
}

func (f *fakeNotifier) NotifyUndelivered(recipientID, senderID, conversationID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, pushRecord{recipientID, senderID, conversationID})
}

func (f *fakeNotifier) all() []pushRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]pushRecord(nil), f.sent...)
}

// waitFor polls until cond() or the deadline; the hub pushes asynchronously.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("condition not met within deadline")
}

// TestHub_OfflineDirectMessageTriggersPush: a directed text to an offline user
// fires a content-blind push (WO-57); a delivered one does not.
func TestHub_OfflineDirectMessageTriggersPush(t *testing.T) {
	h := NewHub()
	notifier := &fakeNotifier{}
	h.SetOfflineNotifier(notifier)

	sender := &Client{hub: h, userID: "did:alice", send: make(chan []byte, 8)}

	// Recipient offline → push.
	sender.routeInbound(WSMessage{Type: "text", From: "did:alice", To: "did:bob", ConversationID: "c1", Payload: json.RawMessage(`"x"`)})
	waitFor(t, func() bool { return len(notifier.all()) == 1 })
	got := notifier.all()[0]
	if got.recipient != "did:bob" || got.sender != "did:alice" || got.conversation != "c1" {
		t.Fatalf("unexpected push record: %+v", got)
	}

	// Recipient online → no additional push.
	registerFakeClient(h, "did:bob")
	sender.routeInbound(WSMessage{Type: "text", From: "did:alice", To: "did:bob", ConversationID: "c1", Payload: json.RawMessage(`"x"`)})
	// Give any erroneous async push a chance to land, then assert count unchanged.
	time.Sleep(20 * time.Millisecond)
	if len(notifier.all()) != 1 {
		t.Fatalf("online recipient must not be pushed, got %d pushes", len(notifier.all()))
	}
}

// TestHub_EphemeralSignalNeverPushes: typing/receipt signals to an offline user
// must not generate a push (they are ephemeral, not messages).
func TestHub_EphemeralSignalNeverPushes(t *testing.T) {
	h := NewHub()
	notifier := &fakeNotifier{}
	h.SetOfflineNotifier(notifier)
	sender := &Client{hub: h, userID: "did:alice", send: make(chan []byte, 8)}

	sender.routeInbound(typingMsg("did:alice", "did:bob")) // bob offline
	time.Sleep(20 * time.Millisecond)
	if len(notifier.all()) != 0 {
		t.Fatalf("ephemeral signal must never push, got %d", len(notifier.all()))
	}
}

// offlinePublisher reports every target as offline.
type offlinePublisher struct{}

func (offlinePublisher) PublishSignal(string, WSMessage) bool { return false }

// TestMessageReact_OfflinePeerPushFires verifies the push path when the live
// publisher reports the peer offline.
func TestMessageReact_OfflinePeerPushFires(t *testing.T) {
	db := database.NewMemoryDB()
	notifier := &fakeNotifier{}
	rt := &Router{V3: &V3Handlers{DB: db, Signals: offlinePublisher{}, Notifier: notifier}}
	mux := v3Mux(rt)
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	body, _ := json.Marshal(map[string]string{"message_id": "m1", "emoji": "👍"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
	mux.ServeHTTP(httptest.NewRecorder(), req)

	got := notifier.all()
	if len(got) != 1 || got[0].recipient != "did:bob" || got[0].sender != "did:alice" {
		t.Fatalf("expected one offline push to did:bob, got %+v", got)
	}
}

// TestMessageReact_OnlinePeerNoPush: when the peer is online (signal delivered),
// no push is sent.
func TestMessageReact_OnlinePeerNoPush(t *testing.T) {
	db := database.NewMemoryDB()
	notifier := &fakeNotifier{}
	// fakeSignalPublisher returns true (delivered) → no push.
	rt := &Router{V3: &V3Handlers{DB: db, Signals: &fakeSignalPublisher{}, Notifier: notifier}}
	mux := v3Mux(rt)
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	body, _ := json.Marshal(map[string]string{"message_id": "m1", "emoji": "👍"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
	mux.ServeHTTP(httptest.NewRecorder(), req)

	if len(notifier.all()) != 0 {
		t.Fatalf("online peer must not be pushed, got %+v", notifier.all())
	}
}
