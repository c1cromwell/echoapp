// Package crypto — Double Ratchet (WO-SX1).
//
// This adds forward secrecy and post-compromise ("break-in") recovery on top of
// the Kinnami primitives (P-256 ECDH + HKDF-SHA256 + AES-256-GCM). It is a
// Signal-style Double Ratchet:
//
//   - A DH ratchet: each party advances a P-256 ratchet key; every time a message
//     arrives carrying a new remote ratchet public key, both root chain and the
//     receiving chain are re-keyed via a fresh ECDH. This is what gives
//     post-compromise recovery — a leaked chain key is healed by the next DH step.
//   - A symmetric-key ratchet: within a chain, each message advances a chain key
//     (KDF_CK) and derives a one-time message key (KDF_MK). Past message keys are
//     not recoverable from a later chain key — that is the forward secrecy.
//
// Out-of-order and dropped messages are handled by deriving and caching skipped
// message keys (bounded by maxSkip) keyed by (ratchetPub, messageNumber).
//
// The shared-secret bootstrap (the initial root key) is expected to come from an
// X3DH-style agreement over the parties' long-term did:key identity keys; this
// module takes that 32-byte secret as input (see NewRatchetSession*). The on-wire
// header is intentionally compatible in spirit with KinnamiEncryptedMessageWithKey
// (it carries a ratchet public key + GCM fields), so the iOS CryptoKit port can
// mirror it the same way kinnami.go mirrors KinnamiEncryption.swift.
package crypto

import (
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	// ratchetRootInfo / ratchetChainInfo separate the two KDF domains so a root
	// secret and a chain key can never collide.
	ratchetRootInfo  = "ECHO-RATCHET-ROOT"
	ratchetChainInfo = "ECHO-RATCHET-CHAIN"

	// ratchetMessageKeyConst / ratchetChainKeyConst are the single-byte constants
	// for the symmetric-key ratchet (KDF_CK), à la the Signal spec: the chain key
	// advances with 0x02, the message key is derived with 0x01.
	ratchetMessageKeyConst = 0x01
	ratchetChainKeyConst   = 0x02

	// RatchetMaxSkip bounds how many message keys we will derive to catch up to an
	// out-of-order message, preventing a malicious header from forcing unbounded work.
	RatchetMaxSkip = 1000

	// RatchetAlgorithm tags the wire format.
	RatchetAlgorithm = "DR-P256-AES256GCM"
)

// RatchetMessage is the on-wire envelope for a Double Ratchet message.
type RatchetMessage struct {
	RatchetPublicKey string `json:"ratchetPublicKey"` // base64 P-256 raw pubkey (X||Y, 64B) of the sender's current ratchet key
	PreviousChainN   uint32 `json:"pn"`               // # messages in the sender's previous sending chain (for skipped-key handling)
	MessageN         uint32 `json:"n"`                // message index within the current sending chain
	Nonce            string `json:"nonce"`            // base64 12-byte GCM nonce
	Ciphertext       string `json:"ciphertext"`       // base64 ciphertext
	Tag              string `json:"tag"`              // base64 16-byte GCM tag
	Algorithm        string `json:"algorithm"`
}

// skippedKey identifies a cached message key for an out-of-order message.
type skippedKey struct {
	ratchetPub string // base64 ratchet pubkey
	n          uint32
}

// RatchetSession is one party's Double Ratchet state for a single conversation.
//
// It is NOT safe for concurrent use; callers serialize access per conversation
// (the relay already routes per-conversation). State is in-memory; persistence
// (encrypted at rest) is the caller's responsibility — see ExportState.
type RatchetSession struct {
	rootKey []byte // RK

	dhSelf   *ecdh.PrivateKey // our current ratchet key pair
	dhRemote []byte           // their current ratchet public key (raw 64B), nil until first received

	sendChainKey []byte // CKs
	recvChainKey []byte // CKr

	sendN     uint32 // Ns
	recvN     uint32 // Nr
	prevSendN uint32 // PN — messages in the previous sending chain
	skipped   map[skippedKey][]byte
	maxSkip   int
}

// ks is a stateless helper reused for ECDH/GCM mechanics.
var ratchetKS = NewKinnamiService()

// NewRatchetSessionInitiator creates the session for the party that sends first
// (Alice). It is seeded with the shared secret from the X3DH-style handshake and
// the responder's initial ratchet public key (Bob's signed pre-key, raw 64B).
func NewRatchetSessionInitiator(sharedSecret, remoteRatchetPub []byte) (*RatchetSession, error) {
	if len(sharedSecret) != 32 {
		return nil, fmt.Errorf("ratchet: shared secret must be 32 bytes, got %d", len(sharedSecret))
	}
	dhSelf, _, err := ratchetKS.GenerateKeyPairP256()
	if err != nil {
		return nil, err
	}
	s := &RatchetSession{
		rootKey:  append([]byte(nil), sharedSecret...),
		dhSelf:   dhSelf,
		dhRemote: append([]byte(nil), remoteRatchetPub...),
		skipped:  make(map[skippedKey][]byte),
		maxSkip:  RatchetMaxSkip,
	}
	// Perform the initial DH ratchet so Alice has a sending chain immediately.
	dh, err := ratchetKS.dhRaw(s.dhSelf, s.dhRemote)
	if err != nil {
		return nil, err
	}
	s.rootKey, s.sendChainKey, err = kdfRootKey(s.rootKey, dh)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// NewRatchetSessionResponder creates the session for the party that receives
// first (Bob). It is seeded with the same shared secret and the ratchet key pair
// whose public half was published in the handshake (Bob's signed pre-key). Bob
// has no sending chain until he performs his first DH ratchet on receipt.
func NewRatchetSessionResponder(sharedSecret []byte, selfRatchet *ecdh.PrivateKey) (*RatchetSession, error) {
	if len(sharedSecret) != 32 {
		return nil, fmt.Errorf("ratchet: shared secret must be 32 bytes, got %d", len(sharedSecret))
	}
	if selfRatchet == nil {
		return nil, errors.New("ratchet: responder requires its published ratchet key pair")
	}
	return &RatchetSession{
		rootKey: append([]byte(nil), sharedSecret...),
		dhSelf:  selfRatchet,
		skipped: make(map[skippedKey][]byte),
		maxSkip: RatchetMaxSkip,
	}, nil
}

// SelfRatchetPublic returns our current ratchet public key (raw 64B), e.g. for
// publishing the responder's pre-key.
func (s *RatchetSession) SelfRatchetPublic() []byte {
	pub := s.dhSelf.PublicKey().Bytes()
	if len(pub) == 65 && pub[0] == 0x04 {
		pub = pub[1:]
	}
	return pub
}

// Encrypt advances the sending chain and seals plaintext under a one-time message key.
func (s *RatchetSession) Encrypt(plaintext []byte) (*RatchetMessage, error) {
	if s.sendChainKey == nil {
		return nil, errors.New("ratchet: no sending chain (responder must receive a message first)")
	}
	ck, mk, err := kdfChainKey(s.sendChainKey)
	if err != nil {
		return nil, err
	}
	s.sendChainKey = ck

	enc, err := ratchetKS.Encrypt(plaintext, mk)
	if err != nil {
		return nil, err
	}
	msg := &RatchetMessage{
		RatchetPublicKey: base64.StdEncoding.EncodeToString(s.SelfRatchetPublic()),
		PreviousChainN:   s.prevSendN,
		MessageN:         s.sendN,
		Nonce:            enc.Nonce,
		Ciphertext:       enc.Ciphertext,
		Tag:              enc.Tag,
		Algorithm:        RatchetAlgorithm,
	}
	s.sendN++
	return msg, nil
}

// Decrypt opens a Double Ratchet message, performing a DH ratchet step if the
// message carries a new remote ratchet key, and handling out-of-order delivery
// via skipped message keys.
func (s *RatchetSession) Decrypt(msg *RatchetMessage) ([]byte, error) {
	remotePub, err := base64.StdEncoding.DecodeString(msg.RatchetPublicKey)
	if err != nil {
		return nil, fmt.Errorf("ratchet: bad ratchet public key: %w", err)
	}

	// 1. If we have a cached skipped key for this exact (ratchetPub, n), use it.
	if pt, ok, err := s.trySkipped(remotePub, msg); err != nil || ok {
		return pt, err
	}

	// 2. If the message advertises a new remote ratchet key, DH-ratchet forward.
	if !ratchetPubEqual(remotePub, s.dhRemote) {
		if err := s.skipMessageKeys(s.recvN, msg.PreviousChainN); err != nil {
			return nil, err
		}
		if err := s.dhRatchet(remotePub); err != nil {
			return nil, err
		}
	}

	// 3. Skip and cache any keys before this message's index in the current chain.
	if err := s.skipMessageKeys(s.recvN, msg.MessageN); err != nil {
		return nil, err
	}

	// 4. Derive this message's key and open.
	if s.recvChainKey == nil {
		return nil, errors.New("ratchet: no receiving chain")
	}
	ck, mk, err := kdfChainKey(s.recvChainKey)
	if err != nil {
		return nil, err
	}
	s.recvChainKey = ck
	s.recvN++

	return ratchetKS.Decrypt(&KinnamiEncryptedMessage{
		Nonce: msg.Nonce, Ciphertext: msg.Ciphertext, Tag: msg.Tag, Algorithm: KinnamiAlgorithm,
	}, mk)
}

// trySkipped looks for a cached out-of-order message key and, if found, opens with it.
func (s *RatchetSession) trySkipped(remotePub []byte, msg *RatchetMessage) ([]byte, bool, error) {
	key := skippedKey{ratchetPub: base64.StdEncoding.EncodeToString(remotePub), n: msg.MessageN}
	mk, ok := s.skipped[key]
	if !ok {
		return nil, false, nil
	}
	pt, err := ratchetKS.Decrypt(&KinnamiEncryptedMessage{
		Nonce: msg.Nonce, Ciphertext: msg.Ciphertext, Tag: msg.Tag, Algorithm: KinnamiAlgorithm,
	}, mk)
	if err != nil {
		return nil, false, err
	}
	delete(s.skipped, key)
	return pt, true, nil
}

// skipMessageKeys derives and caches message keys from index `from` up to `until`
// (exclusive) in the current receiving chain, for later out-of-order delivery.
func (s *RatchetSession) skipMessageKeys(from, until uint32) error {
	if s.recvChainKey == nil {
		return nil // no chain yet (e.g. very first message) — nothing to skip
	}
	if until < from {
		return nil
	}
	if until-from > uint32(s.maxSkip) {
		return fmt.Errorf("ratchet: too many skipped messages (%d > %d)", until-from, s.maxSkip)
	}
	remoteB64 := base64.StdEncoding.EncodeToString(s.dhRemote)
	for i := from; i < until; i++ {
		ck, mk, err := kdfChainKey(s.recvChainKey)
		if err != nil {
			return err
		}
		s.recvChainKey = ck
		s.skipped[skippedKey{ratchetPub: remoteB64, n: i}] = mk
	}
	s.recvN = until
	return nil
}

// dhRatchet performs a DH ratchet step: adopt the remote ratchet key, derive a
// new receiving chain, generate a fresh self ratchet key, and derive a new
// sending chain. This is the post-compromise recovery step.
func (s *RatchetSession) dhRatchet(remotePub []byte) error {
	s.prevSendN = s.sendN
	s.sendN = 0
	s.recvN = 0
	s.dhRemote = append([]byte(nil), remotePub...)

	// Receiving chain from current self key + new remote key.
	dh, err := ratchetKS.dhRaw(s.dhSelf, s.dhRemote)
	if err != nil {
		return err
	}
	s.rootKey, s.recvChainKey, err = kdfRootKey(s.rootKey, dh)
	if err != nil {
		return err
	}

	// New self ratchet key, then sending chain.
	newSelf, _, err := ratchetKS.GenerateKeyPairP256()
	if err != nil {
		return err
	}
	s.dhSelf = newSelf
	dh2, err := ratchetKS.dhRaw(s.dhSelf, s.dhRemote)
	if err != nil {
		return err
	}
	s.rootKey, s.sendChainKey, err = kdfRootKey(s.rootKey, dh2)
	return err
}

// --- KDFs -------------------------------------------------------------------

// kdfRootKey is KDF_RK: HKDF(rootKey, dhOutput) -> (newRootKey, chainKey).
func kdfRootKey(rootKey, dhOutput []byte) (newRoot, chainKey []byte, err error) {
	r := hkdf.New(sha256.New, dhOutput, rootKey, []byte(ratchetRootInfo))
	out := make([]byte, 64)
	if _, err = io.ReadFull(r, out); err != nil {
		return nil, nil, fmt.Errorf("ratchet: root KDF failed: %w", err)
	}
	return out[:32], out[32:], nil
}

// kdfChainKey is KDF_CK: advance the chain key and derive a message key using
// HMAC with single-byte constants (Signal spec).
func kdfChainKey(chainKey []byte) (nextChainKey, messageKey []byte, err error) {
	mac := hmac.New(sha256.New, chainKey)
	if _, err = mac.Write([]byte{ratchetMessageKeyConst}); err != nil {
		return nil, nil, err
	}
	messageKey = mac.Sum(nil)

	mac.Reset()
	if _, err = mac.Write([]byte{ratchetChainKeyConst}); err != nil {
		return nil, nil, err
	}
	nextChainKey = mac.Sum(nil)
	// messageKey is run through HKDF to domain-separate it from the chain.
	mkReader := hkdf.New(sha256.New, messageKey, nil, []byte(ratchetChainInfo))
	mk := make([]byte, 32)
	if _, err = io.ReadFull(mkReader, mk); err != nil {
		return nil, nil, err
	}
	return nextChainKey, mk, nil
}

// --- helpers ----------------------------------------------------------------

// dhRaw performs raw ECDH (no HKDF) between our private key and a raw remote
// public key, returning the shared point — the input to the root KDF.
func (ks *KinnamiService) dhRaw(ourPriv *ecdh.PrivateKey, theirPubRaw []byte) ([]byte, error) {
	var pubBytes []byte
	switch {
	case len(theirPubRaw) == 64:
		pubBytes = append([]byte{0x04}, theirPubRaw...)
	case len(theirPubRaw) == 65 && theirPubRaw[0] == 0x04:
		pubBytes = theirPubRaw
	default:
		return nil, fmt.Errorf("ratchet: invalid public key length %d", len(theirPubRaw))
	}
	theirPub, err := ecdh.P256().NewPublicKey(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("ratchet: parse public key: %w", err)
	}
	return ourPriv.ECDH(theirPub)
}

// ratchetPubEqual reports whether two raw ratchet public keys match. A nil
// (not-yet-known) remote key is treated as "not equal" so the first message
// triggers a DH ratchet.
func ratchetPubEqual(a, b []byte) bool {
	if a == nil || b == nil {
		return false
	}
	return hmac.Equal(a, b)
}
