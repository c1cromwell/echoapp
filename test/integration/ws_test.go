package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thechadcromwell/echoapp/internal/api"
	"github.com/thechadcromwell/echoapp/internal/testutil"
)

// wsConnect opens a WebSocket connection to the test server.
func wsConnect(t *testing.T, baseURL, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws?token=" + token

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("ws dial failed (HTTP %d): %v", resp.StatusCode, err)
		}
		t.Fatalf("ws dial failed: %v", err)
	}
	return conn
}

// readWSMessage reads a single JSON message with a timeout.
func readWSMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) api.WSMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))

	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ws read failed: %v", err)
	}

	var msg api.WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("ws unmarshal failed: %v (raw: %s)", err, string(raw))
	}
	return msg
}

// sendWSMessage sends a JSON message over the WebSocket.
func sendWSMessage(t *testing.T, conn *websocket.Conn, msg api.WSMessage) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("ws marshal failed: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("ws write failed: %v", err)
	}
}

// --- Tests ---

// TestWS_Connect verifies that a client can establish a WebSocket connection.
func TestWS_Connect(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	conn := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer conn.Close()

	// Connection succeeded — no error from wsConnect
}

// TestWS_ConnectRequiresAuth verifies that connecting without a token is rejected.
func TestWS_ConnectRequiresAuth(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	wsURL := "ws" + strings.TrimPrefix(ts.BaseURL, "http") + "/ws"

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected connection to fail without token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// TestWS_DirectMessage verifies that user A can send a message
// directly to user B and that B receives it.
func TestWS_DirectMessage(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	// Connect two users
	connA := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer connA.Close()

	connB := wsConnect(t, ts.BaseURL, ts.IssueTestToken("bob"))
	defer connB.Close()

	// Small delay to ensure both clients are registered
	time.Sleep(50 * time.Millisecond)

	// Alice sends a direct message to Bob
	payload, _ := json.Marshal("Hello Bob, this is Alice!")
	sendWSMessage(t, connA, api.WSMessage{
		Type:           "text",
		To:             "bob",
		ConversationID: "conv-1",
		Payload:        payload,
	})

	// Bob should receive the message
	msg := readWSMessage(t, connB, 3*time.Second)

	if msg.Type != "text" {
		t.Errorf("expected type 'text', got %q", msg.Type)
	}
	if msg.From != "alice" {
		t.Errorf("expected from 'alice', got %q", msg.From)
	}
	if msg.ConversationID != "conv-1" {
		t.Errorf("expected conversation_id 'conv-1', got %q", msg.ConversationID)
	}

	var text string
	json.Unmarshal(msg.Payload, &text)
	if text != "Hello Bob, this is Alice!" {
		t.Errorf("expected payload 'Hello Bob, this is Alice!', got %q", text)
	}
	if msg.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

// TestWS_UnaddressedMessageNotBroadcast verifies that a client message without
// a "to" field is dropped rather than delivered to every connected client.
func TestWS_UnaddressedMessageNotBroadcast(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	connA := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer connA.Close()

	connB := wsConnect(t, ts.BaseURL, ts.IssueTestToken("bob"))
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	payload, _ := json.Marshal("Hello everyone!")
	sendWSMessage(t, connA, api.WSMessage{
		Type:    "text",
		Payload: payload,
	})

	connB.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if _, raw, err := connB.ReadMessage(); err == nil {
		t.Fatalf("Bob must not receive an unaddressed message, got %s", raw)
	}
}

// TestWS_ConnectRejectsInvalidToken verifies that a syntactically present but
// invalid token cannot open a socket (it used to join the hub with an empty
// identity).
func TestWS_ConnectRejectsInvalidToken(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	wsURL := "ws" + strings.TrimPrefix(ts.BaseURL, "http") + "/ws?token=not-a-real-token"

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		conn.Close()
		t.Fatal("expected connection to fail with an invalid token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", resp)
	}
}

// TestWS_PingPong verifies that the server responds to a
// control "ping" message with a "pong".
func TestWS_PingPong(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	conn := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Send a control ping
	pingPayload, _ := json.Marshal(api.WSControlMessage{Action: "ping"})
	sendWSMessage(t, conn, api.WSMessage{
		Type:    "control",
		Payload: pingPayload,
	})

	// Should receive a pong back
	msg := readWSMessage(t, conn, 3*time.Second)

	if msg.Type != "control" {
		t.Fatalf("expected type 'control', got %q", msg.Type)
	}
	if msg.From != "server" {
		t.Errorf("expected from 'server', got %q", msg.From)
	}

	var ctrl api.WSControlMessage
	json.Unmarshal(msg.Payload, &ctrl)
	if ctrl.Action != "pong" {
		t.Errorf("expected action 'pong', got %q", ctrl.Action)
	}
}

// TestWS_MessageNotDeliveredToSelf verifies that a direct message
// is only delivered to the recipient, not echoed to the sender.
func TestWS_MessageNotDeliveredToSelf(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	connA := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer connA.Close()

	connB := wsConnect(t, ts.BaseURL, ts.IssueTestToken("bob"))
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	// Alice sends DM to Bob
	payload, _ := json.Marshal("DM for Bob only")
	sendWSMessage(t, connA, api.WSMessage{
		Type:    "text",
		To:      "bob",
		Payload: payload,
	})

	// Bob should receive it
	_ = readWSMessage(t, connB, 3*time.Second)

	// Alice should NOT receive her own message back (set a short deadline)
	connA.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := connA.ReadMessage()
	if err == nil {
		t.Error("sender should NOT receive their own direct message")
	}
}

// TestWS_MultipleMessagesInOrder verifies that multiple messages
// arrive in the order they were sent.
func TestWS_MultipleMessagesInOrder(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	connA := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer connA.Close()

	connB := wsConnect(t, ts.BaseURL, ts.IssueTestToken("bob"))
	defer connB.Close()

	time.Sleep(50 * time.Millisecond)

	// Alice sends 5 messages to Bob
	for i := 0; i < 5; i++ {
		payload, _ := json.Marshal(i)
		sendWSMessage(t, connA, api.WSMessage{
			Type:    "text",
			To:      "bob",
			Payload: payload,
		})
	}

	// Bob should receive them in order
	for i := 0; i < 5; i++ {
		msg := readWSMessage(t, connB, 3*time.Second)
		var num int
		json.Unmarshal(msg.Payload, &num)
		if num != i {
			t.Errorf("expected message %d, got %d", i, num)
		}
	}
}

// TestWS_DisconnectAndReconnect verifies that a user can
// disconnect and reconnect without issues.
func TestWS_DisconnectAndReconnect(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	// Connect
	conn := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))

	// Disconnect
	conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Reconnect
	conn2 := wsConnect(t, ts.BaseURL, ts.IssueTestToken("alice"))
	defer conn2.Close()

	// Should still work — send/receive a ping/pong
	pingPayload, _ := json.Marshal(api.WSControlMessage{Action: "ping"})
	sendWSMessage(t, conn2, api.WSMessage{
		Type:    "control",
		Payload: pingPayload,
	})

	msg := readWSMessage(t, conn2, 3*time.Second)
	var ctrl api.WSControlMessage
	json.Unmarshal(msg.Payload, &ctrl)
	if ctrl.Action != "pong" {
		t.Errorf("expected pong after reconnect, got %q", ctrl.Action)
	}
}
