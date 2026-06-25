package botsdk_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/botsdk"
)

func TestSendMessageSuccess(t *testing.T) {
	var got struct {
		RecipientDID   string `json:"recipient_did"`
		ConversationID string `json:"conversation_id"`
		Text           string `json:"text"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/bots/relay/message" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-Bot-DID") != "did:bot" {
			t.Fatalf("bot did header missing")
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_ = json.NewEncoder(w).Encode(map[string]bool{"delivered": true})
	}))
	defer srv.Close()

	bot, err := botsdk.New(botsdk.Config{
		BaseURL:  srv.URL,
		BotDID:   "did:bot",
		APIToken: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := bot.SendMessage(context.Background(), "did:user", botsdk.MessageContent{
		ConversationID: "dm:1",
		Text:           "hi",
	}); err != nil {
		t.Fatal(err)
	}
	if got.RecipientDID != "did:user" || got.Text != "hi" {
		t.Fatalf("unexpected body: %+v", got)
	}
}

func TestSendMessageRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("slow down"))
	}))
	defer srv.Close()

	bot, _ := botsdk.New(botsdk.Config{BaseURL: srv.URL, BotDID: "did:bot", APIToken: "x"})
	err := bot.SendMessage(context.Background(), "did:user", botsdk.MessageContent{
		ConversationID: "dm:1",
		Text:           "hi",
	})
	if !botsdk.IsRateLimit(err) {
		t.Fatalf("expected rate limit, got %v", err)
	}
	if botsdk.RetryAfter(err) == 0 {
		t.Fatal("expected retry after")
	}
}
