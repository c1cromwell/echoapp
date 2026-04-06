package staking

import (
	"math/big"
	"time"
)

// StakingTier represents a staking option
type StakingTier struct {
	Name                string
	LockupDays          int
	APYPercent          float64
	GovernanceWeight    float64
	EarlyUnstakePenalty float64
}

// GetStakingTiers returns all available tiers
func GetStakingTiers() []StakingTier {
	return []StakingTier{
		{
			Name:                "Flexible",
			LockupDays:          0,
			APYPercent:          3.0,
			GovernanceWeight:    1.0,
			EarlyUnstakePenalty: 0.0,
		},
		{
			Name:                "30-Day Lock",
			LockupDays:          30,
			APYPercent:          5.0,
			GovernanceWeight:    1.25,
			EarlyUnstakePenalty: 25.0,
		},
		{
			Name:                "90-Day Lock",
			LockupDays:          90,
			APYPercent:          8.0,
			GovernanceWeight:    1.5,
			EarlyUnstakePenalty: 50.0,
		},
		{
			Name:                "180-Day Lock",
			LockupDays:          180,
			APYPercent:          12.0,
			GovernanceWeight:    2.0,
			EarlyUnstakePenalty: 75.0,
		},
		{
			Name:                "365-Day Lock",
			LockupDays:          365,
			APYPercent:          15.0,
			GovernanceWeight:    3.0,
			EarlyUnstakePenalty: 90.0,
		},
	}
}

// Stake represents a user's staked tokens
type Stake struct {
	StakeID           string
	UserID            string
	Amount            *big.Int
	Tier              StakingTier
	CreatedAt         time.Time
	UnlocksAt         time.Time
	LastRewardClaim   time.Time
	CompoundedRewards *big.Int
}

// IsLocked checks if stake is still locked
func (s *Stake) IsLocked() bool {
	return time.Now().Before(s.UnlocksAt)
}

// CalculatePendingReward computes accrued but unclaimed rewards
func (s *Stake) CalculatePendingReward() *big.Int {
	if time.Now().Before(s.CreatedAt) {
		return big.NewInt(0)
	}

	elapsed := time.Since(s.LastRewardClaim)
	days := float64(elapsed.Hours() / 24)

	// Annual reward calculation
	annualRewardFloat := new(big.Float).Mul(
		new(big.Float).SetInt(s.Amount),
		new(big.Float).SetFloat64(s.Tier.APYPercent/100),
	)

	dailyRewardFloat := new(big.Float).Quo(annualRewardFloat, new(big.Float).SetFloat64(365))
	totalRewardFloat := new(big.Float).Mul(dailyRewardFloat, new(big.Float).SetFloat64(days))

	result := new(big.Int)
	totalRewardFloat.Int(result)

	return result
}

// StakingManager handles stake lifecycle
type StakingManager struct {
	Stakes           map[string]*Stake
	UserStakes       map[string][]*Stake
	ValidatorMinimum *big.Int
}

// NewStakingManager creates a manager
func NewStakingManager() *StakingManager {
	minimum := new(big.Int)
	minimum.SetString("5000000000000", 10) // 50,000 ECHO

	return &StakingManager{
		Stakes:           make(map[string]*Stake),
		UserStakes:       make(map[string][]*Stake),
		ValidatorMinimum: minimum,
	}
}

// CreateStake initiates a new stake
func (sm *StakingManager) CreateStake(userID string, amount *big.Int, tierIndex int) *Stake {
	tiers := GetStakingTiers()
	if tierIndex >= len(tiers) {
		return nil
	}

	tier := tiers[tierIndex]
	stake := &Stake{
		StakeID:           generateID(),
		UserID:            userID,
		Amount:            amount,
		Tier:              tier,
		CreatedAt:         time.Now(),
		UnlocksAt:         time.Now().AddDate(0, 0, tier.LockupDays),
		LastRewardClaim:   time.Now(),
		CompoundedRewards: big.NewInt(0),
	}

	sm.Stakes[stake.StakeID] = stake
	sm.UserStakes[userID] = append(sm.UserStakes[userID], stake)

	return stake
}

// GetStake retrieves a stake by ID
func (sm *StakingManager) GetStake(stakeID string) *Stake {
	return sm.Stakes[stakeID]
}

// CanBeValidator checks if user qualifies as validator
func (sm *StakingManager) CanBeValidator(userID string) bool {
	totalStaked := big.NewInt(0)

	for _, stake := range sm.UserStakes[userID] {
		totalStaked.Add(totalStaked, stake.Amount)
	}

	return totalStaked.Cmp(sm.ValidatorMinimum) >= 0
}

// GetGovernanceWeight calculates voting power
func (sm *StakingManager) GetGovernanceWeight(userID string) float64 {
	weight := 0.0

	for _, stake := range sm.UserStakes[userID] {
		if !stake.IsLocked() {
			stakeInECHO := new(big.Float).SetInt(stake.Amount)
			stakeInECHO.Quo(stakeInECHO, new(big.Float).SetInt64(100000000))

			val, _ := stakeInECHO.Float64()
			weight += val * stake.Tier.GovernanceWeight
		}
	}

	return weight
}

// Helper function for ID generation
func generateID() string {
	return time.Now().Format("20060102150405") + "-stub"
}
