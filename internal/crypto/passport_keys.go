package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	// PassportCredentialSyncInfo is the HKDF info string for credential sync keys (WO-294).
	PassportCredentialSyncInfo = "echo-passport-credential-sync"
)

// DerivePassportCredentialSyncKey derives the AES-256-GCM key used to encrypt
// credential wallet sync blobs on the client.
// PassportRootKey lives in Secure Enclave + recovery material (T1); this output is T1.
func (kds *KeyDerivationService) DerivePassportCredentialSyncKey(passportRootKey []byte) ([]byte, error) {
	if len(passportRootKey) < 32 {
		return nil, fmt.Errorf("passport root key must be at least 32 bytes")
	}
	// Fixed salt for deterministic derivation from root (per-device root is already unique).
	salt := []byte("echo-passport-sync-v1")
	hkdfReader := hkdf.New(sha256.New, passportRootKey, salt, []byte(PassportCredentialSyncInfo))
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("failed to derive credential sync key: %w", err)
	}
	return key, nil
}
