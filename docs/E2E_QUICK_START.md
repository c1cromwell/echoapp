# Echo iOS E2E — Quick start

**One page** for TestFlight prep and day-to-day iOS testing. Full detail: [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md).

**Onboarding & login are frozen** — validate the shipped `FirstRunCoordinator` / `GlacialLoginScreen` flows below; do **not** redesign from the React prototype (`docs/Echo Design System Setup_latest/.../onboarding/*`). See [`ux-spec.md`](ux-spec.md) and [`ECHO_IOS_UI_IMPLEMENTATION_SPEC.md`](ECHO_IOS_UI_IMPLEMENTATION_SPEC.md) §0.

---

## Split: agents vs you

| Who | What | Command |
|-----|------|---------|
| **Agent / CI** | Go + iOS unit tests, backend health, compile gate | `make regression` or `make ios-preflight` |
| **You (Xcode)** | Face ID, onboarding taps, two-client chat, TestFlight | ~15 min checklist below |

Agents **cannot** drive the Simulator UI or real biometrics. They *can* prove the stack is up and the app compiles before you open Xcode.

---

## 1. Agent gate (run first)

```bash
make dev                  # backend stack
make ios-preflight        # health + Xcode + scheme checks
make ios-preflight BUILD=1 TESTS=1   # + simulator build + SPM tests
```

**Cursor agent:** MCP `echo-local-dev` → `run_ios_preflight`, `health_backend`, `run_regression`.

If preflight shows **FAIL**, fix that before Xcode — usually `make dev` or `sudo xcode-select -s /Applications/Xcode.app/Contents/Developer`.

---

## 2. You in Xcode (~15 min)

### Setup (once per machine)

1. Open **`ios/Echo/EchoApp.xcodeproj`** (scheme **`EchoApp`**, not "Echo").
2. **Simulator:** scheme already sets `API_URL=http://localhost:8000`.
3. **Physical iPhone:** Product → Scheme → Edit Scheme → Run → Environment → `API_URL` = `http://<Mac-LAN-IP>:8000` (from `ipconfig getifaddr en0`). Same Wi‑Fi as Mac.
4. Simulator: **Features → Face ID → Enrolled**.

### Smoke path (new user) — canonical flow only

Exercise **only** the shipped screens (Welcome → display name → options/Face ID → recovery). Legacy/demo routes (`NameAndKeyView`, `WelcomeCarouselView`) and prototype onboarding pages are **out of scope** for QA and agent work.

| # | Tap / action | Pass if |
|---|----------------|---------|
| 1 | Cold launch | Welcome screen |
| 2 | Set up → enter username | Availability check OK |
| 3 | Continue → Face ID | BiometricEnrollment completes |
| 4 | Recovery → **SMS backup** | Enter phone → OTP (see dev OTP below) → Continue |
| 5 | Main app / Messages tab | No crash; tab bar visible |
| 6 | Settings → Privacy → discovery **ON** | Toggle saves |
| 7 | Contacts → **Contacts on ECHO** → Scan | List or empty state (no crash) |
| 8 | If match → **Add** | Contact added |

**Dev SMS OTP:** `.env` → `DEV_MODE=true`, restart `make dev`. After Send code, read header:

```bash
curl -sD - -o /dev/null -X POST http://localhost:8000/v1/auth/sms-recovery/register \
  -H 'Content-Type: application/json' \
  -d '{"phone_hash":"sha256:abc","phone_raw":"+12125551234","did":"did:key:zTest"}' \
  | grep -i x-dev-otp
```

### Returning user

Force-quit → relaunch → Face ID login → app unlocks.

### Two-client message relay (Wave 0.1)

1. **User A:** New conversation → search **@username** of User B → open chat → send a message.
2. **User B:** Same (or accept invite) — thread id is deterministic from both DIDs.
3. **Pass if:** B sees the message in the open chat (or hub preview updates); Messages tab stays connected to `make dev` WebSocket.

Both devices/simulators need auth tokens (completed onboarding) and the same backend `API_URL`.

### Profile QR (identity share + add contact)

Profile → identity card → tap **QR icon** → share link or **Scan** another user’s code → contact added via `POST /v3/contacts/add` (`added_via: qr_scan`) → open Messages to chat.

### Phase 3 signals (two clients)

Settings → Privacy → toggle **typing indicators** / **read receipts** (persisted locally). Two signed-in clients in the same thread:

| Signal | Pass if |
|--------|---------|
| Typing | Recipient sees indicator; stops on send / idle |
| Read receipts | Open chat → sender checkmarks advance to read |
| Reactions | Long-press 👍 toggles; matches REST |

Detail: [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) Step 5, [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md) §6.4.

### Encrypted message bodies

Outbound chat uses Kinnami P-256 encryption when `GET /identity/resolve/{peerDid}` returns a device key. **Simulator ↔ simulator** decrypt works with software identity keys; physical device decrypt is a follow-up (Secure Enclave agreement export).

### Optional (pre-TestFlight)

- **WO-100 wallet:** `OIDC4VC_ENABLED=true` in `.env` → VIP path → Digital ID.
- **WO-221 PSI (real):** `./scripts/build-echooprf-ios.sh` + embed `EchoOPRF.xcframework` in EchoApp target.
- **Phase 3 signals:** two simulators or sim + device — typing / read receipts ([`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) Step 5).

---

## 3. Troubleshooting (iOS)

| Symptom | Fix |
|---------|-----|
| `cannot find type … in scope` / build fails in Xcode | New Swift file not in **EchoApp** target — add in Project Navigator |
| Registration / API errors on **device** | `API_URL` must be Mac LAN IP, not `localhost` |
| Registration 400 `DID_KEY_MISMATCH` | Pull latest iOS — needs canonical `did:key` (`DidKeyDeriver`) |
| Face ID never prompts | Simulator → Features → Face ID → Enrolled → **Matching Face** |
| SMS code never arrives | `DEV_MODE=true` + use `X-Dev-OTP` header, or configure Twilio |
| Contact discovery always empty | SMS backup done, discovery opt-in ON, both users in phone book, OPRF framework for real PSI |
| `xcodebuild` / `gomobile` fails | Full **Xcode.app**; `sudo xcode-select -s /Applications/Xcode.app/Contents/Developer` |
| Agent `swift test` fails on macOS | Known SPM cross-target noise — use `make ios-preflight BUILD=1` or Xcode build instead |

---

## 4. Launch sign-off (minimal)

Before TestFlight upload:

- [ ] `make regression` green
- [ ] `make ios-preflight BUILD=1` green (or zero FAIL lines)
- [ ] §2 smoke path on **simulator**
- [ ] §2 smoke path on **physical device** (LAN `API_URL`)
- [ ] `make validate-phase1` all steps `ok` (full go/no-go — see full E2E doc)

Full sign-off table: [`E2E_LAUNCH_AND_TESTING.md` §9](E2E_LAUNCH_AND_TESTING.md#9-launch-checklist--sign-off).

---

## Related

| Resource | Use |
|----------|-----|
| [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md) | Full tiers, env vars, TestFlight upload |
| Skill **`echo-testing`** | Agent automation map |
| Skill **`echo-ios-agent-vs-xcode`** | What agents vs Xcode own |
| MCP **`echo-local-dev`** | `run_ios_preflight`, `run_regression`, `health_backend` |
