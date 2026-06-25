package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/bots"
)

// botRelayPayload is opaque to the relay — production bots should send ciphertext only.
type botRelayPayload struct {
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text,omitempty"`
	Ciphertext     []byte `json:"ciphertext,omitempty"`
	FromBot        bool   `json:"from_bot"`
}

func (h *V3Handlers) handleBotsRelaySubroute(w http.ResponseWriter, r *http.Request, subpath string) {
	switch {
	case subpath == "message" && r.Method == http.MethodPost:
		h.handleBotRelayMessage(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown bot relay route", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleBotRelayMessage(w http.ResponseWriter, r *http.Request) {
	if h.Bots == nil || h.BotTokens == nil {
		WriteError(w, http.StatusServiceUnavailable, "BOTS_UNAVAILABLE", "bot relay not configured", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Signals == nil {
		WriteError(w, http.StatusServiceUnavailable, "RELAY_UNAVAILABLE", "signal relay not configured", r.Header.Get("X-Request-ID"))
		return
	}

	botDID := r.Header.Get("X-Bot-DID")
	token := r.Header.Get("X-Bot-Token")
	if botDID == "" || token == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "X-Bot-DID and X-Bot-Token required", r.Header.Get("X-Request-ID"))
		return
	}
	if !h.BotTokens.Validate(botDID, token) {
		WriteError(w, http.StatusForbidden, "INVALID_TOKEN", "bot token invalid", r.Header.Get("X-Request-ID"))
		return
	}
	if h.BotRateLimiter != nil && !h.BotRateLimiter.Allow(botDID) {
		WriteError(w, http.StatusTooManyRequests, botsdkCodeRateLimit, "bot rate limit exceeded", r.Header.Get("X-Request-ID"))
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
