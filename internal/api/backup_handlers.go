package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/passport"
)

// Message history backup (WO-64 / WO-CA2) — client-encrypted blob relay.
//
//	POST /v3/backup/push  body {ciphertext_base64, content_hash?}
//	GET  /v3/backup/pull  -> {ciphertext_base64, storage_uri, content_hash, ...}
//
// Server stores opaque ciphertext only (encblob); decryption uses the user's recovery phrase.

func (h *V3Handlers) handleBackupPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.MessageBackup == nil {
		WriteError(w, http.StatusServiceUnavailable, "BACKUP_UNAVAILABLE", "Message backup not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req passport.PushSyncRequest
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "malformed JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	meta, err := h.MessageBackup.Push(r.Context(), did, req)
	if errors.Is(err, passport.ErrSyncHashMismatch) {
		WriteError(w, http.StatusBadRequest, "HASH_MISMATCH", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BACKUP_PUSH_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, meta)
}

func (h *V3Handlers) handleBackupPull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.MessageBackup == nil {
		WriteError(w, http.StatusServiceUnavailable, "BACKUP_UNAVAILABLE", "Message backup not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	record, err := h.MessageBackup.Pull(r.Context(), did)
	if errors.Is(err, passport.ErrSyncNotFound) {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "No backup for this account", r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "BACKUP_PULL_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"storage_uri":       record.StorageURI,
		"content_hash":      record.ContentHash,
		"byte_size":         record.ByteSize,
		"version":           record.Version,
		"updated_at":        record.UpdatedAt.Format(time.RFC3339),
		"ciphertext_base64": passport.EncodeBase64(record.Ciphertext),
	})
}
