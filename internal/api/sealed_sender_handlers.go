package api

import (
	"encoding/json"
	"github.com/thechadcromwell/echoapp/internal/services/messaging"
	"net/http"
	"strings"
)

// handleSealedToken issues a short-lived delivery token for sealed-sender relay (WO-219).
// POST /v3/messages/sealed-token -> {"delivery_token":"..."}
func (h *V3Handlers) handleSealedToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST required", r.Header.Get("X-Request-ID"))
		return
	}
	if h.SealedTokens == nil {
		WriteError(w, http.StatusServiceUnavailable, "SEALED_SENDER_DISABLED", "sealed sender not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	if h.RateLimiter != nil {
		if err := h.RateLimiter.Check(did, "message_send"); err != nil {
			WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded", r.Header.Get("X-Request-ID"))
			return
		}
	}
	token, err := h.SealedTokens.Issue(did)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "could not issue delivery token", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"delivery_token": token,
		"expires_in":     int(messaging.SealedTokenTTLSeconds()),
	})
}

// sealedTextPayload is the client wire body for Type:"sealed_text".
type sealedTextPayload struct {
	DeliveryToken string `json:"delivery_token"`
	Ciphertext    []byte `json:"ciphertext"`
}

func (h *Hub) SetSealedTokenStore(store *messaging.SealedTokenStore) {
	h.sealedTokens = store
}

// routeSealedText relays a sealed message without exposing sender DID to the recipient (WO-219).
func (c *Client) routeSealedText(msg WSMessage) {
	if msg.To == "" || msg.To == c.userID {
		return
	}
	if c.hub.sealedTokens == nil {
		return
	}
	var payload sealedTextPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.DeliveryToken == "" {
		return
	}
	if !c.hub.sealedTokens.Consume(payload.DeliveryToken, c.userID) {
		return
	}
	relay := msg
	relay.From = ""
	outBytes, err := json.Marshal(relay)
	if err != nil {
		return
	}
	c.hub.deliverOrQueue(msg.To, outBytes, "", msg.ConversationID, msg.Silent)
}

// handleMessageSubroute dispatches /v3/messages/* before message-id parsing.
func (h *V3Handlers) handleMessageSubroute(w http.ResponseWriter, r *http.Request) {
	remainder := strings.TrimPrefix(r.URL.Path, "/v3/messages/")
	if remainder == "sealed-token" {
		h.handleSealedToken(w, r)
		return
	}
	if remainder == "schedule" || strings.HasPrefix(remainder, "schedule/") {
		if remainder == "schedule" {
			h.handleScheduledCollection(w, r)
			return
		}
		h.handleScheduledItem(w, r)
		return
	}
	h.handleMessageReceipt(w, r)
}
