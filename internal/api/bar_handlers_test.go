package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

func TestBarAllowPasskeyAndRevoke(t *testing.T) {
	db := database.NewMemoryDB()
	h := &V3Handlers{Contacts: contacts.NewService(db)}
	rt := buildTestRouter(&singleKeyRegistry{hexKey: vectorPubKeyHex})
	handler := rt.authMiddleware(http.HandlerFunc(h.handleContactsBar))

	body := []byte(`{"contactDid":"did:key:bob"}`)
	unsigned := httptest.NewRequest(http.MethodPost, "/v3/contacts/bar", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, unsigned)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned: %d %s", rec.Code, rec.Body.String())
	}

	rec = serveBar(t, handler, http.MethodPost, "/v3/contacts/bar", body)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"barAllowed":true`) {
		t.Fatalf("allow: %d %s", rec.Code, rec.Body.String())
	}

	rec = serveBar(t, handler, http.MethodGet, "/v3/contacts/bar", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "did:key:bob") {
		t.Fatalf("list: %d %s", rec.Code, rec.Body.String())
	}

	rec = serveBar(t, handler, http.MethodPost, "/v3/contacts/bar/revoke", body)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"revoked":true`) {
		t.Fatalf("revoke: %d %s", rec.Code, rec.Body.String())
	}
	rec = serveBar(t, handler, http.MethodGet, "/v3/contacts/bar", nil)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "did:key:bob") {
		t.Fatalf("list after revoke: %d %s", rec.Code, rec.Body.String())
	}
}

func serveBar(t *testing.T, handler http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	ts := nowTS()
	payload := body
	if payload == nil {
		payload = []byte{}
	}
	sig := vectorSign(t, method, path, ts, payload)
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set(headerSenderDID, vectorDID)
	req.Header.Set(headerSignature, sig)
	req.Header.Set(headerTimestamp, ts)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
