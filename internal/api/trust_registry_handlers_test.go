package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestTrustRegistryListAndGet(t *testing.T) {
	rt := NewRouter([]string{"*"})

	listReq := httptest.NewRequest(http.MethodGet, "/v1/trust-registry/issuers", nil)
	listRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listResp map[string]interface{}
	if err := json.NewDecoder(listRec.Body).Decode(&listResp); err != nil {
		t.Fatal(err)
	}
	if count, _ := listResp["count"].(float64); count < 3 {
		t.Fatalf("expected seeded issuers, got count=%v", listResp["count"])
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/trust-registry/issuers/gvt_us_dmv", nil)
	getRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get by id status = %d", getRec.Code)
	}

	did := "did:key:z6MkhaXgBZDvotDkL5257faWxcqACaZiarbKhaWc6nMWaveJ"
	didReq := httptest.NewRequest(http.MethodGet, "/v1/trust-registry/issuers/"+did, nil)
	didRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(didRec, didReq)
	if didRec.Code != http.StatusOK {
		t.Fatalf("get by did status = %d, body=%s", didRec.Code, didRec.Body.String())
	}
}

func TestTrustRegistryAdminRegister(t *testing.T) {
	t.Setenv("TRUST_REGISTRY_ADMIN_KEY", "test-admin-key")
	rt := NewRouter([]string{"*"})

	body := []byte(`{"issuer_id":"test_reg","name":"Test","issuer_did":"did:key:z6MkTestRegister","issuer_type":"government","trust_level":"high","trusted_credential_types":["passport"]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/trust-registry/issuers", bytes.NewReader(body))
	req.Header.Set("X-Trust-Registry-Admin-Key", "test-admin-key")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", rec.Code, rec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/trust-registry/issuers/test_reg", nil)
	getRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get registered issuer status = %d", getRec.Code)
	}

	suspendReq := httptest.NewRequest(http.MethodPost, "/v1/trust-registry/issuers/test_reg/suspend", nil)
	suspendReq.Header.Set("X-Trust-Registry-Admin-Key", "test-admin-key")
	suspendRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(suspendRec, suspendReq)
	if suspendRec.Code != http.StatusOK {
		t.Fatalf("suspend status = %d", suspendRec.Code)
	}

	getAfter := httptest.NewRequest(http.MethodGet, "/v1/trust-registry/issuers/test_reg", nil)
	getAfterRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(getAfterRec, getAfter)
	if getAfterRec.Code != http.StatusNotFound {
		t.Fatalf("expected suspended issuer hidden from active lookup, got %d", getAfterRec.Code)
	}
}

func TestTrustRegistryAdminForbiddenWithoutKey(t *testing.T) {
	os.Unsetenv("TRUST_REGISTRY_ADMIN_KEY")
	rt := NewRouter([]string{"*"})
	body := []byte(`{"issuer_id":"x","name":"X","issuer_did":"did:key:z6MkX"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/trust-registry/issuers", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without admin key, got %d", rec.Code)
	}
}
