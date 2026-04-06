# API Reference: Messaging Service Enhancements

## Quick Start Examples

### Scheduled Message Enhancements

```go
package main

import (
    "github.com/thechadcromwell/echoapp/internal/services/messaging"
)

// Initialize the enhancements service
sme := messaging.NewScheduledMessageEnhancements()

// 1. Set conditional delivery
sme.SetConditionalTrigger("msg-123", messaging.TriggerWhenOnline)

// 2. Get smart scheduling suggestions
suggestion := sme.GetSmartSchedulingSuggestion("recipient-id", -300) // EST timezone
// Returns: SmartSchedulingSuggestion with SuggestedTime, Confidence, AlternativeTimes

// 3. Handle delivery failures
failure := sme.HandleDeliveryFailure("msg-123", "offline", "recipient offline")
// Returns: DeliveryFailureRecord with retry scheduling

// 4. Record message edits
sme.RecordEditHistory("msg-123", []byte("old"), []byte("new"))

// 5. Check if silent message can be sent
canSend, reason := sme.CanSendSilentMessage("sender-id", "recipient-id", 45) // trust score
// Returns: bool, string explanation

// 6. Validate timezone for waking hours
valid, msg := sme.ValidateScheduleInWakingHours(scheduledTime, -300)
// Returns: bool, string message
```

### Silent Conversation Service

```go
package main

import (
    "github.com/thechadcromwell/echoapp/internal/services/messaging"
    "time"
)

// Initialize the service
scs := messaging.NewSilentConversationService()

// 1. Create silent conversation
mode := messaging.SilentConversationMode{
    MaxDaysVisible:        7,
    HideFromMainList:      true,
    EncryptionRequired:    true,
    TrustLevelMinimum:     string(messaging.TrustMember),
    RequireExitApproval:   false,
    AutoDeleteAfterRead:   true,
}

conv, err := scs.CreateSilentConversation(
    "creator-id",
    []string{"participant-1", "participant-2"},
    "Confidential Discussion",
    mode,
)

// 2. Manage participants
scs.RequestJoinSilentConversation("new-participant", conv.ID)
scs.ApproveMemberJoin("creator-id", "new-participant", conv.ID)

// 3. Exit conversation  
scs.ExitSilentConversation("participant-1", conv.ID)
// If RequireExitApproval, approval needed:
scs.ApproveExit("creator-id", "participant-1", conv.ID)

// 4. Message recall
scs.RecallMessage("message-id", conv.ID, "sender-id", "accidental send")

// 5. Disappearing messages
scs.SetMessageDisappearingTimer("message-id", conv.ID, 1*time.Hour)
// Later, cleanup expired messages
deletedCount := scs.CleanupExpiredMessages()

// 6. End conversation
scs.EndSilentConversation(conv.ID, "project completed")

// 7. Get conversation info
conv, err := scs.GetSilentConversation(conv.ID)
status, err := scs.GetParticipantStatus(conv.ID, "participant-id")

// 8. History recovery
if conv.HistoryPreserved {
    messages, err := scs.RecoverConversationHistory(conv.ID)
}
```

## Type Reference

### ConditionalDeliveryTrigger
```go
type ConditionalDeliveryTrigger string

const (
    TriggerTimeOnly           = "time_only"           // Pure time-based
    TriggerWhenOnline         = "when_online"         // When recipient active
    TriggerAfterLastRead      = "after_last_read"     // After message read
    TriggerNoResponseTimeout  = "no_response_timeout" // If no response
    TriggerConversationEnded  = "conversation_ended"  // When chat ends
)
```

### SilentConversationMode
```go
type SilentConversationMode struct {
    MaxDaysVisible          int    // Days before auto-delete
    RequireExitApproval     bool
    NotifyOnParticipantExit bool
    AutoDeleteAfterRead     bool
    HideFromMainList        bool
    EncryptionRequired      bool
    TrustLevelMinimum       string // "unverified", "newcomer", "member", "trusted", "verified"
}
```

### DeliveryFailureRecord
```go
type DeliveryFailureRecord struct {
    MessageID        string      // ID of failed message
    FailureType      string      // "offline", "blocked", "network", etc.
    FailureReason    string      // Human-readable reason
    FailureTime      time.Time
    RetryCount       int         // Current retry count
    MaxRetries       int         // Maximum retry attempts
    NextRetryTime    *time.Time  // When to retry
    ExpiryTime       time.Time   // When failure record expires
    SenderNotified   bool        // Whether sender was notified
    Status           string      // "queued", "expired", "notified", "resolved"
}
```

### SmartSchedulingSuggestion
```go
type SmartSchedulingSuggestion struct {
    RecipientID         string          // Target recipient
    SuggestedTime       time.Time       // Recommended delivery time
    Confidence          float64         // 0.0-1.0 confidence score
    Reason              string          // Why this time
    AlternativeTimes    []time.Time     // Other suggested times
    BusinessHoursStart  int             // 9
    BusinessHoursEnd    int             // 17
    TypicalActiveHours  []int           // Hours when active
    LastActivityTime    *time.Time
    AverageResponseTime time.Duration
}
```

### SilentConversation
```go
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
    HistoryPreserved  bool
    ConversationEnded bool
    EndedAt           *time.Time
    EndReason         string
}
```

### ParticipantStatus
```go
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
```

## Method Reference

### ScheduledMessageEnhancements Methods

#### Conditional Delivery
- `SetConditionalTrigger(messageID string, trigger ConditionalDeliveryTrigger) error`
- Returns the trigger type that was set

#### Smart Scheduling
- `GetSmartSchedulingSuggestion(recipientID string, recipientTZOffset int) *SmartSchedulingSuggestion`
- Returns suggestion with recommended time and alternatives

#### Failure Management
- `HandleDeliveryFailure(messageID, failureType, failureReason string) *DeliveryFailureRecord`
- `GetDeliveryFailure(messageID string) *DeliveryFailureRecord`
- `MarkDeliveryFailureResolved(messageID string) error`
- `NotifySenderOfFailure(messageID string) error`

#### Edit History
- `RecordEditHistory(messageID string, oldContent, newContent []byte) error`
- `GetEditHistory(messageID string) []EditRecord`

#### Timezone Management
- `SetConversationTimezone(conversationID, participantID string, timezoneOffset int) error`
- `GetConversationTimezone(conversationID, participantID string) (int, error)`
- `ValidateScheduleInWakingHours(scheduledTime time.Time, recipientTZOffset int) (bool, string)`

#### Message Filtering
- `SetRecipientMessageFilter(recipientID string, filter *RecipientMessageFilter) error`
- `GetRecipientMessageFilter(recipientID string) *RecipientMessageFilter`

#### Silent Controls
- `BlockSilentMessages(recipientID, senderID string) error`
- `CanSendSilentMessage(senderID, recipientID string, trustScore int) (bool, string)`
- `GetRecipientSilentControl(recipientID string) *RecipientSilentControl`

### SilentConversationService Methods

#### Conversation Lifecycle
- `CreateSilentConversation(creatorID string, participantIDs []string, topic string, mode SilentConversationMode) (*SilentConversation, error)`
- `GetSilentConversation(conversationID string) (*SilentConversation, error)`
- `EndSilentConversation(conversationID string, reason string) error`

#### Participant Management
- `RequestJoinSilentConversation(participantID, conversationID string) error`
- `ApproveMemberJoin(approverID, participantID, conversationID string) error`
- `GetParticipantStatus(conversationID, participantID string) (*ParticipantStatus, error)`
- `ExitSilentConversation(participantID, conversationID string) error`
- `ApproveExit(approverID, participantID, conversationID string) error`
- `CountActiveParticipants(conversationID string) int`
- `ListPendingExitApprovals(conversationID string) []*ParticipantStatus`

#### Message Management
- `RecallMessage(messageID, conversationID, senderID, reason string) error`
- `GetMessageRecallLog(messageID string) *MessageRecallLog`
- `SetMessageDisappearingTimer(messageID, conversationID string, deleteAfter time.Duration) error`
- `CleanupExpiredMessages() int` (Returns count of deleted messages)

#### History
- `RecoverConversationHistory(conversationID string) ([]string, error)`

## Error Handling

Both services return errors with descriptive messages:

```go
// Example error handling
if err != nil {
    switch err.Error() {
    case "conversation not found":
        log.Printf("Invalid conversation ID")
    case "participant not found":
        log.Printf("Participant is not in conversation")
    case "low-trust senders require recipient approval":
        log.Printf("Need approval for this sender")
    case "history not preserved for this conversation":
        log.Printf("History was not saved for this conversation")
    default:
        log.Printf("Other error: %v", err)
    }
}
```

## Concurrency Notes

Both services are **thread-safe**:
- Use `sync.RWMutex` for concurrent access protection
- Multiple readers can access simultaneously
- Writes have exclusive access
- Safe for use in goroutines

Example:
```go
go func() {
    scs.ExitSilentConversation("user-1", "conv-id")
}()

go func() {
    count := scs.CountActiveParticipants("conv-id")
}()
// Both can run concurrently safely
```

## Best Practices

1. **Waking Hours**: Always validate schedules with `ValidateScheduleInWakingHours()`
2. **Trust Levels**: Check trust scores before allowing sensitive messages
3. **Cleanup**: Call `CleanupExpiredMessages()` periodically (e.g., hourly cron)
4. **History**: Set `HistoryPreserved = true` only for compliance scenarios
5. **Parallel**: Use goroutines for non-blocking operations
6. **Audit**: Review `MessageRecallLog` for security/compliance audits
