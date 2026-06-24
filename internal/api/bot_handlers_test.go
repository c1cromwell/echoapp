package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/bots"
)

func TestBotsCatalog(t *testing.T) {
	h := &V3Handlers{Bots: bots.NewInstallStore()}
	req := httptest.NewRequest(http.MethodGet, "/v3/bots/catalog", nil)
	rec := httptest.NewRecorder()
	h.handleBotsSubroute(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body struct {
		Bots []bots.Manifest `json:"bots"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Bots) < 2 {
		t.Fatalf("expected catalog bots, got %d", len(body.Bots))
	}
}

func TestBotsInstallAndAuthorize(t *testing.T) {
	store := bots.NewInstallStore()
	h := &V3Handlers{Bots: store}
	catalog := bots.DefaultCatalog()
	bot := catalog[0]

	installBody, _ := json.Marshal(map[string]interface{}{
		"bot_did":     bot.BotDID,
		"permissions": bot.RequiredPermissions,
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/bots/install", bytes.NewReader(installBody))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:alice"))
	rec := httptest.NewRecorder()
	h.handleBotsSubroute(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("install status = %d body=%s", rec.Code, rec.Body.String())
	}

	if err := bots.Authorize(store, "did:alice", bot.BotDID, bots.PermSendMessage); err != nil {
		t.Fatalf("authorize send_message: %v", err)
	}
	if err := bots.Authorize(store, "did:alice", bot.BotDID, bots.PermRequestPayment); err == nil {
		t.Fatal("expected payment permission denied")
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/v3/bots/"+url.PathEscape(bot.BotDID)+"/install", nil)
	delReq = delReq.WithContext(context.WithValue(delReq.Context(), ContextKeyUserID, "did:alice"))
	delRec := httptest.NewRecorder()
	h.handleBotsSubroute(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("uninstall status = %d", delRec.Code)
	}
}
