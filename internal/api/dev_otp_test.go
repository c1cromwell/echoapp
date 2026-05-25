package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevOTPEnabled_Gating(t *testing.T) {
	cases := []struct {
		dev, allow, env string
		want            bool
	}{
		{"", "", "", false},                   // nothing set
		{"true", "", "", false},               // DEV_MODE alone is insufficient
		{"", "true", "", false},               // ALLOW_DEV_OTP alone is insufficient
		{"true", "true", "", true},            // both set, non-prod
		{"1", "1", "development", true},       // numeric truthy + dev env
		{"true", "true", "production", false}, // hard-disabled in production
	}
	for _, c := range cases {
		t.Setenv("DEV_MODE", c.dev)
		t.Setenv("ALLOW_DEV_OTP", c.allow)
		t.Setenv("ENVIRONMENT", c.env)
		if got := devOTPEnabled(); got != c.want {
			t.Errorf("devOTPEnabled(DEV_MODE=%q ALLOW_DEV_OTP=%q ENVIRONMENT=%q)=%v want %v",
				c.dev, c.allow, c.env, got, c.want)
		}
	}
}

// TestSMSRecovery_NeverSetsDevOTPHeader is the regression guard for S4: the OTP
// must never be returned in a response header, even with dev echo fully enabled.
func TestSMSRecovery_NeverSetsDevOTPHeader(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	t.Setenv("ALLOW_DEV_OTP", "true")

	rt := newSMSTestRouter()
	phone := "+14155552671"
	body, _ := json.Marshal(map[string]string{
		"phone_hash": phoneHash(phone),
		"phone_raw":  phone,
		"did":        "did:key:zTest",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/sms-recovery/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if got := w.Header().Get("X-Dev-OTP"); got != "" {
		t.Fatalf("X-Dev-OTP header must never be set, got %q", got)
	}
}
