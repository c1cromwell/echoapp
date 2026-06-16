package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
)

func complyMux(h *ComplyHandlers) http.Handler {
	mux := http.NewServeMux()
	h.RegisterComplyRoutes(mux)
	return mux
}

func complyRequest(method, path, orgDID, token string, body any) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("X-Org-DID", orgDID)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestComplyDashboard_ServiceToken(t *testing.T) {
	t.Setenv("COMPLY_SERVICE_TOKEN", "test-comply-token")
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "test-comply-token")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodGet, "/comply/dashboard", "did:org:acme", "test-comply-token", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["activeRetentionPolicies"].(float64) != 0 {
		t.Fatalf("expected zero policies, got %+v", resp)
	}
}

func TestComplyDashboard_Unauthorized(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "secret")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodGet, "/comply/dashboard", "did:org:acme", "wrong", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestComplyRetentionPolicy_CreateAndList(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	body := map[string]any{
		"policy_type":     "litigation_hold",
		"conversation_id": "c1",
		"scope_label":     "matter-42",
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodPost, "/comply/retention/policy", "did:org:acme", "tok", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create policy want 201, got %d: %s", rec.Code, rec.Body.String())
	}

	lrec := httptest.NewRecorder()
	mux.ServeHTTP(lrec, complyRequest(http.MethodGet, "/comply/retention/policy", "did:org:acme", "tok", nil))
	if lrec.Code != http.StatusOK {
		t.Fatalf("list want 200, got %d", lrec.Code)
	}
	var listed struct {
		Policies []database.RetentionPolicy `json:"policies"`
	}
	_ = json.Unmarshal(lrec.Body.Bytes(), &listed)
	if len(listed.Policies) != 1 || listed.Policies[0].PolicyType != database.PolicyLitigationHold {
		t.Fatalf("expected one litigation_hold policy, got %+v", listed.Policies)
	}
	retained, _ := db.IsConversationRetained(t.Context(), "c1")
	if !retained {
		t.Fatal("policy bind should set conversation retained")
	}
}

func TestComplyConversationRetention_Release(t *testing.T) {
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, "tok")
	mux := complyMux(&ComplyHandlers{Comply: svc})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, complyRequest(http.MethodPost, "/comply/conversations/c9/retention", "did:org:acme", "tok",
		map[string]any{"policy_type": "permanent"}))
	if rec.Code != http.StatusOK {
		t.Fatalf("bind want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rrec := httptest.NewRecorder()
	mux.ServeHTTP(rrec, complyRequest(http.MethodPost, "/comply/conversations/c9/release", "did:org:acme", "tok", nil))
	if rrec.Code != http.StatusOK {
		t.Fatalf("release want 200, got %d", rrec.Code)
	}
	retained, _ := db.IsConversationRetained(t.Context(), "c9")
	if retained {
		t.Fatal("release should clear retention flag")
	}
}

func TestComplyServiceAuth_BypassesGatewayAuth(t *testing.T) {
	t.Setenv("COMPLY_SERVICE_TOKEN", "portal-token")
	db := database.NewMemoryDB()
	svc := comply.NewServiceLegacy(db, db, os.Getenv("COMPLY_SERVICE_TOKEN"))
	rt := NewRouter([]string{"http://localhost:3000"})
	rt.Comply = &ComplyHandlers{Comply: svc}

	req := complyRequest(http.MethodGet, "/comply/dashboard", "did:org:acme", "portal-token", nil)
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("full router comply want 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
