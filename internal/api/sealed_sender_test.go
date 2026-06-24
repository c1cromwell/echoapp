package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/messaging"
)

func TestSealedTokenIssueAndConsume(t *testing.T) {
	store := messaging.NewSealedTokenStore()
	token, err := store.Issue("did:example:alice")
	if err != nil {
		t.Fatal(err)
	}
	if !store.Consume(token, "did:example:alice") {
		t.Fatal("expected token to consume for matching sender")
	}
	if store.Consume(token, "did:example:alice") {
		t.Fatal("token must be single-use")
	}
}

func TestHandleSealedTokenRequiresAuth(t *testing.T) {
	store := messaging.NewSealedTokenStore()
	h := &V3Handlers{SealedTokens: store, DB: database.NewMemoryDB()}
	req := httptest.NewRequest(http.MethodPost, "/v3/messages/sealed-token", nil)
	rec := httptest.NewRecorder()
	h.handleSealedToken(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestHub_SealedTextStripsSenderOnRelay(t *testing.T) {
	store := messaging.NewSealedTokenStore()
	token, err := store.Issue("did:example:alice")
	if err != nil {
		t.Fatal(err)
	}

	hub := NewHub()
	hub.SetSealedTokenStore(store)

	recipient := &Client{hub: hub, userID: "did:example:bob", send: make(chan []byte, 4)}
	sender := &Client{hub: hub, userID: "did:example:alice", send: make(chan []byte, 4)}
	hub.clients["did:example:bob"] = recipient
	hub.clients["did:example:alice"] = sender

	payload, _ := json.Marshal(map[string]interface{}{
		"delivery_token": token,
		"ciphertext":     []byte("opaque"),
	})
	msg := WSMessage{
		Type:           "sealed_text",
		To:             "did:example:bob",
		ConversationID: "conv-1",
		Payload:        payload,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
	sender.routeInbound(msg)

	select {
	case raw := <-recipient.send:
		var relayed WSMessage
		if err := json.Unmarshal(raw, &relayed); err != nil {
			t.Fatal(err)
		}
		if relayed.From != "" {
			t.Fatalf("relayed from = %q, want empty sealed sender", relayed.From)
		}
		if relayed.Type != "sealed_text" {
			t.Fatalf("type = %q", relayed.Type)
		}
	default:
		t.Fatal("recipient did not receive sealed relay")
	}
}

func TestHandleSealedTokenIssuesToken(t *testing.T) {
	store := messaging.NewSealedTokenStore()
	h := &V3Handlers{SealedTokens: store, DB: database.NewMemoryDB()}
	req := httptest.NewRequest(http.MethodPost, "/v3/messages/sealed-token", nil)
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:example:alice"))
	rec := httptest.NewRecorder()
	h.handleSealedToken(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		DeliveryToken string `json:"delivery_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.DeliveryToken == "" {
		t.Fatal("expected delivery_token")
	}
}
