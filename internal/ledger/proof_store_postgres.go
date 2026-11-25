package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresProofStore is the durable AnchorStore (migrations 036 and 037).
// Proofs and finality survive gateway restarts and do not depend on the chain
// being reachable.
type PostgresProofStore struct {
	pool *pgxpool.Pool
}

// NewPostgresProofStore wraps a pgx pool.
func NewPostgresProofStore(pool *pgxpool.Pool) *PostgresProofStore {
	return &PostgresProofStore{pool: pool}
}

// Put upserts a per-message proof.
func (s *PostgresProofStore) Put(ctx context.Context, proof MessageAnchorProof) error {
	siblings := proof.Siblings
	if siblings == nil {
		siblings = []string{}
	}
	sib, err := json.Marshal(siblings)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO anchor_message_proofs (message_id, commitment, siblings, leaf_index, merkle_root, submit_tx_ref)
		VALUES ($1, $2, $3::jsonb, $4, $5, $6)
		ON CONFLICT (message_id) DO UPDATE SET
			commitment = EXCLUDED.commitment,
			siblings = EXCLUDED.siblings,
			leaf_index = EXCLUDED.leaf_index,
			merkle_root = EXCLUDED.merkle_root,
			submit_tx_ref = EXCLUDED.submit_tx_ref`,
		proof.MessageID, proof.Commitment, string(sib), proof.MerkleLeafIndex, proof.MerkleRoot, proof.SnapshotHash)
	return err
}

// Get loads a proof by message ID.
func (s *PostgresProofStore) Get(ctx context.Context, messageID string) (MessageAnchorProof, bool, error) {
	p := MessageAnchorProof{MessageID: messageID}
	var sib []byte
	err := s.pool.QueryRow(ctx, `
		SELECT commitment, siblings, leaf_index, merkle_root, submit_tx_ref
		FROM anchor_message_proofs WHERE message_id = $1`, messageID).
		Scan(&p.Commitment, &sib, &p.MerkleLeafIndex, &p.MerkleRoot, &p.SnapshotHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return MessageAnchorProof{}, false, nil
	}
	if err != nil {
		return MessageAnchorProof{}, false, err
	}
	if err := json.Unmarshal(sib, &p.Siblings); err != nil {
		return MessageAnchorProof{}, false, err
	}
	return p, true, nil
}

// RecordRootSubmitted implements RootStore.
func (s *PostgresProofStore) RecordRootSubmitted(ctx context.Context, root string, leafCount int, network string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO anchor_roots (merkle_root, network, leaf_count, submitted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (merkle_root) DO NOTHING`, root, network, leafCount, at)
	return err
}

const anchorRootColumns = `merkle_root, network, leaf_count, submitted_at, finalized, finalized_at,
	block_number, tx_hash, contract, chain_proof`

func scanAnchoredRoot(row pgx.Row) (AnchoredRoot, error) {
	var (
		r                AnchoredRoot
		block            *int64
		txHash, contract *string
		proof            []byte
	)
	if err := row.Scan(&r.MerkleRoot, &r.Network, &r.LeafCount, &r.SubmittedAt, &r.Finalized, &r.FinalizedAt,
		&block, &txHash, &contract, &proof); err != nil {
		return AnchoredRoot{}, err
	}
	if block != nil {
		r.Chain = &ChainAnchor{BlockNumber: *block, Proof: proof}
		if txHash != nil {
			r.Chain.TxHash = *txHash
		}
		if contract != nil {
			r.Chain.Contract = *contract
		}
	}
	return r, nil
}

// GetRoot implements RootStore.
func (s *PostgresProofStore) GetRoot(ctx context.Context, root string) (AnchoredRoot, bool, error) {
	r, err := scanAnchoredRoot(s.pool.QueryRow(ctx,
		`SELECT `+anchorRootColumns+` FROM anchor_roots WHERE merkle_root = $1`, root))
	if errors.Is(err, pgx.ErrNoRows) {
		return AnchoredRoot{}, false, nil
	}
	if err != nil {
		return AnchoredRoot{}, false, err
	}
	return r, true, nil
}

// ListIncompleteRoots implements RootStore.
func (s *PostgresProofStore) ListIncompleteRoots(ctx context.Context, limit int) ([]AnchoredRoot, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx, `SELECT `+anchorRootColumns+` FROM anchor_roots
		WHERE NOT finalized OR chain_proof IS NULL
		ORDER BY submitted_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AnchoredRoot
	for rows.Next() {
		r, err := scanAnchoredRoot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// MarkRootFinalized implements RootStore.
func (s *PostgresProofStore) MarkRootFinalized(ctx context.Context, root string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE anchor_roots SET finalized = TRUE, finalized_at = COALESCE(finalized_at, $2)
		WHERE merkle_root = $1`, root, at)
	return err
}

// SaveChainAnchor implements RootStore.
func (s *PostgresProofStore) SaveChainAnchor(ctx context.Context, root string, anchor ChainAnchor) error {
	var proof *string
	if len(anchor.Proof) > 0 {
		p := string(anchor.Proof)
		proof = &p
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE anchor_roots SET block_number = $2, tx_hash = $3, contract = $4, chain_proof = $5::jsonb
		WHERE merkle_root = $1`, root, anchor.BlockNumber, anchor.TxHash, anchor.Contract, proof)
	return err
}
