package besu

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

// anchoredAtSlot is the storage slot of EchoAnchor.anchoredAt. It is fixed by the order of the
// contract's state variables; the recorded proof in testdata/ fails to verify if it changes.
const anchoredAtSlot = 2

// Proof shows that a root is recorded in the anchor contract as of one block, in a form that can
// be checked with no access to the chain. Nothing in it is trusted by a verifier: the block header
// carries the validators' signatures, and the two Merkle proofs tie the stored entry to that
// header. The verifier's only input besides the proof is the validator addresses it trusts.
type Proof struct {
	Root         common.Hash     `json:"root"`
	Contract     common.Address  `json:"contract"`
	HeaderRLP    hexutil.Bytes   `json:"headerRlp"`
	AccountProof []hexutil.Bytes `json:"accountProof"`
	StorageProof []hexutil.Bytes `json:"storageProof"`
}

// Verified is what a valid proof establishes.
type Verified struct {
	Root           common.Hash
	Contract       common.Address
	AnchoredBlock  uint64           // block in which the root was first recorded
	LeafCount      uint32           // leaves under the root
	ProvenAtBlock  uint64           // block whose signed header the proof is checked against
	ProvenAtTime   time.Time        // that block's timestamp, as agreed by the validators
	Signers        []common.Address // trusted validators whose commit signature was verified
	TrustedCount   int
	QuorumRequired int
}

func storageKey(root common.Hash) common.Hash {
	return crypto.Keccak256Hash(root[:], common.LeftPadBytes(big.NewInt(anchoredAtSlot).Bytes(), 32))
}

// ProveRoot exports a proof that root is anchored, as of the newest block. The storage entry is
// permanent, so a proof can be taken at any time after anchoring; ErrNotAnchored if there is none.
func (c *Client) ProveRoot(ctx context.Context, root string) (*Proof, uint64, error) {
	h, err := parseHash32("root", root)
	if err != nil {
		return nil, 0, err
	}
	ec, err := c.conn(ctx)
	if err != nil {
		return nil, 0, err
	}
	number, err := ec.BlockNumber(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("besu: block number: %w", err)
	}
	block := hexutil.EncodeUint64(number)

	var header hexutil.Bytes
	if err := ec.Client().CallContext(ctx, &header, "debug_getRawHeader", block); err != nil {
		return nil, 0, fmt.Errorf("besu: debug_getRawHeader: %w", err)
	}
	var res struct {
		AccountProof []hexutil.Bytes `json:"accountProof"`
		StorageProof []struct {
			Value *hexutil.Big    `json:"value"`
			Proof []hexutil.Bytes `json:"proof"`
		} `json:"storageProof"`
	}
	if err := ec.Client().CallContext(ctx, &res, "eth_getProof", c.cfg.AnchorContract, []string{storageKey(h).Hex()}, block); err != nil {
		return nil, 0, fmt.Errorf("besu: eth_getProof: %w", err)
	}
	if len(res.StorageProof) != 1 {
		return nil, 0, errors.New("besu: eth_getProof returned no storage proof")
	}
	if res.StorageProof[0].Value == nil || res.StorageProof[0].Value.ToInt().Sign() == 0 {
		return nil, 0, ErrNotAnchored
	}
	return &Proof{Root: h, Contract: c.cfg.AnchorContract, HeaderRLP: header, AccountProof: res.AccountProof, StorageProof: res.StorageProof[0].Proof}, number, nil
}

// qbftExtra is the RLP layout of a QBFT header's extraData.
type qbftExtra struct {
	Vanity     []byte
	Validators []common.Address
	Vote       rlp.RawValue
	Round      rlp.RawValue
	Seals      [][]byte
}

// commitDigest is what each validator signs to commit a block: the header hash computed with the
// commit seals removed from extraData. The round number stays in.
func commitDigest(header *types.Header, extra qbftExtra) (common.Hash, error) {
	stripped, err := rlp.EncodeToBytes([]interface{}{extra.Vanity, extra.Validators, extra.Vote, extra.Round, [][]byte{}})
	if err != nil {
		return common.Hash{}, err
	}
	h := types.CopyHeader(header)
	h.Extra = stripped
	return h.Hash(), nil
}

// Quorum is the number of validators QBFT needs to commit a block: more than two thirds.
func Quorum(validators int) int { return (2*validators + 2) / 3 }

func proofNodes(nodes []hexutil.Bytes) *memorydb.Database {
	db := memorydb.New()
	for _, n := range nodes {
		_ = db.Put(crypto.Keccak256(n), n)
	}
	return db
}

// Verify checks a proof against the validator addresses the caller trusts. It makes no network
// call. It fails unless:
//  1. a quorum of the trusted validators signed the header, and no one else did;
//  2. the contract's storage is part of that header's state; and
//  3. the root is recorded in that storage.
func Verify(p *Proof, trustedValidators []common.Address) (*Verified, error) {
	if p == nil {
		return nil, errors.New("besu: no chain proof")
	}
	if len(trustedValidators) == 0 {
		return nil, errors.New("besu: no trusted validators given")
	}
	trusted := make(map[common.Address]bool, len(trustedValidators))
	for _, v := range trustedValidators {
		trusted[v] = true
	}

	var header types.Header
	if err := rlp.DecodeBytes(p.HeaderRLP, &header); err != nil {
		return nil, fmt.Errorf("besu: decode header: %w", err)
	}
	var extra qbftExtra
	if err := rlp.DecodeBytes(header.Extra, &extra); err != nil {
		return nil, fmt.Errorf("besu: decode QBFT extra data: %w", err)
	}
	digest, err := commitDigest(&header, extra)
	if err != nil {
		return nil, err
	}
	seen := map[common.Address]bool{}
	var signers []common.Address
	for _, seal := range extra.Seals {
		pub, err := crypto.SigToPub(digest[:], seal)
		if err != nil {
			return nil, fmt.Errorf("besu: invalid commit signature: %w", err)
		}
		signer := crypto.PubkeyToAddress(*pub)
		if !trusted[signer] {
			return nil, fmt.Errorf("besu: block signed by %s, which is not a trusted validator", signer.Hex())
		}
		if !seen[signer] {
			seen[signer] = true
			signers = append(signers, signer)
		}
	}
	quorum := Quorum(len(trusted))
	if len(signers) < quorum {
		return nil, fmt.Errorf("besu: %d of %d trusted validators signed the block; %d are required", len(signers), len(trusted), quorum)
	}

	accountRLP, err := trie.VerifyProof(header.Root, crypto.Keccak256(p.Contract[:]), proofNodes(p.AccountProof))
	if err != nil {
		return nil, fmt.Errorf("besu: account proof does not match the block's state root: %w", err)
	}
	if accountRLP == nil {
		return nil, fmt.Errorf("besu: contract %s does not exist in the block's state", p.Contract.Hex())
	}
	var account struct {
		Nonce       uint64
		Balance     *big.Int
		StorageRoot common.Hash
		CodeHash    []byte
	}
	if err := rlp.DecodeBytes(accountRLP, &account); err != nil {
		return nil, fmt.Errorf("besu: decode contract account: %w", err)
	}

	key := storageKey(p.Root)
	valueRLP, err := trie.VerifyProof(account.StorageRoot, crypto.Keccak256(key[:]), proofNodes(p.StorageProof))
	if err != nil {
		return nil, fmt.Errorf("besu: storage proof does not match the contract's storage: %w", err)
	}
	if valueRLP == nil {
		return nil, fmt.Errorf("besu: root %s is not anchored as of block %d", p.Root.Hex(), header.Number)
	}
	var value []byte
	if err := rlp.DecodeBytes(valueRLP, &value); err != nil {
		return nil, fmt.Errorf("besu: decode stored value: %w", err)
	}
	status := unpackAnchoredAt(new(big.Int).SetBytes(value))
	if status.BlockNumber == 0 || status.BlockNumber > header.Number.Uint64() {
		return nil, fmt.Errorf("besu: stored anchor block %d is not consistent with proof block %d", status.BlockNumber, header.Number)
	}

	return &Verified{
		Root:           p.Root,
		Contract:       p.Contract,
		AnchoredBlock:  status.BlockNumber,
		LeafCount:      status.LeafCount,
		ProvenAtBlock:  header.Number.Uint64(),
		ProvenAtTime:   time.Unix(int64(header.Time), 0).UTC(),
		Signers:        signers,
		TrustedCount:   len(trusted),
		QuorumRequired: quorum,
	}, nil
}
