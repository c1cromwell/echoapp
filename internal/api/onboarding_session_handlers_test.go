package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

func testRouterStartTime() time.Time {
	return time.Now()
}
