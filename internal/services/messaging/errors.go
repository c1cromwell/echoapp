package messaging

import "errors"

var (
	ErrInvalidParticipants = errors.New("at least 2 participants required")
	ErrInvalidSender       = errors.New("sender ID required")
	ErrMessageNotFound     = errors.New("message not found")
	ErrConvNotFound        = errors.New("conversation not found")

	// Silent message errors
	ErrSilentMessagesDisabled  = errors.New("silent messages disabled for this conversation")
	ErrSilentBlockedByRecipient = errors.New("recipient has blocked silent messages from this sender")
	ErrSilentRateLimitExceeded = errors.New("silent message rate limit exceeded")

	// Scheduled message errors
	ErrScheduledTimeInPast       = errors.New("scheduled time must be in the future")
	ErrScheduledTimeTooFar       = errors.New("scheduled time exceeds maximum allowed")
	ErrScheduledLimitExceeded    = errors.New("scheduled message limit exceeded")
	ErrScheduledPerRecipientLimit = errors.New("per-recipient scheduled message limit exceeded")
	ErrScheduledNotFound         = errors.New("scheduled message not found")
	ErrScheduledAlreadyDelivered = errors.New("scheduled message already delivered")
	ErrScheduledEditTooLate      = errors.New("cannot edit within 5 minutes of delivery")
	ErrScheduledNotOwner         = errors.New("only the sender can modify a scheduled message")

	// Rate limiting errors
	ErrDailyLimitExceeded        = errors.New("daily message limit exceeded for trust level")
	ErrPerRecipientLimitExceeded = errors.New("per-recipient daily limit exceeded")

	// Abuse errors
	ErrFeatureSuspended = errors.New("feature suspended due to abuse reports")
)
