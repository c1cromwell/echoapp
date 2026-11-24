# T0–T7 Data Classification Guide

> **WO-217** — CI enforcement of the ECHO data classification policy.  
> Violations at tier T0–T3 in on-chain submissions fail the `t0-t7-classification` CI job.

---

## Why This Matters

**Zero PII on any blockchain is a hard system invariant.**  
Echo's chain (a private Hyperledger Besu network, [`BESU_MIGRATION_PLAN.md`](BESU_MIGRATION_PLAN.md)) accepts 32-byte hashes and small numbers only: the contracts have no field that could hold a personal identifier, message content or behavioural data. The Semgrep rules in `.semgrep/t0_t7_rules.yaml` add a check at CI time, before code reaches the chain.

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

Only **T5**, **T6**, and **T7** data is permitted on-chain, and since the move to Besu
(2026-11-24) even T7 identifiers are written as hashes, never as text.

| Record | What the chain stores | What stays off-chain |
|---|---|---|
| Message-batch Merkle root (T5) | The root and a leaf count | The messages and their commitments |
| Trust commitment (T6) | `H(tier \|\| nonce)` | The tier and the nonce |
| Device key | Hash of the DID, hash of the public key | The device label (free text a user chose, e.g. "Alice's iPhone") |
| Username | Hash of the lower-cased username, hash of the DID | — |
| Credential issuance | Hashes of the credential id, subject DID, issuer DID and type | The credential |
| Revocation list | Hash of the issuer DID, a sequence number, hash of the list | The list (served by the gateway, checked against the hash) |
| Group metadata, governance votes | One digest of the whole record | The record, including the admin's or voter's DID |

Identifiers are hashed as `SHA-256("echo/v1/" + domain + ":" + value)` (`internal/besu/ids.go`).

**What a hash does not hide.** A DID is a long random key, so its hash tells nothing to someone
who does not already hold the DID. A username is short and guessable: anyone can hash candidate
names and find which are registered. Username ownership is meant to be checkable, so this is
accepted — but a username must not be treated as private. The same applies to a vote digest: a
vote by a known DID on a known proposal can be confirmed by guessing it.

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

### T7 — public key registration
```go
// ✅ Correct: describe the record; the chain client hashes the DID and the key.
tx := metagraph.DeviceKeyRegistrationUpdate{
    SubjectDID:   "did:key:z...",     // stored on-chain as a hash
    PublicKeyHex: "04abcd...",        // stored on-chain as a hash
    DeviceLabel:  "iPhone 16 Pro",    // NOT sent to the chain
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

## Enforcement on the chain

The contracts in `chain/contracts/` are the last line of defence, and they enforce the rule by
construction, not by inspection:

- Every value they accept is a `bytes32` or an integer. Text cannot be submitted, so there is
  nothing to scan for emails, phone numbers or DIDs.
- Only the gateway's key may write (`NotAuthorisedSigner` otherwise).
- The Go client refuses anything that is not exactly 64 hex characters before it reaches the
  chain (`internal/besu`, `parseHash32`), and refuses update types that have no on-chain form.

`make validate-phase1` step 3 checks a live block for a copy of the DID it just anchored and fails
if it finds one.

---

## Quick Reference: Tier by Field

| Field / Variable Name | Tier | Notes |
|---|---|---|
| `message.plaintext` | T0 | Memory only — never pass to any function that logs or persists |
| `privateKey`, `signingKey` | T0 | Secure Enclave only on iOS; never in Go code |
| `derivedKey`, `storageKey` | T1 | Re-derive on unlock; zero on background |
| `encryptedContent`, `ciphertext` | T2 | Device-local AES-256-GCM only |
| `passport_sync_blob.ciphertext`, `ciphertext_base64` (Passport sync) | T2 | Client-encrypted with `CredentialSyncKey`; server/IPFS store opaque bytes only |
| `passport_recovery_share`, `passport_recovery_policy` (guardian DIDs, share index) | T7 | Metadata only — no Shamir share bytes or recovery secret server-side |
| `offlineQueueBlob` | T3 | Postgres ephemeral TTL; never on-chain |
| `auditCID` | T4 | Content-addressed CID to IPFS encrypted blob |
| `merkleRoot` | T5 | SHA-256 of batch; only hash, never content |
| `trustCommitment` | T6 | H(score‖nonce); only hash, never raw score |
| `did`, `publicKey`, `txHash` | T7 | Public chain data; OK on-chain and in cache |

---

*Last updated: 2026-05-09 — WO-217*
