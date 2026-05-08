# Echo - Product Requirements Document

## Business Problem

Current messaging platforms force users to choose between convenience and privacy, creating significant security vulnerabilities that affect billions of people daily. WhatsApp and Telegram rely on centralized servers that create single points of failure and government surveillance risks, while Signal's focus on privacy comes at the cost of advanced features and user adoption. These platforms cannot provide cryptographic proof of message authenticity, leaving users vulnerable to deepfakes, impersonation attacks, and fraud that costs consumers and businesses billions annually.

The problem extends beyond individual privacy concerns. Financial institutions lose approximately $5 billion yearly to SMS-based phishing attacks because customers cannot verify authentic communications from their banks. Businesses struggle with compliance and audit requirements when using traditional messaging for sensitive communications, as these platforms cannot provide immutable proof of conversations or participant identity verification.

## Vision

A fully decentralized messaging platform that combines:
- **WhatsApp's** seamless UX and media sharing
- **Telegram's** extensibility via bots/channels
- **Signal's** gold-standard privacy
- **X.com-inspired** trust mechanics (verification badges, trusted circles)
- **IRON SPIDR's** blockchain-anchored security (adapted from federal use cases)

## Technical Architecture

### Core Components

| Layer | Technology | Decentralization Level |
|-------|-----------|----------------------|
| Identity & Auth | W3C `did:key`, Identity Metagraph (VCs, StatusList2021), passkeys / Secure Enclave | High |
| Messaging | libp2p, Constellation Hypergraph, Noise Protocol | High |
| Storage | IPFS/Filecoin, OrbitDB | Medium-High |
| Trust Engine | Identity Metagraph commitments; ECHO token logic on Currency metagraph (Phase 1–2) | High |
| Frontend | React Native, Swift, WalletConnect | N/A |

### Key Technical Decisions
- **DIDs (`did:key`)**: Self-certifying identifiers; registration and device bindings via `POST /identity/register` (Phase 1 per ADR-0001 — no pre-ADR-0001 external issuer integration).
- **P2P Messaging**: Relay + E2E encrypted payloads; metagraph for proofs and rewards where applicable.
- **Blockchain / metagraph anchoring**: Message and identity artifacts anchored on Constellation layers per phase roadmap.
- **Zero-Knowledge Proofs**: Privacy-preserving authentication and verification (Phase 3+ evaluation).

## Key Features

### Core Messaging
- 1:1 and group chats (up to 1M members)
- Voice/video calls with screen sharing
- Voice notes, reactions, stickers
- End-to-end encryption by default

### Trust & Verification
- Progressive trust scoring (0-100)
- Verification badges (blue/gold via on-chain credentials)
- Trusted Circles: Inner Circle, Trusted, Acquaintance
- On-chain evidence for reports/blocks

### Privacy & Security
- Disappearing messages with cryptographic verification
- Hidden folders with biometric protection
- Silent and scheduled messages
- No metadata collection

### ECHO Token Economy
- Messaging rewards: 0.1 ECHO per message (capped)
- Payment rail rewards: 1-5 ECHO per transaction
- Referral program: 50 ECHO per verified referral
- Staking: 5-15% APY

## Success Metrics

| Metric | Year 1 Target | Year 2 Target |
|--------|--------------|---------------|
| Monthly Active Users | 100,000 | 1,000,000 |
| Daily Messages/User | 50+ | 75+ |
| 30-Day Retention | 60% | 70% |
| Verified Users | 30% | 50% |
| Enterprise Pilots | 5 | 25 |

## Development Roadmap

### Phase 1: Research & Prototype (1-2 months)
- Validate IRON SPIDR parallels
- Build PoC for `did:key` + Identity Metagraph issuance and P2P / relay messaging
- Security whitepaper

### Phase 2: Core Build (3-5 months)
- Implement messaging stack
- Basic UI development
- Trust scoring contracts
- Alpha release (100 beta users)

### Phase 3: Feature Polish & Launch (2-3 months)
- Bots/channels
- Multi-device sync
- Mainnet launch
- App Store submission

### Phase 4: Scale & Integrate (Ongoing)
- Optimize for 1M+ users
- Bank pilots
- Governance DAO

## Budget Estimate

$500K - $2M for MVP development including:
- Development team (5-10 blockchain + mobile experts)
- Security audits
- Marketing and launch

---

*For detailed API specifications, see [docs/api/openapi.yaml](./api/openapi.yaml)*
