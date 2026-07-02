# Phase 4–7 Gap Audit — Blockchain, Privacy, Calls, Platform

**Date:** 2026-05-29  
**Software Factory sync:** 2026-05-29  
**Scope:** Phases 4–7 (`docs/phase-4-work-orders.md` … `phase-7-work-orders.md`)  
**Messaging MVP closure:** [`ECHO_MESSAGING_LAUNCH_STATUS.md`](ECHO_MESSAGING_LAUNCH_STATUS.md)  
**Cross-product sequencing:** [`CROSS_PRODUCT_GAP_REVIEW.md`](CROSS_PRODUCT_GAP_REVIEW.md) (ADRs 0005/0006)

**Architecture baseline:** Constellation Identity + Data + Currency L1; `did:key` only in Phase 1–2. Cardano WOs in Phase 4 remain **blocked** — do not implement.

---

## Executive summary

| Phase | SF tally | Messaging MVP slice | Primary gaps |
|-------|----------|---------------------|--------------|
| **4** Blockchain & trust | 9 ✅ · 20 backlog | Trust/rewards + **WO-15 anchoring** ✅ | Tokenomics, prod infra, Passport Wave B (298–300) |
| **5** Privacy features | 14 ✅ · 32 backlog | Disappearing, sealed sender, personas ✅ | Scheduled/hidden-folder polish; ZK/PQ/data-sovereignty post-MVP |
| **6** Calls & files | 4 ✅ · 28 backlog | 1:1 calls + groups + voice notes ✅ | Channels, large-file IPFS, VIP StoreKit |
| **7** Platform | 40 ✅ · 1 🔄 · 62 backlog | Rewards/wallet/staking ✅ | **ECHO Comply** verticals + org lifecycle; **ECHO Passport** iOS (WO-297) |

**Active product tracks after messaging MVP:** ECHO Comply (WO-253–307, WO-281–286; WO-309 in progress) and ECHO Passport (WO-297 iOS; Wave B WO-298–300 in Phase 4 export).

---

## Phase 4 — Blockchain & trust infrastructure (29 WOs)

### Shipped (do not rebuild)

| WO | Area | Evidence |
|----|------|----------|
| **8, 27** | Metagraph gateway | `metagraph.MetagraphClient`, `metagraph.Gateway` |
| **15, 94** | Message anchoring + proofs | `anchoring_service.go`, `GET …/merkle-proof`, `GET …/proof` |
| **22, 49, 181** | Trust tiers + dynamic scoring | `internal/services/trust/`, metagraph trust commitments |
| **114, 213** | Reward calculation + volume decay | Reward engine, `make validate-phase1` reward steps |
| **272, 274** | Identity L1 + VC 2.0 | Phase 1 ✅ — `IdentityValidations.scala`, `pkg/credentials/` |
| **(wallet)** | Currency L1 balances / stake | `internal/wallet/`, `/v3/wallet/*`, `scripts/validate-wallet.sh` |

### Removed — Cardano WOs

WO-12, WO-20, WO-37 deleted from active scope — [`PHASE4_CARDANO_REMOVED.md`](PHASE4_CARDANO_REMOVED.md).

### Backlog — post-MVP / go-live

| Cluster | WOs | Notes |
|---------|-----|-------|
| Tokenomics & genesis | 206, 214, 215, 225, 226, 271 | Founder vesting, treasury, quest system |
| Production infra | 210, 231, 232 | K8s, mainnet nodes, CI/CD — pairs with WO-233 App Store gates |
| Metagraph orchestration | 45, 164, 179 | Extended Data L1 app types |
| Privacy data model | 209, 229 | Community relay registry |
| Passport Wave B (Phase 4 export) | **298–300** | Metagraph credential refs, pay-in-chat backend + iOS UX |

---

## Phase 5 — Hidden folders & privacy (46 WOs)

### Shipped (messaging-relevant)

| WO | Feature | Evidence |
|----|---------|----------|
| **7, 30, 47** | Hidden folders + biometric gate | `HiddenFolderSettingsStore`, privacy screens |
| **38, 75, 125** | Disappearing messages + screenshot deterrence | `ChatDetailViewModel.purgeDisappearingMessages`, `DisappearingMessageEnforcer` |
| **56** | Silent message infrastructure | Notification suppression path |
| **72, 82, 102** | Personas + selective visibility | Persona models + settings |
| **219** | Sealed sender | `internal/api` sealed-sender handlers |
| **317–319** | SimpleX SX3–SX6 | Metadata minimization, persona, Tor proxy |

### Backlog (non-blocking for messaging launch)

| Cluster | WOs | Gap |
|---------|-----|-----|
| Hidden-folder polish | 18, 42, 51, 61, 69, 78 | Enhanced encryption, backup, access audit |
| Scheduled / time-locked messages | 65, 67, 76, 87, 96 | UI stubs only |
| Client Merkle verify | **227** | Depends on WO-15 proof API |
| App Store / open source | **233** | Go-live checklist — [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md) |
| ZK / Midnight | 212, 235, 236 | Evaluation only |
| Post-quantum mode | 257, 258, 259 | Future hardening |
| Data sovereignty | 248, 249 | Opt-in query economy |
| Portable identity / deletion | 255, 256, 218 | GDPR hardening |

---

## Phase 6 — Calls & file sharing (32 WOs)

### Shipped (MVP slice)

| WO | Feature | Evidence |
|----|---------|----------|
| **5, 19** | WebRTC infrastructure + call UI | `WebRTCLiveCallEngine`, `WebRTCCallSession`, Phase3 tests |
| **70** | Core groups | `GroupCreateSheet`, `GroupChatView`, group APIs |
| **316** | Voice messages (SX5) | `VoiceNoteRecorder`, group + DM wiring |

### Backlog

| Cluster | WOs | Gap |
|---------|-----|-----|
| **Broadcast channels** | 77, 88, 97, 108, 121, 130, 139, 150, 158, 168, 175 | Channels segment hidden in iOS until shipped |
| **Large file sharing** | 21, 34, 46, 58, 68, 79, 185 | IPFS chunking, cloud connectors, Filecoin |
| **Call polish** | 31, 43, 55, 62, 71 | Screen share, recording, transcription |
| **Group polish** | 81, 90, 101, 112, 123, 131, 148 | Verification badges, moderation, governance |
| **VIP subscription** | **238** | StoreKit IAP — P0 in App Store checklist |
| **Relay registry** | 229 | Phase 4/6 cross-cut |

---

## Phase 7 — Advanced platform (103 WOs)

### Shipped — rewards & wallet (messaging-adjacent)

Large cluster completed in SF: WO-95, 106, 116, 127–127, 133, 138, 142–147, 151–155, 160–161, 166–167, 170–171, 173–176, 184, etc.  
**Code:** `internal/wallet/`, iOS Rewards tab, `scripts/validate-wallet.sh`.

### ECHO Comply

| Layer | WOs | Status |
|-------|-----|--------|
| **Service core** | 250–252 | ✅ `internal/services/comply/`, `cmd/comply` |
| **Web portal** | 308–313 | ✅ `web/` Next.js portal (dashboard, retention, org, eDiscovery shell) |
| **iOS context** | 289 | ✅ Comply coordinator / composer guardrails |
| **eDiscovery / matters UI** | **309** | 🔄 **in_progress** |
| **Vertical packs** | 253–268, 269 | backlog — HIPAA, FOIA, law firm, financial |
| **Org lifecycle** | **281–286** | backlog — org DID, BAA, invites, StatusList, SSO, Stripe seats |
| **FINRA/SEC** | 298–307 | backlog — blueprint WO-298 + 8 build WOs (Phase 7 numbering) |

**Gap:** Comply **operator surface** exists (web wedge shipped); **regulated vertical logic** and **enterprise org onboarding/billing** are the remaining Comply program.

### ECHO Passport

| WO | Wave | Status |
|----|------|--------|
| 293–296 | A — holder backend | ✅ `pkg/passport/`, `/v1/passport/*` |
| **297** | A — **iOS module** | **backlog** |
| 298–300 | B — metagraph + pay-in-chat | backlog (listed in Phase 4 export) |
| 301–307, 316 | C/D — verifier, x402, agentic | backlog — see ADR 0006 |

### Other Phase 7 backlog (non-Comply/Passport)

Bots (40, 52, 63, 74, 85, 186–195), enterprise profiles (80, 89, 98, 107, 119), DeFi bridge (174), governance (177, 290), AI treasury (291), call/enterprise features — all post-messaging MVP.

---

## Go-live vs feature gap (Phases 4–7)

| Item | Type | WO / doc |
|------|------|----------|
| App Store security gates | Go-live | WO-233, [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md) |
| VIP StoreKit | Go-live P0 | WO-238 |
| Two-client E2E | Go-live | [`E2E_QUICK_START.md`](E2E_QUICK_START.md) |
| **Message anchoring (Data L1)** | Feature gap (optional pre-integrity marketing) | WO-15 |
| Comply verticals + org billing | Next product | WO-253–307, 281–286 |
| Passport iOS wallet | Next product | WO-297 |

---

## Recommended sequencing (after messaging MVP)

1. **Go-live:** device E2E, TestFlight, App Store P0s (WO-233, WO-238).
2. **Comply:** finish WO-309 → org lifecycle WO-281–286 → one vertical wedge (e.g. law firm WO-262–265 or FINRA WO-299–301).
3. **Passport:** WO-297 iOS module → Wave B WO-298–300 (pay-in-chat).
4. **Integrity (optional parallel):** complete WO-15 + WO-227 client verify before “blockchain-anchored messaging” claims.

---

## Related

- [`ECHO_MESSAGING_LAUNCH_STATUS.md`](ECHO_MESSAGING_LAUNCH_STATUS.md)
- [`PHASE2_GAP_AUDIT.md`](PHASE2_GAP_AUDIT.md)
- [`ECHO_PASSPORT_PLAN.md`](ECHO_PASSPORT_PLAN.md)
- [`adr/0005-comply-web-admin-portal.md`](adr/0005-comply-web-admin-portal.md)
- [`adr/0006-passport-x402-agentic.md`](adr/0006-passport-x402-agentic.md)
