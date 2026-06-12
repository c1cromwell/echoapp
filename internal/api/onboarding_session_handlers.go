package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/database"
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
	Action       string                   `json:"action"`
	Phone        string                   `json:"phone,omitempty"`
	OTP          string                   `json:"otp,omitempty"`
	ChallengeID  string                   `json:"challenge_id,omitempty"`
	CredentialID string                   `json:"credential_id,omitempty"`
	PublicKeyHex string                   `json:"public_key_hex,omitempty"`
	PasskeyType  string                   `json:"passkey_type,omitempty"`
	DeviceInfo   string                   `json:"device_info,omitempty"`
	Profile      onboardingProfilePayload `json:"profile,omitempty"`
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

	var (
		passkeyChallenge *onboarding.PasskeyChallenge
		devOTP           string
		did              string
		err              error
	)

	switch req.Action {
	case "complete_carousel":
		err = onboardingOrchestrator.CompleteCarousel(sessionID)
	case "skip_carousel":
		err = onboardingOrchestrator.SkipCarousel(sessionID)
	case "start_phone":
		err = onboardingOrchestrator.StartPhoneEntry(sessionID)
	case "submit_phone":
		var verification *onboarding.PhoneVerification
		verification, err = onboardingOrchestrator.SubmitPhone(sessionID, req.Phone)
		if err == nil && verification != nil && enrollmentDevModeEnabled() {
			devOTP = verification.Code
		}
	case "verify_otp":
		err = onboardingOrchestrator.VerifyOTP(sessionID, req.OTP)
	case "setup_passkey":
		passkeyChallenge, err = onboardingOrchestrator.SetupPasskey(sessionID)
	case "complete_passkey":
		passkeyType := onboarding.PasskeyFaceID
		switch strings.ToLower(req.PasskeyType) {
		case "touch_id":
			passkeyType = onboarding.PasskeyTouchID
		case "fingerprint":
			passkeyType = onboarding.PasskeyFingerprint
		case "pin":
			passkeyType = onboarding.PasskeyPIN
		}
		var rawPhone string
		if pre, serr := onboardingOrchestrator.GetSession(sessionID); serr == nil {
			rawPhone = pre.PhoneNumber
		}
		did, err = onboardingOrchestrator.CompletePasskeyHex(
			sessionID,
			req.ChallengeID,
			req.CredentialID,
			req.PublicKeyHex,
			passkeyType,
			req.DeviceInfo,
		)
		if err == nil {
			rt.registerOnboardingDID(r.Context(), did, req.PublicKeyHex)
			rt.ensureOnboardingUser(r.Context(), did)
			rt.commitOnboardingOPRFPhone(r.Context(), rawPhone, did)
		}
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
	case "delete_phone":
		err = onboardingOrchestrator.DeletePhoneNumber(sessionID)
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

	resp := rt.buildOnboardingSessionResponse(session)
	if passkeyChallenge != nil {
		resp["passkey_challenge"] = map[string]interface{}{
			"id":         passkeyChallenge.ID,
			"challenge":  base64.StdEncoding.EncodeToString(passkeyChallenge.Challenge),
			"expires_at": passkeyChallenge.ExpiresAt,
		}
	}
	if devOTP != "" {
		resp["dev_otp"] = devOTP
	}
	if did != "" {
		resp["did"] = did
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (rt *Router) ensureOnboardingUser(ctx context.Context, did string) {
	if rt.V3 == nil || rt.V3.DB == nil || did == "" {
		return
	}
	if _, err := rt.V3.DB.GetUserByDID(ctx, did); err == nil {
		return
	}
	optIn := true
	_ = rt.V3.DB.CreateUser(ctx, &database.User{
		DID:                 did,
		TrustTier:           1,
		PhoneDiscoveryOptIn: &optIn,
	})
}

func (rt *Router) registerOnboardingDID(ctx context.Context, did, publicKeyHex string) {
	if rt.DIDRegistry == nil || did == "" || publicKeyHex == "" {
		return
	}
	if _, err := rt.DIDRegistry.Lookup(ctx, did); err == nil {
		_, _ = rt.DIDRegistry.RegisterAdditionalDevice(ctx, did, publicKeyHex, "onboarding")
		return
	}
	_, _, _ = rt.DIDRegistry.Register(ctx, did, publicKeyHex)
}

func (rt *Router) commitOnboardingOPRFPhone(ctx context.Context, e164, did string) {
	if rt.V3 == nil || rt.V3.Contacts == nil || did == "" || e164 == "" {
		return
	}
	key, err := rt.V3.Contacts.DiscoveryKey(e164)
	if err != nil {
		return
	}
	_ = rt.V3.Contacts.CommitDiscoveryKey(ctx, key, did)
}

func (rt *Router) buildOnboardingSessionResponse(session *onboarding.OnboardingSession) map[string]interface{} {
	score, level, _ := onboardingOrchestrator.GetTrustScore(session.ID)
	security, _ := onboardingOrchestrator.GetSecuritySetup(session.ID)
	nudge, _ := onboardingOrchestrator.HasNudgeCard(session.ID)

	out := map[string]interface{}{
		"session_id":           session.ID,
		"registration_method":  session.RegistrationMethod,
		"current_step":         session.CurrentStep,
		"steps":                session.Steps,
		"profile":              session.Profile,
		"did":                  session.DID,
		"phone_verified":       session.PhoneVerified,
		"phone_decoupled":      session.PhoneDecoupled,
		"trust_score":          score,
		"trust_level":          level,
		"security_setup":       security,
		"show_setup_nudge":     nudge,
		"progressive_identity": onboardingOrchestrator.ProgressiveIdentityPrompts(session),
		"completed_at":         session.CompletedAt,
	}
	if session.PhoneHash != "" && !session.PhoneDecoupled {
		out["phone_hash"] = session.PhoneHash
	}
	return out
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
	WriteJSON(w, http.StatusCreated, rt.buildOnboardingSessionResponse(session))
}

func (rt *Router) handleOnboardingSessionGet(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := onboardingOrchestrator.GetSession(sessionID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, rt.buildOnboardingSessionResponse(session))
}
