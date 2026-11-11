# Echo — End-to-end testing (canonical)

**Single E2E source of truth** for automated gates, manual simulator/device QA, TestFlight bake, and messaging sign-off.

**Go-live calendar:** [`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md)  
**App Store P0–P2:** [`APP_STORE_SUBMISSION.md`](APP_STORE_SUBMISSION.md)  
**Agents:** skill `echo-testing` · MCP `echo-local-dev`

> Older files (`E2E_QUICK_START.md`, `E2E_LAUNCH_AND_TESTING.md`, `E2E_MESSAGING_SIGNOFF_CHECKLIST.md`, `WEEK_A_B_LAUNCH.md`, `TESTFLIGHT_WEEK_A_B.md`, `PHASE1_LAUNCH.md`) now **point here**. Prefer this doc for all new work.

**Frozen UX:** validate `FirstRunCoordinator` / `GlacialLoginScreen` only — do not redesign from the React prototype.

**Messaging tests, in order:** [§5](#5-messaging-tests--run-in-this-order) — (1) two Simulators on this Mac, (2) Simulator + your iPhone, (3) TestFlight upload, (4) two physical iPhones. [`E2E_TWO_DEVICE.md`](E2E_TWO_DEVICE.md) is the longer D0–D6 run after a stage is up (offline, restore, groups, logs).

---

## 0. Quick navigation

| Goal | Section |
|------|---------|
| Daily / milestone commands | [§1](#1-daily-regression-by-milestone) |
| First-time machine setup | [§2](#2-environment-setup) |
| **Start / stop / resume services** | [§2.4](#24-start-stop-resume) |
| **Clean slate (empty DB + new genesis)** | [§2.5](#25-clean-slate--empty-data--new-genesis) |
| Testing tiers & rules | [§3](#3-testing-tiers--gates) |
| Feature delivery matrix | [§4](#4-feature-delivery-matrix) |
| **1 — two Simulators on this Mac** | [§5.3](#53-stage-1--two-simulators-on-this-mac) |
| **2 — Simulator + your iPhone** | [§5.4](#54-stage-2--simulator--your-iphone) |
| **3 — deploy to TestFlight** | [§5.5](#55-stage-3--deploy-to-testflight) · [§9](#9-testflight--backend-deploy) |
| **4 — two physical iPhones** | [§5.6](#56-stage-4--two-physical-iphones) |
| Send the DM (every stage) | [§5.7](#57-find-each-other-and-send-a-dm) |
| Week A checklist (A1–A10) | [§5.8](#58-week-a-checklist-a1a10) |
| Offline / restore / groups / logs | [`E2E_TWO_DEVICE.md`](E2E_TWO_DEVICE.md) |
| Signal / Telegram parity matrices | [`SIGNAL_ECHO_PARITY.md`](SIGNAL_ECHO_PARITY.md) · [`TELEGRAM_ECHO_PARITY.md`](TELEGRAM_ECHO_PARITY.md) |
| Week B contacts (B1–B6) | [§6](#6-week-b--contacts--devices) |
| Phase 3 + ops + groups sign-off | [§7](#7-messaging-sign-off-matrix) |
| Device checks after stage 2 or 4 | [§8](#8-physical-iphone) |
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
| **Week A messaging** | DM / Phase 3 | `make phase3-signals-proof` + `make regression` | [§5 stage 1](#53-stage-1--two-simulators-on-this-mac) | Two-simulator DM + signals |
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

Identity L0/L1 are **not** part of Euclid `hydra` / `make dev`. Start them only when you need VC, Passport, or `validate-phase1` step 3 (trust-tier).

```bash
# Core cluster (Global L0 :9000, Metagraph L0 :9200, Currency L1 :9300, Data L1 :9400 + API :8000)
make dev
make dev-status

# Identity metagraph (optional — separate Docker compose)
cd metagraph && sbt identityL0/assembly identityL1/assembly   # once / after Scala changes
make start-identity   # Identity L0 :9600, Identity L1 :9500
# Stop: make stop-identity

# Then boot two iOS clients — §5
```

| Layer | URL | Started by |
|-------|-----|------------|
| Global L0 | `http://localhost:9000/node/info` | `make dev` (hydra) |
| Metagraph L0 | `http://localhost:9200/node/info` | `make dev` |
| Currency L1 | `http://localhost:9300/node/info` | `make dev` |
| Data L1 | `http://localhost:9400/node/info` | `make dev` |
| Identity L0 | `http://localhost:9600/node/info` | `make start-identity` |
| Identity L1 | `http://localhost:9500/node/info` | `make start-identity` |
| Backend | `http://localhost:8000/health` | `make dev` |

`start-identity` requires: assembly JARs under `metagraph/modules/identity_l*/target/…`, Docker image `metagraph-base-image:latest` (from `make dev`), and Global L0 already up on `:9000`.

**Stop / wipe:** [§2.4](#24-start-stop-resume) and [§2.5](#25-clean-slate--empty-data--new-genesis). Do not use `make clean` — that only deletes Go build artifacts.

Cluster deep dive: [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md).

Scheme: **EchoApp** / product **EchoMessaging**. Agents cannot tap Simulator UI — humans run [§5](#5-messaging-tests--run-in-this-order).

### 2.3 Common env flags

| Variable | Use |
|----------|-----|
| `API_URL` | Gateway base |
| `OIDC4VC_ENABLED` | Wallet enrollment E2E |
| `ECHO_IN_CHAT_PAYMENTS` | Must stay **off** for release |
| `ECHO_WALLET_GENESIS_AUTO` | **off** in prod/TF |

Full env table: historical detail retained in git history of `E2E_LAUNCH_AND_TESTING.md` §3e if needed; prefer `.env.example` / Compose files in repo.

`ECHO_WALLET_GENESIS_AUTO` is **token / wallet minting** for TestFlight — it is **not** Euclid `hydra start-genesis`. Leave it **off** for local E2E.

### 2.4 Start / stop / resume

Two stacks, two lifecycle tools:

| Stack | What | Start | Stop (keep data) |
|-------|------|-------|------------------|
| **Backend** | Gateway `:8000`, Postgres, Redis, NATS, MinIO | `make dev` or `make testnet-up` | `make dev-stop` |
| **Hydra metagraph** | Global L0 `:9000`, Metagraph L0 `:9200`, Currency L1 `:9300`, Data L1 `:9400` | `make dev` runs `hydra start-genesis` | `cd ../euclid-development-environment && scripts/hydra stop` |
| **Identity** (optional) | Identity L0 `:9600`, L1 `:9500` | `make start-identity` | `make stop-identity` |
| **Comply** (optional) | API `:8011`, Postgres `:5433` | `make dev-comply` | `make dev-comply-stop` |

**Everyday:**

```bash
make dev                 # first time today, or after a wipe
make dev-status          # what's up?
make start-identity      # only if you need VC / validate-phase1 step 3
make dev-logs            # backend logs

make dev-stop            # API + Postgres/Redis/MinIO down; hydra still running
make dev-stop-all        # backend + Identity + hydra stop — volumes and last snapshot kept
make dev-resume          # hydra start-rollback + backend — same chain, same DB
```

| You want | Command | Chain | Postgres / MinIO |
|----------|---------|-------|------------------|
| Pause overnight, keep users | `make dev-stop-all` then later `make dev-resume` | last snapshot | kept |
| Restart only the API | `make dev-restart` | unchanged | kept |
| Empty users **and** empty chain | [§2.5](#25-clean-slate--empty-data--new-genesis) | new genesis | wiped |

**Do not** run `make dev` when you meant “resume.” `make dev` always calls `hydra start-genesis` (erases **chain** history). If hydra was already up it may no-op; if hydra was stopped you get a **new chain** on top of the **old Postgres** — DIDs and snapshots will not match.

### 2.5 Clean slate — empty data + new genesis

Use this before a two-device run when leftover accounts, queued messages, or a stale chain are in the way.

```bash
make dev-reset           # wipe backend volumes + Identity volumes + hydra containers
make dev                 # hydra start-genesis (empty chain) + empty Postgres
# make start-identity    # only if you need VC / phase1 step 3 — this is a new Identity genesis too
make dev-status
```

What `make dev-reset` destroys:

| Data | Gone? |
|------|--------|
| Postgres users, threads, queue (`postgres-testnet-data`) | yes |
| Redis sessions | yes |
| MinIO backups / media | yes |
| Identity metagraph volume (`identity-shared` — keys + genesis address) | yes |
| Hydra containers (then `make dev` → `start-genesis`) | yes — **new** Global/Currency/Data genesis |
| Euclid Docker **images** (`hydra purge`) | **no** — next `make dev` stays fast |
| `.keys/identity-service.*` | **no** (optional: `rm -rf .keys/identity-service.*`) |
| Echo on Simulator / iPhone | **no** — delete the app or **Erase All Content and Settings** on each sim |

Then [§5](#5-messaging-tests--run-in-this-order): two **new** first-runs (`alice` / `bob`). Old `@usernames` from a previous wipe do not exist on this backend.

**Contributor copy** of the same wipe (manual compose flags): [`CONTRIBUTING.md` §5](../CONTRIBUTING.md#5-clean-slate-reset-all-data). Production / IntegrationNet genesis is **[`ZERO_TO_PRODUCTION.md`](ZERO_TO_PRODUCTION.md) §4** — do not run `dev-reset` against remote nodes.

---

## 3. Testing tiers & gates

| Tier | Where | When | Automatable |
|------|-------|------|-------------|
| **A — CI** | GitHub Actions | Every PR | ✅ |
| **B — Headless regression** | Mac + Xcode CLI | Weekly / pre-release | ✅ `make regression` |
| **C — Backend E2E** | Docker + Euclid | Metagraph/auth changes | ✅ `make validate-phase1` |
| **D — iOS build** | Xcode.app | After iOS changes | ✅ `make ios-preflight BUILD=1` |
| **E — Simulator manual** | Two iOS Simulators | Feature QA, first | ❌ [§5 stage 1](#53-stage-1--two-simulators-on-this-mac) |
| **F — One phone** | Simulator + iPhone | After stage 1 | ❌ [§5 stage 2](#54-stage-2--simulator--your-iphone) |
| **G — TestFlight** | Device + HTTPS | After stage 2 | ❌ [§5 stage 3](#55-stage-3--deploy-to-testflight) |
| **H — Two phones** | Two iPhones | After TestFlight (or LAN debug) | ❌ [§5 stage 4](#56-stage-4--two-physical-iphones) |

**Rule:** A–D green before E–H. Tier C must be full **GO** (no skips) for TestFlight sign-off.

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

## 5. Messaging tests — run in this order

Run the stages in order. Do not start on a phone or TestFlight until the previous stage’s DM passes ([§5.7](#57-find-each-other-and-send-a-dm)).

| Stage | What you prove | Hardware | Backend |
|-------|----------------|----------|---------|
| **1** | Two Echo apps talk on this laptop | Two iOS Simulators, Xcode | `make dev` · `http://localhost:8000` |
| **2** | A Simulator and your iPhone talk | One Simulator + one iPhone, same Wi-Fi | `make dev` · sim `localhost`, phone LAN IP |
| **3** | A signed build is on TestFlight | Xcode archive → App Store Connect | **HTTPS** staging, not `localhost` |
| **4** | Two real phones talk | Two iPhones | TestFlight (HTTPS), or LAN debug if you have not uploaded yet |

Every stage uses **two different accounts** (`alice` and `bob`). Leave onboarding as shipped (`FirstRunCoordinator` / `GlacialLoginScreen`).

Xcode project: `ios/Echo/EchoApp.xcodeproj`.

| You are | Scheme | Configuration |
|---------|--------|----------------|
| Stages 1, 2, and LAN stage 4 | **EchoApp** | **Debug-Messaging** |
| Stage 3 archive | **EchoMessaging** | **Release-Messaging** |

`API_URL` is injected when you **▶ Run from Xcode**. Tapping the home-screen icon after you unplug a phone drops the scheme env and falls back to the **Debug-Messaging** build setting (`http://localhost:8000`). That works on Simulator and **fails on a phone**.

---

### 5.1 Start the backend (once)

```bash
# Empty world (optional — leftover alice/bob or a stale chain):
# make dev-reset

make dev
make dev-status          # Backend :8000 must be ✓
curl -fsS http://localhost:8000/health
# expect: {"status":"operational", ...}

make phase3-signals-proof
make ios-preflight BUILD=1
```

Identity L0/L1 (`make start-identity`) is **not** required for a DM. Start / stop / wipe: [§2.4](#24-start-stop-resume) / [§2.5](#25-clean-slate--empty-data--new-genesis).

---

### 5.2 Set `API_URL` in Xcode

**Product → Scheme → Edit Scheme… → Run → Arguments → Environment Variables**

| Client | `API_URL` value |
|--------|-----------------|
| Every Simulator | `http://localhost:8000` (already in the shared EchoApp scheme) |
| Every physical iPhone | `http://<Mac-LAN-IP>:8000` |

Mac LAN IP (Wi-Fi):

```bash
ipconfig getifaddr en0
# e.g. 192.168.1.42
```

If you will unplug the phones and tap the icon, also set the **EchoApp** target build setting `API_URL` (Debug-Messaging) to that same LAN URL, then rebuild. Otherwise keep both Xcode Run sessions attached.

---

### 5.3 Stage 1 — two Simulators on this Mac

Do this first. Both clients use `http://localhost:8000`. No phone, no Apple upload.

1. Open **Simulator.app**. Boot **two different** models (File → Open Simulator, or):

   ```bash
   xcrun simctl list devices available
   xcrun simctl boot "iPhone 16"
   xcrun simctl boot "iPhone 16 Pro"
   open -a Simulator
   ```

   Two windows. Same model twice is fine if they are two device entries.

2. Xcode → scheme **EchoApp** → destination **first** iPhone → ▶ Run (`⌘R`).
3. Finish first-run as **alice**. Land on the **Messages** tab (that opens the WebSocket).
4. **Do not stop** that run. Toolbar destination → **second** iPhone → **hold Option and click Run** (or **Product → Perform Action → Run**). Xcode 16: **Window → Devices and Simulators** / run list should show two processes.
5. Finish first-run as **bob**. Land on Messages.
6. Go to [§5.7](#57-find-each-other-and-send-a-dm). Pass that before stage 2.

If Option-Run stops the first client: Edit Scheme → Run → Options and enable multiple simultaneous runs if you see that checkbox, **or** install on both sims then ▶ Run each destination in turn without clicking Stop.

---

### 5.4 Stage 2 — Simulator + your iPhone

Start here only after stage 1’s DM passed. One account stays on the Simulator. The other is your phone, on the **same Wi-Fi** as the Mac, still talking to `make dev`.

1. Simulator: leave `API_URL` as `http://localhost:8000`. Scheme **EchoApp**. ▶ Run as **alice** if that sim is still enrolled; otherwise first-run as **alice**.
2. Mac LAN IP: `ipconfig getifaddr en0`. Prove the phone can reach it **before** you launch the app:

   ```bash
   curl -fsS "http://$(ipconfig getifaddr en0):8000/health"
   ```

   If `localhost:8000` works and this fails, Docker is loopback-only. Stop. Phones will never connect.

   On the iPhone, Safari: `http://<LAN-IP>:8000/health`. Pass = JSON `operational`.
3. Edit Scheme → Run → `API_URL` = `http://<LAN-IP>:8000`. Plug in the phone, pick it as the destination, ▶ Run, first-run as **bob**.
4. You cannot use one scheme env for both in the same launch. Run the phone with the LAN URL first. The running phone keeps that env. Then set the scheme back to `http://localhost:8000` and run the Simulator if it is not already up.
5. Trust the developer cert if iOS asks (Settings → General → VPN & Device Management).
6. Go to [§5.7](#57-find-each-other-and-send-a-dm). Pass that before stage 3.

`Configs/EchoMessaging-Info.plist` allows local-network HTTP (`NSAllowsLocalNetworking`) so Debug builds can use `http://192.168…:8000`. A TestFlight build cannot. It needs HTTPS ([§9](#9-testflight--backend-deploy)).

---

### 5.5 Stage 3 — deploy to TestFlight

Start here only after stage 2’s DM passed. This stage uploads a build. It does not chat over `localhost`.

Follow [§9](#9-testflight--backend-deploy). Short version:

1. Staging gateway is **HTTPS and WSS** from cellular. `ECHO_WALLET_GENESIS_AUTO` stays **off**.
2. On the **EchoApp** target, set build setting `API_URL` for **Release-Messaging** to that `https://…` host. Info.plist reads `$(API_URL)`. If you skip this, the archive falls through to `https://api.echo.local`.
3. Scheme **EchoMessaging** → **Product → Archive** (Release-Messaging) → Organizer → **Distribute App** → App Store Connect → Upload. Bundle id `com.echo.app`.
4. App Store Connect → TestFlight → add yourself as an internal tester → install on the phone from the TestFlight app.
5. Run [§9.3](#93-per-build-tf-smoke-g) on that install. Pass that before stage 4.

Signing, export plist, and the tag path are in [§9.1](#91-archive). Production cluster bring-up is [`ZERO_TO_PRODUCTION.md`](ZERO_TO_PRODUCTION.md), not this section.

---

### 5.6 Stage 4 — two physical iPhones

Start here last.

**After TestFlight (the stage-4 gate).** Install the same build from TestFlight on **both** phones. Both use the HTTPS host baked in at archive time. Complete first-run as **alice** and **bob** (new installs). Then [§5.7](#57-find-each-other-and-send-a-dm) over cellular or Wi-Fi. `localhost` and the Mac LAN IP are not involved.

**LAN debug, only if you have not uploaded yet.** Two phones, Mac, same Wi-Fi, scheme **EchoApp**, `API_URL=http://<Mac-LAN-IP>:8000` on both, Xcode ▶ Run then Option-Run. Safari `/health` on each phone must pass first (same curl as stage 2). This proves hardware against `make dev`. It does not prove the TestFlight build.

---

### 5.7 Find each other and send a DM

Both clients must already be past first-run, on the **Messages** tab, against the **same** `:8000`.

1. **Alice:** compose / new conversation → search `@bob` (or Contacts → username search). Open the thread.  
   Alternate: Alice Profile QR → Bob scans; or share `echo://invite?u=alice` into Bob’s Accept-invite sheet.
2. **Alice:** send `hello`.
3. **Bob:** open the same thread (search `@alice` if it is not in the hub yet).
4. **Pass:** Bob sees `hello` without pulling to refresh. Bob replies. Alice sees the reply live.

Thread id must be `dm:{sorted-did}:{sorted-did}` (`ContactThreadHelper`). If the thread never appears, they are not contacts yet — finish step 1.

Then run the [Week A table](#58-week-a-checklist-a1a10). For offline delivery, phrase restore, groups, APNs, and log capture, continue in [`E2E_TWO_DEVICE.md`](E2E_TWO_DEVICE.md) (D3–D6).

---

### 5.8 Week A checklist (A1–A10)

```bash
make dev
make phase3-signals-proof
make ios-preflight BUILD=1
```

| Step | User A | User B | Pass |
|------|--------|--------|------|
| **A1** | Complete onboarding | Complete onboarding | Both authenticated, distinct `@usernames` |
| **A2** | New conversation → search B `@username` | — | Thread opens |
| **A3** | Send “hello” | — | A shows sent |
| **A4** | — | Open same thread | B sees message |
| **A5** | — | Reply while A chat open | A sees reply live |
| **A6** | Type in field | Chat open | Typing indicator |
| **A7** | Long-press → 👍 | — | Reaction chip both sides |
| **A8** | Privacy → typing/receipts off | Repeat | Signals suppressed |
| **A9** | Reopen chat | — | History persists |
| **A10** | Background / relaunch | — | Thread still listed; unread sane |

**Wiring (code):** `ChatDetailViewModel` + shared WS in `MessagesTabView`; privacy merge via `MessagingPrivacyPreferences`. Skill `echo-phase3-ios-wire`.

---

### 5.9 If it does not work

| Symptom | Fix |
|---------|-----|
| Phone cannot reach API; sim can | `API_URL` is still `localhost`. Use the Mac LAN IP ([§5.2](#52-set-api_url-in-xcode)). Confirm Safari `/health` on the phone. |
| Safari `/health` fails on the phone | Same Wi-Fi? `curl http://$(ipconfig getifaddr en0):8000/health` on the Mac first. If that fails, Docker is loopback-only or the firewall is blocking `:8000`. |
| Both clients look like the same person | Same `@username` or same first-run state. Delete the app on one client and enroll again. |
| Message never arrives | Both on **Messages** (WS connects there). Backend: `make dev-logs` — look for `ws upgrade`, not `ws upgrade error`. |
| Option-Run kills the first sim | Do not click Stop. Use Option-Run or a second destination; see [§5.3](#53-stage-1--two-simulators-on-this-mac) step 4. |
| Tapping the icon on a phone after unplug fails | Scheme env is gone. Rebuild with Debug-Messaging `API_URL` = LAN IP, or keep Xcode attached. |
| Locked phone never wakes | APNs not configured — still pass live DM. Offline-on-open is [`E2E_TWO_DEVICE.md`](E2E_TWO_DEVICE.md) D3. |

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

Stages 2 and 4 are the procedures. This table is the extra device checks after a phone is already talking.

**One phone against `make dev`:** [§5.4](#54-stage-2--simulator--your-iphone).  
**Two phones:** [§5.6](#56-stage-4--two-physical-iphones).  
**Upload:** [§9](#9-testflight--backend-deploy).

| Check | Notes |
|-------|-------|
| `API_URL` | LAN `http://<Mac-IP>:8000` for local; **HTTPS** for TestFlight — never `localhost` on device |
| Face ID / passkey | Real Secure Enclave |
| Background WS resume | Foreground the app → chat continues |
| Push (if configured) | APNs sandbox vs production matches the build |
| Photo / mic permissions | Real OS prompts |
| Two physical devices | Preferred for TestFlight bake ([§9.3](#93-per-build-tf-smoke-g)) |

---

## 9. TestFlight & backend deploy

> **Full production deploy procedure:** [`ZERO_TO_PRODUCTION.md`](ZERO_TO_PRODUCTION.md) — backend on
> Hetzner k3s (§2), Constellation metagraph genesis→integrationnet→mainnet (§4), iOS signed export (§3).
> This section is the TestFlight-bake checklist only.

### 9.1 Archive

Do this only after [stage 2](#54-stage-2--simulator--your-iphone) passed.

1. Xcode → open `ios/Echo/EchoApp.xcodeproj`.
2. Select your Team on the **EchoApp** target (Signing & Capabilities) so Release-Messaging can sign `com.echo.app`.
3. Set the **Release-Messaging** build setting `API_URL` to the public HTTPS gateway (not `http://localhost:8000`).
4. Scheme **EchoMessaging** → **Product → Archive**.
5. Organizer → **Distribute App** → App Store Connect → Upload.
6. Or build an unsigned archive in CI, then sign locally:

   ```bash
   ./scripts/release/product-tag.sh echo-messaging <ver>
   git push origin echo-messaging@v<ver>
   ```

   Download the `.xcarchive` from Actions. Set the Team ID in `ios/ExportOptions.plist`, then:

   ```bash
   xcodebuild -exportArchive -archivePath Echo.xcarchive \
     -exportOptionsPlist ios/ExportOptions.plist -exportPath build/export
   ```

   Upload the `.ipa` with Xcode Organizer or `xcrun altool --upload-app`. Tag details: [`PRODUCT_LAUNCH.md`](PRODUCT_LAUNCH.md). Cluster bring-up: [`ZERO_TO_PRODUCTION.md`](ZERO_TO_PRODUCTION.md) §2–§3.

### 9.2 Backend for TF

- HTTPS + WSS reachable from cellular
- Secrets rotated; genesis wallet auto **off**
- `/health` monitored

### 9.3 Per-build TF smoke (G)

- [ ] Fresh install → onboarding → DM with second TF user
- [ ] Typing / reaction smoke
- [ ] No crash on cold start
- [ ] Legal links open
- [ ] Rewards tab loads (v1: earn, stake, claim; no “buy ECHO”)
- [ ] ECHO VIP subscribe sheet is the App Store purchase for `com.echo.app.vip.monthly`

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
| `make dev` then `make dev-status` before any messaging check. Backend `:8000` must answer `/health`. | Simulator UI, Face ID, two-client chat |
| `make regression*`, phase1, Phase3 unit tests | |
| `ios-preflight`, SPM build | Target membership, signing, Archive |
| `smoke_ios_backend` | TestFlight install, ASC forms |

---

## 12. Document map (merged sources)

| Former doc | Absorbed into |
|------------|---------------|
| `E2E_QUICK_START.md` | §1–§2 |
| `E2E_LAUNCH_AND_TESTING.md` | §3–§4, §8–§10 |
| `WEEK_A_B_LAUNCH.md` / `TESTFLIGHT_WEEK_A_B.md` | §5–§6 (two-client boot + Week A/B) |
| `E2E_MESSAGING_SIGNOFF_CHECKLIST.md` | §7 |
| `PHASE1_LAUNCH.md` | §4 Phase 1 |
| `metagraph-backend-e2e-testing.md` | Linked (cluster-only depth) |
| `TESTING.md` | Prefer this file + `make regression` |

For Sept 1 planning (corporate, competitive, calendar): **[`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md)**.
