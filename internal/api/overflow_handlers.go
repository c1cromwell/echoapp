package api

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

// GET /v3/relay/overflow/{uri} — fetch content-blind overflow blob (WO-237 / M5).
func (h *V3Handlers) handleOverflowBlob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.OverflowStorage == nil {
		WriteError(w, http.StatusServiceUnavailable, "OVERFLOW_UNAVAILABLE", "Overflow storage not configured", r.Header.Get("X-Request-ID"))
		return
	}
	retrievable, ok := h.OverflowStorage.(encblob.RetrievableStorage)
	if !ok {
		WriteError(w, http.StatusServiceUnavailable, "OVERFLOW_UNAVAILABLE", "Overflow storage is not retrievable", r.Header.Get("X-Request-ID"))
		return
	}
	uri := strings.TrimPrefix(r.URL.Path, "/v3/relay/overflow/")
	uri = strings.TrimPrefix(uri, "/")
	if uri == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_URI", "storage uri is required", r.Header.Get("X-Request-ID"))
		return
	}
	blob, err := retrievable.Retrieve(r.Context(), uri)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "overflow blob not found", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"storage_uri":       uri,
		"ciphertext_base64": base64.StdEncoding.EncodeToString(blob),
		"byte_size":         len(blob),
	})
}
