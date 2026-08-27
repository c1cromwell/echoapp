package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/messaging"
)

func scheduledMux() http.Handler {
	mux := http.NewServeMux()
	h := &V3Handlers{
		Scheduled: messaging.NewScheduledMessageService(messaging.NewMessagingService()),
	}
	h.RegisterV3Routes(mux)
	return mux
}

func TestHandleSchedule_RequiresAuth(t *testing.T) {
	mux := scheduledMux()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/messages/schedule", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestHandleSchedule_CreateListGetCancel(t *testing.T) {
	mux := scheduledMux()
	when := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id": "conv-1",
		"content":         "opaque-ciphertext",
		"content_type":    "text",
		"scheduled_at":    when.Format(time.RFC3339),
		"timezone":        "UTC",
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/messages/schedule", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:alice"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created scheduledMessageDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.ConversationID != "conv-1" {
		t.Fatalf("created = %+v", created)
	}
	if string(created.Content) != "opaque-ciphertext" {
		t.Fatalf("content = %q", created.Content)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v3/messages/schedule", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), ContextKeyUserID, "did:alice"))
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRec.Code)
	}
	var listed struct {
		Messages []scheduledMessageDTO `json:"messages"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Messages) != 1 {
		t.Fatalf("listed %d, want 1", len(listed.Messages))
	}
	if listed.Messages[0].Content != nil {
		t.Fatal("list must omit content")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v3/messages/schedule/"+created.ID, nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), ContextKeyUserID, "did:alice"))
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}

	bobGet := httptest.NewRequest(http.MethodGet, "/v3/messages/schedule/"+created.ID, nil)
	bobGet = bobGet.WithContext(context.WithValue(bobGet.Context(), ContextKeyUserID, "did:bob"))
	bobRec := httptest.NewRecorder()
	mux.ServeHTTP(bobRec, bobGet)
	if bobRec.Code != http.StatusForbidden {
		t.Fatalf("bob get status = %d, want 403", bobRec.Code)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/v3/messages/schedule/"+created.ID, nil)
	delReq = delReq.WithContext(context.WithValue(delReq.Context(), ContextKeyUserID, "did:alice"))
	delRec := httptest.NewRecorder()
	mux.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", delRec.Code, delRec.Body.String())
	}

	list2 := httptest.NewRecorder()
	mux.ServeHTTP(list2, listReq)
	if err := json.Unmarshal(list2.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Messages) != 0 {
		t.Fatalf("after cancel listed %d, want 0", len(listed.Messages))
	}
}

func TestHandleSchedule_PastTimeRejected(t *testing.T) {
	mux := scheduledMux()
	body, _ := json.Marshal(map[string]interface{}{
		"conversation_id": "conv-1",
		"content":         "x",
		"scheduled_at":    time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/messages/schedule", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:alice"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
