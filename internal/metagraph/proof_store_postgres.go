package metagraph

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresProofStore is the durable AnchorStore (migration 036_anchor_proofs).
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
	ml0_snapshot_ordinal, ml0_snapshot_hash, ml0_contains_root, ml0_signers, ml0_archive_key, ml0_archive_sha256,
	gl0_snapshot_ordinal, gl0_snapshot_hash, gl0_signers, gl0_archive_key, gl0_archive_sha256`

func scanAnchoredRoot(row pgx.Row) (AnchoredRoot, error) {
	var (
		r                    AnchoredRoot
		mlOrd, glOrd         *int64
		mlHash, mlKey, mlSHA *string
		glHash, glKey, glSHA *string
		mlContains           *bool
		mlSigners, glSigners []byte
		deref                = func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		}
	)
	err := row.Scan(&r.MerkleRoot, &r.Network, &r.LeafCount, &r.SubmittedAt, &r.Finalized, &r.FinalizedAt,
		&mlOrd, &mlHash, &mlContains, &mlSigners, &mlKey, &mlSHA,
		&glOrd, &glHash, &glSigners, &glKey, &glSHA)
	if err != nil {
		return AnchoredRoot{}, err
	}
	if mlOrd != nil {
		ref := &SnapshotRef{Ordinal: *mlOrd, Hash: deref(mlHash), ArchiveKey: deref(mlKey), SHA256: deref(mlSHA), ContainsRoot: mlContains}
		if len(mlSigners) > 0 {
			if err := json.Unmarshal(mlSigners, &ref.Signers); err != nil {
				return AnchoredRoot{}, err
			}
		}
		r.ML0 = ref
	}
	if glOrd != nil {
		ref := &SnapshotRef{Ordinal: *glOrd, Hash: deref(glHash), ArchiveKey: deref(glKey), SHA256: deref(glSHA)}
		if len(glSigners) > 0 {
			if err := json.Unmarshal(glSigners, &ref.Signers); err != nil {
				return AnchoredRoot{}, err
			}
		}
		r.GL0 = ref
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
		WHERE NOT finalized OR ml0_snapshot_ordinal IS NULL
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

// SaveRootSnapshots implements RootStore.
func (s *PostgresProofStore) SaveRootSnapshots(ctx context.Context, root string, ml0, gl0 *SnapshotRef) error {
	nullable := func(v string) *string {
		if v == "" {
			return nil
		}
		return &v
	}
	signers := func(ref *SnapshotRef) (string, error) {
		ids := ref.Signers
		if ids == nil {
			ids = []string{}
		}
		b, err := json.Marshal(ids)
		return string(b), err
	}
	if ml0 != nil {
		sig, err := signers(ml0)
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, `
			UPDATE anchor_roots SET ml0_snapshot_ordinal = $2, ml0_snapshot_hash = $3, ml0_contains_root = $4,
				ml0_signers = $5::jsonb, ml0_archive_key = $6, ml0_archive_sha256 = $7
			WHERE merkle_root = $1`,
			root, ml0.Ordinal, ml0.Hash, ml0.ContainsRoot, sig, nullable(ml0.ArchiveKey), nullable(ml0.SHA256)); err != nil {
			return err
		}
	}
	if gl0 != nil {
		sig, err := signers(gl0)
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, `
			UPDATE anchor_roots SET gl0_snapshot_ordinal = $2, gl0_snapshot_hash = $3,
				gl0_signers = $4::jsonb, gl0_archive_key = $5, gl0_archive_sha256 = $6
			WHERE merkle_root = $1`,
			root, gl0.Ordinal, gl0.Hash, sig, nullable(gl0.ArchiveKey), nullable(gl0.SHA256)); err != nil {
			return err
		}
	}
	return nil
}
