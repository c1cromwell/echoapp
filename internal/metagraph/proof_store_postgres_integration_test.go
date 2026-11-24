//go:build integration

package metagraph

import (
	"context"
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
// The test applies migrations/036_anchor_proofs.sql itself (twice, to check it
// is idempotent) and removes the rows it writes.
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

	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "036_anchor_proofs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("apply migration (pass %d): %v", i+1, err)
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
		if err := store.RecordRootSubmitted(ctx, r, 4, ChainNetworkSovereign, t0.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	// Re-recording must not reset an existing row.
	if err := store.RecordRootSubmitted(ctx, rootA, 99, ChainNetworkConstellation, t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	rec, ok, err := store.GetRoot(ctx, rootA)
	if err != nil || !ok {
		t.Fatalf("get root: ok=%v err=%v", ok, err)
	}
	if rec.LeafCount != 4 || rec.Network != ChainNetworkSovereign || !rec.SubmittedAt.Equal(t0) || rec.Finalized || rec.FinalizedAt != nil || rec.ML0 != nil || rec.GL0 != nil {
		t.Fatalf("fresh root: %+v", rec)
	}
	if _, ok, _ := store.GetRoot(ctx, "it-anchor-root-missing"); ok {
		t.Fatal("unexpected root")
	}

	// Finality without snapshots (no object storage / URLs): archive fields stay NULL.
	if err := store.MarkRootFinalized(ctx, rootA, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkRootFinalized(ctx, rootA, t0.Add(time.Hour)); err != nil { // first finalized_at wins
		t.Fatal(err)
	}
	rec, _, _ = store.GetRoot(ctx, rootA)
	if !rec.Finalized || rec.FinalizedAt == nil || !rec.FinalizedAt.Equal(t0.Add(time.Minute)) || rec.ML0 != nil {
		t.Fatalf("finalized root: %+v", rec)
	}

	// Snapshot evidence: ML0 without archive on B; full ML0+GL0 on A.
	no := false
	if err := store.SaveRootSnapshots(ctx, rootB, &SnapshotRef{Ordinal: 7, Hash: "h7", Signers: []string{"p1"}, ContainsRoot: &no}, nil); err != nil {
		t.Fatal(err)
	}
	rec, _, _ = store.GetRoot(ctx, rootB)
	if rec.ML0 == nil || rec.ML0.ArchiveKey != "" || rec.ML0.SHA256 != "" || rec.ML0.ContainsRoot == nil || *rec.ML0.ContainsRoot || rec.GL0 != nil {
		t.Fatalf("root B: ml0=%+v gl0=%+v", rec.ML0, rec.GL0)
	}
	var nullKey, nullSHA bool
	if err := pool.QueryRow(ctx, `SELECT ml0_archive_key IS NULL, gl0_snapshot_ordinal IS NULL FROM anchor_roots WHERE merkle_root = $1`, rootB).Scan(&nullKey, &nullSHA); err != nil || !nullKey || !nullSHA {
		t.Fatalf("archive columns should be NULL: %v %v err=%v", nullKey, nullSHA, err)
	}

	yes := true
	ml0 := &SnapshotRef{Ordinal: 42, Hash: "mh", Signers: []string{"p1", "p2"}, ArchiveKey: "chain-archive/sovereign/ml0/42-mh.json", SHA256: "s1", ContainsRoot: &yes}
	gl0 := &SnapshotRef{Ordinal: 900, Hash: "gh", Signers: []string{"g1"}, ArchiveKey: "chain-archive/sovereign/gl0/900-gh.json", SHA256: "s2"}
	if err := store.SaveRootSnapshots(ctx, rootA, ml0, gl0); err != nil {
		t.Fatal(err)
	}
	rec, _, _ = store.GetRoot(ctx, rootA)
	if !reflect.DeepEqual(rec.ML0, ml0) || !reflect.DeepEqual(rec.GL0, gl0) {
		t.Fatalf("root A snapshots: ml0=%+v gl0=%+v", rec.ML0, rec.GL0)
	}

	// Incomplete = not finalized OR no ML0 snapshot; newest first.
	// A: finalized + ML0 -> complete. B: ML0 but not finalized. C: untouched.
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
	if err != nil || !b.Finalized || b.MetagraphSnapshot == nil || b.Message == nil {
		t.Fatalf("bundle: %+v err=%v", b, err)
	}
}
