package metagraph

import (
	"errors"
	"math/big"
	"time"
)

// SlashingOffense categorizes the type of validator misbehavior.
type SlashingOffense string

const (
	// OffenseFraudulentReward: validating inflated or exceeded daily cap reward claims.
	OffenseFraudulentReward SlashingOffense = "fraudulent_reward"
	// OffenseInvalidMerkle: submitting malformed Merkle roots or unauthorized sender DIDs.
	OffenseInvalidMerkle SlashingOffense = "invalid_merkle"
	// OffenseExtendedDowntime: >24h continuous offline.
	OffenseExtendedDowntime SlashingOffense = "extended_downtime"
	// OffenseDoubleSigning: conflicting blocks at the same snapshot height.
	OffenseDoubleSigning SlashingOffense = "double_signing"
	// OffenseCollusion: colluding to bypass anti-gaming rules.
	OffenseCollusion SlashingOffense = "collusion"
)

// SlashingPenalty defines the penalty parameters for each offense type.
type SlashingPenalty struct {
	Offense        SlashingOffense `json:"offense"`
	StakePercent   int             `json:"stake_percent"`   // % of staked ECHO slashed
	SuspensionDays int            `json:"suspension_days"` // 0 = no suspension, -1 = permanent ban
	Recoverable    bool           `json:"recoverable"`
	Description    string         `json:"description"`
}

// DefaultSlashingPenalties returns the v3.2 slashing schedule.
// These are governance-adjustable via supermajority vote.
func DefaultSlashingPenalties() map[SlashingOffense]SlashingPenalty {
	return map[SlashingOffense]SlashingPenalty{
		OffenseFraudulentReward: {
			Offense:        OffenseFraudulentReward,
			StakePercent:   10,
			SuspensionDays: 30,
			Recoverable:    true,
			Description:    "Validating fraudulent reward claims (inflated amounts, exceeded daily caps)",
		},
		OffenseInvalidMerkle: {
			Offense:        OffenseInvalidMerkle,
			StakePercent:   5,
			SuspensionDays: 0,
			Recoverable:    true,
			Description:    "Submitting invalid Merkle roots (malformed structure, unauthorized sender DID)",
		},
		OffenseExtendedDowntime: {
			Offense:        OffenseExtendedDowntime,
			StakePercent:   1, // per 24h block
			SuspensionDays: 0,
			Recoverable:    true,
			Description:    "Extended downtime (>24h continuous offline), 1% per 24h block",
		},
		OffenseDoubleSigning: {
			Offense:        OffenseDoubleSigning,
			StakePercent:   50,
			SuspensionDays: -1, // permanent
			Recoverable:    false,
			Description:    "Double-signing (conflicting blocks at same snapshot height)",
		},
		OffenseCollusion: {
			Offense:        OffenseCollusion,
			StakePercent:   25,
			SuspensionDays: -1, // permanent, unless governance reversal
			Recoverable:    false,
			Description:    "Colluding to bypass anti-gaming rules",
		},
	}
}

// SlashingEvent records a validator slashing incident.
type SlashingEvent struct {
	EventID         string          `json:"event_id"`
	ValidatorDID    string          `json:"validator_did"`
	Offense         SlashingOffense `json:"offense"`
	EvidenceHash    string          `json:"evidence_hash"`
	SlashedAmount   *big.Int        `json:"slashed_amount"`
	StakePercent    int             `json:"stake_percent"`
	SuspendedUntil  time.Time       `json:"suspended_until,omitempty"`
	PermanentBan    bool            `json:"permanent_ban"`
	DetectedBy      string          `json:"detected_by"` // peer validator DID or "l0_heartbeat"
	SnapshotHeight  int64           `json:"snapshot_height"`
	TreasuryCredits *big.Int        `json:"treasury_credits"` // slashed tokens go to community treasury
	OccurredAt      time.Time       `json:"occurred_at"`
}

// ValidatorStatus tracks a validator's current standing.
type ValidatorStatus struct {
	ValidatorDID   string           `json:"validator_did"`
	StakedAmount   *big.Int         `json:"staked_amount"`
	IsActive       bool             `json:"is_active"`
	IsSuspended    bool             `json:"is_suspended"`
	IsBanned       bool             `json:"is_banned"`
	SuspendedUntil time.Time        `json:"suspended_until,omitempty"`
	SlashingHistory []SlashingEvent `json:"slashing_history"`
	Layer          L1Layer          `json:"layer"`
	UptimePercent  float64          `json:"uptime_percent"`
	LastHeartbeat  time.Time        `json:"last_heartbeat"`
}

// CalculateSlash computes the slashing amount and creates a SlashingEvent.
func CalculateSlash(validator *ValidatorStatus, offense SlashingOffense, evidenceHash, detectedBy string, snapshotHeight int64) (*SlashingEvent, error) {
	penalties := DefaultSlashingPenalties()
	penalty, ok := penalties[offense]
	if !ok {
		return nil, errors.New("unknown slashing offense")
	}

	slashedAmount := new(big.Int).Mul(validator.StakedAmount, big.NewInt(int64(penalty.StakePercent)))
	slashedAmount.Div(slashedAmount, big.NewInt(100))

	now := time.Now()
	event := &SlashingEvent{
		EventID:         now.Format("20060102150405") + "-slash",
		ValidatorDID:    validator.ValidatorDID,
		Offense:         offense,
		EvidenceHash:    evidenceHash,
		SlashedAmount:   slashedAmount,
		StakePercent:    penalty.StakePercent,
		PermanentBan:    !penalty.Recoverable,
		DetectedBy:      detectedBy,
		SnapshotHeight:  snapshotHeight,
		TreasuryCredits: new(big.Int).Set(slashedAmount),
		OccurredAt:      now,
	}

	if penalty.SuspensionDays > 0 {
		event.SuspendedUntil = now.AddDate(0, 0, penalty.SuspensionDays)
	}

	return event, nil
}

// ApplySlash updates a validator's status after a slashing event.
// Delegators are never slashed — only the validator's own staked ECHO.
func ApplySlash(validator *ValidatorStatus, event *SlashingEvent) {
	validator.StakedAmount.Sub(validator.StakedAmount, event.SlashedAmount)
	validator.SlashingHistory = append(validator.SlashingHistory, *event)

	if event.PermanentBan {
		validator.IsBanned = true
		validator.IsActive = false
	} else if !event.SuspendedUntil.IsZero() {
		validator.IsSuspended = true
		validator.SuspendedUntil = event.SuspendedUntil
		validator.IsActive = false
	}
}
