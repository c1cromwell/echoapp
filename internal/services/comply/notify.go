package comply

import (
	"context"

	"github.com/thechadcromwell/echoapp/internal/services/notification"
)

// PushNotifier sends content-blind system pushes to custodians (WO-251).
type PushNotifier interface {
	NotifyLitigationHold(ctx context.Context, custodianDID, matterID string) error
}

type notificationAdapter struct {
	svc *notification.Service
}

func NewPushNotifier(svc *notification.Service) PushNotifier {
	if svc == nil {
		return nil
	}
	return &notificationAdapter{svc: svc}
}

func (n *notificationAdapter) NotifyLitigationHold(ctx context.Context, custodianDID, matterID string) error {
	_, err := n.svc.Send(ctx, custodianDID, notification.PushPayload{
		Type:           notification.TypeSystem,
		ConversationID: matterID,
	})
	return err
}
