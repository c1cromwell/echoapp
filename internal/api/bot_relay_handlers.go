package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/bots"
	"github.com/thechadcromwell/echoapp/internal/services/media"
)

func (h *V3Handlers) handleBotsRelaySubroute(w http.ResponseWriter, r *http.Request, subpath string) {
	switch {
	case subpath == "message" && r.Method == http.MethodPost:
		h.handleBotRelayMessage(w, r)
	case subpath == "webhook" && r.Method == http.MethodPost:
		h.handleBotRegisterWebhook(w, r)
	case subpath == "payment" && r.Method == http.MethodPost:
		h.handleBotRelayPayment(w, r)
	case subpath == "upload" && r.Method == http.MethodPost:
		h.handleBotRelayUpload(w, r)
	case subpath == "chain" && r.Method == http.MethodGet:
		h.handleBotRelayChain(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown bot relay route", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) authenticateBot(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.BotTokens == nil {
		WriteError(w, http.StatusServiceUnavailable, "BOTS_UNAVAILABLE", "bot relay not configured", r.Header.Get("X-Request-ID"))
		return "", false
	}
	botDID := r.Header.Get("X-Bot-DID")
	token := r.Header.Get("X-Bot-Token")
	if botDID == "" || token == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "X-Bot-DID and X-Bot-Token required", r.Header.Get("X-Request-ID"))
		return "", false
	}
	if !h.BotTokens.Validate(botDID, token) {
		WriteError(w, http.StatusForbidden, "INVALID_TOKEN", "bot token invalid", r.Header.Get("X-Request-ID"))
		return "", false
	}
	if h.BotRateLimiter != nil && !h.BotRateLimiter.Allow(botDID) {
		WriteError(w, http.StatusTooManyRequests, botsdkCodeRateLimit, "bot rate limit exceeded", r.Header.Get("X-Request-ID"))
		return "", false
	}
	return botDID, true
}

func (h *V3Handlers) handleBotRegisterWebhook(w http.ResponseWriter, r *http.Request) {
	botDID, ok := h.authenticateBot(w, r)
	if !ok {
		return
	}
	if h.BotWebhooks == nil {
		WriteError(w, http.StatusServiceUnavailable, "WEBHOOK_UNAVAILABLE", "webhook registry not configured", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		URL string `json:"url"`
	}
	if err := h.readJSON(r, &req); err != nil || req.URL == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "url required", r.Header.Get("X-Request-ID"))
		return
	}
	h.BotWebhooks.Set(botDID, req.URL)
	WriteJSON(w, http.StatusOK, map[string]interface{}{"registered": true})
}

func (h *V3Handlers) handleBotRelayPayment(w http.ResponseWriter, r *http.Request) {
	botDID, ok := h.authenticateBot(w, r)
	if !ok {
		return
	}
	var req struct {
		UserDID string `json:"user_did"`
		Amount  string `json:"amount"`
		Reason  string `json:"reason"`
	}
	if err := h.readJSON(r, &req); err != nil || req.UserDID == "" || req.Amount == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "user_did and amount required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := bots.Authorize(h.Bots, req.UserDID, botDID, bots.PermRequestPayment); err != nil {
		WriteError(w, http.StatusForbidden, "PERMISSION_DENIED", "request_payment not granted", r.Header.Get("X-Request-ID"))
		return
	}
	id := make([]byte, 8)
	_, _ = rand.Read(id)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"payment": map[string]interface{}{
			"id":         "pay-" + hex.EncodeToString(id),
			"user_did":   req.UserDID,
			"amount":     req.Amount,
			"reason":     req.Reason,
			"created_at": time.Now().UTC().Format(time.RFC3339),
			"status":     "pending_user_approval",
		},
	})
}

func (h *V3Handlers) handleBotRelayUpload(w http.ResponseWriter, r *http.Request) {
	botDID, ok := h.authenticateBot(w, r)
	if !ok {
		return
	}
	if h.Media == nil {
		WriteError(w, http.StatusServiceUnavailable, "MEDIA_UNAVAILABLE", "media service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		UserDID    string `json:"user_did"`
		Ciphertext []byte `json:"ciphertext"`
		MimeType   string `json:"mime_type"`
	}
	if err := h.readJSON(r, &req); err != nil || req.UserDID == "" || len(req.Ciphertext) == 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "user_did and ciphertext required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := bots.Authorize(h.Bots, req.UserDID, botDID, bots.PermUploadFile); err != nil {
		WriteError(w, http.StatusForbidden, "PERMISSION_DENIED", "upload_file not granted", r.Header.Get("X-Request-ID"))
		return
	}
	mime := req.MimeType
	if mime == "" {
		mime = "application/octet-stream"
	}
	result, err := h.Media.Upload(r.Context(), media.UploadRequest{
		UploaderDID:   req.UserDID,
		ContentType:   mime,
		EncryptedSize: int64(len(req.Ciphertext)),
		TrustTier:     3,
	}, bytes.NewReader(req.Ciphertext))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "UPLOAD_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"media_id": result.FileID})
}

func (h *V3Handlers) handleBotRelayChain(w http.ResponseWriter, r *http.Request) {
	botDID, ok := h.authenticateBot(w, r)
	if !ok {
		return
	}
	userDID := r.URL.Query().Get("user_did")
	if userDID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_did query required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := bots.Authorize(h.Bots, userDID, botDID, bots.PermReadChainState); err != nil {
		WriteError(w, http.StatusForbidden, "PERMISSION_DENIED", "read_chain_state not granted", r.Header.Get("X-Request-ID"))
		return
	}
	module := r.URL.Query().Get("module")
	key := r.URL.Query().Get("key")
	out := map[string]interface{}{
		"module": module,
		"key":    key,
		"state":  nil,
	}
	if h.IdentityL1 != nil && module == "identity" && key != "" {
		out["state"] = map[string]string{"did": key, "anchored": "unknown"}
	}
	WriteJSON(w, http.StatusOK, out)
}

// botRelayPayload is opaque to the relay — production bots should send ciphertext only.
type botRelayPayload struct {
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text,omitempty"`
	Ciphertext     []byte `json:"ciphertext,omitempty"`
	FromBot        bool   `json:"from_bot"`
}

func (h *V3Handlers) handleBotRelayMessage(w http.ResponseWriter, r *http.Request) {
	if h.Bots == nil {
		WriteError(w, http.StatusServiceUnavailable, "BOTS_UNAVAILABLE", "bot service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Signals == nil {
		WriteError(w, http.StatusServiceUnavailable, "RELAY_UNAVAILABLE", "signal relay not configured", r.Header.Get("X-Request-ID"))
		return
	}
	botDID, ok := h.authenticateBot(w, r)
	if !ok {
		return
	}

	var req struct {
		RecipientDID   string `json:"recipient_did"`
		ConversationID string `json:"conversation_id"`
		Text           string `json:"text,omitempty"`
		Ciphertext     []byte `json:"ciphertext,omitempty"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if req.RecipientDID == "" || req.ConversationID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "recipient_did and conversation_id required", r.Header.Get("X-Request-ID"))
		return
	}
	if req.Text == "" && len(req.Ciphertext) == 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "text or ciphertext required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := bots.Authorize(h.Bots, req.RecipientDID, botDID, bots.PermSendMessage); err != nil {
		WriteError(w, http.StatusForbidden, "PERMISSION_DENIED", "send_message not granted", r.Header.Get("X-Request-ID"))
		return
	}

	payload, err := json.Marshal(botRelayPayload{
		ConversationID: req.ConversationID,
		Text:           req.Text,
		Ciphertext:     req.Ciphertext,
		FromBot:        true,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "RELAY_FAILED", "payload encode failed", r.Header.Get("X-Request-ID"))
		return
	}

	delivered := h.Signals.PublishSignal(req.RecipientDID, WSMessage{
		Type:           "text",
		From:           botDID,
		To:             req.RecipientDID,
		ConversationID: req.ConversationID,
		Payload:        payload,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"delivered": delivered,
	})
}

const botsdkCodeRateLimit = "RATE_LIMIT_EXCEEDED"
