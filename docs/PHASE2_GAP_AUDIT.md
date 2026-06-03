# Phase 2 Gap Audit — Onboarding, Identity & Credentials

**Date:** 2026-05-26 (refreshed)  
**Prior audit:** 2026-05-24  
**Guide / source of truth:** `docs/Echo_Combined_Blueprints-2026-05-08.md`  
**Scope:** All 26 Phase 2 work orders (`docs/phase-2-work-orders.md`)  
**Software Factory sync:** 2026-05-26 — see phase doc headers  
**Architecture baseline:** `did:key` + Constellation Identity Metagraph. **Cardano is removed from Phase 1–2** (blueprint decision 2026-04-25; ADR-0001). Cardano/Midnight survive only as Phase 3+ ZK-circuit evaluation candidates.

**Frozen iOS UX (2026-05):** Shipped onboarding (`FirstRunCoordinator`) and login (`GlacialLoginScreen`) are **correct and must not be redesigned** from the React design prototype. WO-203/204 are backend/orchestration backlog — not a UI replacement mandate. See [`ux-spec.md`](ux-spec.md) and [`ECHO_IOS_UI_IMPLEMENTATION_SPEC.md`](ECHO_IOS_UI_IMPLEMENTATION_SPEC.md) §0.

---

## 1. Executive summary

- **WO-180 and WO-182 must NOT be implemented as written.** Both remain *blocked* in Software Factory. They specify Atala PRISM, `did:prism:cardano:` DIDs, Cardano anchoring, and Plutus-UTXO revocation — all obsolete. Superseded by **WO-273 (did:key)** and **WO-274 (W3C VC 2.0 + StatusList2021 on Constellation)**, both shipped.
- **Identity/VC backend foundations are done.** `pkg/did/`, `pkg/credentials/`, and `pkg/credentials/oidc4vc/` are real (no OIDC4VP always-succeeds stub). Do not rebuild.
- **Wave A integration gaps are largely closed** since the 2026-05-24 audit:
  - `GET /v1/users/check-username` is **mounted and tested** (`internal/api/username_handlers.go`, `router.go`).
  - `POST /v3/auth/refresh` and `/v3/auth/revoke` are **exposed** (`auth_token_handlers.go`, `router.go`).
  - Refresh tokens are **durable via Redis** when configured (`TokenService.SetRedisBackend`, `token_redisbackend_test.go`).
  - OPRF-PSI contact discovery backend is **implemented** (not stubbed): `OPRFService`, `POST /v3/contacts/psi`, durable discovery index (`cf18573`, `5a33252`).
- **Remaining Phase 2 backend gaps** are narrower: enrollment IDV/mDL/VC stubs, universal onboarding orchestration (WO-203), Prove integration, trust-registry durability, profile auto-generation, PSI tier opt-in policy, and credential-engine consolidation.
- **Remaining Phase 2 iOS gaps** are still the largest slice: WO-100 OIDC4VC client, WO-221 PSI client, WO-39 contact use-cases, WO-228 privacy settings consolidation. **WO-14 closed** via formal re-scope to `FirstRunCoordinator` (`docs/WO-14_ONBOARDING_RESCOPE.md`).
- **Cardano cleanup:** codebase remains essentially Cardano-free. Doc-side debt in blueprints/requirements persists (§5).

**Status tally (26 WOs):** ✅ Done 4 · 🟩 Substantially done 4 · 🟨 Partial 10 · 🟥 Stub 3 · ⬜ Missing 3 · ⛔ Obsolete 2  
*(Aligned with Software Factory: 4 completed · 3 in progress · 17 backlog · 2 blocked)*

---

## 2. What already exists (foundations — do not rebuild)

| Capability | Where | State |
|---|---|---|
| `did:key` derivation, registration, resolution, multi-device controller | `pkg/did/`, `pkg/didkey/` | **Done** (= WO-273) |
| W3C VC 2.0 issuance (JSON-LD / JWT / SD-JWT) | `pkg/credentials/issuer_vc2.go`, `formats.go`, `issuer.go` | **Done** (= WO-274) |
| Credential verification w/ real ECDSA, structure, expiry, revocation | `pkg/credentials/verifier.go`, `crypto.go` | **Done** |
| StatusList2021 revocation (Postgres + Identity-L1 publish) | `pkg/credentials/statuslist_pg.go`, `statuslist_l1_*.go` | **Done** |
| OIDC4VC issuer + verifier (VP-JWT, holder-sig check) | `pkg/credentials/oidc4vc/` | **Done** (gated `OIDC4VC_ENABLED`) |
| Passkey body-signature auth middleware | `internal/api/passkey_auth.go` | **Done** |
| JWT issuance + refresh rotation + reuse detection + JTI blocklist | `internal/auth/token.go` | **Done** |
| Refresh/revoke HTTP + Redis-backed refresh store | `auth_token_handlers.go`, `router.go`, `main.go` | **Done** (= WO-287) |
| Username availability (pre-auth) | `username_handlers.go`, `router.go` | **Done** (WO-14 dep) |
| OPRF-PSI server + discovery index | `contacts/oprf.go`, `contacts/service.go`, `/v3/contacts/psi` | **Substantially done** (WO-220) |
| `@username` → DID metagraph anchor | `v3_handlers.go` `anchorUsername`, `username_anchor_test.go` | **Substantially done** (WO-222 D1) |
| Trust scoring / tiers | `internal/services/trust/`, onboarding calculators | **Substantially done** |
| BIP-39 24-word recovery (iOS) | `ios/.../Onboarding/Recovery/*` | **Done** (= WO-234) |
| Phase 1 Glacial first-run onboarding | `FirstRunCoordinator`, WO-292 | **Done** |

---

## 3. Per-WO gap matrix

Legend: ✅ Done · 🟩 Substantially done · 🟨 Partial · 🟥 Stub · ⬜ Missing · ⛔ Obsolete

### Identity / Auth / Onboarding foundations

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **180** | DID Management (Atala PRISM/Cardano) | ⛔ | **Do not implement.** Superseded by WO-273 → `pkg/did/`. |
| **182** | VC Management (Cardano) | ⛔ | **Do not implement.** Superseded by WO-274 → `pkg/credentials/`. |
| **287** | Session JWT refresh / revocation (backend) | ✅ | `POST /v3/auth/refresh`, `/v3/auth/revoke` wired; Redis durable refresh (`SetRedisBackend`); tests in `auth_token_handlers_test.go`, `token_redisbackend_test.go`. SF: **completed**. |
| **14** | Credential-path onboarding (iOS) | ✅ | Re-scoped to `FirstRunCoordinator` (WO-292). `DisplayNameEntryView` polls `GET /v1/users/check-username` via `UsernameAvailabilityService`; `UsernameValidator` matches Go rules. OIDC4VC wallet path = WO-100. See `docs/WO-14_ONBOARDING_RESCOPE.md`. SF: **completed**. |
| **234** | BIP-39 recovery UI (iOS) | ✅ | `ios/.../Onboarding/Recovery/` complete. SF: **completed**. |
| **288** | New-device QR + recovery phrase (iOS) | 🟨 | `DeviceManagementView`, `QRIdentityView`, `RestoreFromPhraseView` exist. **Gap:** E2E controller-signed device transfer vs `/identity/devices` not verified. |

### Credentials / Verification / Trust

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **100** | OIDC4VC protocol client (iOS) | ⬜ | Backend done. **iOS client missing:** no full OIDC4VC registration/presentation flow (`RegisterWithVerifiableCredentialUseCase`, wallet connection). Enrollment scaffold only (`WalletCredentialEnrollmentView`). SF: **in_progress**. |
| **132** | VC issuance + wallet integration (iOS) | 🟨 | `EnrollmentCoordinator`, `EnrollmentAPIClient` exist. **Gap:** not full OIDC4VC credential-offer/wallet path. SF: **completed** (Constellation path; iOS polish remains). |
| **109** | VC verification engine + trust registry | 🟩 | Real engine in `pkg/credentials/verifier.go`. **Gap:** duplicate simulated engine in `onboarding/credentials.go` — consolidate. |
| **118** | Trust registry mgmt | 🟨 | `trust_registry.go` in-memory. **Gap:** durable store + issuer-admin API. |
| **129** | Auto profile generation from VC claims | ⬜ | Not implemented. |
| **199** | ISO 18013-5 mDL verification | 🟥 | `handleEnrollmentMDL` still stub `{"status":"ok"}`. iOS `AppleWalletMDLBridge` + scanner scaffold only. |

### Universal onboarding / SMS / phone

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **202** | SMS verification (Twilio + Prove) | 🟨 | Twilio real; OTP/SMS recovery wired. **Gap:** Prove integration missing. |
| **203** | Universal onboarding orchestration | ⬜ | Backend phone-first coordinator service — **not** a redesign of frozen `FirstRunCoordinator` UI. |
| **204** | Universal onboarding UI (iOS) | ⬜ | Depends on WO-203; **do not** replace shipped welcome/login/recovery screens with prototype onboarding. |
| **205** | Phone number mgmt + deletion | 🟨 | `onboarding/phone.go` exists; deletion flow unverified. |
| **199**/IDV | `/v1/enrollment/{vc,idv,mdl}` | 🟥 | Still stub bodies in `enrollment_handlers.go`. |

### Contacts / discovery / privacy

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **220** | PSI contact discovery backend | 🟨 | **Upgraded from stub:** `OPRFService` (RFC 9497), `OPRFEvaluate`, durable `DiscoveryIndex`, `POST /v3/contacts/psi`, tests (`oprf_test.go`). SMS verify commits OPRF keys (`sms_recovery_handlers.go`). **Gaps:** Tier 3+ discoverability filter, ops key rotation policy, WO spec used `/v1/contacts/psi/blind-eval` path name. SF: **in_progress**. |
| **221** | PSI contact discovery iOS client | ⬜ | No client-side OPRF blind/evaluate/finalize flow on iOS. |
| **222** | Username index, QR, invite links | 🟨 | Username search + invite links (`/v3/contacts/search`, `/invite`); metagraph anchor on register (`anchorUsername`); iOS `QRIdentityView`. **Gaps:** parallel `trustnet/discovery.go` vs `contacts/service.go`; deep links `echo://invite`; public unauthenticated lookup API. SF: **in_progress**. |
| **39** | User mgmt + contacts w/ privacy (iOS) | ⬜ | `ContactDetailView` exists; WO-39 use-case folder (PSI, QR, invite, username search) missing. |
| **187** | User profiles + contact mgmt | 🟨 | `ProfileScreens`, backend `/v3/contacts/*` partial. |
| **190** | Contact blocking + privacy | 🟨 | Backend block/unblock wired; iOS privacy controls partial. |
| **228** | Privacy settings + deletion UI (iOS) | 🟨 | Models/settings refs exist; dedicated screen + account deletion not consolidated. |

### Onboarding support / analytics

| WO | Title | Status | Evidence / Gap |
|---|---|---|---|
| **144** | Onboarding analytics + support | 🟨 | `analytics.go` exists; support side not built. |
| **159** | Verification retry + guidance | ⬜ | Not implemented. |
| **169** | Onboarding help resources | ⬜ | Content/help-center work, not code. |

---

## 4. Endpoint reachability (updated 2026-05-26)

| Endpoint | Prior audit | Now |
|---|---|---|
| `GET /v1/users/check-username` | ❌ Unreachable | ✅ `username_handlers.go` + public route (`b71db7b`) |
| `POST /v3/auth/refresh`, `/v3/auth/revoke` | ❌ No routes | ✅ `auth_token_handlers.go` (`c590dbb`) |
| Refresh token storage | ❌ In-memory only | ✅ Redis when backend configured (`1cf5ac8`) |
| `POST /v3/contacts/psi` | ❌ Stub `PSIDiscovery` | ✅ OPRF evaluate + discovery index (`cf18573`, `5a33252`) |
| `/v1/enrollment/{vc,mdl,idv}` | 🟥 Stub | 🟥 Still stub (`enrollment_handlers.go`) |

---

## 5. Cardano remnant inventory & removal plan

*(Unchanged in substance — code clean; doc debt remains.)*

### 5a. Code / config — **essentially clean**

| Location | Verdict |
|---|---|
| `Makefile` `cardanoidentity` guard | **Removed** |
| `hydra` in Makefile / compose | **Keep** — Euclid metagraph tool, not Cardano |
| `ValidationsSpec.scala` rejects `did:prism` | **Keep** — defensive test |
| Sample username `"ada"` in demos | **Keep** — not ADA token |

### 5b. Docs — **the real debt**

`Echo_Combined_Blueprints-2026-05-08.md` still has ~180 Cardano mentions including stale duplicate DID-management blocks. `phase-*-work-orders.md` WO-180/182 archive blocks are historical. Recommend a focused doc sweep (not bundled with feature work).

---

## 6. Dependency-ordered remediation plan

**Wave A — backend integration (mostly ✅):**

| # | Item | Status |
|---|------|--------|
| A.1 | Mount `GET /v1/users/check-username` | ✅ Done |
| A.2 | Expose refresh/revoke + Redis refresh store | ✅ Done |
| A.3 | Consolidate onboarding credential proof onto `pkg/credentials` crypto | ✅ Done (`credentials_bridge.go`, Ed25519 verify when issuer key published) |

**Wave B — stubs → real implementations:**

| # | Item | Status |
|---|------|--------|
| B.1 | OPRF-PSI backend | 🟨 Core done; tier opt-in + ops hardening remain (WO-220) |
| B.2 | `/v1/enrollment/{idv,mdl,vc}` + Prove/mDL | 🔜 Open (WO-199, WO-202) |
| B.3 | Universal onboarding orchestration + phone deletion | 🔜 Open (WO-203, WO-205) |
| B.4 | Trust registry durability + profile auto-gen | 🔜 Open (WO-118, WO-129) |

**Wave C — iOS (largest remaining effort):**

| # | Item | Status |
|---|------|--------|
| C.1 | WO-14 credential-path onboarding OR formal re-scope to `FirstRunCoordinator` | ✅ Done (re-scope + username availability UI) |
| C.2 | WO-100 OIDC4VC client + WO-132 wallet polish | 🔜 Open / in progress |
| C.3 | WO-221 PSI iOS client + WO-39 contact use-cases | 🔜 Open |
| C.4 | WO-228 privacy settings / account deletion UI | 🔜 Open |

**Wave D — support & docs:**

| # | Item | Status |
|---|------|--------|
| D.1 | WO-144/159/169 support surfaces | 🔜 Open |
| D.2 | Cardano doc sweep (§5b) | 🔜 Open |

---

## 7. Recommended immediate next steps

Pick based on current sprint (Phase 3 Xcode wiring is parallel, not blocking these):

1. **WO-100** — OIDC4VC iOS client against shipped backend (`oidc4vc/`). Highest Phase 2 credential gap.
2. **WO-221** — iOS OPRF-PSI client to consume `/v3/contacts/psi` (unblocks contact discovery E2E with WO-220).
3. ~~**WO-14**~~ — ✅ Re-scoped to `FirstRunCoordinator`; username availability wired in `DisplayNameEntryView`.
4. ~~**Wave A.3**~~ — ✅ Onboarding credential proofs delegate to `pkg/credentials` when issuer publishes `VerificationPublicKeyBase64`.

---

## 8. Related Phase 3 work (out of Phase 2 scope but active)

Software Factory **in_progress** on Phase 3 messaging signals (adjacent to Phase 2 contacts/privacy):

| WO | Title | Note |
|---|---|---|
| **192** | Typing indicators + read receipts | Backend + iOS agent layer shipped; Xcode UI wiring in progress |
| **10** | Emoji reactions | REST + WS backend shipped; iOS UI partial |

See `docs/PHASE3_IOS_UI_SPEC.md` and skill `echo-phase3-ios-wire`.
