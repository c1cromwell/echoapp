// Package crypto — Post-quantum hybrid KEM (WO-SX2).
//
// This adds a hybrid key-agreement that combines classical P-256 ECDH with
// ML-KEM-768 (FIPS 203, crypto/mlkem) so the agreed secret is safe as long as
// EITHER primitive holds. It is "hybrid, never PQ-only" by construction — the
// combiner mixes both shared secrets, so a break of ML-KEM alone (or of P-256
// alone) does not reveal the result.
//
// Where it plugs in: the Double Ratchet (ratchet.go) bootstraps from a 32-byte
// root secret produced by an X3DH-style handshake. Establishing THAT secret is
// the highest-value place to be post-quantum, because a "harvest now, decrypt
// later" adversary records the handshake today to break it once a quantum
// computer exists. NewRatchetInitiatorPQ / NewRatchetResponderPQ derive the
// ratchet's root secret through this hybrid KEM. Per-ratchet-step PQ (re-running
// a KEM on every DH step) is a further extension on top of this primitive.
package crypto

import (
	"crypto/ecdh"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// pqHybridInfo domain-separates the hybrid combiner KDF.
const pqHybridInfo = "ECHO-PQ-HYBRID-KEM-v1"

// HybridPrivateKey holds both secret keys for the hybrid KEM. Keep device-local;
// never serialize the secret halves off-device in plaintext.
type HybridPrivateKey struct {
	ec *ecdh.PrivateKey
	pq *mlkem.DecapsulationKey768
}

// HybridPublicBundle is the published key material a peer encapsulates against.
type HybridPublicBundle struct {
	EC string `json:"ec"` // base64 P-256 raw public key (X||Y, 64B)
	PQ string `json:"pq"` // base64 ML-KEM-768 encapsulation key (1184B)
}

// HybridCiphertext is what the encapsulating party sends to the decapsulating party.
type HybridCiphertext struct {
	EphemeralEC string `json:"eec"` // base64 ephemeral P-256 raw public key (64B)
	PQ          string `json:"pq"`  // base64 ML-KEM-768 ciphertext (1088B)
}

// GenerateHybridKeyPair creates a new hybrid key pair and its public bundle.
func GenerateHybridKeyPair() (*HybridPrivateKey, *HybridPublicBundle, error) {
	ecPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: ec keygen: %w", err)
	}
	pqPriv, err := mlkem.GenerateKey768()
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: ml-kem keygen: %w", err)
	}
	priv := &HybridPrivateKey{ec: ecPriv, pq: pqPriv}
	bundle := &HybridPublicBundle{
		EC: base64.StdEncoding.EncodeToString(rawP256Pub(ecPriv)),
		PQ: base64.StdEncoding.EncodeToString(pqPriv.EncapsulationKey().Bytes()),
	}
	return priv, bundle, nil
}

// HybridEncapsulate derives a shared secret against a peer's public bundle and
// returns the ciphertext the peer needs to derive the same secret.
func HybridEncapsulate(remote *HybridPublicBundle) (*HybridCiphertext, []byte, error) {
	remoteECRaw, err := base64.StdEncoding.DecodeString(remote.EC)
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: decode ec: %w", err)
	}
	remotePQRaw, err := base64.StdEncoding.DecodeString(remote.PQ)
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: decode pq: %w", err)
	}

	// Classical half: ephemeral ECDH.
	ephemeral, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: ephemeral keygen: %w", err)
	}
	ssEC, err := ratchetKS.dhRaw(ephemeral, remoteECRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: ec ecdh: %w", err)
	}

	// PQ half: ML-KEM encapsulation.
	ek, err := mlkem.NewEncapsulationKey768(remotePQRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("pqhybrid: parse ml-kem ek: %w", err)
	}
	ssPQ, ctPQ := ek.Encapsulate()

	ct := &HybridCiphertext{
		EphemeralEC: base64.StdEncoding.EncodeToString(rawP256Pub(ephemeral)),
		PQ:          base64.StdEncoding.EncodeToString(ctPQ),
	}
	secret, err := hybridCombine(ssEC, ssPQ, rawP256Pub(ephemeral), ctPQ)
	if err != nil {
		return nil, nil, err
	}
	return ct, secret, nil
}

// HybridDecapsulate derives the same 32-byte secret from our private key and the
// peer's ciphertext.
func HybridDecapsulate(priv *HybridPrivateKey, ct *HybridCiphertext) ([]byte, error) {
	ephemeralECRaw, err := base64.StdEncoding.DecodeString(ct.EphemeralEC)
	if err != nil {
		return nil, fmt.Errorf("pqhybrid: decode ephemeral ec: %w", err)
	}
	ctPQ, err := base64.StdEncoding.DecodeString(ct.PQ)
	if err != nil {
		return nil, fmt.Errorf("pqhybrid: decode pq ct: %w", err)
	}

	ssEC, err := ratchetKS.dhRaw(priv.ec, ephemeralECRaw)
	if err != nil {
		return nil, fmt.Errorf("pqhybrid: ec ecdh: %w", err)
	}
	ssPQ, err := priv.pq.Decapsulate(ctPQ)
	if err != nil {
		return nil, fmt.Errorf("pqhybrid: ml-kem decapsulate: %w", err)
	}
	return hybridCombine(ssEC, ssPQ, ephemeralECRaw, ctPQ)
}

// hybridCombine mixes the two shared secrets with a KDF, binding the transcript
// (ephemeral EC pubkey + PQ ciphertext) so the result is tied to this exchange.
// Concatenation-into-KDF is a standard hybrid combiner: the output is secure if
// either ssEC or ssPQ is secret.
func hybridCombine(ssEC, ssPQ, ephemeralEC, ctPQ []byte) ([]byte, error) {
	ikm := make([]byte, 0, len(ssEC)+len(ssPQ))
	ikm = append(ikm, ssEC...)
	ikm = append(ikm, ssPQ...)

	transcript := sha256.New()
	transcript.Write(ephemeralEC)
	transcript.Write(ctPQ)
	info := append([]byte(pqHybridInfo), transcript.Sum(nil)...)

	r := hkdf.New(sha256.New, ikm, nil, info)
	out := make([]byte, 32)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, fmt.Errorf("pqhybrid: combiner kdf: %w", err)
	}
	return out, nil
}

// NewRatchetInitiatorPQ bootstraps an initiator (Alice) ratchet session using a
// PQ-hybrid handshake. It returns the session plus the HybridCiphertext the
// responder needs. remoteRatchetPub is the responder's published ratchet pre-key.
func NewRatchetInitiatorPQ(remote *HybridPublicBundle, remoteRatchetPub []byte) (*RatchetSession, *HybridCiphertext, error) {
	ct, secret, err := HybridEncapsulate(remote)
	if err != nil {
		return nil, nil, err
	}
	sess, err := NewRatchetSessionInitiator(secret, remoteRatchetPub)
	if err != nil {
		return nil, nil, err
	}
	return sess, ct, nil
}

// NewRatchetResponderPQ bootstraps a responder (Bob) ratchet session by
// decapsulating the initiator's HybridCiphertext. selfRatchet is the ratchet
// key pair whose public half Bob published as the pre-key.
func NewRatchetResponderPQ(priv *HybridPrivateKey, ct *HybridCiphertext, selfRatchet *ecdh.PrivateKey) (*RatchetSession, error) {
	secret, err := HybridDecapsulate(priv, ct)
	if err != nil {
		return nil, err
	}
	return NewRatchetSessionResponder(secret, selfRatchet)
}

// MarshalBundle / parsing helpers for transport.

// Marshal serializes the public bundle as JSON.
func (b *HybridPublicBundle) Marshal() ([]byte, error) { return json.Marshal(b) }

// ParseHybridPublicBundle parses a JSON-encoded public bundle.
func ParseHybridPublicBundle(data []byte) (*HybridPublicBundle, error) {
	var b HybridPublicBundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("pqhybrid: parse bundle: %w", err)
	}
	return &b, nil
}

// rawP256Pub returns the 64-byte X||Y form of a P-256 public key.
func rawP256Pub(k *ecdh.PrivateKey) []byte {
	pub := k.PublicKey().Bytes()
	if len(pub) == 65 && pub[0] == 0x04 {
		return pub[1:]
	}
	return pub
}
