# Echo — End-to-end testing (canonical)

**Single E2E source of truth** for automated gates, manual simulator/device QA, TestFlight bake, and messaging sign-off.

**Go-live calendar:** [`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md)  
**App Store P0–P2:** [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md)  
**Agents:** skill `echo-testing` · MCP `echo-local-dev`

> Older files (`E2E_QUICK_START.md`, `E2E_LAUNCH_AND_TESTING.md`, `E2E_MESSAGING_SIGNOFF_CHECKLIST.md`, `WEEK_A_B_LAUNCH.md`, `TESTFLIGHT_WEEK_A_B.md`, `PHASE1_LAUNCH.md`) now **point here**. Prefer this doc for all new work.

**Frozen UX:** validate `FirstRunCoordinator` / `GlacialLoginScreen` only — do not redesign from the React prototype.

---

## 0. Quick navigation

| Goal | Section |
|------|---------|
| Daily / milestone commands | [§1](#1-daily-regression-by-milestone) |
| First-time machine setup | [§2](#2-environment-setup) |
| Testing tiers & rules | [§3](#3-testing-tiers--gates) |
| Feature delivery matrix | [§4](#4-feature-delivery-matrix) |
| Week A messaging (A1–A10) | [§5](#5-week-a--messaging-manual) |
| Week B contacts (B1–B6) | [§6](#6-week-b--contacts--devices) |
| Phase 3 + ops + groups sign-off | [§7](#7-messaging-sign-off-matrix) |
| Physical device | [§8](#8-physical-iphone) |
| TestFlight & deploy | [§9](#9-testflight--backend-deploy) |
| Release sign-off template | [§10](#10-release-sign-off-template) |
| Metagraph cluster deep dive | [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) |

---

## 1. Daily regression by milestone

Run the row for your focus **before** Xcode or TestFlight upload.

| Milestone | When | Automated (order) | Manual | Pass if |
|-----------|------|-------------------|--------|---------|
| **Stack health** | Every morning | `make dev` → `make dev-status` | — | Endpoints ✓; backend operational |
| **Backend PR** | Every Go change | `make release-check` | — | Build + race + vet + fmt |
| **Metagraph PR** | Every Scala change | `make metagraph-test` | — | L1 tests green |
| **Phase 1 go/no-go** | Pre-TF / auth-chain changes | `make regression-with-phase1` | — | `validate-phase1` → **GO** |
| **iOS compile** | Every iOS change | `make ios-preflight BUILD=1 TESTS=1` | — | Zero `FAIL` |
| **Screen catalog** | UI changes | `make screen-catalog` | Open `docs/screen_catalog/index.html` | PNGs render |
| **Week A messaging** | DM / Phase 3 | `make phase3-signals-proof` + `make regression` | [§5 A1–A10](#5-week-a--messaging-manual) | Two-client DM + signals |
| **Week B contacts** | After Week A | `make regression` | [§6 B1–B6](#6-week-b--contacts--devices) | Invite, QR, link-device |
| **Wallet / staking** | Balance / claim | `./scripts/validate-wallet.sh` | Rewards → stake → claim | Real balance |
| **Pre-TestFlight** | Before archive | `make regression-with-phase1` | §7 + §10 | Sign-off complete |
| **App Store** | External TF | — | [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md) P0 | All P0 closed |

**PR bar:**

```bash
make fmt && make lint && make vet && make test
cd metagraph && sbt scalafmtCheckAll && sbt test && cd ..
make validate-phase1    # or document skips if simulator-only PR
```

**MCP shortcuts:** `run_ios_preflight`, `run_regression`, `run_validate_phase1`, `run_ios_phase3_tests`, `health_backend`, `cluster_status`, `smoke_ios_backend`.

---

## 2. Environment setup

### 2.1 Prerequisites

| Tool | Version |
|------|---------|
| macOS 14+ (iOS) / Ubuntu 22.04+ (backend OK) | — |
| JDK **21** (Temurin) | Tessellation |
| sbt 1.9+, Scala 2.13.10 | Metagraph |
| Go 1.21+ | Backend |
| Docker Compose v2 | **≥8 GB RAM, ≥4 CPUs** |
| Xcode 15+ | iOS (`ios/Echo/`) |
| `jq`, `yq`, `argc`, Ansible 2.16+, g8 | Euclid |

`metagraph/scripts/setup-euclid.sh` reports missing tools.

### 2.2 Start stack

```bash
make dev                 # metagraph + gateway :8000
make dev-status          # health
# iOS Simulator API_URL (scheme): http://localhost:8000
# Physical device: http://<LAN-IP>:8000
```

Scheme: **EchoApp** / product **EchoMessaging**. Agents cannot tap Simulator UI — humans run §5–§8.

### 2.3 Common env flags

| Variable | Use |
|----------|-----|
| `API_URL` | Gateway base |
| `OIDC4VC_ENABLED` | Wallet enrollment E2E |
| `ECHO_IN_CHAT_PAYMENTS` | Must stay **off** for release |
| `ECHO_WALLET_GENESIS_AUTO` | **off** in prod/TF |

Full env table: historical detail retained in git history of `E2E_LAUNCH_AND_TESTING.md` §3e if needed; prefer `.env.example` / Compose files in repo.

---

## 3. Testing tiers & gates

| Tier | Where | When | Automatable |
|------|-------|------|-------------|
| **A — CI** | GitHub Actions | Every PR | ✅ |
| **B — Headless regression** | Mac + Xcode CLI | Weekly / pre-release | ✅ `make regression` |
| **C — Backend E2E** | Docker + Euclid | Metagraph/auth changes | ✅ `make validate-phase1` |
| **D — iOS build** | Xcode.app | After iOS changes | ✅ `make ios-preflight BUILD=1` |
| **E — Simulator manual** | Xcode Simulator | Feature QA | ❌ §5–§7 |
| **F — Device manual** | Physical iPhone | Pre-TF | ❌ §8 |
| **G — TestFlight** | Device + HTTPS | Per upload | ❌ §9 |

**Rule:** A–D green before E–G. Tier C must be full **GO** (no skips) for TestFlight sign-off.

---

## 4. Feature delivery matrix

### Phase 1 — Foundation (WO-230)

| Set | Capability | Proof |
|-----|------------|-------|
| P1-Identity | `did:key` + trust tier on Identity L1 | validate-phase1 1–3 |
| P1-Relay | WS send/receive | Step 4 |
| P1-Data-L1 | Merkle `/data` | Step 5 |
| P1-Chain | Global L0 snapshot | Step 6 |
| P1-Security | Passkey REST + T0–T7 | `make release-check` |

```bash
make regression-with-phase1   # must print Go/No-Go: GO
```

### Week A — Messaging

| Set | Capability | Automated | Manual |
|-----|------------|-----------|--------|
| A-DM | `dm:` threads; send/receive | `phase3-signals-proof` | A1–A5 |
| A-Chat-UI | ChatView, composer | `ios-preflight BUILD=1` | A2–A5 |
| A-Phase3 | Typing, receipts, reactions | `EchoPhase3Tests` | A6–A9 |
| A-Persist | Local history | — | A10 |

### Week B — Contacts

| Set | Manual |
|-----|--------|
| Invite / QR / detail / SMS / social / link-device | B1–B6 |
| PSI (optional) | `make echooprf-ios` + two devices |

### Optional when in scope

| Set | Proof |
|-----|-------|
| OIDC4VC | Enrollment flow + `OIDC4VC_ENABLED=true` |
| Groups / calls / media / sync / backup | [§7](#7-messaging-sign-off-matrix) M2–M7 |
| Channels beta | Create/list/post only; label non-durable |

Mark **N/A** on sign-off for features not in the release build.

---

## 5. Week A — Messaging manual

**Prep**

```bash
make dev
make phase3-signals-proof
make ios-preflight BUILD=1
```

Two simulators or devices, both signed in against same backend. Thread IDs must be `dm:{sorted-did}:{sorted-did}` via `ContactThreadHelper`.

| Step | User A | User B | Pass |
|------|--------|--------|------|
| **A1** | Complete onboarding | Complete onboarding | Both authenticated |
| **A2** | New conversation → search B `@username` | — | Thread opens |
| **A3** | Send “hello” | — | A shows sent |
| **A4** | — | Open same thread | B sees message |
| **A5** | — | Reply while A chat open | A sees reply live |
| **A6** | Type in field | Chat open | Typing indicator |
| **A7** | Long-press → 👍 | — | Reaction chip both sides |
| **A8** | Privacy → typing/receipts off | Repeat | Signals suppressed |
| **A9** | Reopen chat | — | History persists |
| **A10** | Background / relaunch | — | Thread still listed; unread sane |

**Wiring checklist (code):** `ChatDetailViewModel` + shared WS in `MessagesTabView`; privacy merge via `MessagingPrivacyPreferences`. See skill `echo-phase3-ios-wire`.

---

## 6. Week B — Contacts & devices

| Step | Verify | Pass |
|------|--------|------|
| **B1** | `echo://invite` cold start → post-login sheet | Contact added |
| **B2** | Profile QR → add contact | `POST /v3/contacts/add` |
| **B3** | Contact detail → Message / block / favorite | Hub thread + API |
| **B4** | Add phone backup (`SMSOTPSetupView`) | OTP path works or scoped N/A |
| **B5** | Mutual groups / contacts on detail | Shows when shared membership |
| **B6** | Settings → Link new device QR → scan on login | Second device session |

**Optional:** live OPRF PSI (`make echooprf-ios`); OIDC4VC enrollment.

---

## 7. Messaging sign-off matrix

Run after Week A green. Owner: ________ Date: ________

### Prerequisites

| # | Check | ☐ |
|---|--------|---|
| P1 | `make dev` (or LAN/prod-like API) | |
| P2 | Two signed-in clients A + B | |
| P3 | Xcode build EchoMessaging / EchoApp | |
| P4 | Preflight below green | |

```bash
make release-check
go test ./internal/api/ -run 'Signal|Call|Media|Overflow|Archive|Poll|Sync|Backup'
cd ios/Echo && swift test --filter 'EchoPhase3Tests|EchoSecurityTests'
```

| P5 | WebRTC SPM linked if testing calls ([`ios/WEBRTC_XCODE_SETUP.md`](../ios/WEBRTC_XCODE_SETUP.md)) | |
| P6 | Mic/camera permissions if calls | |

### §7.1 Phase 3 signals (M0)

| # | Step | Expected | ☐ |
|---|------|----------|---|
| 4.1 | A types | B sees typing | |
| 4.2 | A stops ~6s | Clears | |
| 4.3 | B opens / views | A delivered → read | |
| 4.4 | A reacts 👍 | B sees chip; REST/WS match | |
| 4.5 | Non-participant C | No WS signal | |

### §7.2 Message operations (M1)

| # | Step | Expected | ☐ |
|---|------|----------|---|
| 4b.1 | Edit | Peer sees update | |
| 4b.2 | Delete | Tombstone both | |
| 4b.3 | Pin (≤5) | Sync | |
| 4b.4 | Disappearing timer | Expire both | |
| 4b.5 | Reply / forward | Metadata OK | |

### §7.3 Groups (M2)

| # | Step | Expected | ☐ |
|---|------|----------|---|
| 9.1 | Create group + B | Thread opens | |
| 9.2 | Key to B | B can send | |
| 9.3–9.4 | Bidirectional msgs | Decrypt OK | |
| 9.5–9.6 | Remove + rekey | Removed cannot decrypt | |
| 9.7 | Offline queue | Deliver on reconnect | |

### §7.4 Multi-device sync (M3) — N/A for Sept 1 if WO-CA3 not shipped

Mark N/A unless build includes history sync. Link-device login alone is **B6**, not full history sync.

### §7.5 Calls / media / backup — as scoped

Use historical detailed steps in git (`E2E_MESSAGING_SIGNOFF_CHECKLIST.md`) for M4–M7 when those waves are in the RC. For Sept 1 soft launch, minimum is M0 + M1 core + M2 basic unless calls are marketed.

---

## 8. Physical iPhone

| Check | Notes |
|-------|-------|
| Scheme `API_URL` → LAN or HTTPS staging | Not `localhost` |
| Face ID / passkey | Real Secure Enclave |
| Background WS resume | App foreground restores chat |
| Push (if configured) | APNs env correct |
| Photo / mic permissions | Real OS prompts |
| Two physical devices preferred for TF bake | |

---

## 9. TestFlight & backend deploy

> **Full production deploy procedure:** [`ZERO_TO_PRODUCTION.md`](ZERO_TO_PRODUCTION.md) — backend on
> Hetzner k3s (§2), Constellation metagraph genesis→integrationnet→mainnet (§4), iOS signed export (§3).
> This section is the TestFlight-bake checklist only.

### 9.1 Archive

1. Xcode → scheme **EchoMessaging** → **Release-Messaging**
2. Archive → upload to ASC `com.echo.app`
3. Or tag: `./scripts/release/product-tag.sh echo-messaging <ver>` ([`PRODUCT_LAUNCH.md`](PRODUCT_LAUNCH.md))

### 9.2 Backend for TF

- HTTPS + WSS reachable from cellular
- Secrets rotated; genesis wallet auto **off**
- `/health` monitored

### 9.3 Per-build TF smoke (G)

- [ ] Fresh install → onboarding → DM with second TF user
- [ ] Typing / reaction smoke
- [ ] No crash on cold start
- [ ] Legal links open
- [ ] Rewards tab loads without purchase CTA

---

## 10. Release sign-off template

**Build:** ________ **API:** ________ **Date:** ________ **Signer:** ________

| Gate | Result |
|------|--------|
| `make regression-with-phase1` | GO / NO-GO |
| Week A A1–A10 | ☐ |
| Week B (scoped) | ☐ |
| §7.1 Phase 3 | ☐ |
| §7.2 Ops (scoped) | ☐ |
| §7.3 Groups (scoped) | ☐ |
| App Store P0 | ☐ |
| Corporate legal URLs live | ☐ |
| Marketing claims reviewed ([`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md) §8) | ☐ |

**Ship decision:** ☐ Internal TF · ☐ External TF · ☐ Soft launch · ☐ Hold

---

## 11. Agent vs human ownership

| Agent (headless) | Human (Xcode / devices) |
|------------------|-------------------------|
| `make regression*`, phase1, Phase3 unit tests | Simulator UI, Face ID, two-client chat |
| `ios-preflight`, SPM build | Target membership, signing, Archive |
| `smoke_ios_backend` | TestFlight install, ASC forms |

---

## 12. Document map (merged sources)

| Former doc | Absorbed into |
|------------|---------------|
| `E2E_QUICK_START.md` | §1–§2 |
| `E2E_LAUNCH_AND_TESTING.md` | §3–§4, §8–§10 |
| `WEEK_A_B_LAUNCH.md` / `TESTFLIGHT_WEEK_A_B.md` | §5–§6 |
| `E2E_MESSAGING_SIGNOFF_CHECKLIST.md` | §7 |
| `PHASE1_LAUNCH.md` | §4 Phase 1 |
| `metagraph-backend-e2e-testing.md` | Linked (cluster-only depth) |
| `TESTING.md` | Prefer this file + `make regression` |

For Sept 1 planning (corporate, competitive, calendar): **[`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md)**.
