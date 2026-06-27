package wallet

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// CurrencySubmitter submits Currency L1 updates (optional when L1 URL unset).
type CurrencySubmitter interface {
	SubmitTokenLock(ctx context.Context, update metagraph.TokenLockUpdate) (string, error)
	SubmitWithdrawLock(ctx context.Context, update metagraph.WithdrawLockUpdate) (string, error)
	SubmitStakeDelegation(ctx context.Context, update metagraph.StakeDelegationUpdate) (string, error)
	SubmitCurrencyL1(ctx context.Context, tx metagraph.CurrencyL1Transaction) (string, error)
	QueryValidators(ctx context.Context) ([]metagraph.ValidatorSnapshot, error)
}

// LedgerQuerier implements MetagraphQuerier using PG ledger + optional L1 submit.
type LedgerQuerier struct {
	store     Store
	submitter CurrencySubmitter
}

// NewLedgerQuerier wires the wallet store to optional Currency L1 submissions.
func NewLedgerQuerier(store Store, submitter CurrencySubmitter) *LedgerQuerier {
	return &LedgerQuerier{store: store, submitter: submitter}
}

func (q *LedgerQuerier) GetBalance(ctx context.Context, did string) (*BalanceInfo, error) {
	return q.store.GetBalance(ctx, did)
}

func (q *LedgerQuerier) GetTokenLocks(ctx context.Context, did string) ([]TokenLockPos, error) {
	return q.store.ListLocks(ctx, did)
}

func (q *LedgerQuerier) GetDelegations(ctx context.Context, did string) ([]DelegationPos, error) {
	return q.store.ListDelegations(ctx, did)
}

func (q *LedgerQuerier) GetValidators(ctx context.Context) ([]ValidatorInfo, error) {
	if cached, err := q.store.ListValidators(ctx); err == nil && len(cached) > 0 {
		return cached, nil
	}
	if q.submitter == nil {
		return nil, nil
	}
	snaps, err := q.submitter.QueryValidators(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ValidatorInfo, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, ValidatorInfo{
			ID:             s.ID,
			Address:        s.Address,
			Uptime:         s.UptimePercent,
			Commission:     s.CommissionPercent,
			TotalDelegated: s.TotalDelegated,
			DelegatorCount: s.DelegatorCount,
			Layer:          s.Layer,
		})
	}
	_ = q.store.UpsertValidators(ctx, out)
	return out, nil
}

func (q *LedgerQuerier) SubmitTokenLock(ctx context.Context, did string, amount int64, tier StakingTier) (string, error) {
	chain, err := ResolveChainTier(tier.Name)
	if err != nil {
		return "", err
	}
	if amount < chain.MinDatum {
		return "", ErrInsufficientBalance
	}
	if err := q.store.ApplyStake(ctx, did, amount); err != nil {
		return "", err
	}

	lockID := NewLockID()
	txHash := "local-" + lockID
	if q.submitter != nil {
		if h, err := q.submitter.SubmitTokenLock(ctx, metagraph.TokenLockUpdate{
			Amount:   amount,
			TierName: chain.ScalaName,
			LockDays: chain.LockDays,
		}); err == nil && h != "" {
			txHash = h
			lockID = txHash
		}
	}

	pos := TokenLockPos{
		ID:          lockID,
		Amount:      amount,
		Tier:        tier.Name,
		LockedUntil: LockUntil(chain.LockDays),
	}
	if err := q.store.InsertLock(ctx, did, pos); err != nil {
		return "", err
	}
	return txHash, nil
}

func (q *LedgerQuerier) SubmitStakeDelegation(ctx context.Context, delegatorDID, stakeID, validatorID string, amount int64) (string, error) {
	txHash := "local-del-" + uuid.New().String()
	if q.submitter != nil {
		if h, err := q.submitter.SubmitStakeDelegation(ctx, metagraph.StakeDelegationUpdate{
			TokenLockTxID: stakeID,
			ValidatorDid:  validatorID,
			Amount:        amount,
		}); err == nil && h != "" {
			txHash = h
		}
	}
	return txHash, nil
}

func (q *LedgerQuerier) SubmitWithdrawLock(ctx context.Context, did, stakeID string, amount int64) (string, error) {
	txHash := "local-unstake-" + uuid.New().String()
	if q.submitter != nil {
		if h, err := q.submitter.SubmitWithdrawLock(ctx, metagraph.WithdrawLockUpdate{
			TokenLockTxID: stakeID,
			Amount:        amount,
		}); err == nil && h != "" {
			txHash = h
		}
	}
	if err := q.store.ApplyUnstake(ctx, did, amount); err != nil {
		return "", err
	}
	return txHash, nil
}

func (q *LedgerQuerier) SubmitAtomicRewardClaim(ctx context.Context, did string, claims []RewardClaim) (string, error) {
	var total int64
	for _, c := range claims {
		total += c.Amount
	}
	if total <= 0 {
		return "", ErrNoPendingRewards
	}

	txHash := "local-claim-" + uuid.New().String()
	if q.submitter != nil {
		ops := make([]metagraph.AtomicOperation, 0, len(claims))
		for _, c := range claims {
			payload, err := json.Marshal(map[string]interface{}{
				"rewardType": c.RewardType,
				"amount":     c.Amount,
			})
			if err != nil {
				return "", err
			}
			ops = append(ops, metagraph.AtomicOperation{
				Type:    metagraph.OpRewardClaim,
				Layer:   metagraph.CurrencyL1,
				Payload: payload,
			})
		}
		action, err := metagraph.NewAtomicAction(txHash, did, metagraph.CurrencyL1, ops)
		if err == nil {
			if h, err := q.submitter.SubmitCurrencyL1(ctx, metagraph.CurrencyL1Transaction{
				Type:         "atomic_action",
				AtomicAction: action,
			}); err == nil && h != "" {
				txHash = h
			}
		}
	}

	if err := q.store.CreditRewards(ctx, did, total); err != nil {
		return "", err
	}
	return txHash, nil
}

// Ensure LedgerQuerier implements MetagraphQuerier at compile time.
var _ MetagraphQuerier = (*LedgerQuerier)(nil)
