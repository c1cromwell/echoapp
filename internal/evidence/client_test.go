package evidence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient_SubmitFingerprint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/evidence" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("missing bearer token")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"event_id":     "evt-123",
			"explorer_url": "https://example.com/verify/evt-123",
		})
	}))
	defer srv.Close()

	t.Setenv("DIGITAL_EVIDENCE_ENABLED", "true")
	t.Setenv("DIGITAL_EVIDENCE_API_KEY", "test-key")
	t.Setenv("DIGITAL_EVIDENCE_ORG_ID", "org-1")
	t.Setenv("DIGITAL_EVIDENCE_TENANT_ID", "tenant-1")
	t.Setenv("DIGITAL_EVIDENCE_BASE_URL", srv.URL)

	cfg, err := LoadClientConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewHTTPClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	req, err := SmartCheckmarkFingerprint(cfg, "abc123", "2026-05-29T12:00:00Z", "did:key:z6Mk")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.SubmitFingerprint(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.EventID != "evt-123" {
		t.Fatalf("event id = %q", resp.EventID)
	}
}

func TestLoadClientConfigFromEnv_MissingKey(t *testing.T) {
	t.Setenv("DIGITAL_EVIDENCE_ENABLED", "true")
	t.Setenv("DIGITAL_EVIDENCE_API_KEY", "")
	_, err := LoadClientConfigFromEnv()
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

// An API key alone must not enable the Constellation-hosted client.
func TestLoadClientConfigFromEnv_DisabledByDefault(t *testing.T) {
	t.Setenv("DIGITAL_EVIDENCE_API_KEY", "test-key")
	t.Setenv("DIGITAL_EVIDENCE_ORG_ID", "org-1")
	t.Setenv("DIGITAL_EVIDENCE_TENANT_ID", "tenant-1")
	for _, v := range []string{"", "false", "yes"} {
		t.Setenv("DIGITAL_EVIDENCE_ENABLED", v)
		if _, err := LoadClientConfigFromEnv(); err == nil {
			t.Fatalf("DIGITAL_EVIDENCE_ENABLED=%q: expected the client config to be refused", v)
		}
	}
	t.Setenv("DIGITAL_EVIDENCE_ENABLED", "true")
	t.Setenv("DIGITAL_EVIDENCE_BASE_URL", "")
	cfg, err := LoadClientConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != constellationEvidenceBaseURL || cfg.VerifyBaseURL != constellationVerifyBaseURL {
		t.Fatalf("enabled defaults = %q / %q", cfg.BaseURL, cfg.VerifyBaseURL)
	}
}

// A client built without a verify base must not invent an explorer link.
func TestHTTPClient_NoVerifyBaseNoExplorerLink(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"event_id": "evt-9"})
	}))
	defer srv.Close()
	client, err := NewHTTPClient(&ClientConfig{APIKey: "k", OrganizationID: "o", TenantID: "t", BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.SubmitFingerprint(context.Background(), &FingerprintRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExplorerURL != "" {
		t.Fatalf("explorer url = %q, want empty", resp.ExplorerURL)
	}
}
