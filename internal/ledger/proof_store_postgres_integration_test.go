//go:build integration

package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run against a SCRATCH database (never the app's own):
//
//	DATABASE_URL=postgres://... go test -tags integration ./internal/metagraph/ -run PostgresProofStore
//
// The test applies migrations 036 and 037 itself (twice, to check they are
// idempotent) and removes the rows it writes.
func TestPostgresProofStoreIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	for i := 0; i < 2; i++ {
		for _, name := range []string{"036_anchor_proofs.sql", "037_anchor_chain_proof.sql"} {
			migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := pool.Exec(ctx, string(migration)); err != nil {
				t.Fatalf("apply %s (pass %d): %v", name, i+1, err)
			}
		}
	}

	const rootA, rootB, rootC = "it-anchor-root-a", "it-anchor-root-b", "it-anchor-root-c"
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DELETE FROM anchor_roots WHERE merkle_root LIKE 'it-anchor-root-%'`)
		_, _ = pool.Exec(ctx, `DELETE FROM anchor_message_proofs WHERE message_id LIKE 'it-anchor-msg-%'`)
	}
	cleanup()
	defer cleanup()

	var store AnchorStore = NewPostgresProofStore(pool)

	// --- per-message proofs ---
	if _, ok, err := store.Get(ctx, "it-anchor-msg-missing"); err != nil || ok {
		t.Fatalf("missing proof: ok=%v err=%v", ok, err)
	}
	proof := MessageAnchorProof{
		MessageID: "it-anchor-msg-1", Commitment: "c0ffee", Siblings: []string{"aa", "bb"},
		MerkleLeafIndex: 3, SnapshotHash: "tx-1", MerkleRoot: rootA,
	}
	if err := store.Put(ctx, proof); err != nil {
		t.Fatal(err)
	}
	got, ok, err := store.Get(ctx, proof.MessageID)
	if err != nil || !ok || !reflect.DeepEqual(got, proof) {
		t.Fatalf("round trip: got %+v ok=%v err=%v", got, ok, err)
	}
	proof.Siblings, proof.MerkleLeafIndex = nil, 0 // single-leaf batch; Put is an upsert
	if err := store.Put(ctx, proof); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := store.Get(ctx, proof.MessageID); len(got.Siblings) != 0 || got.MerkleLeafIndex != 0 {
		t.Fatalf("upsert: %+v", got)
	}

	// --- anchored roots ---
	t0 := time.Now().UTC().Truncate(time.Microsecond)
	for i, r := range []string{rootA, rootB, rootC} {
		if err := store.RecordRootSubmitted(ctx, r, 4, ChainNetworkBesu, t0.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	// Re-recording must not reset an existing row.
	if err := store.RecordRootSubmitted(ctx, rootA, 99, "other", t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	rec, ok, err := store.GetRoot(ctx, rootA)
	if err != nil || !ok {
		t.Fatalf("get root: ok=%v err=%v", ok, err)
	}
	if rec.LeafCount != 4 || rec.Network != ChainNetworkBesu || !rec.SubmittedAt.Equal(t0) || rec.Finalized || rec.FinalizedAt != nil || rec.Chain != nil {
		t.Fatalf("fresh root: %+v", rec)
	}
	if _, ok, _ := store.GetRoot(ctx, "it-anchor-root-missing"); ok {
		t.Fatal("unexpected root")
	}

	// Finality before the chain proof is recorded: the proof columns stay NULL.
	if err := store.MarkRootFinalized(ctx, rootA, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkRootFinalized(ctx, rootA, t0.Add(time.Hour)); err != nil { // first finalized_at wins
		t.Fatal(err)
	}
	rec, _, _ = store.GetRoot(ctx, rootA)
	if !rec.Finalized || rec.FinalizedAt == nil || !rec.FinalizedAt.Equal(t0.Add(time.Minute)) || rec.Chain != nil {
		t.Fatalf("finalized root: %+v", rec)
	}

	var nullProof bool
	if err := pool.QueryRow(ctx, `SELECT chain_proof IS NULL FROM anchor_roots WHERE merkle_root = $1`, rootA).Scan(&nullProof); err != nil || !nullProof {
		t.Fatalf("chain_proof should be NULL before it is saved: %v err=%v", nullProof, err)
	}

	// Chain evidence: a proof on A (finalized -> complete); a block but no proof yet on B.
	full := ChainAnchor{BlockNumber: 42, TxHash: "0xtx", Contract: "0xc0", Proof: json.RawMessage(`{"root": "0x00"}`)}
	if err := store.SaveChainAnchor(ctx, rootA, full); err != nil {
		t.Fatal(err)
	}
	rec, _, _ = store.GetRoot(ctx, rootA)
	if rec.Chain == nil || rec.Chain.BlockNumber != 42 || rec.Chain.TxHash != "0xtx" || rec.Chain.Contract != "0xc0" || !json.Valid(rec.Chain.Proof) || len(rec.Chain.Proof) == 0 {
		t.Fatalf("root A chain evidence: %+v", rec.Chain)
	}
	if err := store.SaveChainAnchor(ctx, rootB, ChainAnchor{BlockNumber: 7}); err != nil {
		t.Fatal(err)
	}
	rec, _, _ = store.GetRoot(ctx, rootB)
	if rec.Chain == nil || rec.Chain.BlockNumber != 7 || len(rec.Chain.Proof) != 0 {
		t.Fatalf("root B chain evidence: %+v", rec.Chain)
	}

	// Incomplete = not finalized OR no chain proof; newest first.
	// A: finalized + proof -> complete. B: block but no proof, not finalized. C: untouched.
	list, err := store.ListIncompleteRoots(ctx, 1000)
	if err != nil {
		t.Fatal(err)
	}
	var mine []string
	for _, r := range list {
		if r.MerkleRoot == rootA || r.MerkleRoot == rootB || r.MerkleRoot == rootC {
			mine = append(mine, r.MerkleRoot)
		}
	}
	if !reflect.DeepEqual(mine, []string{rootC, rootB}) {
		t.Fatalf("incomplete roots = %v", mine)
	}
	if one, err := store.ListIncompleteRoots(ctx, 1); err != nil || len(one) != 1 {
		t.Fatalf("limit: %d err=%v", len(one), err)
	}

	// The bundle is served from the durable record alone.
	b, err := BuildProofBundle(ctx, store, "", rootA, proof.MessageID)
	if err != nil || !b.Finalized || b.Chain == nil || b.Message == nil {
		t.Fatalf("bundle: %+v err=%v", b, err)
	}
}
