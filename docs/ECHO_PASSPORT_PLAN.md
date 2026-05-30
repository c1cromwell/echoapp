# Echo Passport — Verifiable Credential Wallet (4th Product)

> Status: **Accepted — Wave 0 complete (ADR 0004); Wave A implementation in progress (WO-293+).** Authored 2026-05-29.

## Context

Echo today ships three products on a shared privacy protocol: **Echo Protocol** (open base
layer), **Echo Comply** (enterprise compliance messaging), and **Echo Message** (consumer
messaging). The founder wants a **fourth product**: a verifiable credential vault that lets a
user hold identity, payment, legal, and employment credentials and use them — via the app, an
AI interface, or smart glasses — to access services and transact in daily life, as part of the
**Network State** vision.

**The reframe (honest feedback):** taken literally ("store PII, card numbers, passports,
deeds"), the idea is a centralized honeypot that directly violates Echo's foundational invariant
— *zero readable PII on any server or chain* (the T0–T7 classification, enforced in CI). But
~70% of the *correct* version already exists in the codebase. Echo already issues W3C VC 2.0
credentials, runs StatusList2021 revocation, verifies presentations via OIDC4VCI, derives
self-sovereign `did:key` identities, anchors them on the Identity Metagraph, and has designed the
`AllowSpend`/`SpendTransaction` primitives for consent-bound payments. **Echo Passport is therefore
the holder-side wallet + selective-disclosure + agentic-presentation layer over credentials Echo
mostly already produces — not a new PII database.** This plan re-scopes the idea onto that
foundation, splits it across the roadmap, and wires it into ECHO tokenomics.

### Decisions locked with the founder
- **Custody = hybrid with an opinionated default** (not per-item burden on the user):
  - *Layer 1 — Credentials (signed VCs):* client-encrypted, synced to IPFS/Storj so they're
    available on web / glasses / agent **without the phone in hand**. Echo holds ciphertext only.
  - *Layer 2 — Raw artifacts (passport scan, will PDF, statements):* device-only by default
    (Secure Enclave, T1/T2); **opt-in** client-encrypted backup for high-loss-fear items. Echo
    never holds the key. Key insight: you almost always present the *VC*, not the artifact.
  - *Layer 3 — On-chain:* hashes, status lists, DID only (existing T5–T7 model).
- **Payments = confirm-per-action, bank enforces limits.** Each payment is a fresh,
  time-limited, biometric-gated approval (`AllowSpend` → `SpendTransaction`); the agent/glasses
  never holds standing authority.
- **Echo stays the identity + consent layer, never the money transmitter.** P2P "send money in a
  chat" between Echo users is native (ECHO / stablecoin, self-custodied); merchant payments route
  through licensed rails (card-network tokenization / Open Banking PISP partner).
- **Never store a raw PAN/account number in a VC.** The VC attests *"verified Visa ending 1234"*;
  the instrument is a **tokenized reference** held encrypted in Layer 2, never presented.
- **Recovery is the make-or-break for UX** → social/threshold (Shamir) recovery layered on the
  existing 24-word BIP-39 + SMS flow. This is what makes a non-custodial vault feel custodial-easy.
- **Be the holder, not a regulated issuer of government ID.** Aggregate issuer-signed credentials
  (EUDI Wallet, ISO 18013-5 mDL, Open Banking, OpenID4VC). Echo issues natively only what it is
  authoritative on: trust tier, proof-of-humanity, Comply org roles, Network State citizenship.

### Assumptions (override if wrong)
- **v1 credential scope = Identity & Trust first** (reuses the most existing code) **+ P2P
  pay-in-chat**. Legal documents and employment/income land in later waves.
- **Canonical phases** (from the work-order docs): 1 Foundation · 2 Onboarding/Identity &
  Credentials · 3 Messaging Core · 4 Blockchain & Trust Infrastructure · 5 Hidden Folders &
  Privacy Features (payment rails + revenue) · 6 Calls & File Sharing · 7 Advanced Platform
  Features (marketplace, trust scoring, post-quantum). There is **no numbered "Network State"
  phase** — it is the long-term capstone *beyond Phase 7*. Waves below are mapped onto these real
  phases.

---

## Product framing

**Echo Passport** is a two-sided extension of the existing stack:
- **Holder side (Echo Message consumers):** hold credentials, present with selective disclosure,
  pay in chat, recover across devices.
- **Issuer/verifier side (Echo Comply enterprises):** issue credentials to customers, request
  presentations, pay per-verification in ECHO. This is the natural monetization edge and reuses
  Comply's existing enterprise go-to-market.

The vault is also the **trust-tier engine**: climbing Echo's existing Tier 1→5 ladder *is*
collecting credentials (email/phone → KYC → government ID → peer attestations), which already
multiplies messaging rewards 1.0×→3.0×. So the vault plugs into tokenomics that already exist.

---

## Roadmap (waves mapped to real phases)

### Wave 0 — Recovery & Guardian Mini-Design  *(design spike, early Phase 2)*
A short design-only wave **before** any recovery code, because social/threshold recovery has real
UX and collusion trade-offs and is the make-or-break for vault UX.
- **Deliverables (docs only):** threshold scheme choice (Shamir M-of-N share split); guardian
  model (user devices + optional trusted contacts + optional Comply org as institutional
  guardian); share-distribution & rotation UX; collusion / coercion threat model; how it layers
  on the existing 24-word BIP-39 + SMS recovery; guardian-acceptance-as-VC format.
- **Output:** `docs/adr/0004-echo-passport-recovery-model.md` + a recovery work order that Wave A
  implements against. No server-held key under any option — that's the honeypot line.

### Wave A — Holder Wallet Core  *(Phase 2; trust-anchor pieces in Phase 4)*
The minimum that makes "I have credentials and can show them" real and private.
- **Deliverables:** holder data model + aggregation API; client-side encryption & key hierarchy;
  encrypted credential sync to IPFS/Storj; selective disclosure (SD-JWT full derivation, BBS+
  evaluation); credential management UX (list / detail / present); social-threshold recovery.
- **Credential types (reuse existing):** ProofOfHumanity, KYCLite, HighAssurance, Professional,
  DeviceAttestation, plus ISO 18013-5 mDL (Phase 2 WO-199 already in flight) and verified
  mobile/email/selfie.
- **Reuses:** `pkg/credentials/{models,issuer,verifier,formats,statuslist_l1,revocation}.go`,
  `pkg/credentials/oidc4vc/`, `pkg/didkey/didkey.go`, `internal/crypto/{kinnami,x25519_chacha,keyderivation}.go`,
  `internal/logging/ipfs_clients.go` (generalize the encrypted-blob client out of logging),
  `internal/api/enrollment_vc_handlers.go`, Identity Metagraph state in
  `metagraph/modules/shared_data/.../IdentityTypes.scala` + `IdentityValidations.scala`.
- **New:** `pkg/passport/` (holder/aggregation), `pkg/passport/recovery/` (Shamir + guardians),
  `pkg/passport/disclosure/` (SD-JWT/BBS+), generalized `pkg/storage/encblob/` (from ipfs_clients),
  iOS Passport module.

### Wave B — Pay-in-Chat (P2P, native rails)  *(Phase 4)*
The first "transact in daily life" feature, fully inside Messaging, using rails Echo already owns.
- **Deliverables:** "send money to @user" in a chat; consent sheet (biometric) → `AllowSpend`
  (single-use, time-boxed) → `SpendTransaction` on Currency L1; pay a *verified DID*, not a phone
  number (vault's proof-of-humanity VC is the anti-scam guard).
- **Reuses:** Currency L1 (`metagraph/modules/l1/.../Main.scala` — TokenLock/StakeDelegation/
  RewardClaim patterns), Tessellation v3 `AllowSpend`/`SpendTransaction`/`FeeTransaction`,
  existing trust-tier commitments.
- **New:** `pkg/payments/consent/` (per-action authorization), payment UX in iOS chat.

### Wave C — Verifier Marketplace + External Rails + Recovery Hardening  *(Phase 5; verifier marketplace in Phase 7)*
Make the vault useful *outside* Echo and turn the verifier side into revenue.
- **Deliverables:** relying-party / verifier API (request a presentation, pay per-verification in
  ECHO); merchant payment via licensed rail adapter (card tokenization / Open Banking PISP
  partner — Echo orchestrates consent, partner moves money); tokenized payment-instrument
  references; recovery hardening (guardian bonds, audited Shamir, account-abstraction-style
  device rotation); legal-document VCs (notarized hash + issuer attestation) and
  employment/income credentials (accredited issuers).
- **Reuses:** OIDC4VCI verifier flow, StatusList2021 revocation, trusted-issuer registry
  (`trusted_issuers` table / WO-118), ECHO payment-rail fee model (0.5–1.5%).
- **New:** `pkg/payments/rails/` (adapter interface + 1 reference adapter), `pkg/passport/verifier/`
  (relying-party API + metering).

### Wave D — Network State Capstone  *(Phase 7 → Network State, the post-P7 capstone)*
The original vision: agentic, glasses-native, citizenship-bound.
- **Deliverables:** **Network State citizenship VC** (staking-gated; the credential that gates
  membership-tier physical access); **agentic/glasses presentation + payment** (a GNAP-style
  consent protocol so an agent presents credentials and initiates per-action-confirmed payments
  on the user's behalf); cross-jurisdiction issuer federation; ZK predicate proofs
  ("resident of X", "income > Y") replacing reveal-the-claim where possible.
- **Reuses:** staking tiers + governance weight (`StakedECHO × TrustTierMultiplier`), AllowSpend
  consent model from Wave B/C, DID + selective disclosure from Wave A.
- **New:** `pkg/agent/consent/` (capability-token / GNAP grant model), citizenship issuance under
  governance, glasses/agent SDK surface.

---

## Implementation specs

### 1. Custody & key hierarchy
Extend the existing 3-tier HKDF hierarchy (`internal/crypto/keyderivation.go`,
`MasterKey → ApplicationKey → ContextKey`):
- **PassportRootKey** derived from Secure Enclave P-256 + a recovery secret (BIP-39). Never leaves
  device in plaintext; never sent to server.
- **CredentialSyncKey** = HKDF(PassportRootKey, "echo-passport-credential-sync") → AES-256-GCM for
  Layer-1 blobs replicated to IPFS/Storj.
- **ArtifactKey** = HKDF(PassportRootKey, "echo-passport-artifact-{id}") → per-artifact key for Layer-2
  opt-in backups.
- All blobs content-addressed; server stores ciphertext + CID only. T-classification: synced
  credentials are **T2** (encrypted, off-device allowed only because client-encrypted), artifacts
  **T1/T2**, on-chain refs **T5–T7**. *Update `docs/data-classification.md` to add the
  client-encrypted-sync case explicitly so CI rules don't false-positive.*

### 2. Recovery (the make-or-break)
`pkg/passport/recovery/`: Shamir-split the recovery secret into N shares (user devices + optional
guardians: trusted contacts or a Comply org). Threshold M-of-N reconstruction. Layered on the
existing 24-word BIP-39 + SMS OTP flow (recovery already partially built in Phase 1). Guardian
acceptance is itself a signed VC. **No server-held key, ever** — that's the honeypot line.

### 3. Selective disclosure
`pkg/passport/disclosure/`: complete the stubbed SD-JWT derivation (the format constant already
exists in `pkg/credentials/formats.go`) for per-claim disclosure; add BBS+ for unlinkable
predicate proofs in Wave D. Present minimum claim ("over 21"), never the source document.

### 4. Consent / payment authorization
- **Wave B (P2P):** consent sheet → mint single-use `AllowSpend` (amount-capped, ~60s TTL) →
  `SpendTransaction` on Currency L1. Biometric gate reuses the hidden-persona re-auth pattern.
- **Wave C (merchant):** same consent UX; `pkg/payments/rails/` adapter calls the licensed
  partner; Echo passes a presentation + tokenized instrument ref, never a PAN.
- **Wave D (agent):** GNAP grant defines *what* the agent may present/spend; every payment still
  requires a per-action confirmation (push to phone / glasses tap). No unlimited approval — ever.

### 5. Metagraph state (Identity + Currency)
Extend `IdentityOnChainState` (`IdentityTypes.scala`) with a `passportCredentialRefs:
Map[String, CredentialRef]` (opaque UUID → issuer DID + type + hash + status index; **no PII**)
and validators in `IdentityValidations.scala` (mirror `validateVCIssuance` /
`validateStatusList2021`). Currency L1 (`l1/.../Main.scala`) already validates the spend
primitives. **Build gotcha:** changing the Identity sealed-trait state requires recompiling ALL
metagraph modules (`-Werror` turns non-exhaustive matches into failures).

### 6. APIs (extend `openapi.yaml`)
`/v1/passport/credentials` (list/get), `/v1/passport/present` (begin/accept, reuse OIDC4VCI),
`/v1/passport/recovery/{setup,initiate,complete}`, `/v1/passport/sync` (push/pull ciphertext),
`/v1/pay/p2p` (Wave B), `/v1/verify/request` (Wave C verifier side).

### 7. iOS
New Passport module beside the existing identity/credential code; SwiftData encrypted at rest
(existing AES-256-GCM model); present-flow UI; consent sheets; recovery setup; pay-in-chat entry
point in the conversation view.

---

## Tokenomics integration

**Principle: keep the holder side feeless** (consistent with the existing model where messaging
and identity are feeless); **monetize the verifier/enterprise side, external payment rails, and
premium storage.** Don't over-tokenize the user.

| Mechanism | Token flow | Status |
|---|---|---|
| **Trust-tier engine** — collecting vault credentials raises Tier 1→5 → existing **1.0×–3.0× messaging-reward multiplier** | Earn (existing 400M community pool) | **Reuse, no new mechanic** |
| **ID-verification reward** — 100 ECHO one-time on high-assurance verification | Earn | **Already specified** |
| **Verifier presentation fee** (Wave C) — relying party pays per presentation request | 70% holder / 30% Privacy Commons Treasury (mirror existing Data-Sovereignty split) | New |
| **P2P pay-in-chat** (Wave B) | Native ECHO/stablecoin transfer; minimal/zero protocol fee | Reuse Currency L1 |
| **Merchant payment-rail fee** (Wave C) — 0.5–1.5% per txn | 100% community treasury | **Already specified** |
| **Premium vault** — Layer-2 encrypted-blob backup, more slots, priority recovery | VIP subscription ($9.99/mo) **or** staking-gated tier | New (folds into existing VIP) |
| **Network State citizenship VC** (Wave D) — gates membership-tier physical access | **Staking-gated** (existing Bronze→Platinum lock tiers) | New, ties vault → staking → Network State |
| **Large-artifact storage metering** (optional) | ECHO-metered IPFS/Storj for big Layer-2 backups; credentials always free | Optional |

No change to the 1B fixed supply or the 40/22/18/10/10 allocation. The vault **drives demand**
for existing sinks (staking for citizenship, fees to treasury, trust-tier participation) rather
than minting anything new.

---

## Critical files

**Reuse / extend:** `pkg/credentials/*`, `pkg/credentials/oidc4vc/*`, `pkg/didkey/didkey.go`,
`internal/crypto/{kinnami,x25519_chacha,keyderivation}.go`, `internal/logging/ipfs_clients.go`,
`internal/api/enrollment_vc_handlers.go`, `metagraph/modules/shared_data/.../IdentityTypes.scala`,
`.../IdentityValidations.scala`, `metagraph/modules/l1/.../Main.scala`, `internal/rewards/*`,
`docs/data-classification.md`, `openapi.yaml`, `trusted_issuers` (WO-118).

**Create:** `pkg/passport/` (+ `recovery/`, `disclosure/`, `verifier/`), `pkg/storage/encblob/`
(generalized from ipfs_clients), `pkg/payments/{consent,rails}/`, `pkg/agent/consent/`, new
migrations (`passport_credential_ref`, `passport_sync_blob`, `passport_recovery_share`), iOS Passport module.

**Roadmap docs to update (this task — docs only):** add Echo Passport work orders to
`docs/phase-2-work-orders.md` (Wave 0 mini-design + Wave A holder core), `phase-4-work-orders.md`
(Wave B pay-in-chat + trust anchors), `phase-5-work-orders.md` (Wave C external rails + recovery
hardening + legal/employment creds), `phase-7-work-orders.md` (Wave C verifier marketplace +
Wave D agentic/citizenship → Network State capstone); add the 4th product to the product section
of `docs/Echo_Combined_Requirements.md`; record ADRs
`docs/adr/0003-echo-passport-custody-model.md` and `docs/adr/0004-echo-passport-recovery-model.md`.

---

## Verification

- **Unit:** key-hierarchy derivation determinism; Shamir M-of-N reconstruction round-trip;
  SD-JWT selective-disclosure produces only requested claims; AllowSpend single-use/TTL
  enforcement; presentation verification against StatusList2021 (extend
  `test/tokenomics/` and credential tests).
- **Metagraph:** recompile **all** modules after Identity state change (`-Werror`); validator
  tests for the new `passportCredentialRefs` state mirroring existing `IdentityValidations` tests.
- **E2E:** issue → sync (encrypted) → recover on a second device → present with selective
  disclosure → P2P pay-in-chat → confirm `SpendTransaction` on Currency L1 — run under
  `make dev` per `docs/E2E_LAUNCH_AND_TESTING.md`.
- **CI invariant:** Semgrep T0–T7 must stay green; verify no raw PII or PAN enters any synced
  blob, VC, or chain payload (add fixtures that *should* trip the rules).
- **iOS:** `swift build` library + security-test targets (hard gates per README).

---

## Risks & open items (resolved with founder)
1. **Phase mapping — RESOLVED.** Canonical work-order phases confirmed; waves remapped: Wave 0/A
   → Phase 2, Wave B → Phase 4, Wave C → Phase 5 (+ marketplace Phase 7), Wave D → Phase 7 →
   Network State capstone.
2. **Money-transmitter / PISP licensing — DEFERRED (skip for now).** Not a blocker for the doc/
   roadmap work; revisit when Wave C external rails are actually scheduled.
3. **Recovery guardian model — now Wave 0.** Promoted to an explicit design-only mini-design wave
   (ADR `0004`) ahead of Wave A implementation.
4. **Legal-document validity — RESOLVED: hash.** Deeds/wills are modeled as **notarized-hash
   VCs** (issuer-attested hash + reference), never the legal instrument itself.
5. **Selective disclosure — RESOLVED: SD-JWT.** Wave A ships on SD-JWT (format already stubbed in
   `pkg/credentials/formats.go`). BBS+/ZK predicate proofs are an *optional* Wave D enhancement,
   not a dependency.
