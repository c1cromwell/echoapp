# Echo — E2E Launch & Testing Guide

Single reference for **automated regression** (CI, scripts, agents) and **manual E2E** on Mac/Xcode (simulator + physical iPhone + TestFlight). Use before every TestFlight build and for periodic regression.

**Not** the developer onboarding guide — see [`CONTRIBUTING.md`](../CONTRIBUTING.md) for first-time setup.

---

## 0. How to use this doc

| Goal | Start here |
|------|------------|
| Pre-merge / weekly regression | [§4 Automated regression](#4-automated-regression) → `make regression` |
| TestFlight launch (full path) | [§4](#4-automated-regression) + [§6–8](#6-manual-e2e--ios-simulator) + [§9 Launch checklist](#9-launch-checklist--sign-off) |
| Metagraph + backend E2E | [§3](#3-environment-setup) + [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) |
| Phase 3 messaging signals | [§6.4](#64-phase-3--typing-read-receipts-reactions-wo-192) + [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) |
| OIDC4VC enrollment (WO-100) | [§6.5](#65-oidc4vc-wallet-enrollment-wo-100) |
| PSI contact discovery (WO-221) | [§6.6](#66-psi-contact-discovery-wo-221) |
| Agent automation | Skill **`.cursor/skills/echo-testing`** + MCP `echo-local-dev` |

---

## 1. Testing tiers & gates

| Tier | Runs where | When | Automatable |
|------|------------|------|-------------|
| **A — CI** | GitHub Actions | Every push/PR to `main` / `develop` | ✅ Fully |
| **B — Headless regression** | Mac or Linux (Go); Mac + Xcode (iOS SPM) | Pre-release, weekly | ✅ `make regression` |
| **C — Backend E2E** | Mac + Docker + Euclid | Before TestFlight; after metagraph changes | ✅ `make validate-phase1` |
| **D — iOS build verify** | Mac + Xcode.app | After iOS project changes | ✅ `xcodebuild` |
| **E — Simulator manual E2E** | Mac + Xcode Simulator | Before TestFlight; feature QA | ❌ Checklists §6 |
| **F — Device manual E2E** | Physical iPhone(s) | Before TestFlight; biometrics, PSI, LAN | ❌ Checklists §7 |
| **G — TestFlight regression** | Physical device + prod/staging API | After each uploaded build | ❌ Checklist §8e |

**Rule:** Tiers A–D must be green before running E–G. Tier C may `skip` metagraph steps for simulator-only work; **all steps `ok`** required for launch go/no-go.

---

## 2. Prerequisites & quick-start

```bash
# From repo root — should pass before manual E2E or TestFlight work
make release-check          # Go build + tests + vet + fmt + VERSION
cd ios/Echo && DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer \
  swift build --target Echo --target EchoSecurityTests
```

If both complete without errors, continue to [§3](#3-environment-setup) and [§4](#4-automated-regression). Otherwise complete [`CONTRIBUTING.md`](../CONTRIBUTING.md) first.

---

## 3. Environment setup

### 3a. Local backend stack

```bash
cp .env.example .env    # first time
make dev
make dev-status
curl -s http://localhost:8000/health | jq .status   # → "operational"
```

Optional Identity nodes (VC / trust tier):

```bash
make start-identity   # L0 :9600, L1 :9500
```

### 3b. WO-230 go/no-go validation

```bash
make validate-phase1
```

Expected for a clean Phase 1 build:

```
✓ Step 0 — prerequisites
✓ Step 1 — derive did:key locally
✓ Step 2 — register DID with backend
✓ Step 3 — Identity Metagraph L0/L1 reachable
✓ Step 4 — test message through relay
✓ Step 5 — Merkle root submitted to Data L1
✓ Step 6 — Global L0 snapshot height increments
Go/No-Go: GO ✓
```

Steps 3 and 5 may show `skip` when the Euclid cluster isn't running — fine for simulator-only. **All steps must be `ok` before TestFlight sign-off.**

Details: [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md), skill **`echo-phase1-validate`**.

### 3c. Physical device — LAN `API_URL`

Physical iPhone cannot reach `localhost`. Point at your Mac's LAN IP:

```bash
ipconfig getifaddr en0   # Wi-Fi — e.g. 192.168.1.42

# In .env:
API_URL=http://192.168.1.42:8000
make dev-restart
```

In Xcode scheme **Echo** → Environment / build setting: `API_URL` = same LAN URL. Phone and Mac on **same Wi‑Fi**; macOS firewall allows inbound :8000.

For TestFlight, use your **deployed HTTPS** backend — see [§8b](#8b-backend-deployment).

### 3d. Feature flags

**OIDC4VC (WO-100)** — in `.env`:

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

**SMS recovery (dev):** `DEV_MODE=true` — OTP echoed in `X-Dev-OTP` response header (never in prod).

### 3e. Environment variable reference

All variables live in `.env` (copy from `.env.example`).

| Variable | Example | Required for |
|----------|---------|--------------|
| `API_PORT` | `8000` | Backend listen port |
| `API_URL` | `http://192.168.1.42:8000` | Device / deployed client URL |
| `DATABASE_HOST` | `localhost` | PostgreSQL |
| `REDIS_HOST` | `localhost` | Rate limiting, OTP sessions, device tokens |
| `IDENTITY_L1_URL` | `http://localhost:9500` | WO-274 VC anchor |
| `DATA_L1_URL` | `http://localhost:9400` | WO-230 Merkle root |
| `IDENTITY_SERVICE_DID` | `did:key:z…` | Scala validator authorized-sender check |
| `LOG_MASTER_KEY` | `<64 hex chars>` | WO-53 audit log encryption |
| `OIDC4VC_ENABLED` | `true` | WO-100 wallet enrollment |
| `APNS_KEY_PATH` | `./keys/AuthKey.p8` | iOS push notifications |
| `APNS_KEY_ID` | `ABC123DEFG` | APNs key identifier |
| `APNS_TEAM_ID` | `XYZTEAMID` | Apple Developer team ID |
| `APNS_BUNDLE_ID` | `com.echo.app` | Must match Xcode bundle ID |
| `APNS_ENVIRONMENT` | `sandbox` (TestFlight) / `production` (App Store) | |
| `DEV_MODE` | `true` (dev only!) | OTP in `X-Dev-OTP` header |
| `TWILIO_ACCOUNT_SID` | `ACxxxxxxxx` | SMS OTP (Wave 12) |
| `TWILIO_AUTH_TOKEN` | `…` | SMS OTP |
| `TWILIO_FROM` | `+15005550006` | SMS sender number |

Generate `IDENTITY_SERVICE_DID`:

```bash
go run ./cmd/didkey -out ./.keys/identity-service.pem
# Paste into .env: IDENTITY_SERVICE_DID=did:key:z6Mkh…
```

Generate `LOG_MASTER_KEY`:

```bash
openssl rand -hex 32
# Paste 64-char hex into .env: LOG_MASTER_KEY=<value>
```

---

## 4. Automated regression

### 4a. One command (recommended)

```bash
make regression                    # Go release-check + targeted suites + iOS SPM
make regression-quick              # Go race tests only
make regression-with-phase1        # + validate-phase1 (needs Docker + Euclid)
```

Underlying script: `scripts/run-regression.sh` (`--quick`, `--with-phase1`, `--ios-only`).

### 4b. What `make regression` includes

| Step | Command / target | Hard gate? |
|------|------------------|------------|
| Go build + race tests + vet + fmt | `make release-check` | ✅ |
| Enrollment VC, WS, reactions, contacts | `go test ./internal/api/ … ./internal/services/contacts/… ./mobile/echooprf/…` | ✅ |
| iOS library + security tests | `swift build --target Echo`, `EchoSecurityTests` | ✅ (needs Xcode.app) |
| Phase 3 logic tests | `swift test --filter EchoPhase3Tests` | ✅ |
| OPRF mobile interop | `go test ./mobile/echooprf/...` | ✅ |

### 4c. CI (Tier A)

| Workflow | Gate |
|----------|------|
| Go CI | `go test -race ./internal/... ./pkg/...` |
| iOS SPM (`.github/workflows/ios-ci.yml`) | `EchoSecurityTests`, `EchoPhase3Tests` |
| T0–T7 data classification | `semgrep --config .semgrep/t0_t7_rules.yaml --error .` |

### 4d. API regression (full Go suite)

```bash
go test -race -count=1 ./internal/... ./pkg/... ./test/...
semgrep --config .semgrep/t0_t7_rules.yaml --error .
```

### 4e. MCP automation (Cursor agents)

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

## 5. iOS automated (Mac + Xcode)

Requires **Xcode.app** (not Command Line Tools only).

### 5a. SPM tests (same as CI)

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

### 5b. EchoApp target build (compile gate)

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj -scheme Echo \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  -configuration Debug build
```

### 5c. Live OPRF framework (WO-221)

```bash
./scripts/build-echooprf-ios.sh
```

Then in Xcode: EchoApp → **Frameworks** → add `ios/Echo/Libraries/EchoOPRF.xcframework` → **Embed & Sign**.

Without the framework, `OPRFClientFactory` falls back to `MockOPRFClient` (dev only — not valid for PSI launch QA).

### 5d. Archive (pre-TestFlight compile)

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj \
  -scheme Echo \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination 'generic/platform=iOS' \
  archive
```

Full upload flow: [§8c](#8c-archive--upload-to-app-store-connect).

---

## 6. Manual E2E — iOS Simulator

**Prerequisites:** [§3a](#3a-local-backend-stack) running; Simulator with Face ID enrolled (Features → Face ID → Enrolled).

Set scheme env / build setting `API_URL` = `http://localhost:8000`.

Open `ios/Echo/EchoApp.xcodeproj` → scheme **Echo** → Run on **iPhone 16 Pro** simulator.

### 6.1 Onboarding (new user)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Cold launch (no account) | `EchoWelcomeView` — single screen, three privacy facts |
| 2 | Tap "Set up Echo" | `NameAndKeyView` — username field + privacy receipt |
| 3 | Type a username (e.g. `chad`) | Available checkmark appears |
| 4 | Tap "Continue with Face ID" | Face ID prompt; Secure Enclave key generated |
| 5 | Face ID succeeds | `RecoverySetupView` shown |
| 6 | Tap "Show phrase" | 24-word BIP-39 phrase (blurs on screen record) |
| 7 | Confirm 3 challenge words | Recovery phrase confirmed |
| 8 | Tap "Continue" | Main app tab bar appears |
| **Verify** | Backend | `GET http://localhost:8000/identity/@username` → 200 with DID |
| **Verify** | Keychain | `echo.did.current` and `echo.username.current` set |

Simulator Face ID: Features → Face ID → **Matching Face** when prompted.

**Biometrics unavailable:** Simulator with no Face ID enrolled → `LAError.biometryNotAvailable` message in form, not a crash.

### 6.2 Login (returning user)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Background app, foreground | `StorageLockedView` (storage key zeroed) |
| 2 | Tap "Unlock" | Face ID prompt |
| 3 | Face ID succeeds | Main app shown |
| 4 | Force-quit and relaunch | `GlacialLoginScreen` — "@username" mono label; Face ID auto-triggers ~400ms |
| 5 | Face ID succeeds | Main app |
| 6 | Trigger 5 Face ID failures | `BiometricLockoutView` — requires passcode |

### 6.3 Messaging (backend required)

| Step | Action | Expected |
|------|--------|----------|
| 1 | New conversation → send | Message shows sent status |
| 2 | Receive from second client/curl | Appears in thread |
| 3 | Check conversation header | Verified / key fingerprint visible |
| **Verify** | Backend logs | No plaintext message bodies |

Backend smoke (no iOS):

```bash
make validate-phase1   # Step 4 exercises relay
```

### 6.4 Phase 3 — typing, read receipts, reactions (WO-192)

**Needs two signed-in accounts** (two simulators if supported, or simulator + device).

| Test | Pass criteria |
|------|---------------|
| Typing | A types → B sees indicator; stop/send clears; ~6s safety clear |
| Read receipts | B opens chat → A status → read; privacy off → stays delivered |
| Reactions | Long-press 👍; toggle off; REST matches UI |
| Privacy leak | Account C not in thread receives **no** ephemeral WS signals |

Detail: [`PHASE3_IOS_UI_SPEC.md` Step 5](PHASE3_IOS_UI_SPEC.md), skill **`echo-phase3-ios-wire`**.

### 6.5 OIDC4VC wallet enrollment (WO-100)

**Backend:** `OIDC4VC_ENABLED=true`

| Step | Action | Expected |
|------|--------|----------|
| 1 | Enrollment → Mobile wallet credential | `WalletCredentialEnrollmentView` |
| 2 | Start flow | `POST /v1/enrollment/vc/start` succeeds; wallet / browser opens |
| 3 | Complete presentation | `echo-enroll://` callback or manual finish with VP |
| 4 | Finish | `POST /v1/enrollment/vc/finish` → onboarding continues |
| 5 | Dev fallback | Browser `/verification/ui` handoff page loads |

**Verify URL scheme:** `EchoApp-Info.plist` contains `echo-enroll` (merged via `INFOPLIST_FILE`).

### 6.6 PSI contact discovery (WO-221)

**Backend:** contacts PSI endpoints; both test users SMS-verified (discovery index populated).

**iOS:** `EchoOPRF.xcframework` embedded ([§5c](#5c-live-oprf-framework-wo-221)) for real PSI.

| Step | Action | Expected |
|------|--------|----------|
| 1 | Contacts tab → **Contacts on ECHO** | `ContactDiscoveryView` |
| 2 | Grant Contacts permission | Prompt appears once |
| 3 | Scan | Loading → results or empty state (no crash) |
| 4 | Two users with shared phone book entry | Matched contact shows display name + DID |
| **Verify** | Network | `POST /v3/contacts/psi` — blinded payloads only (no raw phones in logs) |

Unit tests: `go test ./mobile/echooprf/...`

### 6.7 Hidden persona gate

| Step | Action | Expected |
|------|--------|----------|
| 1 | Settings → Personas → Hidden | Lock icon |
| 2 | Open persona | `PersonaGateView`; Face ID ~300ms |
| 3 | Background 2+ min | Re-lock on foreground |

### 6.8 SMS recovery (dev)

Enable `DEV_MODE=true`. Flow: Recovery Setup → Add phone → verify OTP from `X-Dev-OTP` header.

| Step | Action | Expected |
|------|--------|----------|
| 1 | Recovery Setup → "Add phone backup" | `SMSOTPSetupView` |
| 2 | Enter valid E.164 (+12125551234) | "Send verification code" enabled |
| 3 | Tap send | `POST /v1/auth/sms-recovery/register` → 200 |
| 4 | Enter OTP | "Phone backup added ✓" |

### 6.9 Rate limiting (backend resilience)

```bash
for i in $(seq 1 105); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8000/v3/messages
done | sort | uniq -c
# Expected: ~100 × 200, ~5 × 429
```

---

## 7. Manual E2E — Physical iPhone

Simulator cannot exercise Secure Enclave, real Face ID, or realistic PSI with two real phone books.

Use [§3c](#3c-physical-device--lan-api_url) for network setup.

### 7a. Device checklist (minimum launch set)

Run **§6.1–6.3** on device, plus:

- [ ] Real Face ID / Touch ID unlock paths
- [ ] Push notification delivery (APNs sandbox + `APNS_*` in backend `.env`)
- [ ] Background/foreground storage lock
- [ ] WO-100 wallet enrollment with real wallet or dev VP UI
- [ ] WO-221 contact discovery with **live OPRF** + second device
- [ ] Phase 3 two-client test (§6.4) across simulator + device

### 7b. Two-device PSI regression

1. Device A: onboard + SMS verify phone `+1…`
2. Device B: onboard + SMS verify; save A's number in iOS Contacts
3. Both: run contact discovery scan
4. B should see A in **Contacts on ECHO** (and vice versa if symmetric)

---

## 8. TestFlight & App Store

### 8a. iOS signing setup (one-time)

Requires active **Apple Developer Program** membership.

**App Store Connect:**

1. **Identifiers → +** → App ID → Bundle ID: `com.echo.app`
   - Capabilities: **Push Notifications**, **Associated Domains** (passkeys)
2. **Certificates → +** → Apple Distribution → save `.p12`
3. **Provisioning Profiles → +** → App Store → download

**Xcode** (`ios/Echo/EchoApp.xcodeproj`):

1. Target **Echo** → **Signing & Capabilities**
2. **Team** = your Apple Developer team
3. **Bundle Identifier**: `com.echo.app`
4. **Automatically manage signing** (recommended)

**APNs for TestFlight:** App Store Connect → **Keys → +** → APNs → download `.p8`. Set `APNS_KEY_PATH`, `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID=com.echo.app`, `APNS_ENVIRONMENT=sandbox` in backend `.env`.

**TestFlight scheme:**

1. Duplicate `Debug` scheme → name `Echo (TestFlight)`
2. **Run** → Build Configuration → `Release`
3. Set `API_URL` to deployed backend ([§8b](#8b-backend-deployment))

### 8b. Backend deployment

TestFlight app must connect to a real HTTPS backend (not localhost).

| Service | Minimum spec | Notes |
|---------|--------------|-------|
| Go API server | 1 vCPU / 512 MB RAM | fly.io, Render, Railway, VPS |
| PostgreSQL | Managed | Run migrations: `make migrate` |
| Redis | Upstash free tier | Rate limiting + OTP |
| TLS | Let's Encrypt | Required — iOS enforces HTTPS |

Quick Fly.io path:

```bash
brew install flyctl && fly auth login
fly launch --name echo-api --region sjc
fly secrets set DATABASE_HOST=... REDIS_HOST=... IDENTITY_L1_URL=...
fly deploy
curl -s https://echo-api.fly.dev/health | jq .status   # → "operational"
```

Update iOS `Release.xcconfig` or scheme: `API_URL = https://echo-api.fly.dev`

### 8c. Archive & upload to App Store Connect

```bash
cd ios/Echo
xcodebuild \
  -project EchoApp.xcodeproj \
  -scheme "Echo (TestFlight)" \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination "generic/platform=iOS" \
  archive
```

Or Xcode: **Product → Archive** → Organizer → **Distribute App** → **App Store Connect** → **Upload**.

Options: strip Swift symbols, automatically manage signing.

### 8d. Internal & external testing

**Internal** (up to 100 testers, no Apple review):

1. App Store Connect → TestFlight → build appears (~5 min)
2. Add Internal Testers → **Enable** build
3. Testers install via TestFlight app

**External** (requires Beta App Review, 1–3 business days):

- Privacy Policy URL, screenshots, no placeholder content on first launch

### 8e. TestFlight regression (after each upload)

| # | Check | Pass |
|---|-------|------|
| 1 | Install from TestFlight app | Opens without crash |
| 2 | API points to **deployed** HTTPS backend | Health OK in network log |
| 3 | Full onboarding on clean install | §6.1 |
| 4 | Login + biometric | §6.2 |
| 5 | Send/receive message | §6.3 |
| 6 | Phase 3 signals (if build includes) | §6.4 |
| 7 | No crash in Organizer first 24h | Crash rate acceptable |
| 8 | Export compliance answered in ASC | Standard encryption exemption |

Use **Release** + production/staging `API_URL`. APNs: `APNS_ENVIRONMENT=sandbox` for TestFlight.

---

## 9. Launch checklist & sign-off

Use this template for any TestFlight or App Store launch. Adjust dates in your release plan.

### Phase A — Backend + local E2E green

- [ ] `make validate-phase1` passes all 6 steps (or documented skips for simulator-only)
- [ ] `make regression` passed on release commit SHA
- [ ] Physical iPhone on LAN: full onboarding (§6.1) works
- [ ] Login biometric auto-trigger works on device (§6.2)
- [ ] Hidden persona gate works on device (§6.7)
- [ ] T0–T7 Semgrep CI passes on release branch

### Phase B — Signing + backend deployed

- [ ] Apple Developer account active
- [ ] Bundle ID `com.echo.app` registered in App Store Connect
- [ ] Distribution certificate + provisioning profile created
- [ ] Backend deployed to public HTTPS URL
- [ ] TestFlight scheme `API_URL` points to deployed backend
- [ ] Physical device connects to deployed backend (Xcode network logs)
- [ ] APNs key configured

### Phase C — TestFlight build live

- [ ] Archive builds successfully
- [ ] Upload to App Store Connect — build visible in TestFlight
- [ ] Internal testers added, build enabled
- [ ] Full onboarding on TestFlight build (not simulator)
- [ ] Login + biometric on TestFlight build
- [ ] §8e TestFlight regression complete
- [ ] No crashes in first 24h (Xcode Organizer)

**Definition of done:**

> At least one TestFlight internal build installable on a real device, with onboarding → login → send message → hidden persona gate working end-to-end against the deployed backend.

### Release sign-off (copy for WO closure)

```markdown
## Echo release test sign-off — vX.Y.Z — DATE

### Automated (required)
- [ ] CI green on commit SHA: _______
- [ ] `make regression` passed locally
- [ ] `make validate-phase1` all steps ok (launch only)

### Manual simulator (required pre-TestFlight)
- [ ] §6.1 Onboarding
- [ ] §6.2 Login / lockout
- [ ] §6.3 Messaging
- [ ] §6.4 Phase 3 signals (if in scope)
- [ ] §6.5 OIDC4VC (if in scope)
- [ ] §6.6 PSI discovery (if in scope)

### Manual device (required for TestFlight external / App Store)
- [ ] §7a device checklist
- [ ] §8e TestFlight regression on build _______

### Sign-off
Tester: _______  Environment: dev / staging / prod  Result: GO / NO-GO
```

---

## 10. Known issues & workarounds

| Issue | Workaround |
|-------|------------|
| Face ID doesn't trigger in Simulator | Features → Face ID → Enrolled → Matching Face |
| `SecureEnclave` fails in Simulator | Expected — hardware-only; `USE_MOCK_SECURE_ENCLAVE=1` in scheme |
| `POST /identity/register` 400 on device | `API_URL` must be LAN IP, not `localhost` |
| APNs push not delivered | `APNS_ENVIRONMENT=sandbox` for TestFlight |
| TestFlight "Missing Compliance" | ASC → Build → Export Compliance → standard encryption |
| Biometric lockout wrong countdown | Check `BiometricLockState.hardLocked(until:)` in `SecureEnclaveManager` |
| `validate-phase1` step 5 skips | Needs Euclid cluster + `DATA_L1_URL=http://localhost:9400` |
| `gomobile bind` fails (CLT only) | Full Xcode.app; see `./scripts/build-echooprf-ios.sh` |

---

## 11. Command reference

```bash
# ── Automated regression ──────────────────────────────────────────────────
make regression
make regression-quick
make regression-with-phase1
make release-check
./scripts/run-regression.sh --help

# ── Backend ───────────────────────────────────────────────────────────────
make dev && make dev-status
make dev-restart && make dev-logs && make dev-stop
make validate-phase1
make start-identity
go test -race -count=1 ./internal/... ./pkg/... ./test/...
semgrep --config .semgrep/t0_t7_rules.yaml --error .

# ── iOS (Mac + Xcode) ─────────────────────────────────────────────────────
cd ios/Echo && swift test --filter EchoSecurityTests
cd ios/Echo && swift test --filter EchoPhase3Tests
./scripts/build-echooprf-ios.sh
xcodebuild -project ios/Echo/EchoApp.xcodeproj -scheme Echo \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' build
xcodebuild -project ios/Echo/EchoApp.xcodeproj -scheme "Echo (TestFlight)" \
  -configuration Release -archivePath /tmp/Echo.xcarchive \
  -destination 'generic/platform=iOS' archive

# ── Metagraph ─────────────────────────────────────────────────────────────
make metagraph-test
cd metagraph && sbt test
```

---

## 12. Related documentation

| Doc | Contents |
|-----|----------|
| [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) | Phase 3 agent vs Xcode split, Step 5 detail |
| [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) | Hydra cluster, node URLs |
| [`COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`](COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md) | Wave 0+ E2E gates |
| [`CONTRIBUTING.md`](../CONTRIBUTING.md) | First-time dev setup |
| [`AGENTS.md`](../AGENTS.md) | Agent skills index |

**Skills:** `echo-testing`, `echo-phase1-validate`, `echo-ios-agent-vs-xcode`, `echo-phase3-ios-wire`
