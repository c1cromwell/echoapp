# Echo iOS — App Store submission checklist

**Product:** Echo Messaging (`com.echo.app`)  
**Status:** In progress — **not ready for external TestFlight or App Store** until P0 items are closed  
**Last reviewed:** 2026-05-29  
**Related:** [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md) §8–9 · [`E2E_QUICK_START.md`](E2E_QUICK_START.md) · [`data-classification.md`](data-classification.md)

Use this as the **release ticket checklist**. Copy unchecked P0 rows into work orders as needed (VIP → WO-286, deletion → WO-228 follow-on).

---

## Rollout phases

| Phase | Target | Gate |
|-------|--------|------|
| **0 — Internal TestFlight** | Messaging QA, team only | P0 blockers closed |
| **1 — External TestFlight** | Beta App Review | Phase 0 + privacy manifest + legal URLs + deletion hardening |
| **2 — App Store** | Public release | StoreKit VIP (if sold) + full data deletion + review notes |

---

## P0 — Blockers (must fix before external TestFlight)

### Payment & IAP (Guideline 3.1.1)

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P0-1 | Hide or remove VIP **“Subscribe $9.99/mo”** until StoreKit ships | ⬜ | iOS | `VIPSubscriptionView.swift` — today calls `vip-verify` without IAP |
| P0-2 | Ensure no other paid digital goods bypass IAP | ⬜ | iOS | VIP themes/rate limits must not be purchasable off-store |
| P0-3 | Keep `ECHO_IN_CHAT_PAYMENTS` **off** in Release schemes | ⬜ | iOS | `WalletTransferAvailability` — real P2P pay is WO-299 |
| P0-4 | Rewards/Wallet UI: no “buy ECHO” or on-ramp copy | ⬜ | iOS/Product | Frame as earned rewards / staking, not purchasable crypto |

### Privacy manifests & declarations

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P0-5 | Add `PrivacyInfo.xcprivacy` to Echo app target | ⬜ | iOS | Required for App Store; declare Required Reason APIs |
| P0-6 | Add `NSPhotoLibraryUsageDescription` | ⬜ | iOS | `AttachmentPickerView` uses `PhotosPicker` |
| P0-7 | Set `ITSAppUsesNonExemptEncryption` or answer consistently in ASC | ⬜ | iOS | TLS + E2E — standard exemption likely applies |
| P0-8 | Replace placeholder legal URLs (`https://example.com`) | ⬜ | iOS/Web | `AuthScreens.swift`, Settings privacy/terms |
| P0-9 | Publish hosted **Privacy Policy** + **Terms** URLs | ⬜ | Legal/Web | Required for external TestFlight |
| P0-10 | Complete App Store Connect **App Privacy** labels | ⬜ | Release | Must match manifest + [`data-classification.md`](data-classification.md) |

### Account deletion (Guideline 5.1.1(v))

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P0-11 | Wire delete through **StepUp** (passkey + OTP) | ⬜ | iOS | `StepUpAction.deleteAccount` — `PrivacyHubView` skips today |
| P0-12 | Backend: delete profile, devices, discovery, wallet cache, push tokens | ⬜ | Backend | `DELETE /v1/users/account` only revokes refresh tokens today |
| P0-13 | Document retention exceptions (E2E msgs on peers’ devices) | ⬜ | Product/Legal | Already in UI copy — align with policy |

---

## P1 — Before external TestFlight

### Sign-in

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P1-1 | Confirm **Sign in with Apple** not required (passkey-only auth) | ✅ | — | No Google/Facebook login in v1 |
| P1-2 | Re-evaluate SiWA if third-party login added | ⬜ | iOS | Guideline 4.8 |

### Privacy — data inventory (ASC questionnaire)

| Data type | Collected? | Linked to user? | Declare in ASC |
|-----------|------------|-----------------|----------------|
| User ID / DID | Yes | Yes | ⬜ |
| Phone (hashed backup / discovery) | Yes | Yes | ⬜ |
| Contacts (on-device hash for PSI) | Optional | Yes | ⬜ |
| Photos / media (chat) | Yes | Yes | ⬜ |
| Audio (voice, calls) | Yes | Yes | ⬜ |
| Device ID (APNs, link-device) | Yes | Yes | ⬜ |
| Biometrics | No (on-device only) | — | Do not declare as collected |
| Tracking / ATT | No | — | Declare “no tracking” unless added |

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P1-3 | Audit SPM deps (e.g. MLKEMNativeSwift) for privacy manifests | ⬜ | iOS | Fix missing third-party signatures if flagged |
| P1-4 | Support URL in App Store Connect | ⬜ | Release | Required metadata |

### Crypto / wallet review narrative

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P1-5 | App Review notes: wallet = rewards/staking, not IAP crypto purchase | ⬜ | Release | Guidelines 3.1.5 / 4.7 |
| P1-6 | Disable dev genesis credit in production (`ECHO_WALLET_GENESIS_AUTO`) | ⬜ | Backend | TestFlight/prod env |
| P1-7 | Bluetooth background modes: justify mesh or remove from v1 build | ⬜ | iOS/Product | `EchoMessaging-Info.plist` UIBackgroundModes |

---

## P2 — App Store release (when selling VIP)

| # | Item | Status | Owner | Notes |
|---|------|--------|-------|-------|
| P2-1 | StoreKit 2 auto-renewable subscription (WO-286) | ⬜ | iOS | Product ID e.g. `com.echo.app.vip.monthly` |
| P2-2 | Server receipt validation + entitlement sync | ⬜ | Backend | Replace free `vip-verify` activation |
| P2-3 | Restore purchases + manage subscription link | ⬜ | iOS | Required for subscriptions |
| P2-4 | Subscription terms, pricing, ASC subscription group | ⬜ | Release | $9.99/mo as designed |
| P2-5 | Real in-chat ECHO transfer (WO-299) — separate security review | ⬜ | Full stack | Before enabling `ECHO_IN_CHAT_PAYMENTS` |

---

## Metadata & App Store Connect completeness

| # | Item | Status | Notes |
|---|------|--------|-------|
| M-1 | Bundle ID `com.echo.app` matches ASC | ⬜ | |
| M-2 | Version / build incremented per upload | ⬜ | Currently 1.0 (1) |
| M-3 | App description (E2E messaging, not crypto exchange) | ⬜ | |
| M-4 | Keywords & category (Social / Utilities) | ⬜ | |
| M-5 | Age rating questionnaire (UGC, contacts → likely 12+) | ⬜ | |
| M-6 | Screenshots 6.7" + 6.1" (no misleading $/crypto) | ⬜ | |
| M-7 | Review notes + demo account / test steps | ⬜ | Passkey onboarding |
| M-8 | Push capability + APNs `.p8` on backend | ⬜ | See E2E §8a |
| M-9 | Associated Domains (if universal links used) | ⬜ | |
| M-10 | Export compliance answered in ASC | ⬜ | |

**In-app discoverability**

| # | Item | Status | Path |
|---|------|--------|------|
| M-11 | Account deletion ≤ 3 taps | ✅ | Profile → Privacy Hub → Data & deletion |
| M-12 | Privacy policy accessible in app | ⬜ | Settings — fix URL |
| M-13 | Terms accessible in app | ⬜ | Settings — fix URL |

---

## Binary validation (every archive)

Run before **Product → Archive**:

```bash
make release-check
make ios-preflight BUILD=1 TESTS=1
# If wallet in build:
./scripts/validate-wallet.sh
make regression-with-phase1   # pre-TestFlight with cluster
```

| # | Check | Status |
|---|-------|--------|
| B-1 | Scheme **Echo (TestFlight)** / Release | ⬜ |
| B-2 | `API_URL` → HTTPS production (not localhost) | ⬜ |
| B-3 | `ECHO_IN_CHAT_PAYMENTS` unset in Release | ⬜ |
| B-4 | Distribution cert + push entitlement | ⬜ |
| B-5 | Organizer **Validate App** clean | ⬜ |
| B-6 | Upload to App Store Connect | ⬜ |
| B-7 | TestFlight §8e smoke ([`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md)) | ⬜ |
| B-8 | 24h crash rate acceptable in Organizer | ⬜ |

Archive command (reference):

```bash
cd ios/Echo
xcodebuild -project EchoApp.xcodeproj -scheme "Echo (TestFlight)" \
  -configuration Release \
  -archivePath /tmp/Echo.xcarchive \
  -destination 'generic/platform=iOS' \
  archive
```

---

## Code references (audit trail)

| Topic | Location |
|-------|----------|
| VIP without StoreKit | `ios/Echo/Sources/Features/Settings/VIPSubscriptionStore.swift`, `VIPSubscriptionView.swift` |
| In-chat payments gate | `ios/Echo/Sources/Features/Payments/WalletTransfer.swift` |
| Account deletion UI | `ios/Echo/Sources/Features/Settings/PrivacyHubView.swift` |
| Account deletion API | `internal/api/account_handlers.go` — `DELETE /v1/users/account` |
| Step-up delete contract | `ios/Echo/Sources/Features/Auth/Models/AuthState.swift` — `StepUpAction.deleteAccount` |
| Placeholder legal URLs | `ios/Echo/Sources/Presentation/Screens/Onboarding/AuthScreens.swift` |
| Info.plist permissions | `ios/Echo/Configs/EchoMessaging-Info.plist` |
| Wallet / staking launch | [`ECHO_WALLET_STAKING_LAUNCH.md`](ECHO_WALLET_STAKING_LAUNCH.md) |

---

## Sign-off template

Copy into release ticket when phase gate is met:

```text
App Store phase: [ 0 Internal TF | 1 External TF | 2 App Store ]
Build: com.echo.app ___.___ (___)
Commit: _______________

P0 blockers: [ ] all closed
Privacy manifest: [ ] merged
Legal URLs: [ ] live
Deletion: [ ] step-up + server purge
VIP/IAP: [ ] hidden OR StoreKit live
Binary validation B-1–B-8: [ ] pass

Approved by: __________  Date: __________
```
