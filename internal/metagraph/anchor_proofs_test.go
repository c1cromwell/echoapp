package metagraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testLeaves(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = SHA256Hex([]byte(fmt.Sprintf("leaf-%d", i)))
	}
	return out
}

func TestVerifyMerklePath_AllLeavesAllSizes(t *testing.T) {
	for n := 1; n <= 17; n++ {
		leaves := testLeaves(n)
		tree := BuildMerkleTree(leaves)
		for i := range leaves {
			if err := VerifyMerklePath(leaves[i], tree.ProofSiblings(i), i, n, tree.Root); err != nil {
				t.Fatalf("n=%d leaf=%d: %v", n, i, err)
			}
		}
	}
}

func TestVerifyMerklePath_Failures(t *testing.T) {
	leaves := testLeaves(5)
	tree := BuildMerkleTree(leaves)
	sib := tree.ProofSiblings(2)
	other := SHA256Hex([]byte("other"))
	tampered := append([]string(nil), sib...)
	tampered[0] = other

	cases := map[string]error{
		"wrong commitment": VerifyMerklePath(other, sib, 2, 5, tree.Root),
		"wrong root":       VerifyMerklePath(leaves[2], sib, 2, 5, other),
		"tampered sibling": VerifyMerklePath(leaves[2], tampered, 2, 5, tree.Root),
		"wrong index":      VerifyMerklePath(leaves[2], sib, 3, 5, tree.Root),
		"wrong leaf count": VerifyMerklePath(leaves[2], sib, 2, 9, tree.Root),
		"short path":       VerifyMerklePath(leaves[2], sib[:len(sib)-1], 2, 5, tree.Root),
		"long path":        VerifyMerklePath(leaves[2], append(append([]string(nil), sib...), other), 2, 5, tree.Root),
		"index range":      VerifyMerklePath(leaves[2], sib, 5, 5, tree.Root),
		"zero leaf count":  VerifyMerklePath(leaves[2], sib, 2, 0, tree.Root),
	}
	for name, err := range cases {
		if err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseChainNetwork(t *testing.T) {
	for in, want := range map[string]string{
		"": ChainNetworkSovereign, "sovereign": ChainNetworkSovereign,
		" Constellation ": ChainNetworkConstellation, "constellation": ChainNetworkConstellation,
	} {
		got, err := ParseChainNetwork(in)
		if err != nil || got != want {
			t.Errorf("ParseChainNetwork(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"mainnet", "testnet", "sovereign2", "local"} {
		if _, err := ParseChainNetwork(in); err == nil || !strings.Contains(err.Error(), "ECHO_CHAIN_NETWORK") {
			t.Errorf("ParseChainNetwork(%q): want a clear error, got %v", in, err)
		}
	}
	t.Setenv("ECHO_CHAIN_NETWORK", "bogus")
	if _, err := ChainNetworkFromEnv(); err == nil {
		t.Error("ChainNetworkFromEnv should reject bogus value")
	}
}

// snapshotJSON builds a signed-snapshot body shaped like Tessellation's:
// onChainState is a JSON array of byte values holding the state JSON.
func snapshotJSON(ordinal int64, signers []string, roots ...string) []byte {
	value := map[string]any{"ordinal": ordinal, "lastSnapshotHash": "00"}
	if roots != nil {
		mr := map[string]any{}
		for _, r := range roots {
			mr[r] = map[string]any{"root": r, "leafCount": 1}
		}
		state, _ := json.Marshal(map[string]any{"merkleRoots": mr, "trustCommitments": map[string]any{}})
		ints := make([]int, len(state))
		for i, b := range state {
			ints[i] = int(int8(b)) // signed, as circe emits
		}
		value["dataApplication"] = map[string]any{"onChainState": ints, "blocks": []any{}}
	}
	proofs := []map[string]string{}
	for _, s := range signers {
		proofs = append(proofs, map[string]string{"id": s, "signature": "3045"})
	}
	b, _ := json.Marshal(map[string]any{"value": value, "proofs": proofs})
	return b
}

func TestParseSignedSnapshot(t *testing.T) {
	root := SHA256Hex([]byte("r"))
	info, err := ParseSignedSnapshot(snapshotJSON(7, []string{"bb", "aa", "bb"}, root))
	if err != nil {
		t.Fatal(err)
	}
	if info.Ordinal != 7 || !info.HasDataState || !info.ContainsRoot(root) || info.ContainsRoot("nope") {
		t.Fatalf("unexpected info: %+v", info)
	}
	if strings.Join(info.Signers, ",") != "aa,bb" {
		t.Fatalf("signers = %v", info.Signers)
	}
	// Global-style snapshot: no data application.
	info, err = ParseSignedSnapshot(snapshotJSON(9, []string{"cc"}))
	if err != nil || info.HasDataState || info.ContainsRoot(root) {
		t.Fatalf("global snapshot: %+v err=%v", info, err)
	}
	// Opaque (non-JSON) on-chain state is not an error, just undecoded.
	info, err = ParseSignedSnapshot([]byte(`{"value":{"ordinal":1,"dataApplication":{"onChainState":[1,2,-3]}},"proofs":[]}`))
	if err != nil || info.HasDataState {
		t.Fatalf("opaque state: %+v err=%v", info, err)
	}
	for _, bad := range []string{`not json`, `{"proofs":[]}`} {
		if _, err := ParseSignedSnapshot([]byte(bad)); err == nil {
			t.Errorf("ParseSignedSnapshot(%q): expected error", bad)
		}
	}
}

// fakeChain serves the ML0 / Global L0 endpoints the tracker uses.
type fakeChain struct {
	mu        sync.Mutex
	finalized map[string]bool // root -> finalized (absent = 404)
	ordinal   int64
	hash      string
	signers   []string
	roots     []string // roots present in on-chain state; nil = no data application
	prefix    string
	hits      int
}

func (c *fakeChain) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.hits++
		// Mirrors Tessellation 4.0.7, which rejects snapshot reads that carry no Accept header.
		if strings.HasPrefix(r.URL.Path, c.prefix+"/") && r.Header.Get("Accept") == "" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/data-application/merkle-roots/"):
			root := strings.TrimPrefix(r.URL.Path, "/data-application/merkle-roots/")
			fin, ok := c.finalized[root]
			if !ok {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"root": root, "leafCount": 2, "finalized": fin})
		case r.URL.Path == c.prefix+"/latest/metadata":
			_ = json.NewEncoder(w).Encode(map[string]any{"ordinal": c.ordinal, "hash": c.hash, "lastSnapshotHash": "00"})
		case r.URL.Path == fmt.Sprintf("%s/%d", c.prefix, c.ordinal):
			_, _ = w.Write(snapshotJSON(c.ordinal, c.signers, c.roots...))
		default:
			http.NotFound(w, r)
		}
	})
}

type memArchive struct {
	mu      sync.Mutex
	objects map[string][]byte
	err     error
}

func (a *memArchive) Store(_ context.Context, key string, data []byte) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.err != nil {
		return "", a.err
	}
	if a.objects == nil {
		a.objects = map[string][]byte{}
	}
	a.objects[key] = append([]byte(nil), data...)
	return "", nil
}

func seedRoot(t *testing.T, store *MemoryProofStore, root string) {
	t.Helper()
	if err := store.RecordRootSubmitted(context.Background(), root, 2, ChainNetworkSovereign, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestFinalityTracker_FinalizedArchivesSnapshots(t *testing.T) {
	ctx := context.Background()
	root := SHA256Hex([]byte("root-1"))
	ml0 := &fakeChain{finalized: map[string]bool{root: true}, ordinal: 42, hash: "ab12", signers: []string{"peer-b", "peer-a"}, roots: []string{root}, prefix: "/snapshots"}
	gl0 := &fakeChain{ordinal: 900, hash: "cd34", signers: []string{"gpeer"}, prefix: "/global-snapshots"}
	ml0Srv, gl0Srv := httptest.NewServer(ml0.handler()), httptest.NewServer(gl0.handler())
	defer ml0Srv.Close()
	defer gl0Srv.Close()

	store, archive := NewMemoryProofStore(), &memArchive{}
	seedRoot(t, store, root)
	NewFinalityTracker(FinalityTrackerConfig{
		Roots: store, Network: ChainNetworkSovereign, ML0URL: ml0Srv.URL + "/", GlobalL0URL: gl0Srv.URL, Archive: archive,
	}).CheckOnce(ctx)

	rec, _, _ := store.GetRoot(ctx, root)
	if !rec.Finalized || rec.FinalizedAt == nil || rec.ML0 == nil || rec.GL0 == nil {
		t.Fatalf("record incomplete: %+v", rec)
	}
	wantML0Key := "chain-archive/sovereign/ml0/42-ab12.json"
	wantGL0Key := "chain-archive/sovereign/gl0/900-cd34.json"
	if rec.ML0.Ordinal != 42 || rec.ML0.Hash != "ab12" || rec.ML0.ArchiveKey != wantML0Key {
		t.Fatalf("ml0 ref: %+v", rec.ML0)
	}
	if rec.ML0.ContainsRoot == nil || !*rec.ML0.ContainsRoot {
		t.Fatalf("ml0 should record the root as contained: %+v", rec.ML0)
	}
	if strings.Join(rec.ML0.Signers, ",") != "peer-a,peer-b" || strings.Join(rec.GL0.Signers, ",") != "gpeer" {
		t.Fatalf("signers: %v / %v", rec.ML0.Signers, rec.GL0.Signers)
	}
	if rec.GL0.Ordinal != 900 || rec.GL0.ArchiveKey != wantGL0Key || rec.GL0.ContainsRoot != nil {
		t.Fatalf("gl0 ref: %+v", rec.GL0)
	}
	if got := SHA256Hex(archive.objects[wantML0Key]); got != rec.ML0.SHA256 || rec.ML0.SHA256 == "" {
		t.Fatalf("ml0 sha256 %q vs archived %q", rec.ML0.SHA256, got)
	}
	if got := SHA256Hex(archive.objects[wantGL0Key]); got != rec.GL0.SHA256 {
		t.Fatalf("gl0 sha256 %q vs archived %q", rec.GL0.SHA256, got)
	}
	// Completed roots are not polled again.
	if left, _ := store.ListIncompleteRoots(ctx, 10); len(left) != 0 {
		t.Fatalf("root still listed incomplete: %+v", left)
	}
}

func TestFinalityTracker_NotFinalized(t *testing.T) {
	ctx := context.Background()
	pending, unknown := SHA256Hex([]byte("pending")), SHA256Hex([]byte("unknown"))
	ml0 := &fakeChain{finalized: map[string]bool{pending: false}, ordinal: 1, hash: "aa", prefix: "/snapshots"}
	srv := httptest.NewServer(ml0.handler())
	defer srv.Close()
	store, archive := NewMemoryProofStore(), &memArchive{}
	seedRoot(t, store, pending)
	seedRoot(t, store, unknown) // 404 from ML0
	NewFinalityTracker(FinalityTrackerConfig{Roots: store, ML0URL: srv.URL, Archive: archive}).CheckOnce(ctx)
	for _, r := range []string{pending, unknown} {
		if rec, _, _ := store.GetRoot(ctx, r); rec.Finalized || rec.ML0 != nil {
			t.Fatalf("root %s must stay unfinalized: %+v", r, rec)
		}
	}
	if len(archive.objects) != 0 {
		t.Fatalf("nothing should be archived, got %d objects", len(archive.objects))
	}
}

func TestFinalityTracker_ML0Down(t *testing.T) {
	ctx := context.Background()
	root := SHA256Hex([]byte("root-down"))
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close() // connection refused
	store := NewMemoryProofStore()
	seedRoot(t, store, root)
	NewFinalityTracker(FinalityTrackerConfig{Roots: store, ML0URL: url}).CheckOnce(ctx)
	if rec, _, _ := store.GetRoot(ctx, root); rec.Finalized {
		t.Fatal("root must not be finalized when ML0 is unreachable")
	}

	// ML0 answering 500 is also not finality.
	srv500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv500.Close()
	NewFinalityTracker(FinalityTrackerConfig{Roots: store, ML0URL: srv500.URL}).CheckOnce(ctx)
	if rec, _, _ := store.GetRoot(ctx, root); rec.Finalized {
		t.Fatal("root must not be finalized on ML0 errors")
	}
}

func TestFinalityTracker_NoObjectStorageOrGlobalL0(t *testing.T) {
	ctx := context.Background()
	root := SHA256Hex([]byte("root-noarchive"))
	ml0 := &fakeChain{finalized: map[string]bool{root: true}, ordinal: 5, hash: "ee", signers: []string{"p"}, roots: []string{root}, prefix: "/snapshots"}
	srv := httptest.NewServer(ml0.handler())
	defer srv.Close()
	store := NewMemoryProofStore()
	seedRoot(t, store, root)
	NewFinalityTracker(FinalityTrackerConfig{Roots: store, ML0URL: srv.URL}).CheckOnce(ctx)
	rec, _, _ := store.GetRoot(ctx, root)
	if !rec.Finalized || rec.ML0 == nil {
		t.Fatalf("finality must be recorded without object storage: %+v", rec)
	}
	if rec.ML0.ArchiveKey != "" || rec.ML0.SHA256 != "" || rec.GL0 != nil {
		t.Fatalf("archive fields must stay empty: ml0=%+v gl0=%+v", rec.ML0, rec.GL0)
	}
}

func TestFinalityTracker_SnapshotFailuresKeepFinalityAndRetry(t *testing.T) {
	ctx := context.Background()
	root := SHA256Hex([]byte("root-retry"))
	ml0 := &fakeChain{finalized: map[string]bool{root: true}, ordinal: 5, hash: "ee", signers: []string{"p"}, roots: []string{root}, prefix: "/snapshots"}
	srv := httptest.NewServer(ml0.handler())
	defer srv.Close()
	gl0Down := httptest.NewServer(http.NotFoundHandler())
	gl0URL := gl0Down.URL
	gl0Down.Close()

	store, archive := NewMemoryProofStore(), &memArchive{err: errors.New("bucket unavailable")}
	seedRoot(t, store, root)
	tr := NewFinalityTracker(FinalityTrackerConfig{Roots: store, ML0URL: srv.URL, GlobalL0URL: gl0URL, Archive: archive})
	tr.CheckOnce(ctx)
	rec, _, _ := store.GetRoot(ctx, root)
	if !rec.Finalized || rec.ML0 != nil {
		t.Fatalf("archive failure: want finalized without snapshot, got %+v", rec)
	}

	// Storage recovers; Global L0 is still down. The ML0 side is retried and
	// recorded, the Global L0 side stays empty.
	archive.err = nil
	tr.CheckOnce(ctx)
	rec, _, _ = store.GetRoot(ctx, root)
	if rec.ML0 == nil || rec.ML0.ArchiveKey == "" || rec.GL0 != nil {
		t.Fatalf("after retry: ml0=%+v gl0=%+v", rec.ML0, rec.GL0)
	}
}

func TestFinalityTracker_RootNotInSnapshotIsNotClaimed(t *testing.T) {
	ctx := context.Background()
	root := SHA256Hex([]byte("root-absent"))
	ml0 := &fakeChain{finalized: map[string]bool{root: true}, ordinal: 5, hash: "ee", signers: []string{"p"}, roots: []string{}, prefix: "/snapshots"}
	srv := httptest.NewServer(ml0.handler())
	defer srv.Close()
	store := NewMemoryProofStore()
	seedRoot(t, store, root)
	tr := NewFinalityTracker(FinalityTrackerConfig{Roots: store, ML0URL: srv.URL})
	now := time.Now()
	tr.now = func() time.Time { return now }

	// Within the grace period: finalized, but no snapshot is attached yet.
	tr.CheckOnce(ctx)
	rec, _, _ := store.GetRoot(ctx, root)
	if !rec.Finalized || rec.ML0 != nil {
		t.Fatalf("within grace: %+v", rec)
	}

	// After the grace period: the snapshot is recorded as a bound only.
	now = now.Add(rootInSnapshotGrace + time.Second)
	tr.CheckOnce(ctx)
	rec, _, _ = store.GetRoot(ctx, root)
	if rec.ML0 == nil || rec.ML0.ContainsRoot == nil || *rec.ML0.ContainsRoot {
		t.Fatalf("after grace: snapshot must be recorded with containsRoot=false, got %+v", rec.ML0)
	}

	// If the root shows up before the grace expires, it is recorded as contained.
	root2 := SHA256Hex([]byte("root-late"))
	seedRoot(t, store, root2)
	ml0.mu.Lock()
	ml0.finalized[root2] = true
	ml0.mu.Unlock()
	tr.CheckOnce(ctx)
	ml0.mu.Lock()
	ml0.roots = []string{root2}
	ml0.mu.Unlock()
	tr.CheckOnce(ctx)
	rec, _, _ = store.GetRoot(ctx, root2)
	if rec.ML0 == nil || rec.ML0.ContainsRoot == nil || !*rec.ML0.ContainsRoot {
		t.Fatalf("late root: %+v", rec.ML0)
	}
}

func TestFinalityTracker_RunStopsOnCancelAndSkipsWithoutURL(t *testing.T) {
	NewFinalityTracker(FinalityTrackerConfig{Roots: NewMemoryProofStore()}).Run(context.Background()) // returns: no URL

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		NewFinalityTracker(FinalityTrackerConfig{Roots: NewMemoryProofStore(), ML0URL: "http://127.0.0.1:1", Interval: time.Millisecond}).Run(ctx)
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop after cancel")
	}
}

type panicRoots struct{ RootStore }

func (panicRoots) ListIncompleteRoots(context.Context, int) ([]AnchoredRoot, error) {
	panic("store exploded")
}

func TestFinalityTracker_RecoversFromPanic(t *testing.T) {
	NewFinalityTracker(FinalityTrackerConfig{Roots: panicRoots{}, ML0URL: "http://127.0.0.1:1"}).CheckOnce(context.Background())
}

func TestBuildProofBundle(t *testing.T) {
	ctx := context.Background()
	store := NewRedisProofStore(nil, nil) // fallback store must satisfy AnchorStore
	leaves := testLeaves(3)
	tree := BuildMerkleTree(leaves)
	if err := store.RecordRootSubmitted(ctx, tree.Root, 3, ChainNetworkSovereign, time.Now()); err != nil {
		t.Fatal(err)
	}
	_ = store.Put(ctx, MessageAnchorProof{MessageID: "m2", Commitment: leaves[2], Siblings: tree.ProofSiblings(2), MerkleLeafIndex: 2, MerkleRoot: tree.Root})
	_ = store.Put(ctx, MessageAnchorProof{MessageID: "other", Commitment: "x", MerkleRoot: "some-other-root"})

	b, err := BuildProofBundle(ctx, store, "fallback", tree.Root, "m2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Network != ChainNetworkSovereign || b.Finalized || b.LeafCount != 3 || b.Message == nil {
		t.Fatalf("bundle: %+v", b)
	}
	if err := VerifyMerklePath(b.Message.Commitment, b.Message.Siblings, b.Message.LeafIndex, b.LeafCount, b.MerkleRoot); err != nil {
		t.Fatalf("bundle path does not verify: %v", err)
	}
	if _, err := BuildProofBundle(ctx, store, "", "missing", ""); !errors.Is(err, ErrRootNotFound) {
		t.Fatalf("missing root: %v", err)
	}
	for _, id := range []string{"other", "nope"} {
		if _, err := BuildProofBundle(ctx, store, "", tree.Root, id); !errors.Is(err, ErrMessageNotInRoot) {
			t.Fatalf("message %q: %v", id, err)
		}
	}
}
