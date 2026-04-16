// Package metagraph defines Tessellation v3 transaction type models
// for the ECHO metagraph on Constellation's Hypergraph.
//
// These types map directly to native Tessellation v3 primitives to ensure
// interoperability with Stargazer wallet, DAG Explorer, PacaSwap DEX,
// and cross-chain bridges.
package metagraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"time"
)

// L1Layer identifies which L1 layer processes a transaction.
type L1Layer string

const (
	CurrencyL1 L1Layer = "currency_l1"
	DataL1     L1Layer = "data_l1"
)

// TxStatus represents the lifecycle state of a metagraph transaction.
type TxStatus string

const (
	TxPending   TxStatus = "pending"
	TxValidated TxStatus = "validated"
	TxFinalized TxStatus = "finalized"
	TxRejected  TxStatus = "rejected"
)

// SchemaVersion tracks the data schema for L1 validation.
// Validators support current and current-1.
const SchemaVersion = "3.2.0"

// BaseTx contains fields common to all Tessellation v3 transaction types.
type BaseTx struct {
	TxID          string    `json:"tx_id"`
	SchemaVersion string    `json:"schema_version"`
	SenderDID     string    `json:"sender_did"`
	Layer         L1Layer   `json:"layer"`
	Status        TxStatus  `json:"status"`
	SnapshotHash  string    `json:"snapshot_hash,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	FinalizedAt   time.Time `json:"finalized_at,omitempty"`
}

// --- TokenLock ---

// TokenLock represents locking ECHO tokens in the user's own Stargazer wallet
// for staking. Tokens remain in the user's custody.
type TokenLock struct {
	BaseTx
	Amount       *big.Int  `json:"amount"`
	LockDuration int       `json:"lock_duration_days"`
	UnlocksAt    time.Time `json:"unlocks_at"`
	TierName     string    `json:"tier_name"`
	MinimumStake *big.Int  `json:"minimum_stake"`
}

// StakingTierV3 defines a v3-compatible staking tier using TokenLock.
type StakingTierV3 struct {
	Name         string   `json:"name"`
	MinimumStake *big.Int `json:"minimum_stake"` // governance-adjustable
	APRPercent   float64  `json:"apr_percent"`   // 5-15%
	LockDays     int      `json:"lock_days"`
}

// DefaultStakingTiers returns the governance-initial staking tiers.
func DefaultStakingTiers() []StakingTierV3 {
	return []StakingTierV3{
		{Name: "Tier 1", MinimumStake: big.NewInt(10000000000), APRPercent: 5.0, LockDays: 30},       // 100 ECHO
		{Name: "Tier 2", MinimumStake: big.NewInt(100000000000), APRPercent: 8.0, LockDays: 90},      // 1,000 ECHO
		{Name: "Tier 3", MinimumStake: big.NewInt(1000000000000), APRPercent: 10.0, LockDays: 180},   // 10,000 ECHO
		{Name: "Tier 4", MinimumStake: big.NewInt(10000000000000), APRPercent: 12.0, LockDays: 270},  // 100,000 ECHO
		{Name: "Tier 5", MinimumStake: big.NewInt(100000000000000), APRPercent: 15.0, LockDays: 365}, // 1,000,000 ECHO
	}
}

// NewTokenLock creates a TokenLock transaction for the given tier.
func NewTokenLock(txID, senderDID string, amount *big.Int, tier StakingTierV3) (*TokenLock, error) {
	if amount.Cmp(tier.MinimumStake) < 0 {
		return nil, errors.New("amount below minimum stake for tier")
	}
	now := time.Now()
	return &TokenLock{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     senderDID,
			Layer:         CurrencyL1,
			Status:        TxPending,
			CreatedAt:     now,
		},
		Amount:       amount,
		LockDuration: tier.LockDays,
		UnlocksAt:    now.AddDate(0, 0, tier.LockDays),
		TierName:     tier.Name,
		MinimumStake: tier.MinimumStake,
	}, nil
}

// IsLocked returns whether the tokens are still within the lock period.
func (tl *TokenLock) IsLocked() bool {
	return time.Now().Before(tl.UnlocksAt)
}

// --- StakeDelegation ---

// StakeDelegation delegates locked ECHO to an L1 validator.
// Delegators earn proportional emission rewards. Instant re-delegation (no cooldown).
type StakeDelegation struct {
	BaseTx
	TokenLockTxID  string   `json:"token_lock_tx_id"`
	ValidatorDID   string   `json:"validator_did"`
	DelegatedStake *big.Int `json:"delegated_stake"`
}

// NewStakeDelegation creates a delegation from a TokenLock to a validator.
func NewStakeDelegation(txID, senderDID, tokenLockTxID, validatorDID string, amount *big.Int) *StakeDelegation {
	return &StakeDelegation{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     senderDID,
			Layer:         CurrencyL1,
			Status:        TxPending,
			CreatedAt:     time.Now(),
		},
		TokenLockTxID:  tokenLockTxID,
		ValidatorDID:   validatorDID,
		DelegatedStake: amount,
	}
}

// --- WithdrawLock ---

// WithdrawLock initiates ECHO unstaking with a 14-day cooldown (governance-adjustable).
type WithdrawLock struct {
	BaseTx
	TokenLockTxID string    `json:"token_lock_tx_id"`
	Amount        *big.Int  `json:"amount"`
	CooldownDays  int       `json:"cooldown_days"`
	AvailableAt   time.Time `json:"available_at"`
}

// DefaultCooldownDays is the initial unstaking cooldown (governance-adjustable).
const DefaultCooldownDays = 14

// NewWithdrawLock initiates an unstaking withdrawal with the standard cooldown.
func NewWithdrawLock(txID, senderDID, tokenLockTxID string, amount *big.Int) *WithdrawLock {
	now := time.Now()
	return &WithdrawLock{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     senderDID,
			Layer:         CurrencyL1,
			Status:        TxPending,
			CreatedAt:     now,
		},
		TokenLockTxID: tokenLockTxID,
		Amount:        amount,
		CooldownDays:  DefaultCooldownDays,
		AvailableAt:   now.AddDate(0, 0, DefaultCooldownDays),
	}
}

// IsAvailable returns whether the cooldown has elapsed.
func (wl *WithdrawLock) IsAvailable() bool {
	return time.Now().After(wl.AvailableAt)
}

// --- AtomicAction ---

// AtomicAction bundles multiple operations as a single all-or-nothing transaction.
// Used for: reward claim + trust verification + cap update,
// governance vote + stake verification, staking tier change + balance update.
type AtomicAction struct {
	BaseTx
	Operations []AtomicOperation `json:"operations"`
}

// AtomicOpType identifies the type of sub-operation within an AtomicAction.
type AtomicOpType string

const (
	OpRewardClaim         AtomicOpType = "reward_claim"
	OpTrustVerification   AtomicOpType = "trust_verification"
	OpDailyCapUpdate      AtomicOpType = "daily_cap_update"
	OpAutoScaleRateUpdate AtomicOpType = "auto_scale_rate_update"
	OpGovernanceVote      AtomicOpType = "governance_vote"
	OpStakeVerification   AtomicOpType = "stake_verification"
	OpBalanceUpdate       AtomicOpType = "balance_update"
)

// AtomicOperation is a single sub-operation within an AtomicAction.
type AtomicOperation struct {
	Type    AtomicOpType    `json:"type"`
	Layer   L1Layer         `json:"layer"`
	Payload json.RawMessage `json:"payload"` // operation-specific data
}

// NewAtomicAction creates an atomic bundle of operations.
func NewAtomicAction(txID, senderDID string, layer L1Layer, ops []AtomicOperation) (*AtomicAction, error) {
	if len(ops) == 0 {
		return nil, errors.New("atomic action requires at least one operation")
	}
	return &AtomicAction{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     senderDID,
			Layer:         layer,
			Status:        TxPending,
			CreatedAt:     time.Now(),
		},
		Operations: ops,
	}, nil
}

// --- AllowSpend ---

// AllowSpend is a time-limited approval for bot/marketplace payments (Phase 5).
// Explicitly time-limited to avoid Ethereum's unlimited approval vulnerability.
type AllowSpend struct {
	BaseTx
	SpenderDID string    `json:"spender_did"`
	MaxAmount  *big.Int  `json:"max_amount"`
	SpentSoFar *big.Int  `json:"spent_so_far"`
	ExpiresAt  time.Time `json:"expires_at"`
	Purpose    string    `json:"purpose"` // "subscription", "bot_payment", "marketplace_escrow"
}

// NewAllowSpend creates a time-limited spend approval.
func NewAllowSpend(txID, ownerDID, spenderDID, purpose string, maxAmount *big.Int, expiresAt time.Time) (*AllowSpend, error) {
	if expiresAt.Before(time.Now()) {
		return nil, errors.New("expiry must be in the future")
	}
	if maxAmount.Sign() <= 0 {
		return nil, errors.New("max amount must be positive")
	}
	return &AllowSpend{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     ownerDID,
			Layer:         CurrencyL1,
			Status:        TxPending,
			CreatedAt:     time.Now(),
		},
		SpenderDID: spenderDID,
		MaxAmount:  maxAmount,
		SpentSoFar: big.NewInt(0),
		ExpiresAt:  expiresAt,
		Purpose:    purpose,
	}, nil
}

// IsExpired checks whether the approval has expired.
func (as *AllowSpend) IsExpired() bool {
	return time.Now().After(as.ExpiresAt)
}

// RemainingAllowance returns the amount still available to spend.
func (as *AllowSpend) RemainingAllowance() *big.Int {
	if as.IsExpired() {
		return big.NewInt(0)
	}
	return new(big.Int).Sub(as.MaxAmount, as.SpentSoFar)
}

// CanSpend checks whether a given amount can be spent under this approval.
func (as *AllowSpend) CanSpend(amount *big.Int) bool {
	if as.IsExpired() {
		return false
	}
	return as.RemainingAllowance().Cmp(amount) >= 0
}

// --- FeeTransaction ---

// FeeTransaction represents automated snapshot fee payment from ECHO treasury DAG reserves.
type FeeTransaction struct {
	BaseTx
	FeeAmountDAG *big.Int `json:"fee_amount_dag"`
	SnapshotRef  string   `json:"snapshot_ref"`
	TreasuryDID  string   `json:"treasury_did"`
}

// NewFeeTransaction creates a fee payment for metagraph snapshot costs.
func NewFeeTransaction(txID, treasuryDID, snapshotRef string, feeAmount *big.Int) *FeeTransaction {
	return &FeeTransaction{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     treasuryDID,
			Layer:         CurrencyL1,
			Status:        TxPending,
			CreatedAt:     time.Now(),
		},
		FeeAmountDAG: feeAmount,
		SnapshotRef:  snapshotRef,
		TreasuryDID:  treasuryDID,
	}
}

// --- MerkleCommitment (Data L1) ---

// MerkleCommitment represents a batch message integrity anchoring on Data L1.
type MerkleCommitment struct {
	BaseTx
	MerkleRoot      string `json:"merkle_root"`
	CommitmentCount int    `json:"commitment_count"`
	BatchHash       string `json:"batch_hash"`
	IPFSCid         string `json:"ipfs_cid,omitempty"`
}

// ComputeMerkleRoot computes a SHA-256 Merkle root from a set of commitment hashes.
func ComputeMerkleRoot(commitments []string) string {
	if len(commitments) == 0 {
		return ""
	}
	if len(commitments) == 1 {
		return commitments[0]
	}

	// Pad to even count
	if len(commitments)%2 != 0 {
		commitments = append(commitments, commitments[len(commitments)-1])
	}

	var nextLevel []string
	for i := 0; i < len(commitments); i += 2 {
		combined := commitments[i] + commitments[i+1]
		hash := sha256.Sum256([]byte(combined))
		nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
	}

	return ComputeMerkleRoot(nextLevel)
}

// NewMerkleCommitment creates a Data L1 Merkle root submission.
func NewMerkleCommitment(txID, senderDID string, commitments []string) *MerkleCommitment {
	root := ComputeMerkleRoot(commitments)
	batchData := ""
	for _, c := range commitments {
		batchData += c
	}
	batchHash := sha256.Sum256([]byte(batchData))

	return &MerkleCommitment{
		BaseTx: BaseTx{
			TxID:          txID,
			SchemaVersion: SchemaVersion,
			SenderDID:     senderDID,
			Layer:         DataL1,
			Status:        TxPending,
			CreatedAt:     time.Now(),
		},
		MerkleRoot:      root,
		CommitmentCount: len(commitments),
		BatchHash:       hex.EncodeToString(batchHash[:]),
	}
}

// --- TrustCommitment (Data L1) ---

// TrustCommitment represents H(trust_score || nonce) anchored on Data L1.
type TrustCommitment struct {
	BaseTx
	CommitmentHash string `json:"commitment_hash"` // H(score || nonce)
	IssuerDID      string `json:"issuer_did"`
	SubjectDID     string `json:"subject_did"`
}

// ComputeTrustCommitment creates H(score || nonce).
func ComputeTrustCommitment(score string, nonce string) string {
	data := score + nonce
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
