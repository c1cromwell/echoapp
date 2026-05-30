package recovery

import (
	"fmt"

	"github.com/hashicorp/vault/shamir"
)

const secretLen = 32

// SplitRecoverySecret splits a 32-byte recovery secret into n Shamir shares with threshold m.
func SplitRecoverySecret(secret []byte, parts, threshold int) ([][]byte, error) {
	if len(secret) != secretLen {
		return nil, fmt.Errorf("recovery secret must be %d bytes", secretLen)
	}
	if err := ValidatePolicy(threshold, parts); err != nil {
		return nil, err
	}
	return shamir.Split(secret, parts, threshold)
}

// CombineRecoveryShares reconstructs the recovery secret from at least m shares.
func CombineRecoveryShares(shares [][]byte) ([]byte, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("at least one share required")
	}
	return shamir.Combine(shares)
}
