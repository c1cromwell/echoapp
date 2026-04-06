package messaging

import (
	"testing"
	"time"
)

func TestConditionalDeliveryTrigger(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	// Test setting conditional trigger
	err := sme.SetConditionalTrigger("msg-1", TriggerWhenOnline)
	if err != nil {
		t.Errorf("SetConditionalTrigger failed: %v", err)
	}

	// Verify constant values exist
	if TriggerTimeOnly != "time_only" {
		t.Errorf("TriggerTimeOnly constant incorrect")
	}
	if TriggerWhenOnline != "when_online" {
		t.Errorf("TriggerWhenOnline constant incorrect")
	}
	if TriggerAfterLastRead != "after_last_read" {
		t.Errorf("TriggerAfterLastRead constant incorrect")
	}
	if TriggerNoResponseTimeout != "no_response_timeout" {
		t.Errorf("TriggerNoResponseTimeout constant incorrect")
	}
	if TriggerConversationEnded != "conversation_ended" {
		t.Errorf("TriggerConversationEnded constant incorrect")
	}
}

func TestSmartSchedulingSuggestion(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	// Test getting smart scheduling suggestion
	suggestion := sme.GetSmartSchedulingSuggestion("recipient-1", 0)
	if suggestion == nil {
		t.Fatal("expected scheduling suggestion")
	}

	if suggestion.RecipientID != "recipient-1" {
		t.Errorf("expected RecipientID 'recipient-1', got %s", suggestion.RecipientID)
	}

	if suggestion.Confidence < 0.0 || suggestion.Confidence > 1.0 {
		t.Errorf("confidence out of range: %f", suggestion.Confidence)
	}

	if len(suggestion.AlternativeTimes) == 0 {
		t.Error("expected alternative times")
	}

	if suggestion.BusinessHoursStart != 9 || suggestion.BusinessHoursEnd != 17 {
		t.Errorf("unexpected business hours: %d-%d", suggestion.BusinessHoursStart, suggestion.BusinessHoursEnd)
	}

	if suggestion.AverageResponseTime != 4*time.Hour {
		t.Errorf("unexpected average response time: %v", suggestion.AverageResponseTime)
	}
}

func TestDeliveryFailureRecord(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	tests := []struct {
		name        string
		failureType string
		expectRetry bool
	}{
		{"offline", "offline", true},
		{"blocked", "blocked", false},
		{"network", "network", true},
		{"deleted", "account_deleted", false},
		{"key_release", "key_release_failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failure := sme.HandleDeliveryFailure("msg-"+tt.name, tt.failureType, "test failure")

			if failure.MessageID != "msg-"+tt.name {
				t.Errorf("expected MessageID msg-%s, got %s", tt.name, failure.MessageID)
			}

			if failure.FailureType != tt.failureType {
				t.Errorf("expected FailureType %s, got %s", tt.failureType, failure.FailureType)
			}

			if tt.expectRetry && failure.NextRetryTime == nil {
				t.Error("expected retry time but got nil")
			}

			if !tt.expectRetry {
				if failure.Status != "expired" && failure.FailureType != "key_release_failed" {
					t.Errorf("expected expired status, got %s", failure.Status)
				}
			}
		})
	}
}

func TestEditHistory(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	msgID := "msg-123"
	oldContent := []byte("old message")
	newContent := []byte("new message")

	// Record edit
	err := sme.RecordEditHistory(msgID, oldContent, newContent)
	if err != nil {
		t.Errorf("RecordEditHistory failed: %v", err)
	}

	// Get history
	history := sme.GetEditHistory(msgID)
	if len(history) != 1 {
		t.Errorf("expected 1 edit record, got %d", len(history))
	}

	if string(history[0].OldContent) != "old message" {
		t.Errorf("expected old content 'old message', got %s", history[0].OldContent)
	}

	if string(history[0].NewContent) != "new message" {
		t.Errorf("expected new content 'new message', got %s", history[0].NewContent)
	}

	// Record another edit
	err = sme.RecordEditHistory(msgID, newContent, []byte("final message"))
	if err != nil {
		t.Errorf("RecordEditHistory second call failed: %v", err)
	}

	history = sme.GetEditHistory(msgID)
	if len(history) != 2 {
		t.Errorf("expected 2 edit records, got %d", len(history))
	}
}

func TestTimezoneHandling(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	convID := "conv-123"
	participantID := "user-456"
	tzOffset := -300 // EST: UTC-5

	// Set timezone
	err := sme.SetConversationTimezone(convID, participantID, tzOffset)
	if err != nil {
		t.Errorf("SetConversationTimezone failed: %v", err)
	}

	// Get timezone
	offset, err := sme.GetConversationTimezone(convID, participantID)
	if err != nil {
		t.Errorf("GetConversationTimezone failed: %v", err)
	}

	if offset != tzOffset {
		t.Errorf("expected offset %d, got %d", tzOffset, offset)
	}

	// Get non-existent timezone
	_, err = sme.GetConversationTimezone("unknown", "unknown")
	if err == nil {
		t.Error("expected error for non-existent timezone")
	}
}

func TestValidateScheduleInWakingHours(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		hour        int
		tzOffset    int
		expectValid bool
	}{
		{"morning 9 AM", 9, 0, true},
		{"afternoon 3 PM", 15, 0, true},
		{"early 7 AM", 7, 0, true},
		{"late 11 PM", 23, 0, true},
		{"midnight", 0, 0, false},
		{"early morning 6 AM", 6, 0, false},
		{"after midnight 24", 24, 0, false},
		{"morning EST", 14, -300, true}, // 2 PM UTC = 9 AM EST
		{"late EST", 4, -300, true},     // 4 AM UTC = 11 PM EST (valid - still in waking hours)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduledTime := baseTime.Add(time.Duration(tt.hour) * time.Hour)
			valid, msg := sme.ValidateScheduleInWakingHours(scheduledTime, tt.tzOffset)

			if valid != tt.expectValid {
				t.Errorf("expected valid=%v, got %v. Message: %s", tt.expectValid, valid, msg)
			}
		})
	}
}

func TestRecipientMessageFilter(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	recipientID := "user-789"
	filter := &RecipientMessageFilter{
		RecipientID:               recipientID,
		AllowSilentFromTrusted:    true,
		AllowSilentFromMembers:    true,
		AllowSilentFromNewcomers:  false,
		AllowScheduledFromTrusted: true,
		AutoApproveFromTrusted:    true,
	}

	// Set filter
	err := sme.SetRecipientMessageFilter(recipientID, filter)
	if err != nil {
		t.Errorf("SetRecipientMessageFilter failed: %v", err)
	}

	// Get filter
	retrieved := sme.GetRecipientMessageFilter(recipientID)
	if retrieved.RecipientID != recipientID {
		t.Errorf("expected RecipientID %s, got %s", recipientID, retrieved.RecipientID)
	}

	if !retrieved.AllowSilentFromTrusted {
		t.Error("expected AllowSilentFromTrusted=true")
	}

	if retrieved.AllowSilentFromNewcomers {
		t.Error("expected AllowSilentFromNewcomers=false")
	}

	// Get non-existent filter returns default
	newFilter := sme.GetRecipientMessageFilter("unknown-user")
	if newFilter.RecipientID != "unknown-user" {
		t.Errorf("expected new filter with ID unknown-user")
	}
}

func TestRecipientSilentControl(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	recipientID := "user-111"
	senderID1 := "sender-222"
	senderID2 := "sender-333"

	// Block silent messages from sender
	err := sme.BlockSilentMessages(recipientID, senderID1)
	if err != nil {
		t.Errorf("BlockSilentMessages failed: %v", err)
	}

	// Get control
	control := sme.GetRecipientSilentControl(recipientID)
	if len(control.BlockedSilentSenders) != 1 {
		t.Errorf("expected 1 blocked sender, got %d", len(control.BlockedSilentSenders))
	}

	if control.BlockedSilentSenders[0] != senderID1 {
		t.Errorf("expected blocked sender %s, got %s", senderID1, control.BlockedSilentSenders[0])
	}

	// Block another sender
	sme.BlockSilentMessages(recipientID, senderID2)
	control = sme.GetRecipientSilentControl(recipientID)
	if len(control.BlockedSilentSenders) != 2 {
		t.Errorf("expected 2 blocked senders, got %d", len(control.BlockedSilentSenders))
	}
}

func TestCanSendSilentMessage(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	recipientID := "user-444"
	senderID := "sender-555"

	// Initially should allow
	allowed, msg := sme.CanSendSilentMessage(senderID, recipientID, 50)
	if !allowed {
		t.Errorf("expected to allow silent message: %s", msg)
	}

	// Block sender
	sme.BlockSilentMessages(recipientID, senderID)
	allowed, msg = sme.CanSendSilentMessage(senderID, recipientID, 50)
	if allowed {
		t.Error("expected to block silent message after blocking sender")
	}

	// Test with low trust score requiring approval
	newRecipientID := "user-666"
	newSenderID := "sender-777"

	control := &RecipientSilentControl{
		RecipientID:                   newRecipientID,
		RequireApprovalFromNonTrusted: true,
	}
	sme.recipientControls[newRecipientID] = control

	allowed, msg = sme.CanSendSilentMessage(newSenderID, newRecipientID, 30)
	if allowed {
		t.Error("expected to require approval for low-trust sender")
	}

	// High trust should bypass approval
	allowed, msg = sme.CanSendSilentMessage(newSenderID, newRecipientID, 50)
	if !allowed {
		t.Errorf("expected high-trust sender to bypass approval: %s", msg)
	}

	// Globally disabled
	anotherRecipientID := "user-888"
	control = &RecipientSilentControl{
		RecipientID:                 anotherRecipientID,
		GlobalDisableSilentMessages: true,
	}
	sme.recipientControls[anotherRecipientID] = control

	allowed, msg = sme.CanSendSilentMessage("anyone", anotherRecipientID, 100)
	if allowed {
		t.Error("expected to block when globally disabled")
	}
}

func TestDeliveryFailureFunctions(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	msgID := "msg-failure-123"
	failure := sme.HandleDeliveryFailure(msgID, "offline", "user offline")

	if failure.Status != "queued" {
		t.Errorf("expected status 'queued', got %s", failure.Status)
	}

	// Get failure
	retrieved := sme.GetDeliveryFailure(msgID)
	if retrieved == nil {
		t.Fatal("expected failure record")
	}

	// Mark as resolved
	err := sme.MarkDeliveryFailureResolved(msgID)
	if err != nil {
		t.Errorf("MarkDeliveryFailureResolved failed: %v", err)
	}

	retrieved = sme.GetDeliveryFailure(msgID)
	if retrieved.Status != "resolved" {
		t.Errorf("expected status 'resolved', got %s", retrieved.Status)
	}

	// Notify sender
	msgID2 := "msg-failure-456"
	sme.HandleDeliveryFailure(msgID2, "blocked", "user blocked")
	err = sme.NotifySenderOfFailure(msgID2)
	if err != nil {
		t.Errorf("NotifySenderOfFailure failed: %v", err)
	}

	retrieved = sme.GetDeliveryFailure(msgID2)
	if !retrieved.SenderNotified {
		t.Error("expected SenderNotified=true")
	}

	// Mark non-existent as resolved should fail
	err = sme.MarkDeliveryFailureResolved("non-existent")
	if err == nil {
		t.Error("expected error for non-existent failure")
	}
}

func TestConcurrentEnhancements(t *testing.T) {
	sme := NewScheduledMessageEnhancements()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(index int) {
			msgID := "msg-" + string(rune(index))
			sme.HandleDeliveryFailure(msgID, "offline", "test")
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = sme.GetSmartSchedulingSuggestion("recipient", 0)
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}
