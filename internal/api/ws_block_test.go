package api

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

type stubBlockChecker struct {
	blocked map[string]bool
}

func (s *stubBlockChecker) IsEitherBlocked(_ context.Context, a, b string) (bool, error) {
	return s.blocked[a+":"+b] || s.blocked[b+":"+a], nil
}

func TestRouteInboundDropsWhenBlocked(t *testing.T) {
	hub := NewHub()
	hub.blockChecker = &stubBlockChecker{blocked: map[string]bool{"did:bob:did:alice": true}}

	recipient := &Client{hub: hub, userID: "did:bob", send: make(chan []byte, 4)}
	hub.clients["did:bob"] = recipient

	sender := &Client{hub: hub, userID: "did:alice"}

	msg := WSMessage{
		Type:           "text",
		To:             "did:bob",
		ConversationID: "conv-1",
		Payload:        json.RawMessage(`{"text":"hi"}`),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
	sender.routeInbound(msg)

	select {
	case <-recipient.send:
		t.Fatal("blocked message should not be delivered")
	default:
	}
}

func TestRouteInboundDeliversWhenNotBlocked(t *testing.T) {
	hub := NewHub()
	db := database.NewMemoryDB()
	hub.blockChecker = contacts.NewService(db)

	ctx := context.Background()
	db.CreateUser(ctx, &database.User{UserID: "u1", DID: "did:alice", Username: "alice"})
	db.CreateUser(ctx, &database.User{UserID: "u2", DID: "did:bob", Username: "bob"})

	recipient := &Client{hub: hub, userID: "did:bob", send: make(chan []byte, 4)}
	hub.clients["did:bob"] = recipient
	sender := &Client{hub: hub, userID: "did:alice"}

	msg := WSMessage{
		Type:           "text",
		To:             "did:bob",
		ConversationID: "conv-1",
		Payload:        json.RawMessage(`{"text":"hi"}`),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
	sender.routeInbound(msg)

	select {
	case <-recipient.send:
	case <-time.After(time.Second):
		t.Fatal("expected delivery when not blocked")
	}
}
