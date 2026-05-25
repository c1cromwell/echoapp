package api

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// usernameRegex matches a well-formed Echo handle: 3–30 chars, letters/digits/underscore.
// Kept in lockstep with internal/services/onboarding.usernameRegex so the availability
// check and registration agree on what a valid handle is.
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// handleCheckUsername answers GET /v1/users/check-username?username=<handle>.
//
// It is public — the iOS onboarding flow (WO-14 UsernameView) polls it for its
// real-time availability checkmark before any session token exists. The check
// runs against the authoritative users store (the same store registration writes
// to), so a "taken" result reflects a real prior registration.
func (rt *Router) handleCheckUsername(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", reqID)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_USERNAME", "username query parameter is required", reqID)
		return
	}

	if !usernameRegex.MatchString(username) {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"username":   username,
			"available":  false,
			"reason":     "invalid_format",
			"request_id": reqID,
		})
		return
	}

	if rt.V3 == nil || rt.V3.DB == nil {
		WriteError(w, http.StatusServiceUnavailable, "USERNAME_STORE_UNAVAILABLE", "username store not configured", reqID)
		return
	}

	_, err := rt.V3.DB.GetUserByUsername(r.Context(), username)
	switch {
	case err == nil:
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"username":   username,
			"available":  false,
			"reason":     "taken",
			"request_id": reqID,
		})
	case errors.Is(err, database.ErrNotFound):
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"username":   username,
			"available":  true,
			"request_id": reqID,
		})
	default:
		WriteError(w, http.StatusInternalServerError, "USERNAME_LOOKUP_FAILED", "could not check username availability", reqID)
	}
}
