// Command verifyproof checks an Echo anchor proof bundle offline.
//
// Usage:
//
//	verifyproof -bundle bundle.json [-ml0 ml0-snapshot.json] [-gl0 gl0-snapshot.json] [-trusted peers.txt]
//
// The bundle comes from GET /v1/anchors/{root}/proof[?messageId=...]. The
// snapshot files are the archived objects named by the bundle's archiveKey
// fields. The trusted file lists validator peer ids, one per line ('#' starts a
// comment); list both metagraph and Global L0 validators if both snapshots are
// in the bundle.
//
// It checks what can be checked without Tessellation:
//   - the message's Merkle path recomputes to the bundle's root;
//   - each supplied snapshot file's SHA-256 matches the bundle, its ordinal and
//     signer ids match the bundle, and (metagraph snapshot) whether the root is
//     present in its data-application on-chain state;
//   - every signer id is in the trusted set.
//
// It does NOT verify the secp256k1 snapshot signatures, and does not recompute
// the snapshot hashes. Both need Tessellation's canonical serialization. A
// passing run therefore shows the bundle is internally consistent and names
// only trusted signers; it does not show those signers actually signed.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

const signatureDisclaimer = `NOT VERIFIED: snapshot signatures (secp256k1) and snapshot hashes.
  Signer ids are read from the bundle/snapshot "proofs" and compared with the
  trusted list, but the signatures themselves are not checked: that requires
  Tessellation's canonical serialization of the snapshot. Verify them with
  Tessellation tooling before relying on this proof against a party who could
  have altered the archived snapshot files.`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type report struct {
	out    io.Writer
	failed bool
}

func (r *report) ok(format string, a ...any)   { fmt.Fprintf(r.out, "OK    "+format+"\n", a...) }
func (r *report) skip(format string, a ...any) { fmt.Fprintf(r.out, "SKIP  "+format+"\n", a...) }
func (r *report) note(format string, a ...any) { fmt.Fprintf(r.out, "NOTE  "+format+"\n", a...) }
func (r *report) fail(format string, a ...any) {
	r.failed = true
	fmt.Fprintf(r.out, "FAIL  "+format+"\n", a...)
}

// run returns the process exit code: 0 all checks passed, 1 a check failed,
// 2 usage or input error.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("verifyproof", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bundlePath := fs.String("bundle", "", "proof bundle JSON file (required)")
	ml0Path := fs.String("ml0", "", "archived Metagraph L0 snapshot JSON file (optional)")
	gl0Path := fs.String("gl0", "", "archived Global L0 snapshot JSON file (optional)")
	trustedPath := fs.String("trusted", "", "file of trusted validator peer ids, one per line (optional)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *bundlePath == "" {
		fmt.Fprintln(stderr, "verifyproof: -bundle is required")
		fs.Usage()
		return 2
	}
	raw, err := os.ReadFile(*bundlePath)
	if err != nil {
		fmt.Fprintf(stderr, "verifyproof: read bundle: %v\n", err)
		return 2
	}
	var b metagraph.ProofBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		fmt.Fprintf(stderr, "verifyproof: decode bundle: %v\n", err)
		return 2
	}
	if b.MerkleRoot == "" {
		fmt.Fprintln(stderr, "verifyproof: bundle has no merkleRoot")
		return 2
	}
	var trusted map[string]struct{}
	if *trustedPath != "" {
		if trusted, err = readTrusted(*trustedPath); err != nil {
			fmt.Fprintf(stderr, "verifyproof: read trusted peers: %v\n", err)
			return 2
		}
	}

	r := &report{out: stdout}
	fmt.Fprintf(stdout, "network: %s\nroot:    %s\n", b.Network, b.MerkleRoot)

	if b.Finalized {
		r.note("bundle states the root was observed finalized (gateway's record; not independently checkable offline)")
	} else {
		r.fail("bundle does not state the root as finalized")
	}

	checkMessage(r, &b)
	checkSnapshot(r, "metagraph", b.MerkleRoot, b.MetagraphSnapshot, *ml0Path, true)
	checkSnapshot(r, "global", b.MerkleRoot, b.GlobalSnapshot, *gl0Path, false)
	checkSigners(r, &b, trusted, *trustedPath != "")

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, signatureDisclaimer)
	fmt.Fprintln(stdout)
	if r.failed {
		fmt.Fprintln(stdout, "RESULT: FAILED (one or more checks failed)")
		return 1
	}
	fmt.Fprintln(stdout, "RESULT: checks performed passed (signatures NOT verified, see above)")
	return 0
}

func checkMessage(r *report, b *metagraph.ProofBundle) {
	m := b.Message
	if m == nil {
		r.skip("merkle path: bundle has no message (request it with ?messageId=)")
		return
	}
	if err := metagraph.VerifyMerklePath(m.Commitment, m.Siblings, m.LeafIndex, b.LeafCount, b.MerkleRoot); err != nil {
		r.fail("merkle path for message %s: %v", m.MessageID, err)
		return
	}
	r.ok("merkle path for message %s recomputes to the root (leaf %d of %d)", m.MessageID, m.LeafIndex, b.LeafCount)
}

func checkSnapshot(r *report, name, root string, ref *metagraph.SnapshotRef, path string, isMetagraph bool) {
	if ref == nil {
		if path != "" {
			r.fail("%s snapshot: file supplied but the bundle records no %s snapshot", name, name)
		} else {
			r.skip("%s snapshot: none recorded in the bundle", name)
		}
		return
	}
	if isMetagraph {
		if ref.ContainsRoot != nil && *ref.ContainsRoot {
			r.note("metagraph snapshot %d: bundle states the root is in its on-chain state", ref.Ordinal)
		} else {
			r.note("metagraph snapshot %d: root NOT shown inside it; it only bounds finality (observed at or before this ordinal)", ref.Ordinal)
		}
	} else {
		r.note("global snapshot %d: only bounds the observation time; inclusion of the metagraph snapshot in it is not established", ref.Ordinal)
	}
	if path == "" {
		r.skip("%s snapshot: no file supplied; sha256 and contents not checked", name)
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		r.fail("%s snapshot: read %s: %v", name, path, err)
		return
	}
	if ref.SHA256 == "" {
		r.fail("%s snapshot: bundle has no sha256 to compare the file against", name)
	} else if got := metagraph.SHA256Hex(body); got != ref.SHA256 {
		r.fail("%s snapshot: sha256 %s does not match bundle %s", name, got, ref.SHA256)
	} else {
		r.ok("%s snapshot: sha256 matches bundle (%s)", name, got)
	}
	info, err := metagraph.ParseSignedSnapshot(body)
	if err != nil {
		r.fail("%s snapshot: %v", name, err)
		return
	}
	if info.Ordinal != ref.Ordinal {
		r.fail("%s snapshot: file ordinal %d does not match bundle %d", name, info.Ordinal, ref.Ordinal)
	} else {
		r.ok("%s snapshot: ordinal %d matches bundle", name, info.Ordinal)
	}
	if !sameSet(info.Signers, ref.Signers) {
		r.fail("%s snapshot: signer ids in file do not match the bundle", name)
	} else {
		r.ok("%s snapshot: %d signer id(s) in file match the bundle", name, len(info.Signers))
	}
	if !isMetagraph {
		return
	}
	claimed := ref.ContainsRoot != nil && *ref.ContainsRoot
	switch {
	case info.ContainsRoot(root):
		r.ok("metagraph snapshot: root is present in the file's data-application on-chain state")
	case claimed:
		r.fail("metagraph snapshot: bundle states the root is in the snapshot, but it is not in the file")
	case !info.HasDataState:
		r.note("metagraph snapshot: no decodable data-application on-chain state in the file")
	default:
		r.note("metagraph snapshot: root is not in the file's on-chain state (consistent with the bundle)")
	}
}

func checkSigners(r *report, b *metagraph.ProofBundle, trusted map[string]struct{}, haveTrusted bool) {
	if !haveTrusted {
		r.skip("trusted signers: no -trusted file supplied")
		return
	}
	check := func(name string, ref *metagraph.SnapshotRef) {
		if ref == nil {
			return
		}
		if len(ref.Signers) == 0 {
			r.fail("%s snapshot: no signer ids recorded", name)
			return
		}
		var unknown []string
		for _, id := range ref.Signers {
			if _, ok := trusted[strings.ToLower(id)]; !ok {
				unknown = append(unknown, id)
			}
		}
		if len(unknown) > 0 {
			r.fail("%s snapshot: %d signer id(s) not in the trusted set: %s", name, len(unknown), strings.Join(unknown, ", "))
			return
		}
		r.ok("%s snapshot: all %d signer id(s) are in the trusted set (ids only; signatures NOT verified)", name, len(ref.Signers))
	}
	if b.MetagraphSnapshot == nil && b.GlobalSnapshot == nil {
		r.fail("trusted signers: bundle records no snapshot signers to check")
		return
	}
	check("metagraph", b.MetagraphSnapshot)
	check("global", b.GlobalSnapshot)
}

func readTrusted(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]struct{}{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		if line = strings.ToLower(strings.TrimSpace(line)); line != "" {
			out[line] = struct{}{}
		}
	}
	return out, sc.Err()
}

func sameSet(a, b []string) bool {
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
