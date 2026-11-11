package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/agentbus"
	"github.com/thechadcromwell/echoapp/internal/services/bots"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

func agentFixture(t *testing.T) (*V3Handlers, string) {
	t.Helper()
	db := database.NewMemoryDB()
	ctx := context.Background()
	alice := &database.User{DID: "did:key:alice", Username: "alice", TrustTier: 2}
	bob := &database.User{DID: "did:key:bob", Username: "bob", TrustTier: 2}
	if err := db.CreateUser(ctx, alice); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateUser(ctx, bob); err != nil {
		t.Fatal(err)
	}
	bus := agentbus.New()
	bus.RegisterDeviceKey("did:key:alice", vectorPublicKey(t))
	h := &V3Handlers{
		DB:       db,
		Contacts: contacts.NewService(db),
		Agent:    bus,
		Signals:  &fakeBotRelayPublisher{ok: true},
	}
	return h, vectorDID
}

func vectorPublicKey(t *testing.T) *ecdsa.PublicKey {
	t.Helper()
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     mustHexBigInt(t, vectorPubKeyHex[2:66]),
		Y:     mustHexBigInt(t, vectorPubKeyHex[66:]),
	}
}

func signAgent(t *testing.T, method, path string, body []byte) (string, string) {
	t.Helper()
	ts := nowTS()
	return ts, vectorSign(t, method, path, ts, body)
}

func doAgent(t *testing.T, h *V3Handlers, method, path string, body []byte, token string, passkey bool) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Grant "+token)
	}
	if passkey {
		ts, sig := signAgent(t, method, path, body)
		req.Header.Set(headerSenderDID, "did:key:alice")
		req.Header.Set(headerSignature, sig)
		req.Header.Set(headerTimestamp, ts)
	}
	rec := httptest.NewRecorder()
	h.handleAgent(rec, req)
	return rec
}

func issueGrant(t *testing.T, h *V3Handlers, verb string, exp time.Time, signals bool) string {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"agent_did":    "did:key:agent",
		"agent_name":   "Summarizer",
		"verb":         verb,
		"thread_id":    "*",
		"expires_at":   exp.UTC().Format(time.RFC3339),
		"emit_signals": signals,
	})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/grants", body, "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("grant %s: %d %s", verb, rec.Code, rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out.Token
}

func TestAgentResolveAndSealedSend(t *testing.T) {
	h, _ := agentFixture(t)
	token := issueGrant(t, h, agentbus.VerbSend, time.Now().Add(time.Hour), false)
	body, _ := json.Marshal(map[string]string{"name": "@bob"})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/resolve", body, token, false)
	if rec.Code != http.StatusOK {
		t.Fatalf("resolve: %d %s", rec.Code, rec.Body.String())
	}
	var resolved struct {
		DID      string `json:"did"`
		ThreadID string `json:"thread_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resolved); err != nil {
		t.Fatal(err)
	}
	if resolved.DID != "did:key:bob" {
		t.Fatalf("did = %s", resolved.DID)
	}
	wantThread := agentbus.DirectThread("did:key:alice", "did:key:bob")
	if resolved.ThreadID != wantThread {
		t.Fatalf("thread = %s want %s", resolved.ThreadID, wantThread)
	}

	blob := []byte{0x0a, 0x0b, 0x0c}
	sendBody, _ := json.Marshal(map[string]interface{}{
		"to":         "did:key:bob",
		"ciphertext": blob,
	})
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/send", sendBody, token, false)
	if rec.Code != http.StatusOK {
		t.Fatalf("send: %d %s", rec.Code, rec.Body.String())
	}
	q := h.Agent.Queue()
	if len(q) != 1 || !bytes.Equal(q[0].Ciphertext, blob) {
		t.Fatalf("queue = %+v", q)
	}
	if !q[0].Untrusted {
		t.Fatal("a sender who is not a contact was marked trusted")
	}
	if _, err := h.Contacts.AddContact(context.Background(), "did:key:bob", "did:key:alice", "test"); err != nil {
		t.Fatal(err)
	}
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/send", sendBody, token, false)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"untrusted":false`) {
		t.Fatalf("contact send: %d %s", rec.Code, rec.Body.String())
	}

	bad, _ := json.Marshal(map[string]interface{}{"to": "did:key:bob", "text": "hello", "ciphertext": blob})
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/send", bad, token, false)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("plaintext send: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAgentGrantExpiredVerbAndKey(t *testing.T) {
	h, _ := agentFixture(t)
	jwtReq := httptest.NewRequest(http.MethodPost, "/v3/agent/send", strings.NewReader(`{"to":"did:key:bob","ciphertext":"YQ=="}`))
	jwtReq.Header.Set("Authorization", "Bearer user-jwt")
	jwtRec := httptest.NewRecorder()
	h.handleAgent(jwtRec, jwtReq)
	if jwtRec.Code != http.StatusUnauthorized || !strings.Contains(jwtRec.Body.String(), "AGENT_GRANT_REQUIRED") {
		t.Fatalf("user jwt: %d %s", jwtRec.Code, jwtRec.Body.String())
	}

	expired := issueGrant(t, h, agentbus.VerbSend, time.Now().Add(-time.Minute), false)
	body, _ := json.Marshal(map[string]interface{}{"to": "did:key:bob", "ciphertext": []byte{1}})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/send", body, expired, false)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "GRANT_EXPIRED") {
		t.Fatalf("expired: %d %s", rec.Code, rec.Body.String())
	}

	sendTok := issueGrant(t, h, agentbus.VerbSend, time.Now().Add(time.Hour), false)
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/search", []byte(`{}`), sendTok, false)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "VERB_MISMATCH") {
		t.Fatalf("verb: %d %s", rec.Code, rec.Body.String())
	}

	other, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	h.Agent.RegisterDeviceKey("did:key:alice", &other.PublicKey)
	grantBody, _ := json.Marshal(map[string]interface{}{
		"agent_did":  "did:key:agent",
		"verb":       "send",
		"expires_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/grants", grantBody, "", true)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong key: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAgentAuditOmitsPayload(t *testing.T) {
	h, _ := agentFixture(t)
	token := issueGrant(t, h, agentbus.VerbSend, time.Now().Add(time.Hour), false)
	secret := []byte("ciphertext-secret-blob")
	body, _ := json.Marshal(map[string]interface{}{"to": "did:key:bob", "ciphertext": secret})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/send", body, token, false)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	lines := h.Agent.AuditLines()
	if len(lines) != 1 {
		t.Fatalf("audit lines = %v", lines)
	}
	if strings.Contains(lines[0], "ciphertext-secret-blob") || strings.Contains(lines[0], "secret") {
		t.Fatalf("audit leaked payload: %s", lines[0])
	}
	if !strings.Contains(lines[0], "verb=send") {
		t.Fatalf("audit = %s", lines[0])
	}
}

func TestBotRelayRejectsPlaintext(t *testing.T) {
	store := bots.NewInstallStore()
	tokens := bots.NewTokenValidator()
	bot := bots.DefaultCatalog()[0]
	tokens.SetToken(bot.BotDID, "test-token")
	store.Install("did:alice", bot.BotDID, []bots.Permission{bots.PermSendMessage})
	h := &V3Handlers{
		Bots:           store,
		BotTokens:      tokens,
		BotRateLimiter: bots.NewRateLimiter(100, 0),
		Signals:        &fakeBotRelayPublisher{ok: true},
	}
	body, _ := json.Marshal(map[string]interface{}{
		"recipient_did":   "did:alice",
		"conversation_id": "dm:1",
		"text":            "plaintext",
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/bots/relay/message", bytes.NewReader(body))
	req.Header.Set("X-Bot-DID", bot.BotDID)
	req.Header.Set("X-Bot-Token", "test-token")
	rec := httptest.NewRecorder()
	h.handleBotsSubroute(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("plaintext bot: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAgentConfirmPayOnce(t *testing.T) {
	h, _ := agentFixture(t)
	token := issueGrant(t, h, agentbus.VerbPay, time.Now().Add(time.Hour), false)
	blob := []byte{0x11, 0x22}
	body, _ := json.Marshal(map[string]interface{}{"to": "did:key:bob", "ciphertext": blob})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/send", body, token, false)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "needs_confirmation") {
		t.Fatalf("confirm gate: %d %s", rec.Code, rec.Body.String())
	}
	if len(h.Agent.Queue()) != 0 {
		t.Fatal("pay was relayed before confirmation")
	}
	var gate struct {
		ChallengeID string `json:"challenge_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &gate); err != nil || gate.ChallengeID == "" {
		t.Fatal(rec.Body.String())
	}
	// Agent grant alone cannot confirm.
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/confirm", []byte(`{"challenge_id":"`+gate.ChallengeID+`"}`), token, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("agent confirm: %d %s", rec.Code, rec.Body.String())
	}
	confirmBody := []byte(`{"challenge_id":"` + gate.ChallengeID + `"}`)
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/confirm", confirmBody, "", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("user confirm: %d %s", rec.Code, rec.Body.String())
	}
	if len(h.Agent.Queue()) != 1 || !bytes.Equal(h.Agent.Queue()[0].Ciphertext, blob) {
		t.Fatalf("queue after confirm = %+v", h.Agent.Queue())
	}
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/confirm", confirmBody, "", true)
	if rec.Code != http.StatusConflict {
		t.Fatalf("replay: %d %s", rec.Code, rec.Body.String())
	}
	if len(h.Agent.Queue()) != 1 {
		t.Fatal("replay queued a second copy")
	}
}

func TestAgentDeliveryHintAndSignals(t *testing.T) {
	h, _ := agentFixture(t)
	pub := h.Signals.(*fakeBotRelayPublisher)
	token := issueGrant(t, h, agentbus.VerbSend, time.Now().Add(time.Hour), false)
	exp := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"to":          "did:key:bob",
		"ciphertext":  []byte{9},
		"expires_at":  exp,
		"render_hint": "spoken",
	})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/send", body, token, false)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	q := h.Agent.Queue()
	if q[0].RenderHint != "spoken" || q[0].ExpiresAt != exp {
		t.Fatalf("delivery = %+v", q[0])
	}
	var wire struct {
		RenderHint string `json:"render_hint"`
		ExpiresAt  string `json:"expires_at"`
	}
	if err := json.Unmarshal(pub.lastMsg.Payload, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.RenderHint != "spoken" || wire.ExpiresAt != exp {
		t.Fatalf("wire = %+v", wire)
	}
	sigBody, _ := json.Marshal(map[string]string{"to": "did:key:bob", "type": "typing"})
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/signal", sigBody, token, false)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"published":false`) {
		t.Fatalf("silent signals: %d %s", rec.Code, rec.Body.String())
	}
	if pub.lastMsg.Type == "typing" {
		t.Fatal("typing was fanned out without emit_signals")
	}
}

func TestAgentRecoveryDoesNotReconstruct(t *testing.T) {
	h, _ := agentFixture(t)
	if err := h.Agent.Reconstruct(); err != agentbus.ErrNoReconstruct {
		t.Fatalf("reconstruct = %v", err)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"index":      1,
		"holder_did": "did:key:guardian",
	})
	rec := doAgent(t, h, http.MethodPost, "/v3/agent/recovery/shares", body, "", true)
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	meta := h.Agent.ShareMeta()
	if len(meta) != 1 || meta[0].HolderDID != "did:key:guardian" {
		t.Fatalf("meta = %+v", meta)
	}
	leaky, _ := json.Marshal(map[string]interface{}{
		"index":       2,
		"holder_did":  "did:key:guardian",
		"share_bytes": "aabb",
	})
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/recovery/shares", leaky, "", true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("share bytes: %d %s", rec.Code, rec.Body.String())
	}

	token := issueGrant(t, h, agentbus.VerbRotateKey, time.Now().Add(time.Hour), false)
	rec = doAgent(t, h, http.MethodPost, "/v3/agent/rotate-key", []byte(`{}`), token, false)
	if rec.Code != http.StatusConflict {
		t.Fatalf("rotate: %d %s", rec.Code, rec.Body.String())
	}
	if h.Agent.KeyGeneration("did:key:alice") != 0 {
		t.Fatal("grant rotated the key without the user")
	}
}
