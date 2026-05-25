package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func newUsernameTestRouter() (*Router, *database.MemoryDB) {
	db := database.NewMemoryDB()
	rt := &Router{
		AllowedOrigins: []string{"*"},
		V3:             &V3Handlers{DB: db},
	}
	return rt, db
}

// serve drives a request through the full middleware chain so the tests also
// prove the route is wired into Handler() and exempt from auth.
func serveCheckUsername(rt *Router, method, query string) (*httptest.ResponseRecorder, map[string]interface{}) {
	req := httptest.NewRequest(method, "/v1/users/check-username"+query, nil)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	var body map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec, body
}

func TestCheckUsername_AvailableWhenUnregistered(t *testing.T) {
	rt, _ := newUsernameTestRouter()
	rec, body := serveCheckUsername(rt, http.MethodGet, "?username=alice")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	if body["available"] != true {
		t.Fatalf("expected available=true, got %v", body["available"])
	}
}

func TestCheckUsername_TakenWhenRegistered(t *testing.T) {
	rt, db := newUsernameTestRouter()
	if err := db.CreateUser(context.Background(), &database.User{
		UserID:   "user-alice",
		DID:      "did:key:zAlice",
		Username: "alice",
	}); err != nil {
		t.Fatal(err)
	}
	rec, body := serveCheckUsername(rt, http.MethodGet, "?username=alice")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if body["available"] != false {
		t.Fatalf("expected available=false, got %v", body["available"])
	}
	if body["reason"] != "taken" {
		t.Fatalf("expected reason=taken, got %v", body["reason"])
	}
}

func TestCheckUsername_InvalidFormat(t *testing.T) {
	rt, _ := newUsernameTestRouter()
	// "ab" is too short for ^[a-zA-Z0-9_]{3,30}$
	rec, body := serveCheckUsername(rt, http.MethodGet, "?username=ab")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if body["available"] != false || body["reason"] != "invalid_format" {
		t.Fatalf("expected available=false reason=invalid_format, got %v", body)
	}
}

func TestCheckUsername_RejectsDisallowedChars(t *testing.T) {
	rt, _ := newUsernameTestRouter()
	rec, body := serveCheckUsername(rt, http.MethodGet, "?username=bad%20name")
	if rec.Code != http.StatusOK || body["reason"] != "invalid_format" {
		t.Fatalf("expected 200 invalid_format, got %d %v", rec.Code, body)
	}
}

func TestCheckUsername_MissingParam(t *testing.T) {
	rt, _ := newUsernameTestRouter()
	rec, _ := serveCheckUsername(rt, http.MethodGet, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCheckUsername_MethodNotAllowed(t *testing.T) {
	rt, _ := newUsernameTestRouter()
	rec, _ := serveCheckUsername(rt, http.MethodPost, "?username=alice")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

// TestCheckUsername_PublicNoAuth is the regression guard for the audit finding:
// the endpoint must be reachable with no Authorization / X-Sender-DID header.
func TestCheckUsername_PublicNoAuth(t *testing.T) {
	rt, _ := newUsernameTestRouter()
	rec, _ := serveCheckUsername(rt, http.MethodGet, "?username=alice")
	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("endpoint must be public; got 401")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCheckUsername_StoreUnavailable(t *testing.T) {
	rt := &Router{AllowedOrigins: []string{"*"}} // no V3 / DB configured
	rec, _ := serveCheckUsername(rt, http.MethodGet, "?username=alice")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
