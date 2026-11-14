# Changelog

All notable changes to Echo are documented here.

## [1.0.0] — Echo Messaging v1 release candidate (2026-11-14)

Scope and submission gates: [`docs/V1_RELEASE.md`](docs/V1_RELEASE.md).

### Security

- **StoreKit verification** — ECHO VIP transactions are verified against Apple Root CA G3 (x5c chain, Apple marker OIDs, ES256). Refunded transactions are rejected; Sandbox purchases need `VIP_ALLOW_SANDBOX_STOREKIT`. Self-asserted evidence is capped at tier 1.
- **Step-up** requires a passkey-signed request; one shared service instance.
- **Safety reports** carry message / conversation / channel ids and details (migration 035), have an operator queue behind `ADMIN_API_TOKEN`, and survive deletion of the reported account.
- **Wallet** — bounded enrollment map; `link-wallet` returns 409 for an address linked to another DID.
- **iOS certificate pinning** no longer blocks every Release request: real SPKI pins come from Info.plist, with standard TLS validation when none are set.

### Metagraph

- `make dev` builds and runs Echo's metagraph (it had been running the stock Euclid template); the build fails if the jar is not `com.echo.*`.
- Fixed: SHA-256 roots containing a long run of digits were rejected as phone numbers.
- Finality is reported by Metagraph L0, not on L1 acceptance; the combiner is deterministic.
- Identity L1 runs three nodes (a single node never reached consensus); Identity L0 detects a re-genesis of the Global L0.
- `validate-phase1` fails instead of skipping on chain errors and checks on-chain inclusion.

### iOS

- Restore on a new device provisions the `echo-identity-signing` key the rest of the app signs with.
- Report a message, channel or group; chat "Block & report" reports.
- Failed messages can be retried or deleted; rate-limited sends are marked failed.
- Plain-language errors; empty, loading and error states on Messages, Contacts and Rewards; confirmations on destructive actions; VoiceOver labels on the main journeys.
- VIP price comes only from StoreKit; Terms and Privacy links on the purchase screen.
- Release configuration: production API URL, legal links, production push entitlement.
- v1 scope gates: yield percentages, emission internals and unbuilt VIP benefits are not in Release builds.

## [0.1.0] — Phase 1: Foundation & Security Core (2026-05-11)

Phase 1 delivers the complete security foundation, decentralized identity layer,
and privacy-first onboarding flow. All 25 Phase 1 work orders are implemented.

### Identity & Authentication

- **did:key identity** — P-256 Secure Enclave hardware key derives a W3C DID. No Cardano/PRISM dependency (WO-275, WO-278, WO-279, WO-280).
- **ECDSA P-256 passkey middleware** — every authenticated API request carries `X-Sender-DID` + `X-Signature` over `SHA-256(body)`. Redis cache (60s TTL) backs PostgresDIDRegistry (WO-1, WO-2).
- **Multi-device registration** — QR-code token flow: primary device generates 5-min Redis token; secondary device adds key without a registered signature (WO-273).
- **WebAuthn/FIDO2 passkey setup** — `PasskeySetupView` in onboarding for optional WebAuthn enrollment (WO-136).

### iOS App

- **4-step onboarding** — username → biometric enrollment (mandatory Face ID) → recovery setup → main app. No phone number or email required (Wave 12).
- **Signal-style biometric login** — saved username + Face ID auto-triggers 400ms after login screen appears (Wave 12).
- **Hidden persona gate** — `PersonaGateView` re-prompts Face ID for any `Persona` with `visibility == .hidden`; auto-locks after 2-min background (Wave 12).
- **SecureEnclaveManager** — biometric-bound P-256 keys (`biometryCurrentSet`), HKDF purpose-specific key derivation, biometric lockout counter (5→passcode, 10→15-min hard lock), `BiometricLockoutView` (WO-223, WO-211).
- **Storage key derivation** — `deriveStorageKey()` via HKDF-SHA256 from Secure Enclave signature; monthly rotation; background zeroing (WO-224).
- **LocalDatabase** — SwiftData AES-256-GCM at rest; `unlock(storageKey:)` / `lockStorage()` (WO-224).
- **Privacy architecture** — background key purging, `StorageLockedView`, `PrivacySettingsView` toggles, encryption indicator in chat header (WO-208).
- **FileEncryptionService** — per-file AES-256-GCM with HKDF-SHA256 key derivation from conversation key (WO-9).
- **Privacy-safe `EchoLogger`** — PII scrubbing (hex keys, email, E.164 phone) before `os.Logger` output (WO-6).
- **`CertificatePinner`** — TLS 1.3 with SPKI pinning wired as `URLSession` delegate (WO-2).
- **`PasskeySigningInterceptor`** — signs every API request with Secure Enclave key (WO-1 iOS side).
- **`APIError.rateLimited(retryAfter:)`** — reads `Retry-After` header on 429 (WO-44).

### Cryptography

- **Kinnami P-256 + AES-256-GCM** — iOS ↔ backend matching protocol; Go + Swift implementations cross-verified (WO-13, `internal/crypto/kinnami.go`).
- **X25519 + ChaCha20-Poly1305** — canonical WO-13 backend encryption spec; `GET /v1/crypto/server-key` serves static public key; full iOS→backend round-trip integration tested.
- **KinnamiEncryption (iOS)** — P-256 ECDH + AES-256-GCM for message encryption; T0-T7 annotated.

### Backend

- **Server-side validation** — Merkle root (T5), decay factor (T6), DID registration (T7), governance vote, anti-gaming velocity checks (WO-35).
- **W3C VC 2.0 + StatusList2021** — credential issuance, revocation bit-vector, OIDC4VCI VP verification with real JWT parsing and did:key holder signature verification (WO-274).
- **OIDC4VCI issuer** — mounted at `/.well-known/openid-credential-issuer`; VP submission + verification flow.
- **Rate limiting** — per-DID token bucket (base 100/min, VIP 200/min), `Retry-After` header, `X-RateLimit-Remaining` (WO-44).
- **CORS** — `X-Sender-DID`, `X-Signature`, `Retry-After` headers whitelisted.
- **`FromJWT` / `FromSDJWT`** — credential JWT parsing; SD-JWT disclosure stripping.
- **SMS OTP recovery** — 3 endpoints (`/register`, `/verify`, `/challenge`); backend stores only `H(phone)` (sha256); raw number sent only transiently to SMS provider (Wave 12).
- **`TwilioSMSProvider` scaffold** — reads `TWILIO_*` env vars; `StubSMSProvider` for dev/CI.
- **`LogPublisher`** — in-memory buffer, AES-256-GCM batch encryption, monthly HKDF key rotation, `IPFSStorage` interface, `PinataIPFSStorage` + `StorjIPFSStorage` clients (WO-53, WO-33).
- **Privacy-safe Go `Logger`** — JSON-line structured output with PII regex scrubbing (WO-6).

### Metagraph (Scala)

- **Identity Metagraph L1** — `IdentityValidations`: did:key format, authorized-sender gating, 32-byte hex commitment, StatusList2021 vector monotonicity, org role bounds, future expiry (WO-272).
- **87 ScalaTest assertions** — across pure validators, wired dispatch, and identity validations (WO-277).
- **JDK 21 compatibility** — sbt 1.9.9, Scala 2.13.12, semanticdb-scalac 4.12.3 (Wave 4).

### CI / Compliance

- **T0-T7 Semgrep rules** — 9 Go + 5 Swift patterns enforced in `go-ci.yml`; WARNING/ERROR severity split (WO-217).
- **iOS CI** — hard gate on `swift build --target EchoSecurityTests`; full test suite advisory (WO-217).
- **`docs/data-classification.md`** — complete T0-T7 guide with code examples.

### Infrastructure

- **Phase 1 validate script** — `scripts/validate-phase1.sh` — 6-step go/no-go: did:key derivation → DID registration → Identity L1 health → relay message → Merkle root commit → Global L0 snapshot height.
- **`make dev`** — single-command Euclid SDK Docker cluster (6 containers) + backend.
- **`migrations/010_sms_recovery.sql`** — `sms_recovery` table for phone hash registration.

### Removed

- Cardano/Atala PRISM stack (`pkg/cardano/`, `scripts/setup-cardano-testnet.sh`) — replaced by did:key (WO-280).

---

## [Unreleased] — Phase 2

See `docs/phase-2-work-orders.md` for the Phase 2 backlog (infrastructure scale-out, Kubernetes, production metagraph nodes).
