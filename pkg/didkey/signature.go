package didkey

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

// ErrInvalidSignature is returned when ECDSA verification fails.
var ErrInvalidSignature = errors.New("didkey: invalid signature")

// VerifyECDSAP256SHA256 checks an ECDSA P-256 signature over message using SHA-256.
// The signature may be IEEE P1363 raw form (64 bytes, R||S) or ASN.1 DER
// (as produced by crypto/ecdsa.SignASN1).
func VerifyECDSAP256SHA256(pub *ecdsa.PublicKey, message, sig []byte) error {
	if pub == nil {
		return fmt.Errorf("%w: nil public key", ErrInvalidPublicKey)
	}
	digest := sha256.Sum256(message)
	if verifyFromDigest(pub, digest[:], sig) {
		return nil
	}
	return ErrInvalidSignature
}

func verifyFromDigest(pub *ecdsa.PublicKey, digest, sig []byte) bool {
	if len(sig) == 64 {
		r := new(big.Int).SetBytes(sig[:32])
		s := new(big.Int).SetBytes(sig[32:64])
		return ecdsa.Verify(pub, digest, r, s)
	}
	return ecdsa.VerifyASN1(pub, digest, sig)
}
