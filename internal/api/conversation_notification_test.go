package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/messaging"
)

func TestConversationNotificationPrefs_MuteSuppressesPush(t *testing.T) {
	prefs := messaging.NewConversationNotificationPrefsStore()
	hub := NewHub()
	notifier := &fakeNotifier{}
	hub.SetOfflineNotifier(notifier)
	hub.SetConversationNotificationPrefs(prefs)

	prefs.SetMuted("did:bob", "conv-1", true)
	hub.notifyUndelivered("did:bob", "did:alice", "conv-1", false)

	if len(notifier.all()) != 0 {
		t.Fatalf("expected no push for muted conversation, got %d", len(notifier.all()))
	}

	hub.notifyUndelivered("did:bob", "did:alice", "conv-2", false)
	waitFor(t, func() bool { return len(notifier.all()) == 1 })
}

func TestHandleConversationNotificationsPutGet(t *testing.T) {
	prefs := messaging.NewConversationNotificationPrefsStore()
	h := &V3Handlers{ConvNotifPrefs: prefs}

	putReq := httptest.NewRequest(http.MethodPut, "/v3/conversations/conv-1/notifications", strings.NewReader(`{"muted":true}`))
	putReq = putReq.WithContext(context.WithValue(putReq.Context(), ContextKeyUserID, "did:alice"))
	putRec := httptest.NewRecorder()
	h.handleConversationsSubroute(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d", putRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v3/conversations/conv-1/notifications", nil)
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), ContextKeyUserID, "did:alice"))
	getRec := httptest.NewRecorder()
	h.handleConversationsSubroute(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d", getRec.Code)
	}
	if !prefs.IsMuted("did:alice", "conv-1") {
		t.Fatal("expected muted pref stored")
	}
}
