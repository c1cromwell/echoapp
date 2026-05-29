package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/services/onboarding"
)

// onboardingOrchestrator is the in-process WO-203 session coordinator.
var onboardingOrchestrator = onboarding.NewOnboardingService(nil, nil, nil)

func init() {
	onboardingOrchestrator = onboarding.NewOnboardingService(
		onboarding.NewPhoneVerificationService(),
		onboarding.NewPasskeyService(),
		onboarding.NewRecoveryService(),
	)
}

type onboardingAdvanceRequest struct {
	Action string          `json:"action"`
	Phone  string          `json:"phone,omitempty"`
	OTP    string          `json:"otp,omitempty"`
	Profile onboardingProfilePayload `json:"profile,omitempty"`
}

type onboardingProfilePayload struct {
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Bio         string `json:"bio"`
}

func (rt *Router) handleOnboardingSession(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/v1/onboarding/session/start" && r.Method == http.MethodPost:
		rt.handleOnboardingSessionStart(w, r)
	case strings.HasPrefix(path, "/v1/onboarding/session/") && strings.HasSuffix(path, "/advance") && r.Method == http.MethodPost:
		id := strings.TrimPrefix(path, "/v1/onboarding/session/")
		id = strings.TrimSuffix(id, "/advance")
		id = strings.TrimSuffix(id, "/")
		if id == "" {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown onboarding path", r.Header.Get("X-Request-ID"))
			return
		}
		rt.handleOnboardingSessionAdvance(w, r, id)
	case strings.HasPrefix(path, "/v1/onboarding/session/") && r.Method == http.MethodGet:
		id := strings.TrimPrefix(path, "/v1/onboarding/session/")
		if id == "" || strings.Contains(id, "/") {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown onboarding path", r.Header.Get("X-Request-ID"))
			return
		}
		rt.handleOnboardingSessionGet(w, r, id)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown onboarding path", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) handleOnboardingSessionAdvance(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req onboardingAdvanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	var err error
	switch req.Action {
	case "complete_carousel":
		err = onboardingOrchestrator.CompleteCarousel(sessionID)
	case "skip_carousel":
		err = onboardingOrchestrator.SkipCarousel(sessionID)
	case "start_phone":
		err = onboardingOrchestrator.StartPhoneEntry(sessionID)
	case "submit_phone":
		_, err = onboardingOrchestrator.SubmitPhone(sessionID, req.Phone)
	case "verify_otp":
		err = onboardingOrchestrator.VerifyOTP(sessionID, req.OTP)
	case "skip_passkey":
		err = onboardingOrchestrator.SkipPasskey(sessionID)
	case "skip_recovery":
		err = onboardingOrchestrator.SkipRecovery(sessionID)
	case "skip_profile":
		err = onboardingOrchestrator.SkipProfile(sessionID)
	case "setup_profile":
		err = onboardingOrchestrator.SetupProfile(sessionID, &onboarding.Profile{
			DisplayName: req.Profile.DisplayName,
			Username:    req.Profile.Username,
			Bio:         req.Profile.Bio,
		})
	case "complete":
		_, err = onboardingOrchestrator.CompleteOnboarding(sessionID)
	default:
		WriteError(w, http.StatusBadRequest, "UNKNOWN_ACTION", fmt.Sprintf("unknown action %q", req.Action), r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ONBOARDING_STEP_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	session, err := onboardingOrchestrator.GetSession(sessionID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"session_id":   session.ID,
		"current_step": session.CurrentStep,
		"steps":        session.Steps,
		"profile":      session.Profile,
	})
}

type onboardingStartRequest struct {
	Method string `json:"method"` // "phone" | "verifiable_id"
}

func (rt *Router) handleOnboardingSessionStart(w http.ResponseWriter, r *http.Request) {
	var req onboardingStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	method := onboarding.RegistrationPhone
	if req.Method == "verifiable_id" {
		method = onboarding.RegistrationVerifiableID
	}
	session, err := onboardingOrchestrator.StartSession(method)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "ONBOARDING_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"session_id":          session.ID,
		"registration_method": session.RegistrationMethod,
		"current_step":        session.CurrentStep,
		"steps":               session.Steps,
	})
}

func (rt *Router) handleOnboardingSessionGet(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := onboardingOrchestrator.GetSession(sessionID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"session_id":          session.ID,
		"registration_method": session.RegistrationMethod,
		"current_step":        session.CurrentStep,
		"steps":               session.Steps,
		"profile":             session.Profile,
		"completed_at":        session.CompletedAt,
	})
}
