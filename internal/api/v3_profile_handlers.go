package api

import (
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

func (h *V3Handlers) handleProfile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v3/profile")
	switch {
	case path == "" || path == "/":
		h.handleProfileOwn(w, r)
	case path == "/privacy":
		h.handleProfilePrivacy(w, r)
	case path == "/view":
		h.handleProfileView(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown profile route", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleProfileOwn(w http.ResponseWriter, r *http.Request) {
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "CONTACTS_UNAVAILABLE", "Contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", r.Header.Get("X-Request-ID"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		profile, err := h.Contacts.GetProfile(r.Context(), did, did)
		if err != nil {
			WriteError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"profile": profile,
			"privacy": h.Contacts.GetPrivacy(r.Context(), did),
		})
	case http.MethodPatch:
		var req struct {
			DisplayName   *string `json:"displayName"`
			Bio           *string `json:"bio"`
			StatusMessage *string `json:"statusMessage"`
			AvatarURL     *string `json:"avatarURL"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		profile, err := h.Contacts.UpdateOwnProfile(r.Context(), did, req.DisplayName, req.Bio, req.StatusMessage, req.AvatarURL)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "PROFILE_UPDATE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, profile)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET and PATCH are allowed", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleProfilePrivacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PATCH is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "CONTACTS_UNAVAILABLE", "Contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", r.Header.Get("X-Request-ID"))
		return
	}

	var settings contacts.PrivacySettings
	if err := h.readJSON(r, &settings); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	updated, err := h.Contacts.UpdatePrivacy(r.Context(), did, settings)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "PRIVACY_UPDATE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, updated)
}

func (h *V3Handlers) handleProfileView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "CONTACTS_UNAVAILABLE", "Contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	viewer := h.getDID(r)
	target := r.URL.Query().Get("did")
	if viewer == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	if target == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_DID", "did query parameter is required", r.Header.Get("X-Request-ID"))
		return
	}
	profile, err := h.Contacts.GetProfile(r.Context(), viewer, target)
	if err != nil {
		WriteError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

func (h *V3Handlers) handleContactsPrivacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only PATCH is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "CONTACTS_UNAVAILABLE", "Contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	owner := h.getDID(r)
	if owner == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		PeerDID  string                          `json:"peerDid"`
		Override contacts.ContactPrivacyOverride `json:"override"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Contacts.SetContactPrivacyOverride(r.Context(), owner, req.PeerDID, req.Override); err != nil {
		WriteError(w, http.StatusBadRequest, "CONTACT_PRIVACY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
