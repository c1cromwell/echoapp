package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/messaging"
)

// scheduledCreateRequest is the client body for POST /v3/messages/schedule.
// Content is an opaque blob (ciphertext). Handlers never log it (T0).
type scheduledCreateRequest struct {
	ConversationID string          `json:"conversation_id"`
	Content        json.RawMessage `json:"content"`
	ContentType    string          `json:"content_type"`
	ScheduledAt    time.Time       `json:"scheduled_at"`
	Timezone       string          `json:"timezone"`
	Silent         bool            `json:"silent"`
}

type scheduledPatchRequest struct {
	Content     json.RawMessage `json:"content"`
	ScheduledAt *time.Time      `json:"scheduled_at"`
}

type scheduledMessageDTO struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	Timezone       string    `json:"timezone,omitempty"`
	Status         string    `json:"status"`
	Silent         bool      `json:"silent"`
	ContentType    string    `json:"content_type,omitempty"`
	Content        []byte    `json:"content,omitempty"`
}

func scheduledToDTO(msg *messaging.ScheduledMessage, includeContent bool) scheduledMessageDTO {
	dto := scheduledMessageDTO{
		ID:             msg.ID,
		ConversationID: msg.ConvID,
		ScheduledAt:    msg.ScheduledAt,
		Timezone:       msg.Timezone,
		Status:         string(msg.Status),
		Silent:         msg.Silent,
		ContentType:    scheduledContentTypeName(msg.ContentType),
	}
	if includeContent {
		dto.Content = msg.Content
	}
	return dto
}

func scheduledContentTypeName(t messaging.MessageType) string {
	if t == messaging.VoiceMessage {
		return "voice"
	}
	return "text"
}

func parseScheduledContentType(s string) messaging.MessageType {
	if strings.EqualFold(s, "voice") {
		return messaging.VoiceMessage
	}
	return messaging.TextMessage
}

func decodeOpaqueContent(raw json.RawMessage) ([]byte, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, errors.New("content required")
	}
	// JSON strings are UTF-8 ciphertext/previews from clients. []byte would be
	// interpreted as base64 by encoding/json — only use that for JSON arrays.
	if raw[0] == '"' {
		var asString string
		if err := json.Unmarshal(raw, &asString); err != nil {
			return nil, errors.New("content must be a string or byte array")
		}
		if asString == "" {
			return nil, errors.New("content required")
		}
		return []byte(asString), nil
	}
	var asBytes []byte
	if err := json.Unmarshal(raw, &asBytes); err != nil || len(asBytes) == 0 {
		return nil, errors.New("content must be a string or byte array")
	}
	return asBytes, nil
}

func (h *V3Handlers) requireScheduled(w http.ResponseWriter, r *http.Request) (*messaging.ScheduledMessageService, string, bool) {
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return nil, "", false
	}
	if h.Scheduled == nil {
		WriteError(w, http.StatusServiceUnavailable, "SCHEDULE_DISABLED", "scheduled messages not configured", r.Header.Get("X-Request-ID"))
		return nil, "", false
	}
	return h.Scheduled, did, true
}

func writeScheduledError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, messaging.ErrScheduledNotFound):
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "scheduled message not found", r.Header.Get("X-Request-ID"))
	case errors.Is(err, messaging.ErrScheduledNotOwner):
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "only the sender can modify this scheduled message", r.Header.Get("X-Request-ID"))
	case errors.Is(err, messaging.ErrScheduledAlreadyDelivered):
		WriteError(w, http.StatusConflict, "ALREADY_DELIVERED", "scheduled message already delivered or cancelled", r.Header.Get("X-Request-ID"))
	case errors.Is(err, messaging.ErrScheduledEditTooLate):
		WriteError(w, http.StatusConflict, "EDIT_TOO_LATE", "cannot edit within 5 minutes of delivery", r.Header.Get("X-Request-ID"))
	case errors.Is(err, messaging.ErrScheduledTimeInPast):
		WriteError(w, http.StatusBadRequest, "INVALID_TIME", "scheduled time must be in the future", r.Header.Get("X-Request-ID"))
	case errors.Is(err, messaging.ErrInvalidSender):
		WriteError(w, http.StatusBadRequest, "INVALID_SENDER", "sender required", r.Header.Get("X-Request-ID"))
	default:
		WriteError(w, http.StatusBadRequest, "SCHEDULE_ERROR", "could not complete scheduled-message request", r.Header.Get("X-Request-ID"))
	}
}

// handleScheduledCollection POST creates a scheduled message; GET lists the caller's pending items (no content).
func (h *V3Handlers) handleScheduledCollection(w http.ResponseWriter, r *http.Request) {
	svc, did, ok := h.requireScheduled(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		var req scheduledCreateRequest
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if req.ConversationID == "" {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "conversation_id required", r.Header.Get("X-Request-ID"))
			return
		}
		content, err := decodeOpaqueContent(req.Content)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		ct := parseScheduledContentType(req.ContentType)
		tz := req.Timezone
		var msg *messaging.ScheduledMessage
		if req.Silent {
			msg, err = svc.ScheduleSilent(did, req.ConversationID, content, ct, req.ScheduledAt, tz, nil)
		} else {
			msg, err = svc.Schedule(did, req.ConversationID, content, ct, req.ScheduledAt, tz)
		}
		if err != nil {
			writeScheduledError(w, r, err)
			return
		}
		WriteJSON(w, http.StatusCreated, scheduledToDTO(msg, true))
	case http.MethodGet:
		pending := svc.GetPending(did)
		out := make([]scheduledMessageDTO, 0, len(pending))
		for _, msg := range pending {
			out = append(out, scheduledToDTO(msg, false))
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"messages": out})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST required", r.Header.Get("X-Request-ID"))
	}
}

// handleScheduledItem GET/PATCH/DELETE /v3/messages/schedule/{id} and POST .../send-now.
func (h *V3Handlers) handleScheduledItem(w http.ResponseWriter, r *http.Request) {
	svc, did, ok := h.requireScheduled(w, r)
	if !ok {
		return
	}
	remainder := strings.TrimPrefix(r.URL.Path, "/v3/messages/schedule/")
	remainder = strings.Trim(remainder, "/")
	if remainder == "" {
		h.handleScheduledCollection(w, r)
		return
	}
	parts := strings.Split(remainder, "/")
	id := parts[0]
	if id == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_PATH", "scheduled message id required", r.Header.Get("X-Request-ID"))
		return
	}
	if len(parts) == 2 && parts[1] == "send-now" {
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST required", r.Header.Get("X-Request-ID"))
			return
		}
		delivered, err := svc.SendNow(id, did)
		if err != nil {
			writeScheduledError(w, r, err)
			return
		}
		resp := map[string]interface{}{
			"id":     id,
			"status": "delivered",
		}
		if delivered != nil {
			resp["delivered_id"] = delivered.ID
		}
		WriteJSON(w, http.StatusOK, resp)
		return
	}
	if len(parts) != 1 {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown scheduled-message path", r.Header.Get("X-Request-ID"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		msg, err := svc.Get(id)
		if err != nil {
			writeScheduledError(w, r, err)
			return
		}
		if msg.SenderID != did {
			WriteError(w, http.StatusForbidden, "FORBIDDEN", "only the sender can read this scheduled message", r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, scheduledToDTO(msg, true))
	case http.MethodPatch:
		var req scheduledPatchRequest
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		if req.Content == nil && req.ScheduledAt == nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "content or scheduled_at required", r.Header.Get("X-Request-ID"))
			return
		}
		var msg *messaging.ScheduledMessage
		var err error
		if req.Content != nil {
			content, decErr := decodeOpaqueContent(req.Content)
			if decErr != nil {
				WriteError(w, http.StatusBadRequest, "INVALID_BODY", decErr.Error(), r.Header.Get("X-Request-ID"))
				return
			}
			msg, err = svc.Edit(id, did, content)
			if err != nil {
				writeScheduledError(w, r, err)
				return
			}
		}
		if req.ScheduledAt != nil {
			msg, err = svc.Reschedule(id, did, *req.ScheduledAt)
			if err != nil {
				writeScheduledError(w, r, err)
				return
			}
		}
		WriteJSON(w, http.StatusOK, scheduledToDTO(msg, true))
	case http.MethodDelete:
		if err := svc.Cancel(id, did); err != nil {
			writeScheduledError(w, r, err)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"id": id, "status": "cancelled"})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET, PATCH, or DELETE required", r.Header.Get("X-Request-ID"))
	}
}
