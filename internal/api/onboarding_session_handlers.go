package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/services/onboarding"
)

// onboardingOrchestrator is the in-process WO-203 session coordinator (Wave 0.6 MVP).
var onboardingOrchestrator = onboarding.NewOnboardingService(nil, nil, nil)

type onboardingStartRequest struct {
	Method string `json:"method"` // "phone" | "verifiable_id"
}

func (rt *Router) handleOnboardingSession(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/v1/onboarding/session/start" && r.Method == http.MethodPost:
		rt.handleOnboardingSessionStart(w, r)
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
