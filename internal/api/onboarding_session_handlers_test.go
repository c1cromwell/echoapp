package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

const onboardingTestPubKeyHex = "042a8db0febf8361d5b16c0bd5711625a78d22af9559d0e987666be09ed521459873ec2364e35aa21dbfeb8a63a0b52b61e5c56fbe06fc7ad8cc2143cb1929189a"

func TestOnboardingSessionStartAndGet(t *testing.T) {
	rt := &Router{StartTime: testRouterStartTime()}

	startBody, _ := json.Marshal(map[string]string{"method": "phone"})
	req := httptest.NewRequest(http.MethodPost, "/v1/onboarding/session/start", bytes.NewReader(startBody))
	rec := httptest.NewRecorder()
	rt.handleOnboardingSession(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("start status = %d body=%s", rec.Code, rec.Body.String())
	}

	var startResp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &startResp); err != nil {
		t.Fatal(err)
	}
	sessionID, _ := startResp["session_id"].(string)
	if sessionID == "" {
		t.Fatal("expected session_id")
	}
	if startResp["trust_score"] != float64(0) {
		t.Fatalf("trust_score = %v", startResp["trust_score"])
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/onboarding/session/"+sessionID, nil)
	getRec := httptest.NewRecorder()
	rt.handleOnboardingSession(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}
}

func TestOnboardingSessionAdvanceCarousel(t *testing.T) {
	rt := &Router{StartTime: testRouterStartTime()}

	startBody, _ := json.Marshal(map[string]string{"method": "phone"})
	startReq := httptest.NewRequest(http.MethodPost, "/v1/onboarding/session/start", bytes.NewReader(startBody))
	startRec := httptest.NewRecorder()
	rt.handleOnboardingSession(startRec, startReq)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start status = %d", startRec.Code)
	}
	var startResp map[string]interface{}
	if err := json.Unmarshal(startRec.Body.Bytes(), &startResp); err != nil {
		t.Fatal(err)
	}
	sessionID := startResp["session_id"].(string)

	advanceBody, _ := json.Marshal(map[string]string{"action": "complete_carousel"})
	advanceReq := httptest.NewRequest(http.MethodPost, "/v1/onboarding/session/"+sessionID+"/advance", bytes.NewReader(advanceBody))
	advanceRec := httptest.NewRecorder()
	rt.handleOnboardingSession(advanceRec, advanceReq)
	if advanceRec.Code != http.StatusOK {
		t.Fatalf("advance status = %d body=%s", advanceRec.Code, advanceRec.Body.String())
	}
	var advanceResp map[string]interface{}
	if err := json.Unmarshal(advanceRec.Body.Bytes(), &advanceResp); err != nil {
		t.Fatal(err)
	}
	if advanceResp["current_step"] != "welcome" {
		t.Fatalf("expected welcome step, got %v", advanceResp["current_step"])
	}
}

func TestOnboardingSessionPhonePasskeyCompleteFlow(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	rt := NewRouter([]string{"*"})
	rt.DIDRegistry = NewMemoryDIDRegistry()
	oprfSvc, err := contacts.NewOPRFService()
	if err != nil {
		t.Fatal(err)
	}
	db := database.NewMemoryDB()
	contactsSvc := contacts.NewService(db)
	contactsSvc.SetOPRF(oprfSvc)
	rt.V3 = &V3Handlers{DB: db, Contacts: contactsSvc}

	sessionID := startOnboardingSession(t, rt)
	advanceOnboarding(t, rt, sessionID, map[string]string{"action": "complete_carousel"})
	advanceOnboarding(t, rt, sessionID, map[string]string{"action": "start_phone"})
	var submitResp map[string]interface{}
	advanceOnboardingDecode(t, rt, sessionID, map[string]string{
		"action": "submit_phone",
		"phone":  "+15551234001",
	}, &submitResp)
	devOTP, _ := submitResp["dev_otp"].(string)
	if devOTP == "" {
		t.Fatal("expected dev_otp in DEV_MODE submit_phone response")
	}
	advanceOnboarding(t, rt, sessionID, map[string]string{
		"action": "verify_otp",
		"otp":    devOTP,
	})

	var passkeyResp map[string]interface{}
	advanceOnboardingDecode(t, rt, sessionID, map[string]string{"action": "setup_passkey"}, &passkeyResp)
	challenge, _ := passkeyResp["passkey_challenge"].(map[string]interface{})
	challengeID, _ := challenge["id"].(string)
	if challengeID == "" {
		t.Fatal("missing passkey challenge")
	}

	var completeResp map[string]interface{}
	advanceOnboardingDecode(t, rt, sessionID, map[string]interface{}{
		"action":         "complete_passkey",
		"challenge_id":   challengeID,
		"credential_id":  "cred-onboard-1",
		"public_key_hex": onboardingTestPubKeyHex,
		"passkey_type":   "face_id",
		"device_info":    "test",
	}, &completeResp)

	did, _ := completeResp["did"].(string)
	if did == "" {
		t.Fatal("missing did")
	}
	if completeResp["trust_score"] != float64(10) {
		t.Fatalf("trust_score after passkey = %v", completeResp["trust_score"])
	}
	prog, _ := completeResp["progressive_identity"].(map[string]interface{})
	if prog["credential_enrollment_available"] != true {
		t.Fatalf("progressive_identity = %v", prog)
	}

	if _, err := rt.DIDRegistry.Lookup(context.Background(), did); err != nil {
		t.Fatalf("did not registered: %v", err)
	}
	key, err := contactsSvc.DiscoveryKey("+15551234001")
	if err != nil {
		t.Fatal(err)
	}
	idx, err := contactsSvc.DiscoveryIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if idx[key] != did {
		t.Fatalf("oprf index %q = %q, want %q", key, idx[key], did)
	}

	advanceOnboarding(t, rt, sessionID, map[string]string{"action": "delete_phone"})
	getRec := httptest.NewRecorder()
	rt.handleOnboardingSession(getRec, httptest.NewRequest(http.MethodGet, "/v1/onboarding/session/"+sessionID, nil))
	var getResp map[string]interface{}
	_ = json.Unmarshal(getRec.Body.Bytes(), &getResp)
	if getResp["phone_decoupled"] != true {
		t.Fatalf("phone_decoupled = %v", getResp["phone_decoupled"])
	}
}

func startOnboardingSession(t *testing.T, rt *Router) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"method": "phone"})
	rec := httptest.NewRecorder()
	rt.handleOnboardingSession(rec, httptest.NewRequest(http.MethodPost, "/v1/onboarding/session/start", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("start: %d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return resp["session_id"].(string)
}

func advanceOnboarding(t *testing.T, rt *Router, sessionID string, payload interface{}) {
	t.Helper()
	var sink map[string]interface{}
	advanceOnboardingDecode(t, rt, sessionID, payload, &sink)
}

func advanceOnboardingDecode(t *testing.T, rt *Router, sessionID string, payload interface{}, out *map[string]interface{}) {
	t.Helper()
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/onboarding/session/"+sessionID+"/advance", bytes.NewReader(body))
	rt.handleOnboardingSession(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("advance %v: %d %s", payload, rec.Code, rec.Body.String())
	}
	if out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatal(err)
		}
	}
}

func testRouterStartTime() time.Time {
	return time.Now()
}
