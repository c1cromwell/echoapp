---
name: echo-t0-t7-review
description: >-
  Reviews Echo code for T0–T7 data classification violations before chain
  submissions, logging, or persistence. Use when editing Go/Scala metagraph code,
  log publishers, Data L1 handlers, identity L1 updates, or iOS chain-adjacent
  features.
---

# Echo T0–T7 review

## Invariant

**Zero PII on any blockchain.** Only **T5** (Merkle roots), **T6** (trust commitments), **T7** (public chain data) belong in metagraph submissions.

Full table: [`docs/data-classification.md`](../../docs/data-classification.md)

## Quick tier reference

| Tier | Never in logs/DB/chain |
|------|------------------------|
| T0 | Plaintext messages, private keys, decrypted content |
| T1 | HKDF outputs, Secure Enclave key bytes |
| T2 | Encrypted local ciphertext (device only) |
| T3 | Relay queue blobs (ephemeral) |
| T4 | Audit logs — encrypted off-chain; CID only on chain |
| T5–T7 | Allowed on-chain **only** as hashes/commitments/public metadata |

## Pre-PR checklist (Go)

- [ ] No plaintext message content in logs, HTTP responses, or DB columns  
- [ ] No `SenderDID` / `RecipientDID` inside Data L1 **payload** (sender is envelope)  
- [ ] Merkle roots are **32-byte** SHA-256, not raw batch bytes  
- [ ] Trust tier on-chain is `H(tier||nonce)`, not raw tier integer  
- [ ] Operational logs (`internal/logging/`) — no DIDs, phones, emails in events  
- [ ] New env vars / config don't embed secrets in repo  

## Pre-PR checklist (Scala / metagraph)

- [ ] `IdentityValidations` / Data L1 validators reject non-did:key where required  
- [ ] Trust commitment is 64-char hex (32 bytes)  
- [ ] StatusList2021 bit vectors match expected length (131072 bits)  
- [ ] No PII fields added to `IdentityUpdate` case classes  

## Pre-PR checklist (iOS)

- [ ] Private keys never leave Secure Enclave except signatures  
- [ ] `EchoLogger` / privacy scrubbers used before logging user content  
- [ ] SwiftData stores encrypted payloads (T2), not plaintext (T0)  

## Run CI rules locally

```bash
# Go ERROR rules (must pass)
pip install semgrep
semgrep --config .semgrep/t0_t7_rules.yaml --error --severity ERROR \
  --include="*.go" --exclude-dir=vendor .

# Swift (iOS CI runs advisory WARNINGs)
semgrep --config .semgrep/t0_t7_rules.yaml --severity ERROR \
  --include="*.swift" ios/Echo/Sources/
```

Same job as `.github/workflows/go-ci.yml` → `t0-t7-classification`.

## Common violations to catch in review

```go
// ❌ T0 in log
log.Printf("message body: %s", plaintext)

// ❌ PII in L1 submission struct
type Bad struct { Email string `json:"email"` }

// ❌ Raw tier on chain
TrustTierUpdate{ Tier: 4 }  // use commitment hash instead

// ✅ T5
MerkleRoot: sha256.Sum256(batch)[:]
```

## When to escalate

- New on-chain schema fields → confirm tier with `docs/data-classification.md`  
- Cross-service payloads (NATS, IPFS) → T3/T4 rules  
- User asks to "store message for relay" → T3 ephemeral only, never Postgres plaintext  

## Related WOs

- WO-217 — CI Semgrep enforcement  
- WO-35 — server-side pre-validation  
- WO-53 / WO-6 — privacy-safe logging  
