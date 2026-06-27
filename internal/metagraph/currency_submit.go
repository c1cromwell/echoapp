package metagraph

import "context"

// TokenLockUpdate is the Currency L1 JSON body (matches Scala TokenLockUpdate).
type TokenLockUpdate struct {
	Amount   int64  `json:"amount"`
	TierName string `json:"tierName"`
	LockDays int    `json:"lockDays"`
}

// WithdrawLockUpdate is the Currency L1 unstake initiation payload.
type WithdrawLockUpdate struct {
	TokenLockTxID string `json:"tokenLockTxId"`
	Amount        int64  `json:"amount"`
}

// StakeDelegationUpdate delegates a lock to a validator.
type StakeDelegationUpdate struct {
	TokenLockTxID string `json:"tokenLockTxId"`
	ValidatorDid  string `json:"validatorDid"`
	Amount        int64  `json:"amount"`
}

// SubmitTokenLock posts a TokenLockUpdate to Currency L1.
func (c *MetagraphClient) SubmitTokenLock(ctx context.Context, update TokenLockUpdate) (string, error) {
	return c.guarded(func() (string, error) {
		return c.submitTransaction(ctx, c.config.CurrencyL1URL+"/transactions", update)
	})
}

// SubmitWithdrawLock posts a WithdrawLockUpdate to Currency L1.
func (c *MetagraphClient) SubmitWithdrawLock(ctx context.Context, update WithdrawLockUpdate) (string, error) {
	return c.guarded(func() (string, error) {
		return c.submitTransaction(ctx, c.config.CurrencyL1URL+"/transactions", update)
	})
}

// SubmitStakeDelegation posts a StakeDelegationUpdate to Currency L1.
func (c *MetagraphClient) SubmitStakeDelegation(ctx context.Context, update StakeDelegationUpdate) (string, error) {
	return c.guarded(func() (string, error) {
		return c.submitTransaction(ctx, c.config.CurrencyL1URL+"/transactions", update)
	})
}
