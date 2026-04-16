package wallet

import (
	"context"
	"time"
)

// MetagraphQuerier abstracts metagraph on-chain queries and transaction submission.
type MetagraphQuerier interface {
	GetBalance(ctx context.Context, did string) (*BalanceInfo, error)
	GetTokenLocks(ctx context.Context, did string) ([]TokenLockPos, error)
	GetDelegations(ctx context.Context, did string) ([]DelegationPos, error)
	GetValidators(ctx context.Context) ([]ValidatorInfo, error)
	SubmitTokenLock(ctx context.Context, did string, amount int64, tier StakingTier) (string, error)
	SubmitStakeDelegation(ctx context.Context, delegatorDID, stakeID, validatorID string, amount int64) (string, error)
	SubmitWithdrawLock(ctx context.Context, did, stakeID string, amount int64) (string, error)
	SubmitAtomicRewardClaim(ctx context.Context, did string, claims []RewardClaim) (string, error)
}

// RewardClaim is a single reward type + amount for atomic claiming.
type RewardClaim struct {
	RewardType string
	Amount     int64
}

// RewardsQuerier abstracts reward state queries.
type RewardsQuerier interface {
	GetPending(ctx context.Context, did string) (int64, error)
	GetPendingByType(ctx context.Context, did, rewardType string) (int64, error)
	GetAutoScaleState(ctx context.Context, did string) (*AutoScaleState, error)
	ClearPending(ctx context.Context, did string, types []string) error
}

// WalletService aggregates on-chain and cached data for the iOS Wallet tab.
type WalletService struct {
	metagraph MetagraphQuerier
	rewards   RewardsQuerier
}

// NewWalletService creates a new WalletService with the given dependencies.
func NewWalletService(metagraph MetagraphQuerier, rewards RewardsQuerier) *WalletService {
	return &WalletService{
		metagraph: metagraph,
		rewards:   rewards,
	}
}

// GetWalletState returns the complete wallet view for a DID.
func (s *WalletService) GetWalletState(ctx context.Context, did string) (*WalletState, error) {
	balance, err := s.metagraph.GetBalance(ctx, did)
	if err != nil {
		return nil, err
	}

	locks, err := s.metagraph.GetTokenLocks(ctx, did)
	if err != nil {
		return nil, err
	}

	delegations, err := s.metagraph.GetDelegations(ctx, did)
	if err != nil {
		return nil, err
	}

	pending, err := s.rewards.GetPending(ctx, did)
	if err != nil {
		return nil, err
	}

	autoScaleState, err := s.rewards.GetAutoScaleState(ctx, did)
	if err != nil {
		return nil, err
	}

	var vesting *VestingState
	for _, lock := range locks {
		if lock.VestingType == "founder" {
			vesting = s.computeVestingState(lock)
			break
		}
	}

	return &WalletState{
		DID:            did,
		TotalBalance:   balance.Total,
		Available:      balance.Available,
		Staked:         sumLocks(locks),
		PendingRewards: pending,
		Locks:          locks,
		Delegations:    delegations,
		DailyRewards:   autoScaleState,
		Vesting:        vesting,
	}, nil
}

// StakeEcho constructs and submits a TokenLock transaction.
func (s *WalletService) StakeEcho(ctx context.Context, req StakeRequest) (*StakeResult, error) {
	tier, err := ValidateTier(req.Tier)
	if err != nil {
		return nil, err
	}

	txHash, err := s.metagraph.SubmitTokenLock(ctx, req.DID, req.Amount, tier)
	if err != nil {
		return nil, err
	}

	return &StakeResult{TxHash: txHash, Tier: tier}, nil
}

// DelegateToValidator constructs and submits a StakeDelegation transaction.
func (s *WalletService) DelegateToValidator(ctx context.Context, req DelegateRequest) (*DelegateResult, error) {
	txHash, err := s.metagraph.SubmitStakeDelegation(ctx, req.DID, req.StakeID, req.ValidatorID, req.Amount)
	if err != nil {
		return nil, err
	}

	return &DelegateResult{TxHash: txHash}, nil
}

// Unstake constructs and submits a WithdrawLock transaction (14-day cooldown).
func (s *WalletService) Unstake(ctx context.Context, req UnstakeRequest) (*UnstakeResult, error) {
	txHash, err := s.metagraph.SubmitWithdrawLock(ctx, req.DID, req.StakeID, req.Amount)
	if err != nil {
		return nil, err
	}

	return &UnstakeResult{
		TxHash:          txHash,
		CooldownEndDate: time.Now().Add(14 * 24 * time.Hour),
	}, nil
}

// ClaimRewards constructs and submits an AtomicAction for reward claiming.
func (s *WalletService) ClaimRewards(ctx context.Context, did string, types []string) (*ClaimResult, error) {
	var claims []RewardClaim
	for _, rewardType := range types {
		pending, _ := s.rewards.GetPendingByType(ctx, did, rewardType)
		if pending > 0 {
			claims = append(claims, RewardClaim{RewardType: rewardType, Amount: pending})
		}
	}

	if len(claims) == 0 {
		return nil, ErrNoPendingRewards
	}

	txHash, err := s.metagraph.SubmitAtomicRewardClaim(ctx, did, claims)
	if err != nil {
		return nil, err
	}

	_ = s.rewards.ClearPending(ctx, did, types)

	return &ClaimResult{TxHash: txHash}, nil
}

// GetValidators returns active L1 validators with performance metrics.
func (s *WalletService) GetValidators(ctx context.Context) ([]ValidatorInfo, error) {
	return s.metagraph.GetValidators(ctx)
}

func sumLocks(locks []TokenLockPos) int64 {
	var total int64
	for _, l := range locks {
		total += l.Amount
	}
	return total
}

func (s *WalletService) computeVestingState(lock TokenLockPos) *VestingState {
	// Founder vesting: 12-month cliff, 48-month linear vest
	cliffMonths := 12
	vestMonths := 48
	cliffDate := lock.LockedUntil.AddDate(0, -(vestMonths - cliffMonths), 0)
	now := time.Now()
	cliffCompleted := now.After(cliffDate)

	var vestedAmount int64
	var vestingPercent float64
	if !cliffCompleted {
		vestedAmount = 0
		vestingPercent = 0
	} else {
		monthsElapsed := int(now.Sub(cliffDate).Hours() / (24 * 30))
		if monthsElapsed > vestMonths-cliffMonths {
			monthsElapsed = vestMonths - cliffMonths
		}
		vestingPercent = float64(monthsElapsed) / float64(vestMonths-cliffMonths) * 100
		vestedAmount = int64(float64(lock.Amount) * vestingPercent / 100)
	}

	locked := lock.Amount - vestedAmount
	// Withdrawable is vested minus what's already been withdrawn (simplified: assume none withdrawn yet)
	withdrawable := vestedAmount

	// Next unlock: monthly vesting
	var nextDate time.Time
	var nextAmount int64
	if cliffCompleted && vestingPercent < 100 {
		monthlyAmount := lock.Amount / int64(vestMonths-cliffMonths)
		nextAmount = monthlyAmount
		monthsElapsed := int(now.Sub(cliffDate).Hours() / (24 * 30))
		nextDate = cliffDate.AddDate(0, monthsElapsed+1, 0)
	}

	return &VestingState{
		Role:             "founder",
		TotalAllocated:   lock.Amount,
		Vested:           vestedAmount,
		Locked:           locked,
		Withdrawable:     withdrawable,
		NextUnlockAmount: nextAmount,
		NextUnlockDate:   nextDate,
		CliffDate:        cliffDate,
		CliffCompleted:   cliffCompleted,
		VestingPercent:   vestingPercent,
	}
}
