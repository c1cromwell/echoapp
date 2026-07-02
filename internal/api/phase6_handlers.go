package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/internal/validation"
)

func (h *V3Handlers) handleGroupMuteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	h.handleGroupModerationAction(w, r, "mute")
}

func (h *V3Handlers) handleGroupBanMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	h.handleGroupModerationAction(w, r, "ban")
}

func (h *V3Handlers) handleGroupUnmuteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "groups unavailable", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		GroupID  string `json:"groupId"`
		MemberID string `json:"memberId"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", r.Header.Get("X-Request-ID"))
		return
	}
	actor := h.getDID(r)
	if err := h.Groups.AuthorizeAction(req.GroupID, actor, groups.PermissionManageMembers); err != nil {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "admin required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Groups.UnmuteUser(req.GroupID, req.MemberID); err != nil {
		WriteError(w, http.StatusBadRequest, "MODERATION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "unmuted"})
}

func (h *V3Handlers) handleGroupModerationAction(w http.ResponseWriter, r *http.Request, action string) {
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "groups unavailable", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		GroupID       string `json:"groupId"`
		MemberID      string `json:"memberId"`
		DurationHours int    `json:"duration_hours"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateGroupID(req.GroupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	actor := h.getDID(r)
	if err := h.Groups.AuthorizeAction(req.GroupID, actor, groups.PermissionManageMembers); err != nil {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "admin required", r.Header.Get("X-Request-ID"))
		return
	}
	var err error
	switch action {
	case "mute":
		hours := req.DurationHours
		if hours <= 0 {
			hours = 24
		}
		err = h.Groups.MuteUser(req.GroupID, req.MemberID, time.Duration(hours)*time.Hour)
	case "ban":
		err = h.Groups.BanUser(req.GroupID, req.MemberID)
	default:
		WriteError(w, http.StatusBadRequest, "INVALID_ACTION", action, r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "MODERATION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": action + "d"})
}

func (h *V3Handlers) handleBroadcastList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "broadcasts unavailable", r.Header.Get("X-Request-ID"))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("q")
	var channels interface{}
	if query != "" {
		channels = h.Broadcasts.SearchChannels(query, limit)
	} else {
		channels = h.Broadcasts.ListChannels(limit, offset)
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"channels": channels})
}

func (h *V3Handlers) handleBroadcastPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "broadcasts unavailable", r.Header.Get("X-Request-ID"))
		return
	}
	channelID := r.URL.Query().Get("channelId")
	if channelID == "" {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "channelId required", r.Header.Get("X-Request-ID"))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	posts := h.Broadcasts.GetChannelPosts(channelID, limit, offset)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"channelId": channelID,
		"posts":     posts,
	})
}

// WirePhase6Extensions registers broadcast list/posts and group moderation routes.
func (h *V3Handlers) WirePhase6Extensions(mux *http.ServeMux) {
	mux.HandleFunc("/v3/broadcasts/list", h.handleBroadcastList)
	mux.HandleFunc("/v3/broadcasts/posts", h.handleBroadcastPosts)
	mux.HandleFunc("/v3/groups/members/mute", h.handleGroupMuteMember)
	mux.HandleFunc("/v3/groups/members/ban", h.handleGroupBanMember)
	mux.HandleFunc("/v3/groups/members/unmute", h.handleGroupUnmuteMember)
}
