package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/thechadcromwell/echoapp/internal/auth"
	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/wallet"
)

// --- helpers ---

type testWallet struct {
	priv    *secp256k1.PrivateKey
	pubHex  string
	address string
}

func newTestWallet(t *testing.T) testWallet {
	t.Helper()
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubHex := hex.EncodeToString(priv.PubKey().SerializeUncompressed())
	addr := wallet.DagAddressFromPubKey(pubHex)
	if addr == "" {
		t.Fatal("could not derive DAG address")
	}
	return testWallet{priv: priv, pubHex: pubHex, address: addr}
}

// sign mirrors dag4 keyStore.sign(): secp256k1 ECDSA (DER, hex) over SHA-512.
func (w testWallet) sign(message string) string {
	digest := sha512.Sum512([]byte(message))
	return hex.EncodeToString(secp256k1ecdsa.Sign(w.priv, digest[:]).Serialize())
}

func secPostJSON(t *testing.T, rt *Router, path string, body interface{}, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	return rec
}

func accessTokenFor(t *testing.T, rt *Router, did string) string {
	t.Helper()
	tok, _, err := rt.TokenService().IssueAccessToken(did, "dev1", 1, "messaging")
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func restoreChallengeFor(t *testing.T, rt *Router, addr string) string {
	t.Helper()
	rec := secPostJSON(t, rt, "/v1/auth/restore-challenge", map[string]string{"wallet_address": addr}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("restore-challenge: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.Challenge == "" {
		t.Fatalf("challenge decode: %v body=%s", err, rec.Body.String())
	}
	return out.Challenge
}

func linkWalletDirect(rt *Router, did, addr string) {
	rt.enrollmentWalletMu.Lock()
	if rt.enrollmentWalletByDID == nil {
		rt.enrollmentWalletByDID = make(map[string]string)
	}
	rt.enrollmentWalletByDID[did] = addr
	rt.enrollmentWalletMu.Unlock()
}

// --- account restore: proof of wallet ownership ---

// A session for a DID must never be issued without a valid wallet signature
// over the challenge. This used to succeed with any string as the signature.
func TestRestoreDID_RejectsForgedSignature(t *testing.T) {
	rt := NewRouter([]string{"*"})
	victim := newTestWallet(t)
	linkWalletDirect(rt, "did:key:zVictim", victim.address)

	attacker := newTestWallet(t)
	cases := map[string]map[string]string{
		"garbage signature": {
			"wallet_public_key": victim.pubHex,
			"wallet_signature":  "garbage",
		},
		"signed by a different key": {
			"wallet_public_key": victim.pubHex,
			"wallet_signature":  "", // filled per challenge below
		},
		"attacker key for victim address": {
			"wallet_public_key": attacker.pubHex,
			"wallet_signature":  "", // filled per challenge below
		},
	}
	for name, fields := range cases {
		t.Run(name, func(t *testing.T) {
			challenge := restoreChallengeFor(t, rt, victim.address)
			if fields["wallet_signature"] == "" {
				fields["wallet_signature"] = attacker.sign(challenge)
			}
			fields["wallet_address"] = victim.address
			fields["new_device_public_key"] = "aa"
			rec := secPostJSON(t, rt, "/v1/auth/restore-did", fields, "")
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("want 401, got %d %s", rec.Code, rec.Body.String())
			}
			if bytes.Contains(rec.Body.Bytes(), []byte("access_token")) {
				t.Fatal("no token may be issued for a forged restore")
			}
		})
	}
}

// The interim, server-derivable address has no private key, so nothing can be
// restored through it — in particular not by asking the server for it.
func TestRestoreDID_ServerDerivableAddressCannotRestore(t *testing.T) {
	rt := NewRouter([]string{"*"})
	did := "did:key:zNeverMetTheAttacker"

	rec := secPostJSON(t, rt, "/v1/enrollment/wallet", map[string]string{"did": did}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("enrollment/wallet: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Address string `json:"address"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)

	attacker := newTestWallet(t)
	challenge := restoreChallengeFor(t, rt, out.Address)
	rec = secPostJSON(t, rt, "/v1/auth/restore-did", map[string]string{
		"wallet_address":        out.Address,
		"wallet_public_key":     attacker.pubHex,
		"new_device_public_key": "aa",
		"wallet_signature":      attacker.sign(challenge),
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d %s", rec.Code, rec.Body.String())
	}
}

// A challenge is single-use: a captured, valid restore request cannot be replayed.
func TestRestoreDID_ChallengeIsSingleUse(t *testing.T) {
	rt := NewRouter([]string{"*"})
	w := newTestWallet(t)
	linkWalletDirect(rt, "did:key:zOwner", w.address)

	challenge := restoreChallengeFor(t, rt, w.address)
	body := map[string]string{
		"wallet_address":        w.address,
		"wallet_public_key":     w.pubHex,
		"new_device_public_key": p256PublicKeyHex(t),
		"wallet_signature":      w.sign(challenge),
	}
	if rec := secPostJSON(t, rt, "/v1/auth/restore-did", body, ""); rec.Code != http.StatusOK {
		t.Fatalf("first restore: %d %s", rec.Code, rec.Body.String())
	}
	if rec := secPostJSON(t, rt, "/v1/auth/restore-did", body, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("replay: want 401, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestRestoreDID_BindsDeviceKey(t *testing.T) {
	rt := NewRouter([]string{"*"})
	w := newTestWallet(t)
	linkWalletDirect(rt, "did:key:zOwner", w.address)
	deviceKey := p256PublicKeyHex(t)
	challenge := restoreChallengeFor(t, rt, w.address)
	rec := secPostJSON(t, rt, "/v1/auth/restore-did", map[string]string{
		"wallet_address":        w.address,
		"wallet_public_key":     w.pubHex,
		"new_device_public_key": deviceKey,
		"wallet_signature":      w.sign(challenge),
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("restore: %d %s", rec.Code, rec.Body.String())
	}
	devices, err := rt.DIDRegistry.ListDevices(context.Background(), "did:key:zOwner")
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0].PublicKeyHex != deviceKey {
		t.Fatalf("devices = %+v", devices)
	}
}

func TestDeleteAccount_RequiresStepUp(t *testing.T) {
	rt := NewRouter([]string{"*"})
	did := "did:key:zDeleteMe"
	req := httptest.NewRequest(http.MethodDelete, "/v1/users/account", nil)
	req.Header.Set("Authorization", "Bearer "+accessTokenFor(t, rt, did))
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestSafetyReport_PersistsForTheCaller(t *testing.T) {
	rt := NewRouter([]string{"*"})
	db := database.NewMemoryDB()
	rt.V3 = &V3Handlers{DB: db}
	token := accessTokenFor(t, rt, "did:key:alice")
	rec := secPostJSON(t, rt, "/v1/safety/reports", map[string]string{
		"subject_did": "did:key:bob",
		"reason":      "harassment",
		"context":     "contact",
	}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d %s", rec.Code, rec.Body.String())
	}
	listed, err := db.ListSafetyReportsByReporter(context.Background(), "did:key:alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].SubjectDID != "did:key:bob" || listed[0].Reason != "harassment" {
		t.Fatalf("reports = %+v", listed)
	}
}

func TestBodyLimit_RejectsOversizedJSON(t *testing.T) {
	rt := NewRouter([]string{"*"})
	body := bytes.Repeat([]byte("a"), 2<<20)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/restore-did", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("want 400 or 413, got %d", rec.Code)
	}
}

// --- endpoints that used to be unauthenticated ---

func TestVIPVerify_RequiresAuthenticatedOwner(t *testing.T) {
	rt := NewRouter([]string{"*"})
	body := map[string]interface{}{"did": "did:key:zVictim", "trust_tier": 5, "evidence_type": "none"}

	if rec := secPostJSON(t, rt, "/v1/auth/vip-verify", body, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: want 401, got %d %s", rec.Code, rec.Body.String())
	}
	other := accessTokenFor(t, rt, "did:key:zSomeoneElse")
	if rec := secPostJSON(t, rt, "/v1/auth/vip-verify", body, other); rec.Code != http.StatusForbidden {
		t.Fatalf("other user's DID: want 403, got %d %s", rec.Code, rec.Body.String())
	}
	owner := accessTokenFor(t, rt, "did:key:zVictim")
	if rec := secPostJSON(t, rt, "/v1/auth/vip-verify", body, owner); rec.Code != http.StatusOK {
		t.Fatalf("owner: want 200, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestLinkWallet_RequiresAuthenticatedOwner(t *testing.T) {
	rt := NewRouter([]string{"*"})
	body := map[string]string{"did": "did:key:zVictim", "wallet_address": "DAGattacker"}

	if rec := secPostJSON(t, rt, "/v1/identity/link-wallet", body, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: want 401, got %d %s", rec.Code, rec.Body.String())
	}
	other := accessTokenFor(t, rt, "did:key:zSomeoneElse")
	if rec := secPostJSON(t, rt, "/v1/identity/link-wallet", body, other); rec.Code != http.StatusForbidden {
		t.Fatalf("other user's DID: want 403, got %d %s", rec.Code, rec.Body.String())
	}
	rt.enrollmentWalletMu.Lock()
	_, linked := rt.enrollmentWalletByDID["did:key:zVictim"]
	rt.enrollmentWalletMu.Unlock()
	if linked {
		t.Fatal("a rejected request must not bind a wallet")
	}
}

// --- per-IP throttle cannot be dodged with a forged X-Forwarded-For ---

func TestClientIP_IgnoresForwardedForUnlessProxyTrusted(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/users/check-username", nil)
	req.RemoteAddr = "203.0.113.7:54321"
	req.Header.Set("X-Forwarded-For", "10.9.8.7")

	if got := clientIPWithHops(req, 0); got != "203.0.113.7" {
		t.Fatalf("no trusted proxy: want the socket address, got %q", got)
	}

	// One trusted proxy appends the address it saw; anything to its left is
	// whatever the client chose to send.
	req.Header.Set("X-Forwarded-For", "10.9.8.7, 198.51.100.4")
	if got := clientIPWithHops(req, 1); got != "198.51.100.4" {
		t.Fatalf("one trusted hop: want the proxy-appended address, got %q", got)
	}
	if got := clientIPWithHops(req, 2); got != "10.9.8.7" {
		t.Fatalf("two trusted hops: got %q", got)
	}
	// Header shorter than the configured chain: fall back to the socket.
	if got := clientIPWithHops(req, 3); got != "203.0.113.7" {
		t.Fatalf("short header: want the socket address, got %q", got)
	}
}

// --- account deletion actually deletes ---

type fakePurger struct {
	called []string
	err    error
}

func (f *fakePurger) PurgeAccount(_ context.Context, did string) (map[string]int64, error) {
	f.called = append(f.called, did)
	if f.err != nil {
		return nil, f.err
	}
	return map[string]int64{"users": 1, "contacts": 3}, nil
}

func deleteAccount(t *testing.T, rt *Router, did, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	elevated, _, err := rt.TokenService().IssueElevatedToken(did, "dev1", 1, string(auth.StepUpDeleteAccount), 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/v1/users/account", nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("X-Elevated-Token", elevated)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	return rec
}

func p256PublicKeyHex(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(elliptic.Marshal(elliptic.P256(), key.X, key.Y))
}

func TestDeleteAccount_PurgesServerData(t *testing.T) {
	rt := NewRouter([]string{"*"})
	go rt.WSHub.Run()
	purger := &fakePurger{}
	rt.AccountPurger = purger
	did := "did:key:zDeleteMe"
	rt.TokenService().StoreRefreshToken(did, auth.GenerateRefreshToken(), "dev1")
	live := registerFakeClient(rt.WSHub, did)

	rec := deleteAccount(t, rt, did, accessTokenFor(t, rt, did))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d %s", rec.Code, rec.Body.String())
	}
	if len(purger.called) != 1 || purger.called[0] != did {
		t.Fatalf("purge calls = %v", purger.called)
	}
	var body map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["server_data_purged"] != true || body["rows_removed"].(float64) != 4 {
		t.Fatalf("unexpected body: %v", body)
	}
	if n := rt.TokenService().GetActiveRefreshTokenCount(did); n != 0 {
		t.Fatalf("expected 0 active refresh tokens, got %d", n)
	}
	if _, open := <-live.send; open {
		t.Fatal("the deleted account's live socket should be closed")
	}
}

// If the purge fails, the session must survive so the user can retry; reporting
// success or signing them out would strand a half-deleted account.
func TestDeleteAccount_PurgeFailureLeavesSessionIntact(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rt.AccountPurger = &fakePurger{err: errors.New("db down")}
	did := "did:key:zDeleteMe"
	rt.TokenService().StoreRefreshToken(did, auth.GenerateRefreshToken(), "dev1")

	rec := deleteAccount(t, rt, did, accessTokenFor(t, rt, did))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d %s", rec.Code, rec.Body.String())
	}
	if n := rt.TokenService().GetActiveRefreshTokenCount(did); n != 1 {
		t.Fatalf("refresh tokens must not be revoked on failure, got %d active", n)
	}
}

// --- hub: reconnect race and offline replay ---

// A reconnect registers the new socket before the old one notices it is dead.
// The old socket's late unregister must not evict the new one.
func TestHub_StaleUnregisterDoesNotEvictReconnectedClient(t *testing.T) {
	h := NewHub()
	go h.Run()

	oldConn := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 8)}
	newConn := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 8)}
	h.register <- oldConn
	h.register <- newConn
	h.unregister <- oldConn
	h.unregister <- oldConn // idempotent: must not double-close

	deadline := time.Now().Add(time.Second)
	for !h.SendToUser("did:key:bob", []byte("hi")) {
		if time.Now().After(deadline) {
			t.Fatal("reconnected client was evicted by the stale socket's unregister")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := <-newConn.send; string(got) != "hi" {
		t.Fatalf("got %q", got)
	}
}

// A backlog larger than the socket's send buffer must arrive completely and in
// order. It used to be marked delivered in one go and truncated at the buffer.
func TestFlushOffline_LargeBacklogDeliveredInOrder(t *testing.T) {
	const backlog = 700
	db := database.NewMemoryDB()
	h := NewHub()
	h.SetDurableOfflineQueue(db)
	for i := 0; i < backlog; i++ {
		if err := db.Enqueue(context.Background(), &database.QueuedMessage{
			MessageID:    fmt.Sprintf("m-%04d", i),
			SenderDID:    "did:key:alice",
			RecipientDID: "did:key:bob",
			Payload:      []byte(fmt.Sprintf(`{"n":%d}`, i)),
		}); err != nil {
			t.Fatal(err)
		}
	}

	bob := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 256)}
	h.mu.Lock()
	h.clients[bob.userID] = bob
	h.mu.Unlock()

	done := make(chan struct{})
	go func() { h.flushOffline(bob); close(done) }()

	for i := 0; i < backlog; i++ {
		select {
		case got := <-bob.send:
			if want := fmt.Sprintf(`{"n":%d}`, i); string(got) != want {
				t.Fatalf("message %d: got %s want %s", i, got, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("only %d of %d queued messages were delivered", i, backlog)
		}
	}
	<-done
}

// A socket that is not the registered one is never written to and never claims
// anything from the queue, so the backlog is intact for the real reconnect.
func TestFlushOffline_UnregisteredSocketLeavesQueueIntact(t *testing.T) {
	db := database.NewMemoryDB()
	h := NewHub()
	h.SetDurableOfflineQueue(db)
	for i := 0; i < 10; i++ {
		_ = db.Enqueue(context.Background(), &database.QueuedMessage{
			MessageID:    fmt.Sprintf("m-%02d", i),
			SenderDID:    "did:key:alice",
			RecipientDID: "did:key:bob",
			Payload:      []byte(fmt.Sprintf(`{"n":%d}`, i)),
		})
	}

	gone := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 256)}
	h.flushOffline(gone)
	if len(gone.send) != 0 {
		t.Fatalf("nothing should be written to an unregistered socket, got %d", len(gone.send))
	}

	back := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 256)}
	h.mu.Lock()
	h.clients[back.userID] = back
	h.mu.Unlock()
	h.flushOffline(back)
	if len(back.send) != 10 {
		t.Fatalf("want all 10 messages after reconnect, got %d", len(back.send))
	}
	if first := <-back.send; string(first) != `{"n":0}` {
		t.Fatalf("order not preserved, first = %s", first)
	}
}

// A frame rejected by the per-DID send budget is reported back to the sender
// rather than vanishing, so the client can retry it.
func TestRouteInbound_RateLimitedSenderIsTold(t *testing.T) {
	h := NewHub()
	h.SetRateLimiter(infra.NewRateLimiter(map[string]infra.RateLimitConfig{
		"message_send": {MaxRequests: 1, Window: time.Minute},
	}))
	alice := registerFakeClient(h, "did:key:alice")
	bob := registerFakeClient(h, "did:key:bob")

	msg := WSMessage{Type: "text", To: "did:key:bob", ConversationID: "dm:1", Payload: json.RawMessage(`{"n":1}`)}
	alice.routeInbound(msg) // within budget
	alice.routeInbound(msg) // over budget

	if len(bob.send) != 1 {
		t.Fatalf("exactly one message should reach the recipient, got %d", len(bob.send))
	}
	select {
	case raw := <-alice.send:
		var frame WSMessage
		if err := json.Unmarshal(raw, &frame); err != nil || frame.Type != "error" {
			t.Fatalf("want an error frame, got %s", raw)
		}
		var detail struct {
			Code           string `json:"code"`
			ConversationID string `json:"conversation_id"`
			RetryAfter     int    `json:"retry_after_seconds"`
		}
		_ = json.Unmarshal(frame.Payload, &detail)
		if detail.Code != "RATE_LIMITED" || detail.ConversationID != "dm:1" || detail.RetryAfter < 1 {
			t.Fatalf("unexpected error detail: %s", frame.Payload)
		}
	default:
		t.Fatal("the sender was not told its message was rate limited")
	}
}
