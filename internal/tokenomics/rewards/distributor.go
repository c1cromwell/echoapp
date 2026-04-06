package rewards

import (
	"math/big"
	"time"
)

// RewardCalculator computes reward amounts
type RewardCalculator struct {
	BaseRewards map[int]float64 // RewardType -> base amount
}

// NewRewardCalculator creates a calculator with default rates
func NewRewardCalculator() *RewardCalculator {
	return &RewardCalculator{
		BaseRewards: map[int]float64{
			0: 0.01, // Text
			1: 0.02, // Voice
			2: 0.03, // Video
			3: 0.05, // Referral
			4: 0.10, // Governance
			5: 0.00, // Staking (calculated separately)
			6: 0.00, // Burn (calculated separately)
			7: 0.00, // Bridge (calculated separately)
		},
	}
}

// CalculateReward computes reward with multiplier
func (rc *RewardCalculator) CalculateReward(rewardType int, multiplier float64) *big.Int {
	baseAmount := rc.BaseRewards[rewardType]
	finalAmount := baseAmount * multiplier

	// Convert to big.Int with 8 decimals
	result := new(big.Int)
	result.SetString(floatToBigInt(finalAmount, 8), 10)

	return result
}

// floatToBigInt converts float to big.Int with specified decimals
func floatToBigInt(f float64, decimals int) string {
	// Simple conversion for demonstration
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	intVal := int64(f * multiplier)
	return string(rune(intVal))
}

// RewardDistributor manages reward distribution
type RewardDistributor struct {
	DailyLimits map[string]int      // user -> messages rewarded
	EchoEarned  map[string]*big.Int // user -> total earned
	LastUpdated map[string]time.Time
}

// NewRewardDistributor creates a distributor
func NewRewardDistributor() *RewardDistributor {
	return &RewardDistributor{
		DailyLimits: make(map[string]int),
		EchoEarned:  make(map[string]*big.Int),
		LastUpdated: make(map[string]time.Time),
	}
}

// CanDistribute checks if user can receive more rewards today
func (rd *RewardDistributor) CanDistribute(userID string) bool {
	if count, exists := rd.DailyLimits[userID]; exists {
		return count < 500
	}
	return true
}

// IncrementCount tracks distributed rewards
func (rd *RewardDistributor) IncrementCount(userID string) {
	if _, exists := rd.DailyLimits[userID]; !exists {
		rd.DailyLimits[userID] = 0
		rd.EchoEarned[userID] = big.NewInt(0)
	}
	rd.DailyLimits[userID]++
	rd.LastUpdated[userID] = time.Now()
}

// BatchRewardProcessor handles batch reward processing
type BatchRewardProcessor struct {
	BatchSize int
	Queue     []*rewardItem
}

type rewardItem struct {
	UserID string
	Amount *big.Int
}

// NewBatchRewardProcessor creates a processor
func NewBatchRewardProcessor(batchSize int) *BatchRewardProcessor {
	return &BatchRewardProcessor{
		BatchSize: batchSize,
		Queue:     make([]*rewardItem, 0),
	}
}

// Enqueue adds a reward to the processing queue
func (brp *BatchRewardProcessor) Enqueue(userID string, amount *big.Int) {
	brp.Queue = append(brp.Queue, &rewardItem{
		UserID: userID,
		Amount: amount,
	})
}

// ShouldFlush checks if batch is ready to process
func (brp *BatchRewardProcessor) ShouldFlush() bool {
	return len(brp.Queue) >= brp.BatchSize
}

// PoolManager manages reward pools
type PoolManager struct {
	UserRewardPool      *big.Int
	ValidatorRewardPool *big.Int
	EcosystemPool       *big.Int
}

// NewPoolManager creates a pool manager with allocations
func NewPoolManager(total *big.Int) *PoolManager {
	userPool := new(big.Int).Mul(total, big.NewInt(40))
	userPool.Div(userPool, big.NewInt(100))

	validatorPool := new(big.Int).Mul(total, big.NewInt(25))
	validatorPool.Div(validatorPool, big.NewInt(100))

	ecosystemPool := new(big.Int).Mul(total, big.NewInt(20))
	ecosystemPool.Div(ecosystemPool, big.NewInt(100))

	return &PoolManager{
		UserRewardPool:      userPool,
		ValidatorRewardPool: validatorPool,
		EcosystemPool:       ecosystemPool,
	}
}

// CanWithdraw checks if pool has sufficient funds
func (pm *PoolManager) CanWithdraw(poolType string, amount *big.Int) bool {
	switch poolType {
	case "user":
		return pm.UserRewardPool.Cmp(amount) >= 0
	case "validator":
		return pm.ValidatorRewardPool.Cmp(amount) >= 0
	case "ecosystem":
		return pm.EcosystemPool.Cmp(amount) >= 0
	}
	return false
}

// Withdraw reduces pool balance
func (pm *PoolManager) Withdraw(poolType string, amount *big.Int) bool {
	if !pm.CanWithdraw(poolType, amount) {
		return false
	}

	switch poolType {
	case "user":
		pm.UserRewardPool.Sub(pm.UserRewardPool, amount)
	case "validator":
		pm.ValidatorRewardPool.Sub(pm.ValidatorRewardPool, amount)
	case "ecosystem":
		pm.EcosystemPool.Sub(pm.EcosystemPool, amount)
	}

	return true
}
