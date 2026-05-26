package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/messaging"
)

// v3Mux returns the V3 route mux directly so tests exercise handlers without the
// auth middleware (covered separately); the caller DID is injected via context.
func v3Mux(rt *Router) http.Handler {
	mux := http.NewServeMux()
	rt.V3.RegisterV3Routes(mux)
	return mux
}

func withDID(req *http.Request, did string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
}

func reactionRouter() *Router {
	return &Router{V3: &V3Handlers{Reactions: messaging.NewReactionStore()}}
}

func TestMessageReact_AddAndList(t *testing.T) {
	mux := v3Mux(reactionRouter())

	body, _ := json.Marshal(map[string]string{"message_id": "m1", "emoji": "👍"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("react want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	grec := httptest.NewRecorder()
	mux.ServeHTTP(grec, httptest.NewRequest(http.MethodGet, "/v3/messages/reactions?message_id=m1", nil))
	if grec.Code != http.StatusOK {
		t.Fatalf("list want 200, got %d", grec.Code)
	}
	var resp struct {
		Reactions []messaging.ReactionCount `json:"reactions"`
	}
	_ = json.Unmarshal(grec.Body.Bytes(), &resp)
	if len(resp.Reactions) != 1 || resp.Reactions[0].Emoji != "👍" || resp.Reactions[0].Count != 1 {
		t.Fatalf("expected one 👍 reaction, got %+v", resp.Reactions)
	}
}

func TestMessageReact_EmptyEmojiRemoves(t *testing.T) {
	mux := v3Mux(reactionRouter())
	react := func(emoji string) {
		body, _ := json.Marshal(map[string]string{"message_id": "m1", "emoji": emoji})
		req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
		mux.ServeHTTP(httptest.NewRecorder(), req)
	}
	react("👍")
	react("") // remove

	grec := httptest.NewRecorder()
	mux.ServeHTTP(grec, httptest.NewRequest(http.MethodGet, "/v3/messages/reactions?message_id=m1", nil))
	var resp struct {
		Reactions []messaging.ReactionCount `json:"reactions"`
	}
	_ = json.Unmarshal(grec.Body.Bytes(), &resp)
	if len(resp.Reactions) != 0 {
		t.Fatalf("empty emoji should remove the reaction, got %+v", resp.Reactions)
	}
}

func TestMessageReact_MissingMessageID(t *testing.T) {
	mux := v3Mux(reactionRouter())
	body, _ := json.Marshal(map[string]string{"emoji": "👍"})
	req := withDID(httptest.NewRequest(http.MethodPost, "/v3/messages/react", bytes.NewReader(body)), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing message_id want 400, got %d", rec.Code)
	}
}
