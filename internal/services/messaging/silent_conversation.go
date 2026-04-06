package messaging

import (
	"fmt"
	"sync"
	"time"
)

// SilentConversationMode defines the silent settings for a conversation
type SilentConversationMode struct {
	MaxDaysVisible          int // How many days to keep messages visible (0 = immediate auto-delete)
	RequireExitApproval     bool
	NotifyOnParticipantExit bool
	AutoDeleteAfterRead     bool
	HideFromMainList        bool
	EncryptionRequired      bool
	TrustLevelMinimum       string // TrustLevel from rate limiter: "unverified", "newcomer", "member", "trusted", "verified"
}

// SilentConversation represents a conversation with silent/ephemeral message properties
type SilentConversation struct {
	ID                string
	ParticipantIDs    []string
	Mode              SilentConversationMode
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         string
	Topic             string
	AccessToken       string
	LastMessageTime   *time.Time
	HistoryPreserved  bool // Whether to keep history after conversation ends
	ConversationEnded bool
	EndedAt           *time.Time
	EndReason         string
}

// ParticipantStatus tracks a participant's status in a silent conversation
type ParticipantStatus struct {
	ParticipantID       string
	ConversationID      string
	JoinedAt            time.Time
	LastActivityTime    *time.Time
	IsActive            bool
	TrustScore          int
	MessageCount        int
	ApprovedToJoin      bool
	ApprovedBy          string
	ApprovedAt          *time.Time
	ExitApprovalPending bool
	ExitRequestedAt     *time.Time
	ExitApprovedBy      string
	ExitApprovedAt      *time.Time
}

// MessageRecallLog tracks message recalls/deletions in silent conversations
type MessageRecallLog struct {
	MessageID           string
	ConversationID      string
	RecalledBy          string
	RecallTime          time.Time
	Reason              string
	OriginalSender      string
	OriginalTime        time.Time
	ReceiverIDsAffected []string
}

// SilentConversationService manages silent conversation operations
type SilentConversationService struct {
	mu                   sync.RWMutex
	conversations        map[string]*SilentConversation
	participantStatus    map[string]*ParticipantStatus // key: conversationID:participantID
	messageRecallLog     map[string]*MessageRecallLog  // key: messageID
	conversationMessages map[string][]string           // key: conversationID, value: messageIDs
	exitApprovalRequests map[string][]*ParticipantStatus
	disappearingTimers   map[string]time.Time // key: messageID, value: when to delete
}

// NewSilentConversationService creates a new service instance
func NewSilentConversationService() *SilentConversationService {
	return &SilentConversationService{
		conversations:        make(map[string]*SilentConversation),
		participantStatus:    make(map[string]*ParticipantStatus),
		messageRecallLog:     make(map[string]*MessageRecallLog),
		conversationMessages: make(map[string][]string),
		exitApprovalRequests: make(map[string][]*ParticipantStatus),
		disappearingTimers:   make(map[string]time.Time),
	}
}

// CreateSilentConversation creates a new silent conversation
func (scs *SilentConversationService) CreateSilentConversation(
	creatorID string,
	participantIDs []string,
	topic string,
	mode SilentConversationMode,
) (*SilentConversation, error) {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	if len(participantIDs) < 2 {
		return nil, fmt.Errorf("silent conversation requires at least 2 participants")
	}

	// Generate conversation ID
	id := fmt.Sprintf("silent-conv-%d", time.Now().UnixNano())

	conv := &SilentConversation{
		ID:             id,
		ParticipantIDs: participantIDs,
		Mode:           mode,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      creatorID,
		Topic:          topic,
		AccessToken:    fmt.Sprintf("token-%d", time.Now().UnixNano()),
	}

	scs.conversations[id] = conv

	// Initialize participant status
	for _, participantID := range participantIDs {
		status := &ParticipantStatus{
			ParticipantID:  participantID,
			ConversationID: id,
			JoinedAt:       time.Now(),
			IsActive:       true,
		}

		key := fmt.Sprintf("%s:%s", id, participantID)
		scs.participantStatus[key] = status
	}

	return conv, nil
}

// RequestJoinSilentConversation submits a request to join a conversation
func (scs *SilentConversationService) RequestJoinSilentConversation(
	participantID string,
	conversationID string,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	_, exists := scs.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found")
	}

	// Create participant status
	status := &ParticipantStatus{
		ParticipantID:       participantID,
		ConversationID:      conversationID,
		JoinedAt:            time.Now(),
		ApprovedToJoin:      false,
		ExitApprovalPending: false,
	}

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	scs.participantStatus[key] = status

	// Add to approval requests
	scs.exitApprovalRequests[conversationID] = append(scs.exitApprovalRequests[conversationID], status)

	return nil
}

// ApproveMemberJoin approves a member to join the silent conversation
func (scs *SilentConversationService) ApproveMemberJoin(
	approverID string,
	participantID string,
	conversationID string,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	status, exists := scs.participantStatus[key]
	if !exists {
		return fmt.Errorf("participant status not found")
	}

	status.ApprovedToJoin = true
	status.ApprovedBy = approverID
	status.ApprovedAt = &time.Time{}
	*status.ApprovedAt = time.Now()

	return nil
}

// ExitSilentConversation removes a participant from the conversation
func (scs *SilentConversationService) ExitSilentConversation(
	participantID string,
	conversationID string,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	conv, exists := scs.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found")
	}

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	status, exists := scs.participantStatus[key]
	if !exists {
		return fmt.Errorf("participant not in conversation")
	}

	// Request approval if required
	if conv.Mode.RequireExitApproval {
		status.ExitApprovalPending = true
		status.ExitRequestedAt = &time.Time{}
		*status.ExitRequestedAt = time.Now()
		return nil
	}

	// Remove participant
	status.IsActive = false
	newParticipants := []string{}
	for _, p := range conv.ParticipantIDs {
		if p != participantID {
			newParticipants = append(newParticipants, p)
		}
	}
	conv.ParticipantIDs = newParticipants

	return nil
}

// ApproveExit approves a participant's exit request
func (scs *SilentConversationService) ApproveExit(
	approverID string,
	participantID string,
	conversationID string,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	status, exists := scs.participantStatus[key]
	if !exists {
		return fmt.Errorf("participant status not found")
	}

	status.ExitApprovalPending = false
	status.ExitApprovedBy = approverID
	status.ExitApprovedAt = &time.Time{}
	*status.ExitApprovedAt = time.Now()
	status.IsActive = false

	return nil
}

// RecallMessage recalls/deletes a message from the conversation
func (scs *SilentConversationService) RecallMessage(
	messageID string,
	conversationID string,
	senderID string,
	reason string,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	conv, exists := scs.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found")
	}

	// Only creator of message or conversation creator can recall
	// This is simplified - would normally track message sender
	if senderID != conv.CreatedBy {
		// Allow message creator to recall own messages (would check against message metadata)
	}

	// Record in recall log
	log := &MessageRecallLog{
		MessageID:      messageID,
		ConversationID: conversationID,
		RecalledBy:     senderID,
		RecallTime:     time.Now(),
		Reason:         reason,
		OriginalSender: senderID,
		OriginalTime:   time.Now(),
	}

	scs.messageRecallLog[messageID] = log

	// Remove from message list
	messages := scs.conversationMessages[conversationID]
	newMessages := []string{}
	for _, m := range messages {
		if m != messageID {
			newMessages = append(newMessages, m)
		}
	}
	scs.conversationMessages[conversationID] = newMessages

	// Clear disappearing timer
	delete(scs.disappearingTimers, messageID)

	return nil
}

// SetMessageDisappearingTimer sets when a message should auto-delete
func (scs *SilentConversationService) SetMessageDisappearingTimer(
	messageID string,
	conversationID string,
	deleteAfter time.Duration,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	deleteTime := time.Now().Add(deleteAfter)
	scs.disappearingTimers[messageID] = deleteTime

	return nil
}

// CleanupExpiredMessages removes messages that have expired
func (scs *SilentConversationService) CleanupExpiredMessages() int {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	now := time.Now()
	count := 0

	for msgID, deleteTime := range scs.disappearingTimers {
		if now.After(deleteTime) {
			delete(scs.disappearingTimers, msgID)
			count++

			// Remove from all conversation message lists
			for convID, messages := range scs.conversationMessages {
				newMessages := []string{}
				for _, m := range messages {
					if m != msgID {
						newMessages = append(newMessages, m)
					}
				}
				scs.conversationMessages[convID] = newMessages
			}
		}
	}

	return count
}

// GetSilentConversation retrieves a silent conversation
func (scs *SilentConversationService) GetSilentConversation(
	conversationID string,
) (*SilentConversation, error) {
	scs.mu.RLock()
	defer scs.mu.RUnlock()

	conv, exists := scs.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found")
	}

	return conv, nil
}

// GetParticipantStatus retrieves participant's status in a conversation
func (scs *SilentConversationService) GetParticipantStatus(
	conversationID string,
	participantID string,
) (*ParticipantStatus, error) {
	scs.mu.RLock()
	defer scs.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", conversationID, participantID)
	status, exists := scs.participantStatus[key]
	if !exists {
		return nil, fmt.Errorf("participant status not found")
	}

	return status, nil
}

// EndSilentConversation marks conversation as ended
func (scs *SilentConversationService) EndSilentConversation(
	conversationID string,
	reason string,
) error {
	scs.mu.Lock()
	defer scs.mu.Unlock()

	conv, exists := scs.conversations[conversationID]
	if !exists {
		return fmt.Errorf("conversation not found")
	}

	conv.ConversationEnded = true
	now := time.Now()
	conv.EndedAt = &now
	conv.EndReason = reason

	return nil
}

// RecoverConversationHistory returns conversation history (if preserved)
func (scs *SilentConversationService) RecoverConversationHistory(
	conversationID string,
) ([]string, error) {
	scs.mu.RLock()
	defer scs.mu.RUnlock()

	conv, exists := scs.conversations[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation not found")
	}

	if !conv.HistoryPreserved {
		return nil, fmt.Errorf("history not preserved for this conversation")
	}

	messages := scs.conversationMessages[conversationID]
	return messages, nil
}

// GetMessageRecallLog returns recall information for a message
func (scs *SilentConversationService) GetMessageRecallLog(
	messageID string,
) *MessageRecallLog {
	scs.mu.RLock()
	defer scs.mu.RUnlock()

	return scs.messageRecallLog[messageID]
}

// CountActiveParticipants returns the number of active participants
func (scs *SilentConversationService) CountActiveParticipants(
	conversationID string,
) int {
	scs.mu.RLock()
	defer scs.mu.RUnlock()

	count := 0
	for _, status := range scs.participantStatus {
		if status.ConversationID == conversationID && status.IsActive {
			count++
		}
	}

	return count
}

// ListPendingExitApprovals lists pending exit approval requests
func (scs *SilentConversationService) ListPendingExitApprovals(
	conversationID string,
) []*ParticipantStatus {
	scs.mu.RLock()
	defer scs.mu.RUnlock()

	pending := []*ParticipantStatus{}
	for _, status := range scs.exitApprovalRequests[conversationID] {
		if status.ExitApprovalPending {
			pending = append(pending, status)
		}
	}

	return pending
}
