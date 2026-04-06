# Echo - Product Requirements Document (v2.0)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | February 23, 2026 | Resolved messaging transport architecture: client-server relay with decentralized anchoring. Removed libp2p/P2P references. Updated technical architecture, key decisions, privacy section, and roadmap. Added messaging architecture rationale and metadata protection roadmap. |
| 1.0 | February 2026 | Initial PRD |

---

## Business Problem

Current messaging platforms force users to choose between convenience and privacy, creating significant security vulnerabilities that affect billions of people daily. WhatsApp and Telegram rely on centralized servers that create single points of failure and government surveillance risks, while Signal's focus on privacy comes at the cost of advanced features and user adoption. These platforms cannot provide cryptographic proof of message authenticity, leaving users vulnerable to deepfakes, impersonation attacks, and fraud that costs consumers and businesses billions annually.

The problem extends beyond individual privacy concerns. Financial institutions lose approximately $5 billion yearly to SMS-based phishing attacks because customers cannot verify authentic communications from their banks. Businesses struggle with compliance and audit requirements when using traditional messaging for sensitive communications, as these platforms cannot provide immutable proof of conversations or participant identity verification.

## Vision

A messaging platform with decentralized identity, blockchain-anchored security, and end-to-end encryption that combines:

- **WhatsApp's** seamless UX and media sharing
- **Telegram's** extensibility via bots/channels
- **Signal's** gold-standard privacy and E2E encryption model
- **X.com-inspired** trust mechanics (verification badges, trusted circles)
- **IRON SPIDR's** blockchain-anchored security (adapted from federal use cases)

**What makes ECHO decentralized:** ECHO's decentralization comes from three layers that traditional messengers lack entirely. Your identity is self-sovereign (Cardano DIDs — no company owns your account). Your data integrity is blockchain-verified (metagraph consensus — no company can silently alter records). Your message content is mathematically private (E2E encryption — relay servers see only opaque encrypted blobs). The message relay layer uses a client-server model for reliability, but the servers are stateless pipes with no ability to read, alter, or forge message content, and no authority over your identity or data.

## Technical Architecture

### Core Components

| Layer | Technology | Decentralization Level | Notes |
|-------|-----------|----------------------|-------|
| Identity & Auth | Cardano (Veridian/Atala PRISM), KERI | High | Self-sovereign DIDs; no central auth server |
| Message Relay | Go backend WebSocket relay, APNs | Medium | Stateless relay; E2E encrypted; sees only ciphertext |
| Message Integrity | Constellation Hypergraph (Data L1) | High | Merkle roots of message commitments on-chain |
| Storage | IPFS/Storj | Medium-High | Encrypted audit logs; no plaintext stored |
| Trust Engine | Cardano Smart Contracts, Metagraph Data L1 | High | Trust tier on-chain; raw scores off-chain |
| Token Economy | Constellation Metagraph (Currency L1) | High | ECHO token balances, staking, rewards |
| Frontend | Swift (iOS native), SwiftUI | N/A | Secure Enclave integration |

### Key Technical Decisions

- **DIDs on Cardano**: Self-sovereign identity without central auth servers. Users own their identity across applications.
- **E2E Encrypted Relay**: Messages are end-to-end encrypted on the sender's device before transmission. Relay servers transport opaque ciphertext and cannot read, modify, or forge message content. This follows the same model proven by Signal at scale.
- **Blockchain Anchoring**: Message integrity commitments (Merkle roots of hash commitments, never content) are recorded on the Constellation metagraph, providing cryptographic proof of message authenticity and tamper detection.
- **Zero-Knowledge Proofs**: Privacy-preserving authentication and verification. Prove you meet a trust threshold without revealing your exact score. Prove your age without revealing your birthdate.
- **Stateless Backend**: The Go backend is an operational coordinator and hot cache, not an authority. All persistent state lives on-chain (metagraph for app data/rewards, Cardano for identity). PostgreSQL and Redis serve as performance caches only.

### Messaging Architecture Rationale

ECHO uses a client-server relay model rather than peer-to-peer (libp2p) for the following reasons:

**iOS platform constraints make pure P2P unviable.** Apple suspends background network connections within ~30 seconds of app backgrounding. There is no sanctioned workaround. A P2P messenger on iOS cannot receive messages when the app is closed, which is a non-starter for a consumer messaging product. Every P2P messaging project that ships on iOS has had to add server-side relay infrastructure to achieve basic reliability.

**Offline delivery requires store-and-forward infrastructure.** When a recipient's device is offline, someone must hold the encrypted message until the device comes back. In a P2P network, this is handled by relay nodes — which are functionally servers that you operate. The complexity is the same; the reliability is worse.

**Group messaging at scale requires server-side fan-out.** The PRD targets groups of up to 1M members. Coordinating fan-out across a million peers is an unsolved problem in P2P networking. Server-side fan-out is well-understood and scales linearly.

**Push notifications require a server.** APNs (Apple Push Notification service) can only be called from a server-side component. Any iOS messaging app must have server infrastructure for notifications.

**ECHO's decentralization value comes from identity and data integrity, not transport.** The relay server is a "dumb pipe" — it transports encrypted blobs it cannot read. It does not own user identities (Cardano does). It does not control token balances (the metagraph does). It cannot forge messages (clients verify signatures). Removing the server would not meaningfully improve security; it would meaningfully degrade reliability.

**Metadata protection is addressed separately and incrementally:**

| Phase | Metadata Protection | Method |
|-------|-------------------|--------|
| Phase 1–2 | TLS 1.3 transport encryption; server sees sender/recipient DIDs and timestamps | Baseline (comparable to Signal) |
| Phase 3 | Sealed sender | Server knows recipient but not sender (Signal's proven approach) |
| Phase 4 | Federated relay nodes | Multiple independent relay operators; no single operator sees all traffic |
| Phase 4+ | Optional direct P2P for both-online users | Optimization: when both parties are online, relay directly via WebSocket without server hop |

## Key Features

### Core Messaging
- 1:1 and group chats (up to 1M members)
- Voice/video calls with screen sharing
- Voice notes, reactions, stickers
- End-to-end encryption by default (X25519 key agreement + ChaCha20-Poly1305)
- Offline message queuing with encrypted store-and-forward
- Multi-device sync (Phase 3)

### Trust & Verification
- Progressive trust scoring (5 tiers: Unverified → Trusted)
- Verification badges (blue/gold via on-chain Cardano credentials)
- Trusted Circles: Inner Circle, Trusted, Acquaintance
- On-chain evidence for reports/blocks
- Trust tier commitments on-chain (H(score || nonce)); raw scores never on-chain

### Privacy & Security
- Disappearing messages with cryptographic verification of deletion
- Hidden folders with biometric protection (Secure Enclave)
- Silent and scheduled messages
- Minimal metadata collection; sealed sender roadmap (Phase 3)
- Zero PII on any blockchain (enforced by T0–T7 data classification)
- Device-local secrets: passkeys and private keys never leave iOS Secure Enclave

### ECHO Token Economy
- Messaging rewards: 0.1 ECHO per message (daily cap per trust tier)
- Payment rail rewards: 1-5 ECHO per transaction
- Referral program: 50 ECHO per verified referral
- Staking: 5-15% APY across tiered lock durations
- Anti-gaming: trust-multiplied rewards, daily caps, economic micro-fees at scale

## Success Metrics

| Metric | Year 1 Target | Year 2 Target |
|--------|--------------|---------------|
| Monthly Active Users | 100,000 | 1,000,000 |
| Daily Messages/User | 50+ | 75+ |
| 30-Day Retention | 60% | 70% |
| Verified Users (Tier 3+) | 30% | 50% |
| Enterprise Pilots | 5 | 25 |
| Message Delivery Rate | 99.9% | 99.95% |
| End-to-End Finality (on-chain) | < 10s | < 15s |

## Development Roadmap

### Phase 1: Research & Prototype (1-2 months)
- Validate IRON SPIDR parallels
- Build PoC for Cardano DID + E2E encrypted chat via WebSocket relay
- Security whitepaper covering encryption model, relay trust assumptions, and on-chain anchoring
- Constellation metagraph testnet deployment

### Phase 2: Core Build (3-5 months)
- Implement E2E encrypted messaging stack (Kinnami: X25519 + ChaCha20-Poly1305)
- Go backend relay services with WebSocket + APNs push notifications
- iOS native app with Secure Enclave integration and SwiftUI
- Trust scoring: Cardano credential issuance, metagraph trust commitments
- Metagraph Data L1 + Currency L1 custom validation logic
- Offline message queuing (encrypted store-and-forward)
- Alpha release (100 beta users)

### Phase 3: Feature Polish & Launch (2-3 months)
- Sealed sender implementation (metadata protection)
- Bots/channels framework
- Multi-device sync with device-linked key management
- Group messaging optimization (server-side fan-out, on-chain group metadata)
- Client-side verification: Merkle proofs for message anchoring, snapshot hash verification
- Mainnet launch (Constellation metagraph + Cardano mainnet)
- App Store submission

### Phase 4: Scale & Integrate (Ongoing)
- Federated relay nodes (multiple independent operators)
- Optional direct P2P for both-online users
- Optimize for 1M+ users (metagraph sharding, relay scaling)
- Bank pilots with compliance audit trail (IPFS/Storj encrypted logs)
- Governance DAO (stake-weighted voting on protocol upgrades)
- ZK proof system for privacy-preserving verification
- Android support (StrongBox equivalent of Secure Enclave)

## Budget Estimate

$500K - $2M for MVP development including:

- Development team (5-10 blockchain + mobile experts)
- Security audits (E2E encryption, Secure Enclave integration, metagraph validation logic)
- Constellation metagraph node infrastructure
- Cardano transaction fees (credential issuance from platform treasury)
- IPFS/Storj pinning costs
- Marketing and launch

---

*For detailed API specifications, see [docs/api/openapi.yaml](./api/openapi.yaml)*
*For data layer architecture, see [DATA_LAYER_ARCHITECTURE.md](./DATA_LAYER_ARCHITECTURE.md)*
*For iOS frontend architecture, see [ios-frontend-architecture-blueprint-v2.md](./ios-frontend-architecture-blueprint-v2.md)*
*For backend architecture, see [BACKEND_ARCHITECTURE_IMPLEMENTATION.md](./BACKEND_ARCHITECTURE_IMPLEMENTATION.md)*
