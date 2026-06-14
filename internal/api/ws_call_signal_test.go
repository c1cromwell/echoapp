package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCallsICEServers(t *testing.T) {
	mux := http.NewServeMux()
	(&V3Handlers{}).RegisterV3Routes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v3/calls/ice-servers", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	servers, ok := resp["ice_servers"].([]any)
	if !ok || len(servers) == 0 {
		t.Fatal("expected ice_servers array")
	}
}

func TestCallSignalRoutesToPeer(t *testing.T) {
	hub := NewHub()
	alice := &Client{hub: hub, userID: "did:alice", send: make(chan []byte, 2)}
	bob := &Client{hub: hub, userID: "did:bob", send: make(chan []byte, 2)}
	hub.mu.Lock()
	hub.clients["did:bob"] = bob
	hub.mu.Unlock()

	payload, _ := json.Marshal(CallSignal{
		CallID:   "call-1",
		Action:   "offer",
		CallType: "voice",
		SDP:      "v=0",
	})
	msg := WSMessage{
		Type:           "call_signal",
		To:             "did:bob",
		ConversationID: "dm:alice-bob",
		Payload:        payload,
	}
	alice.routeInbound(msg)

	select {
	case raw := <-bob.send:
		var relayed WSMessage
		if err := json.Unmarshal(raw, &relayed); err != nil {
			t.Fatal(err)
		}
		if relayed.Type != "call_signal" || relayed.From != "did:alice" {
			t.Fatalf("unexpected relay: %+v", relayed)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for call_signal delivery")
	}
}

func TestCallSignalDroppedWithoutRecipient(t *testing.T) {
	hub := NewHub()
	alice := &Client{hub: hub, userID: "did:alice", send: make(chan []byte, 1)}
	bob := &Client{hub: hub, userID: "did:bob", send: make(chan []byte, 1)}
	hub.mu.Lock()
	hub.clients["did:bob"] = bob
	hub.mu.Unlock()

	payload, _ := json.Marshal(CallSignal{CallID: "c1", Action: "offer"})
	msg := WSMessage{Type: "call_signal", Payload: payload}
	alice.routeInbound(msg)

	select {
	case <-bob.send:
		t.Fatal("call_signal without to should not deliver")
	default:
	}
}
