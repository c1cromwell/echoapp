// Package rewards provides the reward service layer for the Echo backend.
package rewards

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
	internalrewards "github.com/thechadcromwell/echoapp/internal/rewards"
)

const (
	InitialRatePerMessage = int64(10_000_000) // 0.1 ECHO per message when no activity is recorded
)

// TrustMultiplier maps trust tiers to reward multipliers.
var TrustMultiplier = map[int]float64{
	1: 0.0,
	2: 0.5,
	3: 1.0,
	4: 1.5,
	5: 2.0,
}

var (
	ErrInsufficientTier  = errors.New("trust tier too low for rewards")
	ErrNoPendingRewards  = errors.New("no pending rewards of this type")
	ErrEmissionExhausted = errors.New("daily emission budget exhausted")
)

// ClaimRequest represents a reward claim.
type ClaimRequest struct {
	DID          string `json:"did"`
	RewardType   string `json:"rewardType"`
	TrustTier    int    `json:"trustTier"`
	MessageCount int    `json:"messageCount,omitempty"`
}

// ClaimResult contains the result of a claim.
type ClaimResult struct {
	ClaimID    string    `json:"claimId"`
	DID        string    `json:"did"`
	RewardType string    `json:"rewardType"`
	Amount     int64     `json:"amount"`
	TrustTier  int       `json:"trustTier"`
	Multiplier float64   `json:"multiplier"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

// PendingRewards shows pending rewards for a DID.
type PendingRewards struct {
	DID          string           `json:"did"`
	TrustTier    int              `json:"trustTier"`
	Multiplier   float64          `json:"multiplier"`
	Pending      map[string]int64 `json:"pending"`
	TotalPending int64            `json:"totalPending"`
	Timestamp    time.Time        `json:"timestamp"`
}

// DailyStats shows network-wide daily distribution statistics.
type DailyStats struct {
	Date                 string    `json:"date"`
	TotalDistributed     int64     `json:"totalDistributed"`
	DailyBudget          int64     `json:"dailyBudget"`
	EffectiveDailyBudget int64     `json:"effectiveDailyBudget"`
	RemainingBudget      int64     `json:"remainingBudget"`
	ClaimCount           int       `json:"claimCount"`
	AutoScaleRate        int64     `json:"autoScaleRate"`
	Timestamp            time.Time `json:"timestamp"`
}

// Service manages reward operations.
type Service struct {
	db                  database.DB
	emission            *internalrewards.EmissionSchedule
	antiGaming          *internalrewards.AntiGamingDetector
	mu                  sync.Mutex
	todayDistributed    int64
	todayClaimCount     int
	todayActivityWeight float64
	rolloverBudget      int64
	lastResetDate       string
	currentEmissionYear int
}

// NewService creates a reward service with emission schedule and anti-gaming protections.
func NewService(db database.DB, emission *internalrewards.EmissionSchedule) *Service {
	now := time.Now().UTC()
	return &Service{
		db:                  db,
		emission:            emission,
		antiGaming:          internalrewards.NewAntiGamingDetector(),
		lastResetDate:       now.Format("2006-01-02"),
		currentEmissionYear: emission.CurrentYear(),
	}
}

// Claim processes a reward claim using the AtomicAction pattern:
// verify tier -> calculate amount -> anti-gaming check -> record.
func (s *Service) Claim(ctx context.Context, req ClaimRequest) (*ClaimResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetDailyIfNeeded()

	// Step 1: Verify trust tier eligibility
	multiplier, ok := TrustMultiplier[req.TrustTier]
	if !ok || multiplier == 0 {
		return nil, ErrInsufficientTier
	}

	// Step 2: Calculate reward amount.
	amount := s.calculateReward(req, multiplier)
	if amount <= 0 {
		return nil, ErrNoPendingRewards
	}

	// Step 3: Check emission budget.
	remaining := s.emission.RemainingToday(s.todayDistributed)
	if amount > remaining {
		if remaining > 0 {
			amount = remaining // Partial claim
		} else {
			return nil, ErrEmissionExhausted
		}
	}

	// Step 4: Anti-gaming check
	event := internalrewards.ClaimEvent{
		DID:        req.DID,
		RewardType: req.RewardType,
		Amount:     amount,
		Timestamp:  time.Now(),
	}
	if err := s.antiGaming.CheckAndRecord(event); err != nil {
		return nil, err
	}

	// Step 5: Record claim (atomic)
	s.todayDistributed += amount
	s.todayClaimCount++
	if req.RewardType == "messaging" {
		messageCount := req.MessageCount
		if messageCount <= 0 {
			messageCount = 1
		}
		s.todayActivityWeight += multiplier * float64(messageCount)
	}

	return &ClaimResult{
		ClaimID:    uuid.New().String(),
		DID:        req.DID,
		RewardType: req.RewardType,
		Amount:     amount,
		TrustTier:  req.TrustTier,
		Multiplier: multiplier,
		Status:     "pending_l1",
		Timestamp:  time.Now(),
	}, nil
}

// GetPending returns pending reward summary for a DID.
func (s *Service) GetPending(ctx context.Context, did string, trustTier int) (*PendingRewards, error) {
	multiplier := TrustMultiplier[trustTier]

	// Calculate pending messaging rewards based on the current auto-scale rate,
	// but do not pay above the target base rate when network activity is still low.
	rate := s.AutoScaleRate()
	if rate > InitialRatePerMessage {
		rate = InitialRatePerMessage
	}
	pendingMessaging := int64(float64(rate) * multiplier)

	pending := map[string]int64{
		"messaging": pendingMessaging,
	}
	var total int64
	for _, amt := range pending {
		total += amt
	}

	return &PendingRewards{
		DID:          did,
		TrustTier:    trustTier,
		Multiplier:   multiplier,
		Pending:      pending,
		TotalPending: total,
		Timestamp:    time.Now(),
	}, nil
}

// GetDailyStats returns network-wide daily distribution statistics.
func (s *Service) GetDailyStats(ctx context.Context) (*DailyStats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetDailyIfNeeded()
	effectiveBudget := s.emission.DailyBudget() + s.rolloverBudget
	return &DailyStats{
		Date:                 time.Now().UTC().Format("2006-01-02"),
		TotalDistributed:     s.todayDistributed,
		DailyBudget:          s.emission.DailyBudget(),
		EffectiveDailyBudget: effectiveBudget,
		RemainingBudget:      s.emission.RemainingToday(s.todayDistributed) + s.rolloverBudget,
		ClaimCount:           s.todayClaimCount,
		AutoScaleRate:        s.autoScaleRateLocked(),
		Timestamp:            time.Now(),
	}, nil
}

// AutoScaleRate computes the current per-message reward rate.
// Formula: EffectiveDailyBudget ÷ TotalActivityWeight.
func (s *Service) AutoScaleRate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.autoScaleRateLocked()
}

func (s *Service) autoScaleRateLocked() int64 {
	if s.todayActivityWeight <= 0 {
		return InitialRatePerMessage
	}
	effectiveBudget := float64(s.emission.DailyBudget() + s.rolloverBudget)
	rate := int64(effectiveBudget / s.todayActivityWeight)
	if rate <= 0 {
		return InitialRatePerMessage
	}
	return rate
}

// RecordMessage increments the daily activity weight for auto-scaling.
func (s *Service) RecordMessage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetDailyIfNeeded()
	s.todayActivityWeight += 1.0
}

func (s *Service) calculateReward(req ClaimRequest, multiplier float64) int64 {
	switch req.RewardType {
	case "messaging":
		count := req.MessageCount
		if count <= 0 {
			count = 1
		}
		rate := s.autoScaleRateLocked()
		if rate > InitialRatePerMessage {
			rate = InitialRatePerMessage
		}
		return int64(float64(rate) * multiplier * float64(count))
	case "referral":
		return int64(float64(50_00000000) * multiplier) // 50 ECHO base
	case "staking":
		return 0 // Staking rewards computed by Currency L1
	default:
		return 0
	}
}

// resetDailyIfNeeded resets daily counters at UTC midnight.
func (s *Service) resetDailyIfNeeded() {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	currentYear := s.emission.CurrentYear()
	if s.lastResetDate == "" {
		s.lastResetDate = today
		s.currentEmissionYear = currentYear
		return
	}
	if s.lastResetDate == today {
		return
	}

	if currentYear == s.currentEmissionYear {
		effectiveBudget := s.emission.DailyBudget() + s.rolloverBudget
		unused := effectiveBudget - s.todayDistributed
		if unused < 0 {
			unused = 0
		}
		s.rolloverBudget = unused
	} else {
		s.rolloverBudget = 0
	}

	s.todayDistributed = 0
	s.todayClaimCount = 0
	s.todayActivityWeight = 0
	s.lastResetDate = today
	s.currentEmissionYear = currentYear
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
