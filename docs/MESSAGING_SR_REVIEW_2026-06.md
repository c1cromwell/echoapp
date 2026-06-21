# Echo Messaging — Senior Engineering Review (2026-06)

Reviewer lens: senior backend engineer + senior iOS engineer.
Scope: end-to-end customer journeys (enrollment, login, send / private message, calls, media,
groups, search), iOS architecture, and backend/metagraph architecture.

> **Methodology note.** An initial automated exploration produced a findings table that mixed real
> bugs with over-claims. Every finding below was **re-verified against current source**; where the
> exploration was wrong, the correction is recorded so the engineering record stays honest. Findings
> still unverified at source are tagged **[verify]**.

---

## 1. Verdict

Solid cryptographic and architectural foundations (Secure Enclave, passkeys, content-blind relay,
DAG anchoring), but the **private-message path is not yet correct on real hardware** and there were
**confidentiality fail-open paths** in the 1:1 send/edit flow. Not production-ready for private
messaging until Wave 2 (crypto correctness) lands and is validated through the Xcode build + a
two-device E2E test.

The good news: the group path is already implemented to the right standard (fail-closed), so it is
the reference pattern for fixing the 1:1 path.

---

## 2. Corrections to the exploration pass (important)

| Exploration claim | Reality at source |
|---|---|
| "Encryption failure sends plaintext" — in `TextMessageCrypto` | **Wrong location, real bug.** `TextMessageCrypto.encryptPayload` always encrypts. The real fail-open was in **`ChatDetailViewModel`** send (≈L382) and edit (≈L204): `catch { TextMessagePayload(text: trimmed, encrypted: nil) }`. **Fixed this session.** |
| "No sender verification on receive" | **Wrong.** `MessageRelayManager.handleIncomingMessage` verifies signature + commitment. Real gap: sender pubkey is read from the message, not resolved from the DID. (Moot — see next row.) |
| "Offline queue blocks / silent-provision blocks first run" | **Mostly wrong.** `SilentProvisionService` runs in the background after the user reaches the Messages empty state; first-run works offline. It already writes `echo.username.current` + `echo.did.current` to Keychain (the prior Face ID login bug is **already fixed**, SilentProvisionService.swift:121-122). |
| Two message paths, both "live" | **One is dead.** `MessageRelayManager` is referenced only by a comment in `ChatDetailViewModel:44` and never instantiated; DI wires only `TextMessageCrypto`. The live path is `ConversationSignalService.sendTextMessage` + `TextMessageCrypto`. |

---

## 3. Journey-by-journey assessment

### Enrollment — **good**
`CredentialEnrollmentTailService` + `SilentProvisionService` run a 4-step backgrounded pipeline
(SE key → DID → wallet → passkey) with retry/backoff and Keychain mirroring. Gaps are polish:
per-step progress UI and wiring the existing `retry()` to a button. Add a regression test asserting
`echo.username.current` is written (guards the Face ID login contract in memory
`project_ios_auth_identity_keys`).

### Login — **works, thin recovery**
`LoginViewModel.loginWithPasskey` does challenge → passkey assertion (Face ID) → server verify →
token store. Real gaps: `AUTH_007` (new device) and `AUTH_009` (lockout) set an `errorMessage` but
there is **no recovery navigation**, and the 15-min hard-lockout countdown (state already in
`SecureEnclaveManager.currentLockState()`) is not surfaced.

### Send / private message — **CRITICAL, device-decrypt now FIXED (Option B)**
- **Device decrypt — FIXED via Option B.** Root cause was: messaging ECDH used the P-256 identity
  key, but on device `echo-identity-signing` is a **non-exportable Secure Enclave key**
  (`kSecAttrTokenIDSecureEnclave`; private bytes never leave the enclave — the T0 invariant), and
  `KinnamiEncryption.decryptWithKeyAgreement` requires a CryptoKit `P256.KeyAgreement.PrivateKey` an
  SE key can never provide — so `loadAgreementPrivateKey()` only had a simulator branch. Fixed by
  introducing a **dedicated software P-256 key-agreement key** (`MessagingAgreementKey`, Keychain,
  device-only) used for ECDH on both Simulator and hardware; its public key is registered as a
  labelled device (`MessagingKeyRegistrar` → `POST /identity/devices`, signed by the identity key)
  and `IdentityResolveClient` now resolves peers' messaging key by that label. Verified: iOS-triple
  library build green; round-trip test (`SecurityTests/MessagingAgreementRoundTripTests`) compiles;
  zero new errors vs. baseline. Remaining gate: two-device runtime decrypt.
- **Fail-open to plaintext (P0) — FIXED this session.** 1:1 send and edit now fail **closed**:
  if encryption fails, the message is marked `.failed` / the edit is not transmitted, matching
  `GroupChatViewModel`. (`ChatDetailViewModel.swift`.)
- **Receive-side downgrade (P1) — FIXED this session.** `TextMessageCrypto.decryptPayload` now
  rejects a plaintext-only payload instead of displaying it (`TextMessageCryptoError.missingCiphertext`).
- **Offline-queue dedup (P1) — FIXED this session.** `OfflineQueueManager.enqueue` is now idempotent
  on `messageId`.
- **Peer-key cache (P2) — FIXED this session.** `peerKeyCache` now has a 1-hour TTL + 256-entry
  bound, so peer key rotation is re-fetched and memory is bounded.

### Calls — **skeleton** [verify]
Live/stub WebRTC engine split; simulator is stub-only. Verify the live engine actually negotiates
DTLS-SRTP (no unencrypted RTP), add ringing timeout, and surface the existing `CallHistoryStore`.

### Media — **partial** [verify]
Encryption + relay exist; add per-chunk integrity hashing on download and derive `trustTier` from
enrollment assurance instead of the hardcoded `3`.

### Groups — **good, needs governance** [verify]
`GroupKeyDistributionService` + `GroupKeyManager` do sealed-key distribution and are **fail-closed**.
Add a key-version field and an admin/permissions model (only admins add/remove + trigger rekey);
verify removed members can't derive the post-rekey key.

### Search — **minimal** [verify]
`LocalMessageIndexer` builds an encrypted inverted index; add pruning/compaction so it can't grow
unbounded.

---

## 4. Architecture findings (backend) [verify against Go source]

- `X-Device-Info` is plaintext-trusted in `internal/auth/middleware.go` → integrity flags
  (jailbreak, etc.) are forgeable. Bind with Apple App Attest / DeviceCheck.
- Metagraph submissions go out **unsigned** when `IDENTITY_SERVICE_KEY_PEM` is absent
  (`internal/metagraph/client.go`) → any backend could submit false commitments. Require signing.
- JWT signing key is loaded once with **no rotation** (`internal/auth/token.go`). Add key
  versioning + a JWKS endpoint.
- Per-DID rate limiting is per-instance → multiplies under horizontal scaling. Move to a
  Redis sliding-window/token-bucket.
- Relay can block on synchronous metagraph calls. Add a circuit breaker so a Data L1 timeout
  can't stall message relay (`internal/api/ws.go`, `internal/services/relay/relay.go`).

---

## 5. The device-decrypt fix (the #1 priority) — two options

**Option A — in-enclave ECDH (hardware-correct).** Keep the identity key in the Secure Enclave; do
key agreement with `SecKeyCopyKeyExchangeResult(seKeyRef, .ecdhKeyExchangeStandard, peerPubSecKey,
params, &err)` and HKDF the result. Requires giving `KinnamiEncryption` a shared-secret / `SecKey`
decrypt path (it currently only accepts a CryptoKit `P256.KeyAgreement.PrivateKey`). Strongest
security; touches the crypto core, so must land behind the Xcode build + two-device E2E test.

**Option B — dedicated messaging key-agreement key (pragmatic).** Generate a separate P-256
KeyAgreement keypair as a software key in the Keychain (device-only, biometric-gated access control),
register its public key for messaging in the identity registry, and use it on **both** simulator and
device. Decouples encryption (exportable KA key) from signing (non-exportable SE identity key) —
a common, clean split. Requires an enrollment change to publish the messaging KA public key and an
`IdentityResolveClient` change to return it.

Recommendation: **Option B** for time-to-correct (uniform sim/device path, no Kinnami rewrite),
with Option A as a hardening follow-up. Either way: **must be validated in Xcode + a real
two-device decrypt test** — do not ship a crypto change validated only by headless build.

---

## 6. What changed this session (safe, isolated, no signature changes)

| File | Change |
|---|---|
| `ios/Echo/Sources/Features/Messaging/ChatDetailViewModel.swift` | Removed plaintext fail-open in 1:1 **send** and **edit**; now fail-closed like the group path. |
| `ios/Echo/Sources/Services/TextMessageCrypto.swift` | `decryptPayload` rejects plaintext-only payloads; `peerKeyCache` gains 1-hour TTL + 256-entry bound. |
| `ios/Echo/Sources/Core/Relay/OfflineQueueManager.swift` | `enqueue` is idempotent on `messageId` (dedup). |

These were chosen because they close real confidentiality/correctness gaps **without** touching the
crypto core or requiring deletions — i.e., low build-break risk in a session where the Xcode gate
isn't available.

## 7. Recommended execution order (waves)

1. **Wave 2 / device decrypt (Option B)** — the one fix that makes private messaging work on
   hardware. Highest priority despite being labeled Wave 2; do it next with the Xcode loop.
2. **Delete dead `MessageRelayManager` + `OfflineQueueManager`/`RelayRequest` cascade** — only with
   the Xcode build available to confirm no hidden references (`AnchoringTracker` is live and stays).
3. **Wave 1 polish** — enrollment progress/retry UI, AUTH_007/009 recovery, lockout countdown.
4. **Wave 3 / 4** — feature completion, then backend hardening, per §3–4.

## 8. Verification gate

- Per change: Xcode build (the real gate — headless `swift build` can't load macros) + unit tests.
- Crypto: a **two-device** enroll → send private message → **decrypt on the second device** test.
  This single test catches the device-decrypt break and any dead-path regression. It does not exist
  yet and should be added with Wave 2.
- Two-device run of `docs/E2E_LAUNCH_AND_TESTING.md` §6.4 for typing/receipts.
