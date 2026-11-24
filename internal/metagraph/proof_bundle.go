package metagraph

import (
	"context"
	"errors"
	"time"
)

// ProofBundleVersion is bumped when the bundle layout changes.
const ProofBundleVersion = 1

// ErrRootNotFound means no anchored-root record exists for the root.
var ErrRootNotFound = errors.New("anchored root not found")

// ErrMessageNotInRoot means the message has no proof under the requested root.
var ErrMessageNotInRoot = errors.New("no proof for this message under this root")

// ProofBundle is the self-contained anchor proof served by
// GET /v1/anchors/{root}/proof and checked offline by cmd/verifyproof.
// It is built only from the gateway's own durable record; the chain is not
// contacted. See FinalityTracker for what the snapshot fields mean.
type ProofBundle struct {
	Version     int        `json:"version"`
	Network     string     `json:"network"`
	MerkleRoot  string     `json:"merkleRoot"`
	LeafCount   int        `json:"leafCount"`
	SubmittedAt time.Time  `json:"submittedAt"`
	Finalized   bool       `json:"finalized"`
	FinalizedAt *time.Time `json:"finalizedAt,omitempty"`
	// MetagraphSnapshot / GlobalSnapshot are the signed snapshots captured when
	// finality was observed. Signers are the peer ids listed in the snapshot's
	// proofs; their signatures were not verified by the gateway.
	MetagraphSnapshot *SnapshotRef        `json:"metagraphSnapshot,omitempty"`
	GlobalSnapshot    *SnapshotRef        `json:"globalSnapshot,omitempty"`
	Message           *ProofBundleMessage `json:"message,omitempty"`
}

// ProofBundleMessage is one message's Merkle path to the bundle's root.
type ProofBundleMessage struct {
	MessageID  string   `json:"messageId"`
	Commitment string   `json:"commitment"`
	Siblings   []string `json:"siblings"`
	LeafIndex  int      `json:"leafIndex"`
}

// BuildProofBundle assembles a bundle for root. fallbackNetwork is used only
// for roots recorded without a network. messageID is optional.
func BuildProofBundle(ctx context.Context, store AnchorStore, fallbackNetwork, root, messageID string) (*ProofBundle, error) {
	rec, ok, err := store.GetRoot(ctx, root)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRootNotFound
	}
	network := rec.Network
	if network == "" {
		network = fallbackNetwork
	}
	b := &ProofBundle{
		Version:           ProofBundleVersion,
		Network:           network,
		MerkleRoot:        rec.MerkleRoot,
		LeafCount:         rec.LeafCount,
		SubmittedAt:       rec.SubmittedAt,
		Finalized:         rec.Finalized,
		FinalizedAt:       rec.FinalizedAt,
		MetagraphSnapshot: rec.ML0,
		GlobalSnapshot:    rec.GL0,
	}
	if messageID != "" {
		p, ok, err := store.Get(ctx, messageID)
		if err != nil {
			return nil, err
		}
		if !ok || p.MerkleRoot != root {
			return nil, ErrMessageNotInRoot
		}
		siblings := p.Siblings
		if siblings == nil {
			siblings = []string{}
		}
		b.Message = &ProofBundleMessage{
			MessageID:  p.MessageID,
			Commitment: p.Commitment,
			Siblings:   siblings,
			LeafIndex:  p.MerkleLeafIndex,
		}
	}
	return b, nil
}
