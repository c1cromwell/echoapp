# ECHO Messaging — Launch Status

**Last synced with Software Factory:** 2026-05-29  
**Scope:** Consumer messaging MVP (Phases 1–3 core + Phase 6 groups/calls baseline)

---

## Executive summary

| Track | Status |
|-------|--------|
| **Messaging MVP (code)** | ✅ Shipped in `main` — encryption, DMs, groups, signals, search, voice notes, wallet/rewards |
| **Device E2E / TestFlight** | 🔜 Go-live — [`E2E_TESTING.md`](E2E_TESTING.md) + [`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md) |
| **App Store submission** | 🔜 Go-live — see [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md) (WO-233) |
| **ECHO Comply** | 📋 Active — Phase 7 vertical WOs + org lifecycle (WO-281–307, WO-309 in progress) |
| **ECHO Passport** | 📋 Active — Wave A backend ✅; **WO-297** iOS module backlog |

Phone-first universal onboarding UI (WO-204) is **closed for MVP** — superseded by `FirstRunCoordinator` (WO-292); SMS phone path deferred post-launch.

---

## Messaging MVP — completed work orders

### Phase 1 — Foundation (26/26 ✅)

Identity (`did:key`), passkeys, JWT/WS auth, T0–T7 CI, metagraph validators, WO-230 go/no-go harness.

### Phase 2 — Onboarding & contacts (messaging-relevant)

| WO | Title | SF status |
|----|-------|-----------|
| 14 | Streamlined onboarding UI | ✅ → `FirstRunCoordinator` |
| 39, 187, 190 | Contacts, profiles, blocking | ✅ |
| 100, 109 | OIDC4VC + credential verification | ✅ |
| 203 | Universal onboarding backend | ✅ |
| 204 | Universal onboarding UI | ✅ (MVP superseded by WO-292) |
| 220–222, 221 | PSI + username/QR discovery | ✅ |
| 228, 234, 287, 288 | Privacy settings, recovery, sessions | ✅ |
| 292 | Glacial first-run coordinator | ✅ (Phase 1) |

**Not messaging MVP** (remain backlog): WO-118 trust-registry durability, WO-129 profile auto-gen, WO-199 mDL, WO-202 SMS, WO-205 phone deletion, WO-297 Passport iOS.

### Phase 3 — Messaging core (24/37 ✅ for MVP)

**Shipped:** WO-4, 10, 25, 28, 48, 57, 59, 103, 192, 194, 196–198, 207, 237, 314 (SX1 ratchet), plus search/archive WOs 3, 16, 23, 29, 54, 64, 73, 84.

**Post-MVP backlog (do not block launch):** IDV WOs (17, 26, 104, 113, 120, 126), search extras (41, 83, 92), reactions polish (36, 50), disappearing policies (149), broadcast (172).

### Phase 6 — Groups & calls (MVP slice ✅)

WO-5, 19, 70, 316 (voice messages). Channels, file-sharing, VIP StoreKit (WO-238) = post-MVP.

### Phase 4 — Blockchain & trust (MVP slice partial)

Trust/rewards ✅ (WO-22, 49, 114, 181, 213; wallet `/v3/wallet/*`). **WO-15** message anchoring ✅ (`AnchoringService`, merkle-proof API). Cardano WOs 12/20/37 removed.

### Phase 5 — Privacy (MVP slice ✅)

WO-38, 47, 56, 75, 125, 219, 317–319. Hidden-folder / scheduled-message / ZK / PQ = backlog.

### Phase 7 — Platform (rewards ✅; Comply + Passport active)

Rewards/wallet/staking largely ✅. **Comply:** WO-250–252, 308–313 ✅; WO-309 🔄. **Passport:** WO-293–296 ✅; WO-297 backlog.

Full gap matrix: [`PHASE4_7_GAP_AUDIT.md`](PHASE4_7_GAP_AUDIT.md).

---

## Go-live only (not feature gaps)

| Item | WO / doc |
|------|----------|
| Two-device Phase 3 E2E | [`E2E_TESTING.md`](E2E_TESTING.md) |
| `make validate-phase1` + wallet script | WO-230 |
| TestFlight build + regression | [`echo-phase1-validate`](../.cursor/skills/echo-phase1-validate/SKILL.md) |
| App Store security gates | WO-233, [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md) |
| VIP IAP (StoreKit) | WO-238 (Phase 6 backlog) |

---

## Remaining product tracks (post-messaging MVP)

### ECHO Comply (Phase 7)

**Foundation shipped:** WO-250–252, 289, 308–313.  
**In progress:** WO-309 (eDiscovery / matter UI).  
**Backlog (vertical + org):** WO-253–307, WO-281–286, WO-298–307 (FINRA/SEC blueprint + builds).

### ECHO Passport (Phase 2 Wave A + future waves)

| WO | Title | SF status |
|----|-------|-----------|
| 293–296 | Holder model, sync, SD-JWT, Shamir recovery | ✅ |
| **297** | **iOS Passport module** | **backlog** |
| 298–300+ | Metagraph anchors, pay-in-chat (Phase 4+) | backlog |

See [`ECHO_PASSPORT_PLAN.md`](ECHO_PASSPORT_PLAN.md).

---

## Software Factory phase tallies (2026-05-29)

| Phase | Completed | Backlog | Blocked | In progress |
|-------|-----------|---------|---------|-------------|
| 1 | 26 | 0 | 0 | 0 |
| 2 | 20 | 9 | 2 | 0 |
| 3 | 24 | 13 | 0 | 0 |
| 4 | 9 | 20 | 0 | 0 |
| 5 | 14 | 32 | 0 | 0 |
| 6 | 4 | 28 | 0 | 0 |
| 7 | 40 | 62 | 0 | 1 (WO-309) |

Phase 4–7 detail: [`PHASE4_7_GAP_AUDIT.md`](PHASE4_7_GAP_AUDIT.md).

---

## Related docs

- [`PHASE2_GAP_AUDIT.md`](PHASE2_GAP_AUDIT.md) — onboarding/credential gaps
- [`PHASE4_7_GAP_AUDIT.md`](PHASE4_7_GAP_AUDIT.md) — blockchain, privacy, calls, Comply/Passport
- [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) — signal wiring checklist
- [`WO-14_ONBOARDING_RESCOPE.md`](WO-14_ONBOARDING_RESCOPE.md) — onboarding UI source of truth
