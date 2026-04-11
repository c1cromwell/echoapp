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
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/broadcast_channels"
	"github.com/thechadcromwell/echoapp/internal/services/contacts"
	"github.com/thechadcromwell/echoapp/internal/services/groups"
	"github.com/thechadcromwell/echoapp/internal/services/media"
	"github.com/thechadcromwell/echoapp/internal/services/notification"
	"github.com/thechadcromwell/echoapp/internal/services/rewards"
)

// V3Handlers holds all service dependencies for v3 API routes.
type V3Handlers struct {
	DB           database.DB
	Contacts     *contacts.Service
	Notification *notification.Service
	Media        *media.Service
	Rewards      *rewards.Service
	Groups       *groups.GroupService
	Broadcasts   *broadcast_channels.ChannelService
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
	mux.HandleFunc("/v3/contacts/search", h.handleContactsSearch)
	mux.HandleFunc("/v3/contacts/invite", h.handleContactsInvite)
	mux.HandleFunc("/v3/contacts/verify", h.handleContactsVerify)
	mux.HandleFunc("/v3/contacts/list", h.handleContactsList)
	mux.HandleFunc("/v3/contacts/block", h.handleContactsBlock)
	mux.HandleFunc("/v3/contacts/add", h.handleContactsAdd)

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

	// Message receipt endpoint
	mux.HandleFunc("/v3/messages/", h.handleMessageReceipt)

	// Group endpoints
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

	var req struct {
		PhoneHashes []string `json:"phoneHashes"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	result, err := h.Contacts.PSIDiscovery(r.Context(), h.getDID(r), req.PhoneHashes)
	if err != nil {
		WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, result)
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

// --- Rewards Handlers ---

func (h *V3Handlers) handleRewardsClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req rewards.ClaimRequest
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	req.DID = h.getDID(r)

	result, err := h.Rewards.Claim(r.Context(), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "CLAIM_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
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
	// Extract fileId from path: /v3/media/{fileId} or /v3/media/{fileId}/chunks or /v3/media/{fileId}/scan
	path := r.URL.Path[len("/v3/media/"):]

	switch {
	case len(path) > 7 && path[len(path)-7:] == "/chunks":
		fileID := path[:len(path)-7]
		h.handleMediaChunks(w, r, fileID)
	case len(path) > 5 && path[len(path)-5:] == "/scan":
		fileID := path[:len(path)-5]
		h.handleMediaScan(w, r, fileID)
	default:
		h.handleMediaDownload(w, r, path)
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

	// Create user record
	user := &database.User{
		UserID:   "user-" + req.Username,
		DID:      "did:prism:cardano:" + req.Username,
		Username: req.Username,
	}

	if err := h.DB.CreateUser(r.Context(), user); err != nil {
		WriteError(w, http.StatusConflict, "USER_EXISTS", "Username or DID already registered", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"userId":   user.UserID,
		"did":      user.DID,
		"username": user.Username,
		"tier":     user.TrustTier,
	})
}

func (h *V3Handlers) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	// Passkey verification placeholder
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"verified":  true,
		"did":       h.getDID(r),
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

func (h *V3Handlers) handleMessageReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	// Path: /v3/messages/{messageId}/receipt
	path := r.URL.Path[len("/v3/messages/"):]
	// Extract messageId
	messageID := path
	if idx := indexByte(path, '/'); idx >= 0 {
		messageID = path[:idx]
	}

	var req struct {
		ReceiptType string `json:"receiptType"` // "delivered" or "read"
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if err := h.DB.MarkDelivered(r.Context(), messageID); err != nil {
		WriteError(w, http.StatusNotFound, "MESSAGE_NOT_FOUND", "Message not found", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"messageId":   messageID,
		"receiptType": req.ReceiptType,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
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
	profile := groups.GroupProfile{Name: req.Name, Description: req.Description}
	group, err := h.Groups.CreateGroup(req.GroupID, ownerDID, req.GroupType, profile, req.Requirements)
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
	member, err := h.Groups.AddMember(req.GroupID, req.MemberDID, 0, groups.TrustLevelNewcomer, false)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "MEMBER_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, member)
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
	if err := h.Groups.RemoveMember(req.GroupID, req.MemberID); err != nil {
		WriteError(w, http.StatusNotFound, "MEMBER_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "removed"})
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
