package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// snapshotFile builds a signed-snapshot JSON body; roots == nil means no
// data-application part (Global L0 style).
func snapshotFile(ordinal int64, signers []string, roots []string) []byte {
	value := map[string]any{"ordinal": ordinal}
	if roots != nil {
		mr := map[string]any{}
		for _, r := range roots {
			mr[r] = map[string]any{"root": r}
		}
		state, _ := json.Marshal(map[string]any{"merkleRoots": mr})
		ints := make([]int, len(state))
		for i, b := range state {
			ints[i] = int(b)
		}
		value["dataApplication"] = map[string]any{"onChainState": ints}
	}
	proofs := []map[string]string{}
	for _, s := range signers {
		proofs = append(proofs, map[string]string{"id": s, "signature": "3045"})
	}
	b, _ := json.Marshal(map[string]any{"value": value, "proofs": proofs})
	return b
}

type fixture struct {
	dir                  string
	bundle               metagraph.ProofBundle
	ml0Body, gl0Body     []byte
	bundlePath, ml0, gl0 string
	trustedPath          string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	leaves := make([]string, 5)
	for i := range leaves {
		leaves[i] = metagraph.SHA256Hex([]byte{byte(i)})
	}
	tree := metagraph.BuildMerkleTree(leaves)
	f := &fixture{dir: t.TempDir()}
	f.ml0Body = snapshotFile(42, []string{"peer-a", "peer-b"}, []string{tree.Root})
	f.gl0Body = snapshotFile(900, []string{"gpeer"}, nil)
	yes, now := true, time.Now().UTC()
	f.bundle = metagraph.ProofBundle{
		Version: metagraph.ProofBundleVersion, Network: "sovereign", MerkleRoot: tree.Root, LeafCount: 5,
		SubmittedAt: now, Finalized: true, FinalizedAt: &now,
		MetagraphSnapshot: &metagraph.SnapshotRef{Ordinal: 42, Hash: "mh", Signers: []string{"peer-a", "peer-b"},
			ArchiveKey: "chain-archive/sovereign/ml0/42-mh.json", SHA256: metagraph.SHA256Hex(f.ml0Body), ContainsRoot: &yes},
		GlobalSnapshot: &metagraph.SnapshotRef{Ordinal: 900, Hash: "gh", Signers: []string{"gpeer"},
			ArchiveKey: "chain-archive/sovereign/gl0/900-gh.json", SHA256: metagraph.SHA256Hex(f.gl0Body)},
		Message: &metagraph.ProofBundleMessage{MessageID: "m4", Commitment: leaves[4], Siblings: tree.ProofSiblings(4), LeafIndex: 4},
	}
	f.ml0 = f.write(t, "ml0.json", f.ml0Body)
	f.gl0 = f.write(t, "gl0.json", f.gl0Body)
	f.trustedPath = f.write(t, "trusted.txt", []byte("# validators\nPEER-A\npeer-b  # second\n\ngpeer\n"))
	return f
}

func (f *fixture) write(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(f.dir, name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// verify writes the (possibly mutated) bundle and runs the verifier.
func (f *fixture) verify(t *testing.T, extra ...string) (int, string) {
	t.Helper()
	raw, err := json.Marshal(f.bundle)
	if err != nil {
		t.Fatal(err)
	}
	f.bundlePath = f.write(t, "bundle.json", raw)
	var out, errOut bytes.Buffer
	code := run(append([]string{"-bundle", f.bundlePath}, extra...), &out, &errOut)
	return code, out.String() + errOut.String()
}

func (f *fixture) all() []string {
	return []string{"-ml0", f.ml0, "-gl0", f.gl0, "-trusted", f.trustedPath}
}

func TestVerify_FullBundlePasses(t *testing.T) {
	f := newFixture(t)
	code, out := f.verify(t, f.all()...)
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	for _, want := range []string{
		"merkle path for message m4 recomputes to the root",
		"metagraph snapshot: sha256 matches bundle",
		"global snapshot: sha256 matches bundle",
		"root is present in the file's data-application on-chain state",
		"all 2 signer id(s) are in the trusted set",
		"NOT VERIFIED: snapshot signatures",
		"signatures NOT verified",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "FAIL") {
		t.Errorf("unexpected FAIL line\n%s", out)
	}
}

// The signature disclaimer is printed on every run, pass or fail, with or
// without snapshot files.
func TestVerify_AlwaysPrintsSignatureDisclaimer(t *testing.T) {
	f := newFixture(t)
	code, out := f.verify(t)
	if code != 0 || !strings.Contains(out, "NOT VERIFIED: snapshot signatures") {
		t.Fatalf("bundle only: exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "SKIP  metagraph snapshot: no file supplied") || !strings.Contains(out, "SKIP  trusted signers") {
		t.Errorf("skipped checks must be reported as skipped\n%s", out)
	}
	f.bundle.Message.Commitment = metagraph.SHA256Hex([]byte("forged"))
	code, out = f.verify(t)
	if code != 1 || !strings.Contains(out, "NOT VERIFIED: snapshot signatures") {
		t.Fatalf("failing run: exit %d\n%s", code, out)
	}
}

func TestVerify_Failures(t *testing.T) {
	cases := map[string]struct {
		mutate func(t *testing.T, f *fixture)
		want   string
	}{
		"forged commitment": {func(_ *testing.T, f *fixture) { f.bundle.Message.Commitment = metagraph.SHA256Hex([]byte("x")) }, "FAIL  merkle path"},
		"tampered sibling":  {func(_ *testing.T, f *fixture) { f.bundle.Message.Siblings[0] = metagraph.SHA256Hex([]byte("x")) }, "FAIL  merkle path"},
		"wrong leaf index":  {func(_ *testing.T, f *fixture) { f.bundle.Message.LeafIndex = 0 }, "FAIL  merkle path"},
		"wrong leaf count":  {func(_ *testing.T, f *fixture) { f.bundle.LeafCount = 6 }, "FAIL  merkle path"},
		"different root":    {func(_ *testing.T, f *fixture) { f.bundle.MerkleRoot = metagraph.SHA256Hex([]byte("other root")) }, "FAIL  merkle path"},
		"not finalized":     {func(_ *testing.T, f *fixture) { f.bundle.Finalized = false }, "does not state the root as finalized"},
		"ml0 file altered": {func(t *testing.T, f *fixture) {
			f.ml0 = f.write(t, "ml0.json", append(append([]byte(nil), f.ml0Body...), '\n'))
		}, "FAIL  metagraph snapshot: sha256"},
		"gl0 file altered": {func(t *testing.T, f *fixture) {
			f.gl0 = f.write(t, "gl0.json", snapshotFile(900, []string{"gpeer", "extra"}, nil))
		}, "FAIL  global snapshot: sha256"},
		"bundle sha256 missing": {func(_ *testing.T, f *fixture) { f.bundle.MetagraphSnapshot.SHA256 = "" }, "no sha256 to compare"},
		"untrusted ml0 signer": {func(t *testing.T, f *fixture) {
			f.trustedPath = f.write(t, "trusted.txt", []byte("peer-a\ngpeer\n"))
		}, "FAIL  metagraph snapshot: 1 signer id(s) not in the trusted set: peer-b"},
		"untrusted gl0 signer": {func(t *testing.T, f *fixture) {
			f.trustedPath = f.write(t, "trusted.txt", []byte("peer-a\npeer-b\n"))
		}, "FAIL  global snapshot: 1 signer id(s) not in the trusted set: gpeer"},
		"bundle signers differ from file": {func(_ *testing.T, f *fixture) {
			f.bundle.MetagraphSnapshot.Signers = []string{"peer-a"}
		}, "signer ids in file do not match the bundle"},
		"no signers recorded": {func(_ *testing.T, f *fixture) { f.bundle.GlobalSnapshot.Signers = nil }, "FAIL  global snapshot"},
		"ordinal mismatch":    {func(_ *testing.T, f *fixture) { f.bundle.MetagraphSnapshot.Ordinal = 43 }, "file ordinal 42 does not match bundle 43"},
		"claimed root not in snapshot file": {func(t *testing.T, f *fixture) {
			body := snapshotFile(42, []string{"peer-a", "peer-b"}, []string{})
			f.ml0 = f.write(t, "ml0.json", body)
			f.bundle.MetagraphSnapshot.SHA256 = metagraph.SHA256Hex(body)
		}, "bundle states the root is in the snapshot, but it is not in the file"},
		"snapshot file without bundle record": {func(_ *testing.T, f *fixture) { f.bundle.GlobalSnapshot = nil }, "file supplied but the bundle records no global snapshot"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			tc.mutate(t, f)
			code, out := f.verify(t, f.all()...)
			if code != 1 {
				t.Fatalf("exit %d, want 1\n%s", code, out)
			}
			if !strings.Contains(out, tc.want) || !strings.Contains(out, "RESULT: FAILED") {
				t.Fatalf("output missing %q or RESULT: FAILED\n%s", tc.want, out)
			}
		})
	}
}

// A snapshot that only bounds finality is reported as such, not as inclusion.
func TestVerify_BoundOnlySnapshotIsNotReportedAsInclusion(t *testing.T) {
	f := newFixture(t)
	no := false
	body := snapshotFile(42, []string{"peer-a", "peer-b"}, []string{})
	f.ml0 = f.write(t, "ml0.json", body)
	f.bundle.MetagraphSnapshot.SHA256 = metagraph.SHA256Hex(body)
	f.bundle.MetagraphSnapshot.ContainsRoot = &no
	code, out := f.verify(t, f.all()...)
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "root NOT shown inside it") || strings.Contains(out, "root is present in the file") {
		t.Fatalf("bound-only snapshot misreported\n%s", out)
	}
}

func TestVerify_UsageErrors(t *testing.T) {
	f := newFixture(t)
	var out, errOut bytes.Buffer
	if code := run(nil, &out, &errOut); code != 2 {
		t.Errorf("no args: exit %d", code)
	}
	if code := run([]string{"-bundle", filepath.Join(f.dir, "missing.json")}, &out, &errOut); code != 2 {
		t.Errorf("missing bundle: exit %d", code)
	}
	bad := f.write(t, "bad.json", []byte("{not json"))
	if code := run([]string{"-bundle", bad}, &out, &errOut); code != 2 {
		t.Errorf("bad bundle: exit %d", code)
	}
	if code, _ := f.verify(t, "-trusted", filepath.Join(f.dir, "missing.txt")); code != 2 {
		t.Errorf("missing trusted file: exit %d", code)
	}
	if code, o := f.verify(t, "-ml0", filepath.Join(f.dir, "missing-ml0.json")); code != 1 || !strings.Contains(o, "FAIL  metagraph snapshot: read") {
		t.Errorf("missing snapshot file: exit %d\n%s", code, o)
	}
}
