package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/messaging"
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
		debugTrace("sealed message from %s dropped: no delivery token in it", c.userID)
		return
	}
	if !c.hub.sealedTokens.Consume(payload.DeliveryToken, c.userID) {
		debugTrace("sealed message from %s dropped: its delivery token was not accepted", c.userID)
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

// --- Sealed delivery the server cannot see through (docs/RELAY_METADATA_REVIEW.md §5) ---------
//
// The token path above hides the sender from the recipient's envelope but not from the server:
// the sender is logged in when it asks for a token and when it sends. The path below carries no
// login at all. A sender presents the recipient's delivery key, which only that recipient's
// contacts hold; the server learns the recipient and nothing about the sender.

const (
	sealedDeliverPath = "/v1/sealed/deliver"
	// sealedConversationID is what a sealed message carries where a conversation id would go.
	// The recipient's device works out the conversation from the sender named inside the
	// encrypted payload; the wire names no one.
	sealedConversationID = "dm:sealed"
	maxSealedBody        = 96 * 1024
)

// sealedDeliverRequest is the body of POST /v1/sealed/deliver.
type sealedDeliverRequest struct {
	// DeliveryKey is the recipient's 32-byte delivery key, hex.
	DeliveryKey string `json:"delivery_key"`
	// Payload is opaque to the server: the sealed, encrypted message.
	Payload json.RawMessage `json:"payload"`
	Silent  bool            `json:"silent,omitempty"`
	// MessageID and CommitmentHash are optional. When both are present the message's integrity
	// commitment is anchored under that id, with no sender recorded. The id is random and is
	// otherwise known only to the two devices.
	MessageID      string `json:"message_id,omitempty"`
	CommitmentHash string `json:"commitment_hash,omitempty"`
}

// handleSealedDeliver accepts a sealed message for whoever owns the presented delivery key.
// It is public: there is no token to read and none is looked for. It must never record or
// derive anything about the caller.
func (rt *Router) handleSealedDeliver(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST required", reqID)
		return
	}
	if rt.DeliveryKeys == nil || rt.WSHub == nil {
		WriteError(w, http.StatusServiceUnavailable, "SEALED_DELIVERY_DISABLED", "sealed delivery is not configured", reqID)
		return
	}
	var req sealedDeliverRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxSealedBody)).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", reqID)
		return
	}
	key, err := hex.DecodeString(strings.TrimSpace(req.DeliveryKey))
	if err != nil || len(key) != 32 || len(req.Payload) == 0 || string(req.Payload) == "null" {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "delivery_key must be 32 bytes of hex and payload must be present", reqID)
		return
	}
	hash := messaging.HashDeliveryKey(key)
	recipient, ok, err := rt.DeliveryKeys.Resolve(r.Context(), hash)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL", "could not look up the delivery key", reqID)
		return
	}
	if !ok {
		// The sender's copy of the key is stale (the recipient rotated it) or wrong.
		WriteError(w, http.StatusNotFound, "UNKNOWN_DELIVERY_KEY", "no account accepts this delivery key", reqID)
		return
	}
	// There is no sender to limit, so the limit is on what one recipient can be sent.
	if rt.RateLimiter != nil {
		if err := rt.RateLimiter.Check("sealed:"+hash, "message_send"); err != nil {
			WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "this recipient is receiving too many messages", reqID)
			return
		}
	}
	if rt.WSHub.commitments != nil && req.MessageID != "" && len(req.MessageID) <= 64 {
		if commitment, err := hex.DecodeString(req.CommitmentHash); err == nil && len(commitment) == 32 {
			rt.WSHub.commitments.Add(req.MessageID, "", commitment)
		}
	}
	rt.WSHub.DeliverSealed(recipient, req.Payload, req.Silent)
	WriteJSON(w, http.StatusAccepted, map[string]interface{}{"accepted": true, "request_id": reqID})
}

// DeliverSealed hands a sealed message to a recipient, live or queued. The envelope names the
// recipient and no one else: no sender, and a constant where a conversation id would go.
func (h *Hub) DeliverSealed(recipient string, payload json.RawMessage, silent bool) {
	data, err := json.Marshal(WSMessage{
		Type:           "sealed_text",
		To:             recipient,
		ConversationID: sealedConversationID,
		Payload:        payload,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Silent:         silent,
	})
	if err != nil {
		return
	}
	h.deliverOrQueue(recipient, data, "", "", silent)
}

// handleDeliveryKey registers the caller's delivery key hash.
// PUT /v3/messages/delivery-key {"key_hash":"<64 hex>"}
func (h *V3Handlers) handleDeliveryKey(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodPut {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "PUT required", reqID)
		return
	}
	if h.DeliveryKeys == nil {
		WriteError(w, http.StatusServiceUnavailable, "SEALED_DELIVERY_DISABLED", "sealed delivery is not configured", reqID)
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", reqID)
		return
	}
	var body struct {
		KeyHash string `json:"key_hash"`
		// RevokeOthers makes this the account's only delivery key at once. Used after blocking
		// someone: the copy they hold stops working immediately.
		RevokeOthers bool `json:"revoke_others,omitempty"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", reqID)
		return
	}
	if err := h.DeliveryKeys.Register(r.Context(), did, strings.ToLower(strings.TrimSpace(body.KeyHash)), body.RevokeOthers); err != nil {
		if errors.Is(err, messaging.ErrBadDeliveryKeyHash) {
			WriteError(w, http.StatusBadRequest, "INVALID_KEY_HASH", err.Error(), reqID)
			return
		}
		WriteError(w, http.StatusConflict, "DELIVERY_KEY_REJECTED", "this delivery key cannot be used; generate a new one", reqID)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"registered": true, "request_id": reqID})
}

// handleMessageSubroute dispatches /v3/messages/* before message-id parsing.
func (h *V3Handlers) handleMessageSubroute(w http.ResponseWriter, r *http.Request) {
	remainder := strings.TrimPrefix(r.URL.Path, "/v3/messages/")
	if remainder == "delivery-key" {
		h.handleDeliveryKey(w, r)
		return
	}
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
