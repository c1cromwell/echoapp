package didkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// Method is the DID method name.
const Method = "key"

// Prefix is the canonical did:key URI prefix including the multibase 'z'
// (base58btc) sigil. All P-256 did:keys produced by Derive begin with this.
const Prefix = "did:key:z"

// p256MulticodecVarint is the unsigned-varint encoding of multicodec
// "p256-pub" (0x1200). 0x1200 = (0x24 << 7) | 0x00, so the varint bytes are
// [0x80, 0x24]: low 7 bits 0x00 with continuation set, then high 7 bits 0x24.
var p256MulticodecVarint = []byte{0x80, 0x24}

// compressedP256Len is the length of a SEC1 compressed P-256 public key:
// 1 byte sign indicator (0x02 or 0x03) + 32-byte big-endian X coordinate.
const compressedP256Len = 33

// Errors returned by this package.
var (
	ErrInvalidDID         = errors.New("didkey: not a valid did:key")
	ErrUnsupportedCodec   = errors.New("didkey: only multicodec p256-pub (0x1200) is supported")
	ErrInvalidPublicKey   = errors.New("didkey: invalid P-256 public key")
	ErrInvalidCompression = errors.New("didkey: invalid SEC1 compressed point prefix")
)

// Derive returns the canonical did:key string for the given P-256 public key.
//
// The derivation is deterministic: the same public key always produces the
// same DID. No network call is made; no chain transaction is required.
func Derive(pub *ecdsa.PublicKey) (string, error) {
	if pub == nil || pub.Curve == nil {
		return "", ErrInvalidPublicKey
	}
	if pub.Curve != elliptic.P256() {
		return "", fmt.Errorf("%w: expected P-256, got %s", ErrInvalidPublicKey, pub.Curve.Params().Name)
	}
	if pub.X == nil || pub.Y == nil {
		return "", ErrInvalidPublicKey
	}

	compressed, err := compressP256(pub)
	if err != nil {
		return "", err
	}

	body := make([]byte, 0, len(p256MulticodecVarint)+len(compressed))
	body = append(body, p256MulticodecVarint...)
	body = append(body, compressed...)

	return Prefix + base58Encode(body), nil
}

// DeriveFromPEM is a convenience wrapper that parses a PEM-encoded SubjectPublicKeyInfo
// (the format emitted by `openssl ec -pubout`) and derives the DID. This is
// useful for shell-based test harnesses (e.g. scripts/validate-phase1.sh).
func DeriveFromPEM(pemBytes []byte) (string, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", fmt.Errorf("%w: not a PEM block", ErrInvalidPublicKey)
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPublicKey, err)
	}
	pub, ok := pubAny.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("%w: PEM does not contain an ECDSA public key", ErrInvalidPublicKey)
	}
	return Derive(pub)
}

// DeriveFromPublicKeyHex parses an uncompressed (0x04 || X || Y, 65 bytes) or
// compressed (0x02/0x03 || X, 33 bytes) hex-encoded P-256 public key and
// derives the DID. This matches the wire format used by the
// POST /identity/register handler.
func DeriveFromPublicKeyHex(hexStr string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimPrefix(hexStr, "0x"))
	if err != nil {
		return "", fmt.Errorf("%w: hex decode: %v", ErrInvalidPublicKey, err)
	}
	pub, err := publicKeyFromBytes(raw)
	if err != nil {
		return "", err
	}
	return Derive(pub)
}

// Parse takes a did:key string and returns the embedded P-256 public key.
// Resolution is purely local: no network call is made.
func Parse(did string) (*ecdsa.PublicKey, error) {
	if !strings.HasPrefix(did, Prefix) {
		return nil, fmt.Errorf("%w: missing %q prefix", ErrInvalidDID, Prefix)
	}
	encoded := strings.TrimPrefix(did, Prefix)
	if encoded == "" {
		return nil, ErrInvalidDID
	}

	body, err := base58Decode(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDID, err)
	}
	if len(body) < len(p256MulticodecVarint)+compressedP256Len {
		return nil, fmt.Errorf("%w: payload too short", ErrInvalidDID)
	}

	if body[0] != p256MulticodecVarint[0] || body[1] != p256MulticodecVarint[1] {
		return nil, fmt.Errorf("%w: codec prefix %02x %02x", ErrUnsupportedCodec, body[0], body[1])
	}

	pubBytes := body[len(p256MulticodecVarint):]
	if len(pubBytes) != compressedP256Len {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidPublicKey, compressedP256Len, len(pubBytes))
	}

	return decompressP256(pubBytes)
}

// PublicKeyHexUncompressed returns SEC1 uncompressed P-256 hex (04 || X || Y).
func PublicKeyHexUncompressed(pub *ecdsa.PublicKey) (string, error) {
	if pub == nil || pub.Curve == nil {
		return "", ErrInvalidPublicKey
	}
	raw := elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
	if len(raw) != 65 || raw[0] != 0x04 {
		return "", fmt.Errorf("%w: unexpected uncompressed length", ErrInvalidPublicKey)
	}
	return hex.EncodeToString(raw), nil
}

// MustDerive is the panicking variant of Derive — useful for tests and
// constant-defined fixtures only.
func MustDerive(pub *ecdsa.PublicKey) string {
	d, err := Derive(pub)
	if err != nil {
		panic(err)
	}
	return d
}

// --- internal helpers ---

// compressP256 encodes a P-256 public key in SEC1 compressed form
// (0x02 || X if Y is even, 0x03 || X if Y is odd).
func compressP256(pub *ecdsa.PublicKey) ([]byte, error) {
	xBytes := pub.X.Bytes()
	if len(xBytes) > 32 {
		return nil, fmt.Errorf("%w: X coordinate larger than 32 bytes", ErrInvalidPublicKey)
	}
	out := make([]byte, compressedP256Len)
	if pub.Y.Bit(0) == 0 {
		out[0] = 0x02
	} else {
		out[0] = 0x03
	}
	copy(out[1+(32-len(xBytes)):], xBytes)
	return out, nil
}

// decompressP256 reconstructs a P-256 public key from its SEC1 compressed
// form. Implements the standard square-root-mod-p recovery for the y
// coordinate. Returns ErrInvalidPublicKey if the recovered point is not on
// the curve.
func decompressP256(in []byte) (*ecdsa.PublicKey, error) {
	if len(in) != compressedP256Len {
		return nil, fmt.Errorf("%w: compressed key wrong length", ErrInvalidPublicKey)
	}
	prefix := in[0]
	if prefix != 0x02 && prefix != 0x03 {
		return nil, ErrInvalidCompression
	}
	curve := elliptic.P256()
	params := curve.Params()
	x := new(big.Int).SetBytes(in[1:])
	if x.Sign() < 0 || x.Cmp(params.P) >= 0 {
		return nil, fmt.Errorf("%w: X out of range", ErrInvalidPublicKey)
	}

	// y^2 = x^3 - 3x + b (mod p) for P-256
	xCubed := new(big.Int).Mul(x, x)
	xCubed.Mul(xCubed, x)
	threeX := new(big.Int).Lsh(x, 1)
	threeX.Add(threeX, x)
	rhs := new(big.Int).Sub(xCubed, threeX)
	rhs.Add(rhs, params.B)
	rhs.Mod(rhs, params.P)

	y := new(big.Int).ModSqrt(rhs, params.P)
	if y == nil {
		return nil, fmt.Errorf("%w: no square root for X", ErrInvalidPublicKey)
	}
	if y.Bit(0) != uint(prefix&1) {
		y.Sub(params.P, y)
	}

	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("%w: decompressed point not on curve", ErrInvalidPublicKey)
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

// publicKeyFromBytes parses a raw SEC1 P-256 public key (compressed or
// uncompressed) into an ecdsa.PublicKey.
func publicKeyFromBytes(raw []byte) (*ecdsa.PublicKey, error) {
	switch len(raw) {
	case compressedP256Len:
		if raw[0] != 0x02 && raw[0] != 0x03 {
			return nil, ErrInvalidCompression
		}
		return decompressP256(raw)
	case 65:
		if raw[0] != 0x04 {
			return nil, fmt.Errorf("%w: uncompressed key must start with 0x04", ErrInvalidPublicKey)
		}
		curve := elliptic.P256()
		x := new(big.Int).SetBytes(raw[1:33])
		y := new(big.Int).SetBytes(raw[33:65])
		if !curve.IsOnCurve(x, y) {
			return nil, fmt.Errorf("%w: uncompressed point not on curve", ErrInvalidPublicKey)
		}
		return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
	case 64:
		// Some APIs (incl. WebCrypto raw export) emit X||Y without the 0x04 sigil.
		curve := elliptic.P256()
		x := new(big.Int).SetBytes(raw[:32])
		y := new(big.Int).SetBytes(raw[32:])
		if !curve.IsOnCurve(x, y) {
			return nil, fmt.Errorf("%w: raw X||Y point not on curve", ErrInvalidPublicKey)
		}
		return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported public-key length %d", ErrInvalidPublicKey, len(raw))
	}
}
