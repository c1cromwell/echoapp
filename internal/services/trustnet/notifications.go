package trustnet

import (
	"fmt"
	"sync"
	"time"
)

// NotificationEvent represents a trust-related notification event
type NotificationEvent struct {
	ID        string                 `json:"id"`
	UserDID   string                 `json:"user_did"`
	Type      string                 `json:"type"` // "score_change", "endorsement", "dispute", "auto_promotion", "security"
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`

	// Notification routing
	SendEmail bool `json:"send_email"`
	SendPush  bool `json:"send_push"`
	SendInApp bool `json:"send_in_app"`
}

// NotificationPreferences defines user notification settings
type NotificationPreferences struct {
	UserDID               string
	ScoreChangeAlerts     bool
	ScoreMinimumDelta     float64 // Only alert if change >= this much
	EndorsementReceived   bool
	EndorsementRejected   bool
	DisputeAssignment     bool
	DisputeResolution     bool
	AutoPromotionAlerts   bool
	SecurityAlerts        bool
	SybilWarnings         bool
	BlockChainAnchorNotif bool

	// Delivery preferences
	NotifyViaEmail bool
	NotifyViaPush  bool
	NotifyInApp    bool

	// Frequency
	BundleNotifications bool // Group similar notifications
	BundleIntervalHours int
	QuietHoursStart     string // "20:00"
	QuietHoursEnd       string // "08:00"
}

// NotificationProvider is an interface for sending notifications
type NotificationProvider interface {
	SendEmail(userDID string, title string, message string) error
	SendPush(userDID string, title string, message string) error
	SendInAppNotification(notification *NotificationEvent) error
}

// NotificationService manages trust event notifications
type NotificationService struct {
	mu              sync.RWMutex
	events          map[string]*NotificationEvent // by ID
	userPreferences map[string]*NotificationPreferences
	userEvents      map[string][]string // userDID -> event IDs
	provider        NotificationProvider
	pendingQueue    []*NotificationEvent
	notificationLog map[string][]NotificationEvent // userDID -> events sent
}

// NewNotificationService creates a new notification service
func NewNotificationService(provider NotificationProvider) *NotificationService {
	return &NotificationService{
		events:          make(map[string]*NotificationEvent),
		userPreferences: make(map[string]*NotificationPreferences),
		userEvents:      make(map[string][]string),
		provider:        provider,
		notificationLog: make(map[string][]NotificationEvent),
	}
}

// DefaultPreferences returns sensible defaults
func DefaultPreferences(userDID string) *NotificationPreferences {
	return &NotificationPreferences{
		UserDID:               userDID,
		ScoreChangeAlerts:     true,
		ScoreMinimumDelta:     5.0, // Only alert if ±5 or more
		EndorsementReceived:   true,
		EndorsementRejected:   true,
		DisputeAssignment:     true,
		DisputeResolution:     true,
		AutoPromotionAlerts:   true,
		SecurityAlerts:        true,
		SybilWarnings:         true,
		BlockChainAnchorNotif: false,
		NotifyViaEmail:        true,
		NotifyViaPush:         true,
		NotifyInApp:           true,
		BundleNotifications:   true,
		BundleIntervalHours:   1,
		QuietHoursStart:       "20:00",
		QuietHoursEnd:         "08:00",
	}
}

// SetUserPreferences sets notifications preferences for a user
func (ns *NotificationService) SetUserPreferences(prefs *NotificationPreferences) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.userPreferences[prefs.UserDID] = prefs
}

// GetUserPreferences retrieves preferences for a user
func (ns *NotificationService) GetUserPreferences(userDID string) *NotificationPreferences {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	prefs, exists := ns.userPreferences[userDID]
	if !exists {
		return DefaultPreferences(userDID)
	}

	return prefs
}

// NotifyScoreChange creates a notification for trust score changes
func (ns *NotificationService) NotifyScoreChange(userDID string, oldScore float64, newScore float64, reason string) error {
	ns.mu.Lock()
	prefs := ns.userPreferences[userDID]
	ns.mu.Unlock()

	if prefs == nil {
		prefs = DefaultPreferences(userDID)
	}

	if !prefs.ScoreChangeAlerts {
		return nil
	}

	delta := newScore - oldScore
	if delta < 0 {
		delta = -delta
	}

	if delta < prefs.ScoreMinimumDelta {
		return nil // Below threshold
	}

	direction := "increased"
	if newScore < oldScore {
		direction = "decreased"
	}

	event := &NotificationEvent{
		ID:      generateID("notif_score"),
		UserDID: userDID,
		Type:    "score_change",
		Title:   fmt.Sprintf("Your trust score %s", direction),
		Message: fmt.Sprintf("Your trust score %s from %.1f to %.1f. %s", direction, oldScore, newScore, reason),
		Data: map[string]interface{}{
			"old_score": oldScore,
			"new_score": newScore,
			"delta":     newScore - oldScore,
			"reason":    reason,
		},
		CreatedAt: time.Now(),
		SendEmail: prefs.NotifyViaEmail,
		SendPush:  prefs.NotifyViaPush,
		SendInApp: prefs.NotifyInApp,
	}

	return ns.queueNotification(event)
}

// NotifyEndorsementReceived creates a notification for received endorsements
func (ns *NotificationService) NotifyEndorsementReceived(userDID string, endorserDID string, category string) error {
	ns.mu.Lock()
	prefs := ns.userPreferences[userDID]
	ns.mu.Unlock()

	if prefs == nil {
		prefs = DefaultPreferences(userDID)
	}

	if !prefs.EndorsementReceived {
		return nil
	}

	event := &NotificationEvent{
		ID:      generateID("notif_endorse"),
		UserDID: userDID,
		Type:    "endorsement",
		Title:   fmt.Sprintf("You were endorsed for %s", category),
		Message: fmt.Sprintf("Someone endorsed you as %s", category),
		Data: map[string]interface{}{
			"endorser_did": endorserDID,
			"category":     category,
		},
		CreatedAt: time.Now(),
		SendEmail: prefs.NotifyViaEmail,
		SendPush:  prefs.NotifyViaPush,
		SendInApp: prefs.NotifyInApp,
	}

	return ns.queueNotification(event)
}

// NotifyDisputeAssignment notifies when user is assigned as juror
func (ns *NotificationService) NotifyDisputeAssignment(jurorDID string, disputeID string) error {
	ns.mu.Lock()
	prefs := ns.userPreferences[jurorDID]
	ns.mu.Unlock()

	if prefs == nil {
		prefs = DefaultPreferences(jurorDID)
	}

	if !prefs.DisputeAssignment {
		return nil
	}

	event := &NotificationEvent{
		ID:      generateID("notif_dispute"),
		UserDID: jurorDID,
		Type:    "dispute",
		Title:   "You've been selected as a juror",
		Message: "A trust dispute needs your vote. You have 72 hours to review.",
		Data: map[string]interface{}{
			"dispute_id": disputeID,
			"action_url": fmt.Sprintf("/disputes/%s", disputeID),
		},
		CreatedAt: time.Now(),
		SendEmail: prefs.NotifyViaEmail,
		SendPush:  prefs.NotifyViaPush,
		SendInApp: prefs.NotifyInApp,
	}

	return ns.queueNotification(event)
}

// NotifyAutoPromotion notifies about circle auto-promotions
func (ns *NotificationService) NotifyAutoPromotion(userDID string, contactDID string, newTier CircleTier) error {
	ns.mu.Lock()
	prefs := ns.userPreferences[userDID]
	ns.mu.Unlock()

	if prefs == nil {
		prefs = DefaultPreferences(userDID)
	}

	if !prefs.AutoPromotionAlerts {
		return nil
	}

	event := &NotificationEvent{
		ID:      generateID("notif_promo"),
		UserDID: userDID,
		Type:    "auto_promotion",
		Title:   "Contact promoted",
		Message: fmt.Sprintf("A contact has been automatically promoted to %s based on your interactions", newTier),
		Data: map[string]interface{}{
			"contact_did": contactDID,
			"tier":        newTier,
			"action_url":  fmt.Sprintf("/contacts/%s", contactDID),
		},
		CreatedAt: time.Now(),
		SendEmail: prefs.NotifyViaEmail,
		SendPush:  prefs.NotifyViaPush,
		SendInApp: prefs.NotifyInApp,
	}

	return ns.queueNotification(event)
}

// NotifySecurityAlert sends critical security notifications
func (ns *NotificationService) NotifySecurityAlert(userDID string, alertType string, message string) error {
	ns.mu.Lock()
	prefs := ns.userPreferences[userDID]
	ns.mu.Unlock()

	if prefs == nil {
		prefs = DefaultPreferences(userDID)
	}

	if !prefs.SecurityAlerts {
		return nil
	}

	event := &NotificationEvent{
		ID:      generateID("notif_security"),
		UserDID: userDID,
		Type:    "security",
		Title:   "Security alert",
		Message: message,
		Data: map[string]interface{}{
			"alert_type": alertType,
		},
		CreatedAt: time.Now(),
		SendEmail: true, // Always send security alerts via email
		SendPush:  prefs.NotifyViaPush,
		SendInApp: prefs.NotifyInApp,
	}

	return ns.queueNotification(event)
}

// queueNotification adds a notification to the pending queue
func (ns *NotificationService) queueNotification(event *NotificationEvent) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.events[event.ID] = event
	ns.userEvents[event.UserDID] = append(ns.userEvents[event.UserDID], event.ID)
	ns.pendingQueue = append(ns.pendingQueue, event)

	return nil
}

// SendPendingNotifications processes the pending notification queue
func (ns *NotificationService) SendPendingNotifications() map[string]error {
	ns.mu.Lock()
	pending := ns.pendingQueue
	ns.pendingQueue = []*NotificationEvent{}
	ns.mu.Unlock()

	errors := make(map[string]error)

	for _, event := range pending {
		if event.SendInApp && ns.provider != nil {
			if err := ns.provider.SendInAppNotification(event); err != nil {
				errors[event.ID] = err
				continue
			}
		}

		if event.SendEmail && ns.provider != nil {
			if err := ns.provider.SendEmail(event.UserDID, event.Title, event.Message); err != nil {
				errors[event.ID] = err
				continue
			}
		}

		if event.SendPush && ns.provider != nil {
			if err := ns.provider.SendPush(event.UserDID, event.Title, event.Message); err != nil {
				errors[event.ID] = err
				continue
			}
		}

		// Log successful notification
		ns.mu.Lock()
		ns.notificationLog[event.UserDID] = append(ns.notificationLog[event.UserDID], *event)
		ns.mu.Unlock()
	}

	return errors
}

// GetUserNotifications retrieves notifications for a user
func (ns *NotificationService) GetUserNotifications(userDID string, limit int, offset int) []*NotificationEvent {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	eventIDs := ns.userEvents[userDID]
	result := make([]*NotificationEvent, 0)

	// Iterate in reverse (newest first)
	for i := len(eventIDs) - 1; i >= 0; i-- {
		if offset > 0 {
			offset--
			continue
		}

		if limit <= 0 {
			break
		}

		if event, exists := ns.events[eventIDs[i]]; exists {
			result = append(result, event)
			limit--
		}
	}

	return result
}

// MarkNotificationRead marks a notification as read
func (ns *NotificationService) MarkNotificationRead(notificationID string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	event, exists := ns.events[notificationID]
	if !exists {
		return fmt.Errorf("notification not found")
	}

	event.Read = true
	return nil
}

// GetUnreadCount returns the count of unread notifications
func (ns *NotificationService) GetUnreadCount(userDID string) int {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	eventIDs := ns.userEvents[userDID]
	count := 0

	for _, id := range eventIDs {
		if event, exists := ns.events[id]; exists && !event.Read {
			count++
		}
	}

	return count
}

// ClearOldNotifications removes expired notifications
func (ns *NotificationService) ClearOldNotifications(olderThanDays int) int {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	count := 0

	// Remove from main map
	forDelete := []string{}
	for id, event := range ns.events {
		if event.CreatedAt.Before(cutoff) {
			forDelete = append(forDelete, id)
			count++
		}
	}

	for _, id := range forDelete {
		delete(ns.events, id)
	}

	// Remove from user event lists
	for userDID, eventIDs := range ns.userEvents {
		filtered := []string{}
		for _, id := range eventIDs {
			if _, exists := ns.events[id]; exists {
				filtered = append(filtered, id)
			}
		}
		ns.userEvents[userDID] = filtered
	}

	return count
}
