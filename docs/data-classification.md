# T0–T7 Data Classification Guide

> **WO-217** — CI enforcement of the ECHO data classification policy.  
> Violations at tier T0–T3 in on-chain submissions fail the `t0-t7-classification` CI job.

---

## Why This Matters

**Zero PII on any blockchain is a hard system invariant.**  
The Constellation Identity Metagraph and Data L1 validators reject any submission containing personal identifiers, message content, or user behavioral data beyond the allowances for tiers T5–T7. The Semgrep rules in `.semgrep/t0_t7_rules.yaml` enforce this at CI time so violations are caught before code reaches the chain.

---

## The T0–T7 Table

| Tier | Classification | Examples | On-Chain | Backend DB | IPFS/Storj | Device-Local |
|---|---|---|---|---|---|---|
| **T0** | Secret — never persisted | Message plaintext, private keys, decrypted content | ❌ | ❌ | ❌ | Memory only |
| **T1** | Device-local secret | Derived keys (HKDF output), biometric templates | ❌ | ❌ | ❌ | Secure Enclave only |
| **T2** | Encrypted local | Message ciphertext, SwiftData encrypted rows, Passport credential sync blobs (client AES-256-GCM) | ❌ | Relay only (opaque BYTEA) | Encrypted CID | AES-256-GCM at rest |
| **T3** | Relay-transient | Offline queue blobs, NATS messages | ❌ | Ephemeral TTL | ❌ | ❌ |
| **T4** | Encrypted audit | Operational logs (no content, no DID linkage) | CID only | ❌ | Encrypted | ❌ |
| **T5** | Hash commitment | Merkle roots of message batches (SHA-256) | ✅ | ❌ | — | — |
| **T6** | Trust commitment | `H(trust_score‖nonce)` (SHA-256) | ✅ | ❌ | — | — |
| **T7** | Public chain data | Token transactions, DID documents, governance proposals | ✅ | Cache only | ❌ | ❌ |

---

## What Goes On-Chain

Only **T5**, **T6**, and **T7** data is permitted in metagraph submissions.

### T5 — Merkle root (hash commitment)
```go
// ✅ Correct: submit the 32-byte SHA-256 hash of a message batch
sub := validation.DataL1Submission{
    MerkleRoot:      sha256.Sum256(batchBytes)[:],  // T5: hash only, never the batch
    CommitmentCount: len(batch),
    TimeRange:       timeRange,
    SchemaVersion:   3,
}
```

### T6 — Trust commitment
```go
// ✅ Correct: H(tier || nonce) — never the raw tier value
commitment := metagraph.TrustTierCommitmentHex(tier, nonce)  // returns hex(SHA-256)
```

### T7 — DID document / public key registration
```go
// ✅ Correct: public identity data for the DID document
tx := metagraph.DeviceKeyRegistrationUpdate{
    SubjectDID:   "did:key:z...",     // T7: public DID
    PublicKeyHex: "04abcd...",        // T7: public key (P-256 uncompressed)
    DeviceLabel:  "iPhone 16 Pro",    // T7: device label
    AddedAt:      time.Now().UnixMilli(),
}
```

---

## What MUST NOT Go On-Chain

```go
// ❌ WRONG: T0 plaintext message content
sub := DataL1Submission{SomeField: "Hello Alice, meeting at 3pm"}

// ❌ WRONG: T1 derived key material
sub := DataL1Submission{SomeField: string(derivedKey)}

// ❌ WRONG: T0 PII — email address in data payload
sub := DataL1Submission{UserEmail: "alice@example.com"}

// ❌ WRONG: DID as a data-payload field (belongs in the sender envelope)
sub := DataL1Submission{SenderDID: "did:key:z...", RecipientDID: "did:key:z..."}
```

---

## Code Annotations

Add a brief T0–T7 comment to any function that builds on-chain submissions or touches sensitive data. The Go doc comment convention:

```go
// DataL1Submission accepts T5 (Merkle root hash) data only.
// Strip all T0–T4 fields and PII before constructing this type.
// @DataTier: T5
type DataL1Submission struct {
    MerkleRoot      []byte    // T5: SHA-256 commitment of message batch
    CommitmentCount int
    TimeRange       TimeRange
    SchemaVersion   int
}
```

```swift
// @DataTier: T1 — Storage key is derived in memory on unlock, never persisted.
func deriveStorageKey(keyId: String) -> SymmetricKey
```

---

## CI Enforcement

The `t0-t7-classification` job in `.github/workflows/go-ci.yml` runs Semgrep with `.semgrep/t0_t7_rules.yaml` on every push. **ERROR-severity** rules block merge; **WARNING-severity** rules are advisory.

### Running locally

```bash
# Install Semgrep
pip install semgrep

# Run all T0–T7 rules
semgrep --config .semgrep/t0_t7_rules.yaml --error --severity ERROR .

# Run with advisory warnings
semgrep --config .semgrep/t0_t7_rules.yaml .
```

### Adding new rules

Edit `.semgrep/t0_t7_rules.yaml`. Each rule should:
1. Include `metadata.tier` (e.g. `T0`, `T1`)
2. Reference this document in `message`
3. Target the narrowest pattern possible to avoid false positives

---

## Metagraph Validator Enforcement (Scala)

The Data L1 Scala validator in `metagraph/modules/shared_data/src/main/scala/com/echo/shared_data/validations/Validations.scala` enforces:

- **`requireHex64`** — MerkleRoot and TrustCommitment fields must be exactly 64 lowercase hex characters (pure SHA-256 hash).
- **`rejectPII`** — Rejects email addresses, phone numbers, and embedded DID strings in any free-form field.
- **`rejectUpdatePII`** — Dispatches PII rejection across all `EchoUpdate` variants before structural validation.

These checks run at L1 consensus time. Any transaction that passes CI but fails these checks will be rejected by the metagraph.

---

## Quick Reference: Tier by Field

| Field / Variable Name | Tier | Notes |
|---|---|---|
| `message.plaintext` | T0 | Memory only — never pass to any function that logs or persists |
| `privateKey`, `signingKey` | T0 | Secure Enclave only on iOS; never in Go code |
| `derivedKey`, `storageKey` | T1 | Re-derive on unlock; zero on background |
| `encryptedContent`, `ciphertext` | T2 | Device-local AES-256-GCM only |
| `passport_sync_blob.ciphertext`, `ciphertext_base64` (Passport sync) | T2 | Client-encrypted with `CredentialSyncKey`; server/IPFS store opaque bytes only |
| `offlineQueueBlob` | T3 | Postgres ephemeral TTL; never on-chain |
| `auditCID` | T4 | Content-addressed CID to IPFS encrypted blob |
| `merkleRoot` | T5 | SHA-256 of batch; only hash, never content |
| `trustCommitment` | T6 | H(score‖nonce); only hash, never raw score |
| `did`, `publicKey`, `txHash` | T7 | Public chain data; OK on-chain and in cache |

---

*Last updated: 2026-05-09 — WO-217*
