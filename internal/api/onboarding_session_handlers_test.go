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

func testRouterStartTime() time.Time {
	return time.Now()
}
