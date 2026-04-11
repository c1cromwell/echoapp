// Package notification implements the APNs push notification service.
// Content-blind: only sends conversation IDs and wake-up signals, never message content.
package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// NotificationType identifies the notification category.
type NotificationType string

const (
	TypeMessage  NotificationType = "message"
	TypeGroup    NotificationType = "group"
	TypeChannel  NotificationType = "channel"
	TypeSystem   NotificationType = "system"
	TypeReward   NotificationType = "reward"
	TypeIdentity NotificationType = "identity"
)

var (
	ErrInvalidToken = errors.New("invalid APNs token")
	ErrNoDevices    = errors.New("no registered devices for recipient")
	ErrPushDisabled = errors.New("push notifications disabled by user")
)

// PushPayload is the content-blind notification payload.
type PushPayload struct {
	Type           NotificationType `json:"type"`
	ConversationID string           `json:"conversationId,omitempty"`
	SenderDID      string           `json:"senderDid,omitempty"`
}

// SendResult represents the result of a push notification send.
type SendResult struct {
	RecipientDID string    `json:"recipientDid"`
	DeviceCount  int       `json:"deviceCount"`
	Delivered    int       `json:"delivered"`
	Timestamp    time.Time `json:"timestamp"`
}

// PushSender abstracts the push notification delivery backend.
type PushSender interface {
	SendPush(ctx context.Context, deviceToken string, conversationID string, notifType string) error
}

// Service provides push notification operations.
type Service struct {
	db     database.DB
	pusher PushSender
}

// NewService creates a notification service.
func NewService(db database.DB) *Service {
	return &Service{db: db}
}

// SetPushSender configures a real push delivery backend (e.g., APNs).
func (s *Service) SetPushSender(p PushSender) {
	s.pusher = p
}

// RegisterDevice registers a device for push notifications.
func (s *Service) RegisterDevice(ctx context.Context, did, deviceLabel, publicKey, apnsToken string) (*database.UserDevice, error) {
	if len(apnsToken) < 8 {
		return nil, ErrInvalidToken
	}

	device := &database.UserDevice{
		DeviceID:    uuid.New().String(),
		DID:         did,
		DeviceLabel: deviceLabel,
		PublicKey:   publicKey,
		APNsToken:   apnsToken,
	}

	if err := s.db.RegisterDevice(ctx, device); err != nil {
		return nil, err
	}
	return device, nil
}

// Send sends a content-blind push notification to a recipient.
func (s *Service) Send(ctx context.Context, recipientDID string, payload PushPayload) (*SendResult, error) {
	// Check preferences
	prefs, err := s.db.GetNotificationPrefs(ctx, recipientDID)
	if err != nil {
		return nil, err
	}

	if !prefs.PushEnabled && payload.Type != TypeSystem {
		return nil, ErrPushDisabled
	}

	// Check category preferences
	if !s.categoryAllowed(prefs, payload.Type) {
		return nil, ErrPushDisabled
	}

	// Check quiet hours
	if s.inQuietHours(prefs) && payload.Type != TypeSystem {
		return nil, ErrPushDisabled
	}

	// Get devices
	devices, err := s.db.GetDevicesByDID(ctx, recipientDID)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, ErrNoDevices
	}

	// In production, this would call APNs HTTP/2 API.
	// When a PushSender is configured, use it for real delivery.
	delivered := 0
	for _, d := range devices {
		if len(d.APNsToken) >= 8 {
			if s.pusher != nil {
				if err := s.pusher.SendPush(ctx, d.APNsToken, payload.ConversationID, string(payload.Type)); err != nil {
					continue
				}
			}
			delivered++
		}
	}

	return &SendResult{
		RecipientDID: recipientDID,
		DeviceCount:  len(devices),
		Delivered:    delivered,
		Timestamp:    time.Now(),
	}, nil
}

// SendBatch sends notifications to multiple recipients.
func (s *Service) SendBatch(ctx context.Context, recipientDIDs []string, payload PushPayload) ([]*SendResult, error) {
	var results []*SendResult
	for _, did := range recipientDIDs {
		result, err := s.Send(ctx, did, payload)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

// GetPreferences returns notification preferences for a user.
func (s *Service) GetPreferences(ctx context.Context, did string) (*database.NotificationPrefs, error) {
	return s.db.GetNotificationPrefs(ctx, did)
}

// UpdatePreferences updates notification preferences.
func (s *Service) UpdatePreferences(ctx context.Context, prefs *database.NotificationPrefs) error {
	return s.db.UpdateNotificationPrefs(ctx, prefs)
}

// UpdateAPNsToken updates the APNs token for a device.
func (s *Service) UpdateAPNsToken(ctx context.Context, deviceID, token string) error {
	if len(token) < 8 {
		return ErrInvalidToken
	}
	return s.db.UpdateAPNsToken(ctx, deviceID, token)
}

// QueueDepth returns the number of registered devices for a DID.
func (s *Service) QueueDepth(ctx context.Context, did string) (int, error) {
	devices, err := s.db.GetDevicesByDID(ctx, did)
	if err != nil {
		return 0, err
	}
	return len(devices), nil
}

func (s *Service) categoryAllowed(prefs *database.NotificationPrefs, t NotificationType) bool {
	switch t {
	case TypeSystem:
		return true // System notifications always delivered
	case TypeGroup:
		return prefs.GroupNotifications
	case TypeChannel:
		return prefs.ChannelNotifications
	default:
		return prefs.PushEnabled
	}
}

func (s *Service) inQuietHours(prefs *database.NotificationPrefs) bool {
	if prefs.QuietHoursStart == prefs.QuietHoursEnd {
		return false // Quiet hours disabled
	}
	hour := time.Now().Hour()
	if prefs.QuietHoursStart < prefs.QuietHoursEnd {
		return hour >= prefs.QuietHoursStart && hour < prefs.QuietHoursEnd
	}
	// Wraps midnight (e.g., 22:00 - 07:00)
	return hour >= prefs.QuietHoursStart || hour < prefs.QuietHoursEnd
}
