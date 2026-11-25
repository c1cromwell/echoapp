# Echo — Privacy-First Decentralized Messaging

[![Go CI](https://github.com/c1cromwell/echoapp/actions/workflows/go-ci.yml/badge.svg)](https://github.com/c1cromwell/echoapp/actions/workflows/go-ci.yml)
[![iOS CI](https://github.com/c1cromwell/echoapp/actions/workflows/ios-ci.yml/badge.svg)](https://github.com/c1cromwell/echoapp/actions/workflows/ios-ci.yml)

**Phase 1 complete. Version 0.1.0.**

Echo is an end-to-end encrypted messaging platform. Message integrity and identity records are anchored, as hashes only, on Echo's own private Hyperledger Besu chain (QBFT, four validators). No phone number. No email. Your Secure Enclave P-256 key *is* your identity.

---

## Security Architecture

```
Secure Enclave P-256 key  ←  biometric-bound (Face ID / Touch ID required)
        │
        ├─ ECDSA sign(SHA-256(body))      →  X-Sender-DID + X-Signature on every request
        ├─ derive DID                      →  did:key: identity, anchored as a hash on Echo's chain
        ├─ HKDF("echo-storage-encryption") →  SwiftData AES-256-GCM at rest
        └─ sign("persona-access-{id}")     →  biometric re-auth gate for hidden personas

Message encryption:  X25519 ECDH + ChaCha20-Poly1305  (WO-13 canonical spec)
Credential format:   W3C VC 2.0 + StatusList2021 revocation  (WO-274)
On-chain identity:   EchoIdentity contract on Echo's Besu chain (hashes only)
Privacy invariant:   Zero PII on any blockchain — T0–T7 classification enforced by CI
```

---

## Quick Start

### Prerequisites

- Go 1.21+
- Docker + Docker Compose
- Node.js (only to recompile the contracts: `make contracts`)
- Xcode 16+ (for iOS)

### Backend

Full setup (three products, clean DB, all services): **[`CONTRIBUTING.md`](CONTRIBUTING.md)**.

Daily regression and Xcode smoke: **[`docs/E2E_QUICK_START.md`](docs/E2E_QUICK_START.md)**.

```bash
make install-deps
make dev                  # Messaging: local chain (4 Besu validators) + gateway :8000
make dev-comply           # Comply API :8011
make start-identity       # Identity L0/L1 for VC / Passport
make validate-phase1      # Phase 1 go/no-go
make release-check
```

### iOS

```bash
cd ios/Echo
swift build --target Echo           # Library build (hard gate)
swift build --target EchoSecurityTests  # Security test build (hard gate)
swift test --filter EchoSecurityTests   # Run security tests
```

---

## Phase 1 Feature Summary

| Area | Delivered |
|---|---|
| **Identity** | did:key derivation, POST /identity/register, multi-device QR token flow |
| **Auth** | ECDSA P-256 passkey middleware, JWT fallback, biometric auto-login (iOS) |
| **Onboarding** | 4-step: username → Face ID enrollment → recovery setup (phrase + SMS) |
| **Hidden personas** | Biometric re-auth gate, per-persona HKDF encryption keys |
| **Credentials** | W3C VC 2.0 issuance, StatusList2021, OIDC4VCI VP verification |
| **Chain** | Echo's Besu network: anchoring, identity records as hashes, proofs anyone can verify offline (`cmd/verifyproof`) |
| **Encryption** | Kinnami P-256/AES-256-GCM (iOS↔backend), X25519/ChaCha20-Poly1305 |
| **Privacy** | T0–T7 classification CI (Semgrep), background key purging, PII-safe logging |
| **Recovery** | 24-word BIP-39 phrase + SMS OTP (H(phone) only stored on backend) |
| **Rate limiting** | Per-DID token bucket (100/min base, 200/min VIP), Retry-After header |
| **Logging** | AES-256-GCM batch encryption, Pinata/Storj IPFS clients, HKDF monthly rotation |

---

## Project Structure

```
.
├── cmd/                    CLI tools (did:key derivation, credential issuance)
├── docs/                   ADRs, blueprints, work orders (all 7 phases)
├── internal/
│   ├── api/                HTTP router, passkey auth middleware, handlers
│   ├── crypto/             Kinnami + X25519/ChaCha20 implementations
│   ├── database/           PostgreSQL + in-memory DB interfaces
│   ├── infra/              Redis, rate limiter, SMS provider, IPFS storage
│   ├── logging/            LogPublisher, PII-safe Logger
│   ├── besu/               Client for Echo's chain: anchoring, proofs, checkpoints
│   ├── ledger/             Anchoring service and proof store in front of the chain client
│   └── validation/         T0–T7 pre-validation (rewards, governance, data L1)
├── chain/contracts/        Solidity contracts (EchoAnchor, EchoIdentity, EchoReserve)
├── deploy/besu/            Local chain, host install, backup/restore, key ceremony
├── migrations/             PostgreSQL migration files
├── pkg/
│   ├── credentials/        W3C VC 2.0, StatusList2021, OIDC4VCI
│   └── didkey/             did:key derivation + P-256 signing
├── ios/Echo/               Swift/SwiftUI iOS app (SPM package)
├── scripts/                validate-phase1.sh, setup helpers
└── test/                   Integration + load + tokenomics tests
```

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `API_PORT` | `8000` | HTTP listen port |
| `DATABASE_HOST` | — | PostgreSQL host (in-memory if unset) |
| `REDIS_HOST` | — | Redis host (features disabled if unset) |
| `BESU_RPC_URLS` | — | Echo chain validators' JSON-RPC URLs, comma-separated |
| `BESU_ANCHOR_CONTRACT` / `BESU_IDENTITY_CONTRACT` | — | Contract addresses (written by `make dev` locally) |
| `BESU_VALIDATORS` | — | Validator addresses proofs are checked against |
| `LOG_MASTER_KEY` | — | 32-byte hex key for log batch encryption |
| `TWILIO_ACCOUNT_SID` | — | Twilio SID for SMS OTP (stub if unset) |
| `TWILIO_AUTH_TOKEN` | — | Twilio auth token |
| `TWILIO_FROM` | — | Twilio sender E.164 number |
| `PINATA_API_KEY` | — | IPFS pinning via Pinata |
| `STORJ_ACCESS_KEY` | — | Storj S3-compatible storage |
| `DEV_MODE` | `false` | Echo OTP in X-Dev-OTP header (never in prod) |

---

## License

Proprietary — Echo © 2026 Chad Cromwell. All rights reserved.
