package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

func TestDeleteAccount_RequiresAuth(t *testing.T) {
	rt := NewRouter([]string{"*"})
	req := httptest.NewRequest(http.MethodDelete, "/v1/users/account", nil)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestDeleteAccount_RevokesRefreshTokens(t *testing.T) {
	rt := NewRouter([]string{"*"})
	ts := rt.TokenService()
	userDID := "did:key:zDeleteMe"
	ts.StoreRefreshToken(userDID, auth.GenerateRefreshToken(), "dev1")

	access, _, err := ts.IssueAccessToken(userDID, "dev1", 1, "messaging")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/users/account", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["deleted"] != true {
		t.Fatalf("deleted = %v", body["deleted"])
	}
	if n := ts.GetActiveRefreshTokenCount(userDID); n != 0 {
		t.Fatalf("expected 0 active refresh tokens, got %d", n)
	}
}

func TestLinkWallet_BindsAddress(t *testing.T) {
	rt := NewRouter([]string{"*"})
	payload, _ := json.Marshal(map[string]string{
		"did":            "did:key:zWalletUser",
		"wallet_address": "DAGtestwallet123",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/identity/link-wallet", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rt.enrollmentWalletMu.Lock()
	addr := rt.enrollmentWalletByDID["did:key:zWalletUser"]
	rt.enrollmentWalletMu.Unlock()
	if addr != "DAGtestwallet123" {
		t.Fatalf("stored address = %q", addr)
	}
}
