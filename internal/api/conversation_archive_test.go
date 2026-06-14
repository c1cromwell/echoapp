package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestConversationArchive_SetAndGet(t *testing.T) {
	db := database.NewMemoryDB()
	mux := v3Mux(signalsRouter(db, &fakeSignalPublisher{}))

	rec := postJSON(mux, "/v3/conversations/c1/archive", "did:alice", map[string]any{"archived": true})
	if rec.Code != http.StatusOK {
		t.Fatalf("archive POST want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	grec := httptest.NewRecorder()
	mux.ServeHTTP(grec, httptest.NewRequest(http.MethodGet, "/v3/conversations/c1/archive", nil))
	if grec.Code != http.StatusOK {
		t.Fatalf("archive GET want 200, got %d: %s", grec.Code, grec.Body.String())
	}
	var resp struct {
		ConversationID string `json:"conversation_id"`
		Archived       bool   `json:"archived"`
	}
	_ = json.Unmarshal(grec.Body.Bytes(), &resp)
	if !resp.Archived || resp.ConversationID != "c1" {
		t.Fatalf("expected archived c1, got %+v", resp)
	}
}
