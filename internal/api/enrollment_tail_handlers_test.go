package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnrollmentDID_and_Wallet(t *testing.T) {
	rt := NewRouter([]string{"*"})
	ref := "550e8400-e29b-41d4-a716-446655440000"
	rt.storeEnrollmentVerified(ref, enrollmentVerifiedRecord{
		HolderDID:      "did:key:zHolder",
		AssuranceLevel: "ial2",
		CredentialType: "KYCLite",
		ExpiresAt:      time.Now().Add(5 * time.Minute),
	})

	didBody, _ := json.Marshal(map[string]string{"credential_reference": ref})
	didReq := httptest.NewRequest(http.MethodPost, "/v1/enrollment/did", bytes.NewReader(didBody))
	didReq.Header.Set("Content-Type", "application/json")
	didRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(didRec, didReq)
	if didRec.Code != http.StatusOK {
		t.Fatalf("enrollment/did want 200, got %d: %s", didRec.Code, didRec.Body.String())
	}
	var didResp map[string]interface{}
	_ = json.Unmarshal(didRec.Body.Bytes(), &didResp)
	if didResp["did"] != "did:key:zHolder" {
		t.Fatalf("did = %v", didResp["did"])
	}
	if tier, _ := didResp["trust_tier"].(float64); int(tier) != 4 {
		t.Fatalf("trust_tier = %v, want 4", didResp["trust_tier"])
	}

	walletBody, _ := json.Marshal(map[string]string{"did": "did:key:zHolder"})
	walletReq := httptest.NewRequest(http.MethodPost, "/v1/enrollment/wallet", bytes.NewReader(walletBody))
	walletReq.Header.Set("Content-Type", "application/json")
	walletRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(walletRec, walletReq)
	if walletRec.Code != http.StatusOK {
		t.Fatalf("enrollment/wallet want 200, got %d: %s", walletRec.Code, walletRec.Body.String())
	}
	var walletResp map[string]interface{}
	_ = json.Unmarshal(walletRec.Body.Bytes(), &walletResp)
	addr, _ := walletResp["address"].(string)
	if len(addr) < 4 || addr[:3] != "DAG" {
		t.Fatalf("address = %q", addr)
	}

	// Second wallet call is idempotent.
	walletReq2 := httptest.NewRequest(http.MethodPost, "/v1/enrollment/wallet", bytes.NewReader(walletBody))
	walletReq2.Header.Set("Content-Type", "application/json")
	walletRec2 := httptest.NewRecorder()
	rt.Handler().ServeHTTP(walletRec2, walletReq2)
	var walletResp2 map[string]interface{}
	_ = json.Unmarshal(walletRec2.Body.Bytes(), &walletResp2)
	addr2, _ := walletResp2["address"].(string)
	if addr2 != addr {
		t.Fatalf("expected stable wallet address, got %q vs %q", addr2, addr)
	}
}

func TestEnrollmentDID_registersDeviceKey(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rt.DIDRegistry = NewMemoryDIDRegistry()
	ref := "550e8400-e29b-41d4-a716-446655440001"
	rt.storeEnrollmentVerified(ref, enrollmentVerifiedRecord{
		HolderDID:      "did:key:zHolder",
		AssuranceLevel: "ial2",
		CredentialType: "KYCLite",
		ExpiresAt:      time.Now().Add(5 * time.Minute),
	})

	pubHex := "042a8db0febf8361d5b16c0bd5711625a78d22af9559d0e987666be09ed521459873ec2364e35aa21dbfeb8a63a0b52b61e5c56fbe06fc7ad8cc2143cb1929189a"
	body, _ := json.Marshal(map[string]string{
		"credential_reference": ref,
		"public_key_hex":       pubHex,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/enrollment/did", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("enrollment/did want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	binding, err := rt.DIDRegistry.Lookup(context.Background(), "did:key:zHolder")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if binding.PublicKeyHex != pubHex {
		t.Fatalf("public key = %q", binding.PublicKeyHex)
	}
}

func TestEnrollmentDID_unknownReference(t *testing.T) {
	rt := NewRouter([]string{"*"})
	body, _ := json.Marshal(map[string]string{"credential_reference": "missing"})
	req := httptest.NewRequest(http.MethodPost, "/v1/enrollment/did", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}
