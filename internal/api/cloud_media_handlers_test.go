package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/cloudstorage"
)

func TestCloudOAuthCallbackStub(t *testing.T) {
	t.Setenv("CLOUD_OAUTH_STUB", "true")
	h := &V3Handlers{CloudStorage: cloudstorage.NewService()}
	mux := http.NewServeMux()
	h.WireCloudStorage(mux)
	rec := postJSON(mux, "/v3/integrations/cloud/google_drive/callback", "did:alice", map[string]any{
		"code":         "oauth-code",
		"redirect_uri": "echo://oauth/cloud",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	providers := h.CloudStorage.ListProviders("did:alice")
	if len(providers) != 1 {
		t.Fatalf("providers: %+v", providers)
	}
}

func TestCloudListFilesStub(t *testing.T) {
	t.Setenv("CLOUD_OAUTH_STUB", "true")
	svc := cloudstorage.NewService()
	svc.SaveToken("did:alice", cloudstorage.Token{
		Provider:    cloudstorage.GoogleDrive,
		AccessToken: "tok",
	})
	h := &V3Handlers{CloudStorage: svc}
	mux := http.NewServeMux()
	h.WireCloudStorage(mux)
	req := withDID(httptest.NewRequest(http.MethodGet, "/v3/integrations/cloud/google_drive/files", nil), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "stub-1") {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestCloudStreamStub(t *testing.T) {
	t.Setenv("CLOUD_OAUTH_STUB", "true")
	svc := cloudstorage.NewService()
	svc.SaveToken("did:alice", cloudstorage.Token{
		Provider:    cloudstorage.GoogleDrive,
		AccessToken: "tok",
	})
	h := &V3Handlers{CloudStorage: svc}
	mux := http.NewServeMux()
	h.WireCloudStorage(mux)
	req := withDID(httptest.NewRequest(http.MethodGet, "/v3/integrations/cloud/google_drive/files/stub-1/stream", nil), "did:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "echo-cloud-stub-content") {
		t.Fatalf("body: %s", rec.Body.String())
	}
}
