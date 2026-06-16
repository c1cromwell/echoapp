package api

import (
	"context"

	"github.com/thechadcromwell/echoapp/internal/services/notification"
)

// notificationOfflineNotifier adapts the content-blind notification.Service to the
// OfflineNotifier interface used by the WS hub (new messages) and the reaction
// handler (WO-57). It only ever sends a wake-up push: conversation id + sender,
// never message content.
type notificationOfflineNotifier struct {
	svc *notification.Service
}

// NewOfflineNotifier wraps a notification.Service. Returns nil if svc is nil so
// callers can wire it unconditionally.
func NewOfflineNotifier(svc *notification.Service) OfflineNotifier {
	if svc == nil {
		return nil
	}
	return notificationOfflineNotifier{svc: svc}
}

func (n notificationOfflineNotifier) NotifyUndelivered(recipientID, senderID, conversationID string) {
	if n.svc == nil || recipientID == "" {
		return
	}
	// Best-effort: a missing-device / disabled-push / quiet-hours result is not an error here.
	_, _ = n.svc.Send(context.Background(), recipientID, notification.PushPayload{
		Type:           notification.TypeMessage,
		ConversationID: conversationID,
		SenderDID:      senderID,
	})
}

func (n notificationOfflineNotifier) NotifyMissedCall(recipientID, senderID, callID string) {
	if n.svc == nil || recipientID == "" {
		return
	}
	_, _ = n.svc.Send(context.Background(), recipientID, notification.PushPayload{
		Type:           notification.TypeMissedCall,
		ConversationID: callID,
		SenderDID:      senderID,
	})
}
