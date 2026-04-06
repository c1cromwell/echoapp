package messaging

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// MinEditWindow is the minimum time before delivery when edits are blocked
	MinEditWindow = 5 * time.Minute
	// MaxQueueDays is the default max days a failed delivery stays queued
	MaxQueueDays = 7
	// MaxRetries for delivery attempts
	MaxRetries = 3
)

// ScheduledMessage represents a message queued for future delivery
type ScheduledMessage struct {
	ID              string
	SenderID        string
	ConvID          string
	Content         []byte
	ContentType     MessageType
	ScheduledAt     time.Time
	Timezone        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Status          DeliveryStatus
	DeliveredAt     *time.Time
	RetryCount      int
	FailureReason   string
	Silent          bool
	SilentFlags     *SilentFlags
}

// ScheduledMessageService manages scheduled message operations
type ScheduledMessageService struct {
	mu        sync.RWMutex
	messages  map[string]*ScheduledMessage
	messaging *MessagingService
}

// NewScheduledMessageService creates a new scheduled message service
func NewScheduledMessageService(messaging *MessagingService) *ScheduledMessageService {
	return &ScheduledMessageService{
		messages:  make(map[string]*ScheduledMessage),
		messaging: messaging,
	}
}

// Schedule creates a new scheduled message
func (s *ScheduledMessageService) Schedule(senderID, convID string, content []byte, contentType MessageType, scheduledAt time.Time, timezone string) (*ScheduledMessage, error) {
	if senderID == "" {
		return nil, ErrInvalidSender
	}
	if scheduledAt.Before(time.Now()) {
		return nil, ErrScheduledTimeInPast
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	msg := &ScheduledMessage{
		ID:          uuid.New().String(),
		SenderID:    senderID,
		ConvID:      convID,
		Content:     content,
		ContentType: contentType,
		ScheduledAt: scheduledAt,
		Timezone:    timezone,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      DeliveryPending,
	}

	s.messages[msg.ID] = msg
	return msg, nil
}

// ScheduleSilent creates a scheduled silent message
func (s *ScheduledMessageService) ScheduleSilent(senderID, convID string, content []byte, contentType MessageType, scheduledAt time.Time, timezone string, flags *SilentFlags) (*ScheduledMessage, error) {
	msg, err := s.Schedule(senderID, convID, content, contentType, scheduledAt, timezone)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	msg.Silent = true
	if flags == nil {
		flags = DefaultSilentFlags()
	}
	msg.SilentFlags = flags
	return msg, nil
}

// Get retrieves a scheduled message by ID
func (s *ScheduledMessageService) Get(id string) (*ScheduledMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg, ok := s.messages[id]
	if !ok {
		return nil, ErrScheduledNotFound
	}
	return msg, nil
}

// Edit updates a scheduled message's content
func (s *ScheduledMessageService) Edit(id, senderID string, newContent []byte) (*ScheduledMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := s.messages[id]
	if !ok {
		return nil, ErrScheduledNotFound
	}

	if msg.SenderID != senderID {
		return nil, ErrScheduledNotOwner
	}

	if msg.Status == DeliveryDelivered {
		return nil, ErrScheduledAlreadyDelivered
	}

	if msg.Status == DeliveryCancelled {
		return nil, ErrScheduledAlreadyDelivered
	}

	if time.Until(msg.ScheduledAt) < MinEditWindow {
		return nil, ErrScheduledEditTooLate
	}

	msg.Content = newContent
	msg.UpdatedAt = time.Now()
	return msg, nil
}

// Reschedule changes the delivery time of a scheduled message
func (s *ScheduledMessageService) Reschedule(id, senderID string, newTime time.Time) (*ScheduledMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := s.messages[id]
	if !ok {
		return nil, ErrScheduledNotFound
	}

	if msg.SenderID != senderID {
		return nil, ErrScheduledNotOwner
	}

	if msg.Status == DeliveryDelivered {
		return nil, ErrScheduledAlreadyDelivered
	}

	if newTime.Before(time.Now()) {
		return nil, ErrScheduledTimeInPast
	}

	msg.ScheduledAt = newTime
	msg.UpdatedAt = time.Now()
	msg.Status = DeliveryPending
	return msg, nil
}

// Cancel cancels a scheduled message
func (s *ScheduledMessageService) Cancel(id, senderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := s.messages[id]
	if !ok {
		return ErrScheduledNotFound
	}

	if msg.SenderID != senderID {
		return ErrScheduledNotOwner
	}

	if msg.Status == DeliveryDelivered {
		return ErrScheduledAlreadyDelivered
	}

	msg.Status = DeliveryCancelled
	msg.UpdatedAt = time.Now()
	return nil
}

// SendNow delivers a scheduled message immediately
func (s *ScheduledMessageService) SendNow(id, senderID string) (*Message, error) {
	s.mu.Lock()
	msg, ok := s.messages[id]
	if !ok {
		s.mu.Unlock()
		return nil, ErrScheduledNotFound
	}

	if msg.SenderID != senderID {
		s.mu.Unlock()
		return nil, ErrScheduledNotOwner
	}

	if msg.Status == DeliveryDelivered {
		s.mu.Unlock()
		return nil, ErrScheduledAlreadyDelivered
	}
	s.mu.Unlock()

	return s.deliver(msg)
}

// deliver executes the actual message delivery
func (s *ScheduledMessageService) deliver(scheduled *ScheduledMessage) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var delivered *Message
	var err error

	if scheduled.Silent {
		delivered, err = s.messaging.SendSilentMessage(
			scheduled.SenderID, scheduled.ConvID,
			scheduled.ContentType, scheduled.Content,
			scheduled.SilentFlags,
		)
	} else {
		delivered, err = s.messaging.SendMessage(
			scheduled.SenderID, scheduled.ConvID,
			scheduled.ContentType, scheduled.Content,
		)
	}

	if err != nil {
		scheduled.RetryCount++
		scheduled.FailureReason = err.Error()
		if scheduled.RetryCount >= MaxRetries {
			scheduled.Status = DeliveryFailed
		} else {
			scheduled.Status = DeliveryQueued
		}
		scheduled.UpdatedAt = time.Now()
		return nil, err
	}

	now := time.Now()
	scheduled.Status = DeliveryDelivered
	scheduled.DeliveredAt = &now
	scheduled.UpdatedAt = now
	return delivered, nil
}

// ProcessDue finds and delivers all messages whose scheduled time has passed.
// Returns delivered messages and any errors encountered.
func (s *ScheduledMessageService) ProcessDue() ([]*Message, []error) {
	s.mu.RLock()
	var due []*ScheduledMessage
	now := time.Now()
	for _, msg := range s.messages {
		if msg.Status == DeliveryPending && !msg.ScheduledAt.After(now) {
			due = append(due, msg)
		}
	}
	s.mu.RUnlock()

	var delivered []*Message
	var errs []error
	for _, scheduled := range due {
		msg, err := s.deliver(scheduled)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		delivered = append(delivered, msg)
	}
	return delivered, errs
}

// GetPending returns all pending scheduled messages for a sender
func (s *ScheduledMessageService) GetPending(senderID string) []*ScheduledMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []*ScheduledMessage
	for _, msg := range s.messages {
		if msg.SenderID == senderID && msg.Status == DeliveryPending {
			pending = append(pending, msg)
		}
	}
	return pending
}

// GetByConversation returns all scheduled messages for a conversation by a sender
func (s *ScheduledMessageService) GetByConversation(senderID, convID string) []*ScheduledMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var msgs []*ScheduledMessage
	for _, msg := range s.messages {
		if msg.SenderID == senderID && msg.ConvID == convID {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

// CountPending returns the total pending scheduled messages for a sender
func (s *ScheduledMessageService) CountPending(senderID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, msg := range s.messages {
		if msg.SenderID == senderID && msg.Status == DeliveryPending {
			count++
		}
	}
	return count
}

// CountPendingForRecipient counts pending messages from a sender to a specific conversation
func (s *ScheduledMessageService) CountPendingForRecipient(senderID, convID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, msg := range s.messages {
		if msg.SenderID == senderID && msg.ConvID == convID && msg.Status == DeliveryPending {
			count++
		}
	}
	return count
}
