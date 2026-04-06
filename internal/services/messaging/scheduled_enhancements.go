package messaging

import (
	"fmt"
	"sync"
	"time"
)

// ConditionalDeliveryTrigger extends scheduled messages with conditional delivery
type ConditionalDeliveryTrigger string

const (
	TriggerTimeOnly          ConditionalDeliveryTrigger = "time_only"
	TriggerWhenOnline        ConditionalDeliveryTrigger = "when_online"
	TriggerAfterLastRead     ConditionalDeliveryTrigger = "after_last_read"
	TriggerNoResponseTimeout ConditionalDeliveryTrigger = "no_response_timeout"
	TriggerConversationEnded ConditionalDeliveryTrigger = "conversation_ended"
)

// DeliveryFailureRecord tracks failed message deliveries
type DeliveryFailureRecord struct {
	MessageID      string
	FailureType    string // "offline", "blocked", "network", "key_release_failed", "account_deleted"
	FailureReason  string
	FailureTime    time.Time
	RetryCount     int
	MaxRetries     int
	NextRetryTime  *time.Time
	ExpiryTime     time.Time
	SenderNotified bool
	Status         string // "queued", "expired", "notified", "resolved"
}

// SmartSchedulingSuggestion provides optimal delivery time recommendation
type SmartSchedulingSuggestion struct {
	RecipientID         string
	SuggestedTime       time.Time
	Confidence          float64 // 0.0-1.0
	Reason              string  // "typical_active_hours", "business_hours", "common_response_time"
	AlternativeTimes    []time.Time
	BusinessHoursStart  int
	BusinessHoursEnd    int
	TypicalActiveHours  []int // Hours when recipient is typically active
	LastActivityTime    *time.Time
	AverageResponseTime time.Duration
}

// RecipientMessageFilter defines recipient's message preferences
type RecipientMessageFilter struct {
	RecipientID                 string
	AllowSilentFromTrusted      bool
	AllowSilentFromMembers      bool
	AllowSilentFromNewcomers    bool
	AllowScheduledFromTrusted   bool
	AllowScheduledFromMembers   bool
	AllowScheduledFromNewcomers bool
	AutoApproveFromTrusted      bool
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

// ScheduledMessageEnhancements provides advanced scheduling features
type ScheduledMessageEnhancements struct {
	mu                sync.RWMutex
	conditionalMap    map[string]ConditionalDeliveryTrigger
	failureRecords    map[string]*DeliveryFailureRecord
	suggestions       map[string]*SmartSchedulingSuggestion
	recipientFilters  map[string]*RecipientMessageFilter
	timezoneOffsets   map[string]int // ConvID:ParticipantID -> offset in minutes
	editHistory       map[string][]EditRecord
	recipientControls map[string]*RecipientSilentControl
}

// EditRecord tracks message edit history
type EditRecord struct {
	Timestamp  time.Time
	OldContent []byte
	NewContent []byte
	Reason     string
}

// RecipientSilentControl tracks recipient's silent message preferences
type RecipientSilentControl struct {
	RecipientID                   string
	BlockedSilentSenders          []string
	RequireApprovalFromNonTrusted bool
	ImportantOverrideDaily        int
	ImportantOverridesUsed        int
	ImportantOverrideResetTime    time.Time
	DailyDigestEnabled            bool
	DigestDeliveryTime            time.Time
	EmergencyOverrideContacts     []string
	GlobalDisableSilentMessages   bool
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
}

// NewScheduledMessageEnhancements creates the enhancements module
func NewScheduledMessageEnhancements() *ScheduledMessageEnhancements {
	return &ScheduledMessageEnhancements{
		conditionalMap:    make(map[string]ConditionalDeliveryTrigger),
		failureRecords:    make(map[string]*DeliveryFailureRecord),
		suggestions:       make(map[string]*SmartSchedulingSuggestion),
		recipientFilters:  make(map[string]*RecipientMessageFilter),
		timezoneOffsets:   make(map[string]int),
		editHistory:       make(map[string][]EditRecord),
		recipientControls: make(map[string]*RecipientSilentControl),
	}
}

// SetConditionalTrigger sets a conditional delivery trigger
func (sme *ScheduledMessageEnhancements) SetConditionalTrigger(
	messageID string,
	trigger ConditionalDeliveryTrigger,
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	sme.conditionalMap[messageID] = trigger
	return nil
}

// GetSmartSchedulingSuggestion provides optimal delivery time suggestion
func (sme *ScheduledMessageEnhancements) GetSmartSchedulingSuggestion(
	recipientID string,
	recipientTZOffset int,
) *SmartSchedulingSuggestion {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	if existing, ok := sme.suggestions[recipientID]; ok {
		return existing
	}

	// Suggest 9 AM in recipient's timezone
	now := time.Now()
	suggestedTime := now.AddDate(0, 0, 1) // Tomorrow
	suggestedTime = time.Date(suggestedTime.Year(), suggestedTime.Month(), suggestedTime.Day(),
		9, 0, 0, 0, time.UTC)

	suggestion := &SmartSchedulingSuggestion{
		RecipientID:        recipientID,
		SuggestedTime:      suggestedTime,
		Confidence:         0.75,
		Reason:             "typical_active_hours",
		BusinessHoursStart: 9,
		BusinessHoursEnd:   17,
		AlternativeTimes: []time.Time{
			suggestedTime.Add(1 * time.Hour),
			suggestedTime.Add(2 * time.Hour),
			suggestedTime.Add(3 * time.Hour),
		},
		LastActivityTime:    &now,
		AverageResponseTime: 4 * time.Hour,
	}

	return suggestion
}

// HandleDeliveryFailure records a delivery failure
func (sme *ScheduledMessageEnhancements) HandleDeliveryFailure(
	messageID string,
	failureType string,
	failureReason string,
) *DeliveryFailureRecord {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	failure := &DeliveryFailureRecord{
		MessageID:     messageID,
		FailureType:   failureType,
		FailureReason: failureReason,
		FailureTime:   time.Now(),
		MaxRetries:    3,
		ExpiryTime:    time.Now().Add(7 * 24 * time.Hour),
		Status:        "queued",
	}

	// Set next retry time based on failure type
	switch failureType {
	case "offline":
		nextRetry := time.Now().Add(1 * time.Hour)
		failure.NextRetryTime = &nextRetry
	case "blocked":
		failure.Status = "expired"
	case "account_deleted":
		failure.Status = "expired"
		failure.SenderNotified = true
	case "network":
		nextRetry := time.Now().Add(5 * time.Minute)
		failure.NextRetryTime = &nextRetry
	}

	sme.failureRecords[messageID] = failure
	return failure
}

// RecordEditHistory tracks message edits
func (sme *ScheduledMessageEnhancements) RecordEditHistory(
	messageID string,
	oldContent []byte,
	newContent []byte,
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	record := EditRecord{
		Timestamp:  time.Now(),
		OldContent: oldContent,
		NewContent: newContent,
	}

	sme.editHistory[messageID] = append(sme.editHistory[messageID], record)
	return nil
}

// GetEditHistory returns edit history for a message
func (sme *ScheduledMessageEnhancements) GetEditHistory(messageID string) []EditRecord {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	if history, ok := sme.editHistory[messageID]; ok {
		return history
	}
	return []EditRecord{}
}

// SetConversationTimezone stores timezone offset for a conversation participant
func (sme *ScheduledMessageEnhancements) SetConversationTimezone(
	conversationID string,
	participantID string,
	timezoneOffset int, // in minutes from UTC
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	sme.timezoneOffsets[key] = timezoneOffset
	return nil
}

// GetConversationTimezone retrieves timezone offset
func (sme *ScheduledMessageEnhancements) GetConversationTimezone(
	conversationID string,
	participantID string,
) (int, error) {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	if offset, ok := sme.timezoneOffsets[key]; ok {
		return offset, nil
	}
	return 0, fmt.Errorf("timezone not set")
}

// ValidateScheduleInWakingHours checks if scheduled time is reasonable
func (sme *ScheduledMessageEnhancements) ValidateScheduleInWakingHours(
	scheduledTime time.Time,
	recipientTZOffset int,
) (bool, string) {
	// Convert to recipient's timezone
	recipientTime := scheduledTime.Add(time.Duration(recipientTZOffset) * time.Minute)
	hour := recipientTime.Hour()

	// Consider reasonable hours: 7 AM to 11 PM
	if hour < 7 || hour > 23 {
		return false, fmt.Sprintf("scheduled time (%d:%00d) is outside typical waking hours (7 AM-11 PM) for recipient",
			hour, recipientTime.Minute())
	}

	return true, ""
}

// SetRecipientMessageFilter sets message filtering preferences
func (sme *ScheduledMessageEnhancements) SetRecipientMessageFilter(
	recipientID string,
	filter *RecipientMessageFilter,
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	sme.recipientFilters[recipientID] = filter
	return nil
}

// GetRecipientMessageFilter retrieves message filter for recipient
func (sme *ScheduledMessageEnhancements) GetRecipientMessageFilter(
	recipientID string,
) *RecipientMessageFilter {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	if filter, ok := sme.recipientFilters[recipientID]; ok {
		return filter
	}

	return &RecipientMessageFilter{
		RecipientID: recipientID,
		CreatedAt:   time.Now(),
	}
}

// BlockSilentMessages blocks silent messages from a specific sender
func (sme *ScheduledMessageEnhancements) BlockSilentMessages(
	recipientID, senderID string,
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	control := sme.recipientControls[recipientID]
	if control == nil {
		control = &RecipientSilentControl{
			RecipientID: recipientID,
			CreatedAt:   time.Now(),
		}
		sme.recipientControls[recipientID] = control
	}

	control.BlockedSilentSenders = append(control.BlockedSilentSenders, senderID)
	control.UpdatedAt = time.Now()

	return nil
}

// GetRecipientSilentControl retrieves silent message controls
func (sme *ScheduledMessageEnhancements) GetRecipientSilentControl(
	recipientID string,
) *RecipientSilentControl {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	if control, ok := sme.recipientControls[recipientID]; ok {
		return control
	}

	return &RecipientSilentControl{
		RecipientID: recipientID,
		CreatedAt:   time.Now(),
	}
}

// CanSendSilentMessage checks if silent message can be sent
func (sme *ScheduledMessageEnhancements) CanSendSilentMessage(
	senderID, recipientID string,
	trustScore int,
) (bool, string) {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	control := sme.recipientControls[recipientID]
	if control != nil {
		// Check if globally disabled
		if control.GlobalDisableSilentMessages {
			return false, "recipient has disabled silent messages"
		}

		// Check if sender is blocked
		for _, blocked := range control.BlockedSilentSenders {
			if blocked == senderID {
				return false, "sender is blocked from sending silent messages"
			}
		}

		// Check if approval required for low-trust senders
		if control.RequireApprovalFromNonTrusted && trustScore < 40 {
			return false, "low-trust senders require recipient approval"
		}
	}

	return true, ""
}

// GetDeliveryFailure retrieves delivery failure information
func (sme *ScheduledMessageEnhancements) GetDeliveryFailure(
	messageID string,
) *DeliveryFailureRecord {
	sme.mu.RLock()
	defer sme.mu.RUnlock()

	if failure, ok := sme.failureRecords[messageID]; ok {
		return failure
	}

	return nil
}

// MarkDeliveryFailureResolved marks a failure as resolved
func (sme *ScheduledMessageEnhancements) MarkDeliveryFailureResolved(
	messageID string,
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	if failure, ok := sme.failureRecords[messageID]; ok {
		failure.Status = "resolved"
		return nil
	}

	return fmt.Errorf("delivery failure not found")
}

// NotifySenderOfFailure marks sender as notified of delivery failure
func (sme *ScheduledMessageEnhancements) NotifySenderOfFailure(
	messageID string,
) error {
	sme.mu.Lock()
	defer sme.mu.Unlock()

	if failure, ok := sme.failureRecords[messageID]; ok {
		failure.SenderNotified = true
		return nil
	}

	return fmt.Errorf("delivery failure not found")
}
