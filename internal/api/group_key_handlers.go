package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/internal/validation"
)

// handleGroupKeyDistribute fans out opaque per-member group key packages over WS.
// POST /v3/groups/key/distribute — content-blind relay (WO-207 / M2).
func (h *V3Handlers) handleGroupKeyDistribute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Groups service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Signals == nil {
		WriteError(w, http.StatusServiceUnavailable, "SIGNALS_UNAVAILABLE", "WebSocket signal hub not configured", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		GroupID       string `json:"group_id"`
		Version       int    `json:"version"`
		DistributedBy string `json:"distributed_by"`
		Packages      []struct {
			To      string `json:"to"`
			Payload []byte `json:"payload"`
		} `json:"packages"`
	}
	if err := h.readJSON(r, &req); err != nil || req.GroupID == "" || req.Version < 1 || len(req.Packages) == 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "group_id, version, and packages are required", r.Header.Get("X-Request-ID"))
		return
	}

	caller := h.getDID(r)
	if caller == "" {
		WriteError(w, http.StatusUnauthorized, "MISSING_AUTH", "authenticated DID required", r.Header.Get("X-Request-ID"))
		return
	}
	if !h.enforceDIDRateLimit(w, r, caller, "group_mutation") {
		return
	}
	if err := validation.ValidateGroupID(req.GroupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	distributor := req.DistributedBy
	if distributor == "" {
		distributor = caller
	}

	ok, err := h.Groups.HasPermission(req.GroupID, distributor, groups.PermissionManageMembers)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "GROUP_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if !ok {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "only admins can distribute group keys", r.Header.Get("X-Request-ID"))
		return
	}

	delivered := 0
	ts := time.Now().UTC().Format(time.RFC3339)
	convID := "group:" + req.GroupID

	for _, pkg := range req.Packages {
		if pkg.To == "" || len(pkg.Payload) == 0 {
			continue
		}
		payload, err := json.Marshal(GroupKeySignal{
			GroupID:       req.GroupID,
			Version:       req.Version,
			EncryptedKey:  pkg.Payload,
			DistributedBy: distributor,
		})
		if err != nil {
			continue
		}
		if h.Signals.PublishSignal(pkg.To, WSMessage{
			Type:           "group_key",
			From:           distributor,
			To:             pkg.To,
			ConversationID: convID,
			Payload:        payload,
			Timestamp:      ts,
		}) {
			delivered++
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"group_id":  req.GroupID,
		"version":   req.Version,
		"delivered": delivered,
		"total":     len(req.Packages),
	})
}
