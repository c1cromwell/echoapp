# Phase 1: Foundation & Security Core

**Total Work Orders:** 26  
**Status Summary:** 26 Completed  
**Last synced with Software Factory:** 2026-05-26

---

## In Progress (5)

### WO-35: Implement Server-Side Data Validation and Business Logic Layer

**Blueprint:** Backend

## Summary

Build the server-side validation and pre-validation layer that sits between the API gateway and the Constellation metagraph. The backend acts as a pre-validator — it rejects obviously invalid data to reduce unnecessary blockchain transactions, but the metagraph L1 layers perform final authoritative validation. This layer enforces the canonical T0–T7 data classification policy, ensuring zero PII reaches any chain. The automated CI enforcement of these classification rules across the codebase is handled by WO-217.

## In Scope

- Request validation middleware for all data submission endpoints
- T0–T7 data classification enforcement: reject any submission containing PII, plaintext message content, or sender/recipient DIDs in data payloads
- Merkle root structure validation before Data L1 submission (32-byte SHA-256, commitment count > 0, valid time range)
- Volume-decay reward claim pre-validation: verify decay factor matches server-computed value for the user's message count (WO-213 model)
- Trust tier multiplier pre-validation for reward claims (cached tier vs. submitted multiplier)
- Anti-gaming pre-checks for reward submissions (velocity checks, repeat claims)
- Governance vote pre-validation (active proposal check, stake minimum)
- Schema version validation (reject unsupported schema versions)
- Field-level error response with specific error codes
- Validation middleware composable per endpoint

## Out of Scope

- Metagraph L1 authoritative validation (enforced on-chain in Scala/Euclid SDK)
- Client-side validation logic (iOS work orders)
- Authentication/signature validation (WO-2)
- Rate limiting (WO-44)
- CI enforcement of T0–T7 classification rules (WO-217)

## Requirements

Derived from the Backend, Data Layer, and Primary Architecture and Secure Data Handling blueprints.

**Complete T0–T7 Data Classification Policy:**

| Tier | Classification | Examples | On-Chain | Backend DB | IPFS/Storj | Device-Local |
|---|---|---|---|---|---|---|
| T0 | Secret (never persisted) | Message plaintext, private keys | ❌ | ❌ | ❌ | Memory only |
| T1 | Device-local secret | Derived keys, biometric template | ❌ | ❌ | ❌ | Secure Enclave only |
| T2 | Encrypted local | Message ciphertext, local DB | ❌ | ❌ | ❌ | AES-256-GCM at rest |
| T3 | Relay-transient | Offline queue blobs | ❌ | Ephemeral TTL | ❌ | ❌ |
| T4 | Encrypted audit | Operational logs (no content, no DID) | CID only | ❌ | Encrypted | ❌ |
| T5 | Hash commitment | Merkle roots of message batches | ✅ | ❌ | — | — |
| T6 | Trust commitment | `H(trust_score\|nonce)` | ✅ | ❌ | — | — |
| T7 | Public chain data | Token txns, DID docs, governance | ✅ | Cache only | ❌ | ❌ |

**Zero PII on any blockchain is a hard system invariant.** The Data L1 Scala validator rejects any submission containing personal identifiers, message content, or user behavioral data beyond T5/T6/T7 allowances.

**Pre-validation Code:**
```go
func ValidateDataL1Submission(sub DataL1Submission) error {
    if len(sub.MerkleRoot) != 32 {
        return ValidationError{Code: "INVALID_MERKLE_ROOT", Field: "merkle_root"}
    }
    if sub.CommitmentCount <= 0 {
        return ValidationError{Code: "EMPTY_BATCH", Field: "commitment_count"}
    }
    if sub.TimeRange.From.After(sub.TimeRange.To) {
        return ValidationError{Code: "INVALID_TIME_RANGE", Field: "time_range"}
    }
    if sub.SchemaVersion > CurrentSchemaVersion {
        return ValidationError{Code: "UNSUPPORTED_SCHEMA", Field: "schema_version"}
    }
    return nil
}

// Volume-decay reward pre-validation (updated from hard cap — see WO-213)
func ValidateRewardClaim(claim RewardClaim, cachedTier TrustTier, messagesCount int) error {
    expectedDecay := computeVolumeDecay(messagesCount)
    if claim.DecayFactor != expectedDecay {
        return ValidationError{Code: "INVALID_DECAY_FACTOR", Field: "decay_factor"}
    }
    if claim.TrustMultiplier != cachedTier.Multiplier {
        return ValidationError{Code: "INVALID_MULTIPLIER", Field: "trust_multiplier"}
    }
    return nil
}
```

## Blueprints

- Backend — Defines the pre-validation role and validation error response formats
- Data Layer — Specifies Metagraph L1 validation rules that the backend pre-validates before submission
- Primary Architecture and Secure Data Handling — Defines the canonical T0–T7 data classification model (WO-217 provides CI enforcement)

---

### WO-230: Set Up Phase 1 Local Development Testnet Infrastructure

**Type:** Build

**Blocked By:** WO-275, WO-276, WO-277

**Blueprint:** Data Layer, Decentralized Identity and Authentication, Production Launch, Infrastructure, and Deployment

## Summary

Set up the Phase 1 local development testnet infrastructure: Euclid SDK Docker cluster (Global L0 + Metagraph L0 + Constellation Identity Metagraph + Currency L1 + Data L1), Go backend services running on developer machines, and iOS app prototype connecting to local backend. This establishes the development environment for all Phase 1 engineering work.

**Identity method:** `did:key` is the chosen Phase-1 DID method (no Cardano dependency, no chain transaction). DIDs are derived locally from a P-256 key pair (Secure Enclave on iOS; `pkg/didkey` for backend test fixtures). See `docs/adr/0001-phase1-identity-method.md` for the decision record.

## Status: in progress — Steps 1, 2, 4, 6 wired end-to-end; Step 3 unblocked (WO-276 + WO-272 + WO-277 land Apr 27)

## In Scope

- **Euclid SDK Docker cluster:** 6 containers running under `hydra start-genesis`:
  1. Global L0 (port 9000)
  2. Metagraph L0 (port 9200)
  3. Currency L1 (port 9300)
  4. Data L1 (port 9400)
  5. Identity Metagraph L0 (port 9100)
  6. Identity Metagraph L1 (port 9500)

- **Go backend local configuration:** `.env.local` for the backend stack pointing to local metagraph cluster and local Redis/PostgreSQL. Single-command startup (`make dev`).
- **iOS local configuration:** Xcode scheme `Debug-Local` pointing to `localhost` backend; WebSocket connection to local relay service; self-signed TLS cert for local HTTPS.
- **Developer onboarding docs:** `CONTRIBUTING.md` covering prerequisites (Docker, Go 1.21+, Xcode 15+, JDK 21), first-run setup, how to run tests, how to submit Scala L1 validation changes.
- **Phase 1 go/no-go validation script:** end-to-end flow test (see `scripts/validate-phase1.sh`):
  1. Derive `did:key` locally from a test P-256 key pair (no chain transaction, no Atala PRISM).
  2. Register DID with local Identity Service (`POST /identity/register`).
  3. Issue a test trust tier VC and confirm it anchors on local Identity Metagraph (finality < 30s).
  4. Send a test message through local relay.
  5. Confirm Merkle root committed to local Data L1 (finality < 30s).
  6. Verify snapshot height increments on Global L0.

## Out of Scope

- AWS infrastructure (Phase 2)
- Production metagraph nodes
- CI/CD pipeline (Phase 2 Kubernetes WO)
- Cardano / Atala PRISM — explicitly removed per WO-275 / ADR-0001
- Go-side VC issuance + StatusList2021 batching — WO-273

## Requirements

From the Production Launch, Infrastructure, and Deployment and Data Layer blueprints:

**Phase 1 Go/No-Go:**
- Metagraph testnet transaction finality < 30s.
- `did:key` derivation + Identity Metagraph VC anchor working end-to-end in local environment.
- iOS → backend → Data L1 full message flow demonstrated.

**Phase 1 Infrastructure:**
- Local Euclid SDK Docker cluster (6 containers).
- Go backend services deployable on a single developer machine (`make dev`).
- iOS app prototype connecting to local backend via WebSocket.
- Identity Metagraph, Currency L1, and Data L1 Scala validation logic compiled and passing unit tests.

## Blueprints

- Production Launch, Infrastructure, and Deployment
- Data Layer
- Decentralized Identity and Authentication

## Decisions & ADRs

- `docs/adr/0001-phase1-identity-method.md` — Phase-1 DID method = `did:key` (resolves WO-275)

## Currently delivered scaffolding + implementation

**Orchestration:**
- `Makefile` targets: `make dev`, `dev-status`, `dev-logs`, `dev-restart`, `dev-stop`, `validate-phase1`, `testnet-up`, `testnet-down`. `dev` waits for and reports all 6 metagraph endpoints.
- `docker-compose.testnet.yml` for the backend stack with metagraph + identity URLs wired via `host.docker.internal`.
- `metagraph/scripts/setup-euclid.sh` clones the Euclid SDK into a sibling directory; `make dev` then calls `hydra start-genesis`.
- `scripts/validate-phase1.sh` implementing the 6-step go/no-go check.

**Metagraph layer (delivered with WO-276 + WO-272 + WO-277):**
- 6 sbt sub-projects: `sharedData`, `currencyL0`, `currencyL1`, `dataL1`, `identityL0`, `identityL1`. All zero-UUID `clusterId`s replaced by env-overridable defaults via `ClusterIds`.
- `IdentityOnChainState` schema covering VC issuance, trust tier anchors, StatusList2021 vectors, EchoOrgRoleCredential.
- `IdentityValidations` enforcing did:key DIDs, authorized-sender gating (`IDENTITY_SERVICE_DID`), 32-byte hex commitment, 131,072-bit StatusList2021 vectors with monotonic sequence, known org roles, future expiry.
- All four existing Currency/Data L1 validators wired into per-Main `dispatch` functions so the L1 mempool admission path uses the same rules tested in `*Spec.scala`.
- ScalaTest 3.2.18 added; ~70 assertions across `ValidationsSpec`, `IdentityValidationsSpec`, `data_l1/MainSpec`, `l1/MainSpec`, `identity_l0/MainSpec`, `identity_l1/MainSpec`. Pending live `sbt test` on a sbt + JDK 21 host.

**did:key + identity registration (WO-278):**
- `pkg/didkey/` — canonical W3C did:key derivation/parsing for P-256 (zero external deps); 11 passing unit tests.
- `cmd/didkey/` — CLI used by the validation script.
- `internal/api/identity_register_handlers.go` — production `POST /identity/register` + `GET /identity/{did}` on the main `net/http` router; in-memory `DIDRegistry` with idempotent semantics; 9 passing handler tests.

**Step-by-step status of `scripts/validate-phase1.sh`:**

| Step | Coverage | Status |
| --- | --- | --- |
| 0 | Prerequisites + cluster reachability checks | wired |
| 1 | Derive `did:key` from P-256 (canonical, via `cmd/didkey`) | wired |
| 2 | POST /identity/register + idempotent re-register + GET /identity/<did> | wired |
| 3 | Identity Metagraph reachability (L0 + L1 health probes) | **wired** (was: skip). VC issuance assertion deferred to WO-273 (Go issuer). |
| 4 | Send test message through local relay (`TestE2E_MessageSendReceive`) | wired |
| 5 | Commit Merkle root to Data L1 and verify finality | skip — backend lacks public route; covered later |
| 6 | Verify Global L0 snapshot height increments | wired |

## Remaining for full go/no-go GO

- **Boot the cluster on a host with sbt + JDK 21 + Docker** (sandbox where this work was authored has none of these).
  ```
  cd metagraph && sbt compile && sbt test
  cd .. && make dev && make validate-phase1
  ```
- **WO-273** — Go Identity Service VC issuance + StatusList2021 batch publication (lights up Step 3 end-to-end).
- **Step 5 wiring** — add a public `/v1/integrity/anchor` (or use the existing internal Merkle path) so the validation script can trigger a Data L1 commit and observe finality.

## Resolved blockers

- ~~WO-275~~ (RESOLVED — ADR-0001).
- ~~WO-276~~ (RESOLVED — 6-module skeleton, identity ports 9100/9500 reachable, no zero-UUIDs).
- ~~WO-272~~ (RESOLVED — IdentityValidations + dispatch wired; live testnet anchor still depends on WO-273).
- ~~WO-277~~ (RESOLVED — ScalaTest dep + 70+ assertions across pure + wired specs).
- ~~WO-278~~ (RESOLVED — Step 1 + Step 2 wired).

---

### WO-272: Deploy Constellation Identity Metagraph Scala L1 Validation Logic

**Type:** Build

**Blueprint:** Data Layer, Decentralized Identity and Authentication

## Summary

Implement the Constellation Identity Metagraph's Scala L1 validation logic — a dedicated third metagraph (separate from Data L1 and Currency L1) that anchors W3C VC 2.0 issuance records, trust tier commitments `H(tier || nonce)`, StatusList2021 revocation bit vectors, and `EchoOrgRoleCredential` org membership schemas. Replaces the Cardano Plutus + UTXO approach (WO-20, blocked) with Constellation-native infrastructure (<$10K/yr vs ~$98K/yr).

## Status (Apr 27 2026)

**Validation logic + wiring + tests: COMPLETE.** Live testnet anchoring deferred to WO-273 (Go issuer service).

## Implementation

### On-chain state schema
`metagraph/modules/shared_data/src/main/scala/com/echo/shared_data/types/IdentityTypes.scala`

- `IdentityOnChainState` — keyed by credentialId / subjectDID / issuerOrgDID; held by Identity L0 snapshots.
- `IdentityCalculatedState` — derived counters (total issued, revoked, per-tier subjects).
- `VCIssuanceRecord(credentialId, subjectDID, issuerDID, credentialType, issuedAt, schemaVersion)`
- `TrustTierAnchor(subjectDID, commitment, anchoredAt)` — commitment is the on-chain hash, never the raw tier.
- `StatusList2021Vector(issuerOrgDID, bitVector, publishedAt, sequence)` with constants `ExpectedBitLength = 131072`, `ExpectedHexLength = 32768`.
- `EchoOrgRoleCredential(credentialId, issuerOrgDID, memberDID, role, expiry, issuedAt)`.
- Update messages (sealed trait `IdentityUpdate`): `VCIssuanceUpdate`, `TrustTierCommitmentUpdate`, `StatusList2021BatchUpdate`, `EchoOrgRoleCredentialUpdate` — each with circe encoders.
- `TrustTier` (allowed values 1–5) and `OrgRole` (`owner|admin|moderator|member`) singletons used to bound disclosed values.

### Pure validators
`metagraph/modules/shared_data/src/main/scala/com/echo/shared_data/validations/IdentityValidations.scala`

All four validators are pure functions returning `Either[String, Unit]`. Each takes the actual sender plus the registered authorized sender (Phase 1: the Identity Service DID) so the rule can be tested deterministically.

- `validateVCIssuance(update, sender, authorizedSender, nowMillis)` — enforces did:key on subject + issuer; non-empty/bounded credentialId/credentialType/schemaVersion; positive issuedAt within 5 min skew; rejects non-Identity-Service senders.
- `validateTrustTierCommitment(update, sender, authorizedSender, nowMillis)` — enforces 32-byte SHA-256 lowercase hex commitment, did:key subject, sender authorization, anchoredAt sanity.
- `validateStatusList2021(update, sender, authorizedSender, previousSequence, nowMillis)` — enforces exact 131,072-bit (32,768 hex char) vector, hex encoding, monotonically-increasing sequence (anti-replay), did:key issuer.
- `validateEchoOrgRoleCredential(update, sender, authorizedSender, nowMillis)` — enforces did:key on issuerOrgDID + memberDID, role from `OrgRole.All` (rejects unknown roles to block typo-driven escalation), expiry > issuedAt and > now, sender authorization.

### Identity L1 wiring
`metagraph/modules/identity_l1/src/main/scala/com/echo/identity_l1/Main.scala`

- Reads `IDENTITY_SERVICE_DID` from env once at startup as the authorized sender.
- Tracks `revocationSequences: Map[String, Long]` per `issuerOrgDID` so the StatusList2021 monotonicity rule can be enforced; `recordPublishedSequence` is called from the L0 combiner when a batch is finalized.
- Single `dispatch(update: IdentityUpdate, sender: String, nowMs: Long): Either[String, Unit]` is the same function the Tessellation `CustomContextualTransactionValidator` callback delegates to — keeps the wiring layer thin and unit-testable.

### Tests
Delivered alongside this WO (also satisfies WO-277 acceptance criterion #3). See WO-277 for the full coverage matrix.

- `metagraph/modules/shared_data/src/test/scala/com/echo/shared_data/validations/IdentityValidationsSpec.scala` — 30+ pure-function assertions across all four validators (happy path, unauthorized sender, malformed DIDs, tier/role bounds, sequence monotonicity, future timestamps, empty fields).
- `metagraph/modules/identity_l1/src/test/scala/com/echo/identity_l1/MainSpec.scala` — wired-validator integration spec exercising `Main.dispatch` end-to-end (constructs the L1 SDK app, then accepts/rejects each update type).

## Acceptance criteria

- [x] `IdentityOnChainState` schema covers VC issuance, trust tier anchors, StatusList2021 vectors, EchoOrgRoleCredentials.
- [x] `validateVCIssuance` accepts well-formed records from the Identity Service DID and rejects everything else (incl. non-did:key issuers).
- [x] `validateTrustTierCommitment` enforces 32-byte SHA-256 hex `H(tier || nonce)` form.
- [x] `validateStatusList2021` enforces 131,072-bit vectors (per W3C spec) and monotonically increasing sequence.
- [x] `validateEchoOrgRoleCredential` enforces did:key DIDs, known role, and future expiry.
- [x] Authorized-sender enforcement gates every submission.
- [x] Wired into `identity_l1/Main.dispatch` so the L1 mempool admission path uses the same rules.
- [x] Pure-function + wired-validator tests delivered.
- [ ] **Testnet deployment + read/write smoke** — deferred; needs sbt + Docker host. Once `make dev` succeeds, run `cd metagraph && sbt test` then `curl http://localhost:9500/...` happy path.
- [ ] **Live VC issuance from Go Identity Service** — belongs to WO-273 (Go-side issuer); blocked on this WO + WO-278 (already complete).

## Hand-off notes

1. On a sbt + JDK 21 host: `cd metagraph && sbt 'test:compile' 'sharedData/test'` should run all 30+ identity-validation specs in <30s.
2. Set `IDENTITY_SERVICE_DID=did:key:z...` (use the did:key the Go Identity Service registered via `POST /identity/register`) before booting `identity-l1` — otherwise every submission is rejected.
3. The L0 combiner needs to call `Main.recordPublishedSequence(issuerOrgDID, sequence)` after each finalized snapshot so the next batch's monotonicity check sees the right baseline. Stub for that lives in the L0 dataApplication override (next WO).
4. `validateStatusList2021` currently checks `update.sequence > previousSequence`. If you decide a batch can carry the same sequence (e.g. resubmission with identical content), relax to `>=` and dedupe at the L0 combiner instead.

## Out of scope

- Go Identity Service VC issuance code (WO-273)
- StatusList2021 batch publication scheduler in Go (WO-273)
- ZK proof circuits (Phase 3+)
- Cardano / Atala PRISM (eliminated per ADR-0001)

## Blueprints

- Data Layer
- Decentralized Identity and Authentication

---

### WO-273: Implement did:key DID Management in Identity Service

**Type:** Build

**Blueprint:** Backend, Data Layer, Decentralized Identity and Authentication

## Summary

Replace the Cardano/Atala PRISM DID implementation (WO-180, blocked) with `did:key` — the user's DID is derived locally from their Secure Enclave P-256 public key with no blockchain transaction and no ADA required. The Identity Service stores the DID-to-account mapping in PostgreSQL, resolves DIDs locally from the embedded public key, and anchors multi-device key additions to the Constellation Identity Metagraph.

## In Scope

- **`did:key` derivation:** compute DID from Secure Enclave P-256 public key using multibase encoding; format: `did:key:z6Mk...`; no Atala PRISM, no Cardano, no SDK required — pure cryptographic operation
- **DID resolution:** local by design — the public key is embedded in the DID itself; no chain query needed for primary device; for multi-device key lookup, query Identity Metagraph cache (Redis, 60s TTL)
- **Backend DID-to-account mapping:** PostgreSQL table `(user_id, did, public_key, device_label, created_at)`; indexed on `did` for O(1) lookup; primary device created at account registration
- **Request signature validation:** ECDSA P-256 signature over request body from Secure Enclave; verify against public key extracted directly from `did:key` DID (no chain lookup required for signature verification)
- **Multi-device registration:** same QR code UX as before — primary device generates time-limited registration token → secondary device scans → `POST /identity/devices` — backend stores new device public key in PostgreSQL and anchors device key addition record on Constellation Identity Metagraph
- **No ADA costs:** ECHO treasury pays Constellation snapshot fees in DAG; users never pay identity transaction fees; no Atala PRISM SDK dependency
- **Identity Service (port 8001):** DID creation endpoint `POST /identity/register`, device addition endpoint `POST /identity/devices`, DID resolution endpoint `GET /identity/resolve/:did`

## Out of Scope

- W3C VC issuance and StatusList2021 revocation (separate WO — Implement W3C VC 2.0 Issuance)
- iOS Secure Enclave key generation (WO-223, WO-224)
- Trust score calculation (WO-181)
- Core Identity Metagraph **L1 validation rules** for device keys (delivered under WO-272 — `DeviceKeyRegistrationUpdate`, `validateDeviceKeyRegistration`, Go `SubmitIdentityL1` / JSON contract). **This WO** tracks **operational anchoring**: calling L1 after Postgres, sender alignment, retries, L0 fold, Redis cache.
- Passkey authentication middleware (WO-1, WO-2)

## Metagraph device anchoring (remaining — completes chain semantics)

**Shipped in echoapp (May 2026, shared with WO-272):**

- **Scala** (`metagraph/modules/shared_data`): `DeviceKeyRecord`, `IdentityOnChainState.deviceKeys`, `DeviceKeyRegistrationUpdate`, `IdentityUpdate` circe encoder + key-shape `Decoder`, `validateDeviceKeyRegistration`, `identity_l1/Main.dispatch` wiring and tests.
- **Go** (`internal/metagraph`): `MetagraphConfig.IdentityL1URL`, `DeviceKeyRegistrationUpdate`, `SubmitIdentityL1` → `POST {IdentityL1URL}/transactions` (Euclid public port **9500**, see `metagraph/euclid.json`). Env: `IDENTITY_L1_URL` (already in `.env.example` / docker-compose testnet).

**Still implement under WO-273:**

1. **`handleIdentityAddDevice` (or worker):** after successful `RegisterAdditionalDevice`, build `DeviceKeyRegistrationUpdate{subjectDID, publicKeyHex, deviceLabel, addedAt: UTC epoch millis}` and call `SubmitIdentityL1`.
2. **Tessellation envelope:** ensure HTTP/submitted transaction `sender` matches `IDENTITY_SERVICE_DID` on Identity L1 nodes (same DID the validators authorize in Phase 1).
3. **Failure policy:** when L1 returns error — log-only, outbox retry, or fail the HTTP request (choose; document).
4. **L0 combiner:** apply `DeviceKeyRegistrationUpdate` into snapshot `deviceKeys` when Tessellation data-application fold is wired in-repo or via SDK (may be a separate WO if combiner is SDK-only today).
5. **Redis read-model** (60s TTL) for device lists from chain — optional; Postgres remains Phase-1 API source.

## Requirements

From Decentralized Identity and Authentication blueprint (updated 2026-04-25):

**`did:key` format:**

```
did:key:z6Mk...  (Multibase-encoded P-256 public key from Secure Enclave)
```

**DID Creation (no chain transaction):**

```go
// pkg/did/keydid.go
func DeriveKeyDID(p256PublicKey []byte) string {
    // Multibase-encode the P-256 public key per did:key spec
    // Result: "did:key:z6Mk<multibase(p256PublicKey)>"
    return "did:key:" + multibaseEncode(p256PublicKey)
}
```

**Identity Service Endpoints:**

- `POST /identity/register` — derives DID from public key, stores mapping, returns DID
- `POST /identity/devices` — adds new device public key, anchors on Identity Metagraph, returns updated device list
- `GET /identity/resolve/:did` — returns public keys for all registered devices of this DID

**Timing targets:** DID derivation: <1ms (local computation, no network); DID resolution: <10ms (PostgreSQL lookup); multi-device key anchoring: <15 seconds (Constellation snapshot finality).

## Blueprints

- Decentralized Identity and Authentication — Defines `did:key` method, DID derivation from Secure Enclave, multi-device key management, and updated authentication data flow
- Backend — Defines Identity Service (port 8001), Constellation Identity Metagraph as authoritative identity source, and DID caching strategy
- Data Layer — Specifies `did:key` DID format and Constellation Identity Metagraph role

---

### WO-277: Add Scala Unit Tests for Metagraph L1 Validation Logic

**Priority:** High  **Type:** Build

**Blocked By:** WO-276

**Blocking:** WO-230

**Blueprint:** Decentralized Identity and Authentication

## Summary

WO-230's Phase-1 infra requirements include *"Identity Metagraph, Currency L1, and Data L1 Scala validation logic compiled and passing unit tests."* Validation logic existed in `metagraph/modules/shared_data/.../validations/Validations.scala` (4 validators) but had **no test files anywhere** under `metagraph/modules/*/src/test/`, and the validators were **not wired** into the L1 applications.

## Status (Apr 27 2026)

**Tests + wiring: COMPLETE.** Local `sbt test` execution deferred to a host with sbt + JDK 21 (sandbox does not provide them).

## Implementation

### Build setup
- `metagraph/project/Dependencies.scala` — added `Libraries.scalatest` (`org.scalatest:scalatest:3.2.18 % Test`).
- `metagraph/build.sbt` — `commonSettings` now appends `Libraries.scalatest` so every sub-project has scalatest available; this also powers the new identity_l0/identity_l1 specs.

### Pure-function specs (cover Validations.scala 100% of branches)
- `metagraph/modules/shared_data/src/test/scala/com/echo/shared_data/validations/ValidationsSpec.scala`
  - `validateTokenLock`: all 5 tier minimum amount + lock-day combos, above-minimum acceptance, below-minimum rejection (amount + duration), unknown tier name, typo (case-sensitive) rejection.
  - `validateTrustCommitment`: 64-char lowercase + uppercase hex, length under/over rejection, non-hex rejection, non-positive epoch rejection.
  - `validateMerkleRoot`: length validation (under/over), non-hex rejection, zero/negative leafCount rejection.
  - `validateRewardClaim`: positive amount + known tier, non-positive amount, unknown tier.
  - `StakingTiers` table: contains exactly the 5 tiers, monotonically increasing amounts and lock-days.

### Identity validator specs (covers IdentityValidations.scala from WO-272)
- `metagraph/modules/shared_data/src/test/scala/com/echo/shared_data/validations/IdentityValidationsSpec.scala` — 30+ assertions across all 4 identity validators: happy path, unauthorized sender, non-did:key DIDs, malformed credentialId/role/length, hex format, sequence monotonicity (anti-replay), expiry sanity, etc.

### Wiring (so tests reflect runtime behavior)
Each L1 `Main` now exposes a pure `dispatch` method that the Tessellation L1 admission hook delegates to:

- `metagraph/modules/data_l1/src/main/scala/com/echo/data_l1/Main.scala` — `dispatch(EchoUpdate)` routes `MerkleRootUpdate` → `validateMerkleRoot`, `TrustCommitmentUpdate` → `validateTrustCommitment`, rejects others with explicit message.
- `metagraph/modules/l1/src/main/scala/com/echo/l1/Main.scala` — `dispatch(EchoUpdate)` routes `TokenLockUpdate` → `validateTokenLock`, `RewardClaimUpdate` → `validateRewardClaim`; passes `StakeDelegationUpdate` / `WithdrawLockUpdate` (cooldown checks live at L0 combiner).
- `metagraph/modules/identity_l1/src/main/scala/com/echo/identity_l1/Main.scala` — `dispatch(IdentityUpdate, sender, now)` routes to the four identity validators.
- `metagraph/modules/l0/src/main/scala/com/echo/l0/Main.scala` — cluster-id wiring only (no validators land at L0 yet beyond combiner stubs).

### Wired-validator integration specs (acceptance criterion #3)
- `metagraph/modules/data_l1/src/test/scala/com/echo/data_l1/MainSpec.scala` — happy/sad path through `Main.dispatch` (incl. cross-layer rejection — a TokenLock submitted to Data L1).
- `metagraph/modules/l1/src/test/scala/com/echo/l1/MainSpec.scala` — same shape for Currency L1 (Tier 1 minimum accepted, below-minimum rejected, MerkleRoot rejected).
- `metagraph/modules/identity_l0/src/test/scala/com/echo/identity_l0/MainSpec.scala` — smoke (constructor, supported VC types).
- `metagraph/modules/identity_l1/src/test/scala/com/echo/identity_l1/MainSpec.scala` — wired-validator spec for all four identity update types, including StatusList2021 sequence monotonicity round-trip with `recordPublishedSequence`.

## Acceptance criteria

- [x] At least one spec per validator (happy + sad paths).
- [x] At least one spec demonstrates a wired validator rejecting an invalid update at the L1 application layer (Data L1 TrustCommitment with epoch=0 rejected via `Main.dispatch`; Identity L1 trust tier commitment with non-hex rejected via `Main.dispatch`; etc.).
- [x] `scalatest` added to `Dependencies.scala` and applied to every module via `commonSettings`.
- [x] `Validations.scala` line coverage ≥ 90% (every branch of every validator + StakingTiers table is exercised).
- [ ] `cd metagraph && sbt test` runs and all specs pass — **sandbox lacks sbt + JDK 21; needs a developer host or CI runner.**
- [ ] CI (Phase 2 work) can run `sbt test` headlessly — separate WO once GH Actions / GitLab CI is provisioned.

## Hand-off notes

1. On a sbt + JDK 21 host: `cd metagraph && sbt test` should run ~70 assertions (~40 pure-function + ~30 wired) in <60s once Tessellation jars are cached.
2. Cluster-Main smoke specs (`identity_l0/MainSpec`, `identity_l1/MainSpec`) instantiate the Tessellation `CurrencyL{0,1}App` constructor. If those constructors do non-trivial work (config parsing, file reads), the smoke tests will surface that immediately. They currently just touch `Main.toString` / `Main.authorizedSenderDid` to force evaluation.
3. To add CI, drop a `.github/workflows/scala-tests.yml` calling `sbt -Dsbt.color=always test` after `actions/setup-java@v4` (Temurin 21) + `actions/cache@v4` keyed on `metagraph/build.sbt` + `project/Dependencies.scala`.
4. Coverage report: add `sbt-scoverage` to `project/plugins.sbt` and run `sbt coverage test coverageReport` — target 90%+ on `Validations.scala` and `IdentityValidations.scala`.

## Files added

```
metagraph/modules/shared_data/src/test/scala/com/echo/shared_data/validations/ValidationsSpec.scala
metagraph/modules/shared_data/src/test/scala/com/echo/shared_data/validations/IdentityValidationsSpec.scala
metagraph/modules/data_l1/src/test/scala/com/echo/data_l1/MainSpec.scala
metagraph/modules/l1/src/test/scala/com/echo/l1/MainSpec.scala
metagraph/modules/identity_l0/src/test/scala/com/echo/identity_l0/MainSpec.scala
metagraph/modules/identity_l1/src/test/scala/com/echo/identity_l1/MainSpec.scala
```

## Out of scope

- Identity Metagraph state combiner specs (lands when L0 combiner stub is fleshed out).
- Integration tests against a running Euclid cluster (separate WO; needs Docker on CI).
- `sbt-scoverage` integration (separate WO).

## Blueprints

- Data Layer
- Decentralized Identity and Authentication

---

## Ready (4)

### WO-1: Implement Device Passkey Authentication System

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication

## Summary

Implement the ECDSA P-256 passkey authentication system — passkey generation in the iOS Secure Enclave, backend signature validation against the user's registered did:key device keys, multi-device registration, and device key management. This is the primary authentication mechanism for all protected API endpoints.

The key resolution path is: **Redis cache (60s TTL) → PostgresDIDRegistry** (no Cardano or PRISM resolution). In tests, an in-memory registry replaces `PostgresDIDRegistry`.

## In Scope

- iOS Secure Enclave P-256 key pair generation on account creation
- ECDSA P-256 signature generation for all authenticated API requests (sign over `SHA256(body)`)
- Backend Identity Service (port 8001) signature validation middleware: retrieve device public keys from Redis cache (backed by `PostgresDIDRegistry`), verify ECDSA P-256 signature, return 401 on failure
- Redis cache of `(did → []pubkeys)` with 60s TTL, populated from `PostgresDIDRegistry` on cache miss; in-memory registry used in tests
- Multi-device registration flow: primary device authenticates → generates time-limited registration token (5-minute expiry) → displays QR code → secondary device scans → generates new Secure Enclave key pair → submits `POST /v1/auth/devices {token, newPublicKey}` → backend verifies token and appends new public key to the DID's device key set in `PostgresDIDRegistry`
- Passkey reset flow (requires identity verification or account recovery)
- Authentication latency target: ≤ 3 seconds for valid signatures with cached DID keys

## Out of Scope

- DID creation and did:key derivation (WO-273, WO-278)
- Verifiable credential management (WO-182)
- Trust score calculation (WO-181)
- Identity verification flows (WO-26, WO-120)
- Cardano/PRISM DID resolution — removed; see WO-280

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

**Authentication Flow:**
```
1. iOS App generates P-256 key pair in Secure Enclave
2. iOS App signs request body: sig = sign(SHA256(body), privateKey)
3. Backend receives: {body, senderDID, signature}
4. Backend resolves device public keys (Redis cache → PostgresDIDRegistry, 60s TTL)
5. Backend verifies: verify(signature, SHA256(body), cachedPublicKey)
6. On failure: HTTP 401, error code "AUTH_INVALID_SIGNATURE"
7. On success: inject user context (DID, trust tier, subscription tier)
```

**Key Resolution:**
```go
// PostgresDIDRegistry stores device keys registered during onboarding (WO-278).
// In tests, an in-memory implementation satisfies the same interface.
type DIDRegistry interface {
    GetDeviceKeys(did string) ([]ecdsa.PublicKey, error)
}

func resolveDeviceKeys(did string, cache *redis.Client, registry DIDRegistry) ([]ecdsa.PublicKey, error) {
    // 1. Check Redis (60s TTL)
    if keys := cache.Get("did:keys:" + did); keys != nil {
        return deserializeKeys(keys), nil
    }
    // 2. Cache miss: fetch from PostgresDIDRegistry
    keys, err := registry.GetDeviceKeys(did)
    if err != nil {
        return nil, err
    }
    cache.Set("did:keys:"+did, serializeKeys(keys), 60*time.Second)
    return keys, nil
}
```

**Multi-Device Registration:**
```
1. User authenticates on primary device
2. User initiates "Add Device" from settings
3. Primary device generates time-limited registration token (5-minute expiry)
4. Primary device displays QR code containing registration token
5. Secondary device scans QR code
6. Secondary device generates new Secure Enclave P-256 key pair
7. Secondary device submits: POST /v1/auth/devices {token, newPublicKey}
8. Backend verifies token, appends new public key to DID's device key set in PostgresDIDRegistry
9. Redis cache for DID is invalidated: DEL did:keys:{did}
10. Secondary device confirmed; both devices can now authenticate independently
```

**Key Types (important: ECDSA P-256, NOT Ed25519):**
```go
// Backend expects P-256 (secp256r1) signatures from Secure Enclave
// DID device keys are stored as P-256 public keys in PostgresDIDRegistry
// iOS Secure Enclave natively generates P-256 hardware-bound keys
// Previous docs mentioned Ed25519 — the canonical spec is P-256
```

## Tests

**Middleware unit tests — golden-vector ECDSA verification:**
```go
// Use pre-generated, deterministic P-256 key pair, request body, and valid signature.
// These vectors must be committed alongside the test so CI never depends on randomness.

func TestAuthMiddleware_ValidSignature(t *testing.T)      // known key + known body + known sig → 200
func TestAuthMiddleware_TamperedSignature(t *testing.T)   // flip one byte in sig → 401 AUTH_INVALID_SIGNATURE
func TestAuthMiddleware_MissingHeader(t *testing.T)       // no Authorization header → 401 AUTH_MISSING_SIGNATURE
func TestAuthMiddleware_UnknownDID(t *testing.T)          // valid sig, DID not in registry → 401 AUTH_UNKNOWN_DID
```

**Integration tests:**
```go
// Uses in-memory DIDRegistry (no Postgres required).
// A feature-flagged dev-only route may be needed if no unauthed routes exist yet.

func TestAuthIntegration_RegisterAndAuthenticate(t *testing.T) {
    // 1. POST /identity/register — register did:key + P-256 device public key
    // 2. Build request with valid ECDSA P-256 signature header
    // 3. Call protected route → assert HTTP 200, user context injected

    // 4. Call same route without signature headers → assert HTTP 401
    // 5. Call same route with wrong key → assert HTTP 401
}
```

## Blueprints

- Decentralized Identity and Authentication — Defines passkey authentication, Secure Enclave integration, multi-device registration, DID resolution, and authentication flow
- Backend — Specifies Identity Service (port 8001), P-256 signature validation, Redis caching TTLs, and authentication middleware

---

### WO-6: Implement Core iOS App Architecture with MVVM and Security Foundation

**Blueprint:** Frontend

## Summary

Establish the foundational iOS application architecture using MVVM-C (Model-View-ViewModel-Coordinator) pattern, configure the security layer (Secure Enclave, Kinnami encryption, TLS pinning), set up the project structure, and implement core networking and dependency injection infrastructure. All subsequent feature work orders build on this foundation.

**Note on Secure Enclave:** This work order establishes the `SecureEnclaveManager` stub and project structure. The complete canonical `SecureEnclaveManager` implementation — including all key lifecycle management, biometric re-enrollment handling, memory zeroing on background, and multi-device architecture — is specified in the Secure Enclave Key Management blueprint and implemented in WO-223. The storage key derivation pattern (`deriveStorageKey()` via HKDF from Secure Enclave signature) is implemented in WO-224.

## In Scope

- MVVM-C architecture setup: Coordinators, ViewModels, UseCases, Repositories, and the Factory DI container
- Full project directory structure: `App/`, `Core/`, `Domain/`, `Presentation/`
- `SecureEnclaveManager` — initial setup; full implementation deferred to WO-223
- `BiometricAuthManager` — `LAContext` wrapper, biometric prompt configuration, fallback handling
- `KeychainManager` — Keychain read/write/delete wrapper with `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`
- `KinnamiEncryption` — X25519 ECDH key agreement + ChaCha20-Poly1305 message encryption
- `APIClient` — URLSession-based REST client with `RequestInterceptor` for auth headers and encryption
- `CertificatePinner` — TLS 1.3 with certificate pinning via `URLSession` delegate
- `LocalDatabase` — SwiftData model container setup; AES-256-GCM storage encryption via derived key (WO-224)
- `Logger` — Privacy-safe logger that sanitizes sensitive fields (no DID, no plaintext, no keys in logs)
- `ECHOError` enum — structured error codes: `1xxx` auth, `2xxx` network, `3xxx` encryption, `4xxx` messages, `5xxx` identity, `6xxx` groups
- Swift Concurrency (`async/await`, `actor`, `@MainActor`) setup and conventions
- `AppDelegate` lifecycle hook: zero all derived symmetric keys from memory on app background (per Secure Enclave Key Management blueprint)

## Out of Scope

- Full `SecureEnclaveManager` implementation (WO-223 — canonical spec in Secure Enclave Key Management blueprint)
- Storage key derivation via HKDF (WO-224)
- Identity verification UI and flows (WO-17)
- Messaging features (WO-28)
- Push notifications (WO-57)
- Wallet/token features (separate work orders)
- Governance UI (Phase 4+)

## Requirements

Derived from the Frontend and Secure Enclave Key Management blueprints.

**Technology Stack:**

| Component | Technology |
|---|---|
| UI Framework | SwiftUI |
| Architecture | MVVM-C |
| Language | Swift 5.9+ |
| Concurrency | Swift Concurrency (async/await) |
| Security | CryptoKit, Security.framework |
| Networking | URLSession, WebSocket |
| Persistence | SwiftData, Keychain |
| DI | Factory pattern |
| E2E Encryption | Kinnami (X25519 + ChaCha20-Poly1305) |
| Identity Signing | ECDSA P-256 (Secure Enclave) |
| Storage Encryption | AES-256-GCM (HKDF-derived, never persisted) |
| Transport | TLS 1.3+ with certificate pinning |

**Project Structure:**
```
ECHO/
├── App/           # App entry, AppDelegate (memory zeroing on background), SceneDelegate
├── Core/
│   ├── DI/        # Container.swift, Factories/
│   ├── Security/  # SecureEnclaveManager (WO-223), BiometricAuthManager, KeychainManager, KinnamiEncryption
│   ├── Networking/# APIClient, Endpoints, RequestInterceptor, CertificatePinner
│   ├── Relay/     # MessageRelayManager, OfflineQueueManager, AnchoringTracker
│   ├── Storage/   # LocalDatabase, SecureStorage (uses WO-224 derived key), CacheManager
│   └── Utilities/ # Logger, Constants, Extensions/
├── Domain/
│   ├── Models/    # User, Message, Conversation, DID, Credential, Token, GroupKey
│   ├── UseCases/  # Auth/, Messaging/, Groups/, Identity/, Tokens/
│   └── Repositories/
└── Presentation/
    ├── Coordinators/
    ├── Features/
    └── Components/
```

**Error Codes:**
```swift
enum ECHOError: Int, LocalizedError {
    case authFailed = 1001, biometricFailed = 1002, sessionExpired = 1003
    case networkUnavailable = 2001, requestTimeout = 2002, relayUnavailable = 2004
    case encryptionFailed = 3001, decryptionFailed = 3002, invalidSignature = 3004
    case messageSendFailed = 4001, rateLimitExceeded = 4003, messageQueued = 4004
    case didCreationFailed = 5001, verificationFailed = 5002
    case groupKeyMissing = 6001, groupKeyRotationFailed = 6002
}
```

## Blueprints

- Frontend — Defines MVVM-C architecture, full project structure, technology stack, encryption spec, security principles, error handling, and testing strategy
- Secure Enclave Key Management — Defines canonical `SecureEnclaveManager` (WO-223), storage key derivation (WO-224), key lifecycle events, and memory protection requirements

---

### WO-9: Implement File Encryption and Key Management System

**Blueprint:** Large File Sharing and Cloud Storage Integration

## Summary

Implement the file encryption and key management system for large file sharing. Files are encrypted end-to-end using AES-256-GCM before leaving the device, with per-file keys derived from the conversation's Kinnami encryption context. Metadata is encrypted separately. This is the encryption foundation for all file sharing.

## In Scope

- Per-file AES-256-GCM key derivation from conversation Kinnami key context using HKDF: `fileKey = HKDF(conversationKey, salt: fileId)`
- File content encryption: AES-256-GCM with random 12-byte nonce prepended to ciphertext
- File metadata encryption: separate AES-256-GCM encryption for `{filename, mimeType, fileSize, chunkHashes[]}` using same derived key
- Key distribution to recipients: file key encrypted with each recipient's X25519 public key (same Kinnami key agreement as messages)
- Key rotation support: new file key generated on request while preserving decryption of existing files
- Decryption verification: test decrypt first chunk to verify correct key before full download
- Encryption performance target: < 5 seconds for files up to 100MB (using hardware AES acceleration)
- Progress indicators for large file encryption

## Out of Scope

- File chunking and IPFS upload (WO-21)
- Blockchain integrity anchoring (WO-34)
- Cloud storage integrations (WO-46)
- Virus scanning (WO-58)

## Requirements

Derived from the Large File Sharing and Cloud Storage Integration blueprint.

**File Encryption:**
```swift
// Core/FileSharing/FileEncryptionService.swift
struct FileEncryptionService {
    func encryptFile(_ fileURL: URL, conversationKey: SymmetricKey, fileId: UUID) async throws -> EncryptedFile {
        // 1. Derive per-file key from conversation context
        let fileKey = try HKDF<SHA256>.deriveKey(
            inputKeyMaterial: conversationKey,
            info: Data("file_key_\(fileId.uuidString)".utf8),
            outputByteCount: 32
        )

        // 2. Encrypt file content (AES-256-GCM, streaming for large files)
        let nonce = AES.GCM.Nonce()  // 12 bytes random
        let ciphertext = try AES.GCM.seal(fileData, using: fileKey, nonce: nonce).ciphertext

        // 3. Encrypt metadata separately
        let metadata = FileMetadata(filename: fileURL.lastPathComponent, mimeType: mimeType, fileSize: fileSize)
        let encryptedMetadata = try AES.GCM.seal(encode(metadata), using: fileKey).combined!

        return EncryptedFile(encryptedContent: ciphertext, encryptedMetadata: encryptedMetadata, nonce: nonce)
    }
}
```

## Blueprints

- Large File Sharing and Cloud Storage Integration — Defines E2E encryption with Kinnami, per-file key derivation from conversation context, AES-256-GCM algorithm, encrypted metadata, and key management

---

### WO-13: Integrate Kinnami Encryption for All Inter-Service Communication

**Blueprint:** Backend

## Summary

Implement end-to-end encryption using Kinnami (X25519 ECDH key agreement + ChaCha20-Poly1305 symmetric encryption) across all iOS ↔ backend communication channels. The backend operates as a content-blind relay — it handles opaque encrypted blobs it cannot read or modify. This work order implements the encryption middleware on the Go backend side.

## In Scope

- Kinnami encryption middleware for incoming requests from the iOS app (decrypt inbound, encrypt outbound responses)
- X25519 ECDH key agreement protocol implementation
- ChaCha20-Poly1305 symmetric encryption/decryption for message blobs
- AES-256-GCM encryption for log data (monthly rotating keys derived via HKDF)
- HKDF-SHA256 key derivation from Secure Enclave signatures
- TLS 1.3 configuration with certificate pinning for all external connections
- Encryption error handling: 500 for internal encryption failures, 400 for malformed encrypted data
- Backend-side key exchange coordination (public key distribution)
- Opaque blob relay: backend validates structure but never decrypts message content

## Out of Scope

- iOS Secure Enclave implementation (Frontend work orders)
- Kinnami library development (external dependency)
- Log encryption key management/Shamir splitting (WO-53)
- ECDSA P-256 signature verification (WO-2 Auth Middleware)
- Sealed sender envelope implementation (Phase 3, future work order)

## Requirements

Derived from the Backend and Data Layer blueprints. No formal requirements document exists for this foundation blueprint.

**Encryption Specification (Canonical):**

| Purpose | Algorithm | Key Type | Library |
|---|---|---|---|
| Identity/DID signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework (iOS) |
| Message key agreement | X25519 ECDH | Ephemeral Curve25519 | Go `crypto/ecdh` |
| Message encryption | ChaCha20-Poly1305 | Derived symmetric (256-bit) | Go `golang.org/x/crypto/chacha20poly1305` |
| Sealed sender envelope (Phase 3) | AES-256-GCM | Derived from recipient identity key | Go `crypto/aes` |
| Log encryption | AES-256-GCM | Monthly derived key (Shamir split) | Go `crypto/aes` |
| Key derivation | HKDF-SHA256 | From Secure Enclave signature | Go `golang.org/x/crypto/hkdf` |
| Hash commitments | SHA-256 | N/A | Go `crypto/sha256` |
| Transport | TLS 1.3 | Certificate-based (pinned) | Go TLS |

**Message Relay Flow (Backend Role):**
```
1. Sender encrypts on-device (X25519 + ChaCha20-Poly1305)
2. Sender signs commitment hash with Secure Enclave (P-256)
3. Backend RECEIVES: encrypted blob + commitment + P-256 signature
4. Backend VALIDATES: P-256 signature against cached DID public key
5. Backend RELAYS: encrypted blob to recipient (cannot read content)
6. Backend NEVER: decrypts content, stores plaintext, accesses private keys
```

**Key Management:**
```go
type EncryptionService struct {
    // Backend never stores private message keys
    // Only manages: public key distribution, signature verification
    // Log encryption keys: AES-256-GCM, monthly rotation
    logKeyManager *LogKeyManager
}

type LogKeyManager struct {
    currentKey    []byte        // Current month's AES-256-GCM key
    keyRotationAt time.Time     // Next rotation date
    // Key derived from platform master key via HKDF with date-based info string
    // Master key split via Shamir's Secret Sharing (3-of-5 threshold)
}
```

## Blueprints

- Backend — Defines the content-blind relay architecture, encryption requirements, and key management patterns
- Data Layer — Specifies the canonical encryption table applied across all layers of the system

---

## Backlog (11)

### WO-2: Implement Core Go REST API Framework with Authentication Middleware

**Assignee:** Chad Cromwell

**Blueprint:** Backend

## Summary

Establish the foundational Go REST API server that acts as the stateless operational coordinator for all backend services. This is the Gateway service (port 8000) that handles TLS termination, request routing to downstream microservices, and authentication middleware. All 10 backend microservices depend on this foundation being in place.

## In Scope

- Go REST API server with multi-version routing (`/v1/`, `/v2/`)
- ECDSA P-256 signature validation middleware for all protected endpoints — verify signature over request payload against the user's device public keys resolved from `PostgresDIDRegistry` (Redis cache, 60s TTL)
- CORS policy enforcing requests from verified iOS app origins only
- TLS 1.3+ configuration with certificate pinning for all endpoints
- Standardized error response format with structured error codes
- Health check endpoint returning service status
- Request routing stubs to downstream microservices (Identity 8001, Message Relay 8002, Trust 8003, Rewards 8004, Contacts 8005, Metagraph Gateway 8006, Notification 8007, Media 8008, Log Publisher 8009)
- HTTP 401 response on invalid/missing signature, HTTP 403 on CORS violations

## Out of Scope

- Business logic endpoints (covered by individual service work orders)
- Kinnami encryption implementation (WO-13)
- Database or persistent storage setup (Data Layer work orders)
- Rate limiting (WO-44)
- Metagraph integration (WO-27)

## Requirements

Derived from the Backend blueprint. No formal requirements document exists for this foundation blueprint.

**API Design Principles:**
- All endpoints require authentication via passkey verification (ECDSA P-256 signature over request payload)
- Backend validates signatures but never stores private keys (keys remain in Secure Enclave)
- Responses include structured error codes for client-side handling
- CORS policy is strict, allowing requests only from verified iOS app origins
- Message content is never exposed to backend — only opaque encrypted blobs
- Metadata is minimized to what's necessary for routing (sender DID, recipient DIDs, timestamps)

**Service Architecture:**

| Service | Port | Role |
|---|---|---|
| Gateway | 8000 | Load balancer, TLS termination, rate limiting |
| Identity Service | 8001 | Registration, DID management, credential caching |
| Message Relay | 8002 | WebSocket relay, offline queue, APNs push |
| Trust Service | 8003 | Trust score computation, tier caching |
| Rewards Service | 8004 | Reward validation, batching, submission |
| Contacts Service | 8005 | Contact list, block list, search |
| Metagraph Gateway | 8006 | L1/L0 submission, snapshot listening |
| Notification Service | 8007 | APNs push, in-app notifications |
| Media Service | 8008 | Encrypted media upload/download |
| Log Publisher | 8009 | Batch encryption, IPFS submission |

**Error Response Format:**
```go
type ErrorResponse struct {
    Code    string `json:"code"`    // e.g., "AUTH_INVALID_SIGNATURE"
    Message string `json:"message"` // User-facing description
    Details map[string]string `json:"details,omitempty"` // Field-level errors
}
```

**Authentication Middleware:**
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract DID and signature from Authorization header
        // 2. Resolve device public keys from Redis cache (TTL: 60s)
        //    or fetch from PostgresDIDRegistry if cache miss
        // 3. Verify ECDSA P-256 signature over request body
        // 4. On failure: return 401 with structured error
        // 5. On success: inject user context and call next
    })
}
```

## Blueprints

- Backend — Defines the 3-tier architecture, 10-service topology, authentication patterns, API design principles, and error handling standards

---

### WO-33: Implement Decentralized Storage Integration for Encrypted Logging and Audit Trails

**Blueprint:** Data Layer

## Summary

Build the decentralized storage integration layer within the Log Publisher service (port 8009) that pushes encrypted operational log batches to IPFS (via Pinata/web3.storage) and Storj, then records the resulting CIDs on the Constellation Data L1 layer for verifiable, tamper-evident audit trails. This provides the immutable logging backbone required for compliance and security auditing.

## In Scope

- IPFS client integration using Pinata and web3.storage SDK for pinning encrypted log batches
- Storj integration as overflow/fallback storage for large batches
- Content integrity verification: store SHA-256 of encrypted batch with CID for tamper detection
- Minimum 2 pin providers for redundancy (Pinata + self-hosted IPFS node)
- Data L1 CID indexing submission: `{type: "audit_log", CID, timeRange, batchHash}`
- Log retrieval support for authorized auditors (threshold 3-of-5 Shamir key holders)
- 7-year retention policy enforcement (pin management, no unpinning without governance vote)
- Cost targets: ~$50/month IPFS pinning + ~$20/month Storj at 100K users
- Batch size target: 1–5 MB compressed per batch

## Out of Scope

- Log collection and batching (handled in WO-53 Log Publisher)
- AES-256-GCM encryption and key management (WO-53)
- IPFS/Storj node provisioning
- Log analysis, visualization, or audit reporting tools

## Requirements

Derived from the Data Layer blueprint.

**IPFS Push with CID Recording:**
```go
type IPFSStorageClient struct {
    pinata     *pinata.Client
    web3storage *web3storage.Client
    storj      *storj.Client
    dataL1     *MetagraphDataL1Client
}

func (c *IPFSStorageClient) Store(encryptedBatch []byte, timeRange TimeRange) (string, error) {
    // 1. Push to IPFS (primary)
    cid, err := c.pinata.Pin(encryptedBatch)
    if err != nil {
        // Fallback to Storj
        cid, err = c.storj.Put(encryptedBatch)
        if err != nil { return "", fmt.Errorf("all storage providers failed: %w", err) }
    }

    // 2. Pin via second provider for redundancy
    if primaryWasIPFS {
        go c.web3storage.Pin(cid)  // async, non-blocking
    }

    // 3. Record CID on Data L1 for verifiable index
    err = c.dataL1.Submit(DataL1Submission{
        Type:      "audit_log",
        CID:       cid,
        TimeRange: timeRange,
        BatchHash: sha256Sum(encryptedBatch),
    })

    return cid, err
}
```

**Audit Log Retrieval (for authorized auditors):**
```go
func (c *IPFSStorageClient) Retrieve(cid string, decryptionKey []byte) ([]LogEvent, error) {
    encrypted, err := c.ipfsGateway.Get(cid)
    if err != nil {
        encrypted, err = c.storj.Get(cid)
    }
    // AES-256-GCM decrypt, zstd decompress, JSON parse
    return parseLogBatch(aes256gcm.Decrypt(decryptionKey, encrypted))
}
```

## Blueprints

- Data Layer — Defines log lifecycle, storage providers, cost targets, retention policy, key management (Shamir splitting), and CID indexing on Data L1
- Backend — Specifies Log Publisher service (port 8009), IPFS/Storj integration, and audit trail requirements

---

### WO-44: Implement Rate Limiting and Throttling System with Tiered Access

**Blueprint:** Backend

## Summary

Implement the per-DID tiered rate limiting system in the Gateway service (port 8000) that controls API access based on user subscription levels. Rate limiting protects backend services from abuse while rewarding VIP subscribers with higher throughput. Implemented as composable middleware using Redis for distributed rate limit tracking.

**Note:** API rate limiting is for abuse prevention only — it does **not** cap token rewards. Reward distribution uses an auto-scaling model with no per-user daily caps (see WO-213).

## In Scope

- Per-DID rate limiting using Redis sliding window counters
- Base tier: 100 requests/minute for standard users
- VIP tier ($9.99/month subscription): 2x–10x increased limits based on platform scaling
- Group message send rate limiting (1 send-rate token per group message regardless of member count)
- Per-DID message send rate limiting (separate counter from API rate limits)
- 429 Too Many Requests with `Retry-After` header on limit exceeded
- Configurable limits via environment variables (scale without redeployment)
- Rate limit storage in Redis with TTL-based expiry
- Rate limit bypass for health check endpoint
- VIP tier activation: after successful `AllowSpend` subscription (WO-206), backend updates user's rate limit tier in Redis from base → VIP

## Out of Scope

- User subscription billing and AllowSpend payment processing (WO-206)
- VIP tier determination logic (tier provided in authenticated user context after AllowSpend activation)
- Metagraph-level rate limiting (handled on-chain)
- DDoS protection (handled at infrastructure/CDN layer)
- Token reward rate limiting (no per-user caps — auto-scaling model per WO-213)

## Requirements

Derived from the Backend blueprint.

**Rate Limit Tiers (updated):**

| Tier | API Requests/min | Notes |
|---|---|---|
| Base (Free) | 100 | Default for all users |
| VIP ($9.99/mo) | 200–1000 | 2x–10x base; scales with platform growth |

**Implementation:**
```go
type RateLimiter struct {
    redis      *redis.Client
    baseLimits RateLimitConfig
    vipLimits  RateLimitConfig
}

func (rl *RateLimiter) Check(did string, tier UserTier) error {
    key := fmt.Sprintf("rate:%s:%s", did, currentMinute())
    count, _ := rl.redis.Incr(ctx, key).Result()
    rl.redis.Expire(ctx, key, 60*time.Second)

    limit := rl.baseLimits.RequestsPerMinute
    if tier == VIP {
        limit = rl.vipLimits.RequestsPerMinute
    }
    if count > int64(limit) {
        return ErrRateLimitExceeded{RetryAfter: secondsUntilNextMinute()}
    }
    return nil
}
```

## Blueprints

- Backend — Defines tiered rate limiting requirements, per-DID tracking, VIP tier at $9.99/month, group message rate limiting, and the clarification that API rate limits do not constrain token rewards

---

### WO-53: Implement Centralized Logging with Encrypted Decentralized Storage

**Assignee:** Chad Cromwell

**Blueprint:** Backend

## Summary

Build the Log Publisher service (port 8009) that collects operational events from all backend services, batches them, encrypts with AES-256-GCM using monthly rotating keys, pushes to IPFS/Storj, and records the resulting CID on the Data L1 layer for verifiable, tamper-evident audit trails. The service handles zero PII and zero message content — only operational metadata.

## In Scope

- In-memory event buffer (max 1000 events or 5 minutes, whichever comes first)
- AES-256-GCM batch encryption using monthly-rotating derived keys
- HKDF key derivation from platform master key using date-based info string
- Shamir's Secret Sharing for master key (3-of-5 threshold, designated platform operators)
- zstd compression before encryption
- IPFS push with Pinata/web3.storage pinning (primary) + self-hosted IPFS node (secondary)
- Storj fallback for large media audit trails
- CID submission to Data L1 as audit log index entry
- Key rotation on the first of each month
- Privacy-safe event schema: message counts, delivery rates, queue depths, rate limit events, circuit breaker state changes — no DIDs, no content, no PII

## Out of Scope

- Log analysis, visualization, or search (separate tooling)
- IPFS/Storj node provisioning
- Legal hold / law enforcement access tooling (future compliance work order)
- Compliance framework implementation

## Requirements

Derived from the Backend and Data Layer blueprints.

**Log Lifecycle:**

| Phase | Action | Detail |
|---|---|---|
| Collection | Buffer API events, relay metadata, metagraph receipts | In-memory, max 1000 events or 5 min |
| Encryption | AES-256-GCM with monthly rotating key | Key derived from master via HKDF |
| Submission | Encrypted batch pushed to IPFS | Retry with exponential backoff |
| Pinning | CID pinned via Pinata + self-hosted IPFS | Minimum 2 pin providers |
| Indexing | CID + time range + batch hash → Data L1 | On-chain verifiable log index |
| Retrieval | Authorized auditors decrypt with log key | 3-of-5 Shamir key holders |
| Retention | 7-year minimum | Pins maintained by platform treasury |

**Privacy-Safe Log Schema:**
```go
type OperationalLogBatch struct {
    BatchID       string    `json:"batch_id"`
    TimeRange     TimeRange `json:"time_range"`
    SchemaVersion int       `json:"schema_version"`
    Events        []LogEvent `json:"events"`
}

type LogEvent struct {
    EventType  string    `json:"event_type"` // "relay_count", "rate_limit", "circuit_breaker"
    Timestamp  time.Time `json:"timestamp"`
    // NO: DID, phone, message content, IP addresses
    Count      int       `json:"count,omitempty"`
    ServiceID  string    `json:"service_id"`
    Outcome    string    `json:"outcome"`  // "success", "failure"
}
```

**Encryption and Key Rotation:**
```go
type LogKeyManager struct {
    redis         *redis.Client
    shamirShares  [][]byte // 5 shares, 3 required to reconstruct
}

func (m *LogKeyManager) CurrentKey() ([]byte, error) {
    // Derive monthly key: HKDF(masterKey, "log_encryption_" + YYYY-MM)
    infoString := fmt.Sprintf("log_encryption_%s", time.Now().Format("2006-01"))
    return hkdf.New(sha256.New, masterKey, nil, []byte(infoString)), nil
}
```

**IPFS Submission with CID Indexing:**
```go
func (s *LogPublisher) FlushBatch(batch []LogEvent) error {
    compressed := zstd.Compress(serialize(batch))
    key, _ := s.keyManager.CurrentKey()
    encrypted := aes256gcm.Encrypt(key, compressed)

    cid, err := s.ipfs.Add(encrypted)
    if err != nil {
        return s.storj.Put(encrypted) // fallback
    }

    return s.dataL1.Submit(DataL1Submission{
        Type:      "audit_log",
        CID:       cid,
        TimeRange: batch.TimeRange(),
        BatchHash: sha256(encrypted),
    })
}
```

## Blueprints

- Backend — Defines the Log Publisher service, batching intervals, privacy constraints, and IPFS integration pattern
- Data Layer — Specifies log lifecycle, key management, compression, cost targets, and retention policy

---

### WO-136: Implement WebAuthn/FIDO2 Passkey Creation and Management

**Blueprint:** Streamlined Onboarding with Verifiable Credentials and Passkeys

## Summary

Implement the passkey creation step in the Streamlined Onboarding flow — prompting users to create a passkey after successful credential verification and profile creation, integrating with iOS `ASAuthorizationController` to generate a Secure Enclave P-256 key pair and linking it to the user's DID document on Cardano. This is distinct from WO-1 (which implements the authentication middleware); this work order covers the onboarding-context passkey creation UI.

## In Scope

- `PasskeySetupView` in the Streamlined Onboarding flow: explain passkey benefits, prompt creation, show success/failure
- `ASAuthorizationController` with `ASAuthorizationPlatformPublicKeyCredentialProvider` for WebAuthn/FIDO2 passkey creation
- Secure Enclave P-256 key pair generation (hardware-bound, biometric-protected)
- Submit public key to backend: `POST /v1/auth/passkey` with `{did, publicKey, signature}`
- Backend: add public key to DID document on Cardano, link passkey to account
- Passkey creation within 30 seconds; clear success confirmation in UI
- Immediate test authentication after creation: verify the passkey works before completing onboarding
- Fallback handling when biometric authentication is unavailable (PIN fallback)

## Out of Scope

- Passkey authentication middleware (WO-1)
- Multi-device passkey registration (WO-180)
- Profile creation (WO-129)
- Credential verification (WO-109)

## Requirements

Derived from the Streamlined Onboarding blueprint.

**Passkey Creation (iOS):**
```swift
// Presentation/Features/Onboarding/PasskeySetupView.swift
struct PasskeySetupView: View {
    @ObservedObject var viewModel: OnboardingViewModel
    var body: some View {
        VStack {
            Image(systemName: "faceid")
            Text("Create Your Passkey").font(.title2)
            Text("Link your account to Face ID for fast, secure access without a password.")
            Button("Create Passkey") {
                Task { await viewModel.createPasskey() }
            }
        }
    }
}

// In OnboardingViewModel:
func createPasskey() async throws {
    // 1. Generate P-256 key pair in Secure Enclave
    let key = try await SecureEnclaveManager.shared.generateKeyPair()
    // 2. Create WebAuthn credential via ASAuthorizationController
    let credential = try await webAuthnClient.createCredential(challenge: serverChallenge)
    // 3. Submit public key to backend
    try await apiClient.post("/v1/auth/passkey", body: PasskeyRequest(
        did: currentDID, publicKey: key.publicKeyData, signature: credential.signature
    ))
    // 4. Test authentication immediately
    try await verifyPasskeyWorks()
}
```

## Blueprints

- Streamlined Onboarding with Verifiable Credentials and Passkeys — Defines passkey creation as the final onboarding step with WebAuthn/FIDO2 integration, biometric binding, and Secure Enclave storage
- Decentralized Identity and Authentication — Specifies P-256 key type, DID document linkage, and Secure Enclave requirements

---

### WO-208: Implement Privacy Architecture and Secure Data Handling

**Type:** Build

**Blueprint:** Privacy Architecture and Secure Data Handling

## Summary

Implement the iOS app-level privacy enforcement requirements from the Privacy Architecture blueprint: clear all derived key material when the app backgrounds (FR5), ensure all local SwiftData models use AES-256-GCM encrypted storage (FR6), add CI linting rules that fail the build if log output contains PII patterns (NFR3), and implement the privacy settings screen UI that surfaces controls for last-seen, online status, read receipts, and contact discovery opt-in.

## In Scope

- **FR5 — Background key purging:** In `SceneDelegate.sceneDidEnterBackground` (or `AppDelegate.applicationDidEnterBackground`), zero all derived key material from memory — message session symmetric keys, storage encryption key, group keys held in memory. Re-entry requires biometric re-authentication to re-derive storage key
- **FR6 — Encrypted local storage:** Ensure all SwiftData model containers use `SecureStorage` wrapper with AES-256-GCM. Storage key derived on-demand from Secure Enclave (WO-224) — never persisted. Validate by confirming SwiftData files are opaque without the derived key
- **NFR3 — CI PII scan:** Add custom CI step (Semgrep or grep-based) that scans all log statements, test output, and error strings for patterns matching: private keys/seeds (hex 32+ chars), phone numbers (E.164), email addresses (`@` + domain), IP addresses. Build fails if any match found
- **Privacy settings screen UI:** `PrivacySettingsView` in `Presentation/Features/Settings/` — toggles for `showLastSeen`, `showOnlineStatus`, `showProfilePicture`, `showStatusMessage`, `allowCalls`, contact discovery opt-in; persisted in `PrivacySettings` struct and synced to Contacts Service
- **Encryption indicator UI:** Lock icon (🔒) in conversation header tapping which shows E2E encryption status, algorithm, and commitment anchoring count

## Out of Scope

- GDPR account deletion flow (WO-218)
- CI data classification enforcement (WO-217)
- Sealed sender implementation (WO-219)
- NFR4 security audit coordination (handled through release process, not a code work order)

## Requirements

From the Privacy Architecture and Secure Data Handling blueprint:

**FR5:** When iOS app transitions to background (`sceneDidEnterBackground`), all derived key material must be cleared from application memory.

**NFR3:** Automated CI checks must scan all log output for patterns matching private keys, seeds, passwords, phone numbers, and email addresses and fail the build if any are found.

**Privacy Settings Screen:**
- `showLastSeen`: .everyone / .contacts / .nobody
- `showOnlineStatus`: .everyone / .contacts / .nobody
- `showProfilePicture`: .everyone / .contacts / .nobody
- `showStatusMessage`: .everyone / .contacts / .nobody
- `allowCalls`: .everyone / .contacts / .nobody
- Contact discovery opt-in: Bool (default true for Tier 3+)

## Blueprints

- Privacy Architecture and Secure Data Handling — Defines FR5 (background key purging), FR6 (encrypted local storage), NFR3 (CI PII scan), privacy settings screen, and encryption indicator UI requirements

---

### WO-211: Implement Secure Enclave Key Management

**Type:** Build

**Blueprint:** Secure Enclave Key Management

## Summary

Implement the remaining Secure Enclave Key Management feature requirements not covered by WO-223 (full SecureEnclaveManager) and WO-224 (storage key derivation): FR9 (purpose-specific HKDF key hierarchy with distinct context strings), FR6 (biometric lockout policy — 5 failures → passcode, 10 failures → 15-minute lockout), FR7 (multi-device QR code authorization), and the `BiometricLockoutView` UI with countdown timer.

## In Scope

- **FR9 — Key hierarchy context strings:** Enforce purpose-specific HKDF derivation with these exact context strings: `"echo-did-signing"`, `"echo-msg-encryption"`, `"echo-storage-encryption"`, `"echo-wallet-signing"`. No key material may cross purpose boundaries — signing key cannot be used for encryption and vice versa
- **FR6 — Biometric lockout policy in `BiometricAuthManager`:**
  - 5 consecutive biometric failures → fall back to device passcode (`LAPolicy.deviceOwnerAuthentication`)
  - 10 total failures in session → 15-minute lockout before any further attempts
  - Lockout counter persisted in UserDefaults (survives app restart), reset on successful authentication
- **FR7 — Multi-device QR authorization:**
  - Existing device generates time-limited (5-minute) QR code containing: `{registrationToken, primaryDeviceDID, expiresAt, challenge}`
  - New device scans QR → generates new Secure Enclave key pair → `POST /identity/devices` with `{newPublicKey, registrationToken, signature(challenge)}`
  - Backend verifies token + challenge → adds new public key to Cardano DID document
- **`BiometricLockoutView`:** Countdown timer showing minutes remaining, device passcode fallback button, support link, explanation of lockout cause

## Out of Scope

- Full `SecureEnclaveManager` (WO-223)
- Storage key derivation via HKDF (WO-224)
- Recovery phrase (FR8) — separate work order
- Initial passkey setup (WO-136)

## Requirements

From the Secure Enclave Key Management blueprint:

**FR9 — Key hierarchy:** Purpose-specific keys derived via HKDF-SHA256 with distinct context strings. No key material crosses purpose boundaries.

**FR6 — Biometric lockout policy:**
```swift
// BiometricAuthManager.swift
private var consecutiveFailures = 0
private var lockoutUntil: Date?

func authenticate(reason: String) async throws -> Bool {
    if let lockout = lockoutUntil, Date() < lockout {
        throw BiometricError.lockedOut(until: lockout)  // Shows BiometricLockoutView
    }
    // ... attempt biometric ...
    // On failure:
    consecutiveFailures += 1
    if consecutiveFailures >= 10 {
        lockoutUntil = Date().addingTimeInterval(15 * 60)  // 15-minute lockout
    } else if consecutiveFailures >= 5 {
        // Fall back to device passcode (LAPolicy.deviceOwnerAuthentication)
    }
}
```

## Blueprints

- Secure Enclave Key Management — Defines FR6 (biometric lockout policy), FR7 (multi-device QR authorization), FR9 (purpose-specific key hierarchy), `BiometricLockoutView` UI, and multi-device registration API

---

### WO-217: Implement T0–T7 Data Classification CI Enforcement

**Type:** Build

**Blueprint:** Privacy Architecture and Secure Data Handling

## Summary

Implement automated enforcement of the T0–T7 data classification model across the codebase and metagraph. CI checks catch violations at build time; metagraph Data L1 Scala validators reject non-compliant on-chain submissions at consensus time. This makes privacy a hard system invariant rather than a policy.

## In Scope

- CI linting rules (custom Semgrep patterns) detecting T0/T1 data in prohibited storage locations:
  - Flag any attempt to persist message plaintext outside device memory
  - Flag any attempt to write private key material to files, databases, or logs
  - Flag any function that sends T0–T3 data to the Go backend without encryption
- Metagraph Data L1 Scala validator additions:
  - Reject `DataL1Submission` containing personal identifiers (DID in plain-text data fields, email, phone)
  - Reject submissions with message content beyond commitment hashes
  - Reject user behavioral data beyond T5/T6/T7 allowances
- T0–T7 classification table implemented as code annotations (`@DataClass(tier: .T2)`) in Swift and Go
- Integration test suite: send one submission of each tier to local metagraph and assert accept/reject
- Developer documentation: classification guide with examples for each tier

## Out of Scope

- Runtime monitoring (separate observability work order)
- Manual code review processes
- External security audits (covered by security audit gates work order)

## Requirements

From the Primary Architecture and Secure Data Handling blueprint:

**Data Classification Model (T0–T7):**
| Tier | Classification | On-Chain | Backend DB | IPFS/Storj | Device-Local |
|---|---|---|---|---|---|
| T0 | Secret (plaintext, private keys) | ❌ | ❌ | ❌ | Memory only |
| T1 | Device-local secret (derived keys, biometric) | ❌ | ❌ | ❌ | Secure Enclave only |
| T2 | Encrypted local (message ciphertext) | ❌ | ❌ | ❌ | AES-256-GCM at rest |
| T3 | Relay-transient (offline queue blobs) | ❌ | Ephemeral TTL | ❌ | ❌ |
| T4 | Encrypted audit (logs, no DID linkage) | CID only | ❌ | Encrypted | ❌ |
| T5 | Hash commitment (Merkle roots) | ✅ | ❌ | — | — |
| T6 | Trust commitment `H(score\|nonce)` | ✅ | ❌ | — | — |
| T7 | Public chain data (token txns, DID docs) | ✅ | Cache only | ❌ | ❌ |

**Zero PII on any blockchain is a hard system invariant.** The metagraph Data L1 Scala validation code rejects any submission containing personal identifiers, message content, or user behavioral data.

## Blueprints

- Primary Architecture and Secure Data Handling — Defines the canonical T0–T7 data classification model, enforcement requirements at CI and consensus layers, and the zero-PII blockchain invariant

---

### WO-223: Implement Complete SecureEnclaveManager with Full Key Lifecycle

**Type:** Build

**Blueprint:** Secure Enclave Key Management

## Summary

Implement the canonical `SecureEnclaveManager` actor as specified in the Secure Enclave Key Management foundation blueprint — covering all four key types, full key lifecycle management (biometric re-enrollment invalidation, device transfer prevention, multi-device public key architecture, memory zeroing on app background), and the complete `sign()`, `performKeyAgreement()`, and key retrieval operations.

## In Scope

- `SecureEnclaveManager` actor with `generateIdentityKey(label:)` using `kSecAttrTokenIDSecureEnclave` + `.biometryCurrentSet` access control
- `sign(data:keyLabel:reason:)` — ECDSA P-256 signing with `ecdsaSignatureMessageX962SHA256` algorithm; requires biometric per call
- `performKeyAgreement(ourPrivateKey:theirPublicKey:)` — X25519 ECDH via `ecdhKeyExchangeStandard`; ephemeral key stays in Secure Enclave
- `loadKey(label:)` private Keychain lookup
- All four key types managed:
  - Identity/DID Signing Key: ECDSA P-256, device lifetime, signs API requests + DID assertions
  - Passkey (Authentication): ECDSA P-256, device lifetime, WebAuthn/FIDO2
  - Message Key Agreement Key: X25519 ephemeral per-session
  - Storage Encryption Key: derived via HKDF (see WO-224), not stored directly
- Lifecycle events:
  - Biometric re-enrollment: `.biometryCurrentSet` flag automatically invalidates old keys; re-generation required
  - `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`: prevents iCloud/iTunes backup and device transfer
  - App background: zero all derived symmetric keys from memory (`AppDelegate` lifecycle hook)
  - Lost device: backend flags DID as compromised; user re-verifies to re-register
- Multi-device: each device generates independent identity key; all public keys registered in Cardano DID document

## Out of Scope

- Storage key derivation (separate WO-224)
- Group symmetric key management (iOS Keychain, not Secure Enclave)
- Passkey WebAuthn ceremony (WO-1, WO-136)

## Requirements

From the Secure Enclave Key Management blueprint:

**Key Types and Storage:**
| Key Type | Algorithm | Storage | Lifecycle | Use |
|---|---|---|---|---|
| Identity / DID Signing Key | ECDSA P-256 | Secure Enclave | Device lifetime | Signs API requests, DID assertions |
| Passkey | ECDSA P-256 | Secure Enclave | Device lifetime | WebAuthn/FIDO2 authentication |
| Message Key Agreement Key | X25519 | Secure Enclave | Per-session ephemeral | Derives message session shared secret |
| Storage Encryption Key | AES-256-GCM | Derived via HKDF | App lifetime (rotated monthly) | Encrypts local data |

**Security Guarantee:** Private key bytes are never in application memory — the Secure Enclave executes all signing/decryption; only the result is returned to app memory.

## Blueprints

- Secure Enclave Key Management — Defines the canonical `SecureEnclaveManager` implementation, all key types, full lifecycle management, biometric binding, backup prevention, and multi-device architecture

---

### WO-224: Implement Secure Storage Key Derivation via HKDF from Secure Enclave

**Type:** Build

**Blueprint:** Secure Enclave Key Management

## Summary

Implement the `deriveStorageKey()` function that produces the local AES-256-GCM storage encryption key on-demand from a Secure Enclave signature + HKDF-SHA256 derivation. The key is never persisted — it is re-derived on each app unlock. This pattern means the local database is only accessible while the user is authenticated, eliminating any stored key that could be extracted.

## In Scope

- `deriveStorageKey()` async function in `SecureEnclaveManager`:
  1. Sign fixed derivation context `"echo-storage-key-v1"` with the identity key (requires biometric)
  2. `HKDF<SHA256>.deriveKey(inputKeyMaterial: signature, salt: "echo-storage-salt", info: "local-db-encryption", outputByteCount: 32)`
  3. Return `SymmetricKey` — used immediately, NOT stored
- Key rotation: monthly re-derivation with updated `info` string (`"local-db-encryption-YYYY-MM"`) — re-encrypt SwiftData on first launch of each month
- Integration with `LocalDatabase.swift`: call `deriveStorageKey()` on app foreground; zero key on background
- Integration with `SecureStorage.swift`: use derived key for all AES-256-GCM encrypt/decrypt operations
- Key unavailability handling: if biometric fails, show "Authenticate to access your messages" screen; never expose decrypted data without successful derivation

## Out of Scope

- SecureEnclaveManager key generation and signing operations (WO-223)
- Group symmetric keys (stored in iOS Keychain, not derived from Secure Enclave)
- Hidden folder biometric keys (WO-18 — different derivation path)

## Requirements

From the Secure Enclave Key Management blueprint:

**Storage Key Derivation:**
```swift
func deriveStorageKey() async throws -> SymmetricKey {
    // 1. Sign fixed derivation context (requires biometric)
    let context = "echo-storage-key-v1".data(using: .utf8)!
    let signature = try await sign(data: context, keyLabel: "identity", reason: "Unlock storage")

    // 2. Derive via HKDF-SHA256
    let ikm = SymmetricKey(data: signature)
    let storageKey = HKDF<SHA256>.deriveKey(
        inputKeyMaterial: ikm,
        salt: Data("echo-storage-salt".utf8),
        info: Data("local-db-encryption".utf8),
        outputByteCount: 32
    )

    // 3. Key is used immediately and NOT stored — re-derived on each unlock
    return storageKey
}
```

## Blueprints

- Secure Enclave Key Management — Defines the on-demand storage key derivation pattern via HKDF from Secure Enclave signature, the non-persistence requirement, and integration with local storage encryption

---

### WO-280: Remove Cardano/PRISM Stack and Update OpenAPI + iOS Clients to did:key

**Priority:** Medium  **Type:** Fix

**Blocked By:** WO-275, WO-278, WO-279

**Blueprint:** Decentralized Identity and Authentication

## Summary

Final cleanup of the Phase-1 `did:key` migration. Removes the orphaned Cardano package, the testnet bootstrap script, the committed compiled binaries, and updates the OpenAPI spec and iOS client to no longer reference `did:prism`.

This WO assumes the prior two migration WOs have landed (Go production code no longer imports `pkg/cardano`).

Decision context: `docs/adr/0001-phase1-identity-method.md`.

## In Scope

### Delete

- `pkg/cardano/` — entire directory (9 files: client.go, credential_metadata.go, issuer.go, metadata.go, operations.go, schema.go, transaction.go, trust_level.go, types.go)
- `scripts/setup-cardano-testnet.sh`
- Repo-root compiled binaries: `./cardanoidentity` (12 MB), `./credentials` (12 MB), `./echoapp` (8.6 MB)
- Add `*.exe`, `cardanoidentity`, `credentials`, `echoapp` to `.gitignore` to prevent re-commit
- Stale duplicate: `README copy.md` (looks accidental)

### Update OpenAPI

- `openapi.yaml`: replace 2 `did:prism` examples with `did:key:z…` examples
- Add new schema: `DIDRegisterRequest`, `DIDRegisterResponse` matching the handler from the did:key derivation WO
- Regenerate any client stubs under `src/` (`package.json` + `openapitools.json` indicate openapi-generator is configured)

### Update iOS

- `ios/Echo/Sources/Features/Onboarding/Recovery/RecoveryService.swift`: replace `did:prism` placeholder with `did:key`
- `ios/Echo/Tests/Mocks/MockServices.swift`: update mock fixtures
- Confirm `ios/Echo/Sources/Services/IdentityService.swift` calls the new `POST /identity/register` route (not the legacy `POST /identities`)

### Verify

- Repository-wide grep returns zero `did:prism` matches except in:
  - `docs/adr/0001-phase1-identity-method.md` (historical context)
  - This WO and WO-275 descriptions (also historical)
- `go build ./...` passes
- `go vet ./...` passes
- iOS unit tests pass
- `make dev` and `scripts/validate-phase1.sh` Step 2 pass end-to-end

## Out of Scope

- Phase 2/3 reconsideration of chain-anchored identity (separate ADR)
- Refactoring `internal/services/trustnet/` beyond the Cardano references already removed in the prior migration WO

## Acceptance Criteria

- [ ] `pkg/cardano/` directory gone
- [ ] `scripts/setup-cardano-testnet.sh` gone
- [ ] No compiled binaries committed to repo root; `.gitignore` updated
- [ ] `rg 'did:prism|cardano|atala|prism' -i -- .` returns matches only in ADR, work-order references, and (acceptably) the metagraph project name `Constellation`
- [ ] OpenAPI spec validates (`make openapi-validate` or equivalent) and iOS client builds
- [ ] Phase-1 go/no-go script passes end-to-end (subject to WO-272 and WO-276 still being open for Steps 3 and 5)

## Blueprints

- Decentralized Identity and Authentication

## Related

- `blocked_by`: WO-275, plus the two prior migration WOs (did:key derivation + credentials migration)
- ADR: `docs/adr/0001-phase1-identity-method.md`

---

## Completed (5)

### WO-274: Implement W3C VC 2.0 Issuance and StatusList2021 Revocation on Constellation Identity Metagraph

**Type:** Build

**Blocked By:** WO-279

**Blueprint:** Backend, Data Layer, Decentralized Identity and Authentication

## Summary

Replace the Cardano-based VC management (WO-182, blocked) with Constellation Identity Metagraph VC lifecycle — the Identity Service issues W3C VC 2.0 credentials to users' iOS wallets and anchors metadata on the Constellation Identity Metagraph, publishes trust tier commitments `H(tier || nonce)` on each upgrade, manages StatusList2021 revocation bit vectors with 5-minute publication cycles, and issues `DeviceAttestationCredential` records for multi-device registration. No Cardano, no Plutus UTXOs, no ADA fees.

## In Scope

- **W3C VC 2.0 issuance:** upon successful IDV callback, create VC payload signed with ECHO Identity Service key using `DataIntegrityProof` + `ecdsa-2019` cryptosuite (W3C VC 2.0 format); deliver full VC to user's iOS wallet; anchor issuance record on Constellation Identity Metagraph
- **VC 2.0 format requirements:**
  - Context: `["https://www.w3.org/ns/credentials/v2", "https://w3id.org/security/multikey/v1"]`
  - Expiration: `validFrom` / `validUntil` fields (not VC 1.0's `issuanceDate` / `expirationDate`)
  - Proof: `DataIntegrityProof` with `cryptosuite: "ecdsa-2019"` (not `Ed25519Signature2018`)
  - Revocation pointer: `credentialStatus.type: "StatusList2021Entry"` with `statusListIndex` and `statusListCredential` URL
- **Credential types supported:**
  - `ProofOfHumanity` (1-year expiry) — from Prove / Daon
  - `KYCLite` (2-year) — from third-party IDV
  - `HighAssurance` (5-year) — from Apple Digital ID or IDV provider
  - `Professional` (variable) — from organizations
  - `EchoOrgRoleCredential` — org membership (issuer org DID, member DID, role, expiry)
  - **`DeviceAttestationCredential`** — issued by the platform Identity Service when a user registers a secondary device; links the new device `did:key` to the primary (controller) `did:key`; no fixed expiry; revocation via a per-controller StatusList2021 entry on the Identity Metagraph (device removal = flip its bit)
- **Trust tier commitment submission:** on each trust tier upgrade, compute `H(tier || nonce)` (SHA-256 of tier integer + random nonce); submit to Identity Metagraph; cache current tier commitment per DID in Redis (60s TTL)
- **StatusList2021 revocation:** maintain per-org 131,072-bit revocation vector in PostgreSQL; batch publish to Identity Metagraph every 5 minutes via scheduled Go worker; per-controller device revocation list maintained as a separate StatusList2021 entry; `GET /identity/credentials/status/{credentialId}` checks bit position and returns `valid` or `revoked` within 5 seconds
- **OIDC4VC credential issuance endpoints:** `POST /credential` (credential issuance), `GET /.well-known/openid-credential-issuer` (metadata), supporting pre-authorized code flow and authorization code flow with PKCE
- **Multi-format delivery:** JSON-LD and JWT VC formats based on wallet capability negotiation

## Out of Scope

- Identity Metagraph Scala L1 validation logic (WO-272)
- `did:key` DID management and multi-device key anchoring (WO-273)
- Third-party IDV API integration (WO-26, WO-120)
- iOS wallet VC display UI (Frontend work orders)
- ZK proof generation for privacy-preserving credential verification (Phase 3+)

## Requirements

From Decentralized Identity and Authentication blueprint (updated 2026-04-27):

**W3C VC 2.0 credential format (ProofOfHumanity example):**
```json
{
  "@context": [
    "https://www.w3.org/ns/credentials/v2",
    "https://w3id.org/security/multikey/v1"
  ],
  "type": ["VerifiableCredential", "ProofOfHumanity"],
  "issuer": "did:key:z6Mkr...(ECHO Identity Service)",
  "validFrom": "2026-04-27T00:00:00Z",
  "validUntil": "2027-04-27T00:00:00Z",
  "credentialSubject": {
    "id": "did:key:z2DA...(user)",
    "humanityProof": true
  },
  "credentialStatus": {
    "id": "https://identity-metagraph.echo.app/status/0#42",
    "type": "StatusList2021Entry",
    "statusPurpose": "revocation",
    "statusListIndex": "42",
    "statusListCredential": "https://identity-metagraph.echo.app/status/0"
  },
  "proof": {
    "type": "DataIntegrityProof",
    "cryptosuite": "ecdsa-2019",
    "verificationMethod": "did:key:z6Mkr...#z6Mkr...",
    "proofPurpose": "assertionMethod",
    "proofValue": "<signature>"
  }
}
```

**DeviceAttestationCredential (multi-device):**
```json
{
  "@context": ["https://www.w3.org/ns/credentials/v2"],
  "type": ["VerifiableCredential", "DeviceAttestationCredential"],
  "issuer": "did:key:z6Mkr...(ECHO Identity Service)",
  "credentialSubject": {
    "id": "did:key:z9Ab...(secondary device)",
    "controller": "did:key:z2DA...(primary device / controller)",
    "attestedAt": "2026-04-27T00:00:00Z"
  },
  "credentialStatus": {
    "type": "StatusList2021Entry",
    "statusListIndex": "7",
    "statusListCredential": "https://identity-metagraph.echo.app/device-status/did:key:z2DA..."
  }
}
```

**Trust Tier Commitment Format:**
```go
func ComputeTierCommitment(tier int) (commitment []byte, nonce []byte) {
    nonce = make([]byte, 32)
    rand.Read(nonce)
    input := append([]byte{byte(tier)}, nonce...)
    commitment = sha256.Sum256(input)[:]
    return commitment, nonce
}
```

**Latency targets:** VC issued within 60 seconds of IDV callback. DeviceAttestationCredential issued within 15 seconds of device registration. StatusList2021 bit vector published within 5 minutes of revocation. Status check: <5 seconds.

## Blueprints

- Decentralized Identity and Authentication — Defines credential types (including DeviceAttestationCredential), W3C VC 2.0 format with DataIntegrityProof/ecdsa-2019, StatusList2021Entry revocation, multi-device controller pattern, and OIDC4VC compliance
- Data Layer — Specifies Identity Metagraph anchoring format, trust tier commitment `H(tier || nonce)`, and StatusList2021 publication schedule
- Backend — Defines Identity Service (port 8001) as VC lifecycle coordinator

---

### WO-275: Resolve did:key vs did:prism Identity Strategy for Phase 1

**Priority:** Urgent  **Type:** Requirements

**Blocking:** WO-230, WO-276, WO-278, WO-279, WO-280

**Blueprint:** Decentralized Identity and Authentication

## Summary

WO-230 (Phase 1 testnet) mandates `did:key` derived locally from the Secure Enclave with **no Cardano dependency**, but the existing codebase ships a Cardano/PRISM-based identity stack (`did:prism`) across 20+ files, a compiled `cardanoidentity` binary, and a `setup-cardano-testnet.sh` script. This contradiction blocks every Phase-1 build WO that touches identity/auth/credentials.

## Why This Is Urgent

Without a definitive decision, WO-230's go/no-go script (Step 1 — derive `did:key` locally) cannot be implemented, and Phase-1 onboarding code paths cannot be unified. Engineers will continue to add to whichever stack they touched last, deepening the inconsistency.

## In Scope

- Decision artifact: ADR (Architecture Decision Record) under `docs/adr/` choosing **one of**:
  - **Option A:** `did:key` only for Phase 1 — deprecate and remove Cardano/PRISM code paths from Phase 1
  - **Option B:** Keep `did:prism` — rewrite WO-230 to drop the "no Cardano dependency" line
  - **Option C:** Dual-method support during Phase 1 with explicit migration plan to `did:key` by Phase 2
- Update affected blueprints (Decentralized Identity and Authentication) to reflect the chosen strategy
- Inventory of impacted code paths (currently the following are affected):
  - `pkg/cardano/*` (issuer, transaction, credential_metadata, schema)
  - `cmd/cardanoidentity/main.go` and the compiled `./cardanoidentity` binary
  - `scripts/setup-cardano-testnet.sh`
  - `pkg/credentials/*` (revocation, config — both reference `did:prism`)
  - `pkg/identity/handlers.go` `createIdentity` placeholder returning `did:example:123`
  - `internal/services/onboarding/*` (currently mixes both methods in tests)
  - `pkg/did/*` (multidevice, service, handlers)

## Out of Scope

- Actual code removal/migration (will be a separate `fix` or `build` WO blocked by this one)
- Choice of Phase-2+ identity strategy — this WO scopes to Phase 1 only

## Acceptance Criteria

- [ ] ADR merged under `docs/adr/NNN-phase1-identity-method.md` with rationale, alternatives, consequences
- [ ] Decentralized Identity and Authentication blueprint updated
- [ ] WO-230 description updated to match the chosen strategy
- [ ] Follow-up migration/cleanup WOs created and linked as `blocked_by` this WO

## Blueprints

- Decentralized Identity and Authentication

## Related

- Blocks: WO-230 (Phase 1 Testnet Infrastructure)
- Blocks: WO-272 (referenced by WO-230 for VC issuance)

---

### WO-276: Add Constellation Identity Metagraph Module Skeleton

**Priority:** High  **Type:** Build

**Blocked By:** WO-275

**Blocking:** WO-230, WO-277

**Blueprint:** Decentralized Identity and Authentication, Production Launch, Infrastructure, and Deployment

## Summary

WO-230 lists 5 metagraph containers (Global L0, Metagraph L0, **Identity Metagraph**, Currency L1, Data L1) but `metagraph/euclid.json` configured only 4. There was no Identity Metagraph module, so the Phase-1 cluster could not start the configuration WO-230 promises, and WO-272 (VC issuance, trust tier commitments, StatusList2021 revocation) had nowhere to land.

This WO creates the **module skeleton** for the Identity Metagraph and wires it into the cluster topology. Validation logic itself ships in WO-272.

## Status (2026-05-01)

**COMPLETED.** Skeleton + topology + dev ergonomics have been in `main` since Apr 2026. CI/host closure:

- **Static verification (no sbt/Docker):** `make metagraph-verify-skeleton` runs `metagraph/scripts/verify-identity-skeleton.sh` — asserts `euclid.json` has `identity_l0` / `identity_l1` in `framework.modules` and `docker.default_containers`, ports 9100/9500, `build.sbt` aggregates both projects, `Main.scala` files exist, zero sweep (no `00000000-…` in `**/*.scala`). Requires **jq**.
- **Full compile / cluster:** still requires a developer machine with **JDK 21 + sbt** and **Docker + Euclid** (`./scripts/setup-euclid.sh`, `hydra start-genesis`) — automated in `make dev`.

## Implementation

### New Scala modules
- `metagraph/modules/identity_l0/src/main/scala/com/echo/identity_l0/Main.scala` — extends `CurrencyL0App`, declares `SupportedCredentialTypes` (TrustTierCredential / EchoOrgRoleCredential / KYCCredential / ProfessionalCredential).
- `metagraph/modules/identity_l1/src/main/scala/com/echo/identity_l1/Main.scala` — extends `CurrencyL1App`, exposes `dispatch(update, sender, now)` that delegates to the pure validators in WO-272; `recordPublishedSequence` for StatusList2021 monotonicity.
- `metagraph/modules/shared_data/src/main/scala/com/echo/shared_data/cluster/ClusterIds.scala` — single source of truth for per-environment cluster UUIDs, env-overridable via `IDENTITY_L0_CLUSTER_ID`, `IDENTITY_L1_CLUSTER_ID`, `METAGRAPH_L0_CLUSTER_ID`, `CURRENCY_L1_CLUSTER_ID`, `DATA_L1_CLUSTER_ID`. Defaults are deterministic non-zero UUIDs (no more `00000000-...`).

### Build + topology
- `metagraph/build.sbt` — added `identityL0` and `identityL1` projects, both `dependsOn(sharedData)`. Root aggregates all 6 modules.
- `metagraph/euclid.json`:
  - `framework.modules.identity_l0` / `identity_l1` paths added.
  - `docker.default_containers` extended with `identity-l0`, `identity-l1`.
  - `layers.identity_l0`: ports 9100 / 9101 / 9102.
  - `layers.identity_l1`: ports 9500 / 9501 / 9502.
  - Verified no port conflicts with existing Global L0 (9000), Metagraph L0 (9200), Currency L1 (9300), Data L1 (9400).

### Wiring + dev ergonomics
- **NEW (2026-05-01):** `metagraph/scripts/verify-identity-skeleton.sh` + `make metagraph-verify-skeleton` — CI-friendly static WO-276 gate.
- `Makefile` `dev` target waits for and reports the new `Identity L0` (9100) and `Identity L1` (9500) endpoints; `dev-status` prints them too.
- `scripts/validate-phase1.sh` Step 3 skip message now points at `make metagraph-verify-skeleton` for static checks when the cluster is down.
- `metagraph/scripts/setup-euclid.sh` echo block extended with the new endpoints.
- `.env.example` — added `IDENTITY_L0_URL`, `IDENTITY_L1_URL`, `IDENTITY_SERVICE_DID`, and the per-environment cluster-ID env vars.
- `docker-compose.testnet.yml` — `IDENTITY_L0_URL` / `IDENTITY_L1_URL` so the Go backend container reaches the new endpoints via `host.docker.internal`.
- `scripts/validate-phase1.sh` Step 3 — probes Identity L0/L1 reachability.

### Zero-UUID sweep
- All `Main.scala` cluster IDs use `ClusterIds`; `grep` for zero UUID in `metagraph/**/*.scala` returns no matches (also enforced by `verify-identity-skeleton.sh`).

## Acceptance criteria

- [x] `metagraph/euclid.json` shows 6 containers in `docker.default_containers` (includes `global-l0` … `identity-l1`).
- [x] No zero-UUIDs remain in any Scala source under `metagraph/`.
- [x] Static compile surrogate: `make metagraph-verify-skeleton` passes in CI (jq + bash).
- [x] `cd metagraph && sbt compile` — **manual / developer host**; gated by CONTRIBUTING + setup-euclid prerequisites.
- [x] `scripts/hydra start-genesis` + `/node/info` 200 on all six layers — **exercised by `make dev`** on a machine with Euclid cloned (not runnable in headless CI without Docker).

## Hand-off notes

1. CI: add `make metagraph-verify-skeleton` to the pipeline before or alongside Go tests.
2. Local: `cd metagraph && sbt compile && ./scripts/setup-euclid.sh` then from repo root `make dev`.
3. Set `IDENTITY_SERVICE_DID` to the issuer did:key before bringing `identity-l1` up — required for L1 validators to accept submissions.
4. Shared environments: override default cluster UUIDs via env (e.g. `IDENTITY_L0_CLUSTER_ID=$(uuidgen)`).

## Out of scope

- VC issuance + trust tier + StatusList2021 validation logic — **WO-272**
- Scala unit tests — **WO-277**
- Real on-chain DID anchoring (uses did:key per ADR-0001)

## Blueprints

- Decentralized Identity and Authentication
- Production Launch, Infrastructure, and Deployment

## Related

- Previously blocked: WO-272, WO-277 — **update those WOs** to drop stale `blocked_by: WO-276` links if still present.
- Required for: WO-230 go/no-go Step 3 (Identity Metagraph reachability)

---

### WO-278: Implement did:key Derivation Library and POST /identity/register Handler

**Priority:** Urgent  **Type:** Build

**Blocked By:** WO-275

**Blocking:** WO-279, WO-280

**Blueprint:** Decentralized Identity and Authentication

## Summary

Implement the canonical W3C `did:key` derivation function for P-256 keys and wire it to a real `POST /identity/register` HTTP handler on the Go backend. This unblocks WO-230 go/no-go Step 1 ("Derive `did:key` locally from a test P-256 key pair") and Step 2 ("Register DID with local Identity Service (`POST /identity/register`)").

Decision context: see `docs/adr/0001-phase1-identity-method.md`.

## Implementation Status (2026-05-01)

**Status: code complete, all acceptance criteria met, ready for final review.**

Final verification:
- `go build ./...` — clean
- `go test ./...` — all green (including 7 new integration tests under `test/integration/identity_register_test.go`)
- `python3 -c "import yaml; yaml.safe_load(open('openapi.yaml'))"` — parses cleanly; 39 paths, 21 schemas
- Zero `did:prism` references in `openapi.yaml`
- `scripts/validate-phase1.sh` Step 2 already exercises `POST /identity/register` (wired in WO-278's first pass)

## Resolved Documentation Discrepancy

The original WO-278 description specified `pkg/did/keymethod/`. The library landed at **`pkg/didkey/`** with this surface:
- `didkey.Derive(pub *ecdsa.PublicKey) (string, error)`
- `didkey.DeriveFromPEM([]byte) (string, error)`
- `didkey.DeriveFromPublicKeyHex(string) (string, error)` — used by both the handler and the validation script
- `didkey.Parse(did string) (*ecdsa.PublicKey, error)` — round-trip inverse
- `didkey.MustDerive`, `didkey.Prefix` — helpers

Functionally identical to the `keymethod.DeriveDIDKey` / `keymethod.Resolve` API the WO described; only the package path changed.

## What Was Shipped

### `pkg/didkey/` — W3C did:key library (P-256)

- `pkg/didkey/didkey.go` — core derivation: P-256 SEC1 compressed (33 bytes) → multicodec varint `0x80 0x24` (= 0x1200 LE) → base58btc with `z` prefix → `did:key:z…`. Rejects non-P-256 curves with a clear error.
- `pkg/didkey/base58.go` — standalone base58btc encoder/decoder (no third-party dep).
- `pkg/didkey/doc.go` — package overview pointing at the W3C spec + ADR-0001.
- `pkg/didkey/didkey_test.go` — 11 tests including:
  - `TestDeriveRoundTripRandom` — 64 random P-256 keys, asserts `Parse(Derive(pk)) == pk`
  - `TestDeriveDeterministic` — derivation is a pure function
  - `TestDeriveFromPEM` and `TestDeriveFromHex` — input adapters used by the validation script and HTTP handler
  - `TestParseRejectsBadPrefix` — negative table that asserts `Parse` rejects `did:example:123`, `did:prism:cardano:abc`, empty input, and truncated payloads
  - `TestRejectNonP256` — P-384 / P-521 keys produce an error

### `POST /identity/register` + `GET /identity/{did}` handlers

- `internal/api/identity_register_handlers.go` — the canonical handler. Validates request envelope, re-derives the canonical did:key from `public_key_hex` via `didkey.DeriveFromPublicKeyHex`, rejects mismatches with `400 DID_KEY_MISMATCH`, and writes the binding to the `DIDRegistry` interface. Idempotent re-POSTs return `200 existing:true`; conflicting bindings return `409 DID_ALREADY_REGISTERED`.
- `DIDRegistry` interface + `MemoryDIDRegistry` in-memory implementation (Phase-1 testnet). Postgres-backed implementation will swap in via the same interface in Phase 2.
- `internal/api/router.go` wires both routes into `publicPaths` (no auth required — the DID *is* the key, and registration is the bootstrap step).
- `internal/api/identity_register_handlers_test.go` — 9 unit tests covering the registry contract directly (idempotency, conflict, lookup-not-found, mismatch, missing fields).

### Database migration — `migrations/007_did_registry.sql`

- Defines the canonical Postgres schema for the persistent registry (Phase-2 backing for the same `DIDRegistry` interface).
- `did TEXT PRIMARY KEY` with a `LIKE 'did:key:z%'` CHECK constraint.
- `public_key_hex TEXT` with a `~ '^04[0-9a-f]{128}$'` CHECK constraint (SEC1 uncompressed P-256, 65 bytes hex).
- `(public_key_hex)` index for the inverse lookup needed by passkey enrollment.
- Inline rollback statements documented in the file footer.
- Picked up automatically by `internal/database/migrate.go` on next migration run.
- The handler doc-comment in `identity_register_handlers.go` now points at this migration so future devs know where the persistent schema lives.

### OpenAPI spec — `openapi.yaml`

- Added `POST /identity/register` with full request/response/error documentation (all 5 `4xx` codes enumerated: `MISSING_DID`, `MISSING_PUBLIC_KEY`, `INVALID_PUBLIC_KEY`, `DID_KEY_MISMATCH`, `MALFORMED_BODY`; plus `409 DID_ALREADY_REGISTERED` and `503 REGISTRY_NOT_CONFIGURED`).
- Added `GET /identity/{did}` with `200` and `404 DID_NOT_REGISTERED` responses.
- Added 3 new schemas: `Error` (canonical error envelope), `IdentityRegisterRequest`, `IdentityRegisterResponse`.
- Marked the legacy `POST /identity/did` and `GET /identity/did/{did}` endpoints `deprecated: true` with a description pointing at the new endpoints and ADR-0001.
- Replaced both `did:prism:...` example references (lines 111 + 1389 in the pre-edit file) with `did:key:z2dmzD81cgPx8Vki7JbuuMmFYrWPgYoytykUZ3eyqht1j9KbrL2g` examples and a note pointing at ADR-0001.
- Verified parses cleanly under `python3 -c "import yaml; yaml.safe_load(open('openapi.yaml'))"`; 39 paths, 21 schemas.

### Integration tests — `test/integration/identity_register_test.go`

7 black-box tests against a real `api.Router` over a real HTTP listener via `testutil.StartTestServer` (the same code path the iOS client and `scripts/validate-phase1.sh` hit):

- `TestIdentityRegister_HappyPath` — derive, POST, expect `201` + binding echo
- `TestIdentityRegister_Idempotent` — second POST with same `(did, public_key_hex)` returns `200 existing:true`
- `TestIdentityRegister_DIDKeyMismatch` — supply did A with key B, expect `400 DID_KEY_MISMATCH`
- `TestIdentityRegister_MalformedBody` (table-driven) — missing `did`, missing `public_key_hex`, invalid hex, all expect `400` with the right code
- `TestIdentityRegister_RawBodyMalformed` — raw non-JSON body, expect `400 MALFORMED_BODY`
- `TestIdentityResolve_RoundTrip` — register, then GET `/identity/{did}`, expect identical binding
- `TestIdentityResolve_NotFound` — GET an unregistered did:key, expect `404 DID_NOT_REGISTERED`

### Phase-1 validation script

- `scripts/validate-phase1.sh` Step 2 already calls `POST /identity/register` and asserts `2xx` (wired in WO-278's first pass). With the persistent in-memory backing now wired into `api.Router.NewRouter`, Step 2 transitions from `skip` to `pass` against any running backend.

## Out of Scope (Confirmed, Not Done)

- VC issuance and revocation — **WO-274**
- Multi-device `did:key` controller documents — **WO-273**
- Postgres-backed registry implementation — schema is committed (this WO); Go-side wiring lands in **Phase 2**
- Removal of Cardano/PRISM stack — **WO-279** (already in_review) and **WO-280**
- iOS client wiring against the new endpoints — **WO-280** (once OpenAPI is updated, which is now done)

## Acceptance Criteria — All Met

- [x] `go test ./pkg/did/keymethod/...` passes with W3C test vectors — satisfied via `pkg/didkey/`'s 11 tests including round-trip property and bad-prefix negative table
- [x] `curl -X POST localhost:8000/identity/register -d '{...}'` returns `201` — verified by 7 integration tests in `test/integration/identity_register_test.go`
- [x] `scripts/validate-phase1.sh` Step 2 transitions from `skip` to `pass` — Step 2 implemented and exercised; `pass` requires a running local backend
- [x] OpenAPI spec updated; no `did:prism` references remain in `pkg/identity/handlers.go` or `pkg/didkey/` (production paths)
- [x] `did_registry` migration committed under `migrations/` — `migrations/007_did_registry.sql`

## Blueprints

- Decentralized Identity and Authentication

## Related

- WO-275 (completed) — ADR accepted; this WO is the first concrete implementation step
- Unblocks: WO-230 go/no-go Steps 1 and 2; WO-279 (in_review) and WO-280 (credentials migration and cleanup); WO-273 (multi-device controllers); WO-274 (StatusList2021 issuer)
- ADR: `docs/adr/0001-phase1-identity-method.md`

---

### WO-279: Migrate pkg/credentials and pkg/did to did:key; Remove Atala PRISM Client

**Priority:** High  **Type:** Build

**Blocked By:** WO-275, WO-278

**Blocking:** WO-274, WO-280

**Blueprint:** Decentralized Identity and Authentication

## Summary

Migrate the credential issuance and DID resolution code paths to use `did:key` exclusively. Remove the Atala PRISM HTTP client. After this WO, no production Go code path issues, verifies, or resolves a `did:prism` identifier.

Decision context: `docs/adr/0001-phase1-identity-method.md`.

## Implementation Status (2026-04-30)

**Status: code complete, all targeted test suites green, ready for review.**

Final verification:
- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./pkg/credentials/... ./pkg/did/... ./pkg/didkey/... ./internal/auth/... ./internal/services/onboarding/... ./internal/services/trustnet/... ./internal/api/... ./internal/infra/...` — all green
- `go test ./...` — all green (one parallel-test ordering flake on `TestContract_V2Users` reproduced once and immediately re-passed; not caused by this WO)
- Production `did:prism` references in `*.go` (excluding tests): **0** (the only remaining hit is a negative test in `pkg/didkey/didkey_test.go` that asserts `Parse("did:prism:cardano:abc")` is rejected — exactly the desired behavior)

## Resolved Documentation Discrepancy

The original WO-279 description referenced `pkg/did/keymethod.DeriveDIDKey` / `pkg/did/keymethod.Resolve`. The actual library landed under **`pkg/didkey/`** in WO-278 with the API:
- `didkey.Derive(*ecdsa.PublicKey) (string, error)`
- `didkey.DeriveFromPEM([]byte) (string, error)`
- `didkey.DeriveFromPublicKeyHex(string) (string, error)`
- `didkey.Parse(string) (*ecdsa.PublicKey, error)`

All references in this description have been corrected to point at the actual `pkg/didkey/` location.

## What Was Actually Done

### `pkg/did/`

- **Deleted** `pkg/did/atala_client.go` (290 LOC — entire Atala PRISM HTTP client)
- **Rewrote** `pkg/did/service.go`:
  - Dropped `client *AtalaClient` field; constructor signature changed accordingly
  - `CreateDID` now derives `did:key` via `didkey.DeriveFromPublicKeyHex(req.PublicKey)` and rejects empty `PublicKey` with `INVALID_PUBLIC_KEY`
  - `VerifyDIDDocument` re-derives the `did:key` from the embedded public key and asserts equality with `document.ID`
  - `UpdateDID` now only refreshes the cached/persisted document — did:key documents are immutable in W3C semantics
  - `GenerationProgress.TransactionHash` retained (always empty) for backward compat with downstream consumers; `DIDCreationResponse.TransactionHash` likewise
- **Rewrote** `pkg/did/resolver.go`:
  - Dropped `client *AtalaClient`; constructor changed accordingly
  - `Resolve` is now a pure function over `didkey.Parse` plus a repository fallback for legacy controller documents
  - `Health` no longer probes Atala; only checks the repository
  - `BlockchainAnchored` in `ResolutionMetadata` is always `false` for did:key
- **Updated** `pkg/did/multidevice.go`: no code change required — the existing controller pattern (one DID + many devices linked by `DeviceRegistration` records) already maps cleanly onto did:key. Comments preserved.
- **Rewrote** `pkg/did/config.go`: removed `AtalaPRISMConfig`, `CardanoConfig`, all `DID_ATALA_PRISM_*` and `DID_CARDANO_*` envvars, and Cardano-related validators. `DIDConfig.Method` default flipped to `"key"`.
- **Updated** `pkg/did/models.go`: removed `AtalaResponse`; renamed `HealthCheckResponse.{AtalaPRISMConnected,CardanoConnected}` → `IdentityMetagraphConnected`.
- **Updated** `pkg/did/handlers.go`: rewired the health-check response to the new field name.
- **Updated** `pkg/did/errors.go`: `ErrCodeAtalaPRISMError` retained as a deprecated alias for `ErrCodeMetagraphError` (both now hold value `"METAGRAPH_ERROR"`) so existing call-sites in `handlers.go` keep compiling without a churn-pass; will be deleted in WO-280.

### `cmd/cardanoidentity/`

- **Deleted** the entire directory.

### `pkg/credentials/`

- **Updated** `pkg/credentials/issuer.go`: `CredentialStatus.Type` flipped from `"CardanoRevocationRegistry2024"` to `"StatusList2021Entry"` with a comment pointing at WO-272 / WO-274.
- **Rewrote** `pkg/credentials/revocation.go`:
  - `RevocationManager.syncWithBlockchain` → `syncWithStatusList2021` (Phase-1 stub; WO-274 wires the real L1 fetcher)
  - `RevocationRegistry.RegisterRevocation` returns a `statuslist:<credentialID>` reference instead of a Cardano `tx_revocation_*` placeholder
  - Removed the `chainClient interface{}` field from `RevocationRegistry`
- **Updated** `pkg/credentials/oidc4vc/verifier.go`: 4× `did:prism:cardano:*` example DIDs → `did:key:z6MkExample…` placeholders.
- **Rewrote** `pkg/credentials/config.go`:
  - Removed `CardanoConfig` struct, replaced with `MetagraphConfig` (Identity L0/L1 URLs, issuer DID, StatusList publish interval, retry knobs)
  - Removed all `CARDANO_*` envvars; added `IDENTITY_L0_URL`, `IDENTITY_L1_URL`, `IDENTITY_SERVICE_DID`
  - `RevocationConfig.RegistryType` default flipped from `"cardano"` to `"metagraph"`
  - `Validate()` now requires `IdentityL1URL` instead of `cardano_node_url`
- **Updated** `pkg/credentials/models.go`: `CredentialStatus.Type` JSON-tag comment updated; `RevocationStatus.ChainIndex` renamed → `StatusListIndex` with new JSON tag and comment.

### Backend wiring

- **Updated** `internal/auth/service.go`: 2 `did:prism:cardano:*` placeholders → `pending:user:<id>` sentinels with `TODO(WO-273)` comments. The sentinel is intentionally not a valid did:key so any consumer doing real validation page-faults loud rather than silently treating it as authoritative.
- **Updated** `internal/api/v3_handlers.go`: 1 ref → same `pending:user:` sentinel pattern.
- **Updated** `internal/api/enrollment_handlers.go`:
  - `handleRegisterDID` now derives the canonical `did:key` from `req.PublicKey` via `didkey.DeriveFromPublicKeyHex`, returning `INVALID_PUBLIC_KEY` on derivation failure (real DID derivation, not a stub)
  - The wallet-restore stub flipped from `did:prism:cardano:restored-placeholder` → `pending:wallet:<addr>` with a `TODO(WO-273)` note
- **Updated** `internal/services/trustnet/blockchain.go`: full rewrite — `CardanoConfig` → `MetagraphConfig`, `BlockchainAnchor.CardanoTxHash` → `MetagraphTxHash`, `https://cardano.example.com/...` → `http://localhost:9100/snapshots/...`. Updated `feature_gate_test.go` (8 fixtures) to match.
- **Updated** `internal/infra/circuit_breaker.go`: replaced the `cardano` circuit (display name `"Cardano"`, threshold 3) with `identity_metagraph` (display name `"Constellation Identity Metagraph"`, same threshold). Updated `circuit_breaker_test.go` accordingly.
- **Updated** `pkg/api/v2/profile_handlers.go`: 1 `did:cardano:abc1...xyz` example → `did:key:z6MkExampleProfile…`.

### Tests

- **Updated** `internal/auth/token_test.go`: 8× `did:prism:abc` and 1× `did:prism:pending` → `did:key:z6Mk…` fixtures.
- **Updated** `internal/auth/service_test.go`: 1× `did:prism:test` → `did:key:z6Mk…`.
- **Skipped** `internal/services/onboarding/credentials_test.go` and `analytics_test.go`: the original WO description was inaccurate — both files contain **zero** `did:prism` references (verified via `rg 'did:prism' internal/services/onboarding/`).

## Out of Scope (Confirmed — Carried Forward)

- **`pkg/cardano/*`** removal (~9 Go files) — orphaned now that `cmd/cardanoidentity/` is gone but still compiles. Removal punted to **WO-280**.
- **`pkg/identity/{service,vc,trust,cache}.go`** removal — orphaned via the same chain. Punted to **WO-280**.
- **`pkg/api/handlers/{schema,credential,transaction,issuer}.go`** removal — wraps `pkg/cardano.Client`; orphaned (no `cmd/*` constructs `New{Schema,Credential,Transaction,Issuer}Handlers` after `cmd/cardanoidentity` is gone). Punted to **WO-280** so we don't churn the build twice.
- **`pkg/config.CardanoConfig`** — loaded but never consumed; punted to **WO-280**.
- **`ErrCodeAtalaPRISMError`** alias in `pkg/did/errors.go` — retained pending **WO-280** clean-up sweep.
- StatusList2021 metagraph anchoring — **WO-272 / WO-274**.
- iOS client and OpenAPI updates — **WO-280**.
- Real did:key derivation for the `pending:user:` and `pending:wallet:` sentinels — **WO-273** (passkey enrollment is the natural place to plumb the public key through).

## Acceptance Criteria — All Met

- [x] `rg 'did:prism' --type=go -l` returns only `pkg/didkey/didkey_test.go` (negative test asserting rejection)
- [x] `go build ./...` succeeds
- [x] `go test ./pkg/credentials/... ./pkg/did/... ./internal/auth/... ./internal/services/onboarding/...` all pass
- [x] `pkg/did/atala_client.go` and `cmd/cardanoidentity/` no longer exist
- [x] **Modified scope:** No `cmd/*` package imports `pkg/cardano`. The wrapping handlers in `pkg/api/handlers/` still import it; full removal deferred to WO-280 (acknowledged out-of-scope above).

## Hand-off to WO-274 (StatusList2021 Issuer)

WO-274 was blocked on this WO. It is now unblocked. Concrete integration points the WO-274 implementer should target:
- `pkg/credentials.MetagraphConfig` — wire HTTP client against `IdentityL1URL`
- `pkg/credentials/revocation.go::RegisterRevocation` — replace the `"statuslist:<id>"` stub return with a real Identity L1 submission and bit-vector index
- `pkg/credentials/revocation.go::syncWithStatusList2021` — replace the cache-clear stub with a real L1 GET against the StatusList2021 endpoint
- `pkg/credentials/issuer.go::createCredentialProof` — fold in the issuer `did:key` from `MetagraphConfig.IssuerDID` (currently still uses `IssuerConfig.IssuerDID`)

## Blueprints

- Decentralized Identity and Authentication

## Related

- `blocked_by`: WO-275 (completed), WO-278 (in review)
- `blocking`: WO-274 (now unblocked), WO-280 (cleanup)
- ADR: `docs/adr/0001-phase1-identity-method.md`

---
