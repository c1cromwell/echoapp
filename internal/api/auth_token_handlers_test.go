package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

func postRefresh(rt *Router, refreshToken, deviceID string) (*httptest.ResponseRecorder, map[string]interface{}) {
	b, _ := json.Marshal(map[string]string{"refresh_token": refreshToken, "device_id": deviceID})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/refresh", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	var body map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec, body
}

func TestAuthRefresh_RotatesAndDetectsReuse(t *testing.T) {
	rt := NewRouter([]string{"*"})
	ts := rt.TokenService()
	refresh := auth.GenerateRefreshToken()
	ts.StoreRefreshToken("did:key:zUser", refresh, "dev1")

	rec, body := postRefresh(rt, refresh, "dev1")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	newRefresh, _ := body["refresh_token"].(string)
	if newRefresh == "" || newRefresh == refresh {
		t.Fatalf("expected a rotated refresh token, got %q", newRefresh)
	}
	if at, _ := body["access_token"].(string); at == "" {
		t.Fatal("expected an access_token in the response")
	}

	// Reusing the now-consumed refresh token must fail (reuse detection).
	rec2, _ := postRefresh(rt, refresh, "dev1")
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("reused refresh token want 401, got %d", rec2.Code)
	}
}

func TestAuthRefresh_InvalidToken(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rec, _ := postRefresh(rt, "rt_does-not-exist", "dev1")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown refresh token want 401, got %d", rec.Code)
	}
}

func TestAuthRevoke_RequiresAuth(t *testing.T) {
	rt := NewRouter([]string{"*"})
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/revoke", nil)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoke without auth want 401, got %d", rec.Code)
	}
}

func TestAuthRevoke_RevokesUserRefreshTokens(t *testing.T) {
	rt := NewRouter([]string{"*"})
	ts := rt.TokenService()
	access, _, err := ts.IssueAccessToken("did:key:zUser", "dev1", 3, "messaging")
	if err != nil {
		t.Fatal(err)
	}
	ts.StoreRefreshToken("did:key:zUser", auth.GenerateRefreshToken(), "dev1")
	if n := ts.GetActiveRefreshTokenCount("did:key:zUser"); n != 1 {
		t.Fatalf("precondition: want 1 active refresh token, got %d", n)
	}

	req := httptest.NewRequest(http.MethodPost, "/v3/auth/revoke", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if n := ts.GetActiveRefreshTokenCount("did:key:zUser"); n != 0 {
		t.Fatalf("expected all refresh tokens revoked, got %d active", n)
	}
}
