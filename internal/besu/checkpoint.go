package besu

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// A Checkpoint is one block of Echo's chain, with the validators' signatures on it, written down
// somewhere Echo does not control.
//
// Echo runs every validator, so the validators could in principle agree to replace history. A
// proof checked against today's chain would not notice. A checkpoint published outside Echo
// does: if the chain ever shows a different block at a checkpointed height, history was rewritten
// after that checkpoint was published, and the two signed headers prove it.
type Checkpoint struct {
	ChainID   uint64        `json:"chainId"`
	Number    uint64        `json:"number"`
	Hash      common.Hash   `json:"hash"`
	Time      time.Time     `json:"time"`
	HeaderRLP hexutil.Bytes `json:"headerRlp"`
}

// ErrHistoryRewritten means the chain holds a different block at a checkpoint's height.
var ErrHistoryRewritten = errors.New("besu: the chain's block at this height is not the checkpointed block")

func rawHeader(ctx context.Context, rc *rpc.Client, number uint64) (hexutil.Bytes, error) {
	var header hexutil.Bytes
	if err := rc.CallContext(ctx, &header, "debug_getRawHeader", hexutil.EncodeUint64(number)); err != nil {
		return nil, fmt.Errorf("besu: debug_getRawHeader: %w", err)
	}
	return header, nil
}

// TakeCheckpoint reads the newest block from a validator. With QBFT a block is final as soon as it
// exists, so the newest block is a valid checkpoint.
func TakeCheckpoint(ctx context.Context, rpcURL string) (*Checkpoint, error) {
	rc, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("besu: dial %s: %w", rpcURL, err)
	}
	defer rc.Close()
	var chainID, number hexutil.Uint64
	if err := rc.CallContext(ctx, &chainID, "eth_chainId"); err != nil {
		return nil, fmt.Errorf("besu: eth_chainId: %w", err)
	}
	if err := rc.CallContext(ctx, &number, "eth_blockNumber"); err != nil {
		return nil, fmt.Errorf("besu: eth_blockNumber: %w", err)
	}
	header, err := rawHeader(ctx, rc, uint64(number))
	if err != nil {
		return nil, err
	}
	cp := &Checkpoint{ChainID: uint64(chainID), Number: uint64(number), HeaderRLP: header}
	if err := cp.fillFromHeader(); err != nil {
		return nil, err
	}
	return cp, nil
}

// fillFromHeader sets Hash and Time from the header, so they cannot disagree with it.
func (cp *Checkpoint) fillFromHeader() error {
	header, _, err := decodeHeader(cp.HeaderRLP)
	if err != nil {
		return err
	}
	if header.Number == nil || header.Number.Uint64() != cp.Number {
		return fmt.Errorf("besu: header is for block %v, not %d", header.Number, cp.Number)
	}
	if cp.Hash, err = blockHash(cp.HeaderRLP); err != nil {
		return err
	}
	cp.Time = time.Unix(int64(header.Time), 0).UTC()
	return nil
}

// VerifyCheckpoint checks, with no network call, that a quorum of the trusted validators signed
// the checkpointed block and that the recorded number, hash and time are the header's own.
func VerifyCheckpoint(cp *Checkpoint, trustedValidators []common.Address) ([]common.Address, error) {
	if cp == nil {
		return nil, errors.New("besu: no checkpoint")
	}
	if len(trustedValidators) == 0 {
		return nil, errors.New("besu: no trusted validators given")
	}
	header, signers, err := verifyHeader(cp.HeaderRLP, trustedValidators)
	if err != nil {
		return nil, err
	}
	if header.Number == nil || header.Number.Uint64() != cp.Number {
		return nil, fmt.Errorf("besu: checkpoint says block %d but its header is block %v", cp.Number, header.Number)
	}
	got, err := blockHash(cp.HeaderRLP)
	if err != nil {
		return nil, err
	}
	if got != cp.Hash {
		return nil, fmt.Errorf("besu: checkpoint says hash %s but its header hashes to %s", cp.Hash.Hex(), got.Hex())
	}
	if !cp.Time.Equal(time.Unix(int64(header.Time), 0)) {
		return nil, errors.New("besu: checkpoint time is not the header's time")
	}
	return signers, nil
}

// CompareWithChain asks a validator for its block at the checkpoint's height. It returns
// ErrHistoryRewritten if that block is not the checkpointed one.
func CompareWithChain(ctx context.Context, cp *Checkpoint, rpcURL string) error {
	rc, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("besu: dial %s: %w", rpcURL, err)
	}
	defer rc.Close()
	var chainID hexutil.Uint64
	if err := rc.CallContext(ctx, &chainID, "eth_chainId"); err != nil {
		return fmt.Errorf("besu: eth_chainId: %w", err)
	}
	if uint64(chainID) != cp.ChainID {
		return fmt.Errorf("besu: checkpoint is for chain %d, this node is chain %d", cp.ChainID, uint64(chainID))
	}
	// Ask for the block by number and read the hash the node itself reports for it.
	var block *struct {
		Hash common.Hash `json:"hash"`
	}
	if err := rc.CallContext(ctx, &block, "eth_getBlockByNumber", hexutil.EncodeUint64(cp.Number), false); err != nil {
		return fmt.Errorf("besu: eth_getBlockByNumber: %w", err)
	}
	if block == nil {
		return fmt.Errorf("%w: the chain has no block %d", ErrHistoryRewritten, cp.Number)
	}
	if block.Hash != cp.Hash {
		return fmt.Errorf("%w: block %d is %s on the chain, %s in the checkpoint", ErrHistoryRewritten, cp.Number, block.Hash.Hex(), cp.Hash.Hex())
	}
	return nil
}
