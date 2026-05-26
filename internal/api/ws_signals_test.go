package api

import (
	"encoding/json"
	"testing"
)

func registerFakeClient(h *Hub, userID string) *Client {
	c := &Client{hub: h, userID: userID, send: make(chan []byte, 8)}
	h.mu.Lock()
	h.clients[userID] = c
	h.mu.Unlock()
	return c
}

func typingMsg(from, to string) WSMessage {
	payload, _ := json.Marshal(TypingSignal{ConversationID: "conv1", State: "start"})
	return WSMessage{Type: "typing", From: from, To: to, Payload: payload}
}

// TestRouteInbound_TypingRelayedToRecipientOnly: a typing signal reaches only its
// recipient — never other connected clients, never the broadcast channel.
func TestRouteInbound_TypingRelayedToRecipientOnly(t *testing.T) {
	h := NewHub()
	b := registerFakeClient(h, "B")
	c := registerFakeClient(h, "C")
	sender := &Client{hub: h, userID: "A", send: make(chan []byte, 8)}

	sender.routeInbound(typingMsg("A", "B"))

	if len(b.send) != 1 {
		t.Fatalf("recipient B should get exactly 1 signal, got %d", len(b.send))
	}
	if len(c.send) != 0 {
		t.Fatalf("uninvolved client C must get nothing, got %d", len(c.send))
	}
	if len(h.broadcast) != 0 {
		t.Fatalf("ephemeral signal must never hit the broadcast channel, got %d", len(h.broadcast))
	}
}

// TestRouteInbound_TypingWithoutRecipientDropped: a typing signal with no `to` is
// dropped, NOT broadcast to everyone (the privacy-critical guard).
func TestRouteInbound_TypingWithoutRecipientDropped(t *testing.T) {
	h := NewHub()
	b := registerFakeClient(h, "B")
	sender := &Client{hub: h, userID: "A", send: make(chan []byte, 8)}

	sender.routeInbound(typingMsg("A", "")) // no recipient

	if len(h.broadcast) != 0 {
		t.Fatalf("typing without recipient must be dropped, not broadcast (got %d)", len(h.broadcast))
	}
	if len(b.send) != 0 {
		t.Fatalf("no client should receive a target-less typing signal, got %d", len(b.send))
	}
}

// TestRouteInbound_EphemeralNotEchoedToSender: a signal addressed to self is dropped.
func TestRouteInbound_EphemeralNotEchoedToSender(t *testing.T) {
	h := NewHub()
	sender := registerFakeClient(h, "A")

	sender.routeInbound(typingMsg("A", "A"))

	if len(sender.send) != 0 {
		t.Fatalf("a self-addressed ephemeral signal must be dropped, got %d", len(sender.send))
	}
}

// TestRouteInbound_ReadReceiptRelayedToRecipientOnly mirrors typing for receipts.
func TestRouteInbound_ReadReceiptRelayedToRecipientOnly(t *testing.T) {
	h := NewHub()
	sender := &Client{hub: h, userID: "B", send: make(chan []byte, 8)}
	a := registerFakeClient(h, "A")
	other := registerFakeClient(h, "C")

	payload, _ := json.Marshal(ReadReceiptSignal{ConversationID: "conv1", MessageIDs: []string{"m1", "m2"}, ReadAt: "2026-05-25T00:00:00Z"})
	sender.routeInbound(WSMessage{Type: "read_receipt", From: "B", To: "A", Payload: payload})

	if len(a.send) != 1 {
		t.Fatalf("sender A should receive the read receipt, got %d", len(a.send))
	}
	if len(other.send) != 0 || len(h.broadcast) != 0 {
		t.Fatalf("read receipt must not leak to others/broadcast")
	}
}

// TestRouteInbound_TextStillBroadcasts confirms non-ephemeral messages keep the
// existing direct/broadcast behavior.
func TestRouteInbound_TextStillBroadcasts(t *testing.T) {
	h := NewHub()
	sender := &Client{hub: h, userID: "A", send: make(chan []byte, 8)}

	sender.routeInbound(WSMessage{Type: "text", From: "A", To: "", Payload: json.RawMessage(`"hi"`)})
	if len(h.broadcast) != 1 {
		t.Fatalf("a broadcast text (no recipient) should hit the broadcast channel, got %d", len(h.broadcast))
	}

	b := registerFakeClient(h, "B")
	sender.routeInbound(WSMessage{Type: "text", From: "A", To: "B", Payload: json.RawMessage(`"hi"`)})
	if len(b.send) != 1 {
		t.Fatalf("a directed text should reach recipient B, got %d", len(b.send))
	}
}
