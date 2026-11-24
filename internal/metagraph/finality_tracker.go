package metagraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultFinalityInterval = 30 * time.Second
	finalityBatchLimit      = 200
	maxSnapshotBytes        = 128 << 20
	// rootInSnapshotGrace is how long after finality the tracker keeps waiting
	// for a metagraph snapshot whose on-chain state contains the root before
	// settling for one that only bounds finality.
	rootInSnapshotGrace = 2 * time.Minute
)

// SnapshotArchive stores raw snapshot bodies in object storage.
// media.StorageBackend satisfies it.
type SnapshotArchive interface {
	Store(ctx context.Context, key string, data []byte) (string, error)
}

// FinalityTrackerConfig wires the finality tracker.
type FinalityTrackerConfig struct {
	Roots       RootStore
	Network     string          // ECHO_CHAIN_NETWORK; part of the archive key
	ML0URL      string          // Metagraph L0 base URL (required for the tracker to run)
	GlobalL0URL string          // optional; no Global L0 snapshot is captured without it
	Archive     SnapshotArchive // optional; archive key/sha256 stay empty without it
	Interval    time.Duration
	HTTP        *http.Client
}

// FinalityTracker follows anchored roots until Metagraph L0 reports them
// finalized, then captures the signed snapshots current at that moment so the
// record stays checkable after the chain (or its in-memory root index) is gone.
//
// What the captured evidence means:
//   - "finalized" is the answer of ML0's /data-application/merkle-roots/{root}
//     at poll time. That index is in memory on the chain side, which is why the
//     answer is persisted here.
//   - The ML0 snapshot is the latest one at capture time, not necessarily the
//     one whose data block introduced the root. Echo's on-chain state is
//     cumulative, so the tracker checks whether the root is present in that
//     snapshot's decoded on-chain state and records the result
//     (SnapshotRef.ContainsRoot). When it is absent or cannot be decoded, the
//     snapshot only means "observed finalized at or before this ordinal".
//   - The Global L0 snapshot is the latest at capture time. It is never
//     claimed to include the metagraph snapshot: it only bounds the time of
//     observation ("at or before this global ordinal").
//   - Snapshot hashes come from the node's /latest/metadata endpoint and are
//     not recomputed; snapshot signatures are not verified (see
//     SignedSnapshotInfo).
type FinalityTracker struct {
	cfg  FinalityTrackerConfig
	http *http.Client
	now  func() time.Time
}

// NewFinalityTracker builds a tracker.
func NewFinalityTracker(cfg FinalityTrackerConfig) *FinalityTracker {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultFinalityInterval
	}
	cfg.ML0URL = strings.TrimRight(strings.TrimSpace(cfg.ML0URL), "/")
	cfg.GlobalL0URL = strings.TrimRight(strings.TrimSpace(cfg.GlobalL0URL), "/")
	client := cfg.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &FinalityTracker{cfg: cfg, http: client, now: time.Now}
}

// Run blocks until ctx is cancelled, checking on each tick. It no-ops without
// a root store or a Metagraph L0 URL (nothing to ask about finality).
func (t *FinalityTracker) Run(ctx context.Context) {
	if t == nil || t.cfg.Roots == nil || t.cfg.ML0URL == "" {
		log.Println("anchor finality tracker: no root store or METAGRAPH_L0_URL; skipping")
		return
	}
	ticker := time.NewTicker(t.cfg.Interval)
	defer ticker.Stop()
	log.Printf("anchor finality tracker: started (interval=%s, network=%s, archive=%t, globalL0=%t)",
		t.cfg.Interval, t.cfg.Network, t.cfg.Archive != nil, t.cfg.GlobalL0URL != "")
	for {
		select {
		case <-ctx.Done():
			log.Println("anchor finality tracker: stopped")
			return
		case <-ticker.C:
			t.CheckOnce(ctx)
		}
	}
}

// tickSnapshots caches the snapshot capture for one pass, so many roots
// finalized together share one fetch and one archived object.
type tickSnapshots struct {
	attempted bool
	ml0       *SnapshotRef
	ml0Info   SignedSnapshotInfo
	gl0       *SnapshotRef
}

// CheckOnce runs a single pass over roots that still need work.
func (t *FinalityTracker) CheckOnce(ctx context.Context) {
	// Background goroutine: contain a panic to this pass instead of taking the
	// API server down.
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("anchor finality tracker: recovered from panic: %v", rec)
		}
	}()
	if t.cfg.Roots == nil || t.cfg.ML0URL == "" {
		return
	}
	roots, err := t.cfg.Roots.ListIncompleteRoots(ctx, finalityBatchLimit)
	if err != nil {
		log.Printf("anchor finality tracker: list roots failed: %v", err)
		return
	}
	var snaps tickSnapshots
	for _, r := range roots {
		if ctx.Err() != nil {
			return
		}
		if !r.Finalized {
			finalized, err := t.pollFinalized(ctx, r.MerkleRoot)
			if err != nil {
				// ML0 unreachable or erroring: the rest of the batch would
				// fail the same way, so try again next tick.
				log.Printf("anchor finality tracker: finality poll failed: %v", err)
				return
			}
			if !finalized {
				continue
			}
			at := t.now().UTC()
			// Finality is recorded first and on its own: it must not depend on
			// snapshot capture or object storage succeeding.
			if err := t.cfg.Roots.MarkRootFinalized(ctx, r.MerkleRoot, at); err != nil {
				log.Printf("anchor finality tracker: mark finalized %s failed: %v", r.MerkleRoot, err)
				continue
			}
			r.Finalized, r.FinalizedAt = true, &at
			log.Printf("anchor finality tracker: root %s finalized", r.MerkleRoot)
		}
		t.recordSnapshots(ctx, r, &snaps)
	}
}

// recordSnapshots attaches snapshot evidence to a finalized root. Failures
// leave the root in the incomplete list, so it is retried on a later tick.
func (t *FinalityTracker) recordSnapshots(ctx context.Context, r AnchoredRoot, snaps *tickSnapshots) {
	if !snaps.attempted {
		snaps.attempted = true
		ml0, info, err := t.capture(ctx, t.cfg.ML0URL, "/snapshots", "ml0")
		if err != nil {
			log.Printf("anchor finality tracker: metagraph snapshot capture failed: %v", err)
		} else {
			snaps.ml0, snaps.ml0Info = ml0, info
			if t.cfg.GlobalL0URL != "" {
				gl0, _, err := t.capture(ctx, t.cfg.GlobalL0URL, "/global-snapshots", "gl0")
				if err != nil {
					// Recorded without the Global L0 side; not retried.
					log.Printf("anchor finality tracker: global snapshot capture failed: %v", err)
				} else {
					snaps.gl0 = gl0
				}
			}
		}
	}
	if snaps.ml0 == nil {
		return
	}
	contains := snaps.ml0Info.ContainsRoot(r.MerkleRoot)
	if !contains && r.FinalizedAt != nil && t.now().Sub(*r.FinalizedAt) < rootInSnapshotGrace {
		// ML0 can report finality before the snapshot carrying the root is
		// served. Wait for a snapshot that actually contains it rather than
		// archiving one that only bounds finality.
		return
	}
	ml0 := *snaps.ml0
	// contains=false is stored as such: this snapshot does not show the root
	// (absent, or the on-chain state could not be decoded) and only means
	// "observed finalized at or before this ordinal".
	ml0.ContainsRoot = &contains
	if err := t.cfg.Roots.SaveRootSnapshots(ctx, r.MerkleRoot, &ml0, snaps.gl0); err != nil {
		log.Printf("anchor finality tracker: save snapshots for %s failed: %v", r.MerkleRoot, err)
	}
}

// pollFinalized asks ML0 whether a root is finalized. 404 means ML0 does not
// know the root (not folded yet, or its in-memory index was reset).
func (t *FinalityTracker) pollFinalized(ctx context.Context, root string) (bool, error) {
	body, status, err := t.get(ctx, t.cfg.ML0URL+"/data-application/merkle-roots/"+url.PathEscape(root), 1<<20)
	if err != nil {
		return false, err
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	if status != http.StatusOK {
		return false, fmt.Errorf("merkle-roots status %d", status)
	}
	var st struct {
		Root      string `json:"root"`
		Finalized bool   `json:"finalized"`
	}
	if err := json.Unmarshal(body, &st); err != nil {
		return false, fmt.Errorf("decode merkle-root status: %w", err)
	}
	return st.Finalized && st.Root == root, nil
}

// capture fetches the latest signed snapshot of one layer (by metadata, then by
// ordinal, so ordinal/hash/body agree) and archives its raw body if configured.
func (t *FinalityTracker) capture(ctx context.Context, base, prefix, layer string) (*SnapshotRef, SignedSnapshotInfo, error) {
	metaBody, status, err := t.get(ctx, base+prefix+"/latest/metadata", 1<<20)
	if err != nil {
		return nil, SignedSnapshotInfo{}, err
	}
	if status != http.StatusOK {
		return nil, SignedSnapshotInfo{}, fmt.Errorf("%s metadata status %d", layer, status)
	}
	var meta struct {
		Ordinal *int64 `json:"ordinal"`
		Hash    string `json:"hash"`
	}
	if err := json.Unmarshal(metaBody, &meta); err != nil || meta.Ordinal == nil || !isHex(meta.Hash) {
		return nil, SignedSnapshotInfo{}, fmt.Errorf("%s metadata: unexpected body", layer)
	}
	ordinal := strconv.FormatInt(*meta.Ordinal, 10)
	body, status, err := t.get(ctx, base+prefix+"/"+ordinal, maxSnapshotBytes)
	if err != nil {
		return nil, SignedSnapshotInfo{}, err
	}
	if status != http.StatusOK {
		return nil, SignedSnapshotInfo{}, fmt.Errorf("%s snapshot %s status %d", layer, ordinal, status)
	}
	info, err := ParseSignedSnapshot(body)
	if err != nil {
		return nil, SignedSnapshotInfo{}, fmt.Errorf("%s snapshot %s: %w", layer, ordinal, err)
	}
	if info.Ordinal != *meta.Ordinal {
		return nil, SignedSnapshotInfo{}, fmt.Errorf("%s snapshot ordinal %d does not match metadata %d", layer, info.Ordinal, *meta.Ordinal)
	}
	ref := &SnapshotRef{Ordinal: *meta.Ordinal, Hash: meta.Hash, Signers: info.Signers}
	if t.cfg.Archive != nil {
		key := SnapshotArchiveKey(t.cfg.Network, layer, *meta.Ordinal, meta.Hash)
		if _, err := t.cfg.Archive.Store(ctx, key, body); err != nil {
			return nil, SignedSnapshotInfo{}, fmt.Errorf("archive %s: %w", key, err)
		}
		ref.ArchiveKey = key
		ref.SHA256 = SHA256Hex(body)
	}
	return ref, info, nil
}

// SnapshotArchiveKey is the deterministic object key for an archived snapshot.
// layer is "ml0" or "gl0".
func SnapshotArchiveKey(network, layer string, ordinal int64, hash string) string {
	return fmt.Sprintf("chain-archive/%s/%s/%d-%s.json", network, layer, ordinal, hash)
}

func (t *FinalityTracker) get(ctx context.Context, u string, limit int64) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	// Tessellation 4.0.7 answers 415 to a snapshot request that carries no Accept header
	// (it negotiates JSON vs binary); Go's client sends none by default.
	req.Header.Set("Accept", "application/json")
	resp, err := t.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, 0, err
	}
	if int64(len(body)) > limit {
		return nil, 0, fmt.Errorf("response from %s exceeds %d bytes", u, limit)
	}
	return body, resp.StatusCode, nil
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
