// Package api provides v3 API handlers that connect to backend services.
// These implement the full blueprint API endpoints for:
// - Identity Service (port 8001)
// - Message Relay (port 8002)
// - Trust Service (port 8003)
// - Rewards Service (port 8004)
// - Contacts Service (port 8005)
// - Notification Service (port 8007)
// - Media Service (port 8008)
// - Log Publisher (port 8009)
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/services/bots"
	"github.com/thechadcromwell/echoapp/internal/services/broadcast_channels"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/internal/services/media"
	"github.com/thechadcromwell/echoapp/internal/services/messaging"
	"github.com/thechadcromwell/echoapp/internal/services/notification"
	"github.com/thechadcromwell/echoapp/internal/services/rewards"
	"github.com/thechadcromwell/echoapp/internal/validation"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
	"github.com/thechadcromwell/echoapp/pkg/passport"
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

// V3Handlers holds all service dependencies for v3 API routes.
type V3Handlers struct {
	DB              database.DB
	Contacts        *contacts.Service
	Notification    *notification.Service
	Media           *media.Service
	Rewards         *rewards.Service
	Groups          *groups.GroupService
	Broadcasts      *broadcast_channels.ChannelService
	RateLimiter     *infra.RateLimiter                            // optional; enforces per-DID claim velocity (WO-35)
	IdentityL1      *metagraph.MetagraphClient                    // optional; anchors @username -> DID on the Identity Metagraph (D1)
	Signals         SignalPublisher                               // optional; pushes live typing/receipt/reaction signals over WS (WO-10/192)
	Notifier        OfflineNotifier                               // optional; content-blind push when a signal target is offline (WO-57)
	MessageBackup   *passport.SyncService                         // optional; WO-64/CA2 client-encrypted history backup relay
	OverflowStorage encblob.Storage                               // optional; WO-237 overflow blob retrieval
	Comply          *comply.Service                               // optional; WO-250 retention enforcement
	SealedTokens    *messaging.SealedTokenStore                   // optional; WO-219 sealed-sender tokens
	ConvNotifPrefs  *messaging.ConversationNotificationPrefsStore // optional; WO-56 mute prefs
	Bots            *bots.InstallStore                            // optional; Stage 4 bot installs
	BotTokens       *bots.TokenValidator                          // optional; WO-11 bot API tokens
	BotRateLimiter  *bots.RateLimiter                             // optional; WO-11 bot send velocity
	BotWebhooks     *bots.WebhookRegistry                         // optional; WO-11 inbound webhooks
}

// RegisterV3Routes adds all v3 API routes to the router.
func (h *V3Handlers) RegisterV3Routes(mux *http.ServeMux) {
	// Identity endpoints
	mux.HandleFunc("/v3/auth/register", h.handleAuthRegister)
	mux.HandleFunc("/v3/auth/verify", h.handleAuthVerify)
	mux.HandleFunc("/v3/identity/", h.handleIdentityResolve)

	// Trust endpoints
	mux.HandleFunc("/v3/trust/", h.handleTrustScore)
	mux.HandleFunc("/v3/trust/scores", h.handleTrustScoreBatch)

	// Contacts endpoints
	mux.HandleFunc("/v3/contacts/psi", h.handleContactsPSI)
	mux.HandleFunc("/v3/contacts/discovery-settings", h.handleContactsDiscoverySettings)
	mux.HandleFunc("/v3/contacts/search", h.handleContactsSearch)
	mux.HandleFunc("/v3/contacts/invite", h.handleContactsInvite)
	mux.HandleFunc("/v3/contacts/verify", h.handleContactsVerify)
	mux.HandleFunc("/v3/contacts/list", h.handleContactsList)
	mux.HandleFunc("/v3/contacts/block", h.handleContactsBlock)
	mux.HandleFunc("/v3/contacts/add", h.handleContactsAdd)
	mux.HandleFunc("/v3/contacts/relationship", h.handleContactsRelationship)

	// Rewards endpoints
	mux.HandleFunc("/v3/rewards/claim", h.handleRewardsClaim)
	mux.HandleFunc("/v3/rewards/pending/", h.handleRewardsPending)
	mux.HandleFunc("/v3/rewards/daily-stats", h.handleRewardsDailyStats)
	mux.HandleFunc("/v3/rewards/auto-scale-rate", h.handleRewardsAutoScaleRate)

	// Notification endpoints
	mux.HandleFunc("/v3/notifications/register", h.handleNotificationsRegister)
	mux.HandleFunc("/v3/notifications/send", h.handleNotificationsSend)
	mux.HandleFunc("/v3/notifications/preferences/", h.handleNotificationsPreferences)

	// Media endpoints
	mux.HandleFunc("/v3/media/upload", h.handleMediaUpload)
	mux.HandleFunc("/v3/media/", h.handleMediaGet)

	// Message receipt + ops endpoints (receipt/status/edit/delete/pin/unpin/history)
	mux.HandleFunc("/v3/messages/react", h.handleMessageReact)
	mux.HandleFunc("/v3/messages/reactions", h.handleMessageReactions)
	mux.HandleFunc("/v3/messages/", h.handleMessageSubroute)

	// Conversation-scoped endpoints (pins list, retention flag)
	mux.HandleFunc("/v3/conversations/", h.handleConversationsSubroute)

	// Device history sync (WO-CA3) — content-blind per-device streams
	mux.HandleFunc("/v3/sync/push", h.handleSyncPush)
	mux.HandleFunc("/v3/sync/pull", h.handleSyncPull)
	mux.HandleFunc("/v3/sync/head", h.handleSyncHead)
	mux.HandleFunc("/v3/sync/revoke", h.handleSyncRevoke)

	// Encrypted message backup (WO-64 / WO-CA2) — phrase-encrypted blob relay
	mux.HandleFunc("/v3/backup/push", h.handleBackupPush)
	mux.HandleFunc("/v3/backup/pull", h.handleBackupPull)

	// Offline queue overflow blobs (WO-237 / M5)
	mux.HandleFunc("/v3/relay/overflow/", h.handleOverflowBlob)

	// WebRTC call signaling (M4)
	mux.HandleFunc("/v3/calls/ice-servers", h.handleCallsICEServers)

	// Bots (Stage 4 / WO-11 foundation)
	mux.HandleFunc("/v3/bots/", h.handleBotsSubroute)

	// Group endpoints
	mux.HandleFunc("/v3/groups/key/distribute", h.handleGroupKeyDistribute)
	mux.HandleFunc("/v3/groups/create", h.handleGroupCreate)
	mux.HandleFunc("/v3/groups/members/add", h.handleGroupAddMember)
	mux.HandleFunc("/v3/groups/members/remove", h.handleGroupRemoveMember)
	mux.HandleFunc("/v3/groups/members", h.handleGroupMembers)
	mux.HandleFunc("/v3/groups/", h.handleGroupGet)

	// Broadcast channel endpoints
	mux.HandleFunc("/v3/broadcasts/create", h.handleBroadcastCreate)
	mux.HandleFunc("/v3/broadcasts/post", h.handleBroadcastPost)
	mux.HandleFunc("/v3/broadcasts/subscribe", h.handleBroadcastSubscribe)
	mux.HandleFunc("/v3/broadcasts/unsubscribe", h.handleBroadcastUnsubscribe)
	mux.HandleFunc("/v3/broadcasts/", h.handleBroadcastGet)
}

// --- Helpers ---

func (h *V3Handlers) getDID(r *http.Request) string {
	if did := r.Context().Value(ContextKeyUserID); did != nil {
		return did.(string)
	}
	return ""
}

func (h *V3Handlers) ownerTrustLevel(r *http.Request, ownerDID string) groups.TrustLevel {
	if h.DB != nil {
		user, err := h.DB.GetUserByDID(r.Context(), ownerDID)
		if err == nil {
			return groups.TrustLevelFromTier(user.TrustTier)
		}
	}
	return groups.TrustLevelNewcomer
}

func (h *V3Handlers) readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v)
}

// --- Contacts Handlers ---

func (h *V3Handlers) handleContactsPSI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	did := h.getDID(r)
	if !h.enforceDIDRateLimit(w, r, did, "psi_discovery") {
		return
	}

	// OPRF-PSI (D2): the client sends base64 blinded phone elements; the server
	// evaluates them under its secret key (learning nothing) and returns the
	// evaluations plus the {hex(OPRF_k(phone)) -> DID} index. The client finalizes
	// and matches LOCALLY — it must never send its OPRF outputs back, since the
	// server could brute-force them to phone numbers with its key.
	var req struct {
		Blinded []string `json:"blinded"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	evaluated, err := h.Contacts.OPRFEvaluate(req.Blinded)
	if err != nil {
		switch {
		case errors.Is(err, contacts.ErrRateLimited):
			WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", err.Error(), r.Header.Get("X-Request-ID"))
		case errors.Is(err, contacts.ErrOPRFUnavailable):
			WriteError(w, http.StatusServiceUnavailable, "DISCOVERY_UNAVAILABLE", err.Error(), r.Header.Get("X-Request-ID"))
		default:
			WriteError(w, http.StatusBadRequest, "INVALID_BLINDED", err.Error(), r.Header.Get("X-Request-ID"))
		}
		return
	}

	index, err := h.Contacts.DiscoveryIndex(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DISCOVERY_INDEX_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"evaluated":  evaluated,
		"index":      index,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

func (h *V3Handlers) handleContactsDiscoverySettings(w http.ResponseWriter, r *http.Request) {
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "DISCOVERY_UNAVAILABLE", "Contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "MISSING_DID", "Authenticated DID required", r.Header.Get("X-Request-ID"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		settings, err := h.Contacts.GetPhoneDiscoverySettings(r.Context(), did)
		if err != nil {
			WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, settings)
	case http.MethodPatch, http.MethodPut:
		var req struct {
			OptIn bool `json:"phone_discovery_opt_in"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if err := h.Contacts.SetPhoneDiscoveryOptIn(r.Context(), did, req.OptIn); err != nil {
			WriteError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		settings, err := h.Contacts.GetPhoneDiscoverySettings(r.Context(), did)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "SETTINGS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, settings)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleContactsSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	handle := r.URL.Query().Get("handle")
	if handle == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_PARAM", "handle parameter required", r.Header.Get("X-Request-ID"))
		return
	}

	result, err := h.Contacts.SearchByUsername(r.Context(), h.getDID(r), handle)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "SEARCH_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *V3Handlers) handleContactsInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	invite, err := h.Contacts.CreateInviteLink(r.Context(), h.getDID(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INVITE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusCreated, invite)
}

func (h *V3Handlers) handleContactsVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	invite, err := h.Contacts.AcceptInvite(r.Context(), req.Code, h.getDID(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVITE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, invite)
}

func (h *V3Handlers) handleContactsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	contactsList, err := h.Contacts.GetContacts(r.Context(), h.getDID(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CONTACTS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"contacts":  contactsList,
		"count":     len(contactsList),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *V3Handlers) handleContactsBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		ContactDID string `json:"contactDid"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if err := h.Contacts.BlockContact(r.Context(), h.getDID(r), req.ContactDID); err != nil {
		WriteError(w, http.StatusInternalServerError, "BLOCK_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "blocked"})
}

func (h *V3Handlers) handleContactsAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		ContactDID string `json:"contactDid"`
		AddedVia   string `json:"addedVia"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if req.AddedVia == "" {
		req.AddedVia = "manual"
	}

	contact, err := h.Contacts.AddContact(r.Context(), h.getDID(r), req.ContactDID, req.AddedVia)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ADD_CONTACT_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusCreated, contact)
}

func (h *V3Handlers) handleContactsRelationship(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	caller := h.getDID(r)
	peer := r.URL.Query().Get("peer_did")
	if caller == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	if peer == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_PEER", "peer_did query parameter is required", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Contacts == nil {
		WriteError(w, http.StatusServiceUnavailable, "CONTACTS_UNAVAILABLE", "Contacts service not configured", r.Header.Get("X-Request-ID"))
		return
	}

	mutualGroups := make([]map[string]interface{}, 0)
	if h.Groups != nil {
		for _, g := range h.Groups.MutualGroups(caller, peer) {
			mutualGroups = append(mutualGroups, map[string]interface{}{
				"groupId":      g.GroupID,
				"name":         g.Name,
				"type":         g.Type,
				"member_count": g.MemberCount,
			})
		}
	}

	mutualContacts, err := h.Contacts.MutualContacts(r.Context(), caller, peer, 20)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "RELATIONSHIP_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	contactsPayload := make([]map[string]interface{}, 0, len(mutualContacts))
	for _, c := range mutualContacts {
		entry := map[string]interface{}{"did": c.DID}
		if c.Username != "" {
			entry["username"] = c.Username
		}
		contactsPayload = append(contactsPayload, entry)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"peer_did":              peer,
		"mutual_groups":         mutualGroups,
		"mutual_groups_count":   len(mutualGroups),
		"mutual_contacts":       contactsPayload,
		"mutual_contacts_count": len(contactsPayload),
		"request_id":            r.Header.Get("X-Request-ID"),
	})
}

// --- Rewards Handlers ---

func (h *V3Handlers) handleRewardsClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	did := h.getDID(r)
	reqID := r.Header.Get("X-Request-ID")

	// Anti-gaming: rate-limit reward claims at the HTTP layer (WO-35).
	// The rewards service also runs velocity + duplicate checks internally.
	if h.RateLimiter != nil {
		if err := h.RateLimiter.Check(did, "reward_claim"); err != nil {
			WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "reward claim rate limit exceeded", reqID)
			return
		}
	}

	var req rewards.ClaimRequest
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", reqID)
		return
	}
	req.DID = did

	result, err := h.Rewards.Claim(r.Context(), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "CLAIM_ERROR", err.Error(), reqID)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *V3Handlers) handleRewardsPending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	did := h.getDID(r)
	tierStr := r.URL.Query().Get("tier")
	tier := 1
	if tierStr != "" {
		if t, err := strconv.Atoi(tierStr); err == nil {
			tier = t
		}
	}

	result, err := h.Rewards.GetPending(r.Context(), did, tier)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REWARDS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *V3Handlers) handleRewardsDailyStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	stats, err := h.Rewards.GetDailyStats(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "STATS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, stats)
}

func (h *V3Handlers) handleRewardsAutoScaleRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ratePerMessage": h.Rewards.AutoScaleRate(),
		"decayFactor":    h.Rewards.MessagingDecayFactor(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}

// --- Notification Handlers ---

func (h *V3Handlers) handleNotificationsRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		DeviceLabel string `json:"deviceLabel"`
		PublicKey   string `json:"publicKey"`
		APNsToken   string `json:"apnsToken"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	device, err := h.Notification.RegisterDevice(r.Context(), h.getDID(r), req.DeviceLabel, req.PublicKey, req.APNsToken)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "REGISTER_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusCreated, device)
}

func (h *V3Handlers) handleNotificationsSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		RecipientDID   string `json:"recipientDid"`
		ConversationID string `json:"conversationId"`
		Type           string `json:"type"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	payload := notification.PushPayload{
		Type:           notification.NotificationType(req.Type),
		ConversationID: req.ConversationID,
		SenderDID:      h.getDID(r),
	}

	result, err := h.Notification.Send(r.Context(), req.RecipientDID, payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "NOTIFICATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *V3Handlers) handleNotificationsPreferences(w http.ResponseWriter, r *http.Request) {
	did := h.getDID(r)

	switch r.Method {
	case http.MethodGet:
		prefs, err := h.Notification.GetPreferences(r.Context(), did)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "PREFS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, prefs)

	case http.MethodPut:
		var prefs database.NotificationPrefs
		if err := h.readJSON(r, &prefs); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		prefs.DID = did
		if err := h.Notification.UpdatePreferences(r.Context(), &prefs); err != nil {
			WriteError(w, http.StatusInternalServerError, "PREFS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, prefs)

	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or PUT required", r.Header.Get("X-Request-ID"))
	}
}

// --- Media Handlers ---

func (h *V3Handlers) handleMediaUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	contentType := r.Header.Get("Content-Type")
	sizeStr := r.Header.Get("X-Encrypted-Size")
	tierStr := r.Header.Get("X-Trust-Tier")

	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	tier, _ := strconv.Atoi(tierStr)
	if tier == 0 {
		tier = 1
	}

	req := media.UploadRequest{
		UploaderDID:   h.getDID(r),
		ContentType:   contentType,
		EncryptedSize: size,
		TrustTier:     tier,
	}

	result, err := h.Media.Upload(r.Context(), req, r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "UPLOAD_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusCreated, result)
}

func (h *V3Handlers) handleMediaGet(w http.ResponseWriter, r *http.Request) {
	// /v3/media/{fileId} | /chunks | /chunks/{index} | /scan
	path := strings.TrimPrefix(r.URL.Path, "/v3/media/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		WriteError(w, http.StatusNotFound, "FILE_NOT_FOUND", "missing file id", r.Header.Get("X-Request-ID"))
		return
	}
	fileID := parts[0]
	switch {
	case len(parts) == 1:
		h.handleMediaDownload(w, r, fileID)
	case len(parts) == 2 && parts[1] == "chunks":
		h.handleMediaChunks(w, r, fileID)
	case len(parts) == 3 && parts[1] == "chunks":
		h.handleMediaChunkData(w, r, fileID, parts[2])
	case len(parts) == 2 && parts[1] == "scan":
		h.handleMediaScan(w, r, fileID)
	default:
		WriteError(w, http.StatusNotFound, "FILE_NOT_FOUND", "unknown media path", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleMediaDownload(w http.ResponseWriter, r *http.Request, fileID string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	_, meta, err := h.Media.Download(r.Context(), fileID, h.getDID(r))
	if err != nil {
		WriteError(w, http.StatusNotFound, "FILE_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, meta)
}

func (h *V3Handlers) handleMediaChunks(w http.ResponseWriter, r *http.Request, fileID string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	chunks, err := h.Media.GetChunks(r.Context(), fileID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "FILE_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"fileId": fileID,
		"chunks": chunks,
		"count":  len(chunks),
	})
}

func (h *V3Handlers) handleMediaChunkData(w http.ResponseWriter, r *http.Request, fileID, indexStr string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "Media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_INDEX", "chunk index must be non-negative integer", r.Header.Get("X-Request-ID"))
		return
	}
	data, err := h.Media.RetrieveChunk(r.Context(), fileID, index)
	if err != nil {
		WriteError(w, http.StatusNotFound, "CHUNK_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *V3Handlers) handleMediaScan(w http.ResponseWriter, r *http.Request, fileID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	if err := h.Media.SubmitForScan(r.Context(), fileID); err != nil {
		WriteError(w, http.StatusNotFound, "FILE_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "scan_submitted"})
}

// --- Identity Handlers (stubs connecting to existing DID service) ---

func (h *V3Handlers) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		Username  string `json:"username"`
		PublicKey string `json:"publicKey"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if req.Username == "" || req.PublicKey == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "username and publicKey are required", r.Header.Get("X-Request-ID"))
		return
	}

	pubHex := strings.TrimSpace(req.PublicKey)
	did, derr := didkey.DeriveFromPublicKeyHex(pubHex)
	if derr != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_PUBLIC_KEY", "publicKey must be SEC1 P-256 hex (uncompressed or compressed)", r.Header.Get("X-Request-ID"))
		return
	}

	user := &database.User{
		UserID:   "user-" + req.Username,
		DID:      did,
		Username: req.Username,
	}

	if err := h.DB.CreateUser(r.Context(), user); err != nil {
		WriteError(w, http.StatusConflict, "USER_EXISTS", "Username or DID already registered", r.Header.Get("X-Request-ID"))
		return
	}

	// Anchor the @username -> DID binding on the Identity Metagraph (D1). The
	// Postgres row just created is a read-through cache of this public index;
	// anchoring is best-effort and never blocks or fails registration.
	h.anchorUsername(r.Context(), user.DID, user.Username)

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"userId":   user.UserID,
		"did":      user.DID,
		"username": user.Username,
		"tier":     user.TrustTier,
	})
}

// anchorUsername submits a public username->DID binding to the Identity
// Metagraph. No-op when no Identity L1 client is configured; errors are logged,
// not surfaced, since the Postgres cache remains usable until the anchor lands.
func (h *V3Handlers) anchorUsername(ctx context.Context, did, username string) {
	if h.IdentityL1 == nil {
		return
	}
	_, err := h.IdentityL1.SubmitIdentityL1(ctx, metagraph.UsernameRegistrationUpdate{
		SubjectDID:   did,
		Username:     username,
		RegisteredAt: time.Now().UnixMilli(),
	})
	if err != nil {
		log.Printf("v3: username anchor failed for %q (%s): %v", username, did, err)
	}
}

// handleAuthVerify confirms the caller's identity after authentication.
// By the time this handler runs, the auth middleware (WO-1) has already
// verified the passkey or JWT.  This endpoint exists as a client-callable
// "am I authenticated?" check — it echoes the verified DID back.
func (h *V3Handlers) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	did := h.getDID(r)
	if did == "" {
		// Should never happen — auth middleware would have rejected without a DID —
		// but guard defensively.
		WriteError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"verified":  true,
		"did":       did,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *V3Handlers) handleIdentityResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	did := r.URL.Path[len("/v3/identity/"):]
	if did == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_DID", "DID path parameter required", r.Header.Get("X-Request-ID"))
		return
	}

	user, err := h.DB.GetUserByDID(r.Context(), did)
	if err != nil {
		WriteError(w, http.StatusNotFound, "DID_NOT_FOUND", "DID not found", r.Header.Get("X-Request-ID"))
		return
	}

	ts, _ := h.DB.GetTrustScore(r.Context(), did)
	creds, _ := h.DB.GetCredentialsByDID(r.Context(), did)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":         user.DID,
		"username":    user.Username,
		"trustTier":   user.TrustTier,
		"trustScore":  ts,
		"credentials": creds,
	})
}

// --- Trust Handlers ---

func (h *V3Handlers) handleTrustScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	did := r.URL.Path[len("/v3/trust/"):]
	if did == "" || did == "scores" {
		return // Handled by batch endpoint
	}

	ts, err := h.DB.GetTrustScore(r.Context(), did)
	if err != nil {
		WriteError(w, http.StatusNotFound, "TRUST_NOT_FOUND", "Trust score not found or expired", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, ts)
}

func (h *V3Handlers) handleTrustScoreBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	didsParam := r.URL.Query().Get("dids")
	if didsParam == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_PARAM", "dids query parameter required", r.Header.Get("X-Request-ID"))
		return
	}

	var dids []string
	for _, d := range splitCSV(didsParam) {
		if d != "" {
			dids = append(dids, d)
		}
	}

	scores, err := h.DB.GetTrustScoreBatch(r.Context(), dids)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "TRUST_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"scores": scores,
		"count":  len(scores),
	})
}

// --- Message Receipt Handler ---

// handleMessageReact adds, replaces, or removes (empty emoji) the caller's emoji
// reaction on a message and returns the updated aggregated reactions. The live
// update to the peer travels over the WS "reaction" signal; this endpoint is the
// durable source of truth.
func (h *V3Handlers) handleMessageReact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.DB == nil {
		WriteError(w, http.StatusServiceUnavailable, "REACTIONS_UNAVAILABLE", "reactions not configured", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		MessageID string `json:"message_id"`
		Emoji     string `json:"emoji"`
	}
	if err := h.readJSON(r, &req); err != nil || req.MessageID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "message_id is required", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var err error
	if req.Emoji == "" {
		err = h.DB.RemoveReaction(r.Context(), req.MessageID, did)
	} else {
		err = h.DB.AddReaction(r.Context(), req.MessageID, did, req.Emoji)
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REACTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	rows, err := h.DB.GetReactions(r.Context(), req.MessageID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REACTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	// Server-authoritative live fan-out (WO-10): push the reaction to the
	// counterparty over WS so it lands even if the reacting client never sends a
	// WS signal itself. The durable store above remains the source of truth.
	h.publishReactionSignal(r.Context(), did, req.MessageID, req.Emoji)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": req.MessageID,
		"reactions":  messaging.AggregateReactions(rows),
	})
}

// publishReactionSignal relays a reaction (or its removal, when emoji is empty) to
// the message's counterparty over WS. Best-effort and nil-safe; if the message is
// unknown to the relay (e.g. not queued here) it silently does nothing.
func (h *V3Handlers) publishReactionSignal(ctx context.Context, reactorDID, messageID, emoji string) {
	if h.Signals == nil && h.Notifier == nil {
		return
	}
	meta, err := h.DB.GetMessageMeta(ctx, messageID)
	if err != nil {
		return
	}
	// The counterparty is whichever participant isn't the reactor.
	target := meta.RecipientDID
	if target == reactorDID {
		target = meta.SenderDID
	}
	if target == "" || target == reactorDID {
		return
	}
	payload, err := json.Marshal(ReactionSignal{
		ConversationID: meta.ConversationID,
		MessageID:      messageID,
		Emoji:          emoji,
	})
	if err != nil {
		return
	}
	delivered := false
	if h.Signals != nil {
		delivered = h.Signals.PublishSignal(target, WSMessage{
			Type:           "reaction",
			From:           reactorDID,
			To:             target,
			ConversationID: meta.ConversationID,
			Payload:        payload,
		})
	}
	// If the peer wasn't connected, wake them with a content-blind push (WO-57).
	// A removed reaction (empty emoji) is not worth a push.
	if !delivered && emoji != "" && h.Notifier != nil {
		h.Notifier.NotifyUndelivered(target, reactorDID, meta.ConversationID, false)
	}
}

// handleMessageReactions returns the aggregated reactions for a message.
func (h *V3Handlers) handleMessageReactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.DB == nil {
		WriteError(w, http.StatusServiceUnavailable, "REACTIONS_UNAVAILABLE", "reactions not configured", r.Header.Get("X-Request-ID"))
		return
	}
	messageID := r.URL.Query().Get("message_id")
	if messageID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_MESSAGE_ID", "message_id query parameter is required", r.Header.Get("X-Request-ID"))
		return
	}
	rows, err := h.DB.GetReactions(r.Context(), messageID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REACTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": messageID,
		"reactions":  messaging.AggregateReactions(rows),
	})
}

// handleMessageReceipt is mounted on the "/v3/messages/" prefix and serves two
// routes (the /react and /reactions exact patterns take precedence):
//
//	POST /v3/messages/{id}/receipt   body {"receiptType":"delivered"|"read"}
//	GET  /v3/messages/{id}/status    -> durable delivery state (for reconnect sync)
//
// On a "read" receipt it durably records read state (WO-192) and pushes a live
// read_receipt signal to the original sender (best-effort). The GET lets a client
// reconcile receipts it may have missed while offline (WO-48).
func (h *V3Handlers) handleMessageReceipt(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/v3/messages/"):]
	messageID := path
	action := ""
	if idx := indexByte(path, '/'); idx >= 0 {
		messageID = path[:idx]
		action = path[idx+1:]
	}
	if messageID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_MESSAGE_ID", "message id is required", r.Header.Get("X-Request-ID"))
		return
	}

	// M1 message-ops actions (WO-25/84/59) dispatch ahead of receipt/status.
	switch action {
	case "edit":
		h.handleMessageEdit(w, r, messageID)
		return
	case "delete":
		h.handleMessageDelete(w, r, messageID)
		return
	case "pin":
		h.handleMessagePin(w, r, messageID, true)
		return
	case "unpin":
		h.handleMessagePin(w, r, messageID, false)
		return
	case "history":
		h.handleMessageEditHistory(w, r, messageID)
		return
	case "refs":
		if r.Method == http.MethodGet {
			h.handleMessageRefsGet(w, r, messageID)
		} else if r.Method == http.MethodPost {
			h.handleMessageRefsPut(w, r, messageID)
		} else {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST required", r.Header.Get("X-Request-ID"))
		}
		return
	}

	// GET .../status — return durable delivery state so a reconnecting client syncs.
	if r.Method == http.MethodGet && action == "status" {
		meta, err := h.DB.GetMessageMeta(r.Context(), messageID)
		if err != nil {
			WriteError(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Message not found", r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"messageId":      meta.MessageID,
			"conversationId": meta.ConversationID,
			"status":         meta.Status,
			"deliveredAt":    meta.DeliveredAt,
			"readAt":         meta.ReadAt,
		})
		return
	}

	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		ReceiptType string `json:"receiptType"` // "delivered" or "read"
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	var err error
	read := req.ReceiptType == "read"
	if read {
		err = h.DB.MarkRead(r.Context(), messageID)
	} else {
		req.ReceiptType = "delivered"
		err = h.DB.MarkDelivered(r.Context(), messageID)
	}
	if err != nil {
		WriteError(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Message not found", r.Header.Get("X-Request-ID"))
		return
	}

	// On a read receipt, nudge the original sender live (durable state above is truth).
	if read {
		h.publishReadReceipt(r.Context(), h.getDID(r), messageID)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"messageId":   messageID,
		"receiptType": req.ReceiptType,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// publishReadReceipt sends a live read_receipt signal to the message's sender so
// their UI flips to "Read" without polling. Best-effort and nil-safe.
func (h *V3Handlers) publishReadReceipt(ctx context.Context, readerDID, messageID string) {
	if h.Signals == nil {
		return
	}
	meta, err := h.DB.GetMessageMeta(ctx, messageID)
	if err != nil {
		return
	}
	// Notify the counterparty (the sender), not the reader themselves.
	target := meta.SenderDID
	if target == "" || target == readerDID {
		return
	}
	payload, err := json.Marshal(ReadReceiptSignal{
		ConversationID: meta.ConversationID,
		MessageIDs:     []string{messageID},
		ReadAt:         time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return
	}
	h.Signals.PublishSignal(target, WSMessage{
		Type:           "read_receipt",
		From:           readerDID,
		To:             target,
		ConversationID: meta.ConversationID,
		Payload:        payload,
	})
}

// --- M1 message ops: edit (WO-25), delete (WO-84), pin (WO-59) ---

// messageOpRequest is the shared body for the message-ops endpoints. `Ciphertext`
// (edit only) is opaque and JSON-encoded as base64.
type messageOpRequest struct {
	ConversationID string `json:"conversation_id"`
	Ciphertext     []byte `json:"ciphertext"`
}

// conversationIDForMessage prefers an explicit conversation id, falling back to the
// stored message metadata when available.
func (h *V3Handlers) conversationIDForMessage(ctx context.Context, messageID, provided string) string {
	if provided != "" {
		return provided
	}
	if meta, err := h.DB.GetMessageMeta(ctx, messageID); err == nil {
		return meta.ConversationID
	}
	return ""
}

// publishOpSignal fans a message-op event out to the counterparty over WS, falling
// back to a content-blind push if they are offline. Best-effort; nil-safe.
func (h *V3Handlers) publishOpSignal(ctx context.Context, actorDID, messageID, conversationID, opType string, payload any) {
	if h.Signals == nil && h.Notifier == nil {
		return
	}
	target := ""
	if meta, err := h.DB.GetMessageMeta(ctx, messageID); err == nil {
		target = meta.RecipientDID
		if target == actorDID {
			target = meta.SenderDID
		}
	}
	if target == "" || target == actorDID {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	delivered := false
	if h.Signals != nil {
		delivered = h.Signals.PublishSignal(target, WSMessage{
			Type:           opType,
			From:           actorDID,
			To:             target,
			ConversationID: conversationID,
			Payload:        raw,
		})
	}
	if !delivered && h.Notifier != nil {
		h.Notifier.NotifyUndelivered(target, actorDID, conversationID, false)
	}
}

// publishToPeer sends a conversation-level signal directly to a known peer DID
// (used where there is no message id to resolve participants from). Best-effort.
func (h *V3Handlers) publishToPeer(peerDID, actorDID, conversationID, opType string, payload any) {
	if peerDID == "" || peerDID == actorDID {
		return
	}
	if h.Signals == nil && h.Notifier == nil {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	delivered := false
	if h.Signals != nil {
		delivered = h.Signals.PublishSignal(peerDID, WSMessage{
			Type:           opType,
			From:           actorDID,
			To:             peerDID,
			ConversationID: conversationID,
			Payload:        raw,
		})
	}
	if !delivered && h.Notifier != nil {
		h.Notifier.NotifyUndelivered(peerDID, actorDID, conversationID, false)
	}
}

// handleMessageEdit edits a message (WO-25). Per the hybrid model, an immutable
// version is persisted only when the conversation is under retention; otherwise the
// edit is relayed and clients hold the history. The new ciphertext is fanned out to
// the peer (server stores no plaintext, ever).
func (h *V3Handlers) handleMessageEdit(w http.ResponseWriter, r *http.Request, messageID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req messageOpRequest
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	convID := h.conversationIDForMessage(r.Context(), messageID, req.ConversationID)
	retained, _ := h.DB.IsConversationRetained(r.Context(), convID)
	version := 0
	if retained {
		v, err := h.DB.AppendEditVersion(r.Context(), &database.MessageEdit{
			MessageID:      messageID,
			ConversationID: convID,
			EditorDID:      did,
			Ciphertext:     req.Ciphertext,
		})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "EDIT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		version = v
	}
	h.publishOpSignal(r.Context(), did, messageID, convID, "edit", EditSignal{
		ConversationID: convID,
		MessageID:      messageID,
		Ciphertext:     req.Ciphertext,
		Version:        version,
	})
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": messageID,
		"edited":     true,
		"retained":   retained,
		"version":    version,
	})
}

// handleMessageDelete records a synchronized-delete tombstone (WO-84) and fans the
// delete out to the peer. Under retention (litigation hold) the edit history is
// preserved for eDiscovery; otherwise it is purged with the message.
func (h *V3Handlers) handleMessageDelete(w http.ResponseWriter, r *http.Request, messageID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req messageOpRequest
	_ = h.readJSON(r, &req)
	convID := h.conversationIDForMessage(r.Context(), messageID, req.ConversationID)
	if h.Comply != nil && h.Comply.BlocksDeletion(r.Context(), convID) {
		WriteError(w, http.StatusForbidden, "RETENTION_POLICY_ACTIVE", "retention_policy_active", r.Header.Get("X-Request-ID"))
		return
	}
	retained, _ := h.DB.IsConversationRetained(r.Context(), convID)
	if err := h.DB.MarkMessageDeleted(r.Context(), messageID, retained); err != nil {
		WriteError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	h.publishOpSignal(r.Context(), did, messageID, convID, "delete", DeleteSignal{
		ConversationID: convID,
		MessageID:      messageID,
	})
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": messageID,
		"deleted":    true,
		"retained":   retained,
	})
}

// handleMessagePin pins or unpins a message (WO-59, max 5 per conversation) and
// fans the change out to the peer so their pinned bar updates.
func (h *V3Handlers) handleMessagePin(w http.ResponseWriter, r *http.Request, messageID string, pin bool) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req messageOpRequest
	_ = h.readJSON(r, &req)
	convID := h.conversationIDForMessage(r.Context(), messageID, req.ConversationID)
	if convID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_CONVERSATION", "conversation_id is required", r.Header.Get("X-Request-ID"))
		return
	}
	var err error
	if pin {
		err = h.DB.PinMessage(r.Context(), convID, messageID, did)
	} else {
		err = h.DB.UnpinMessage(r.Context(), convID, messageID)
	}
	if errors.Is(err, database.ErrPinLimitReached) {
		WriteError(w, http.StatusConflict, "PIN_LIMIT_REACHED",
			"conversation already has the maximum number of pinned messages", r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PIN_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	h.publishOpSignal(r.Context(), did, messageID, convID, "pin", PinSignal{
		ConversationID: convID,
		MessageID:      messageID,
		Pinned:         pin,
	})
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": messageID,
		"pinned":     pin,
	})
}

// handleMessageEditHistory returns the immutable edit versions for a message
// (eDiscovery; only populated for retained conversations). Ciphertext is opaque.
func (h *V3Handlers) handleMessageEditHistory(w http.ResponseWriter, r *http.Request, messageID string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	versions, err := h.DB.GetEditHistory(r.Context(), messageID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "HISTORY_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if versions == nil {
		versions = []*database.MessageEdit{}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id": messageID,
		"versions":   versions,
	})
}

// handleMessageRefsPut stores reply/forward thread metadata (WO-59). Preview text
// must stay inside client-encrypted payloads; only message-id refs are persisted.
func (h *V3Handlers) handleMessageRefsPut(w http.ResponseWriter, r *http.Request, messageID string) {
	author := h.getDID(r)
	if author == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	if !h.enforceDIDRateLimit(w, r, author, "message_send") {
		return
	}
	var req struct {
		ConversationID              string `json:"conversation_id"`
		ReplyToMessageID            string `json:"reply_to_message_id"`
		ForwardedFromMessageID      string `json:"forwarded_from_message_id"`
		ForwardedFromConversationID string `json:"forwarded_from_conversation_id"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateMessageRefs(
		messageID,
		req.ConversationID,
		req.ReplyToMessageID,
		req.ForwardedFromMessageID,
		req.ForwardedFromConversationID,
	); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.DB.PutMessageRefs(r.Context(), &database.MessageRefs{
		MessageID:                   messageID,
		ConversationID:              req.ConversationID,
		AuthorDID:                   author,
		ReplyToMessageID:            req.ReplyToMessageID,
		ForwardedFromMessageID:      req.ForwardedFromMessageID,
		ForwardedFromConversationID: req.ForwardedFromConversationID,
		CreatedAt:                   time.Now(),
	}); err != nil {
		WriteError(w, http.StatusInternalServerError, "REFS_PUT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message_id":      messageID,
		"conversation_id": req.ConversationID,
		"stored":          true,
	})
}

func (h *V3Handlers) handleMessageRefsGet(w http.ResponseWriter, r *http.Request, messageID string) {
	if err := validation.ValidateMessageID(messageID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	refs, err := h.DB.GetMessageRefs(r.Context(), messageID)
	if errors.Is(err, database.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "message refs not found", r.Header.Get("X-Request-ID"))
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REFS_GET_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, refs)
}

// handleConversationsSubroute serves conversation-scoped endpoints:
//
//	GET  /v3/conversations/{id}/pins        -> pinned messages
//	POST /v3/conversations/{id}/retention   body {"retained":bool}  (Comply gate)
func (h *V3Handlers) handleConversationsSubroute(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/v3/conversations/"):]
	convID := path
	action := ""
	if idx := indexByte(path, '/'); idx >= 0 {
		convID = path[:idx]
		action = path[idx+1:]
	}
	if convID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_CONVERSATION", "conversation id is required", r.Header.Get("X-Request-ID"))
		return
	}

	switch {
	case action == "notifications" && (r.Method == http.MethodGet || r.Method == http.MethodPut):
		h.handleConversationNotifications(w, r, convID)
	case action == "pins" && r.Method == http.MethodGet:
		pins, err := h.DB.GetPinnedMessages(r.Context(), convID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "PINS_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		if pins == nil {
			pins = []*database.PinnedMessage{}
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"pins":            pins,
		})
	case action == "archive" && r.Method == http.MethodGet:
		archived, err := h.DB.IsConversationArchived(r.Context(), convID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "ARCHIVE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"archived":        archived,
		})
	case action == "archive" && r.Method == http.MethodPost:
		var req struct {
			Archived bool `json:"archived"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if err := h.DB.SetConversationArchived(r.Context(), convID, req.Archived); err != nil {
			WriteError(w, http.StatusInternalServerError, "ARCHIVE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"archived":        req.Archived,
		})
	case action == "retention" && r.Method == http.MethodPost:
		var req struct {
			Retained   bool   `json:"retained"`
			OrgDID     string `json:"org_did"`
			PolicyType string `json:"policy_type"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if req.Retained {
			if h.Comply != nil && req.OrgDID != "" {
				policyType := database.RetentionPolicyType(req.PolicyType)
				if !policyType.Valid() {
					policyType = database.PolicyPermanent
				}
				actor := h.getDID(r)
				if actor == "" {
					actor = req.OrgDID
				}
				if _, err := h.Comply.CreateRetentionPolicy(r.Context(), comply.CreatePolicyInput{
					OrgDID:         req.OrgDID,
					PolicyType:     policyType,
					ConversationID: convID,
					CreatedByDID:   actor,
				}); err != nil {
					WriteError(w, http.StatusInternalServerError, "RETENTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
					return
				}
			} else if err := h.DB.SetConversationRetention(r.Context(), convID, true); err != nil {
				WriteError(w, http.StatusInternalServerError, "RETENTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
		} else {
			if h.Comply != nil {
				_ = h.Comply.ReleaseConversation(r.Context(), convID)
			} else if err := h.DB.SetConversationRetention(r.Context(), convID, false); err != nil {
				WriteError(w, http.StatusInternalServerError, "RETENTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
				return
			}
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"retained":        req.Retained,
		})
	case action == "disappearing" && r.Method == http.MethodGet:
		ttl, err := h.DB.GetDisappearingTTL(r.Context(), convID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "DISAPPEARING_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"ttl_seconds":     ttl,
		})
	case action == "disappearing" && r.Method == http.MethodPost:
		var req struct {
			TTLSeconds int    `json:"ttl_seconds"`
			PeerDID    string `json:"peer_did"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if req.TTLSeconds > 0 && h.Comply != nil && h.Comply.BlocksDisappearing(r.Context(), convID) {
			WriteError(w, http.StatusForbidden, "RETENTION_POLICY_ACTIVE", "retention_policy_active", r.Header.Get("X-Request-ID"))
			return
		}
		if err := h.DB.SetDisappearingTTL(r.Context(), convID, req.TTLSeconds); err != nil {
			WriteError(w, http.StatusInternalServerError, "DISAPPEARING_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		// Fan the timer change to the peer so both ends apply it (peer_did optional).
		h.publishToPeer(req.PeerDID, h.getDID(r), convID, "disappearing_config", DisappearingSignal{
			ConversationID: convID,
			TTLSeconds:     req.TTLSeconds,
		})
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"ttl_seconds":     req.TTLSeconds,
		})
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown conversation route", r.Header.Get("X-Request-ID"))
	}
}

// splitCSV splits a comma-separated string.
func splitCSV(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// --- Group Handlers ---

func (h *V3Handlers) handleGroupCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Groups service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		GroupID      string                          `json:"groupId"`
		GroupType    groups.GroupType                `json:"groupType"`
		Name         string                          `json:"name"`
		Description  string                          `json:"description"`
		Requirements groups.VerificationRequirements `json:"requirements"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	ownerDID := h.getDID(r)
	if !h.enforceDIDRateLimit(w, r, ownerDID, "group_mutation") {
		return
	}
	if err := validation.ValidateGroupCreate(req.GroupID, ownerDID, req.GroupType, req.Name, req.Description); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	profile := groups.GroupProfile{Name: req.Name, Description: req.Description}
	ownerTrust := h.ownerTrustLevel(r, ownerDID)
	group, err := h.Groups.CreateGroup(req.GroupID, ownerDID, req.GroupType, profile, req.Requirements, ownerTrust)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "GROUP_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, group)
}

func (h *V3Handlers) handleGroupGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Groups service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	groupID := r.URL.Path[len("/v3/groups/"):]
	if err := validation.ValidateGroupID(groupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	group, err := h.Groups.GetGroup(groupID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "GROUP_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, group)
}

func (h *V3Handlers) handleGroupAddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Groups service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		GroupID   string `json:"groupId"`
		MemberDID string `json:"memberDid"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	// Only members with manage-members permission (owner/admin) may add members.
	actor := h.getDID(r)
	if !h.enforceDIDRateLimit(w, r, actor, "group_mutation") {
		return
	}
	if err := validation.ValidateGroupID(req.GroupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateGroupMemberDID(req.MemberDID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Groups.AuthorizeAction(req.GroupID, actor, groups.PermissionManageMembers); err != nil {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "only group admins can manage members", r.Header.Get("X-Request-ID"))
		return
	}
	member, err := h.Groups.AddMember(req.GroupID, req.MemberDID, 0, groups.TrustLevelNewcomer, false)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "MEMBER_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"member":         member,
		"requires_rekey": true,
	})
}

func (h *V3Handlers) handleGroupRemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Groups service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		GroupID  string `json:"groupId"`
		MemberID string `json:"memberId"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	actor := h.getDID(r)
	if !h.enforceDIDRateLimit(w, r, actor, "group_mutation") {
		return
	}
	if err := validation.ValidateGroupID(req.GroupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := validation.ValidateGroupMemberDID(req.MemberID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Groups.AuthorizeAction(req.GroupID, actor, groups.PermissionManageMembers); err != nil {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "only group admins can manage members", r.Header.Get("X-Request-ID"))
		return
	}
	if err := h.Groups.RemoveMember(req.GroupID, req.MemberID); err != nil {
		WriteError(w, http.StatusNotFound, "MEMBER_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "removed",
		"requires_rekey": true,
	})
}

func (h *V3Handlers) handleGroupMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Groups == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Groups service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	groupID := r.URL.Query().Get("groupId")
	if err := validation.ValidateGroupID(groupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	members, err := h.Groups.GetGroupMembers(groupID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "GROUP_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"groupId": groupID,
		"members": members,
		"count":   len(members),
	})
}

// --- Broadcast Channel Handlers ---

func (h *V3Handlers) handleBroadcastCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Broadcast service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		Name        string                         `json:"name"`
		Topic       string                         `json:"topic"`
		ChannelType broadcast_channels.ChannelType `json:"channelType"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	creatorDID := h.getDID(r)
	channel, err := h.Broadcasts.CreateChannel(req.Name, req.Topic, creatorDID, req.ChannelType)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "CHANNEL_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, channel)
}

func (h *V3Handlers) handleBroadcastGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Broadcast service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	channelID := r.URL.Path[len("/v3/broadcasts/"):]
	channel, err := h.Broadcasts.GetChannel(channelID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "CHANNEL_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, channel)
}

func (h *V3Handlers) handleBroadcastPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Broadcast service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		ChannelID   string                         `json:"channelId"`
		Content     string                         `json:"content"`
		ContentType broadcast_channels.ContentType `json:"contentType"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	creatorDID := h.getDID(r)
	post, err := h.Broadcasts.CreatePost(req.ChannelID, creatorDID, req.Content, req.ContentType)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "POST_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, post)
}

func (h *V3Handlers) handleBroadcastSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Broadcast service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		ChannelID string `json:"channelId"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	subscriberDID := h.getDID(r)
	sub, err := h.Broadcasts.Subscribe(req.ChannelID, subscriberDID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "SUBSCRIBE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, sub)
}

func (h *V3Handlers) handleBroadcastUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Broadcasts == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Broadcast service not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		ChannelID string `json:"channelId"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	subscriberDID := h.getDID(r)
	if err := h.Broadcasts.Unsubscribe(req.ChannelID, subscriberDID); err != nil {
		WriteError(w, http.StatusNotFound, "UNSUBSCRIBE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "unsubscribed"})
}
