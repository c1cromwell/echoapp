# Echo — E2E launch, testing & sign-off

**Full reference** for automated regression, manual E2E, TestFlight upload, and **release sign-off with delivered feature sets**.

| Doc | Audience |
|-----|----------|
| **[`E2E_QUICK_START.md`](E2E_QUICK_START.md)** | Daily regression, setup, full rebuild (Go + Scala + iOS) |
| **This file** | Launch tiers, step-by-step procedures, TestFlight, sign-off |
| [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md) | External tester scripts |

**Frozen UX:** onboarding/login flows are canonical as shipped — validate only; do not redesign ([`ux-spec.md`](ux-spec.md) §2.1–2.2).

---

## 0. How to use this doc

| Goal | Start here |
|------|------------|
| First-time setup / daily gates | [`E2E_QUICK_START.md`](E2E_QUICK_START.md) §0–5 |
| Pre-merge / weekly regression | [§4 Automated regression](#4-automated-regression) |
| Feature QA on simulator | [§6 Manual E2E — Simulator](#6-manual-e2e--ios-simulator) |
| Device + PSI + biometrics | [§7 Manual E2E — Physical iPhone](#7-manual-e2e--physical-iphone) |
| TestFlight upload | [§8 TestFlight & App Store](#8-testflight--app-store) |
| Release sign-off | [§9 Launch checklist & sign-off](#9-launch-checklist--sign-off) |
| Metagraph cluster detail | [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) |

---

## 1. Feature delivery matrix

What each milestone **ships** and which tests prove it.

### Phase 1 — Foundation (WO-230 go/no-go)

| Feature set | Delivered capability | Automated proof | Launch required |
|-------------|---------------------|-----------------|-----------------|
| **P1-Identity** | `did:key` derive + register; trust-tier `H(tier‖nonce)` on Identity L1 | `validate-phase1` Steps 1–3 | ✅ |
| **P1-Relay** | WebSocket message send/receive | Step 4 + `TestE2E_MessageSendReceive` | ✅ |
| **P1-Data-L1** | Merkle root `POST /data` + finality poll | Step 5 | ✅ |
| **P1-Chain** | Global L0 snapshot advancement | Step 6 | ✅ |
| **P1-Security** | Passkey-signed REST; T0–T7 pre-validation | `make release-check` + Semgrep CI | ✅ |

```bash
make regression-with-phase1   # must print Go/No-Go: GO, all steps ok
```

### Week A — Messaging (Wave 0.1)

| Feature set | Delivered capability | Automated | Manual (required) |
|-------------|---------------------|-----------|---------------------|
| **A-DM** | Deterministic `dm:` threads; send/receive | `make phase3-signals-proof` | A1–A5 |
| **A-Chat-UI** | `ChatView`, composer, tab bar hide | `make ios-preflight BUILD=1` | A2–A5 |
| **A-Phase3** | Typing, read receipts, reactions | `EchoPhase3Tests` + Go WS tests | A6–A9 |
| **A-Persist** | Local thread history | — | A10 |

Tester script: [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md) **A1–A10**.

### Week B — Contacts & discovery

| Feature set | Delivered capability | Manual |
|-------------|---------------------|--------|
| **B-Invite** | `echo://invite` deep link | B1 |
| **B-QR** | Profile QR add contact | B2 |
| **B-Detail** | Block, favorite, Message | B3 |
| **B-SMS** | Phone backup + OTP | B4 |
| **B-Social** | Mutual contacts / groups in common | B5 |
| **B-Link-device** | WO-288 QR link new device | B6 |
| **B-PSI** | OPRF contact discovery (optional) | `make echooprf-ios` + two devices |

### Phase 2+ (when in scope)

| Feature set | Proof |
|-------------|-------|
| OIDC4VC wallet (WO-100) | §6.5 + `OIDC4VC_ENABLED=true` |
| VIP / credentials | `go test ./internal/api/ -run EnrollmentVC` |
| Passport / sync | Phase 2 work orders + manual QA |

Mark **N/A** on sign-off for features not in the release build.

---

## 2. Testing tiers & gates

| Tier | Runs where | When | Automatable |
|------|------------|------|-------------|
| **A — CI** | GitHub Actions | Every push/PR | ✅ |
| **B — Headless regression** | Mac/Linux + Xcode | Pre-release, weekly | ✅ `make regression` |
| **C — Backend E2E** | Mac + Docker + Euclid | After metagraph changes | ✅ `make validate-phase1` |
| **D — iOS build verify** | Mac + Xcode.app | After iOS changes | ✅ `make ios-preflight BUILD=1` |
| **E — Simulator manual** | Xcode Simulator | Feature QA | ❌ §6 |
| **F — Device manual** | Physical iPhone | Pre-TestFlight | ❌ §7 |
| **G — TestFlight** | Device + HTTPS API | Per uploaded build | ❌ §8e |

**Rule:** Tiers **A–D** green before **E–G**. Tier **C** must be full **GO** (no skips) for TestFlight sign-off.

---

## 3. Environment setup

### 3a. Local backend stack

```bash
cp .env.example .env
make dev
make dev-status
curl -s http://localhost:8000/health | jq .status   # → "operational"
```

### 3b. Identity metagraph (Step 3)

```bash
make start-identity
curl -s http://localhost:9600/node/info | jq .state
curl -s http://localhost:9500/node/info | jq .state
```

Separate from hydra — see [ADR-0002](adr/0002-identity-metagraph-deployment.md).

### 3c. WO-230 validation

```bash
make validate-phase1
```

Expected for launch:

```
✓ Steps 0–6 all ok
Go/No-Go: GO
```

Full rebuild if Steps 3/5 fail: [`E2E_QUICK_START.md` §3b](E2E_QUICK_START.md#3b-scala-metagraph-assembly-jars).

### 3d. Physical device — LAN `API_URL`

```bash
ipconfig getifaddr en0
# .env: API_URL=http://192.168.x.x:8000
make dev-restart
```

Xcode scheme **EchoApp** → `API_URL` = same LAN URL.

### 3e. Environment variable reference

| Variable | Example | Required for |
|----------|---------|--------------|
| `API_URL` | `http://192.168.1.42:8000` | Device clients |
| `IDENTITY_L1_URL` | `http://localhost:9500` | VC / trust tier |
| `DATA_L1_URL` | `http://localhost:9400` | Merkle anchor |
| `IDENTITY_SERVICE_DID` | `did:key:z…` | L1 authorized sender |
| `IDENTITY_SERVICE_KEY_PEM` | `/app/.keys/identity-service.pem` | Signed `POST /data` |
| `OIDC4VC_ENABLED` | `true` | WO-100 |
| `DEV_MODE` | `true` (dev only) | `X-Dev-OTP` header |
| `APNS_*` | sandbox key | Push on device |

Generate keys: [`E2E_QUICK_START.md` §2–3](E2E_QUICK_START.md).

---

## 4. Automated regression

```bash
make regression                 # Go + iOS SPM
make regression-quick             # Go race only
make regression-with-phase1       # + validate-phase1
make release-check                # Go gate
./scripts/run-regression.sh --help
```

| Step | Gate |
|------|------|
| `make release-check` | ✅ |
| API + contacts + OPRF Go tests | ✅ |
| `EchoSecurityTests` + `EchoPhase3Tests` | ✅ (Xcode) |

**MCP `echo-local-dev`:** `run_regression`, `run_validate_phase1`, `run_ios_preflight`, `cluster_status`.

---

## 5. iOS automated (Mac + Xcode)

```bash
make ios-preflight BUILD=1 TESTS=1
make phase3-signals-proof
make echooprf-ios    # WO-221 PSI framework
```

| Target | Role |
|--------|------|
| `EchoSecurityTests` | Crypto, Secure Enclave, passkey |
| `EchoPhase3Tests` | Typing, receipts, reactions |

Archive preflight: [§8c](#8c-archive--upload-to-app-store-connect).

---

## 6. Manual E2E — iOS Simulator

**Prerequisites:** `make dev` up; Simulator Face ID enrolled; scheme **EchoApp**; `API_URL=http://localhost:8000`.

### 6.1 Onboarding (frozen flow)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Cold launch | Welcome → set up |
| 2 | Username | Availability OK |
| 3 | Face ID | DID registered |
| 4 | SMS recovery | OTP verified |
| 5 | Continue | Messages tab |

### 6.2 Login (returning user)

| Step | Action | Expected |
|------|--------|----------|
| 1 | Foreground after background | Storage lock → unlock |
| 2 | Relaunch | `GlacialLoginScreen` → Face ID |
| 3 | 5 failures | `BiometricLockoutView` |

### 6.3 Messaging

| Step | Action | Expected |
|------|--------|----------|
| 1 | Send message | Sent status |
| 2 | Second client receives | In thread |
| 3 | Header | Verified / fingerprint |

### 6.4 Phase 3 signals (WO-192)

Two signed-in clients, same `dm:` thread.

| Test | Pass |
|------|------|
| Typing | Indicator; ~6s clear |
| Read receipts | Delivered → read |
| Reactions | 👍 toggle; REST match |
| Privacy | Non-participant gets no WS signal |

Spec: [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md).

### 6.5 OIDC4VC (WO-100)

`OIDC4VC_ENABLED=true` → enrollment → wallet → `echo-enroll://` callback.

### 6.6 PSI discovery (WO-221)

`EchoOPRF.xcframework` embedded; both users SMS-verified; `POST /v3/contacts/psi` (blinded only).

### 6.7 Hidden persona

Settings → Hidden persona → Face ID gate; re-lock after background.

### 6.8 SMS recovery (dev)

`DEV_MODE=true` → `X-Dev-OTP` on register endpoint.

---

## 7. Manual E2E — Physical iPhone

Minimum **device** set before TestFlight external:

- [ ] §6.1–6.3 on device (LAN `API_URL`)
- [ ] Real Face ID / Touch ID
- [ ] APNs push (`APNS_ENVIRONMENT=sandbox`)
- [ ] WO-100 / WO-221 if in release scope
- [ ] Phase 3 two-client (sim + device)
- [ ] Two-device PSI ([§7b](#7b-two-device-psi-regression))

### 7b. Two-device PSI regression

1. Device A + B: onboard, SMS verify, discovery opt-in.
2. Save each other's numbers in iOS Contacts.
3. Both scan — matched contacts appear.

---

## 8. TestFlight & App Store

### 8a. Signing (one-time)

- Bundle ID `com.echo.app` in App Store Connect
- Push + Associated Domains capabilities
- Distribution cert + profile
- APNs `.p8` key in backend `.env`

### 8b. Backend deployment

TestFlight requires **HTTPS** API (not localhost). Deploy Go API + Postgres + Redis; run `make migrate`; set secrets; point iOS `API_URL` to production URL.

### 8c. Archive & upload

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj -scheme "Echo (TestFlight)" \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination 'generic/platform=iOS' \
  archive
```

Xcode: **Product → Archive → Distribute → App Store Connect**.

### 8d. Internal / external testing

- **Internal:** up to 100 testers, no review
- **External:** Beta App Review; privacy policy + screenshots required

### 8e. TestFlight regression (per build)

| # | Check | Pass |
|---|-------|------|
| 1 | Install from TestFlight | No crash on launch |
| 2 | HTTPS backend health | Network log OK |
| 3 | Onboarding clean install | §6.1 |
| 4 | Login + biometric | §6.2 |
| 5 | Send/receive message | §6.3 |
| 6 | Phase 3 (if in build) | §6.4 |
| 7 | 24h crash rate | Acceptable in Organizer |
| 8 | Export compliance | Answered in ASC |

---

## 9. Launch checklist & sign-off

Complete for every TestFlight or App Store release. Copy the template into your release ticket.

### Phase A — Automated + local E2E

| # | Check | Feature set | Pass |
|---|-------|-------------|:----:|
| A1 | `make regression` on release SHA | P1-Security, A-Phase3 headless | ☐ |
| A2 | `make validate-phase1` all steps `ok` | P1-Identity, P1-Data-L1, P1-Chain | ☐ |
| A3 | Semgrep T0–T7 CI green | P1-Security | ☐ |
| A4 | Simulator §6.1–6.3 | A-DM, A-Chat-UI | ☐ |
| A5 | Week A A1–A10 (if messaging release) | A-* | ☐ |
| A6 | Week B B1–B6 (if contacts release) | B-* | ☐ |

### Phase B — Deploy + device

| # | Check | Pass |
|---|-------|:----:|
| B1 | Apple Developer + bundle ID ready | ☐ |
| B2 | Backend HTTPS deployed + migrations | ☐ |
| B3 | TestFlight scheme `API_URL` → deployed API | ☐ |
| B4 | Device §7 minimum on LAN or staging | ☐ |
| B5 | APNs configured | ☐ |

### Phase C — TestFlight live

| # | Check | Pass |
|---|-------|:----:|
| C1 | Archive + upload succeeded | ☐ |
| C2 | Internal testers enabled | ☐ |
| C3 | §8e regression on TestFlight build | ☐ |
| C4 | No blocker crashes 24h | ☐ |

**Definition of done:**

> At least one TestFlight internal build on a real device: onboarding → login → send message → (scoped features) working against the deployed HTTPS backend.

### Release sign-off template

```markdown
## Echo release sign-off — vX.Y.Z — DATE

**Commit SHA:** _______
**TestFlight build:** _______
**Backend URL:** _______

### Feature sets in this release
- [ ] P1-Identity  - [ ] P1-Relay  - [ ] P1-Data-L1  - [ ] P1-Chain
- [ ] Week A messaging  - [ ] Week B contacts
- [ ] Phase 3 signals  - [ ] WO-100 OIDC4VC  - [ ] WO-221 PSI

### Automated (required)
- [ ] CI green on SHA
- [ ] `make regression` passed
- [ ] `make validate-phase1` → GO (all ok)

### Manual simulator
- [ ] §6.1 Onboarding  - [ ] §6.2 Login  - [ ] §6.3 Messaging
- [ ] §6.4 Phase 3 (if shipped)  - [ ] §6.5 OIDC4VC (if shipped)  - [ ] §6.6 PSI (if shipped)

### Manual device / TestFlight
- [ ] §7a device checklist
- [ ] §8e TestFlight regression

**Tester:** _______  **Environment:** dev / staging / prod  **Result:** GO / NO-GO
```

---

## 10. Known issues & workarounds

| Issue | Workaround |
|-------|------------|
| Overwhelming test matrix | [`E2E_QUICK_START.md` §0](E2E_QUICK_START.md) daily table |
| Wrong scheme | Use **EchoApp**, not Echo |
| Face ID in Simulator | Features → Face ID → Enrolled |
| Device `localhost` API | LAN IP in scheme `API_URL` |
| Step 5 validate fail | Rebuild Data L1 JAR — quick start §3b |
| Step 3 skip | `make start-identity` |
| `gomobile` fails | Full Xcode.app |

---

## 11. Command reference

```bash
make dev && make dev-status && make validate-phase1
make regression && make regression-with-phase1
make ios-preflight BUILD=1 TESTS=1
make phase3-signals-proof && make metagraph-test
make start-identity && make stop-identity
```

---

## 12. Related documentation

| Doc | Contents |
|-----|----------|
| [`E2E_QUICK_START.md`](E2E_QUICK_START.md) | Setup, rebuild, daily regression |
| [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) | Phase 3 wiring |
| [`metagraph-backend-e2e-testing.md`](metagraph-backend-e2e-testing.md) | Hydra cluster |
| [`WEEK_A_B_LAUNCH.md`](WEEK_A_B_LAUNCH.md) | Sprint execution |
| [`AGENTS.md`](../AGENTS.md) | Agent skills |

**Skills:** `echo-testing`, `echo-phase1-validate`, `echo-ios-agent-vs-xcode`, `echo-phase3-ios-wire`
