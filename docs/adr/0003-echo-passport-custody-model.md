# ADR 0003 — Echo Passport Custody Model: hybrid, client-encrypted, holder-not-issuer

- **Status:** Accepted
- **Date:** 2026-05-29
- **Deciders:** Product / Platform / Identity
- **Related work order:** WO-292 … WO-308 (Echo Passport, 4th product)
- **Blocks / unblocks:** WO-293 (holder model), WO-294 (sync), WO-298 (metagraph anchors)
- **Supersedes:** N/A

## Context

The founder proposed a **fourth product**: a verifiable credential vault to hold identity,
payment, legal, and employment data and use it — via app, AI agent, or smart glasses — to access
services and transact in daily life, as part of the Network State vision (full plan:
`docs/ECHO_PASSPORT_PLAN.md`).

Taken literally — "store PII, card numbers, passports, deeds" — this is a centralized honeypot
that directly violates Echo's foundational invariant: **zero readable PII on any server or
chain**, the T0–T7 data classification enforced in CI (`docs/data-classification.md`, Semgrep).
It would also make Echo a regulated data controller of plaintext PII and a prime breach target.

At the same time, ~70% of the *correct* version already exists: W3C VC 2.0 issuance
(`pkg/credentials/`), StatusList2021 revocation (`pkg/credentials/statuslist_l1.go`), OIDC4VCI
presentation (`pkg/credentials/oidc4vc/`), `did:key` identity (`pkg/didkey/`), Identity Metagraph
anchoring (`metagraph/.../IdentityTypes.scala`), and the Tessellation v3 `AllowSpend` /
`SpendTransaction` consent-bound payment primitives.

The real architectural variable is not "custodial vs non-custodial" — it is **where the
encryption key lives and how recovery works**. The honeypot risk is about *plaintext access*, not
storage location: client-encrypted data the provider can never decrypt is not the same liability
as a readable PII database.

## Decision

**Echo Passport is a holder-side credential wallet + selective-disclosure + agentic-presentation
layer — not a PII database. Custody is hybrid with an opinionated default (split by data kind,
not by per-item user choice), and Echo never holds a decryption key.**

### Three layers

1. **Layer 1 — Credentials (signed VCs).** Client-encrypted and synced to IPFS/Storj so they are
   available on web / glasses / agent **without the phone in hand**. Echo stores ciphertext + CID
   only. This is what you actually present 99% of the time. Classified **T2** (the
   client-encrypted-sync case is added to `data-classification.md`).
2. **Layer 2 — Raw artifacts (passport scan, will PDF, statements).** Device-only by default
   (Secure Enclave, T1/T2). **Opt-in** client-encrypted backup for high-loss-fear items. Echo
   never holds the key. Key realization: you almost always present the *VC derived from* an
   artifact, not the artifact.
3. **Layer 3 — On-chain (Identity Metagraph).** Hashes, status lists, DID only — existing T5–T7
   model. Credential references are opaque UUIDs → issuer DID + type + hash + status index, **no
   PII**.

### Key hierarchy

Extends the existing `MasterKey → ApplicationKey → ContextKey` hierarchy
(`internal/crypto/keyderivation.go`):

- `PassportRootKey` = Secure Enclave P-256 + a recovery secret (BIP-39); never leaves the device
  in plaintext, never sent to the server.
- `CredentialSyncKey = HKDF(PassportRootKey, "echo-passport-credential-sync")` → AES-256-GCM for
  Layer-1 blobs.
- `ArtifactKey = HKDF(PassportRootKey, "echo-passport-artifact-{id}")` → per-artifact Layer-2
  backups.

### Holder, not issuer

Echo **aggregates** issuer-signed credentials (EUDI Wallet, ISO 18013-5 mDL, Open Banking,
OpenID4VC) and **natively issues only what it is authoritative on**: trust tier, proof-of-humanity,
Comply org roles, and the Network State citizenship VC. Echo is never the issuer of government ID.

### Payments

Confirm-per-action, bank enforces limits. Echo is the identity + consent layer, **never the money
transmitter**. P2P between Echo users is native (ECHO/stablecoin, self-custodied,
`AllowSpend`→`SpendTransaction`); merchant payments route through licensed rails. **No raw PAN in
a VC** — the VC attests "verified Visa ending 1234"; the instrument is a tokenized reference held
encrypted in Layer 2, never presented. (Recovery is the subject of ADR 0004.)

## Consequences

### Positive

- Preserves the zero-PII invariant: Echo holds no decryption key and no readable PII, so it is not
  a data controller of plaintext and not a breach target.
- Credentials are available across surfaces (web/glasses/agent) because Layer 1 syncs.
- Reuses ~70% of existing credential/identity/payment code; the vault is additive plumbing.
- Selective disclosure means verifiers see "over 21", never the document.

### Negative

- **Recovery becomes the make-or-break UX problem** (no server backstop). Addressed by ADR 0004
  (social/threshold recovery); under-investing there pushes users toward losing data or demanding
  a backdoor (which would recreate the honeypot).
- Two storage code paths (synced credentials vs device-only artifacts) and two threat models.
- External payment rails (Wave C) require a licensed partner; jurisdiction/licensing is an open
  item (see WO-302).

### Neutral

- "Vault" is the concept; the product brand is **Echo Passport** (ties to Network State
  citizenship and avoids the storage-honeypot connotation).

## Alternatives considered

### Option B — Custodial encrypted blobs with a server-recoverable key
Rejected. Any server-held key (even "for convenience") recreates the honeypot and makes Echo a
data controller of recoverable PII, violating the T0–T7 invariant.

### Option C — Pure device-only, no sync
Rejected. Fails the core use case (web/glasses/agent access without the phone) and makes
device loss catastrophic. Layer 1 client-encrypted sync gives availability without a server key.

### Option D — Echo issues all credentials natively (incl. government ID)
Rejected. Makes Echo a regulated IDV provider / data controller with full KYC liability and a
chicken-and-egg trust problem. Being the holder rides existing regulatory rails (EUDI, mDL,
Open Banking) instead.

## Implementation status

- [x] `pkg/passport/` holder model + aggregation API (WO-293 — in progress)
- [ ] `pkg/storage/encblob/` + client-encrypted sync; `data-classification.md` T2 case (WO-294)
- [ ] `passportCredentialRefs` Identity Metagraph state + validators (WO-298)
- [ ] Selective disclosure (SD-JWT) (WO-295)
