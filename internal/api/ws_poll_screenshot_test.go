package api

import (
	"encoding/json"
	"testing"
)

func pollMsg(from, to string) WSMessage {
	payload, _ := json.Marshal(PollSignal{
		ConversationID: "conv1",
		PollID:         "poll-1",
		Action:         "vote",
		OptionID:       "opt-a",
	})
	return WSMessage{Type: "poll", From: from, To: to, Payload: payload}
}

func screenshotMsg(from, to string) WSMessage {
	payload, _ := json.Marshal(ScreenshotAlertSignal{
		ConversationID: "conv1",
		AlertedAt:      "2026-05-29T12:00:00Z",
	})
	return WSMessage{Type: "screenshot_alert", From: from, To: to, Payload: payload}
}

func TestRouteInbound_PollRelayedToRecipientOnly(t *testing.T) {
	h := NewHub()
	b := registerFakeClient(h, "B")
	c := registerFakeClient(h, "C")
	sender := &Client{hub: h, userID: "A", send: make(chan []byte, 8)}

	sender.routeInbound(pollMsg("A", "B"))

	if len(b.send) != 1 {
		t.Fatalf("recipient B should get exactly 1 poll signal, got %d", len(b.send))
	}
	if len(c.send) != 0 || len(h.broadcast) != 0 {
		t.Fatalf("poll must not leak to others/broadcast")
	}
}

func TestRouteInbound_ScreenshotAlertRelayedToRecipientOnly(t *testing.T) {
	h := NewHub()
	b := registerFakeClient(h, "B")
	c := registerFakeClient(h, "C")
	sender := &Client{hub: h, userID: "A", send: make(chan []byte, 8)}

	sender.routeInbound(screenshotMsg("A", "B"))

	if len(b.send) != 1 {
		t.Fatalf("recipient B should get exactly 1 screenshot alert, got %d", len(b.send))
	}
	if len(c.send) != 0 || len(h.broadcast) != 0 {
		t.Fatalf("screenshot alert must not leak to others/broadcast")
	}
}

func TestRouteInbound_PollWithoutRecipientDropped(t *testing.T) {
	h := NewHub()
	b := registerFakeClient(h, "B")
	sender := &Client{hub: h, userID: "A", send: make(chan []byte, 8)}

	sender.routeInbound(pollMsg("A", ""))

	if len(h.broadcast) != 0 || len(b.send) != 0 {
		t.Fatalf("poll without recipient must be dropped")
	}
}
