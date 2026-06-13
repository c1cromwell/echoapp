package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/passport"
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

func backupRouter() *Router {
	store := passport.NewMemSyncStore()
	svc := passport.NewSyncService(store, encblob.NewStubStorage())
	return &Router{V3: &V3Handlers{MessageBackup: svc}}
}

func TestBackupPushPullRoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	backupRouter().V3.RegisterV3Routes(mux)

	body, _ := json.Marshal(map[string]string{
		"ciphertext_base64": base64.RawURLEncoding.EncodeToString([]byte("encrypted-backup")),
	})
	req := httptest.NewRequest(http.MethodPost, "/v3/backup/push", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, withDID(req, "did:alice"))
	if rec.Code != http.StatusOK {
		t.Fatalf("push status %d: %s", rec.Code, rec.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/v3/backup/pull", nil)
	grec := httptest.NewRecorder()
	mux.ServeHTTP(grec, withDID(get, "did:alice"))
	if grec.Code != http.StatusOK {
		t.Fatalf("pull status %d: %s", grec.Code, grec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(grec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ciphertext_base64"] == "" {
		t.Fatal("expected ciphertext_base64 in pull response")
	}
}

func TestBackupPullNotFound(t *testing.T) {
	mux := http.NewServeMux()
	backupRouter().V3.RegisterV3Routes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v3/backup/pull", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, withDID(req, "did:nobody"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
