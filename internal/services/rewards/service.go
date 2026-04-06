// Package rewards provides the reward service layer for the Echo backend.
package rewards

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
	internalrewards "github.com/thechadcromwell/echoapp/internal/rewards"
)

const (
	BaseRatePerMessage = int64(1_00000000) // 1 ECHO per message (8 decimals)
	DecayThreshold     = 100               // Messages before decay starts
	DecayFactor        = 0.01              // 1% decay per message over threshold
	MinRateMultiplier  = 0.01              // Floor: 1% of base rate
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
	Date             string    `json:"date"`
	TotalDistributed int64     `json:"totalDistributed"`
	DailyBudget      int64     `json:"dailyBudget"`
	RemainingBudget  int64     `json:"remainingBudget"`
	ClaimCount       int       `json:"claimCount"`
	AutoScaleRate    int64     `json:"autoScaleRate"`
	Timestamp        time.Time `json:"timestamp"`
}

// Service manages reward operations.
type Service struct {
	db               database.DB
	emission         *internalrewards.EmissionSchedule
	antiGaming       *internalrewards.AntiGamingDetector
	mu               sync.Mutex
	todayDistributed int64
	todayClaimCount  int
	todayMessages    int
	lastResetDate    string
}

// NewService creates a reward service with emission schedule and anti-gaming protections.
func NewService(db database.DB, emission *internalrewards.EmissionSchedule) *Service {
	return &Service{
		db:         db,
		emission:   emission,
		antiGaming: internalrewards.NewAntiGamingDetector(),
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

	// Step 2: Calculate reward amount
	amount := s.calculateReward(req, multiplier)
	if amount <= 0 {
		return nil, ErrNoPendingRewards
	}

	// Step 3: Check emission budget
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

	// Calculate pending messaging rewards based on auto-scale rate
	rate := s.AutoScaleRate()
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

	return &DailyStats{
		Date:             time.Now().UTC().Format("2006-01-02"),
		TotalDistributed: s.todayDistributed,
		DailyBudget:      s.emission.DailyBudget(),
		RemainingBudget:  s.emission.RemainingToday(s.todayDistributed),
		ClaimCount:       s.todayClaimCount,
		AutoScaleRate:    s.AutoScaleRate(),
		Timestamp:        time.Now(),
	}, nil
}

// AutoScaleRate computes the current per-message reward rate.
// Formula: base_rate * max(0.01, 1.0 - 0.01 * max(0, msgs_today - 100))
func (s *Service) AutoScaleRate() int64 {
	decayInput := float64(intMax(0, s.todayMessages-DecayThreshold))
	decayMult := math.Max(MinRateMultiplier, 1.0-DecayFactor*decayInput)
	return int64(float64(BaseRatePerMessage) * decayMult)
}

// RecordMessage increments the daily message counter for auto-scaling.
func (s *Service) RecordMessage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetDailyIfNeeded()
	s.todayMessages++
}

func (s *Service) calculateReward(req ClaimRequest, multiplier float64) int64 {
	switch req.RewardType {
	case "messaging":
		count := req.MessageCount
		if count <= 0 {
			count = 1
		}
		rate := s.AutoScaleRate()
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
	today := time.Now().UTC().Format("2006-01-02")
	if s.lastResetDate != today {
		s.todayDistributed = 0
		s.todayClaimCount = 0
		s.todayMessages = 0
		s.lastResetDate = today
	}
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
