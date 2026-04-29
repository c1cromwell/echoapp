package didkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"testing"
)

// TestDeriveRoundTripRandom generates 64 P-256 keys, derives a did:key for
// each, and asserts that Parse recovers the exact same public key.
func TestDeriveRoundTripRandom(t *testing.T) {
	for i := 0; i < 64; i++ {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		pub := &priv.PublicKey

		did, err := Derive(pub)
		if err != nil {
			t.Fatalf("Derive: %v", err)
		}

		if !strings.HasPrefix(did, Prefix) {
			t.Fatalf("DID %q missing prefix %q", did, Prefix)
		}

		got, err := Parse(did)
		if err != nil {
			t.Fatalf("Parse(%q): %v", did, err)
		}
		if got.X.Cmp(pub.X) != 0 || got.Y.Cmp(pub.Y) != 0 {
			t.Fatalf("round-trip mismatch:\n got x=%x y=%x\nwant x=%x y=%x",
				got.X, got.Y, pub.X, pub.Y)
		}
	}
}

// TestDeriveDeterministic asserts the derivation is a pure function: the same
// key produces the same DID across invocations.
func TestDeriveDeterministic(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	a := MustDerive(&priv.PublicKey)
	b := MustDerive(&priv.PublicKey)
	if a != b {
		t.Fatalf("derive not deterministic:\n a=%s\n b=%s", a, b)
	}
}

// TestDeriveFromPEM verifies the convenience PEM helper used by the bash
// validation harness produces the same DID as Derive.
func TestDeriveFromPEM(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal pkix: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

	gotPEM, err := DeriveFromPEM(pemBytes)
	if err != nil {
		t.Fatalf("DeriveFromPEM: %v", err)
	}
	gotKey := MustDerive(&priv.PublicKey)
	if gotPEM != gotKey {
		t.Fatalf("PEM derivation mismatch:\n pem=%s\n key=%s", gotPEM, gotKey)
	}
}

// TestDeriveFromHex covers the hex API used by POST /identity/register.
func TestDeriveFromHex(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	uncompressed := elliptic.Marshal(elliptic.P256(), priv.X, priv.Y) // 0x04 || X || Y
	hexStr := hex.EncodeToString(uncompressed)

	gotHex, err := DeriveFromPublicKeyHex(hexStr)
	if err != nil {
		t.Fatalf("DeriveFromPublicKeyHex: %v", err)
	}
	want := MustDerive(&priv.PublicKey)
	if gotHex != want {
		t.Fatalf("hex derivation mismatch:\n hex=%s\n key=%s", gotHex, want)
	}

	gotHex0x, err := DeriveFromPublicKeyHex("0x" + hexStr)
	if err != nil {
		t.Fatalf("DeriveFromPublicKeyHex with 0x prefix: %v", err)
	}
	if gotHex0x != want {
		t.Fatalf("0x-prefix derivation mismatch")
	}
}

// TestParseRejectsNonP256Codec ensures Parse refuses keys with a multicodec
// other than 0x1200 (p256-pub). Constructed by manually encoding an Ed25519
// codec prefix (0xed) followed by 32 random bytes.
func TestParseRejectsNonP256Codec(t *testing.T) {
	body := append([]byte{0xed, 0x01}, make([]byte, 32)...) // Ed25519 codec
	bad := Prefix + base58Encode(body)
	if _, err := Parse(bad); err == nil {
		t.Fatalf("expected ErrUnsupportedCodec, got nil")
	}
}

// TestParseRejectsTruncated ensures Parse refuses a payload that is too
// short to contain the full compressed key.
func TestParseRejectsTruncated(t *testing.T) {
	body := append([]byte{}, p256MulticodecVarint...)
	body = append(body, 0x02, 0x01, 0x02, 0x03) // way under 33 bytes
	short := Prefix + base58Encode(body)
	if _, err := Parse(short); err == nil {
		t.Fatalf("expected truncated payload to fail Parse")
	}
}

// TestParseRejectsBadPrefix ensures Parse refuses URIs that aren't did:key.
func TestParseRejectsBadPrefix(t *testing.T) {
	for _, in := range []string{
		"",
		"did:example:123",
		"did:prism:cardano:abc",
		"key:zABC",
		"did:key:",
	} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) should fail but did not", in)
		}
	}
}

// TestDeriveRejectsNonP256Curve ensures Derive refuses non-P-256 curves.
func TestDeriveRejectsNonP256Curve(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("generate p384: %v", err)
	}
	if _, err := Derive(&priv.PublicKey); err == nil {
		t.Fatalf("Derive should reject P-384 keys")
	}
}

// TestDIDKeyFormat asserts DIDs match the expected printable shape for
// downstream consumers (regex-friendly, URL-safe).
func TestDIDKeyFormat(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	did := MustDerive(&priv.PublicKey)

	if !strings.HasPrefix(did, "did:key:z") {
		t.Fatalf("missing did:key:z prefix: %s", did)
	}
	body := strings.TrimPrefix(did, "did:key:z")
	for _, c := range body {
		if !strings.ContainsRune(base58btcAlphabet, c) {
			t.Fatalf("non-base58btc character %q in DID body %s", c, body)
		}
	}
	// Encoded payload is 35 bytes; base58btc length = ceil(35 * log(256)/log(58)) ≈ 47-48 chars.
	if l := len(body); l < 46 || l > 49 {
		t.Fatalf("unexpected DID body length %d (expected 46-49) for %s", l, did)
	}
}

// TestBase58RoundTrip exercises the local base58 implementation.
func TestBase58RoundTrip(t *testing.T) {
	cases := [][]byte{
		{},
		{0},
		{0, 0, 0, 1, 2, 3},
		{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa},
	}
	for i := 0; i < 32; i++ {
		buf := make([]byte, 35)
		if _, err := rand.Read(buf); err != nil {
			t.Fatalf("rand: %v", err)
		}
		cases = append(cases, buf)
	}
	for _, in := range cases {
		out, err := base58Decode(base58Encode(in))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if string(out) != string(in) {
			t.Fatalf("base58 round-trip mismatch:\n in=%x\nout=%x", in, out)
		}
	}
}

// TestBase58DecodeRejectsInvalid ensures characters outside the alphabet are
// surfaced as errors (not silently mapped).
func TestBase58DecodeRejectsInvalid(t *testing.T) {
	for _, bad := range []string{"0OIl", "abc!def", "abc def"} {
		if _, err := base58Decode(bad); err == nil {
			t.Fatalf("base58Decode(%q) should fail", bad)
		}
	}
}
