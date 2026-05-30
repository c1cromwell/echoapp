package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/passport/recovery"
)

func testPassportRecoveryRouter(t *testing.T) *Router {
	t.Helper()
	rt := testPassportRouter(t)
	rt.PassportRecovery = recovery.NewService(recovery.NewMemStore())
	return rt
}

func TestPassportRecoverySetupInitiateComplete(t *testing.T) {
	rt := testPassportRecoveryRouter(t)

	setupBody, _ := json.Marshal(map[string]interface{}{
		"threshold_m": 2,
		"total_n":     3,
		"shareholders": []map[string]interface{}{
			{"share_index": 1, "guardian_did": "did:key:zDev1", "role": "device", "status": "active"},
			{"share_index": 2, "guardian_did": "did:key:zDev2", "role": "device", "status": "active"},
			{"share_index": 3, "guardian_did": "did:key:zFriend", "role": "contact", "status": "active"},
		},
	})
	setup := httptest.NewRequest(http.MethodPost, "/v1/passport/recovery/setup", bytes.NewReader(setupBody))
	setup.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, setup)
	if w.Code != http.StatusCreated {
		t.Fatalf("setup expected 201, got %d: %s", w.Code, w.Body.String())
	}

	initiate := httptest.NewRequest(http.MethodPost, "/v1/passport/recovery/initiate", nil)
	initiate.Header.Set("Authorization", "Bearer test-token")
	w2 := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w2, initiate)
	if w2.Code != http.StatusOK {
		t.Fatalf("initiate expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var initResp recovery.InitiateResponse
	if err := json.NewDecoder(w2.Body).Decode(&initResp); err != nil {
		t.Fatal(err)
	}

	completeBody, _ := json.Marshal(map[string]string{
		"session_id":          initResp.Session.SessionID,
		"root_key_commitment": recovery.RootKeyCommitment([]byte("passport-root-key-32-bytes!!!!!")),
	})
	complete := httptest.NewRequest(http.MethodPost, "/v1/passport/recovery/complete", bytes.NewReader(completeBody))
	complete.Header.Set("Authorization", "Bearer test-token")
	w3 := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w3, complete)
	if w3.Code != http.StatusOK {
		t.Fatalf("complete expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
}

func TestPassportRecoveryRejectsShareMaterial(t *testing.T) {
	rt := testPassportRecoveryRouter(t)
	body, _ := json.Marshal(map[string]interface{}{
		"threshold_m": 2,
		"total_n":     3,
		"share_bytes": "must-not-accept",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/passport/recovery/setup", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 honeypot rejection, got %d", w.Code)
	}
}
