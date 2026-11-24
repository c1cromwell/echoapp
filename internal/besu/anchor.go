package besu

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ErrNotAnchored means the chain has no record of the value.
var ErrNotAnchored = errors.New("besu: not anchored")

// parseHash32 accepts exactly 64 hex characters (optionally 0x-prefixed) and rejects the zero hash.
// Everything Echo anchors is a 32-byte hash; nothing else can be written to the contract.
func parseHash32(field, value string) (common.Hash, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "0x")
	if len(raw) != 64 {
		return common.Hash{}, fmt.Errorf("besu: %s must be 64 hex characters, got %d", field, len(raw))
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		return common.Hash{}, fmt.Errorf("besu: %s is not hex: %w", field, err)
	}
	h := common.BytesToHash(b)
	if h == (common.Hash{}) {
		return common.Hash{}, fmt.Errorf("besu: %s must not be zero", field)
	}
	return h, nil
}

// AnchorRoot records a Merkle root. It returns once the root is final.
func (c *Client) AnchorRoot(ctx context.Context, root string, leafCount int) (Receipt, error) {
	h, err := parseHash32("root", root)
	if err != nil {
		return Receipt{}, err
	}
	if leafCount <= 0 || int64(leafCount) > int64(^uint32(0)) {
		return Receipt{}, fmt.Errorf("besu: leaf count %d is out of range", leafCount)
	}
	return c.transact(ctx, c.cfg.AnchorContract, anchorABI, "anchorRoot", h, uint32(leafCount))
}

// AnchorCommitment records a trust commitment for an epoch. It returns once it is final.
func (c *Client) AnchorCommitment(ctx context.Context, commitment string, epoch int64) (Receipt, error) {
	h, err := parseHash32("commitment", commitment)
	if err != nil {
		return Receipt{}, err
	}
	if epoch <= 0 {
		return Receipt{}, errors.New("besu: epoch must be positive")
	}
	return c.transact(ctx, c.cfg.AnchorContract, anchorABI, "anchorCommitment", h, uint64(epoch))
}

// RootStatus is the chain's record of a root.
type RootStatus struct {
	BlockNumber uint64 // block the root was anchored in
	LeafCount   uint32
}

// unpackAnchoredAt splits the contract's packed value: (block number << 32) | leaf count.
func unpackAnchoredAt(packed *big.Int) RootStatus {
	return RootStatus{
		BlockNumber: new(big.Int).Rsh(packed, 32).Uint64(),
		LeafCount:   uint32(new(big.Int).And(packed, big.NewInt(0xffffffff)).Uint64()),
	}
}

// RootStatus reads whether root is anchored. ErrNotAnchored if it is not.
func (c *Client) RootStatus(ctx context.Context, root string) (RootStatus, error) {
	h, err := parseHash32("root", root)
	if err != nil {
		return RootStatus{}, err
	}
	ec, err := c.conn(ctx)
	if err != nil {
		return RootStatus{}, err
	}
	var out []interface{}
	if err := bind.NewBoundContract(c.cfg.AnchorContract, anchorABI, ec, ec, ec).Call(&bind.CallOpts{Context: ctx}, &out, "anchoredAt", h); err != nil {
		return RootStatus{}, fmt.Errorf("besu: read anchoredAt: %w", err)
	}
	packed, ok := out[0].(*big.Int)
	if !ok || packed.Sign() == 0 {
		return RootStatus{}, ErrNotAnchored
	}
	return unpackAnchoredAt(packed), nil
}
