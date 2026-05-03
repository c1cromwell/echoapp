package didkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ParseECPrivateKeyPEM decodes a PEM block containing either a PKCS #8 or SEC1
// EC private key and returns the P-256 private key.
func ParseECPrivateKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("%w: no PEM block", ErrInvalidPublicKey)
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ec, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("%w: PKCS#8 is not ECDSA", ErrInvalidPublicKey)
		}
		if ec.Curve != elliptic.P256() {
			return nil, fmt.Errorf("%w: expected P-256 private key", ErrInvalidPublicKey)
		}
		return ec, nil
	}
	ec, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPublicKey, err)
	}
	if ec.Curve != elliptic.P256() {
		return nil, fmt.Errorf("%w: expected P-256 private key", ErrInvalidPublicKey)
	}
	return ec, nil
}

// SignECDSAP256SHA256ASN1 returns ECDSA P-256 / SHA-256 signature in ASN.1 DER form
// over canonicalMessage (the message is hashed with SHA-256 before signing).
func SignECDSAP256SHA256ASN1(priv *ecdsa.PrivateKey, canonicalMessage []byte) ([]byte, error) {
	if priv == nil {
		return nil, fmt.Errorf("%w: nil private key", ErrInvalidPublicKey)
	}
	sum := sha256.Sum256(canonicalMessage)
	return ecdsa.SignASN1(rand.Reader, priv, sum[:])
}
