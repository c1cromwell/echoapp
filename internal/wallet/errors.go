package wallet

import "errors"

var (
	ErrInvalidTier         = errors.New("invalid staking tier")
	ErrInsufficientBalance = errors.New("insufficient available balance")
	ErrStakeNotFound       = errors.New("staking position not found")
	ErrValidatorNotFound   = errors.New("validator not found")
	ErrNoPendingRewards    = errors.New("no pending rewards to claim")
	ErrVestingNotFound     = errors.New("no vesting position for this DID")

	ErrSignedSubmitUnavailable = errors.New("signed submission requires a Currency L1 connection")
	ErrSignedSubmitFailed      = errors.New("signed transaction submission failed")
)
