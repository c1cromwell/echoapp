package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/credentials"
)

func (rt *Router) handleCredentialVCStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	credID := strings.TrimPrefix(r.URL.Path, "/identity/credentials/status/")
	if credID == "" || strings.Contains(credID, "/") {
		WriteError(w, http.StatusBadRequest, "INVALID_CREDENTIAL_ID", "credential id required", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.CredentialStatusPool == nil {
		WriteError(w, http.StatusServiceUnavailable, "STATUS_STORE_UNAVAILABLE", "Credential status store is not configured", r.Header.Get("X-Request-ID"))
		return
	}
	idx, revoked, found, err := credentials.QueryCredentialVCStatus(r.Context(), rt.CredentialStatusPool, credID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "STATUS_QUERY_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"credential_id":     credID,
		"found":             found,
		"status_list_index": idx,
		"revoked":           revoked,
		"request_id":        r.Header.Get("X-Request-ID"),
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	})
}
