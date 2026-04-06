package metagraph

import (
	"math/big"
	"testing"
	"time"
)

// --- TokenLock Tests ---

func TestNewTokenLock_ValidTier(t *testing.T) {
	tiers := DefaultStakingTiers()
	amount := big.NewInt(10000000000) // 100 ECHO (meets Tier 1 minimum)

	lock, err := NewTokenLock("tx-001", "did:dag:sender1", amount, tiers[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lock.TxID != "tx-001" {
		t.Errorf("expected tx-001, got %s", lock.TxID)
	}
	if lock.Layer != CurrencyL1 {
		t.Errorf("expected CurrencyL1, got %s", lock.Layer)
	}
	if lock.Status != TxPending {
		t.Errorf("expected pending, got %s", lock.Status)
	}
	if lock.SchemaVersion != SchemaVersion {
		t.Errorf("expected %s, got %s", SchemaVersion, lock.SchemaVersion)
	}
	if lock.LockDuration != 30 {
		t.Errorf("expected 30 day lock, got %d", lock.LockDuration)
	}
	if lock.TierName != "Tier 1" {
		t.Errorf("expected Tier 1, got %s", lock.TierName)
	}
}

func TestNewTokenLock_BelowMinimum(t *testing.T) {
	tiers := DefaultStakingTiers()
	amount := big.NewInt(100) // way below 100 ECHO minimum

	_, err := NewTokenLock("tx-002", "did:dag:sender1", amount, tiers[0])
	if err == nil {
		t.Fatal("expected error for amount below minimum")
	}
}

func TestTokenLock_IsLocked(t *testing.T) {
	tiers := DefaultStakingTiers()
	amount := big.NewInt(10000000000)

	lock, _ := NewTokenLock("tx-003", "did:dag:sender1", amount, tiers[0])
	if !lock.IsLocked() {
		t.Error("newly created lock should be locked")
	}

	// Simulate expired lock
	lock.UnlocksAt = time.Now().Add(-1 * time.Hour)
	if lock.IsLocked() {
		t.Error("expired lock should not be locked")
	}
}

func TestDefaultStakingTiers_Count(t *testing.T) {
	tiers := DefaultStakingTiers()
	if len(tiers) != 5 {
		t.Errorf("expected 5 tiers, got %d", len(tiers))
	}
}

func TestDefaultStakingTiers_APRRange(t *testing.T) {
	tiers := DefaultStakingTiers()
	for _, tier := range tiers {
		if tier.APRPercent < 5.0 || tier.APRPercent > 15.0 {
			t.Errorf("tier %s APR %f outside 5-15%% range", tier.Name, tier.APRPercent)
		}
	}
}

func TestDefaultStakingTiers_AscendingMinimums(t *testing.T) {
	tiers := DefaultStakingTiers()
	for i := 1; i < len(tiers); i++ {
		if tiers[i].MinimumStake.Cmp(tiers[i-1].MinimumStake) <= 0 {
			t.Errorf("tier %d minimum should exceed tier %d", i, i-1)
		}
	}
}

// --- StakeDelegation Tests ---

func TestNewStakeDelegation(t *testing.T) {
	amount := big.NewInt(10000000000)
	del := NewStakeDelegation("tx-010", "did:dag:user1", "tx-001", "did:dag:validator1", amount)

	if del.TokenLockTxID != "tx-001" {
		t.Errorf("expected token lock tx-001, got %s", del.TokenLockTxID)
	}
	if del.ValidatorDID != "did:dag:validator1" {
		t.Errorf("expected validator DID did:dag:validator1, got %s", del.ValidatorDID)
	}
	if del.DelegatedStake.Cmp(amount) != 0 {
		t.Errorf("delegated stake mismatch")
	}
	if del.Layer != CurrencyL1 {
		t.Errorf("expected CurrencyL1")
	}
}

// --- WithdrawLock Tests ---

func TestNewWithdrawLock_DefaultCooldown(t *testing.T) {
	amount := big.NewInt(5000000000)
	wl := NewWithdrawLock("tx-020", "did:dag:user1", "tx-001", amount)

	if wl.CooldownDays != DefaultCooldownDays {
		t.Errorf("expected %d day cooldown, got %d", DefaultCooldownDays, wl.CooldownDays)
	}
	if wl.IsAvailable() {
		t.Error("newly created withdraw should not be available yet")
	}
}

func TestWithdrawLock_IsAvailable_AfterCooldown(t *testing.T) {
	amount := big.NewInt(5000000000)
	wl := NewWithdrawLock("tx-021", "did:dag:user1", "tx-001", amount)

	// Simulate cooldown elapsed
	wl.AvailableAt = time.Now().Add(-1 * time.Hour)
	if !wl.IsAvailable() {
		t.Error("should be available after cooldown")
	}
}

// --- AllowSpend Tests ---

func TestNewAllowSpend_Valid(t *testing.T) {
	maxAmount := big.NewInt(100000000) // 1 ECHO
	expiresAt := time.Now().Add(24 * time.Hour)

	as, err := NewAllowSpend("tx-030", "did:dag:owner", "did:dag:spender", "subscription", maxAmount, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if as.Purpose != "subscription" {
		t.Errorf("expected subscription, got %s", as.Purpose)
	}
	if as.SpentSoFar.Sign() != 0 {
		t.Error("spent_so_far should be 0 initially")
	}
}

func TestNewAllowSpend_PastExpiry(t *testing.T) {
	maxAmount := big.NewInt(100000000)
	expiresAt := time.Now().Add(-1 * time.Hour) // in the past

	_, err := NewAllowSpend("tx-031", "did:dag:owner", "did:dag:spender", "bot_payment", maxAmount, expiresAt)
	if err == nil {
		t.Fatal("expected error for past expiry")
	}
}

func TestNewAllowSpend_ZeroAmount(t *testing.T) {
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := NewAllowSpend("tx-032", "did:dag:owner", "did:dag:spender", "test", big.NewInt(0), expiresAt)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestAllowSpend_CanSpend(t *testing.T) {
	maxAmount := big.NewInt(100000000) // 1 ECHO
	expiresAt := time.Now().Add(24 * time.Hour)
	as, _ := NewAllowSpend("tx-033", "did:dag:owner", "did:dag:spender", "subscription", maxAmount, expiresAt)

	// Can spend full amount
	if !as.CanSpend(big.NewInt(100000000)) {
		t.Error("should be able to spend full allowance")
	}

	// Can spend partial
	if !as.CanSpend(big.NewInt(50000000)) {
		t.Error("should be able to spend partial amount")
	}

	// Cannot spend more than max
	if as.CanSpend(big.NewInt(200000000)) {
		t.Error("should not be able to exceed max amount")
	}

	// After partial spend, remaining is reduced
	as.SpentSoFar = big.NewInt(80000000)
	remaining := as.RemainingAllowance()
	if remaining.Cmp(big.NewInt(20000000)) != 0 {
		t.Errorf("expected 20000000 remaining, got %s", remaining.String())
	}
}

func TestAllowSpend_Expired(t *testing.T) {
	maxAmount := big.NewInt(100000000)
	expiresAt := time.Now().Add(24 * time.Hour)
	as, _ := NewAllowSpend("tx-034", "did:dag:owner", "did:dag:spender", "subscription", maxAmount, expiresAt)

	// Simulate expiry
	as.ExpiresAt = time.Now().Add(-1 * time.Second)
	if !as.IsExpired() {
		t.Error("should be expired")
	}
	if as.CanSpend(big.NewInt(1)) {
		t.Error("cannot spend after expiry")
	}
	remaining := as.RemainingAllowance()
	if remaining.Sign() != 0 {
		t.Error("remaining should be 0 after expiry")
	}
}

// --- FeeTransaction Tests ---

func TestNewFeeTransaction(t *testing.T) {
	feeAmount := big.NewInt(1000000) // micro DAG amount
	ft := NewFeeTransaction("tx-040", "did:dag:treasury", "snapshot-123", feeAmount)

	if ft.TreasuryDID != "did:dag:treasury" {
		t.Errorf("expected treasury DID, got %s", ft.TreasuryDID)
	}
	if ft.SnapshotRef != "snapshot-123" {
		t.Errorf("expected snapshot-123, got %s", ft.SnapshotRef)
	}
	if ft.FeeAmountDAG.Cmp(feeAmount) != 0 {
		t.Error("fee amount mismatch")
	}
}

// --- AtomicAction Tests ---

func TestNewAtomicAction_Valid(t *testing.T) {
	ops := []AtomicOperation{
		{Type: OpRewardClaim, Layer: CurrencyL1, Payload: []byte(`{"amount":100}`)},
		{Type: OpTrustVerification, Layer: DataL1, Payload: []byte(`{"tier":3}`)},
		{Type: OpDailyCapUpdate, Layer: CurrencyL1, Payload: []byte(`{"remaining":5}`)},
	}

	aa, err := NewAtomicAction("tx-050", "did:dag:user1", CurrencyL1, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aa.Operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(aa.Operations))
	}
}

func TestNewAtomicAction_Empty(t *testing.T) {
	_, err := NewAtomicAction("tx-051", "did:dag:user1", CurrencyL1, []AtomicOperation{})
	if err == nil {
		t.Fatal("expected error for empty operations")
	}
}

// --- Merkle Commitment Tests ---

func TestComputeMerkleRoot_Single(t *testing.T) {
	root := ComputeMerkleRoot([]string{"abc123"})
	if root != "abc123" {
		t.Errorf("single item should return itself, got %s", root)
	}
}

func TestComputeMerkleRoot_Empty(t *testing.T) {
	root := ComputeMerkleRoot([]string{})
	if root != "" {
		t.Errorf("empty should return empty, got %s", root)
	}
}

func TestComputeMerkleRoot_Multiple(t *testing.T) {
	commitments := []string{"hash1", "hash2", "hash3", "hash4"}
	root := ComputeMerkleRoot(commitments)
	if root == "" {
		t.Error("root should not be empty for multiple commitments")
	}

	// Same input should produce same root (deterministic)
	root2 := ComputeMerkleRoot([]string{"hash1", "hash2", "hash3", "hash4"})
	if root != root2 {
		t.Error("merkle root should be deterministic")
	}

	// Different input should produce different root
	root3 := ComputeMerkleRoot([]string{"hash1", "hash2", "hash3", "hash5"})
	if root == root3 {
		t.Error("different inputs should produce different roots")
	}
}

func TestNewMerkleCommitment(t *testing.T) {
	commitments := []string{"h1", "h2", "h3"}
	mc := NewMerkleCommitment("tx-060", "did:dag:relay1", commitments)

	if mc.CommitmentCount != 3 {
		t.Errorf("expected 3 commitments, got %d", mc.CommitmentCount)
	}
	if mc.MerkleRoot == "" {
		t.Error("merkle root should not be empty")
	}
	if mc.BatchHash == "" {
		t.Error("batch hash should not be empty")
	}
	if mc.Layer != DataL1 {
		t.Errorf("expected DataL1, got %s", mc.Layer)
	}
}

// --- TrustCommitment Tests ---

func TestComputeTrustCommitment(t *testing.T) {
	hash1 := ComputeTrustCommitment("85", "nonce123")
	hash2 := ComputeTrustCommitment("85", "nonce123")
	hash3 := ComputeTrustCommitment("85", "nonce456")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different nonces should produce different hashes")
	}
	if len(hash1) != 64 { // SHA-256 hex = 64 chars
		t.Errorf("expected 64-char hex hash, got %d chars", len(hash1))
	}
}
