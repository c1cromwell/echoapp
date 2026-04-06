package messaging

import (
	"time"

	"github.com/google/uuid"
)

type MessageType int

const (
	TextMessage MessageType = iota
	VoiceMessage
)

// SilentDuration represents how long silent mode is active
type SilentDuration int

const (
	SilentOff     SilentDuration = 0
	Silent1Hour   SilentDuration = 1
	Silent8Hours  SilentDuration = 8
	Silent24Hours SilentDuration = 24
	Silent7Days   SilentDuration = 168
	SilentAlways  SilentDuration = -1
)

// DeliveryStatus tracks the state of a message delivery
type DeliveryStatus string

const (
	DeliveryPending     DeliveryStatus = "pending"
	DeliveryQueued      DeliveryStatus = "queued"
	DeliveryDelivered   DeliveryStatus = "delivered"
	DeliveryFailed      DeliveryStatus = "failed"
	DeliveryExpired     DeliveryStatus = "expired"
	DeliveryCancelled   DeliveryStatus = "cancelled"
	DeliveryUnreachable DeliveryStatus = "unreachable"
)

type Message struct {
	ID             string
	SenderID       string
	ConvID         string
	Content        []byte
	Type           MessageType
	Time           time.Time
	IsRead         bool
	Silent         bool
	SilentFlags    *SilentFlags
	BlockchainTxID string
	ExpiresAt      *time.Time
	EditedAt       *time.Time
	Deleted        bool
}

// SilentFlags controls notification behavior for a silent message
type SilentFlags struct {
	SuppressPush     bool
	SuppressBadge    bool
	SuppressTyping   bool
	SuppressReceipts bool
	SuppressPreview  bool
}

// DefaultSilentFlags returns flags that suppress all notifications
func DefaultSilentFlags() *SilentFlags {
	return &SilentFlags{
		SuppressPush:     true,
		SuppressBadge:    true,
		SuppressTyping:   true,
		SuppressReceipts: true,
		SuppressPreview:  true,
	}
}

// ConversationSilentSettings holds per-conversation silent mode config
type ConversationSilentSettings struct {
	Enabled   bool
	Duration  SilentDuration
	ExpiresAt *time.Time
}

type Conversation struct {
	ID             string
	Users          []string
	Time           time.Time
	SilentSettings *ConversationSilentSettings
}

type MessagingService struct {
	msgs  map[string]*Message
	convs map[string]*Conversation
}

func NewMessagingService() *MessagingService {
	return &MessagingService{
		msgs:  make(map[string]*Message),
		convs: make(map[string]*Conversation),
	}
}

func (m *MessagingService) CreateConversation(users []string) (*Conversation, error) {
	if len(users) < 2 {
		return nil, ErrInvalidParticipants
	}

	c := &Conversation{
		ID:    uuid.New().String(),
		Users: users,
		Time:  time.Now(),
	}

	m.convs[c.ID] = c
	return c, nil
}

// CreateConv is an alias for backwards compatibility
func (m *MessagingService) CreateConv(users []string) (*Conversation, error) {
	return m.CreateConversation(users)
}

func (m *MessagingService) SendMessage(senderID, convID string, msgType MessageType, content []byte) (*Message, error) {
	if senderID == "" {
		return nil, ErrInvalidSender
	}
	if _, ok := m.convs[convID]; !ok {
		return nil, ErrConvNotFound
	}

	msg := &Message{
		ID:       uuid.New().String(),
		SenderID: senderID,
		ConvID:   convID,
		Content:  content,
		Type:     msgType,
		Time:     time.Now(),
	}

	m.msgs[msg.ID] = msg
	return msg, nil
}

// SendSilentMessage sends a message with silent notification flags
func (m *MessagingService) SendSilentMessage(senderID, convID string, msgType MessageType, content []byte, flags *SilentFlags) (*Message, error) {
	if senderID == "" {
		return nil, ErrInvalidSender
	}
	conv, ok := m.convs[convID]
	if !ok {
		return nil, ErrConvNotFound
	}

	// Check if conversation-level silent mode overrides
	if conv.SilentSettings != nil && conv.SilentSettings.Enabled {
		if conv.SilentSettings.ExpiresAt != nil && time.Now().After(*conv.SilentSettings.ExpiresAt) {
			conv.SilentSettings.Enabled = false
		}
	}

	if flags == nil {
		flags = DefaultSilentFlags()
	}

	msg := &Message{
		ID:          uuid.New().String(),
		SenderID:    senderID,
		ConvID:      convID,
		Content:     content,
		Type:        msgType,
		Time:        time.Now(),
		Silent:      true,
		SilentFlags: flags,
	}

	m.msgs[msg.ID] = msg
	return msg, nil
}

// Send is an alias for backwards compatibility
func (m *MessagingService) Send(senderID, convID string, content []byte) (*Message, error) {
	return m.SendMessage(senderID, convID, TextMessage, content)
}

func (m *MessagingService) GetMessage(id string) (*Message, error) {
	msg, ok := m.msgs[id]
	if !ok {
		return nil, ErrMessageNotFound
	}
	return msg, nil
}

func (m *MessagingService) GetConversation(id string) (*Conversation, error) {
	conv, ok := m.convs[id]
	if !ok {
		return nil, ErrConvNotFound
	}
	return conv, nil
}

// SetConversationSilent enables/disables silent mode for a conversation
func (m *MessagingService) SetConversationSilent(convID string, duration SilentDuration) error {
	conv, ok := m.convs[convID]
	if !ok {
		return ErrConvNotFound
	}

	if duration == SilentOff {
		conv.SilentSettings = nil
		return nil
	}

	settings := &ConversationSilentSettings{
		Enabled:  true,
		Duration: duration,
	}

	if duration != SilentAlways {
		expires := time.Now().Add(time.Duration(duration) * time.Hour)
		settings.ExpiresAt = &expires
	}

	conv.SilentSettings = settings
	return nil
}

// GetConversationMessages returns all messages for a conversation
func (m *MessagingService) GetConversationMessages(convID string) ([]*Message, error) {
	if _, ok := m.convs[convID]; !ok {
		return nil, ErrConvNotFound
	}

	var msgs []*Message
	for _, msg := range m.msgs {
		if msg.ConvID == convID && !msg.Deleted {
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

// GetSilentMessages returns only silent messages for a conversation
func (m *MessagingService) GetSilentMessages(convID string) ([]*Message, error) {
	if _, ok := m.convs[convID]; !ok {
		return nil, ErrConvNotFound
	}

	var msgs []*Message
	for _, msg := range m.msgs {
		if msg.ConvID == convID && msg.Silent && !msg.Deleted {
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}
