package besu

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// testdata/proof.json is a real proof recorded from a four-validator chain
// (BESU_UPDATE_FIXTURE=1 go test ./internal/besu -run TestLiveAnchorProveVerify).
func loadFixture(t *testing.T) (*Proof, []common.Address) {
	t.Helper()
	raw, err := os.ReadFile("testdata/proof.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Validators []common.Address `json:"validators"`
		Proof      *Proof           `json:"proof"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f.Proof, f.Validators
}

// clone deep-copies a proof so a test can tamper with it.
func clone(t *testing.T, p *Proof) *Proof {
	t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var out Proof
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return &out
}

func TestVerifyRecordedProof(t *testing.T) {
	p, validators := loadFixture(t)
	v, err := Verify(p, validators)
	if err != nil {
		t.Fatalf("recorded proof must verify offline: %v", err)
	}
	if v.Root != p.Root || v.LeafCount != 7 || v.AnchoredBlock == 0 || v.AnchoredBlock > v.ProvenAtBlock {
		t.Fatalf("verified = %+v", v)
	}
	if v.TrustedCount != 4 || v.QuorumRequired != 3 || len(v.Signers) < 3 {
		t.Fatalf("signers %d, quorum %d of %d", len(v.Signers), v.QuorumRequired, v.TrustedCount)
	}
}

func TestVerifyRejectsTampering(t *testing.T) {
	good, validators := loadFixture(t)

	// reseal rewrites the header with fn applied and returns the new RLP.
	rewriteHeader := func(t *testing.T, p *Proof, fn func(h *types.Header, extra *qbftExtra)) {
		t.Helper()
		var h types.Header
		if err := rlp.DecodeBytes(p.HeaderRLP, &h); err != nil {
			t.Fatal(err)
		}
		var extra qbftExtra
		if err := rlp.DecodeBytes(h.Extra, &extra); err != nil {
			t.Fatal(err)
		}
		fn(&h, &extra)
		enc, err := rlp.EncodeToBytes([]interface{}{extra.Vanity, extra.Validators, extra.Vote, extra.Round, extra.Seals})
		if err != nil {
			t.Fatal(err)
		}
		h.Extra = enc
		out, err := rlp.EncodeToBytes(&h)
		if err != nil {
			t.Fatal(err)
		}
		p.HeaderRLP = out
	}

	cases := []struct {
		name    string
		tamper  func(t *testing.T, p *Proof)
		trusted []common.Address
		want    string
	}{
		{"a different root", func(t *testing.T, p *Proof) { p.Root[0] ^= 1 }, validators, "is not anchored"},
		{"a different contract", func(t *testing.T, p *Proof) { p.Contract[0] ^= 1 }, validators, "does not exist"},
		{"state root changed after signing", func(t *testing.T, p *Proof) {
			rewriteHeader(t, p, func(h *types.Header, _ *qbftExtra) { h.Root[0] ^= 1 })
		}, validators, "not a trusted validator"},
		{"timestamp changed after signing", func(t *testing.T, p *Proof) {
			rewriteHeader(t, p, func(h *types.Header, _ *qbftExtra) { h.Time++ })
		}, validators, "not a trusted validator"},
		{"signatures stripped below quorum", func(t *testing.T, p *Proof) {
			rewriteHeader(t, p, func(_ *types.Header, e *qbftExtra) { e.Seals = e.Seals[:2] })
		}, validators, "are required"},
		{"one signature repeated to fake a quorum", func(t *testing.T, p *Proof) {
			rewriteHeader(t, p, func(_ *types.Header, e *qbftExtra) { e.Seals = [][]byte{e.Seals[0], e.Seals[0], e.Seals[0]} })
		}, validators, "are required"},
		{"storage proof node altered", func(t *testing.T, p *Proof) {
			n := p.StorageProof[len(p.StorageProof)-1]
			n[len(n)-1] ^= 1
		}, validators, "storage proof"},
		{"verifier trusts different validators", func(*testing.T, *Proof) {}, []common.Address{common.HexToAddress("0x01"), common.HexToAddress("0x02"), common.HexToAddress("0x03"), common.HexToAddress("0x04")}, "not a trusted validator"},
		{"verifier trusts only two of the real validators", func(*testing.T, *Proof) {}, validators[:2], "not a trusted validator"},
		{"no trusted validators", func(*testing.T, *Proof) {}, nil, "no trusted validators"},
		{"empty header", func(t *testing.T, p *Proof) { p.HeaderRLP = hexutil.Bytes{} }, validators, "decode header"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := clone(t, good)
			tc.tamper(t, p)
			_, err := Verify(p, tc.trusted)
			if err == nil {
				t.Fatal("tampered proof verified")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestQuorum(t *testing.T) {
	for n, want := range map[int]int{1: 1, 2: 2, 3: 2, 4: 3, 5: 4, 6: 4, 7: 5} {
		if got := Quorum(n); got != want {
			t.Errorf("Quorum(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestParseHash32(t *testing.T) {
	ok := strings.Repeat("ab", 32)
	if _, err := parseHash32("root", ok); err != nil {
		t.Fatal(err)
	}
	if _, err := parseHash32("root", "0x"+ok); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]string{
		"short":        ok[:62],
		"long":         ok + "ab",
		"not hex":      strings.Repeat("zz", 32),
		"zero":         strings.Repeat("00", 32),
		"an email":     "someone@example.com",
		"a did":        "did:key:zDnaerDaTF5BXEavCrfRZEk316dpbLsfPDZ3WJ5hRTPFU2169",
		"empty string": "",
	} {
		if _, err := parseHash32("root", bad); err == nil {
			t.Errorf("%s: accepted %q", name, bad)
		}
	}
}

func TestAnchorRejectsBadInputBeforeTouchingTheChain(t *testing.T) {
	c, err := New(Config{RPCURLs: []string{"http://127.0.0.1:9"}, AnchorContract: common.HexToAddress("0x01")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AnchorRoot(t.Context(), "nope", 1); err == nil || !strings.Contains(err.Error(), "64 hex") {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.AnchorRoot(t.Context(), strings.Repeat("ab", 32), 0); err == nil || !strings.Contains(err.Error(), "leaf count") {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.AnchorCommitment(t.Context(), strings.Repeat("ab", 32), 0); err == nil || !strings.Contains(err.Error(), "epoch") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewRequiresEndpointAndContract(t *testing.T) {
	if _, err := New(Config{AnchorContract: common.HexToAddress("0x01")}); err == nil {
		t.Fatal("no RPC URL accepted")
	}
	if _, err := New(Config{RPCURLs: []string{"http://x"}}); err == nil {
		t.Fatal("no contract accepted")
	}
}

// The gateway's key file is an OpenSSL "EC PRIVATE KEY"; its address is what the contract authorises.
func TestParseKeyPEM(t *testing.T) {
	pemBytes := []byte(throwawayKeyPEM)
	key, err := ParseKeyPEM(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	// Signing with the parsed key must recover to the parsed key's address.
	digest := crypto.Keccak256([]byte("echo"))
	sig, err := crypto.Sign(digest, key)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := crypto.SigToPub(digest, sig)
	if err != nil {
		t.Fatal(err)
	}
	if crypto.PubkeyToAddress(*pub) != crypto.PubkeyToAddress(key.PublicKey) {
		t.Fatal("parsed key does not sign as its own address")
	}
	if got := crypto.PubkeyToAddress(key.PublicKey).Hex(); got != throwawayKeyAddress {
		t.Fatalf("address = %s, want %s", got, throwawayKeyAddress)
	}
	if _, err := ParseKeyPEM([]byte("not a pem")); err == nil {
		t.Fatal("garbage accepted as a key")
	}
}

func TestParseAddresses(t *testing.T) {
	got, err := ParseAddresses(" 0x0000000000000000000000000000000000000001, 0x0000000000000000000000000000000000000002 ,")
	if err != nil || len(got) != 2 {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := ParseAddresses("0x01,nonsense"); err == nil {
		t.Fatal("malformed address accepted")
	}
}

// A test vector, not a secret: a key made with `openssl ecparam -name secp256k1 -genkey -noout`
// for this test only, and the address it must produce. It controls nothing.
const throwawayKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHQCAQEEIBVTVa3+ntL15gMbZvb6V9cNRWwhm8qi8lJEUt4oVR9OoAcGBSuBBAAK
oUQDQgAEVZpRYOFAMptTU8Hw7vI7EjrfoQxg/kj0Wy/eOkISwLr9ZDUBVRg0EOxD
9jhWrn507g/Ibot+wxXXYjNGtLMieA==
-----END EC PRIVATE KEY-----
`

const throwawayKeyAddress = "0x29d1DFBFAB99d15A532d000f88924D1fDF3aA3B1"
