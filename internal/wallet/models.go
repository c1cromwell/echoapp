// Package wallet provides the Wallet API service supporting the iOS Wallet tab.
// It aggregates on-chain and cached data for balance, staking, delegation, and rewards.
package wallet

import "time"

// WalletState is the complete wallet view for a DID.
type WalletState struct {
	DID            string          `json:"did"`
	TotalBalance   int64           `json:"totalBalance"`
	Available      int64           `json:"available"`
	Staked         int64           `json:"staked"`
	PendingRewards int64           `json:"pendingRewards"`
	Locks          []TokenLockPos  `json:"locks"`
	Delegations    []DelegationPos `json:"delegations"`
	DailyRewards   *AutoScaleState `json:"dailyRewards"`
	Vesting        *VestingState   `json:"vesting,omitempty"`
}

// TokenLockPos represents a staking position (on-chain TokenLock mirror).
type TokenLockPos struct {
	ID          string    `json:"id"`
	Amount      int64     `json:"amount"`
	Tier        string    `json:"tier"`
	LockedUntil time.Time `json:"lockedUntil"`
	VestingType string    `json:"vestingType,omitempty"`
	DelegatedTo string    `json:"delegatedTo,omitempty"`
}

// DelegationPos represents a validator delegation position.
type DelegationPos struct {
	ID          string        `json:"id"`
	StakeID     string        `json:"stakeId"`
	ValidatorID string        `json:"validatorId"`
	Validator   ValidatorInfo `json:"validator"`
	Amount      int64         `json:"amount"`
	Since       time.Time     `json:"since"`
}

// ValidatorInfo contains L1 validator performance metrics.
type ValidatorInfo struct {
	ID             string  `json:"id"`
	Address        string  `json:"address"`
	Uptime         float64 `json:"uptimePercent"`
	Commission     float64 `json:"commissionPercent"`
	TotalDelegated int64   `json:"totalDelegated"`
	DelegatorCount int     `json:"delegatorCount"`
	Layer          string  `json:"layer"` // currency_l1, data_l1
	EstimatedAPR   float64 `json:"estimatedApr"`
}

// VestingState represents a founder vesting position.
type VestingState struct {
	Role             string    `json:"role"`
	TotalAllocated   int64     `json:"totalAllocated"`
	Vested           int64     `json:"vested"`
	Locked           int64     `json:"locked"`
	Withdrawable     int64     `json:"withdrawable"`
	NextUnlockAmount int64     `json:"nextUnlockAmount"`
	NextUnlockDate   time.Time `json:"nextUnlockDate"`
	CliffDate        time.Time `json:"cliffDate"`
	CliffCompleted   bool      `json:"cliffCompleted"`
	VestingPercent   float64   `json:"vestingPercent"`
	ExplorerURL      string    `json:"explorerUrl"`
}

// AutoScaleState replaces DailyCapState.
// Per PRD v2.5.1: auto-scaling model adopted, daily caps removed.
type AutoScaleState struct {
	CurrentRate          int64   `json:"currentRate"`          // Current per-message base rate (8 decimals)
	DailyBudget          int64   `json:"dailyBudget"`          // Today's emission budget from annual curve
	EffectiveDailyBudget int64   `json:"effectiveDailyBudget"` // Including rollover from low-activity days
	BudgetUsedToday      int64   `json:"budgetUsedToday"`      // Total distributed today
	RemainingToday       int64   `json:"remainingToday"`       // Budget remaining today
	TotalActivityWeight  float64 `json:"totalActivityWeight"`  // Sum of tier-weighted messages today
	LastUpdated          string  `json:"lastUpdated"`          // ISO timestamp of last rate recalculation
}

// StakeRequest is the input for staking ECHO via TokenLock.
type StakeRequest struct {
	DID    string `json:"did"`
	Amount int64  `json:"amount"`
	Tier   string `json:"tier"`
}

// StakeResult is returned after a successful TokenLock submission.
type StakeResult struct {
	TxHash string      `json:"txHash"`
	Tier   StakingTier `json:"tier"`
}

// DelegateRequest is the input for delegating staked ECHO to a validator.
type DelegateRequest struct {
	DID         string `json:"did"`
	ValidatorID string `json:"validatorId"`
	StakeID     string `json:"stakeId"`
	Amount      int64  `json:"amount"`
}

// DelegateResult is returned after a successful StakeDelegation submission.
type DelegateResult struct {
	TxHash string `json:"txHash"`
}

// UnstakeRequest is the input for initiating unstaking with cooldown.
type UnstakeRequest struct {
	DID     string `json:"did"`
	StakeID string `json:"stakeId"`
	Amount  int64  `json:"amount"`
}

// UnstakeResult is returned after a successful WithdrawLock submission.
type UnstakeResult struct {
	TxHash          string    `json:"txHash"`
	CooldownEndDate time.Time `json:"cooldownEndDate"`
}

// ClaimResult is returned after a successful reward claim.
type ClaimResult struct {
	TxHash string `json:"txHash"`
}

// StakingTier defines a staking tier with duration and APR.
type StakingTier struct {
	Name         string  `json:"name"`
	DurationDays int     `json:"durationDays"`
	APR          float64 `json:"apr"`
}

// StakingTiers maps tier names to their configuration.
var StakingTiers = map[string]StakingTier{
	"bronze":   {Name: "bronze", DurationDays: 30, APR: 5.0},
	"silver":   {Name: "silver", DurationDays: 90, APR: 8.0},
	"gold":     {Name: "gold", DurationDays: 180, APR: 12.0},
	"platinum": {Name: "platinum", DurationDays: 365, APR: 15.0},
}

// ValidateTier returns the StakingTier for the given name or an error.
func ValidateTier(name string) (StakingTier, error) {
	tier, ok := StakingTiers[name]
	if !ok {
		return StakingTier{}, ErrInvalidTier
	}
	return tier, nil
}

// BalanceInfo is the raw balance data from the metagraph.
type BalanceInfo struct {
	Total     int64 `json:"total"`
	Available int64 `json:"available"`
}
