package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/messaging"
	"github.com/thechadcromwell/echoapp/internal/services/relay"
)

// sealedFixture is a gateway with one recipient (Bob) who has registered a delivery key.
func sealedFixture(t *testing.T) (rt *Router, db *database.MemoryDB, bobKey []byte) {
	t.Helper()
	db = database.NewMemoryDB()
	rt = NewRouter(nil)
	rt.WSHub = NewHub()
	rt.WSHub.SetDurableOfflineQueue(db)
	rt.DeliveryKeys = messaging.NewMemoryDeliveryKeyStore()
	bobKey = make([]byte, 32)
	if _, err := rand.Read(bobKey); err != nil {
		t.Fatal(err)
	}
	if err := rt.DeliveryKeys.Register(context.Background(), "did:key:zBob", messaging.HashDeliveryKey(bobKey), false); err != nil {
		t.Fatal(err)
	}
	return rt, db, bobKey
}

func deliver(rt *Router, key []byte, payload string, headers map[string]string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]interface{}{"delivery_key": hex.EncodeToString(key), "payload": json.RawMessage(payload)})
	req := httptest.NewRequest(http.MethodPost, sealedDeliverPath, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	rt.handleSealedDeliver(rec, req)
	return rec
}

// The whole point: a message arrives for Bob, and neither the envelope nor the stored row says
// who sent it or which conversation it belongs to.
func TestSealedDeliver_ReachesRecipientWithoutNamingTheSender(t *testing.T) {
	rt, db, bobKey := sealedFixture(t)
	rec := deliver(rt, bobKey, `{"v":2,"ciphertext":"b3BhcXVl"}`, nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	rows, err := db.Dequeue(context.Background(), "did:key:zBob", 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queued for Bob: %d rows, %v", len(rows), err)
	}
	if rows[0].SenderDID != "" || rows[0].ConversationID != "" {
		t.Fatalf("stored row names sender %q / conversation %q", rows[0].SenderDID, rows[0].ConversationID)
	}
	var env WSMessage
	if err := json.Unmarshal(rows[0].Payload, &env); err != nil {
		t.Fatal(err)
	}
	if env.Type != "sealed_text" || env.To != "did:key:zBob" {
		t.Fatalf("envelope: %+v", env)
	}
	if env.From != "" {
		t.Fatalf("envelope names a sender: %q", env.From)
	}
	if env.ConversationID != sealedConversationID {
		t.Fatalf("conversation id on the wire = %q, want the constant %q", env.ConversationID, sealedConversationID)
	}
	if string(env.Payload) != `{"v":2,"ciphertext":"b3BhcXVl"}` {
		t.Fatalf("payload was altered: %s", env.Payload)
	}
}

// A logged-in caller is treated exactly like an anonymous one: the endpoint must not pick up an
// identity from a token, a header or the request context and attach it to the message.
func TestSealedDeliver_IgnoresAnyIdentityTheCallerPresents(t *testing.T) {
	rt, db, bobKey := sealedFixture(t)
	body, _ := json.Marshal(map[string]interface{}{"delivery_key": hex.EncodeToString(bobKey), "payload": json.RawMessage(`{"ciphertext":"eA=="}`)})
	req := httptest.NewRequest(http.MethodPost, sealedDeliverPath, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer some-token")
	req.Header.Set("X-Sender-DID", "did:key:zAlice")
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zAlice"))
	rec := httptest.NewRecorder()
	rt.handleSealedDeliver(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	rows, _ := db.Dequeue(context.Background(), "did:key:zBob", 10)
	if len(rows) != 1 {
		t.Fatalf("%d rows", len(rows))
	}
	stored, _ := json.Marshal(rows[0])
	if strings.Contains(string(stored), "zAlice") || strings.Contains(string(rows[0].Payload), "zAlice") {
		t.Fatalf("the caller's identity reached storage: %s / %s", stored, rows[0].Payload)
	}
}

func TestSealedDeliver_IsReachableWithoutLogin(t *testing.T) {
	if !publicPaths[sealedDeliverPath] {
		t.Fatal("sealed delivery requires a login, which would tell the server who is sending")
	}
}

func TestSealedDeliver_Refusals(t *testing.T) {
	rt, db, bobKey := sealedFixture(t)
	stranger := make([]byte, 32)
	_, _ = rand.Read(stranger)

	cases := []struct {
		name string
		rec  *httptest.ResponseRecorder
		want int
	}{
		{"a key no account registered", deliver(rt, stranger, `{"ciphertext":"eA=="}`, nil), http.StatusNotFound},
		{"a short key", deliver(rt, bobKey[:16], `{"ciphertext":"eA=="}`, nil), http.StatusBadRequest},
		{"no payload", deliver(rt, bobKey, `null`, nil), http.StatusBadRequest},
	}
	for _, c := range cases {
		if c.rec.Code != c.want {
			t.Errorf("%s: status %d, want %d (%s)", c.name, c.rec.Code, c.want, c.rec.Body.String())
		}
	}
	get := httptest.NewRecorder()
	rt.handleSealedDeliver(get, httptest.NewRequest(http.MethodGet, sealedDeliverPath, nil))
	if get.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET: status %d", get.Code)
	}
	huge := deliver(rt, bobKey, `{"ciphertext":"`+strings.Repeat("A", maxSealedBody)+`"}`, nil)
	if huge.Code != http.StatusBadRequest {
		t.Errorf("oversized body: status %d", huge.Code)
	}
	if rows, _ := db.Dequeue(context.Background(), "did:key:zBob", 10); len(rows) != 0 {
		t.Fatalf("a refused request still delivered %d message(s)", len(rows))
	}
}

func TestDeliveryKeyRegistration(t *testing.T) {
	h := &V3Handlers{DeliveryKeys: messaging.NewMemoryDeliveryKeyStore()}
	put := func(did, hash string) int {
		body, _ := json.Marshal(map[string]string{"key_hash": hash})
		req := httptest.NewRequest(http.MethodPut, "/v3/messages/delivery-key", bytes.NewReader(body))
		if did != "" {
			req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
		}
		rec := httptest.NewRecorder()
		h.handleDeliveryKey(rec, req)
		return rec.Code
	}
	hash := func(b byte) string { return messaging.HashDeliveryKey(bytes.Repeat([]byte{b}, 32)) }

	if code := put("", hash(1)); code != http.StatusUnauthorized {
		t.Errorf("registering without a login: %d", code)
	}
	if code := put("did:key:zBob", "not-a-hash"); code != http.StatusBadRequest {
		t.Errorf("malformed hash: %d", code)
	}
	if code := put("did:key:zBob", hash(1)); code != http.StatusOK {
		t.Fatalf("first key: %d", code)
	}
	if code := put("did:key:zBob", hash(1)); code != http.StatusOK {
		t.Errorf("registering the same key again must succeed: %d", code)
	}
	// Someone else cannot claim Bob's key and so redirect his mail to themselves.
	if code := put("did:key:zMallory", hash(1)); code != http.StatusConflict {
		t.Errorf("another account registering Bob's key: %d, want 409", code)
	}
	// Rotation: the two newest keys work, the oldest stops.
	put("did:key:zBob", hash(2))
	put("did:key:zBob", hash(3))
	ctx := context.Background()
	for b, want := range map[byte]bool{1: false, 2: true, 3: true} {
		_, ok, _ := h.DeliveryKeys.Resolve(ctx, hash(b))
		if ok != want {
			t.Errorf("key %d valid = %v, want %v", b, ok, want)
		}
	}
}

// Through the real middleware chain: no token, and the message is accepted. The neighbouring
// registration route, which does need to know who is calling, still refuses without one.
func TestSealedDeliver_ThroughTheRealHandlerChain(t *testing.T) {
	rt, db, bobKey := sealedFixture(t)
	body, _ := json.Marshal(map[string]interface{}{"delivery_key": hex.EncodeToString(bobKey), "payload": json.RawMessage(`{"ciphertext":"eA=="}`)})
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, sealedDeliverPath, bytes.NewReader(body)))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("anonymous sealed delivery: status %d: %s", rec.Code, rec.Body.String())
	}
	if rows, _ := db.Dequeue(context.Background(), "did:key:zBob", 10); len(rows) != 1 {
		t.Fatalf("%d message(s) reached Bob", len(rows))
	}
	rec = httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/v3/messages/delivery-key", strings.NewReader(`{"key_hash":"x"}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("registering a delivery key without a login: status %d, want 401", rec.Code)
	}
}

// Blocking: registering a new key with revoke_others makes every earlier key stop at once, so the
// copy a blocked contact holds is useless immediately.
func TestDeliveryKeyRevokeOthers(t *testing.T) {
	store := messaging.NewMemoryDeliveryKeyStore()
	ctx := context.Background()
	hash := func(b byte) string { return messaging.HashDeliveryKey(bytes.Repeat([]byte{b}, 32)) }
	for _, b := range []byte{1, 2} {
		if err := store.Register(ctx, "did:key:zBob", hash(b), false); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Register(ctx, "did:key:zBob", hash(3), true); err != nil {
		t.Fatal(err)
	}
	for b, want := range map[byte]bool{1: false, 2: false, 3: true} {
		if _, ok, _ := store.Resolve(ctx, hash(b)); ok != want {
			t.Errorf("key %d valid = %v, want %v", b, ok, want)
		}
	}
}

// A sealed message can still be anchored: the commitment is recorded under the message's random
// id with no sender beside it.
func TestSealedDeliver_AnchorsWithoutASender(t *testing.T) {
	rt, _, bobKey := sealedFixture(t)
	batch := relay.NewCommitmentBatch()
	rt.WSHub.SetCommitmentBatch(batch)
	body, _ := json.Marshal(map[string]interface{}{
		"delivery_key":    hex.EncodeToString(bobKey),
		"payload":         json.RawMessage(`{"ciphertext":"eA=="}`),
		"message_id":      "6f1c2a34-0000-4000-8000-000000000001",
		"commitment_hash": strings.Repeat("ab", 32),
	})
	rec := httptest.NewRecorder()
	rt.handleSealedDeliver(rec, httptest.NewRequest(http.MethodPost, sealedDeliverPath, bytes.NewReader(body)))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d", rec.Code)
	}
	entries := batch.Flush()
	if len(entries) != 1 || entries[0].MessageID != "6f1c2a34-0000-4000-8000-000000000001" {
		t.Fatalf("commitments: %+v", entries)
	}
	if entries[0].SenderDID != "" {
		t.Fatalf("commitment names a sender: %q", entries[0].SenderDID)
	}
}
