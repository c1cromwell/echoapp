# ECHO — repo review, backend test, and App Store readiness (2026-11-10)

## Follow-up (2026-11-11)

The defects below that were still open in code are now implemented and covered by tests. `make dev` was run: Global L0, Metagraph L0, Currency L1, Data L1, Postgres, Redis, NATS, MinIO, and the gateway are up. `http://localhost:8000/health` returned `operational`. Migration `034_safety_reports.sql` applied on boot. Identity L0/L1 stay down until `make start-identity`; a direct message does not need them.

`go test -race -count=1 ./internal/... ./pkg/... ./test/...` passed. `swift test --filter 'RelayErrorNotice|BarSession'` passed. Unsigned `POST /v1/safety/reports` and `DELETE /v1/users/account` return 401 on the live server.

Rewards ships in v1 as the existing earn / stake / claim tab. There is no “buy ECHO” control. ECHO VIP is a StoreKit subscription, product `com.echo.app.vip.monthly`, with Restore and Manage on the VIP screen. `POST /v1/auth/vip-verify` with `evidence_type=vip_subscription` requires that transaction. Create the same auto-renewable product in App Store Connect before a Release purchase can complete. The local scheme uses `ios/Echo/Configs/EchoVIP.storekit`.

Still yours, and not changed here: Apple team and entitlements, Terms and Privacy URLs (they still point at `https://example.com`), the export-compliance answer, and the iOS 26 deployment target. Two signed-in simulators sending a message still needs your taps.

Scope: Go backend (109K LOC), iOS app (84K LOC Swift, 511 files), Xcode project, deploy config.
Reviewed against the working tree on `main` (111 uncommitted files at the time).

## Verdict

**Not ready to submit.** The backend is now in much better shape than it was this morning, but the
iOS app has open App Review blockers that need your Apple account, your legal URLs, product
decisions, and work I could not build or run from here. The list is in "What still blocks
submission".

## What was tested, and how

| Check | Result |
|---|---|
| `go build ./...` | Pass |
| `go vet ./...` | Pass |
| `gofmt -l .` | Was failing on 3 files (so `make release-check` failed); fixed |
| `go test -race ./internal/... ./pkg/... ./test/...` | Pass before and after my changes |
| Live server: Postgres 16 + Redis, all 35 migrations | Starts, serves, but see crash below |
| Live two-device smoke test (enroll, session, WebSocket, 400-message offline replay, delete account) | Pass after fixes |
| Account purge against the real schema | Pass (`-tags integration`) |

Limits you should know about:

- **`make dev` itself was not run.** It needs Docker and the Euclid/hydra metagraph cluster on your
  Mac; my shell on your machine is a sandbox without Docker or Go. I ran the same backend binary
  against local Postgres and Redis in my own sandbox instead. NATS, MinIO and the metagraph L0/L1
  nodes were absent, so anchoring, staking submission and credential issuance were not exercised.
- **Go 1.26.3, not 1.25.3.** I built the compiler from your Homebrew source; the module cache came
  from `~/go/pkg/mod`.
- **Nothing on iOS was compiled or run.** No Swift toolchain here, and Xcode is view-and-click only
  for me. Every Swift and project-file change below is unbuilt. Screen review is from the existing
  `docs/screen_catalog` PNGs (last regenerated July), not a live simulator.
- The wallet cross-validation tests skipped in my sandbox (they read an iOS test vector I did not
  copy over); they will run on your machine.

## Backend: defects found and fixed

All confirmed against a running server first, then fixed with regression tests. The hub tests were
also run against the original code to confirm they fail there.

1. **Account takeover for any DID (critical).** `POST /v1/auth/restore-did` never checked
   `wallet_signature`, and `/v1/enrollment/wallet` returns the wallet address for any DID with no
   auth. Three anonymous requests returned a valid access token for an arbitrary DID; I used it to
   delete that account. Now the caller must present the wallet public key, it must derive the
   address, and the signature must verify over the single-use challenge. The restore endpoints are
   also IP-throttled.
2. **WebSocket accepted invalid tokens.** A garbage or expired token got `101 Switching Protocols`
   and joined the hub as user `""`, exempt from rate limits and block lists. Now 401.
3. **Any client could broadcast to every connected user.** A frame with no `to` fanned out to all
   sockets. Now dropped. No iOS code path sends such frames.
4. **`/v1/auth/vip-verify` and `/v1/identity/link-wallet` were unauthenticated.** Anyone could set
   any DID's trust tier or rebind its wallet. Both now require the DID's own authenticated request.
5. **Server crashed five minutes after start** whenever Postgres was configured and
   `CURRENCY_L1_URL` was not: a nil `*MetagraphClient` boxed in an interface passed the `== nil`
   guard and the wallet reconciler dereferenced it. Normalised at construction; the reconciler also
   recovers from panics.
6. **Offline messages were lost on reconnect.** The queue claimed up to 1,000 messages as delivered
   and pushed them into a 256-slot buffer; the rest were dropped (the original code delivers 256 of
   700). Replay is now paged, undelivered messages are requeued, and order is restored after
   Postgres `UPDATE … RETURNING`.
7. **Reconnect race silently disconnected users.** An old socket's late unregister evicted the new
   one, leaving the user connected but undeliverable. Also closed a send-on-closed-channel race.
8. **Per-IP rate limit was bypassable** with a made-up `X-Forwarded-For`. The header is now ignored
   unless `TRUSTED_PROXY_HOPS` is set. **Deploy note:** I set it to `"1"` in
   `deploy/k8s/base/configmap.yaml`; with the Hetzner LB in front of ingress-nginx, check the
   comment there or every client will share one bucket.
9. **Account deletion deleted nothing.** `DELETE /v1/users/account` only revoked refresh tokens.
   It now purges 34 tables in one transaction, drops cached keys, revokes tokens and closes the
   socket. What it deliberately leaves (owned groups/channels, votes, staking rows, Comply records,
   media bytes in S3) is listed in `internal/database/account_purge.go` for you to decide.
10. **New installs could never open the realtime socket.** Only account recovery issued an access
    token, it lasts 15 minutes, and no refresh token came with it. First-run enrollment issued none
    at all. Added passkey-signed `POST /v1/auth/session` and an iOS bootstrap that uses it.
11. **Rate-limited messages vanished silently.** The sender now gets an `error` frame
    (`RATE_LIMITED`). iOS does not handle that frame yet.
12. `build/Dockerfile` was gitignored, so `make dev` and the Backend Image workflow fail from a
    clean clone. Un-ignored. Added `ReadHeaderTimeout`/`IdleTimeout` to the HTTP server.

## Backend and protocol gaps not fixed

- **Delivery is at-most-once.** A message handed to a socket that dies before the write is gone;
  there is no client ack. This is the biggest remaining reliability gap versus Signal/WhatsApp.
- **One live socket per account.** The hub maps DID to a single client, so a second device replaces
  the first. Multi-device fan-out does not exist at the transport level.
- **Delivered ciphertext and its metadata stay in `message_queue` for 30 days.** Sender, recipient,
  conversation and timestamps are retained after delivery. Delete on delivery, or keep hours.
- **60 frames/minute covers everything on the socket**, including typing, receipts and WebRTC
  signalling. A call setup plus normal chat can exceed it.
- **Recovery on a new phone is unfinished.** iOS sends a zeroed device key (`createIdentityKey`
  returns `Data(count: 65)`, `PasskeyManager.register` is empty) and the server does not bind a new
  device key on restore, so a restored device cannot sign requests.
- Step-up tokens are minted by a per-request `StepUpService` and never verified on the routes they
  are meant to protect. Access tokens are passed in the WebSocket URL query (they end up in proxy
  logs). 33 handlers decode request bodies with no size limit.
- Many services are in-memory (`MessagingService`, scheduled messages, enrollment wallet map,
  restore challenges): state is lost on restart and not shared across replicas.
- I did not audit the E2EE protocol itself (X25519/ChaCha20, the SX1 ratchet, ML-KEM hybrid,
  sealed sender, safety numbers). That needs a dedicated review before you claim Signal parity.

## iOS: changes made (unbuilt)

- `RecoveryService.swift` sends `wallet_public_key` (required by fix 1).
- `SilentProvisionService.swift`, `VIPSubscriptionStore.swift` sign the `vip-verify` call.
- `MessageRelaySession.swift` gains `SessionTokenBootstrap` (stored token, then refresh, then
  passkey-signed session).
- `VIPSubscriptionView.swift`: the "Subscribe for $9.99/mo" button is DEBUG-only; Release shows
  "Coming soon". The price is still displayed on that screen and in the Settings row.
- Added `Resources/PrivacyInfo.xcprivacy`, `Resources/Assets.xcassets` with an `AppIcon` rendered
  from your ripple mark (there was no asset catalog and no icon at all), and
  `NFCReaderUsageDescription` (the binary links CoreNFC for mDL enrollment).
- `project.pbxproj` edited by script to register the two resources; original saved as
  `project.pbxproj.pre-claude.bak`.

**Build `EchoMessaging` in Xcode before anything else.** If the project fails to open, restore the
`.bak`.

## What still blocks submission

Hard blockers (rejection or upload failure):

1. **No entitlements file.** No `aps-environment`, so push registration fails; no NFC entitlement;
   no app group. There is also no Notification Service Extension target, though
   `NotificationService.swift` exists, and no `voip`/`audio` background mode for CallKit calls.
2. **`DEVELOPMENT_TEAM` is empty** and the bundle ID `com.echo.app` must be one you own.
3. **Terms and Privacy links point to `https://example.com`** (`AuthScreens.swift:168,174`).
   The default auth base URL is `https://api.echo.app/v1`, and `Endpoints.swift` still lists routes
   the backend does not have (`/auth/login`, `/messages/send`, `/users/profile`).
4. **User-generated content rules (1.2).** "Report Contact" sets a flag nothing reads; there is no
   report endpoint, no message or channel reporting, no moderation queue, no published contact.
   Blocking exists. Channels and groups make this mandatory.
5. **Rewards and staking (3.1.5).** A top-level Rewards tab with quests, claims and 8/12/18% APR
   staking is the pattern Apple rejects: tokens for completing tasks, and yield claims. It also
   requires an organisation developer account. Decide whether v1 ships without it.
6. **VIP (3.1.1)** needs StoreKit, or the price must disappear from Release entirely.
7. **Export compliance.** `ITSAppUsesNonExemptEncryption` is unset. I left it for you: it is a legal
   declaration, and the third-party ML-KEM implementation affects the answer.
8. **Deployment target is iOS 26.0** for the app target. Nothing rejects that, but it excludes most
   devices; the package itself declares iOS 17.
9. Privacy manifest and App Privacy labels need your sign-off; `MLKEMNativeSwift` needs its own
   manifest check.

Design and accessibility, from the screen catalog and source:

- **831 fixed `.font(.system(size:))` calls and no Dynamic Type**; 26 `accessibilityLabel`s in the
  whole app; zero localisation. VoiceOver and large-text users are not served.
- **Light mode is forced** in five places; there is no dark appearance.
- The tab bar is custom with 10pt labels, low-contrast inactive icons and hard-coded 28pt bottom
  padding rather than the safe area.
- Rewards, Settings, the VIP screen and the chat thread render blank in the catalog: either the
  fixtures are stale or those screens have no loading or empty state.
- Login reads "FaceID, TouchID" (Apple's names are "Face ID" and "Touch ID") and the legal line is
  about 10pt grey. The message-actions sheet carries the stray line "Yes! End-to-end encrypted on
  this thread."
- `StubProvisionPasskey()` is wired into the production app initialiser.

## Files

Backend changes are applied in the working tree and also saved as
`tmp/claude-stage/echo-backend-fixes.patch`. New tests: `internal/api/security_regression_test.go`,
`internal/api/auth_session_test.go`, `internal/database/account_purge_integration_test.go`,
`internal/wallet/typed_nil_test.go`. Two existing tests that asserted the global broadcast were
rewritten to assert the opposite. Nothing was committed.

`tmp/claude-stage/` holds emptied staging files I could not delete; remove the folder.
