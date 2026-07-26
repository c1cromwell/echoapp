package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/validation"
)

// Device history sync (WO-CA3) — content-blind, per-device-addressed streams.
//
// All routes are scoped by the caller's controller DID (from the access token):
// a client can only push to / pull from / revoke streams under its own account.
// The server stores opaque ciphertext only; entries are wrapped to the target
// device's key by the writer (pairwise ECDH).
//
//	POST /v3/sync/push    body {target_device_id, ciphertext, entry_type?}  -> {seq}
//	GET  /v3/sync/pull?device_id=&after=&limit=                              -> {entries, next_cursor}
//	GET  /v3/sync/head?device_id=                                           -> {seq}
//	POST /v3/sync/ack     body {device_id, through_seq}                       -> {acked:true}
//	POST /v3/sync/revoke  body {target_device_id}                          -> {revoked:true}

func (h *V3Handlers) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	if !h.enforceDIDRateLimit(w, r, did, "sync_push") {
		return
	}
	var req struct {
		TargetDeviceID string `json:"target_device_id"`
		EntryType      string `json:"entry_type"`
		Ciphertext     []byte `json:"ciphertext"` // JSON base64
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "malformed JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateSyncPush(req.TargetDeviceID, req.Ciphertext); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	seq, err := h.DB.AppendSyncEntry(r.Context(), &database.SyncEntry{
		ControllerDID:  did,
		TargetDeviceID: req.TargetDeviceID,
		EntryType:      req.EntryType,
		Ciphertext:     req.Ciphertext,
		CreatedAt:      time.Now(),
	})
	if errors.Is(err, database.ErrDeviceRevoked) {
		WriteError(w, http.StatusForbidden, "DEVICE_REVOKED", "target device sync stream is revoked", r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "SYNC_PUSH_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"target_device_id": req.TargetDeviceID,
		"seq":              seq,
	})
}

func (h *V3Handlers) handleSyncPull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	deviceID := r.URL.Query().Get("device_id")
	if err := validation.ValidateDeviceID(deviceID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	after := parseInt64(r.URL.Query().Get("after"), 0)
	limit := int(parseInt64(r.URL.Query().Get("limit"), 100))

	entries, err := h.DB.PullSyncEntries(r.Context(), did, deviceID, after, limit)
	if errors.Is(err, database.ErrDeviceRevoked) {
		WriteError(w, http.StatusForbidden, "DEVICE_REVOKED", "this device sync stream is revoked", r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "SYNC_PULL_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if entries == nil {
		entries = []*database.SyncEntry{}
	}
	nextCursor := after
	if n := len(entries); n > 0 {
		nextCursor = entries[n-1].Seq
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"device_id":   deviceID,
		"entries":     entries,
		"next_cursor": nextCursor,
	})
}

func (h *V3Handlers) handleSyncHead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	deviceID := r.URL.Query().Get("device_id")
	if err := validation.ValidateDeviceID(deviceID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	seq, err := h.DB.SyncHead(r.Context(), did, deviceID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "SYNC_HEAD_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"device_id": deviceID,
		"seq":       seq,
	})
}

func (h *V3Handlers) handleSyncAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		DeviceID   string `json:"device_id"`
		ThroughSeq int64  `json:"through_seq"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "malformed JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateDeviceID(req.DeviceID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if req.ThroughSeq < 0 {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "through_seq must be >= 0", r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.DB.AckSyncEntries(r.Context(), did, req.DeviceID, req.ThroughSeq); err != nil {
		WriteError(w, http.StatusInternalServerError, "SYNC_ACK_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"device_id":   req.DeviceID,
		"through_seq": req.ThroughSeq,
		"acked":       true,
	})
}

func (h *V3Handlers) handleSyncRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		TargetDeviceID string `json:"target_device_id"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "malformed JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateDeviceID(req.TargetDeviceID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.DB.RevokeDeviceStream(r.Context(), did, req.TargetDeviceID); err != nil {
		WriteError(w, http.StatusInternalServerError, "SYNC_REVOKE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"target_device_id": req.TargetDeviceID,
		"revoked":          true,
	})
}

func parseInt64(s string, def int64) int64 {
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}
