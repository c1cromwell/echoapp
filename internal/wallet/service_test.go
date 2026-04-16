package wallet

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- Mock MetagraphQuerier ---

type mockMetagraph struct {
	balance     *BalanceInfo
	locks       []TokenLockPos
	delegations []DelegationPos
	validators  []ValidatorInfo
	txHash      string
	err         error
}

func (m *mockMetagraph) GetBalance(_ context.Context, _ string) (*BalanceInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.balance, nil
}

func (m *mockMetagraph) GetTokenLocks(_ context.Context, _ string) ([]TokenLockPos, error) {
	return m.locks, nil
}

func (m *mockMetagraph) GetDelegations(_ context.Context, _ string) ([]DelegationPos, error) {
	return m.delegations, nil
}

func (m *mockMetagraph) GetValidators(_ context.Context) ([]ValidatorInfo, error) {
	return m.validators, nil
}

func (m *mockMetagraph) SubmitTokenLock(_ context.Context, _ string, _ int64, _ StakingTier) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.txHash, nil
}

func (m *mockMetagraph) SubmitStakeDelegation(_ context.Context, _, _, _ string, _ int64) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.txHash, nil
}

func (m *mockMetagraph) SubmitWithdrawLock(_ context.Context, _, _ string, _ int64) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.txHash, nil
}

func (m *mockMetagraph) SubmitAtomicRewardClaim(_ context.Context, _ string, _ []RewardClaim) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.txHash, nil
}

// --- Mock RewardsQuerier ---

type mockRewards struct {
	pending        int64
	pendingMap     map[string]int64
	autoScaleState *AutoScaleState
	err            error
}

func (m *mockRewards) GetPending(_ context.Context, _ string) (int64, error) {
	return m.pending, m.err
}

func (m *mockRewards) GetPendingByType(_ context.Context, _ string, rewardType string) (int64, error) {
	if m.pendingMap != nil {
		return m.pendingMap[rewardType], nil
	}
	return m.pending, nil
}

func (m *mockRewards) GetAutoScaleState(_ context.Context, _ string) (*AutoScaleState, error) {
	return m.autoScaleState, m.err
}

func (m *mockRewards) ClearPending(_ context.Context, _ string, _ []string) error {
	return nil
}

// --- Tests ---

func TestGetWalletState(t *testing.T) {
	mg := &mockMetagraph{
		balance:     &BalanceInfo{Total: 1000_00000000, Available: 500_00000000},
		locks:       []TokenLockPos{{ID: "lock1", Amount: 500_00000000, Tier: "gold"}},
		delegations: []DelegationPos{{ID: "del1", Amount: 300_00000000}},
		txHash:      "hash1",
	}
	rw := &mockRewards{
		pending: 50_00000000,
		autoScaleState: &AutoScaleState{
			CurrentRate:          10000000,     // 0.1 ECHO
			DailyBudget:          100000000000, // 1000 ECHO
			EffectiveDailyBudget: 100000000000,
			BudgetUsedToday:      10000000000, // 100 ECHO
			RemainingToday:       90000000000,
			TotalActivityWeight:  1000.0,
			LastUpdated:          "2024-01-01T12:00:00Z",
		},
	}

	svc := NewWalletService(mg, rw)
	state, err := svc.GetWalletState(context.Background(), "did:echo:test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.TotalBalance != 1000_00000000 {
		t.Errorf("expected total balance 1000_00000000, got %d", state.TotalBalance)
	}
	if state.Available != 500_00000000 {
		t.Errorf("expected available 500_00000000, got %d", state.Available)
	}
	if state.Staked != 500_00000000 {
		t.Errorf("expected staked 500_00000000, got %d", state.Staked)
	}
	if state.PendingRewards != 50_00000000 {
		t.Errorf("expected pending rewards 50_00000000, got %d", state.PendingRewards)
	}
	if len(state.Locks) != 1 {
		t.Errorf("expected 1 lock, got %d", len(state.Locks))
	}
	if state.Vesting != nil {
		t.Error("expected no vesting for non-founder lock")
	}
}

func TestGetWalletState_FounderVesting(t *testing.T) {
	mg := &mockMetagraph{
		balance: &BalanceInfo{Total: 60_000_000_00000000, Available: 0},
		locks: []TokenLockPos{{
			ID:          "founder1",
			Amount:      60_000_000_00000000,
			Tier:        "platinum",
			VestingType: "founder",
			LockedUntil: time.Now().Add(36 * 30 * 24 * time.Hour), // ~36 months from now
		}},
	}
	rw := &mockRewards{
		pending:        0,
		autoScaleState: &AutoScaleState{},
	}

	svc := NewWalletService(mg, rw)
	state, err := svc.GetWalletState(context.Background(), "did:echo:founder")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Vesting == nil {
		t.Fatal("expected vesting state for founder lock")
	}
	if state.Vesting.TotalAllocated != 60_000_000_00000000 {
		t.Errorf("expected total allocated 60_000_000_00000000, got %d", state.Vesting.TotalAllocated)
	}
}

func TestStakeEcho(t *testing.T) {
	mg := &mockMetagraph{txHash: "stake_hash_1"}
	rw := &mockRewards{}
	svc := NewWalletService(mg, rw)

	result, err := svc.StakeEcho(context.Background(), StakeRequest{
		DID:    "did:echo:test",
		Amount: 100_00000000,
		Tier:   "gold",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TxHash != "stake_hash_1" {
		t.Errorf("expected hash stake_hash_1, got %s", result.TxHash)
	}
	if result.Tier.DurationDays != 180 {
		t.Errorf("expected gold tier 180 days, got %d", result.Tier.DurationDays)
	}
}

func TestStakeEcho_InvalidTier(t *testing.T) {
	mg := &mockMetagraph{}
	rw := &mockRewards{}
	svc := NewWalletService(mg, rw)

	_, err := svc.StakeEcho(context.Background(), StakeRequest{
		DID:    "did:echo:test",
		Amount: 100_00000000,
		Tier:   "diamond", // invalid
	})

	if !errors.Is(err, ErrInvalidTier) {
		t.Errorf("expected ErrInvalidTier, got %v", err)
	}
}

func TestDelegateToValidator(t *testing.T) {
	mg := &mockMetagraph{txHash: "delegate_hash_1"}
	rw := &mockRewards{}
	svc := NewWalletService(mg, rw)

	result, err := svc.DelegateToValidator(context.Background(), DelegateRequest{
		DID:         "did:echo:test",
		ValidatorID: "validator1",
		StakeID:     "lock1",
		Amount:      300_00000000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TxHash != "delegate_hash_1" {
		t.Errorf("expected delegate_hash_1, got %s", result.TxHash)
	}
}

func TestUnstake(t *testing.T) {
	mg := &mockMetagraph{txHash: "unstake_hash_1"}
	rw := &mockRewards{}
	svc := NewWalletService(mg, rw)

	result, err := svc.Unstake(context.Background(), UnstakeRequest{
		DID:     "did:echo:test",
		StakeID: "lock1",
		Amount:  100_00000000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TxHash != "unstake_hash_1" {
		t.Errorf("expected unstake_hash_1, got %s", result.TxHash)
	}
	// Cooldown should be ~14 days from now
	expected := time.Now().Add(14 * 24 * time.Hour)
	if result.CooldownEndDate.Sub(expected) > time.Minute {
		t.Errorf("cooldown end date too far from expected: %v", result.CooldownEndDate)
	}
}

func TestClaimRewards(t *testing.T) {
	mg := &mockMetagraph{txHash: "claim_hash_1"}
	rw := &mockRewards{
		pendingMap: map[string]int64{
			"messaging": 50_00000000,
			"staking":   25_00000000,
		},
	}
	svc := NewWalletService(mg, rw)

	result, err := svc.ClaimRewards(context.Background(), "did:echo:test", []string{"messaging", "staking"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TxHash != "claim_hash_1" {
		t.Errorf("expected claim_hash_1, got %s", result.TxHash)
	}
}

func TestClaimRewards_NoPending(t *testing.T) {
	mg := &mockMetagraph{txHash: "claim_hash_1"}
	rw := &mockRewards{
		pendingMap: map[string]int64{},
	}
	svc := NewWalletService(mg, rw)

	_, err := svc.ClaimRewards(context.Background(), "did:echo:test", []string{"messaging"})
	if !errors.Is(err, ErrNoPendingRewards) {
		t.Errorf("expected ErrNoPendingRewards, got %v", err)
	}
}

func TestGetValidators(t *testing.T) {
	mg := &mockMetagraph{
		validators: []ValidatorInfo{
			{ID: "v1", Uptime: 99.5, Commission: 5.0, Layer: "currency_l1"},
			{ID: "v2", Uptime: 98.0, Commission: 3.0, Layer: "data_l1"},
		},
	}
	rw := &mockRewards{}
	svc := NewWalletService(mg, rw)

	validators, err := svc.GetValidators(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(validators) != 2 {
		t.Errorf("expected 2 validators, got %d", len(validators))
	}
}

func TestValidateTier(t *testing.T) {
	tests := []struct {
		name    string
		tier    string
		wantErr bool
		wantAPR float64
	}{
		{"bronze", "bronze", false, 5.0},
		{"silver", "silver", false, 8.0},
		{"gold", "gold", false, 12.0},
		{"platinum", "platinum", false, 15.0},
		{"invalid", "diamond", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier, err := ValidateTier(tt.tier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTier(%s) error = %v, wantErr %v", tt.tier, err, tt.wantErr)
			}
			if !tt.wantErr && tier.APR != tt.wantAPR {
				t.Errorf("ValidateTier(%s) APR = %v, want %v", tt.tier, tier.APR, tt.wantAPR)
			}
		})
	}
}

func TestGetWalletState_BalanceError(t *testing.T) {
	mg := &mockMetagraph{err: errors.New("metagraph down")}
	rw := &mockRewards{}
	svc := NewWalletService(mg, rw)

	_, err := svc.GetWalletState(context.Background(), "did:echo:test")
	if err == nil {
		t.Error("expected error when metagraph is down")
	}
}
