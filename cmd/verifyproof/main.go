// Command verifyproof checks an Echo anchor proof bundle with no access to Echo or its chain.
//
// Usage:
//
//	verifyproof -bundle bundle.json -validators 0xAAA...,0xBBB...,0xCCC...,0xDDD...
//	verifyproof -bundle bundle.json -validators-file validators.txt
//
// The bundle comes from GET /v1/anchors/{root}/proof[?messageId=...]. The validator addresses are
// the verifier's only trust anchor: obtain them from a source you trust (Echo publishes them), not
// from the bundle.
//
// A passing run establishes, by checking signatures and hashes only:
//  1. (when the bundle carries a message) the message commitment is a leaf under the root;
//  2. the root is recorded in Echo's anchor contract; and
//  3. the block that record is proven against was signed by a quorum of the given validators.
//
// It does not establish what the committed message said (the commitment is a hash), nor that the
// validators are independent of Echo: Echo runs them.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/thechadcromwell/echoapp/internal/besu"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type report struct {
	out    io.Writer
	failed bool
}

func (r *report) ok(format string, a ...any)   { fmt.Fprintf(r.out, "OK    "+format+"\n", a...) }
func (r *report) skip(format string, a ...any) { fmt.Fprintf(r.out, "SKIP  "+format+"\n", a...) }
func (r *report) fail(format string, a ...any) {
	r.failed = true
	fmt.Fprintf(r.out, "FAIL  "+format+"\n", a...)
}

// run returns the process exit code: 0 every check passed, 1 a check failed, 2 usage or input error.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("verifyproof", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bundlePath := fs.String("bundle", "", "proof bundle JSON file (required)")
	validatorList := fs.String("validators", "", "comma-separated validator addresses you trust")
	validatorFile := fs.String("validators-file", "", "file of validator addresses you trust, one per line ('#' starts a comment)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	usage := func(msg string) int {
		fmt.Fprintln(stderr, "verifyproof:", msg)
		fs.Usage()
		return 2
	}
	if *bundlePath == "" {
		return usage("-bundle is required")
	}
	list := *validatorList
	if *validatorFile != "" {
		fromFile, err := readList(*validatorFile)
		if err != nil {
			return usage("read validators file: " + err.Error())
		}
		list = strings.Join(append(strings.Split(list, ","), fromFile...), ",")
	}
	validators, err := besu.ParseAddresses(list)
	if err != nil {
		return usage(err.Error())
	}
	if len(validators) == 0 {
		return usage("no trusted validators given (-validators or -validators-file); without them nothing can be verified")
	}
	raw, err := os.ReadFile(*bundlePath)
	if err != nil {
		return usage("read bundle: " + err.Error())
	}
	var b metagraph.ProofBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return usage("decode bundle: " + err.Error())
	}
	if b.MerkleRoot == "" {
		return usage("bundle has no merkleRoot")
	}

	r := &report{out: stdout}
	fmt.Fprintf(stdout, "network: %s\nroot:    %s\n", b.Network, b.MerkleRoot)

	// Link 1: message -> root.
	if b.Message == nil {
		r.skip("no message in the bundle; only the root itself is checked")
	} else if err := metagraph.VerifyMerklePath(b.Message.Commitment, b.Message.Siblings, b.Message.LeafIndex, b.LeafCount, b.MerkleRoot); err != nil {
		r.fail("message %s is not under the root: %v", b.Message.MessageID, err)
	} else {
		r.ok("message %s (commitment %s) is leaf %d of %d under the root", b.Message.MessageID, b.Message.Commitment, b.Message.LeafIndex, b.LeafCount)
	}

	// Link 2: root -> a block signed by the validators.
	verifyChain(r, &b, validators)

	if r.failed {
		fmt.Fprintln(stdout, "RESULT: FAILED")
		return 1
	}
	fmt.Fprintln(stdout, "RESULT: VERIFIED")
	return 0
}

func verifyChain(r *report, b *metagraph.ProofBundle, validators []common.Address) {
	if b.Chain == nil || len(b.Chain.Proof) == 0 {
		r.fail("the bundle carries no chain proof, so nothing shows the root is anchored")
		return
	}
	var proof besu.Proof
	if err := json.Unmarshal(b.Chain.Proof, &proof); err != nil {
		r.fail("chain proof is malformed: %v", err)
		return
	}
	if !strings.EqualFold(strings.TrimPrefix(proof.Root.Hex(), "0x"), strings.TrimPrefix(b.MerkleRoot, "0x")) {
		r.fail("chain proof is for root %s, not the bundle's root", proof.Root.Hex())
		return
	}
	v, err := besu.Verify(&proof, validators)
	if err != nil {
		r.fail("chain proof: %v", err)
		return
	}
	r.ok("root is recorded in contract %s, anchored in block %d", v.Contract.Hex(), v.AnchoredBlock)
	r.ok("block %d (%s) carries valid signatures from %d of your %d validators (%d required)",
		v.ProvenAtBlock, v.ProvenAtTime.Format(time.RFC3339), len(v.Signers), v.TrustedCount, v.QuorumRequired)
	if int(v.LeafCount) != b.LeafCount {
		r.fail("the chain recorded %d leaves under this root; the bundle says %d", v.LeafCount, b.LeafCount)
	} else {
		r.ok("leaf count on the chain matches the bundle (%d)", v.LeafCount)
	}
	if b.Chain.BlockNumber != 0 && uint64(b.Chain.BlockNumber) != v.AnchoredBlock {
		r.fail("the bundle says block %d; the chain proof says block %d", b.Chain.BlockNumber, v.AnchoredBlock)
	}
	if b.Chain.Contract != "" && !strings.EqualFold(b.Chain.Contract, v.Contract.Hex()) {
		r.fail("the bundle names contract %s; the chain proof is for %s", b.Chain.Contract, v.Contract.Hex())
	}
}

func readList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(strings.SplitN(sc.Text(), "#", 2)[0])
		if line != "" {
			out = append(out, line)
		}
	}
	return out, sc.Err()
}
