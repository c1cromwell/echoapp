package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestRestoreDID_LooksUpLinkedWalletAndIssuesJWT(t *testing.T) {
	rt := NewRouter([]string{"*"})
	w := newTestWallet(t)
	linkWalletDirect(rt, "did:key:zRestoreMe", w.address)

	challenge := restoreChallengeFor(t, rt, w.address)
	restoreRec := secPostJSON(t, rt, "/v1/auth/restore-did", map[string]string{
		"wallet_address":        w.address,
		"wallet_public_key":     w.pubHex,
		"new_device_public_key": p256PublicKeyHex(t),
		"wallet_signature":      w.sign(challenge),
	}, "")
	if restoreRec.Code != http.StatusOK {
		t.Fatalf("restore-did: %d %s", restoreRec.Code, restoreRec.Body.String())
	}
	var resp struct {
		DID         string `json:"did"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(restoreRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.DID != "did:key:zRestoreMe" {
		t.Fatalf("did = %s", resp.DID)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected access_token so the new phone can pull /v3/backup")
	}
}

func TestRestoreDID_UnknownWallet(t *testing.T) {
	rt := NewRouter([]string{"*"})
	w := newTestWallet(t) // a real wallet that no identity is linked to

	challenge := restoreChallengeFor(t, rt, w.address)
	restoreRec := secPostJSON(t, rt, "/v1/auth/restore-did", map[string]string{
		"wallet_address":        w.address,
		"wallet_public_key":     w.pubHex,
		"new_device_public_key": "aa",
		"wallet_signature":      w.sign(challenge),
	}, "")
	if restoreRec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d %s", restoreRec.Code, restoreRec.Body.String())
	}
}

func TestFlushOffline_DurableQueueSurvivesHub(t *testing.T) {
	db := database.NewMemoryDB()
	h := NewHub()
	h.SetDurableOfflineQueue(db)

	payload := []byte(`{"type":"text","to":"did:key:bob"}`)
	if h.deliverOrQueue("did:key:bob", payload, "did:key:alice", "dm:1", false) {
		t.Fatal("offline recipient should not be live-delivered")
	}

	bob := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, 8)}
	h.mu.Lock()
	h.clients[bob.userID] = bob // flushOffline only writes to the registered socket
	h.mu.Unlock()
	h.flushOffline(bob)
	select {
	case got := <-bob.send:
		if string(got) != string(payload) {
			t.Fatalf("payload = %s", got)
		}
	default:
		t.Fatal("expected durable queue flush on reconnect")
	}
}
