package integration

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/services/relay"
	"github.com/thechadcromwell/echoapp/internal/testutil"
)

func TestE2E_MessageSendReceive(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	tokenA := ts.IssueTestToken("alice")
	tokenB := ts.IssueTestToken("bob")

	connA := wsConnectE2E(t, ts.BaseURL, tokenA)
	defer connA.Close()
	connB := wsConnectE2E(t, ts.BaseURL, tokenB)
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	payload, _ := json.Marshal(map[string]string{
		"text":  "Hello Bob!",
		"nonce": "abc123",
	})
	msg := api.WSMessage{
		Type:           "text",
		To:             "bob",
		ConversationID: "conv-e2e-1",
		Payload:        payload,
	}
	data, _ := json.Marshal(msg)
	connA.WriteMessage(websocket.TextMessage, data)

	connB.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, raw, err := connB.ReadMessage()
	if err != nil {
		t.Fatalf("Bob failed to receive message: %v", err)
	}

	var received api.WSMessage
	json.Unmarshal(raw, &received)

	if received.From != "alice" {
		t.Errorf("expected from=alice, got %s", received.From)
	}
	if received.ConversationID != "conv-e2e-1" {
		t.Errorf("expected conv-e2e-1, got %s", received.ConversationID)
	}
	if received.Timestamp == "" {
		t.Error("timestamp should be set by server")
	}

	var content map[string]string
	json.Unmarshal(received.Payload, &content)
	if content["text"] != "Hello Bob!" {
		t.Errorf("payload mismatch: %v", content)
	}
}

func TestE2E_RelayOfflineQueue(t *testing.T) {
	rl := infra.NewRateLimiter(infra.DefaultRateLimits())
	svc := relay.NewRelayService(rl)

	svc.Connect("did:echo:alice")

	for i := 0; i < 3; i++ {
		result, err := svc.Relay(relay.RelayMessage{
			MessageID:     "msg-" + string(rune('A'+i)),
			SenderDID:     "did:echo:alice",
			RecipientDIDs: []string{"did:echo:bob"},
			ContentType:   "application/octet-stream",
			EncryptedBlob: []byte("encrypted-content"),
			Commitment:    []byte("commitment-hash"),
			Timestamp:     time.Now(),
		})
		if err != nil {
			t.Fatalf("relay message %d failed: %v", i, err)
		}
		if result.Recipients["did:echo:bob"] != "queued" {
			t.Errorf("msg %d: expected queued, got %s", i, result.Recipients["did:echo:bob"])
		}
	}

	svc.Connect("did:echo:bob")
	messages, err := svc.DrainOfflineQueue("did:echo:bob")
	if err != nil {
		t.Fatalf("drain failed: %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 queued messages, got %d", len(messages))
	}

	remaining, _ := svc.DrainOfflineQueue("did:echo:bob")
	if len(remaining) != 0 {
		t.Errorf("expected empty queue after drain, got %d", len(remaining))
	}
}

func TestE2E_RelayExpiredMessage(t *testing.T) {
	rl := infra.NewRateLimiter(infra.DefaultRateLimits())
	svc := relay.NewRelayService(rl)

	pastTime := time.Now().Add(-1 * time.Hour)
	_, err := svc.Relay(relay.RelayMessage{
		MessageID:     "msg-expired",
		SenderDID:     "did:echo:alice",
		RecipientDIDs: []string{"did:echo:bob"},
		EncryptedBlob: []byte("old-content"),
		Timestamp:     time.Now(),
		ExpiresAt:     &pastTime,
	})

	if err == nil {
		t.Fatal("expired message should be rejected")
	}
}

func TestE2E_BidirectionalConversation(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	connA := wsConnectE2E(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer connA.Close()
	connB := wsConnectE2E(t, ts.BaseURL, ts.IssueTestToken("bob"))
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	sendJSON(t, connA, api.WSMessage{
		Type:           "text",
		To:             "bob",
		ConversationID: "conv-bidir",
		Payload:        json.RawMessage(`"Hey Bob"`),
	})

	msgFromAlice := readJSON(t, connB, 3*time.Second)
	if msgFromAlice.From != "alice" {
		t.Errorf("expected from=alice, got %s", msgFromAlice.From)
	}

	sendJSON(t, connB, api.WSMessage{
		Type:           "text",
		To:             "alice",
		ConversationID: "conv-bidir",
		Payload:        json.RawMessage(`"Hey Alice!"`),
	})

	msgFromBob := readJSON(t, connA, 3*time.Second)
	if msgFromBob.From != "bob" {
		t.Errorf("expected from=bob, got %s", msgFromBob.From)
	}

	var text string
	json.Unmarshal(msgFromBob.Payload, &text)
	if text != "Hey Alice!" {
		t.Errorf("expected 'Hey Alice!', got %q", text)
	}
}

func wsConnectE2E(t *testing.T, baseURL, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws?token=" + token
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("ws dial (HTTP %d): %v", resp.StatusCode, err)
		}
		t.Fatalf("ws dial: %v", err)
	}
	return conn
}

func sendJSON(t *testing.T, conn *websocket.Conn, msg api.WSMessage) {
	t.Helper()
	data, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("ws send: %v", err)
	}
}

func readJSON(t *testing.T, conn *websocket.Conn, timeout time.Duration) api.WSMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ws read: %v", err)
	}
	var msg api.WSMessage
	json.Unmarshal(raw, &msg)
	return msg
}
