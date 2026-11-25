package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/ledger"
)

// testdata/bundle.json is a real bundle: three message commitments anchored on a four-validator
// chain through the gateway's own code path, as served by GET /v1/anchors/{root}/proof
// (BESU_UPDATE_FIXTURE=1 go test ./internal/metagraph -run TestLiveAnchorBatch).
func fixture(t *testing.T) (*ledger.ProofBundle, string) {
	t.Helper()
	raw, err := os.ReadFile("testdata/bundle.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Validators string              `json:"validators"`
		Bundle     *ledger.ProofBundle `json:"bundle"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f.Bundle, f.Validators
}

// runWith writes the bundle to a temp file and runs the verifier on it.
func runWith(t *testing.T, b *ledger.ProofBundle, extra ...string) (int, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bundle.json")
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := run(append([]string{"-bundle", path}, extra...), &out, &errOut)
	return code, out.String() + errOut.String()
}

// copyBundle deep-copies so a test can tamper with it.
func copyBundle(t *testing.T, b *ledger.ProofBundle) *ledger.ProofBundle {
	t.Helper()
	raw, _ := json.Marshal(b)
	var out ledger.ProofBundle
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return &out
}

func TestVerify_RealBundlePasses(t *testing.T) {
	b, validators := fixture(t)
	code, out := runWith(t, b, "-validators", validators)
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	for _, want := range []string{
		"is leaf 1 of 3 under the root",
		"root is recorded in contract",
		"carries valid signatures from",
		"leaf count on the chain matches the bundle (3)",
		"RESULT: VERIFIED",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "FAIL") {
		t.Errorf("a passing run printed FAIL\n%s", out)
	}
}

func TestVerify_ValidatorsFromFile(t *testing.T) {
	b, validators := fixture(t)
	path := filepath.Join(t.TempDir(), "validators.txt")
	body := "# Echo validators\n" + strings.ReplaceAll(validators, ",", "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, out := runWith(t, b, "-validators-file", path); code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
}

func TestVerify_RootOnlyBundle(t *testing.T) {
	b, validators := fixture(t)
	b = copyBundle(t, b)
	b.Message = nil
	code, out := runWith(t, b, "-validators", validators)
	if code != 0 || !strings.Contains(out, "SKIP  no message in the bundle") {
		t.Fatalf("exit %d\n%s", code, out)
	}
}

func TestVerify_Failures(t *testing.T) {
	good, validators := fixture(t)
	strangers := "0x0000000000000000000000000000000000000001,0x0000000000000000000000000000000000000002,0x0000000000000000000000000000000000000003,0x0000000000000000000000000000000000000004"
	twoOfFour := strings.Join(strings.Split(validators, ",")[:2], ",")

	cases := []struct {
		name       string
		tamper     func(b *ledger.ProofBundle)
		validators string
		want       string
	}{
		{"validators the block was not signed by", func(*ledger.ProofBundle) {}, strangers, "not a trusted validator"},
		{"only two of the four real validators trusted", func(*ledger.ProofBundle) {}, twoOfFour, "not a trusted validator"},
		{"chain proof removed", func(b *ledger.ProofBundle) { b.Chain = nil }, validators, "carries no chain proof"},
		{"chain proof emptied", func(b *ledger.ProofBundle) { b.Chain.Proof = nil }, validators, "carries no chain proof"},
		{"chain proof is not JSON for a proof", func(b *ledger.ProofBundle) { b.Chain.Proof = json.RawMessage(`"x"`) }, validators, "malformed"},
		{"bundle claims a different root than the proof", func(b *ledger.ProofBundle) {
			b.MerkleRoot = strings.Repeat("ab", 32)
		}, validators, "FAIL"},
		{"bundle claims more leaves than the chain recorded", func(b *ledger.ProofBundle) {
			b.LeafCount = 4
		}, validators, "FAIL"},
		{"bundle claims a different block", func(b *ledger.ProofBundle) { b.Chain.BlockNumber++ }, validators, "the bundle says block"},
		{"bundle names a different contract", func(b *ledger.ProofBundle) {
			b.Chain.Contract = "0x00000000000000000000000000000000000000ff"
		}, validators, "the bundle names contract"},
		{"message commitment swapped", func(b *ledger.ProofBundle) {
			b.Message.Commitment = strings.Repeat("cd", 32)
		}, validators, "is not under the root"},
		{"message path sibling altered", func(b *ledger.ProofBundle) {
			b.Message.Siblings[0] = strings.Repeat("ef", 32)
		}, validators, "is not under the root"},
		{"message claimed at another position", func(b *ledger.ProofBundle) { b.Message.LeafIndex = 0 }, validators, "is not under the root"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := copyBundle(t, good)
			tc.tamper(b)
			code, out := runWith(t, b, "-validators", tc.validators)
			if code != 1 {
				t.Fatalf("exit %d, want 1\n%s", code, out)
			}
			if !strings.Contains(out, tc.want) || !strings.Contains(out, "RESULT: FAILED") {
				t.Fatalf("output should mention %q and RESULT: FAILED\n%s", tc.want, out)
			}
		})
	}
}

func TestVerify_UsageErrors(t *testing.T) {
	good, validators := fixture(t)
	if code, _ := runWith(t, good); code != 2 {
		t.Errorf("no validators: exit %d, want 2 (nothing can be verified without a trust anchor)", code)
	}
	if code, _ := runWith(t, good, "-validators", "not-an-address"); code != 2 {
		t.Errorf("malformed validator: exit %d, want 2", code)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"-validators", validators}, &out, &errOut); code != 2 {
		t.Errorf("no bundle: exit %d, want 2", code)
	}
	if code := run([]string{"-bundle", "/nonexistent", "-validators", validators}, &out, &errOut); code != 2 {
		t.Errorf("missing file: exit %d, want 2", code)
	}
	noRoot := copyBundle(t, good)
	noRoot.MerkleRoot = ""
	if code, _ := runWith(t, noRoot, "-validators", validators); code != 2 {
		t.Errorf("bundle with no root: exit %d, want 2", code)
	}
}
