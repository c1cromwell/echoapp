# Phase 2 Gap Audit — Onboarding, Identity & Credentials

**Date:** 2026-05-24
**Guide / source of truth:** `docs/Echo_Combined_Blueprints-2026-05-08.md`
**Scope:** All 26 Phase 2 work orders (`docs/phase-2-work-orders.md`)
**Architecture baseline:** `did:key` + Constellation Identity Metagraph. **Cardano is removed from Phase 1–2** (blueprint decision 2026-04-25; ADR-0001). Cardano/Midnight survive only as Phase 3+ ZK-circuit evaluation candidates.

---

## 1. Executive summary

- **WO-180 and WO-182 must NOT be implemented as written.** Both are explicitly marked *"⚠ Blocked — Do not implement"* in the WO doc. They specify Atala PRISM, `did:prism:cardano:` DIDs, Cardano anchoring, and Plutus-UTXO revocation — all obsolete. They were **superseded by WO-273 (did:key) and WO-274 (W3C VC 2.0 + StatusList2021 on Constellation)**.
- **Those replacements are already built and tested** in `pkg/did/`, `pkg/credentials/`, and `pkg/credentials/oidc4vc/`. There is **no greenfield "backend foundation" to build** for identity/VC. The G5 "OIDC4VP always-succeeds" stub noted in prior Phase-1 analysis is **resolved** — real VP/credential signature verification now exists.
- **The real Phase 2 backend gaps** are a thin integration layer the iOS work orders depend on: an unreachable username-availability endpoint, unexposed session refresh/revoke endpoints, a stubbed PSI discovery query, stubbed IDV/mDL enrollment handlers, and missing orchestration/profile-generation/Prove integrations.
- **The real Phase 2 iOS gaps** are larger: WO-14's named onboarding coordinator/views, the WO-100 OIDC4VC client, and the WO-39/221 contact use-cases do not exist yet (equivalent Phase 1 `FirstRunCoordinator` onboarding does exist).
- **Cardano cleanup:** the **codebase is already essentially Cardano-free** (one vestigial Makefile string, now removed). The remaining Cardano debt is **doc-side** — primarily the blueprint, which still contains a stale duplicate DID-management section and ~180 Cardano mentions. See §5.

**Status tally (26 WOs):** Done 1 · Substantially done 4 · Partial 9 · Stub 4 · Missing 6 · Obsolete (do-not-implement) 2.

---

## 2. What already exists (foundations — do not rebuild)

| Capability | Where | State |
|---|---|---|
| `did:key` derivation, registration, resolution, multi-device controller pattern | `pkg/did/` (`service.go`, `multidevice.go`, `resolver.go`, `repository.go`, `cache.go`, `handlers.go`); `pkg/didkey/` | **Done** (= WO-273) |
| W3C VC 2.0 issuance (JSON-LD / JWT / SD-JWT) | `pkg/credentials/issuer_vc2.go`, `formats.go`, `issuer.go` | **Done** (= WO-274) |
| Credential verification w/ **real ECDSA** signature check, structure, expiry, revocation | `pkg/credentials/verifier.go:301`, `crypto.go:315` | **Done** |
| StatusList2021 revocation (Postgres + Identity-L1 publish) | `pkg/credentials/statuslist_pg.go`, `statuslist_l1_*.go`, `identity_l1_publish.go` | **Done** |
| OIDC4VC issuer + verifier (real `VerifyPresentation`, VP-JWT parse, holder-sig check) | `pkg/credentials/oidc4vc/` (`issuer.go`, `verifier.go:190`, `metadata.go`, `flows.go`) | **Done** (gated by `OIDC4VC_ENABLED`; mounted via `router.OIDC` in `main.go`) |
| Passkey body-signature auth middleware (X-Sender-DID + X-Signature) | `internal/api/passkey_auth.go`, `router.go:258` | **Done** |
| ES256 JWT issuance + refresh-token rotation w/ reuse detection + JTI blocklist | `internal/auth/token.go` | **Logic done, not fully wired** — see WO-287 |
| Trust scoring / tiers (0–100 → Tier 1–5) | `internal/services/trust/`, `internal/services/onboarding/credentials.go` (TrustScoreCalculator) | **Substantially done** |
| BIP-39 24-word recovery (iOS) | `ios/.../Onboarding/Recovery/*` | **Done** (= WO-234) |

---

## 3. Per-WO gap matrix

Legend: ✅ Done · 🟩 Substantially done · 🟨 Partial · 🟥 Stub · ⬜ Missing · ⛔ Obsolete

### Identity / Auth / Onboarding foundations

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **180** | DID Management (Atala PRISM/Cardano) | ⛔ | **Do not implement.** Obsolete. Superseded by WO-273 → `pkg/did/` (done). |
| **182** | VC Management (Cardano) | ⛔ | **Do not implement.** Obsolete. Superseded by WO-274 → `pkg/credentials/` (done). |
| **287** | Session JWT issuance / refresh / revocation (backend) | 🟨 | `internal/auth/token.go` has issue, `RotateRefreshToken` (reuse-detection → revoke), `BlocklistToken`. **Gaps:** refresh tokens stored in an **in-memory map** (`refreshTokens`), not durable; **no HTTP routes** (`/v3/auth/refresh`, `/v3/auth/revoke` absent from `v3_handlers.go`). |
| **14** | Streamlined onboarding flow (iOS) | 🟨 | Equivalent Phase 1 flow exists (`FirstRunCoordinator`, `NameAndKeyView`, `PasskeySetupView`, recovery, VIP). **Gap:** WO-14's named `OnboardingCoordinator`, `UsernameView`, `DIDCreationProgressView`, `OptionalVerificationView`, `OnboardingCompleteView` do not exist; `GET /v1/users/check-username` is **unreachable** (see below). |
| **234** | 24-word BIP-39 recovery UI (iOS) | ✅ | `ios/.../Onboarding/Recovery/` complete (wordlist, display, confirm, restore, scheduler). |
| **288** | New-device login via QR transfer + recovery phrase (iOS) | 🟨 | `DeviceManagementView`, `QRIdentityView`, `MDLQRScannerView`, `EnrollmentCoordinator`, `RestoreFromPhraseView` exist. **Gap:** end-to-end QR device-transfer handshake (controller-signed nonce → secondary `did:key` → `DeviceAttestationCredential`) not verified wired against backend `/identity/devices`. |

### Credentials / Verification / Trust

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **100** | OIDC4VC protocol client (iOS) | ⬜ | Backend verifier done (`oidc4vc/verifier.go`). **iOS client missing:** no `RegisterWithVerifiableCredentialUseCase`, `WalletConnectionView`, OIDC4VC request construction, or `submitPresentation` call. |
| **132** | VC issuance + wallet integration (iOS) | 🟨 | `WalletCredentialEnrollmentView`, `EnrollmentCoordinator`, `EnrollmentModels`, `EnrollmentAPIClient` exist. **Gap:** not a full OIDC4VC credential-offer/wallet flow. |
| **109** | VC verification engine + trust registry | 🟩 | Real engine in `pkg/credentials/verifier.go`. **Note:** a second, in-memory engine with *simulated* signature check exists in `internal/services/onboarding/credentials.go:183` — consolidation needed to avoid divergence. |
| **118** | Trust registry mgmt for issuer verification | 🟨 | `internal/services/onboarding/trust_registry.go` (`GetIssuer`, issuer status/level). **Gap:** appears in-memory; no durable store or admin/issuer-onboarding API surface verified. |
| **129** | Automatic profile generation w/ verified data | ⬜ | No profile-generation service found. Trust badges/score exist but auto-population of profile from verified VC claims is absent. |
| **199** | ISO/IEC 18013-5 mDL verification | 🟥 | `handleEnrollmentMDL` returns `{"status":"ok"}` (`enrollment_handlers.go:220`). iOS has `AppleWalletMDLBridge` + `MDLQRScannerView` scaffold. Real mDL device-engagement / issuer-data verification missing. |

### Universal onboarding / SMS / phone

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **202** | SMS verification (Twilio + Prove) | 🟨 | Twilio **real** (`internal/infra/sms_provider.go` `TwilioSMSProvider`), SMS-recovery handlers wired, OTP flow works. **Gap:** **Prove** identity integration missing. |
| **203** | Universal onboarding backend orchestration | ⬜ | No orchestration service. Component pieces (phone, passkey, DID, VC) exist independently; the phone-number-first universal flow coordinator is absent. |
| **204** | Universal onboarding UI flow (iOS) | ⬜ | Depends on WO-203. Phone-number onboarding UI not present (current onboarding is username + passkey first-run). |
| **205** | Phone number mgmt + optional deletion | 🟨 | `internal/services/onboarding/phone.go` exists. **Gap:** deletion endpoint/flow and management UI not verified. |
| **199**/IDV | `/v1/enrollment/vc|idv` tail | 🟥 | `handleEnrollmentVC`, `handleEnrollmentIDV` are stubs returning `{"status":"ok"}` (`enrollment_handlers.go:211,229`). No Prove/Daon/Darwinium integration. |

### Contacts / discovery / privacy

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **220** | PSI contact discovery backend | 🟥 | `internal/services/contacts/service.go:44` `PSIDiscovery` is a **stub** (`_ = hash // In production, this would query the hashed phone index`). `HashPhone` (Argon2id) is real. `/v3/contacts/psi` is wired but returns stub data. No OPRF. |
| **221** | PSI contact discovery iOS client | ⬜ | No Argon2id hashing or PSI client on iOS. |
| **222** | Username index, QR, invite links | 🟨 | DB-backed invite links + username search (`contacts/service.go`), plus in-memory `trustnet/discovery.go` (QR, username). **Gap:** two parallel impls; username index on Identity Metagraph not present; consolidation needed. |
| **39** | User mgmt + contact system w/ privacy controls (iOS) | ⬜ | `ContactDetailView` exists, but the WO-39 `Contacts/` feature folder + 4 use-cases (`ContactDiscoveryUseCase` w/ Argon2id, `QRContactExchangeUseCase`, `InviteLinkUseCase`, `UsernameSearchUseCase`) are missing. |
| **187** | User profiles + contact mgmt | 🟨 | `ProfileScreens`, `ContactDetailView`, backend contacts CRUD (`/v3/contacts/*`). Partial UI. |
| **190** | Contact blocking + privacy control | 🟨 | Backend `BlockContact`/`UnblockContact` + `/v3/contacts/block` wired. iOS per-setting privacy controls partial. |
| **228** | Privacy settings screen + encryption indicator + account deletion UI (iOS) | 🟨 | References in `ProfileScreens`, `Models`, `ViewModels`, `Endpoints`. **Gap:** dedicated privacy-settings screen, encryption indicator, and account-deletion UI not consolidated. |

### Onboarding support / analytics

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **144** | Onboarding analytics + support | 🟨 | `internal/services/onboarding/analytics.go` exists. Support side not built. |
| **159** | Verification retry + error handling + user guidance | ⬜ | No dedicated retry/error-guidance service found. |
| **169** | Onboarding support docs + help resources | ⬜ | Not present (largely content/help-center, not backend code). |

---

## 4. Unreachable / unexposed endpoints (live router vs. blueprint)

Implemented logic exists but is **not reachable** on the running server (`internal/api/router.go`):

1. **`GET /v1/users/check-username`** (WO-14 dependency) — handler exists at `pkg/api/v2/profile_handlers.go:455` but **`pkg/api/v2` is never imported/mounted**; `OnboardingService.CheckUsernameAvailability` (`onboarding.go:613`) is not on a route. `handleV1`/`handleV2` only serve `/v1/users`, `/v1/users/profile`, `/v2/users`, `/v2/users/profile`.
2. **`POST /v3/auth/refresh` and `/v3/auth/revoke`** (WO-287) — `token.go` has the rotation/revocation logic; no routes register it (`v3_handlers.go:45`).
3. **`/v1/enrollment/{vc,mdl,idv}`** — wired but **stub bodies** (`{"status":"ok"}`).
4. **`/v3/contacts/psi`** — wired but backed by the stubbed `PSIDiscovery`.

---

## 5. Cardano remnant inventory & removal plan

### 5a. Code / config — **essentially clean**

| Location | Match | Verdict |
|---|---|---|
| `Makefile:179-180` | `cardanoidentity` stale-binary guard | **Removed** in this pass (obsolete binary). |
| `Makefile`, `docker-compose.*.yml`, `trustnet/blockchain.go:239` | `hydra` | **Keep — not Cardano.** `hydra` = Constellation/Euclid metagraph deploy tool. |
| `IdentityCardView.swift:204`, `DemoFlows.swift:286` | `username: "ada"` | **Keep — not Cardano.** Sample username (Ada Lovelace), not the ADA token. |
| `metagraph/.../ValidationsSpec.scala:190` | `"rejects did:prism strings"` | **Keep — defensive.** Test asserts `did:prism` is *rejected* as obsolete PII. |
| `CHANGELOG.md:12,71` | "No Cardano/PRISM dependency", "replaced by did:key" | **Keep — historical record** of the removal. |

### 5b. Docs — **the real debt** (recommended follow-up sweep)

The guide doc (`Echo_Combined_Blueprints-2026-05-08.md`, dated *after* the 2026-04-25 removal) still has **~180 Cardano mentions**. Two classes:

- **KEEP / lightly trim** — deliberate "Cardano not used in Phase 1–2; Phase 3 ZK-eval only" statements that document the decision: lines ~1952, 2076, 2107–2122, 2130, 2568, 2610, 4201, 4231, 4518.
- **REMOVE / rewrite** — stale content that presents Cardano as live Phase 1–2 spec:
  - **`#### Decentralized Identifier (DID) Management` duplicate at lines ~4655–4710** — an entire **stale legacy copy** ("generates a new DID using Atala PRISM infrastructure", "anchored to Cardano blockchain", `did:prism:cardano:abc123` doc + `verifiableCredential` embedded in DID doc). Directly contradicts the canonical did:key section at line 4227. **Delete this duplicate block.** (Note: subsections "Streamlined Onboarding" / "In-App High-Assurance" also appear twice — 4562/5116 and 4577/5130 — indicating a broader stale mirror region to reconcile.)
  - Lines 4480 ("Credential is returned to backend and stored on Cardano") and 4481 ("added to user's DID document") → should read "anchored on the Constellation Identity Metagraph" and drop DID-doc embedding (per the note at line 4279).
  - Backend/Data-Layer/Secure-Enclave/Privacy stale refs: ~192, 209, 517, 2150, 2198, 2202, 2411, 2466, 2597, 2604, 2815, 2818, 2912, 2976.
  - Later-phase feature sections (out of Phase 2 scope but heavily Cardano): Portable Social Graph (~3656–3767), PQ Mode (~3866–3914), Privacy Commons Treasury (~3973–4056).
- **Also:** `docs/Echo_Combined_Requirements.md` (~77 mentions) and `docs/phase-1..7-work-orders.md` carry Cardano; WO-180/182 "original content preserved for reference" blocks are historical archives — flag, don't gut.

> **Recommendation:** the doc sweep is a bounded but sizable edit to the canonical source-of-truth (~180 mentions, multiple duplicate regions, all 7 phases). It deserves its own focused pass with a reviewable diff rather than being bundled with code work. Code/config cleanup is complete.

---

## 6. Dependency-ordered remediation plan

**Wave A — backend integration gaps (unblocks iOS; small, testable in Go):**
1. Mount `GET /v1/users/check-username` (reuse `OnboardingService.CheckUsernameAvailability`) + tests. *(WO-14 dep)*
2. Expose `POST /v3/auth/refresh` + `/v3/auth/revoke` over `token.go`; move refresh-token store to durable backing (Postgres/Redis) + tests. *(WO-287)*
3. Consolidate the two credential-verification engines onto `pkg/credentials/verifier.go`; retire the simulated one in `onboarding/credentials.go`. *(WO-109)*

**Wave B — real implementations behind existing stubs:**
4. Implement PSI discovery query (OPRF over the Argon2id hashed-phone index) behind `/v3/contacts/psi`. *(WO-220)*
5. Implement `/v1/enrollment/{idv,mdl,vc}` against a real IDV provider (Prove) + mDL ISO 18013-5 verification. *(WO-199, WO-202)*
6. Universal onboarding orchestration service + phone management/deletion. *(WO-203, WO-205)*
7. Trust-registry durability + issuer-admin API. *(WO-118)*; automatic profile generation. *(WO-129)*

**Wave C — iOS feature builds (largest effort):**
8. WO-14 `OnboardingCoordinator` + named views (or formally re-scope to the existing `FirstRunCoordinator`).
9. WO-100 OIDC4VC client + WO-132 wallet credential flow.
10. WO-39/WO-221 contacts feature folder + 4 use-cases (Argon2id) + PSI client.
11. WO-228 privacy settings / encryption indicator / account deletion UI; WO-204 universal onboarding UI.

**Wave D — support & docs:**
12. WO-144 support, WO-159 retry/guidance, WO-169 help resources.
13. Full Cardano doc sweep (§5b).

---

## 7. Recommended immediate next step

Start **Wave A.1** — wire `GET /v1/users/check-username` with tests. It's the smallest concrete gap, unblocks WO-14, is fully testable in Go here, and validates the audit's "logic exists but unreachable" finding end-to-end.
