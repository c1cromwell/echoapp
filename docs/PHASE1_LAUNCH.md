# Phase 1 — End-to-End Testing & TestFlight Launch Guide
## Target: TestFlight internal build live by June 1, 2026

This document is the **launch checklist**, not the developer onboarding guide
(that's `CONTRIBUTING.md`). Read this when you want to:

- Confirm the Phase 1 deliverable works end-to-end on a real device.
- Set up code signing and submit to TestFlight.
- Know exactly what still needs to happen before June 1.

---

## Quick-start: are you already set up?

```bash
# From the repo root — should all pass before you touch anything else.
make release-check          # Go build + tests + vet + fmt + VERSION
cd ios/Echo && DEVELOPER_DIR=/Applications/Xcode.app \
  swift build --target Echo --target EchoSecurityTests
```

If both complete without errors, jump to [Section 3](#3-end-to-end-test-flows).
If not, complete `CONTRIBUTING.md` first, then come back here.

---

## 1. Local backend setup (5 minutes)

### 1a. Start the full Phase 1 stack

```bash
# Copy env file if you haven't already
cp .env.example .env

# Bring up: Postgres · Redis · NATS · MinIO · echoapp :8000
# (Euclid metagraph is optional for most iOS testing — see 1b)
make dev
```
# 3. (Optional) Start Identity nodes for VC / trust-tier features
#    Identity L0 + L1 are custom Echo modules not managed by Euclid hydra.
#    They run via docker-compose.identity.yml using the sbt assembly JARs.
make start-identity

Verify:

```bash
make dev-status
curl -s http://localhost:8000/health | jq .status   # → "operational"
```

### 1b. Run the Phase 1 go/no-go validation script (optional — needs Docker + JDK 21)

This exercises all 6 WO-230 steps end-to-end including the Constellation
Euclid metagraph cluster. Skip on first setup; run once you have everything up.

```bash
# Prerequisite: Euclid cluster must be running
# (See CONTRIBUTING.md §3 — takes ~5 min first time)
make validate-phase1
```

Expected output for a clean Phase 1 build:

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

Steps 3 and 5 may show `skip` when the Euclid cluster isn't running — that's
fine for simulator testing. All steps must be `ok` before the TestFlight build.

### 1c. Configure your `.env` for device testing

When testing on a **physical iPhone** the device can't reach `localhost`.
Point it at your Mac's LAN IP instead:

```bash
# Find your Mac's IP
ipconfig getifaddr en0    # Wi-Fi — e.g. 192.168.1.42

# In .env set:
API_URL=http://192.168.1.42:8000

# Restart backend
make dev-restart
```

For TestFlight (which connects to your *production* backend), replace this with
your deployed API URL once you have a server. See [Section 5](#5-backend-deployment).

---

## 2. iOS signing setup (one-time, ~30 minutes)

You need an active **Apple Developer Program** membership ($99/yr) to submit to
TestFlight. If you don't have one: enroll at developer.apple.com and wait for
approval (usually instant for individuals, 1–2 days for organisations).

### 2a. Certificates and identifiers in App Store Connect

1. Sign in at [appstoreconnect.apple.com](https://appstoreconnect.apple.com).
2. **Certificates, IDs & Profiles → Identifiers → +**
   - Type: **App ID**
   - Bundle ID: `com.echo.app` (explicit, matches `EchoApp.xcodeproj`)
   - Capabilities to enable:
     - **Push Notifications** (for APNs OTP and message delivery)
     - **Associated Domains** (for passkey credential provider)
     - **Sign In with Apple** (optional, Phase 2)
3. **Certificates → +** → Apple Distribution → follow the CSR steps.
   Save the `.p12` file and its password somewhere safe (Keychain, 1Password).
4. **Provisioning Profiles → +** → App Store → select the App ID → select
   your distribution certificate → name it `Echo App Store` → download.

### 2b. Xcode project signing

Open `ios/Echo/EchoApp.xcodeproj` in Xcode 16+:

1. Select the `Echo` target → **Signing & Capabilities**
2. Set **Team** to your Apple Developer team
3. **Bundle Identifier**: `com.echo.app`
4. **Signing Certificate**: `Apple Distribution` (for TestFlight/release)
   or `Apple Development` for local device testing
5. Check **Automatically manage signing** — Xcode handles the provisioning
   profile for you once the bundle ID is registered.

For the **EchoDemo** test app (if you want to ship that too):

1. `ios/EchoDemo/EchoDemo.xcodeproj` → target `EchoDemo`
2. Bundle ID: `com.echo.EchoDemo`
3. Same team, same automatic signing

> **Note on Push Notifications + APNs:** `CONTRIBUTING.md` has the APNs env
> vars. For TestFlight, create an APNs key in App Store Connect:
> **Keys → +** → APNs → download the `.p8` file. Set `APNS_KEY_PATH`,
> `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID=com.echo.app`,
> `APNS_ENVIRONMENT=sandbox` in your backend `.env`.

### 2c. Scheme configuration for TestFlight

In Xcode, duplicate the `Debug` scheme:

1. **Product → Scheme → Manage Schemes**
2. Duplicate `Echo` → name it `Echo (TestFlight)`
3. **Run**: Build Configuration → `Release`
4. Set the `API_URL` build variable to your deployed backend URL
   (Environment Variables in the scheme's Run tab, or better: an
   `xcconfig` file in `ios/Echo/Config/`)

---

## 3. End-to-end test flows

Run each of these on **both the iOS Simulator and a physical iPhone** before
submitting to TestFlight.

### 3.1 · Onboarding (new user)

| Step | Action | Expected |
|---|---|---|
| 1 | Launch app cold (no account) | `EchoWelcomeView` — single screen, three privacy facts |
| 2 | Tap "Set up Echo" | `NameAndKeyView` — username field + privacy receipt |
| 3 | Type a username (e.g. `chad`) | Available checkmark appears |
| 4 | Tap "Continue with Face ID" | Face ID prompt fires; Secure Enclave key generated |
| 5 | Face ID succeeds | `RecoverySetupView` shown |
| 6 | Tap "Show phrase" | 24-word BIP-39 phrase displayed (blurs on screen record) |
| 7 | Confirm 3 challenge words | Recovery phrase confirmed |
| 8 | Tap "Continue" | Main app tab bar appears |
| **Verify** | Backend | `GET http://localhost:8000/identity/@username` → 200 with DID |
| **Verify** | Keychain | `echo.did.current` and `echo.username.current` set |

**Test biometrics unavailable (simulator):**
- Tap "Continue with Face ID" on a simulator with no Face ID enrolled
- Should show: `LAError.biometryNotAvailable` — error message in the form, not a crash

### 3.2 · Login (returning user)

| Step | Action | Expected |
|---|---|---|
| 1 | Background app, foreground | `StorageLockedView` shown (storage key zeroed) |
| 2 | Tap "Unlock" | Face ID prompt |
| 3 | Face ID succeeds | Main app shown |
| 4 | Force-quit and relaunch | `GlacialLoginScreen` — "@chad" mono label, Face ID auto-triggers in 400ms |
| 5 | Face ID succeeds | Main app |
| 6 | Trigger 5 Face ID failures | `BiometricLockoutView` — requires passcode |

### 3.3 · Messaging (requires running backend)

| Step | Action | Expected |
|---|---|---|
| 1 | Start a new conversation | Contact picker → select or enter DID |
| 2 | Send a message | Message appears with `✓` sent status |
| 3 | Receive a message (send from another client/curl) | Delivered to conversation |
| 4 | Check conversation header | `verified · key xx:xx:xx` in Geist Mono |
| **Verify** | Backend log | No plaintext message content in logs |

Quick backend smoke test (no iOS needed):

```bash
# Register two DIDs and send a message between them
DID=$(go run ./cmd/didkey -out /tmp/key.pem)
curl -s -X POST http://localhost:8000/identity/register \
  -H "Content-Type: application/json" \
  -d "{\"did\":\"$DID\",\"public_key_hex\":\"$(go run ./cmd/didkey -pubhex /tmp/key.pem)\"}"

# The relay flow is exercised by validate-phase1 step 4:
make validate-phase1
```

### 3.4 · Hidden persona gate

| Step | Action | Expected |
|---|---|---|
| 1 | Navigate to Settings → Personas | Persona list |
| 2 | Create a persona, set visibility to Hidden | Persona saved with lock icon |
| 3 | Navigate back to persona | `PersonaGateView` — night surface, "Verify to continue." |
| 4 | Face ID auto-triggers in 300ms | Content revealed |
| 5 | Background app for 2+ minutes | Re-lock on foreground |

### 3.5 · SMS recovery setup (optional flow)

| Step | Action | Expected |
|---|---|---|
| 1 | Recovery Setup → "Add phone backup" | `SMSOTPSetupView` |
| 2 | Enter valid E.164 number (+12125551234) | "Send verification code" enabled |
| 3 | Tap send | Backend: `POST /v1/auth/sms-recovery/register` → 200 |
| 4 | Check `X-Dev-OTP` response header (dev mode) | 6-digit OTP |
| 5 | Enter OTP in the app | "Phone backup added ✓" |
| **Enable dev mode** | In `.env`: `DEV_MODE=true` | OTP echoed in response (never in prod) |

### 3.6 · Rate limiting (backend resilience)

```bash
# Should return 429 after 100 requests/minute per DID
for i in $(seq 1 105); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8000/v3/messages
done | sort | uniq -c
# Expected: ~100 lines "200", ~5 lines "429"
```

### 3.7 · API regression tests

```bash
# Full suite (run after every change)
cd /path/to/echoapp
go test -race -count=1 ./internal/... ./pkg/... ./test/...

# T0–T7 data classification (Semgrep)
semgrep --config .semgrep/t0_t7_rules.yaml --error .
```

---

## 4. TestFlight submission walkthrough

### 4a. Build the archive

```bash
# From repo root — creates a Release build and opens Xcode Organizer
cd ios/Echo
xcodebuild \
  -project EchoApp.xcodeproj \
  -scheme "Echo (TestFlight)" \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination "generic/platform=iOS" \
  archive
```

Or via Xcode GUI:
1. Select `Echo (TestFlight)` scheme
2. **Product → Archive**
3. In the Organizer that opens, click **Distribute App**

### 4b. Upload to App Store Connect

In the Xcode Organizer → **Distribute App**:
1. **App Store Connect** → Next
2. **Upload** (not Export — Upload goes directly to TestFlight)
3. Choose options:
   - ✅ Strip Swift symbols
   - ✅ Include bitcode (if targeting iOS < 16)
   - ✅ Automatically manage signing
4. Click **Upload**

Or use `xcrun altool` / **Transporter** for CI automation.

### 4c. TestFlight internal testing

In App Store Connect → your app → **TestFlight**:

1. The build appears in **Builds** after upload (< 5 min usually).
2. Add yourself + team as **Internal Testers** (up to 100 people, no Apple review).
3. Click the build → **Enable** for testing.
4. Testers get an email invite → install via TestFlight app.

Internal testing is instant — no Apple review required. Use this for your
first round.

### 4d. External testing (optional for June 1 scope)

External TestFlight requires a **Beta App Review** (1–3 business days):
1. Add External Testers group (up to 10,000)
2. Submit for Beta App Review

External review checklist:
- [ ] Privacy Policy URL in App Store Connect
- [ ] App description and screenshot (at least 1 per device size)
- [ ] No placeholder or dummy content visible on first launch
- [ ] TestFlight beta feedback questionnaire configured

---

## 5. Backend deployment (pre-TestFlight)

The TestFlight app must connect to a real backend (not localhost). Minimum
infrastructure for the June 1 build:

| Service | Minimum spec | Notes |
|---|---|---|
| Go API server | 1 vCPU / 512 MB RAM | `fly.io`, `Render`, `Railway`, or any VPS |
| PostgreSQL | Managed (Supabase free tier works) | Run migrations: `make migrate` |
| Redis | Upstash free tier | Rate limiting + OTP sessions |
| TLS | Let's Encrypt via Caddy / nginx | Required — iOS enforces HTTPS |

Quick Fly.io deployment (fastest path):

```bash
# Install flyctl
brew install flyctl && fly auth login

# From repo root
fly launch --name echo-api --region sjc
# Fly generates fly.toml — edit to set PORT=8000

# Set secrets from .env
fly secrets set \
  DATABASE_HOST=... DATABASE_PORT=5432 DATABASE_NAME=echoapp \
  DATABASE_USER=... DATABASE_PASSWORD=... \
  REDIS_HOST=... REDIS_PORT=6379 \
  IDENTITY_L1_URL=...

# Deploy
fly deploy
```

After deploy:

```bash
# Smoke test
curl -s https://echo-api.fly.dev/health | jq .status   # → "operational"

# Update iOS scheme to use deployed URL
# ios/Echo/Config/Release.xcconfig (create if absent):
# API_URL = https://echo-api.fly.dev
```

---

## 6. June 1 launch checklist

**Today is May 11. You have 21 days.**

### Week 1: May 11–17 — Backend + E2E green

- [ ] `make validate-phase1` passes all 6 steps
- [ ] Physical iPhone on local Wi-Fi: full onboarding flow works
- [ ] Login biometric auto-trigger works on device
- [ ] Hidden persona gate (night surface) works on device
- [ ] All Go tests pass: `make release-check`
- [ ] T0–T7 Semgrep CI passes on main branch

### Week 2: May 18–24 — Signing + backend deployed

- [ ] Apple Developer account active
- [ ] Bundle ID `com.echo.app` registered in App Store Connect
- [ ] Distribution certificate + provisioning profile created
- [ ] Backend deployed to a public HTTPS URL
- [ ] TestFlight scheme `API_URL` points to deployed backend
- [ ] Physical device connects to deployed backend (check network logs in Xcode)
- [ ] APNs key configured (for push notification delivery)

### Week 3: May 25–31 — TestFlight internal build live

- [ ] Archive builds successfully: `xcodebuild archive ...`
- [ ] Upload to App Store Connect — build visible in TestFlight tab
- [ ] Internal testers added, build enabled
- [ ] Full onboarding flow completed on TestFlight build (not simulator)
- [ ] Login + biometric works on TestFlight build
- [ ] No crashes in first 24 hours (check Xcode Organizer Crashes)
- [ ] EchoDemo also on TestFlight (optional — good for UX review partners)

**June 1 definition of done:**
> At least 1 TestFlight internal build installable by external testers (yourself + 1 other real device), with onboarding → login → send message → hidden persona gate working end-to-end against the deployed backend.

---

## 7. Known issues and workarounds

| Issue | Workaround |
|---|---|
| Face ID doesn't trigger in Simulator | Simulator → Features → Face ID → Enrolled, then Matching Face |
| `SecureEnclave` operations fail in Simulator | Expected — Secure Enclave is hardware-only; use `MockSecureEnclaveManager` (set `USE_MOCK_SECURE_ENCLAVE=1` in scheme env vars) |
| `POST /identity/register` returns 400 on device | Check `API_URL` env var in the run scheme — must be reachable from device (not `localhost`) |
| APNs push not delivered | Verify `APNS_ENVIRONMENT=sandbox` for TestFlight, `production` for App Store |
| TestFlight build shows "Missing Compliance" | Navigate to App Store Connect → Build → Export Compliance → select "Yes, uses standard encryption" |
| Biometric lockout shows wrong countdown | `BiometricLockoutView` reads `BiometricLockState.hardLocked(until:)` — set the date correctly in `SecureEnclaveManager.currentLockState()` |
| `validate-phase1 step 5` always skips | Requires live Euclid cluster + `DATA_L1_URL` in `.env` pointing at `localhost:9400` |

---

## 8. Useful commands reference

```bash
# ── Backend ────────────────────────────────────────────────────────────────
make dev                     # Start full Phase 1 stack
make dev-status              # Check all container health
make dev-logs                # Tail backend logs
make dev-restart             # Restart backend only (keep metagraph up)
make dev-stop                # Tear down backend (metagraph stays)
make validate-phase1         # 6-step WO-230 go/no-go
make release-check           # Pre-release sanity (build+test+vet+fmt)
go test -race ./...          # Full Go test suite
semgrep --config .semgrep/t0_t7_rules.yaml --error .   # T0–T7 CI check

# ── iOS ────────────────────────────────────────────────────────────────────
# Build SPM library (macOS — for CI and quick syntax checks)
cd ios/Echo && DEVELOPER_DIR=/Applications/Xcode.app \
  swift build --target Echo --target EchoSecurityTests

# Build and run on simulator
xcodebuild -project ios/Echo/EchoApp.xcodeproj \
  -scheme Echo -destination "platform=iOS Simulator,name=iPhone 16 Pro" \
  -configuration Debug build

# Archive for TestFlight
xcodebuild -project ios/Echo/EchoApp.xcodeproj \
  -scheme "Echo (TestFlight)" \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination "generic/platform=iOS" archive

# ── Metagraph ──────────────────────────────────────────────────────────────
cd metagraph && sbt test                    # All Scala tests
cd metagraph && sbt 'sharedData/test'       # Fast pure-validator tests
make metagraph-test                         # WO-272/277 wired tests
```

---

## 9. Environment variable reference

All variables live in `.env` (copy from `.env.example`). Critical ones for Phase 1:

| Variable | Example | Required for |
|---|---|---|
| `API_PORT` | `8000` | Backend listen port |
| `DATABASE_HOST` | `localhost` | PostgreSQL |
| `REDIS_HOST` | `localhost` | Rate limiting, OTP sessions, device tokens |
| `IDENTITY_L1_URL` | `http://localhost:9500` | WO-274 VC anchor |
| `DATA_L1_URL` | `http://localhost:9400` | WO-230 Merkle root |
| `IDENTITY_SERVICE_DID` | `did:key:z…` | Scala validator authorized-sender check |
| `LOG_MASTER_KEY` | `<64 hex chars>` | WO-53 audit log encryption |
| `APNS_KEY_PATH` | `./keys/AuthKey.p8` | iOS push notifications |
| `APNS_KEY_ID` | `ABC123DEFG` | APNs key identifier |
| `APNS_TEAM_ID` | `XYZTEAMID` | Apple Developer team ID |
| `APNS_BUNDLE_ID` | `com.echo.app` | Must match Xcode bundle ID |
| `APNS_ENVIRONMENT` | `sandbox` (TestFlight) / `production` (App Store) | |
| `DEV_MODE` | `true` (dev only!) | Echo OTP in X-Dev-OTP response header |
| `TWILIO_ACCOUNT_SID` | `ACxxxxxxxx` | SMS OTP (Wave 12) |
| `TWILIO_AUTH_TOKEN` | `…` | SMS OTP |
| `TWILIO_FROM` | `+15005550006` | SMS sender number |

Generate `IDENTITY_SERVICE_DID`:

```bash
# From repo root — generates a fresh P-256 key and prints the did:key
go run ./cmd/didkey -out ./.keys/identity-service.pem
# Output: did:key:z6Mkh…
# Paste into .env: IDENTITY_SERVICE_DID=did:key:z6Mkh…
```

Generate `LOG_MASTER_KEY`:

```bash
openssl rand -hex 32
# Paste 64-char hex string into .env: LOG_MASTER_KEY=<value>
# Keep this in your secret manager — losing it means old audit logs can't be decrypted.
```
