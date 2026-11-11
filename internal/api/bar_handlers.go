package api

import (
	"errors"
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
)

func (h *V3Handlers) handleContactsBar(w http.ResponseWriter, r *http.Request) {
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "CONTACTS_UNAVAILABLE", "contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	caller := h.getDID(r)
	if caller == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	switch {
	case r.URL.Path == "/v3/contacts/bar" && r.Method == http.MethodGet:
		h.listBar(w, r, caller)
	case r.URL.Path == "/v3/contacts/bar" && r.Method == http.MethodPost:
		h.allowBar(w, r, caller)
	case r.URL.Path == "/v3/contacts/bar/revoke" && r.Method == http.MethodPost:
		h.revokeBar(w, r, caller)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown bar route", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) listBar(w http.ResponseWriter, r *http.Request, caller string) {
	list, err := h.Contacts.BarContacts(r.Context(), caller)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "BAR_LIST_ERROR", "could not list bar contacts", r.Header.Get("X-Request-ID"))
		return
	}
	if list == nil {
		list = []*database.Contact{}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"contacts": list,
		"count":    len(list),
	})
}

func (h *V3Handlers) allowBar(w http.ResponseWriter, r *http.Request, caller string) {
	var req struct {
		ContactDID string `json:"contactDid"`
	}
	if err := h.readJSON(r, &req); err != nil || req.ContactDID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "contactDid required", r.Header.Get("X-Request-ID"))
		return
	}
	contact, err := h.Contacts.AllowBar(r.Context(), caller, req.ContactDID)
	if err != nil {
		if errors.Is(err, contacts.ErrSelfContact) {
			WriteError(w, http.StatusBadRequest, "SELF_CONTACT", "cannot allow yourself", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusBadRequest, "BAR_ALLOW_ERROR", "could not allow contact", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, contact)
}

func (h *V3Handlers) revokeBar(w http.ResponseWriter, r *http.Request, caller string) {
	var req struct {
		ContactDID string `json:"contactDid"`
	}
	if err := h.readJSON(r, &req); err != nil || req.ContactDID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "contactDid required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Contacts.RevokeBar(r.Context(), caller, req.ContactDID); err != nil {
		WriteError(w, http.StatusNotFound, "BAR_REVOKE_ERROR", "contact is not on the bar", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"revoked": true})
}
