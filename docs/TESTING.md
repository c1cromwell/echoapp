# Echo Testing Guide — Launch & Regression

Single reference for **automated gates** (CI, scripts, agents) and **manual E2E** on Mac/Xcode (simulator + physical iPhone + TestFlight). Use this before every TestFlight build and for periodic regression.

**Quick links**

| Goal | Start here |
|------|------------|
| Pre-merge / weekly regression | [§2 Automated regression](#2-automated-regression-headless) → `make regression` |
| Phase 1 TestFlight launch | [§2](#2-automated-regression-headless) + [§5–7](#5-manual-e2e--xcode-ios-simulator) + [`PHASE1_LAUNCH.md`](PHASE1_LAUNCH.md) |
| Metagraph + backend E2E | [§3](#3-backend--metagraph-e2e) + [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) |
| Phase 3 messaging signals | [§5.4](#54-phase-3--typing-read-receipts-reactions-wo-192) + [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) |
| OIDC4VC enrollment (WO-100) | [§5.5](#55-oidc4vc-wallet-enrollment-wo-100) |
| PSI contact discovery (WO-221) | [§5.6](#56-psi-contact-discovery-wo-221) |
| Agent automation | Skill **`.cursor/skills/echo-testing`** + MCP `echo-local-dev` |

---

## 1. Testing tiers

| Tier | Runs where | When | Automatable |
|------|------------|------|-------------|
| **A — CI** | GitHub Actions | Every push/PR to `main` / `develop` | ✅ Fully |
| **B — Headless regression** | Mac or Linux (Go); Mac + Xcode (iOS SPM) | Pre-release, weekly | ✅ `make regression` |
| **C — Backend E2E** | Mac + Docker + Euclid | Before TestFlight; after metagraph changes | ✅ `make validate-phase1` |
| **D — iOS build verify** | Mac + Xcode.app | After iOS project changes | ✅ `xcodebuild` |
| **E — Simulator manual E2E** | Mac + Xcode Simulator | Before TestFlight; feature QA | ❌ Checklists §5 |
| **F — Device manual E2E** | Physical iPhone(s) | Before TestFlight; biometrics, PSI, LAN | ❌ Checklists §6 |
| **G — TestFlight regression** | Physical device + prod/staging API | After each uploaded build | ❌ Checklist §7 |

**Rule:** Tiers A–D must be green before running E–G. Tier C may `skip` metagraph steps for simulator-only work; **all steps `ok`** required for launch go/no-go.

---

## 2. Automated regression (headless)

### 2a. One command (recommended)

```bash
# From repo root — Go release-check + targeted suites + iOS SPM tests
make regression

# Go race tests only (no iOS, no validate-phase1)
make regression-quick

# Also run WO-230 validate-phase1 (needs Docker + Euclid cluster)
make regression-with-phase1
```

Underlying script: `scripts/run-regression.sh` (`--quick`, `--with-phase1`, `--ios-only`).

### 2b. What `make regression` includes

| Step | Command / target | Hard gate? |
|------|------------------|------------|
| Go build + race tests + vet + fmt | `make release-check` | ✅ |
| Enrollment VC, WS, reactions, contacts | `go test ./internal/api/ … ./internal/services/contacts/… ./mobile/echooprf/…` | ✅ |
| iOS library + security tests | `swift build --target Echo`, `EchoSecurityTests` | ✅ (needs Xcode.app) |
| Phase 3 logic tests | `swift test --filter EchoPhase3Tests` | ✅ |
| OPRF mobile interop | `go test ./mobile/echooprf/...` | ✅ |

### 2c. CI (Tier A — runs without your Mac)

| Workflow | Path | Gate |
|----------|------|------|
| Go tests | `.github/workflows/` (Go CI) | `go test -race ./internal/... ./pkg/...` |
| iOS SPM | `.github/workflows/ios-ci.yml` | `EchoSecurityTests`, `EchoPhase3Tests` |
| T0–T7 data classification | Semgrep | `.semgrep/t0_t7_rules.yaml` |

```bash
semgrep --config .semgrep/t0_t7_rules.yaml --error .
```

### 2d. MCP automation (Cursor agents)

Server: **`echo-local-dev`** (see `tools/echo-local-dev-mcp/README.md`).

| Tool | Equivalent |
|------|------------|
| `run_release_check` | `make release-check` |
| `run_validate_phase1` | `make validate-phase1` |
| `run_ios_phase3_tests` | `swift test --filter EchoPhase3Tests` |
| `health_backend` | `curl localhost:8000/health` |
| `cluster_status` | `make dev-status` |

Agent skill: **`.cursor/skills/echo-testing`**.

---

## 3. Backend + metagraph E2E

### 3a. Start stack

```bash
cp .env.example .env    # first time
make dev
make dev-status
curl -s http://localhost:8000/health | jq .status   # → "operational"
```

Optional Identity nodes (VC / trust tier):

```bash
make start-identity
```

### 3b. WO-230 go/no-go (6 steps)

```bash
make validate-phase1
```

Expected: all steps `ok` for launch. Steps 3/5 may `skip` without Euclid — acceptable for simulator-only, **not** for TestFlight sign-off.

Details: [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md), skill **`echo-phase1-validate`**.

### 3c. OIDC4VC backend (WO-100)

In `.env`:

```bash
OIDC4VC_ENABLED=true
OIDC4VC_VERIFIER_BASE_URL=http://localhost:8000   # or public URL
```

Smoke:

```bash
curl -s -o /dev/null -w "%{http_code}\n" \
  -X POST http://localhost:8000/v1/enrollment/vc/start \
  -H 'Content-Type: application/json' \
  -d '{"requested_claims":{"givenName":true,"familyName":true,"ageOver18":true}}'
# → 200 with session_id + wallet URL when verifier mounted
```

Unit tests: `go test ./internal/api/ -run EnrollmentVC -count=1`

---

## 4. iOS automated (Mac + Xcode)

Requires **Xcode.app** (not Command Line Tools only).

### 4a. SPM tests (same as CI)

```bash
cd ios/Echo
export DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer

swift build --target Echo --target EchoSecurityTests --target EchoPhase3Tests
swift test --filter EchoSecurityTests
swift test --filter EchoPhase3Tests
```

| Target | Work orders | Role |
|--------|-------------|------|
| `EchoSecurityTests` | WO-208/211/223/224 | Crypto, Secure Enclave, passkey |
| `EchoPhase3Tests` | WO-192 | Typing, receipts, reactions logic |
| `EchoTests` | Legacy | Advisory only — may not compile |

### 4b. EchoApp target build (compile gate)

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj -scheme Echo \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  -configuration Debug build
```

### 4c. Live OPRF framework (WO-221)

```bash
./scripts/build-echooprf-ios.sh
```

Then in Xcode: EchoApp → **Frameworks** → add `ios/Echo/Libraries/EchoOPRF.xcframework` → **Embed & Sign**.

Without the framework, `OPRFClientFactory` falls back to `MockOPRFClient` (dev only — not valid for PSI launch QA).

### 4d. Archive (pre-TestFlight compile)

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj \
  -scheme Echo \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination 'generic/platform=iOS' \
  archive
```

Full upload flow: [`PHASE1_LAUNCH.md` §4](PHASE1_LAUNCH.md).

---

## 5. Manual E2E — Xcode (iOS Simulator)

**Prerequisites:** `make dev` running; Simulator with Face ID enrolled (Features → Face ID → Enrolled).

Set scheme env / build setting `API_URL` = `http://localhost:8000` (simulator can reach host localhost).

Open `ios/Echo/EchoApp.xcodeproj` → scheme **Echo** → Run on **iPhone 16 Pro** simulator.

### 5.1 Onboarding (new user)

| # | Action | Pass |
|---|--------|------|
| 1 | Cold launch (no account) | `EchoWelcomeView` |
| 2 | Set up Echo → username + Face ID | Key generated; no crash |
| 3 | Recovery phrase flow | 24 words; confirm 3 words |
| 4 | Reach main tab bar | — |
| 5 | Backend | `GET /identity/@<username>` → 200 |

Simulator Face ID: Features → Face ID → **Matching Face** when prompted.

### 5.2 Login (returning user)

| # | Action | Pass |
|---|--------|------|
| 1 | Background → foreground | `StorageLockedView` |
| 2 | Unlock with Face ID | Main app |
| 3 | Force-quit → relaunch | `GlacialLoginScreen`; auto Face ID ~400ms |
| 4 | 5× Face ID fail | `BiometricLockoutView` |

### 5.3 Messaging (backend required)

| # | Action | Pass |
|---|--------|------|
| 1 | New conversation → send | Message shows sent status |
| 2 | Receive from second client/curl | Appears in thread |
| 3 | Header | Verified / key fingerprint visible |
| 4 | Backend logs | No plaintext message bodies |

### 5.4 Phase 3 — typing, read receipts, reactions (WO-192)

**Needs two signed-in accounts** (two simulators if supported, or simulator + device).

| Test | Pass criteria |
|------|---------------|
| Typing | A types → B sees indicator; stop/send clears; ~6s safety clear |
| Read receipts | B opens chat → A status → read; privacy off → stays delivered |
| Reactions | Long-press 👍; toggle off; REST matches UI |
| Privacy leak | Account C not in thread receives **no** ephemeral WS signals |

Detail: [`PHASE3_IOS_UI_SPEC.md` Step 5](PHASE3_IOS_UI_SPEC.md), skill **`echo-phase3-ios-wire`**.

### 5.5 OIDC4VC wallet enrollment (WO-100)

**Backend:** `OIDC4VC_ENABLED=true`

| # | Action | Pass |
|---|--------|------|
| 1 | Enrollment → Mobile wallet credential | `WalletCredentialEnrollmentView` |
| 2 | Start flow | `POST /v1/enrollment/vc/start` succeeds; wallet / browser opens |
| 3 | Complete presentation | `echo-enroll://` callback or manual finish with VP |
| 4 | Finish | `POST /v1/enrollment/vc/finish` → onboarding continues |
| 5 | Dev fallback | Browser `/verification/ui` handoff page loads |

**Verify URL scheme:** `EchoApp-Info.plist` contains `echo-enroll` (merged via `INFOPLIST_FILE`).

### 5.6 PSI contact discovery (WO-221)

**Backend:** contacts PSI endpoints; both test users SMS-verified (discovery index populated).

**iOS:** `EchoOPRF.xcframework` embedded (§4c) for real PSI — mock OPRF will not match production index.

| # | Action | Pass |
|---|--------|------|
| 1 | Contacts tab → **Contacts on ECHO** | `ContactDiscoveryView` |
| 2 | Grant Contacts permission | Prompt appears once |
| 3 | Scan | Loading → results or empty state (no crash) |
| 4 | Two users with shared phone book entry | Matched contact shows display name + DID |
| 5 | Network | `POST /v3/contacts/psi` — blinded payloads only (no raw phones in logs) |

Unit tests (no Xcode UI): `swift test --filter ContactDiscoveryLogicTests` (when target wired) + `go test ./mobile/echooprf/...`

### 5.7 Hidden persona gate

| # | Action | Pass |
|---|--------|------|
| 1 | Settings → Personas → Hidden | Lock icon |
| 2 | Open persona | `PersonaGateView`; Face ID ~300ms |
| 3 | Background 2+ min | Re-lock on foreground |

### 5.8 SMS recovery (dev)

Enable `DEV_MODE=true`; OTP in `X-Dev-OTP` response header. Flow: Recovery Setup → Add phone → verify OTP.

---

## 6. Manual E2E — Physical iPhone

Simulator cannot exercise Secure Enclave, real Face ID, or realistic PSI with two real phone books.

### 6a. Network setup

```bash
ipconfig getifaddr en0   # e.g. 192.168.1.42
```

In `.env`: `API_URL=http://192.168.1.42:8000` → `make dev-restart`

In Xcode scheme **Echo** → Environment / build setting: `API_URL` = same LAN URL.

Phone and Mac on **same Wi‑Fi**; macOS firewall allows inbound :8000.

### 6b. Device checklist (minimum launch set)

Run **§5.1–5.3** on device, plus:

- [ ] Real Face ID / Touch ID unlock paths
- [ ] Push notification delivery (APNs sandbox + `APNS_*` in backend `.env`)
- [ ] Background/foreground storage lock
- [ ] WO-100 wallet enrollment with real wallet or dev VP UI
- [ ] WO-221 contact discovery with **live OPRF** + second device
- [ ] Phase 3 two-client test (§5.4) across simulator + device

### 6c. Two-device PSI regression

1. Device A: onboard + SMS verify phone `+1…`
2. Device B: onboard + SMS verify; save A's number in iOS Contacts
3. Both: run contact discovery scan
4. B should see A in **Contacts on ECHO** (and vice versa if symmetric)

---

## 7. TestFlight regression

After each upload to App Store Connect (internal testing):

| # | Check | Pass |
|---|-------|------|
| 1 | Install from TestFlight app | Opens without crash |
| 2 | API points to **deployed** HTTPS backend (not localhost) | Health OK in network log |
| 3 | Full onboarding on clean install | §5.1 |
| 4 | Login + biometric | §5.2 |
| 5 | Send/receive message | §5.3 |
| 6 | Phase 3 signals (if build includes) | §5.4 |
| 7 | No crash in Organizer first 24h | Crash rate acceptable |
| 8 | Export compliance answered in ASC | Standard encryption exemption |

Scheme: use **Release** + production/staging `API_URL`. APNs: `APNS_ENVIRONMENT=sandbox` for TestFlight.

Upload walkthrough: [`PHASE1_LAUNCH.md` §4](PHASE1_LAUNCH.md).

---

## 8. Launch & regression sign-off

Copy for release notes / WO closure:

```markdown
## Echo release test sign-off — vX.Y.Z — DATE

### Automated (required)
- [ ] CI green on commit SHA: _______
- [ ] `make regression` passed locally
- [ ] `make validate-phase1` all steps ok (launch only)

### Manual simulator (required pre-TestFlight)
- [ ] §5.1 Onboarding
- [ ] §5.2 Login / lockout
- [ ] §5.3 Messaging
- [ ] §5.4 Phase 3 signals (if in scope)
- [ ] §5.5 OIDC4VC (if in scope)
- [ ] §5.6 PSI discovery (if in scope)

### Manual device (required for TestFlight external / App Store)
- [ ] §6b device checklist
- [ ] §7 TestFlight regression on build _______

### Sign-off
Tester: _______  Environment: dev / staging / prod  Result: GO / NO-GO
```

---

## 9. Command reference

```bash
# ── Automated ─────────────────────────────────────────────────────────────
make regression                    # Headless regression (Go + iOS SPM)
make regression-quick              # Go only
make regression-with-phase1        # + validate-phase1
make release-check                 # Pre-tag Go gate
./scripts/run-regression.sh --help

# ── Backend ───────────────────────────────────────────────────────────────
make dev && make dev-status
make validate-phase1
go test -race -count=1 ./internal/... ./pkg/... ./test/...

# ── iOS (Mac + Xcode) ─────────────────────────────────────────────────────
cd ios/Echo && swift test --filter EchoSecurityTests
cd ios/Echo && swift test --filter EchoPhase3Tests
./scripts/build-echooprf-ios.sh
xcodebuild -project ios/Echo/EchoApp.xcodeproj -scheme Echo \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' build

# ── Metagraph ─────────────────────────────────────────────────────────────
make metagraph-test
cd metagraph && sbt test
```

---

## 10. Related documentation

| Doc | Contents |
|-----|----------|
| [`PHASE1_LAUNCH.md`](PHASE1_LAUNCH.md) | TestFlight, signing, June 1 checklist, env vars |
| [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) | Phase 3 agent vs Xcode split, Step 5 detail |
| [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) | Hydra cluster, node URLs |
| [`CONTRIBUTING.md`](../CONTRIBUTING.md) | First-time dev setup |
| [`AGENTS.md`](../AGENTS.md) | Agent skills index |

**Skills:** `echo-testing`, `echo-phase1-validate`, `echo-ios-agent-vs-xcode`, `echo-phase3-ios-wire`
