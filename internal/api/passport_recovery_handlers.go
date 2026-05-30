package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/pkg/passport/recovery"
)

func (rt *Router) handlePassportRecovery(w http.ResponseWriter, r *http.Request, action string) {
	if rt.PassportRecovery == nil {
		WriteError(w, http.StatusServiceUnavailable, "PASSPORT_RECOVERY_UNAVAILABLE", "Passport recovery not configured", r.Header.Get("X-Request-ID"))
		return
	}
	holderDID, ok := rt.authenticatedHolderDID(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost && action != "status" {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	switch action {
	case "setup":
		var raw map[string]interface{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&raw); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if err := recovery.RejectSecretFields(raw); err != nil {
			WriteError(w, http.StatusBadRequest, "FORBIDDEN_FIELD", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		body, _ := json.Marshal(raw)
		var req recovery.SetupRequest
		if err := json.Unmarshal(body, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid setup body", r.Header.Get("X-Request-ID"))
			return
		}
		policy, shareholders, err := rt.PassportRecovery.Setup(r.Context(), holderDID, req)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "SETUP_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusCreated, map[string]interface{}{
			"policy":       policy,
			"shareholders": shareholders,
		})
	case "initiate":
		resp, err := rt.PassportRecovery.Initiate(r.Context(), holderDID)
		if err != nil {
			if recovery.ErrIsNotConfigured(err) {
				WriteError(w, http.StatusNotFound, "NOT_CONFIGURED", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
			WriteError(w, http.StatusBadRequest, "INITIATE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, resp)
	case "complete":
		var raw map[string]interface{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&raw); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if err := recovery.RejectSecretFields(raw); err != nil {
			WriteError(w, http.StatusBadRequest, "FORBIDDEN_FIELD", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		body, _ := json.Marshal(raw)
		var req recovery.CompleteRequest
		if err := json.Unmarshal(body, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid complete body", r.Header.Get("X-Request-ID"))
			return
		}
		session, err := rt.PassportRecovery.Complete(r.Context(), holderDID, req)
		if err != nil {
			switch {
			case errors.Is(err, recovery.ErrSessionNotFound):
				WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
			case errors.Is(err, recovery.ErrSessionExpired):
				WriteError(w, http.StatusBadRequest, "SESSION_EXPIRED", err.Error(), r.Header.Get("X-Request-ID"))
			default:
				WriteError(w, http.StatusBadRequest, "COMPLETE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			}
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"session": session,
			"message": "Recovery completed; pull credential sync blob on the new device",
		})
	case "status":
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
			return
		}
		policy, shareholders, err := rt.PassportRecovery.GetStatus(r.Context(), holderDID)
		if err != nil {
			if recovery.ErrIsNotConfigured(err) {
				WriteError(w, http.StatusNotFound, "NOT_CONFIGURED", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
			WriteError(w, http.StatusInternalServerError, "STATUS_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"policy":       policy,
			"shareholders": shareholders,
		})
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown recovery path", r.Header.Get("X-Request-ID"))
	}
}

func recoveryActionFromPath(path string) string {
	return strings.TrimPrefix(path, "recovery/")
}
