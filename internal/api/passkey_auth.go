package api

// passkeyAuth implements WO-1 ECDSA P-256 request-signature authentication.
//
// Protocol:
//   iOS sends two extra headers on every authenticated request:
//     X-Sender-DID:  the caller's did:key (e.g. did:key:zDn...)
//     X-Signature:   base64-std-encoded ECDSA P-256 signature over SHA-256(body)
//
// Key resolution: Redis cache (60s TTL) → PostgresDIDRegistry.
// Any one of the DID's registered device keys may produce a valid signature.
//
// Error codes returned in APIError.Code:
//   AUTH_MISSING_SIGNATURE  — X-Signature header absent while X-Sender-DID present
//   AUTH_UNKNOWN_DID        — no device keys found for the supplied DID
//   AUTH_INVALID_SIGNATURE  — signature does not verify against any registered key

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

const (
	headerSenderDID = "X-Sender-DID"
	headerSignature = "X-Signature"
	headerTimestamp = "X-Timestamp"

	// maxClockSkewSeconds bounds how far a request timestamp may be from server
	// time; outside this window the request is rejected as stale (anti-replay).
	maxClockSkewSeconds = 120
	// replayWindow is how long a used signature is remembered. It must exceed
	// the freshness window so a signature cannot be reused before it goes stale.
	replayWindow = 2 * maxClockSkewSeconds * time.Second
)

// canonicalSigningString builds the bytes that the client signs and the server
// verifies. Binding method + path + timestamp + body-hash prevents replaying a
// captured signature on a different endpoint, at a different time, or (for
// empty-body GETs) at all once the freshness window passes.
func canonicalSigningString(method, path, timestamp string, body []byte) []byte {
	bodyHash := sha256.Sum256(body)
	return []byte(method + "\n" + path + "\n" + timestamp + "\n" + hex.EncodeToString(bodyHash[:]))
}

// resolveDeviceKeys returns the public keys for did from the Redis cache; on cache miss
// it falls back to the DIDRegistry and populates the cache.
func (rt *Router) resolveDeviceKeys(ctx context.Context, did string) ([]*ecdsa.PublicKey, error) {
	if rt.Redis != nil {
		cached, err := rt.Redis.GetDIDDeviceKeys(ctx, did)
		if err == nil && len(cached) > 0 {
			return hexSliceToPublicKeys(cached)
		}
	}

	if rt.DIDRegistry == nil {
		return nil, fmt.Errorf("no DID registry configured")
	}
	bindings, err := rt.DIDRegistry.ListDevices(ctx, did)
	if err != nil {
		return nil, err
	}

	var hexKeys []string
	var pubs []*ecdsa.PublicKey
	for _, b := range bindings {
		pub, err := hexToPublicKey(b.PublicKeyHex)
		if err != nil {
			continue
		}
		pubs = append(pubs, pub)
		hexKeys = append(hexKeys, b.PublicKeyHex)
	}
	if len(pubs) == 0 {
		return nil, fmt.Errorf("no valid keys for DID")
	}

	if rt.Redis != nil {
		_ = rt.Redis.SetDIDDeviceKeys(ctx, did, hexKeys)
	}
	return pubs, nil
}

// verifyPasskeyAuth validates the X-Sender-DID / X-Signature / X-Timestamp
// headers against the request. The signature must cover the canonical string
// (method, path, timestamp, body-hash), the timestamp must be fresh (±120s),
// and the exact signature must not have been seen before within the window.
//
// The function peeks the body without consuming it so downstream handlers still
// receive the full body via r.Body.
func (rt *Router) verifyPasskeyAuth(r *http.Request, keys []*ecdsa.PublicKey) error {
	sigB64 := r.Header.Get(headerSignature)
	if sigB64 == "" {
		return passkeyError("AUTH_MISSING_SIGNATURE", "X-Signature header required")
	}

	ts := r.Header.Get(headerTimestamp)
	if ts == "" {
		return passkeyError("AUTH_MISSING_TIMESTAMP", "X-Timestamp header required")
	}
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return passkeyError("AUTH_STALE_TIMESTAMP", "X-Timestamp must be unix seconds")
	}
	if skew := time.Now().Unix() - tsInt; skew > maxClockSkewSeconds || skew < -maxClockSkewSeconds {
		return passkeyError("AUTH_STALE_TIMESTAMP", "request timestamp outside the allowed window")
	}

	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		// Try URL-safe encoding as fallback (iOS may use either)
		sig, err = base64.URLEncoding.DecodeString(sigB64)
		if err != nil {
			return passkeyError("AUTH_INVALID_SIGNATURE", "X-Signature is not valid base64")
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return passkeyError("AUTH_INVALID_SIGNATURE", "could not read request body")
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	msg := canonicalSigningString(r.Method, r.URL.Path, ts, body)

	verified := false
	for _, pub := range keys {
		if err := didkey.VerifyECDSAP256SHA256(pub, msg, sig); err == nil {
			verified = true
			break
		}
	}
	if !verified {
		return passkeyError("AUTH_INVALID_SIGNATURE", "signature does not match any registered device key")
	}

	// Replay protection: a given signature may be presented at most once within
	// the window. Combined with the freshness check, this closes the capture-and-
	// replay hole (including for empty-body GETs).
	if rt.tokenService != nil {
		sigHash := sha256.Sum256(sig)
		replayKey := "pkreplay:" + hex.EncodeToString(sigHash[:])
		if !rt.tokenService.CheckAndStoreNonceTTL(replayKey, replayWindow) {
			return passkeyError("AUTH_REPLAY", "request signature already used")
		}
	}
	return nil
}

// passkeyAuthError is a sentinel error type carrying the API error code.
type passkeyAuthError struct {
	code string
	msg  string
}

func (e *passkeyAuthError) Error() string { return e.code + ": " + e.msg }

func passkeyError(code, msg string) *passkeyAuthError {
	return &passkeyAuthError{code: code, msg: msg}
}

// hexToPublicKey parses a hex-encoded uncompressed or compressed P-256 public key.
func hexToPublicKey(hexStr string) (*ecdsa.PublicKey, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	curve := elliptic.P256()
	switch len(b) {
	case 65: // uncompressed: 0x04 || X || Y
		if b[0] != 0x04 {
			return nil, fmt.Errorf("unexpected prefix byte 0x%02x", b[0])
		}
		x := new(big.Int).SetBytes(b[1:33])
		y := new(big.Int).SetBytes(b[33:65])
		pub := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
		if !curve.IsOnCurve(x, y) {
			return nil, fmt.Errorf("point not on P-256 curve")
		}
		return pub, nil
	case 33: // compressed: 0x02 or 0x03 || X
		x, y := elliptic.UnmarshalCompressed(curve, b)
		if x == nil {
			return nil, fmt.Errorf("could not decompress P-256 point")
		}
		return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
	default:
		return nil, fmt.Errorf("unexpected key length %d", len(b))
	}
}

func hexSliceToPublicKeys(hexKeys []string) ([]*ecdsa.PublicKey, error) {
	out := make([]*ecdsa.PublicKey, 0, len(hexKeys))
	for _, h := range hexKeys {
		pub, err := hexToPublicKey(h)
		if err != nil {
			continue
		}
		out = append(out, pub)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid public keys in cache entry")
	}
	return out, nil
}
