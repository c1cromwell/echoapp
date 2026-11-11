package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	if int(resp["trust_tier"].(float64)) != 2 {
		t.Errorf("unexpected trust_tier: %v", resp["trust_tier"])
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

func TestVIPVerify_SubscriptionAcceptsCurrentTransaction(t *testing.T) {
	rt := &Router{}
	expires := time.Now().Add(30 * 24 * time.Hour).UnixMilli()
	payload, _ := json.Marshal(map[string]interface{}{
		"transactionId": "2000001",
		"bundleId":      vipBundleID,
		"productId":     vipMonthlyProductID,
		"expiresDate":   expires,
	})
	jws := "e30." + base64.RawURLEncoding.EncodeToString(payload) + ".c2ln"
	body, _ := json.Marshal(map[string]interface{}{
		"did":                "did:key:zVIP",
		"trust_tier":         3,
		"evidence_type":      "vip_subscription",
		"signed_transaction": jws,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/vip-verify", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.handleVIPVerify(w, asCaller(req, "did:key:zVIP"))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d %s", w.Code, w.Body.String())
	}
}
