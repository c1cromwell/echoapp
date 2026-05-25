package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/infra"
)

// TestPublicRateLimit_ThrottlesByIP verifies S5: a pre-auth public endpoint is
// throttled per client IP once the limit is exceeded.
func TestPublicRateLimit_ThrottlesByIP(t *testing.T) {
	rt := &Router{
		AllowedOrigins: []string{"*"},
		V3:             &V3Handlers{DB: database.NewMemoryDB()},
		PublicRateLimiter: infra.NewRateLimiter(map[string]infra.RateLimitConfig{
			"public_pre_auth": {MaxRequests: 3, Window: time.Minute},
		}),
	}
	handler := rt.Handler()

	serve := func() int {
		req := httptest.NewRequest(http.MethodGet, "/v1/users/check-username?username=alice", nil)
		req.RemoteAddr = "203.0.113.7:5555"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	for i := 0; i < 3; i++ {
		if code := serve(); code != http.StatusOK {
			t.Fatalf("request %d: want 200, got %d", i+1, code)
		}
	}
	if code := serve(); code != http.StatusTooManyRequests {
		t.Fatalf("4th request from same IP: want 429, got %d", code)
	}
}

// TestPublicRateLimit_DisabledWhenNil confirms requests pass when the limiter is
// not configured (the default in tests/dev without the limiter wired).
func TestPublicRateLimit_DisabledWhenNil(t *testing.T) {
	rt := &Router{
		AllowedOrigins: []string{"*"},
		V3:             &V3Handlers{DB: database.NewMemoryDB()},
	}
	handler := rt.Handler()
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/users/check-username?username=bob", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: want 200 (no limiter), got %d", i+1, rec.Code)
		}
	}
}
