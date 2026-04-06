package messaging

import (
	"sync"
	"time"
)

// TrustLevel mirrors auth.TrustScoreLevel for decoupling
type TrustLevel string

const (
	TrustUnverified TrustLevel = "unverified"
	TrustNewcomer   TrustLevel = "newcomer"
	TrustMember     TrustLevel = "member"
	TrustTrusted    TrustLevel = "trusted"
	TrustVerified   TrustLevel = "verified"
)

// TrustLimits defines messaging limits for a trust level
type TrustLimits struct {
	SilentMessagesPerDay    int
	ScheduledMessagesPerDay int
	MaxScheduleAheadDays    int
	MaxPerRecipientPerDay   int
	MaxTotalScheduled       int
}

// DefaultTrustLimits returns the trust-based rate limits from the blueprint
var DefaultTrustLimits = map[TrustLevel]TrustLimits{
	TrustUnverified: {
		SilentMessagesPerDay:    5,
		ScheduledMessagesPerDay: 3,
		MaxScheduleAheadDays:    1,
		MaxPerRecipientPerDay:   1,
		MaxTotalScheduled:       3,
	},
	TrustNewcomer: {
		SilentMessagesPerDay:    20,
		ScheduledMessagesPerDay: 10,
		MaxScheduleAheadDays:    7,
		MaxPerRecipientPerDay:   3,
		MaxTotalScheduled:       10,
	},
	TrustMember: {
		SilentMessagesPerDay:    50,
		ScheduledMessagesPerDay: 25,
		MaxScheduleAheadDays:    14,
		MaxPerRecipientPerDay:   5,
		MaxTotalScheduled:       25,
	},
	TrustTrusted: {
		SilentMessagesPerDay:    100,
		ScheduledMessagesPerDay: 50,
		MaxScheduleAheadDays:    30,
		MaxPerRecipientPerDay:   10,
		MaxTotalScheduled:       50,
	},
	TrustVerified: {
		SilentMessagesPerDay:    -1, // unlimited
		ScheduledMessagesPerDay: -1, // unlimited
		MaxScheduleAheadDays:    30,
		MaxPerRecipientPerDay:   20,
		MaxTotalScheduled:       100,
	},
}

// PerRecipientDailyLimit is the hard cap regardless of trust level
const PerRecipientDailyLimit = 10

// usageRecord tracks daily usage for a sender
type usageRecord struct {
	SilentCount    int
	ScheduledCount int
	// recipientID -> count
	PerRecipient map[string]int
	Date         string // YYYY-MM-DD
}

// RateLimiter enforces trust-based rate limits on messaging features
type RateLimiter struct {
	mu     sync.Mutex
	usage  map[string]*usageRecord // senderID -> daily usage
	limits map[TrustLevel]TrustLimits
}

// NewRateLimiter creates a rate limiter with default trust limits
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		usage:  make(map[string]*usageRecord),
		limits: DefaultTrustLimits,
	}
}

// NewRateLimiterWithLimits creates a rate limiter with custom limits
func NewRateLimiterWithLimits(limits map[TrustLevel]TrustLimits) *RateLimiter {
	return &RateLimiter{
		usage:  make(map[string]*usageRecord),
		limits: limits,
	}
}

func (r *RateLimiter) getOrCreateUsage(senderID string) *usageRecord {
	today := time.Now().Format("2006-01-02")
	record, ok := r.usage[senderID]
	if !ok || record.Date != today {
		record = &usageRecord{
			PerRecipient: make(map[string]int),
			Date:         today,
		}
		r.usage[senderID] = record
	}
	return record
}

// CheckSilentLimit verifies if a sender can send a silent message
func (r *RateLimiter) CheckSilentLimit(senderID, recipientID string, level TrustLevel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	limits, ok := r.limits[level]
	if !ok {
		limits = r.limits[TrustUnverified]
	}

	record := r.getOrCreateUsage(senderID)

	// Check daily silent limit (skip if unlimited)
	if limits.SilentMessagesPerDay >= 0 && record.SilentCount >= limits.SilentMessagesPerDay {
		return ErrSilentRateLimitExceeded
	}

	// Check per-recipient hard cap
	if record.PerRecipient[recipientID] >= PerRecipientDailyLimit {
		return ErrPerRecipientLimitExceeded
	}

	return nil
}

// RecordSilent records a silent message send
func (r *RateLimiter) RecordSilent(senderID, recipientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := r.getOrCreateUsage(senderID)
	record.SilentCount++
	record.PerRecipient[recipientID]++
}

// CheckScheduledLimit verifies if a sender can schedule a message
func (r *RateLimiter) CheckScheduledLimit(senderID string, level TrustLevel, scheduledAt time.Time, totalPending int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	limits, ok := r.limits[level]
	if !ok {
		limits = r.limits[TrustUnverified]
	}

	record := r.getOrCreateUsage(senderID)

	// Check daily scheduled limit (skip if unlimited)
	if limits.ScheduledMessagesPerDay >= 0 && record.ScheduledCount >= limits.ScheduledMessagesPerDay {
		return ErrScheduledLimitExceeded
	}

	// Check max schedule ahead
	maxAhead := time.Duration(limits.MaxScheduleAheadDays) * 24 * time.Hour
	if time.Until(scheduledAt) > maxAhead {
		return ErrScheduledTimeTooFar
	}

	// Check total pending limit
	if totalPending >= limits.MaxTotalScheduled {
		return ErrScheduledLimitExceeded
	}

	return nil
}

// RecordScheduled records a scheduled message creation
func (r *RateLimiter) RecordScheduled(senderID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := r.getOrCreateUsage(senderID)
	record.ScheduledCount++
}

// GetLimits returns the limits for a trust level
func (r *RateLimiter) GetLimits(level TrustLevel) TrustLimits {
	if limits, ok := r.limits[level]; ok {
		return limits
	}
	return r.limits[TrustUnverified]
}

// GetUsage returns the current daily usage for a sender
func (r *RateLimiter) GetUsage(senderID string) (silent, scheduled int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record := r.getOrCreateUsage(senderID)
	return record.SilentCount, record.ScheduledCount
}

// GetTrustLevelFromScore maps a numeric trust score to a TrustLevel
func GetTrustLevelFromScore(score int) TrustLevel {
	switch {
	case score >= 80:
		return TrustVerified
	case score >= 60:
		return TrustTrusted
	case score >= 40:
		return TrustMember
	case score >= 20:
		return TrustNewcomer
	default:
		return TrustUnverified
	}
}
