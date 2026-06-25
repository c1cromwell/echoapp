package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/bots"
)

type fakeBotRelayPublisher struct {
	lastTo  string
	lastMsg WSMessage
	ok      bool
}

func (f *fakeBotRelayPublisher) PublishSignal(to string, msg WSMessage) bool {
	f.lastTo = to
	f.lastMsg = msg
	return f.ok
}

func TestBotRelayMessageRequiresInstall(t *testing.T) {
	store := bots.NewInstallStore()
	tokens := bots.NewTokenValidator()
	bot := bots.DefaultCatalog()[0]
	tokens.SetToken(bot.BotDID, "test-token")

	pub := &fakeBotRelayPublisher{ok: true}
	h := &V3Handlers{
		Bots:           store,
		BotTokens:      tokens,
		BotRateLimiter: bots.NewRateLimiter(100, 0),
		Signals:        pub,
	}

	body, _ := json.Marshal(map[string]interface{}{
		"recipient_did":   "did:alice",
		"conversation_id": "dm:alice-bot",
		"text":            "hello from bot",
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/bots/relay/message", bytes.NewReader(body))
	req.Header.Set("X-Bot-DID", bot.BotDID)
	req.Header.Set("X-Bot-Token", "test-token")
	rec := httptest.NewRecorder()
	h.handleBotsSubroute(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without install, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestBotRelayMessageDelivers(t *testing.T) {
	store := bots.NewInstallStore()
	tokens := bots.NewTokenValidator()
	bot := bots.DefaultCatalog()[0]
	tokens.SetToken(bot.BotDID, "test-token")
	store.Install("did:alice", bot.BotDID, []bots.Permission{bots.PermSendMessage, bots.PermReadMessages})

	pub := &fakeBotRelayPublisher{ok: true}
	h := &V3Handlers{
		Bots:           store,
		BotTokens:      tokens,
		BotRateLimiter: bots.NewRateLimiter(100, 0),
		Signals:        pub,
	}

	body, _ := json.Marshal(map[string]interface{}{
		"recipient_did":   "did:alice",
		"conversation_id": "dm:alice-bot",
		"text":            "reminder: stand up",
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/bots/relay/message", bytes.NewReader(body))
	req.Header.Set("X-Bot-DID", bot.BotDID)
	req.Header.Set("X-Bot-Token", "test-token")
	rec := httptest.NewRecorder()
	h.handleBotsSubroute(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if pub.lastTo != "did:alice" {
		t.Fatalf("expected deliver to alice, got %q", pub.lastTo)
	}
	if pub.lastMsg.Type != "text" || pub.lastMsg.From != bot.BotDID {
		t.Fatalf("unexpected ws msg: %+v", pub.lastMsg)
	}
	var delivered struct {
		Delivered bool `json:"delivered"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &delivered); err != nil {
		t.Fatal(err)
	}
	if !delivered.Delivered {
		t.Fatal("expected delivered=true")
	}
}

func TestBotRelayRateLimit(t *testing.T) {
	store := bots.NewInstallStore()
	tokens := bots.NewTokenValidator()
	bot := bots.DefaultCatalog()[0]
	tokens.SetToken(bot.BotDID, "test-token")
	store.Install("did:alice", bot.BotDID, []bots.Permission{bots.PermSendMessage, bots.PermReadMessages})

	pub := &fakeBotRelayPublisher{ok: true}
	h := &V3Handlers{
		Bots:           store,
		BotTokens:      tokens,
		BotRateLimiter: bots.NewRateLimiter(1, 0),
		Signals:        pub,
	}

	body, _ := json.Marshal(map[string]interface{}{
		"recipient_did":   "did:alice",
		"conversation_id": "dm:1",
		"text":            "x",
	})
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v3/bots/relay/message", bytes.NewReader(body))
		req = req.WithContext(context.Background())
		req.Header.Set("X-Bot-DID", bot.BotDID)
		req.Header.Set("X-Bot-Token", "test-token")
		rec := httptest.NewRecorder()
		h.handleBotsSubroute(rec, req)
		if i == 0 && rec.Code != http.StatusOK {
			t.Fatalf("first send: %d", rec.Code)
		}
		if i == 1 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("second send should rate limit, got %d", rec.Code)
		}
	}
}
