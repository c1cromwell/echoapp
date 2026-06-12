package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnrollmentMDL_StartFinish_DevCallback(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	rt := NewRouter([]string{"*"})

	startBody, _ := json.Marshal(map[string]interface{}{
		"transport": "web_dcapi",
		"requested_claims": map[string]bool{
			"givenName": true, "familyName": true, "ageOver18": true,
		},
	})
	startReq := httptest.NewRequest(http.MethodPost, "/v1/enrollment/mdl/start", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("mdl/start want 200, got %d: %s", startRec.Code, startRec.Body.String())
	}
	var startResp map[string]interface{}
	_ = json.Unmarshal(startRec.Body.Bytes(), &startResp)
	sessionID, _ := startResp["id"].(string)
	if sessionID == "" {
		t.Fatal("missing session id")
	}

	devDR := base64.StdEncoding.EncodeToString([]byte("dev-mdl-device-response"))
	devST := base64.StdEncoding.EncodeToString([]byte("dev-mdl-session-transcript"))
	finishBody, _ := json.Marshal(map[string]interface{}{
		"session_id":             sessionID,
		"transport":              "web_dcapi",
		"device_response_b64":    devDR,
		"session_transcript_b64": devST,
	})
	finishReq := httptest.NewRequest(http.MethodPost, "/v1/enrollment/mdl/finish", bytes.NewReader(finishBody))
	finishReq.Header.Set("Content-Type", "application/json")
	finishRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(finishRec, finishReq)
	if finishRec.Code != http.StatusOK {
		t.Fatalf("mdl/finish want 200, got %d: %s", finishRec.Code, finishRec.Body.String())
	}
	var finishResp map[string]interface{}
	_ = json.Unmarshal(finishRec.Body.Bytes(), &finishResp)
	ev, _ := finishResp["evidence"].(map[string]interface{})
	if ev["kind"] != "mdoc" {
		t.Fatalf("evidence kind = %v", ev["kind"])
	}
	if finishResp["assurance_level"] != "ial2" {
		t.Fatalf("ial = %v", finishResp["assurance_level"])
	}
}

func TestEnrollmentIDV_StartAwait_DevStub(t *testing.T) {
	rt := NewRouter([]string{"*"})

	startReq := httptest.NewRequest(http.MethodPost, "/v1/enrollment/idv/start", bytes.NewReader([]byte("{}")))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("idv/start want 200, got %d: %s", startRec.Code, startRec.Body.String())
	}
	var startResp map[string]interface{}
	_ = json.Unmarshal(startRec.Body.Bytes(), &startResp)
	if startResp["provider"] != "dev_stub" {
		t.Fatalf("provider = %v", startResp["provider"])
	}
	sessionID, _ := startResp["id"].(string)

	awaitBody, _ := json.Marshal(map[string]string{"session_id": sessionID})
	awaitReq := httptest.NewRequest(http.MethodPost, "/v1/enrollment/idv/await", bytes.NewReader(awaitBody))
	awaitReq.Header.Set("Content-Type", "application/json")
	awaitRec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(awaitRec, awaitReq)
	if awaitRec.Code != http.StatusOK {
		t.Fatalf("idv/await want 200, got %d: %s", awaitRec.Code, awaitRec.Body.String())
	}
	var awaitResp map[string]interface{}
	_ = json.Unmarshal(awaitRec.Body.Bytes(), &awaitResp)
	ev, _ := awaitResp["evidence"].(map[string]interface{})
	if ev["kind"] != "idv" {
		t.Fatalf("evidence kind = %v", ev["kind"])
	}
}

func TestEnrollmentMDL_FinishRequiresEvidence(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rt.storeEnrollmentMDLSession("sess-1", enrollmentMDLSession{
		ID: "sess-1", Transport: "qr_ble", ExpiresAt: mustFutureTime(),
	})
	body, _ := json.Marshal(map[string]string{"session_id": "sess-1"})
	req := httptest.NewRequest(http.MethodPost, "/v1/enrollment/mdl/finish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func mustFutureTime() (t time.Time) {
	return time.Now().Add(5 * time.Minute)
}
