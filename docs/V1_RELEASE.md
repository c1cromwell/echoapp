# Echo Messaging v1.0 — release definition

**Product:** Echo Messaging (`com.echo.app`, scheme `EchoMessaging`, configuration `Release-Messaging`)
**Defined:** 2026-11-14 · supersedes the scope table in [`GO_LIVE_SEPT_1_2026.md`](GO_LIVE_SEPT_1_2026.md) §0
**Status of this build:** code-complete for the scope below; **not yet submittable** — see "Before submission".

v1 is one promise: **a private messenger where you own your identity.** Everything in the build
either delivers that or is hidden. Echo Comply and Echo Passport ship separately.

## What ships

| Area | In v1 | Notes |
|------|-------|-------|
| **Identity & sign-in** | `did:key` identity in the Secure Enclave, passkey sign-in with Face ID / Touch ID, no phone number or email | Frozen first-run and login screens |
| **Recovery** | Restore on a new phone from the recovery phrase; the restored device gets a working signing key | Fixed for v1 (key id mismatch) — needs one manual device pass |
| **1:1 messaging** | End-to-end encrypted text, replies, edits, pins, reactions, typing and read signals, voice notes, attachments, search and archive | Failed messages show "Not sent · Tap to retry" with retry / delete |
| **Groups & channels** | Group chat, broadcast channels with comments | Group chat is functional but less polished than 1:1 (see "Known limits") |
| **Contacts** | Username / QR / invite-link discovery, block, trust badges | |
| **Safety (App Review 1.2)** | Report a contact, message, channel or group; block; operator moderation queue (`GET /v1/admin/safety/reports`); published support contact | Reports survive deletion of the reported account |
| **Privacy controls** | Privacy hub, account deletion with passkey step-up (server purges 35 tables), data export copy | App returns to first run after deletion |
| **Rewards** | Echo Score, quests, earned ECHO, leaderboard, staking positions | **Value-free:** no purchase, no redemption, no transfer, no yield figures |
| **ECHO VIP** | StoreKit auto-renewable `com.echo.app.vip.monthly`, Restore and Manage, price from the App Store only | Server verifies Apple's signature; refunded and (optionally) sandbox purchases are rejected |
| **Integrity anchoring** | Message-batch Merkle roots on Data L1, trust-tier commitments on the Identity metagraph | Invisible to users; verified end to end on the local cluster |

## Hidden in the v1 Release build

Compiled out or gated by `V1ReleaseScope` (`ios/Echo/Sources/Core/Networking/EchoAPIBaseURL.swift`); all remain in Debug builds.

| Hidden | Why |
|--------|-----|
| "% APR" / "APY" on staking tiers and validators | A yield figure is a promise of value. It contradicts the launch posture (non-redeemable, non-transferable, no value promise) and is the pattern App Review rejects under 3.1.5 |
| Emission gauge on the wallet | Token-supply internals; not a user feature |
| VIP benefits not built yet (rate-limit boost, themes / app icon, priority delivery) | A paid subscription may only list what it delivers. Add each row back when the feature ships |
| "Set up a new account", "Wallet & Advanced" (real-funds switch), Developer section | QA and development tools |
| Mock settings screens (Appearance, Storage & Data, old Account screen) | Not implemented; rows are hidden rather than dead |
| In-chat payments (`ECHO_IN_CHAT_PAYMENTS`), real-funds wallet (`ECHO_WALLET_REAL_FUNDS`) | Custody not ready (see wallet custody status) |

## Deferred to v1.1+

Multi-device message fan-out (one live socket per account today) · delivery acknowledgements
(delivery is at-most-once) · cloud chat backup · Dynamic Type across all screens, localisation and
dark mode · Notification Service Extension (rich / decrypted push) · NFC identity verification ·
in-chat tips and payments · private AI · token redemption or transfer of any kind.

## Known limits users can hit in v1

- A message handed to a socket that drops before the write is lost; the sender sees it as sent.
- Signing in on a second device disconnects the first from realtime delivery.
- Group chat has no failed-send state or retry yet.
- "Import from Phone" contacts are not linked to Echo accounts; three of the four contact trust filters never match.
- People without a username appear as a shortened identifier (blocked list, leaderboard).
- A failed attachment has to be picked again.

## Before submission

Owner actions — none of these can be done from the codebase:

1. **Publish the legal pages.** The build links to `https://echo.app/join/terms` and
   `https://echo.app/join/privacy` and mails `support@echo.app`. All three must be live.
2. **Deploy the API at `https://api.echo.app`** (the Release `API_URL`). Add TLS pins to the
   Info.plist array `EchoPinnedSPKIHashes` (command in `CertificatePinner.swift`); until then the app
   uses standard TLS validation, which is safe but unpinned.
3. **Apple account:** set `DEVELOPMENT_TEAM` for `Release-Messaging`; enable Push and App Groups for
   `com.echo.app`; create `com.echo.app.vip.monthly` in App Store Connect; answer export compliance
   (`ITSAppUsesNonExemptEncryption` is `true` in the plist).
4. **Backend config:** `ADMIN_API_TOKEN` (32+ characters) for the moderation queue;
   `VIP_ALLOW_SANDBOX_STOREKIT=true` while in TestFlight / App Review, then `false`;
   `JWT_SIGNING_KEY`; `ECHO_WALLET_GENESIS_AUTO` off; `TRUSTED_PROXY_HOPS` matched to the load balancer.
5. **Metagraph:** production genesis allocation (the local cluster uses the Euclid template's
   30,000 tokens, not the 1B ECHO supply); an authorised-signer check on Data L1 before any public network.
6. **Manual pass on two devices** — [`E2E_TESTING.md`](E2E_TESTING.md) §5, plus: restore from phrase and
   relaunch, each report flow, retry of a failed message, account deletion, VIP purchase in sandbox.
7. **Decide:** deployment target is iOS 26.0, which excludes most devices in use.

## Verification at definition time

| Check | Result |
|-------|--------|
| `go build`, `go vet`, `gofmt`, `go test ./internal/... ./pkg/... ./test/...` | Pass |
| Account-purge SQL against real Postgres (`-tags integration`) | Pass |
| Metagraph Scala tests (all six modules) | Pass |
| `make validate-phase1` on a cluster running Echo's metagraph | 26 / 26, nothing skipped |
| iOS `Debug-Messaging` and `Release-Messaging` simulator builds | Succeed |
| iOS `EchoUnitTests` | 269 run, 0 failures, 2 skipped |
| Two real devices, TestFlight, StoreKit sandbox | **Not run** |
