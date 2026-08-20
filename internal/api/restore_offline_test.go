package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestRestoreDID_LooksUpLinkedWalletAndIssuesJWT(t *testing.T) {
	rt := NewRouter([]string{"*"})
	linkBody, _ := json.Marshal(map[string]string{
		"did":            "did:key:zRestoreMe",
		"wallet_address": "DAGrestoreaddr",
	})
	linkReq := httptest.NewRequest(http.MethodPost, "/v1/identity/link-wallet", bytes.NewReader(linkBody))
	linkRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(linkRec, linkReq)
	if linkRec.Code != http.StatusOK {
		t.Fatalf("link-wallet: %d %s", linkRec.Code, linkRec.Body.String())
	}

	chalBody, _ := json.Marshal(map[string]string{"wallet_address": "DAGrestoreaddr"})
	chalReq := httptest.NewRequest(http.MethodPost, "/v1/auth/restore-challenge", bytes.NewReader(chalBody))
	chalRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(chalRec, chalReq)
	if chalRec.Code != http.StatusOK {
		t.Fatalf("restore-challenge: %d %s", chalRec.Code, chalRec.Body.String())
	}
	var chal struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(chalRec.Body.Bytes(), &chal); err != nil || chal.Challenge == "" {
		t.Fatalf("challenge decode: %v body=%s", err, chalRec.Body.String())
	}

	restoreBody, _ := json.Marshal(map[string]string{
		"wallet_address":        "DAGrestoreaddr",
		"new_device_public_key": "aa",
		"wallet_signature":      "sig",
	})
	restoreReq := httptest.NewRequest(http.MethodPost, "/v1/auth/restore-did", bytes.NewReader(restoreBody))
	restoreRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(restoreRec, restoreReq)
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
	chalBody, _ := json.Marshal(map[string]string{"wallet_address": "DAGnobody"})
	chalReq := httptest.NewRequest(http.MethodPost, "/v1/auth/restore-challenge", bytes.NewReader(chalBody))
	rt.Handler().ServeHTTP(httptest.NewRecorder(), chalReq)

	restoreBody, _ := json.Marshal(map[string]string{
		"wallet_address":        "DAGnobody",
		"new_device_public_key": "aa",
		"wallet_signature":      "sig",
	})
	restoreReq := httptest.NewRequest(http.MethodPost, "/v1/auth/restore-did", bytes.NewReader(restoreBody))
	restoreRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(restoreRec, restoreReq)
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
