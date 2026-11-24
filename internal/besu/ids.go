package besu

import (
	"crypto/sha256"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// On-chain identifiers.
//
// Nothing readable is written to Echo's chain. Wherever a record refers to a DID, a username, a
// device key or a credential, the chain holds
//
//	SHA-256("echo/v1/" + domain + ":" + value)
//
// The domain keeps hashes of different kinds of value apart (a username can never collide with a
// DID). Anyone who knows the value can recompute the hash and look the record up; that is the
// point of anchoring it. Short, guessable values (usernames) are therefore findable by trial, and
// must not be treated as hidden.
const idPrefix = "echo/v1/"

// Domains of on-chain identifiers.
const (
	domainDID            = "did"
	domainUsername       = "username"
	domainDeviceKey      = "device-key"
	domainCredential     = "credential"
	domainCredentialKind = "credential-kind"
	domainAttestation    = "attestation"
)

func id(domain, value string) common.Hash {
	return sha256.Sum256([]byte(idPrefix + domain + ":" + value))
}

// SubjectID is the on-chain identifier of a DID.
func SubjectID(did string) common.Hash { return id(domainDID, did) }

// UsernameID is the on-chain identifier of a username. Usernames are case-insensitive.
func UsernameID(username string) common.Hash {
	return id(domainUsername, strings.ToLower(strings.TrimSpace(username)))
}

// DeviceKeyID is the on-chain identifier of a device public key given as hex.
func DeviceKeyID(publicKeyHex string) common.Hash {
	return id(domainDeviceKey, strings.ToLower(strings.TrimPrefix(strings.TrimSpace(publicKeyHex), "0x")))
}

// CredentialID is the on-chain identifier of a credential id (for example "urn:uuid:…").
func CredentialID(credentialID string) common.Hash { return id(domainCredential, credentialID) }

// CredentialKindID is the on-chain identifier of a credential type name.
func CredentialKindID(kind string) common.Hash { return id(domainCredentialKind, kind) }

// AttestationKindID is the on-chain identifier of an attestation kind (for example "group_metadata").
func AttestationKindID(kind string) common.Hash { return id(domainAttestation, kind) }

// ListHash is the hash under which a revocation list is published: SHA-256 of its bytes.
func ListHash(list []byte) common.Hash { return sha256.Sum256(list) }
