package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
)

// fakeSignalPublisher captures signals a handler pushes, so tests can assert
// server-authoritative fan-out without a real WebSocket hub.
type fakeSignalPublisher struct {
	mu   sync.Mutex
	sent []sentSignal
}

type sentSignal struct {
	to  string
	msg WSMessage
}

func (f *fakeSignalPublisher) PublishSignal(to string, msg WSMessage) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sentSignal{to: to, msg: msg})
	return true
}

func (f *fakeSignalPublisher) all() []sentSignal {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sentSignal(nil), f.sent...)
}

func signalsRouter(db database.DB, pub SignalPublisher) *Router {
	return &Router{V3: &V3Handlers{DB: db, Signals: pub}}
}

func complySignalsRouter(db database.DB, pub SignalPublisher, svc *comply.Service) *Router {
	return &Router{V3: &V3Handlers{DB: db, Signals: pub, Comply: svc}}
}

func enqueue(t *testing.T, db database.DB, id, conv, sender, recipient string) {
	t.Helper()
	if err := db.Enqueue(context.Background(), &database.QueuedMessage{
		MessageID:      id,
		ConversationID: conv,
		SenderDID:      sender,
		RecipientDID:   recipient,
		Payload:        []byte("ciphertext"),
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
}

// TestMessageReceipt_ReadIsDurableAndSignalsSender: a "read" receipt durably
// records read state (WO-192) and pushes a live read_receipt to the sender.
func TestMessageReceipt_ReadIsDurableAndSignalsSender(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))

	// bob sent m1 to alice; alice (the reader) marks it read.
	enqueue(t, db, "m1", "c1", "did:bob", "did:alice")

	body, _ := json.Marshal(map[string]string{"receiptType": "read"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/m1/receipt", bytes.NewReader(body)), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("receipt want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	meta, err := db.GetMessageMeta(context.Background(), "m1")
	if err != nil {
		t.Fatalf("meta: %v", err)
	}
	if meta.Status != "read" || meta.ReadAt == nil || meta.DeliveredAt == nil {
		t.Fatalf("expected durable read state (read implies delivered), got %+v", meta)
	}

	sent := pub.all()
	if len(sent) != 1 || sent[0].to != "did:bob" || sent[0].msg.Type != "read_receipt" {
		t.Fatalf("expected one read_receipt signal to sender did:bob, got %+v", sent)
	}
}

// TestMessageReceipt_DeliveredDefault: an unspecified/"delivered" receiptType marks
// delivered (not read) and does not signal the sender.
func TestMessageReceipt_DeliveredDefault(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))
	enqueue(t, db, "m1", "c1", "did:bob", "did:alice")

	body, _ := json.Marshal(map[string]string{"receiptType": "delivered"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/m1/receipt", bytes.NewReader(body)), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("receipt want 200, got %d", rec.Code)
	}

	meta, _ := db.GetMessageMeta(context.Background(), "m1")
	if meta.Status != "delivered" || meta.ReadAt != nil {
		t.Fatalf("expected delivered (not read), got %+v", meta)
	}
	if len(pub.all()) != 0 {
		t.Fatalf("delivered receipt must not signal the sender, got %+v", pub.all())
	}
}

// TestMessageStatus_GETSyncsState: GET .../status returns durable delivery state so
// a reconnecting client can reconcile receipts it missed offline (WO-48).
func TestMessageStatus_GETSyncsState(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(signalsRouter(db, &fakeSignalPublisher{}))
	enqueue(t, db, "m1", "c1", "did:bob", "did:alice")
	if err := db.MarkRead(context.Background(), "m1"); err != nil {
		t.Fatalf("markread: %v", err)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/messages/m1/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		MessageID string  `json:"messageId"`
		Status    string  `json:"status"`
		ReadAt    *string `json:"readAt"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.MessageID != "m1" || resp.Status != "read" || resp.ReadAt == nil {
		t.Fatalf("expected read status with readAt, got %+v", resp)
	}
}

// TestMessageStatus_GETUnknown returns 404 for an unknown message.
func TestMessageStatus_GETUnknown(t *testing.T) {
	mux := v3Mux(signalsRouter(database.NewMemoryDB(), nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/messages/nope/status", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown status want 404, got %d", rec.Code)
	}
}

// TestMessageReact_FansOutToCounterparty: a reaction is pushed to the message's
// counterparty over WS by the server (WO-10), independent of the client.
func TestMessageReact_FansOutToCounterparty(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))
	// alice sent m1 to bob; alice reacts -> bob should be notified.
	enqueue(t, db, "m1", "c1", "did:alice", "did:bob")

	body, _ := json.Marshal(map[string]string{"message_id": "m1", "emoji": "👍"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("react want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	sent := pub.all()
	if len(sent) != 1 || sent[0].to != "did:bob" || sent[0].msg.Type != "reaction" {
		t.Fatalf("expected one reaction signal to did:bob, got %+v", sent)
	}
	var sig ReactionSignal
	_ = json.Unmarshal(sent[0].msg.Payload, &sig)
	if sig.MessageID != "m1" || sig.Emoji != "👍" {
		t.Fatalf("reaction signal payload wrong: %+v", sig)
	}
}

// TestMessageReact_UnknownMessageNoSignal: reacting on a message the relay doesn't
// know (e.g. not queued here) still succeeds durably but pushes no signal.
func TestMessageReact_UnknownMessageNoSignal(t *testing.T) {
	db := database.NewMemoryDB()
	pub := &fakeSignalPublisher{}
	mux := v3Mux(signalsRouter(db, pub))

	body, _ := json.Marshal(map[string]string{"message_id": "ghost", "emoji": "👍"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("react want 200, got %d", rec.Code)
	}
	if len(pub.all()) != 0 {
		t.Fatalf("no signal expected for unknown message, got %+v", pub.all())
	}
}
