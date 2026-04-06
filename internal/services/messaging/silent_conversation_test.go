package messaging

import (
	"strings"
	"testing"
	"time"
)

func TestCreateSilentConversation(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2", "user-3"}
	topic := "confidential discussion"

	mode := SilentConversationMode{
		MaxDaysVisible:      7,
		RequireExitApproval: false,
		HideFromMainList:    true,
		EncryptionRequired:  true,
		TrustLevelMinimum:   string(TrustMember),
	}

	conv, err := scs.CreateSilentConversation(creatorID, participantIDs, topic, mode)
	if err != nil {
		t.Errorf("CreateSilentConversation failed: %v", err)
	}

	if conv.ID == "" {
		t.Error("expected conversation ID")
	}

	if conv.CreatedBy != creatorID {
		t.Errorf("expected createdBy %s, got %s", creatorID, conv.CreatedBy)
	}

	if conv.Topic != topic {
		t.Errorf("expected topic %s, got %s", topic, conv.Topic)
	}

	if len(conv.ParticipantIDs) != len(participantIDs) {
		t.Errorf("expected %d participants, got %d", len(participantIDs), len(conv.ParticipantIDs))
	}

	if !conv.Mode.HideFromMainList {
		t.Error("expected HideFromMainList=true")
	}

	if !conv.Mode.EncryptionRequired {
		t.Error("expected EncryptionRequired=true")
	}
}

func TestCreateSilentConversationValidation(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	mode := SilentConversationMode{}

	// Too few participants
	_, err := scs.CreateSilentConversation(creatorID, []string{"user-1"}, "topic", mode)
	if err == nil {
		t.Error("expected error for < 2 participants")
	}

	// Empty participants
	_, err = scs.CreateSilentConversation(creatorID, []string{}, "topic", mode)
	if err == nil {
		t.Error("expected error for empty participants")
	}
}

func TestGetSilentConversation(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	created, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	// Retrieve conversation
	conv, err := scs.GetSilentConversation(created.ID)
	if err != nil {
		t.Errorf("GetSilentConversation failed: %v", err)
	}

	if conv.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, conv.ID)
	}

	// Non-existent conversation
	_, err = scs.GetSilentConversation("non-existent")
	if err == nil {
		t.Error("expected error for non-existent conversation")
	}
}

func TestRequestAndApproveJoin(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	newParticipantID := "user-new"
	err := scs.RequestJoinSilentConversation(newParticipantID, conv.ID)
	if err != nil {
		t.Errorf("RequestJoinSilentConversation failed: %v", err)
	}

	// Approve join
	err = scs.ApproveMemberJoin(creatorID, newParticipantID, conv.ID)
	if err != nil {
		t.Errorf("ApproveMemberJoin failed: %v", err)
	}

	// Check participant status
	status, err := scs.GetParticipantStatus(conv.ID, newParticipantID)
	if err != nil {
		t.Errorf("GetParticipantStatus failed: %v", err)
	}

	if !status.ApprovedToJoin {
		t.Error("expected ApprovedToJoin=true")
	}

	if status.ApprovedBy != creatorID {
		t.Errorf("expected ApprovedBy=%s, got %s", creatorID, status.ApprovedBy)
	}
}

func TestExitSilentConversation(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2", "user-3"}
	mode := SilentConversationMode{
		RequireExitApproval: false,
	}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	// Exit without approval required
	err := scs.ExitSilentConversation("user-1", conv.ID)
	if err != nil {
		t.Errorf("ExitSilentConversation failed: %v", err)
	}

	// Verify participant count
	activeCount := scs.CountActiveParticipants(conv.ID)
	if activeCount != 2 {
		t.Errorf("expected 2 active participants, got %d", activeCount)
	}
}

func TestExitWithApproval(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{
		RequireExitApproval: true,
	}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	exitUser := "user-1"
	err := scs.ExitSilentConversation(exitUser, conv.ID)
	if err != nil {
		t.Errorf("ExitSilentConversation failed: %v", err)
	}

	// Check exit approval pending
	status, _ := scs.GetParticipantStatus(conv.ID, exitUser)
	if !status.ExitApprovalPending {
		t.Error("expected ExitApprovalPending=true")
	}

	// Approve exit
	err = scs.ApproveExit(creatorID, exitUser, conv.ID)
	if err != nil {
		t.Errorf("ApproveExit failed: %v", err)
	}

	status, _ = scs.GetParticipantStatus(conv.ID, exitUser)
	if status.ExitApprovalPending {
		t.Error("expected ExitApprovalPending=false after approval")
	}

	if !status.IsActive == false {
		t.Error("expected IsActive=false after approved exit")
	}
}

func TestRecallMessage(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	messageID := "msg-123"
	reason := "accidental send"

	// Manually add message to conversation (normally done by messaging service)
	scs.conversationMessages[conv.ID] = []string{messageID}

	// Recall message
	err := scs.RecallMessage(messageID, conv.ID, creatorID, reason)
	if err != nil {
		t.Errorf("RecallMessage failed: %v", err)
	}

	// Verify recall log
	log := scs.GetMessageRecallLog(messageID)
	if log == nil {
		t.Fatal("expected recall log")
	}

	if log.Reason != reason {
		t.Errorf("expected reason %s, got %s", reason, log.Reason)
	}

	if log.RecalledBy != creatorID {
		t.Errorf("expected recalled by %s, got %s", creatorID, log.RecalledBy)
	}

	// Verify message removed from conversation
	messages := scs.conversationMessages[conv.ID]
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

func TestDisappearingMessageTimer(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	messageID := "msg-disappear"
	scs.conversationMessages[conv.ID] = []string{messageID}

	// Set disappearing timer for 1 hour
	deleteAfter := 1 * time.Hour
	err := scs.SetMessageDisappearingTimer(messageID, conv.ID, deleteAfter)
	if err != nil {
		t.Errorf("SetMessageDisappearingTimer failed: %v", err)
	}

	// Cleanup should not delete (timer is in future)
	deleted := scs.CleanupExpiredMessages()
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}

	// Manually set past timer for testing
	scs.disappearingTimers[messageID] = time.Now().Add(-1 * time.Hour)

	// Cleanup should delete
	deleted = scs.CleanupExpiredMessages()
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	// Message should be removed from conversation
	messages := scs.conversationMessages[conv.ID]
	if len(messages) != 0 {
		t.Errorf("expected 0 messages after cleanup, got %d", len(messages))
	}
}

func TestEndSilentConversation(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	reason := "topic concluded"
	err := scs.EndSilentConversation(conv.ID, reason)
	if err != nil {
		t.Errorf("EndSilentConversation failed: %v", err)
	}

	// Verify conversation ended
	conv, _ = scs.GetSilentConversation(conv.ID)
	if !conv.ConversationEnded {
		t.Error("expected ConversationEnded=true")
	}

	if conv.EndReason != reason {
		t.Errorf("expected reason %s, got %s", reason, conv.EndReason)
	}

	if conv.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

func TestCountActiveParticipants(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2", "user-3"}
	mode := SilentConversationMode{
		RequireExitApproval: false,
	}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	// Initial count
	count := scs.CountActiveParticipants(conv.ID)
	if count != 3 {
		t.Errorf("expected 3 active participants, got %d", count)
	}

	// Exit one participant
	scs.ExitSilentConversation("user-1", conv.ID)

	count = scs.CountActiveParticipants(conv.ID)
	if count != 2 {
		t.Errorf("expected 2 active participants, got %d", count)
	}

	// Exit another
	scs.ExitSilentConversation("user-2", conv.ID)

	count = scs.CountActiveParticipants(conv.ID)
	if count != 1 {
		t.Errorf("expected 1 active participant, got %d", count)
	}
}

func TestListPendingExitApprovals(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2", "user-3"}
	mode := SilentConversationMode{
		RequireExitApproval: true,
	}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	// Request exits
	scs.ExitSilentConversation("user-1", conv.ID)
	scs.ExitSilentConversation("user-2", conv.ID)

	// Check that exit approvals are pending (they were requested with exits)
	status1, _ := scs.GetParticipantStatus(conv.ID, "user-1")
	status2, _ := scs.GetParticipantStatus(conv.ID, "user-2")

	if !status1.ExitApprovalPending {
		t.Error("expected user-1 exit approval pending")
	}
	if !status2.ExitApprovalPending {
		t.Error("expected user-2 exit approval pending")
	}

	// Approve one exit
	scs.ApproveExit(creatorID, "user-1", conv.ID)

	status1, _ = scs.GetParticipantStatus(conv.ID, "user-1")
	if status1.ExitApprovalPending {
		t.Error("expected user-1 exit approval to be resolved")
	}
}

func TestRecoverConversationHistory(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	// Try to recover without preservation
	_, err := scs.RecoverConversationHistory(conv.ID)
	if err == nil || !strings.Contains(err.Error(), "not preserved") {
		t.Error("expected error about history not preserved")
	}

	// Enable history preservation
	conv.HistoryPreserved = true

	// Add some messages
	messageIDs := []string{"msg-1", "msg-2", "msg-3"}
	scs.conversationMessages[conv.ID] = messageIDs

	// Recover history
	messages, err := scs.RecoverConversationHistory(conv.ID)
	if err != nil {
		t.Errorf("RecoverConversationHistory failed: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}

	for i, msgID := range messages {
		if msgID != messageIDs[i] {
			t.Errorf("expected message %s, got %s", messageIDs[i], msgID)
		}
	}
}

func TestGetParticipantStatus(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	// Get status
	status, err := scs.GetParticipantStatus(conv.ID, "user-1")
	if err != nil {
		t.Errorf("GetParticipantStatus failed: %v", err)
	}

	if status.ParticipantID != "user-1" {
		t.Errorf("expected ParticipantID user-1, got %s", status.ParticipantID)
	}

	if status.ConversationID != conv.ID {
		t.Errorf("expected ConversationID %s, got %s", conv.ID, status.ConversationID)
	}

	if !status.IsActive {
		t.Error("expected IsActive=true")
	}

	// Get non-existent status
	_, err = scs.GetParticipantStatus(conv.ID, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent participant")
	}
}

func TestSilentConversationModeConfiguration(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}

	tests := []struct {
		name string
		mode SilentConversationMode
	}{
		{
			name: "minimal",
			mode: SilentConversationMode{},
		},
		{
			name: "auto_delete_after_read",
			mode: SilentConversationMode{
				AutoDeleteAfterRead: true,
				MaxDaysVisible:      0,
				HideFromMainList:    true,
				EncryptionRequired:  true,
				TrustLevelMinimum:   string(TrustTrusted),
			},
		},
		{
			name: "with_exit_approval",
			mode: SilentConversationMode{
				RequireExitApproval:     true,
				NotifyOnParticipantExit: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv, err := scs.CreateSilentConversation(creatorID, participantIDs, "topic", tt.mode)
			if err != nil {
				t.Errorf("CreateSilentConversation failed: %v", err)
			}

			if conv.Mode.MaxDaysVisible != tt.mode.MaxDaysVisible {
				t.Errorf("expected MaxDaysVisible %d, got %d", tt.mode.MaxDaysVisible, conv.Mode.MaxDaysVisible)
			}

			if conv.Mode.RequireExitApproval != tt.mode.RequireExitApproval {
				t.Errorf("expected RequireExitApproval %v, got %v", tt.mode.RequireExitApproval, conv.Mode.RequireExitApproval)
			}

			if conv.Mode.EncryptionRequired != tt.mode.EncryptionRequired {
				t.Errorf("expected EncryptionRequired %v, got %v", tt.mode.EncryptionRequired, conv.Mode.EncryptionRequired)
			}
		})
	}
}

func TestConcurrentSilentConversationOperations(t *testing.T) {
	scs := NewSilentConversationService()

	creatorID := "user-creator"
	participantIDs := []string{"user-1", "user-2"}
	mode := SilentConversationMode{}

	conv, _ := scs.CreateSilentConversation(creatorID, participantIDs, "topic", mode)

	done := make(chan bool)

	// Multiple message recalls
	for i := 0; i < 10; i++ {
		go func(index int) {
			msgID := conv.ID + "-msg-" + string(rune(48+index))
			// Manually add message to conversation using proper locking
			scs.mu.Lock()
			scs.conversationMessages[conv.ID] = append(scs.conversationMessages[conv.ID], msgID)
			scs.mu.Unlock()

			scs.RecallMessage(msgID, conv.ID, creatorID, "test")
			done <- true
		}(i)
	}

	// Multiple status checks
	for i := 0; i < 10; i++ {
		go func() {
			_ = scs.CountActiveParticipants(conv.ID)
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}
