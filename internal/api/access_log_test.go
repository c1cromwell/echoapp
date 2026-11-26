package api

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedactPath(t *testing.T) {
	for in, want := range map[string]string{
		"/v1/sealed/deliver": "/v1/sealed/deliver",
		"/identity/resolve/did:key:zDnaeibPk6wtLbF24tAb4A32q":      "/identity/resolve/{id}",
		"/v3/messages/6f1c2a34-0000-4000-8000-000000000001/status": "/v3/messages/{id}/status",
		"/v1/anchors/" + strings.Repeat("ab", 32) + "/proof":       "/v1/anchors/{id}/proof",
		"/v1/gamification/quests/identity_builder/claim":           "/v1/gamification/quests/identity_builder/claim",
	} {
		if got := redactPath(in); got != want {
			t.Errorf("redactPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// The access log must never say who called: not the DID in a header or path, not a token, not a
// query string, not an address.
func TestAccessLogNamesNoCaller(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	handler := accessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	req := httptest.NewRequest(http.MethodPost, "/identity/devices/did:key:zSecretPerson?token=abc123secret", nil)
	req.Header.Set("X-Sender-DID", "did:key:zSecretPerson")
	req.Header.Set("Authorization", "Bearer abc123secret")
	req.RemoteAddr = "203.0.113.9:4444"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	line := buf.String()
	if !strings.Contains(line, "POST /identity/devices/{id} 202") {
		t.Fatalf("unexpected line: %q", line)
	}
	for _, secret := range []string{"zSecretPerson", "abc123secret", "203.0.113.9", "token="} {
		if strings.Contains(line, secret) {
			t.Fatalf("the access log contains %q: %s", secret, line)
		}
	}
}
