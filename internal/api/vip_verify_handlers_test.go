package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// asCaller marks the request as authenticated for did, as the auth middleware would.
func asCaller(req *http.Request, did string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, did))
}

func TestVIPVerify_HappyPath(t *testing.T) {
	rt := &Router{DIDRegistry: NewMemoryDIDRegistry()}
	body, _ := json.Marshal(map[string]interface{}{
		"did":           "did:key:z6MkTestVIP",
		"trust_tier":    2,
		"evidence_type": "standard_idv",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/vip-verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rt.handleVIPVerify(w, asCaller(req, "did:key:z6MkTestVIP"))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["did"] != "did:key:z6MkTestVIP" {
		t.Errorf("unexpected did: %v", resp["did"])
	}
	// standard_idv is self-asserted here: the server holds no proof of it, so
	// the recorded tier is the baseline, and the response says so.
	if int(resp["trust_tier"].(float64)) != vipMaxSelfAssertedTier {
		t.Errorf("self-asserted evidence must be capped at tier %d, got %v", vipMaxSelfAssertedTier, resp["trust_tier"])
	}
	if int(resp["requested_trust_tier"].(float64)) != 2 || resp["capped"] != true {
		t.Errorf("response should report the cap: %v", resp)
	}
}

func TestVIPVerify_InvalidTier(t *testing.T) {
	rt := &Router{DIDRegistry: NewMemoryDIDRegistry()}
	body, _ := json.Marshal(map[string]interface{}{
		"did":        "did:key:z6MkTest",
		"trust_tier": 99,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/vip-verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rt.handleVIPVerify(w, asCaller(req, "did:key:z6MkTest"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestVIPVerify_MissingDID(t *testing.T) {
	rt := &Router{DIDRegistry: NewMemoryDIDRegistry()}
	body, _ := json.Marshal(map[string]interface{}{"trust_tier": 2})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/vip-verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rt.handleVIPVerify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestVIPVerify_MethodNotAllowed(t *testing.T) {
	rt := &Router{DIDRegistry: NewMemoryDIDRegistry()}
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/vip-verify", nil)
	w := httptest.NewRecorder()

	rt.handleVIPVerify(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestVIPVerify_SubscriptionRequiresStoreKit(t *testing.T) {
	rt := &Router{}
	body, _ := json.Marshal(map[string]interface{}{
		"did":           "did:key:zVIP",
		"trust_tier":    3,
		"evidence_type": "vip_subscription",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/vip-verify", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.handleVIPVerify(w, asCaller(req, "did:key:zVIP"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d %s", w.Code, w.Body.String())
	}
}
