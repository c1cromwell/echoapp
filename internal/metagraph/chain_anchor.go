package metagraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/thechadcromwell/echoapp/internal/besu"
)

// DataSubmitter writes data-layer updates to Echo's chain and returns a reference to the write.
// The chain accepts 32-byte hashes only: Merkle roots, trust commitments, and the digest of any
// other record (group metadata, governance votes).
type DataSubmitter interface {
	SubmitDataL1(ctx context.Context, tx interface{}) (string, error)
}

// SubmitterFunc adapts a function to DataSubmitter, for tests and small adapters.
type SubmitterFunc func(ctx context.Context, tx interface{}) (string, error)

// SubmitDataL1 implements DataSubmitter.
func (f SubmitterFunc) SubmitDataL1(ctx context.Context, tx interface{}) (string, error) {
	return f(ctx, tx)
}

// RootProver exports the chain's evidence that a root is anchored.
type RootProver interface {
	ProveRoot(ctx context.Context, root string) (ChainAnchor, error)
}

// ErrUnsupportedUpdate means the update type has no on-chain form yet.
var ErrUnsupportedUpdate = errors.New("chain: this update type is not anchored on-chain yet")

// BesuAnchor adapts the Besu client to the gateway's anchoring interfaces.
type BesuAnchor struct {
	Client *besu.Client
}

// NewBesuAnchor wraps a Besu client.
func NewBesuAnchor(client *besu.Client) *BesuAnchor { return &BesuAnchor{Client: client} }

// SubmitDataL1 implements DataSubmitter. It returns once the write is final, with the
// transaction hash as the reference.
func (a *BesuAnchor) SubmitDataL1(ctx context.Context, tx interface{}) (string, error) {
	if a == nil || a.Client == nil {
		return "", errors.New("chain: not configured")
	}
	var (
		receipt besu.Receipt
		err     error
	)
	switch u := tx.(type) {
	case DataL1MerkleRootUpdate:
		receipt, err = a.Client.AnchorRoot(ctx, u.Root, u.LeafCount)
	case *DataL1MerkleRootUpdate:
		receipt, err = a.Client.AnchorRoot(ctx, u.Root, u.LeafCount)
	case DataL1TrustCommitmentUpdate:
		receipt, err = a.Client.AnchorCommitment(ctx, u.Commitment, u.Epoch)
	case *DataL1TrustCommitmentUpdate:
		receipt, err = a.Client.AnchorCommitment(ctx, u.Commitment, u.Epoch)
	case GroupMetadataUpdate:
		return a.attest(ctx, u.Type, u)
	case GovernanceVoteAnchor:
		return a.attest(ctx, u.Type, u)
	case map[string]interface{}:
		kind, _ := u["type"].(string)
		return a.attest(ctx, kind, u)
	default:
		return "", fmt.Errorf("%w (%T)", ErrUnsupportedUpdate, tx)
	}
	if err != nil {
		return "", err
	}
	return receipt.TxHash.Hex(), nil
}

// ProveRoot implements RootProver.
func (a *BesuAnchor) ProveRoot(ctx context.Context, root string) (ChainAnchor, error) {
	if a == nil || a.Client == nil {
		return ChainAnchor{}, errors.New("chain: not configured")
	}
	status, err := a.Client.RootStatus(ctx, root)
	if err != nil {
		return ChainAnchor{}, err
	}
	proof, _, err := a.Client.ProveRoot(ctx, root)
	if err != nil {
		return ChainAnchor{}, err
	}
	raw, err := json.Marshal(proof)
	if err != nil {
		return ChainAnchor{}, err
	}
	return ChainAnchor{BlockNumber: int64(status.BlockNumber), Contract: a.Client.AnchorContract().Hex(), Proof: raw}, nil
}

// AnchorRoot submits a Merkle root and records it durably together with the chain's proof, so
// that GET /v1/anchors/{root}/proof can serve a bundle that verifies with the chain switched off.
// A failure to record or prove after a successful submit is logged, not returned: the root is on
// the chain, and BackfillChainProofs completes the record later.
func AnchorRoot(ctx context.Context, submitter DataSubmitter, store RootStore, network, root string, leafCount int) (string, error) {
	if submitter == nil {
		return "", errors.New("chain: not configured")
	}
	txRef, err := submitter.SubmitDataL1(ctx, DataL1MerkleRootUpdate{Root: root, LeafCount: leafCount})
	if err != nil {
		return "", err
	}
	if store == nil {
		return txRef, nil
	}
	now := time.Now()
	if err := store.RecordRootSubmitted(ctx, root, leafCount, network, now); err != nil {
		log.Printf("anchoring: record root %s failed: %v", root, err)
		return txRef, nil
	}
	// The submitter returns only after the write is final (QBFT has no forks).
	if err := store.MarkRootFinalized(ctx, root, now); err != nil {
		log.Printf("anchoring: mark root %s final failed: %v", root, err)
	}
	if prover, ok := submitter.(RootProver); ok {
		if err := recordChainProof(ctx, prover, store, root, txRef); err != nil {
			log.Printf("anchoring: chain proof for root %s not recorded yet: %v", root, err)
		}
	}
	return txRef, nil
}

func recordChainProof(ctx context.Context, prover RootProver, store RootStore, root, txRef string) error {
	anchor, err := prover.ProveRoot(ctx, root)
	if err != nil {
		return err
	}
	anchor.TxHash = txRef
	return store.SaveChainAnchor(ctx, root, anchor)
}

// BackfillChainProofs completes records whose chain proof could not be taken when the root was
// anchored (for example the gateway stopped between the two steps). The storage entry on the
// chain is permanent, so a proof can be taken at any later time. Runs until ctx is cancelled.
func BackfillChainProofs(ctx context.Context, prover RootProver, store RootStore, every time.Duration) {
	if prover == nil || store == nil {
		return
	}
	if every <= 0 {
		every = time.Minute
	}
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			backfillOnce(ctx, prover, store)
		}
	}
}

func backfillOnce(ctx context.Context, prover RootProver, store RootStore) {
	roots, err := store.ListIncompleteRoots(ctx, 200)
	if err != nil {
		log.Printf("anchoring: list incomplete roots: %v", err)
		return
	}
	for _, r := range roots {
		txRef := ""
		if r.Chain != nil {
			txRef = r.Chain.TxHash
		}
		if err := recordChainProof(ctx, prover, store, r.MerkleRoot, txRef); err != nil {
			if !errors.Is(err, besu.ErrNotAnchored) {
				log.Printf("anchoring: backfill proof for root %s: %v", r.MerkleRoot, err)
			}
			continue
		}
		if !r.Finalized {
			if err := store.MarkRootFinalized(ctx, r.MerkleRoot, time.Now()); err != nil {
				log.Printf("anchoring: mark root %s final failed: %v", r.MerkleRoot, err)
			}
		}
	}
}

// BesuAnchorFromEnv builds the chain anchor from the environment:
//
//	BESU_RPC_URLS             comma-separated JSON-RPC endpoints of Echo's validators
//	BESU_ANCHOR_CONTRACT      address of the EchoAnchor contract
//	BESU_IDENTITY_CONTRACT    address of the EchoIdentity contract (optional: identity records)
//	IDENTITY_SERVICE_KEY_PEM  path to the gateway's secp256k1 signing key
//
// It returns (nil, nil) when BESU_RPC_URLS is unset: the chain is optional in development, and
// anchoring is then off. Any other missing or malformed setting is an error, so a half-configured
// chain stops startup instead of silently not anchoring.
func BesuAnchorFromEnv() (*BesuAnchor, error) {
	urls := strings.TrimSpace(os.Getenv("BESU_RPC_URLS"))
	if urls == "" {
		return nil, nil
	}
	contract := strings.TrimSpace(os.Getenv("BESU_ANCHOR_CONTRACT"))
	if !common.IsHexAddress(contract) {
		return nil, fmt.Errorf("BESU_ANCHOR_CONTRACT %q is not an address", contract)
	}
	keyPath := strings.TrimSpace(os.Getenv("IDENTITY_SERVICE_KEY_PEM"))
	if keyPath == "" {
		return nil, errors.New("IDENTITY_SERVICE_KEY_PEM is not set; the gateway cannot sign chain writes")
	}
	key, err := besu.LoadKeyFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("IDENTITY_SERVICE_KEY_PEM: %w", err)
	}
	cfg := besu.Config{RPCURLs: strings.Split(urls, ","), AnchorContract: common.HexToAddress(contract), Signer: key}
	if identity := strings.TrimSpace(os.Getenv("BESU_IDENTITY_CONTRACT")); identity != "" {
		if !common.IsHexAddress(identity) {
			return nil, fmt.Errorf("BESU_IDENTITY_CONTRACT %q is not an address", identity)
		}
		cfg.IdentityContract = common.HexToAddress(identity)
	}
	client, err := besu.New(cfg)
	if err != nil {
		return nil, err
	}
	return NewBesuAnchor(client), nil
}
