# Echo - Requirements Documentation

## Table of Contents

### Overview Documents
- [Business Problem](#business-problem)
- [Current State](#current-state)
- [Product Description](#product-description)
- [Personas](#personas)
- [Success Metrics](#success-metrics)
- [Technical Requirements](#technical-requirements)
- [Product Features](#product-features)
- [Development Roadmap](#development-roadmap)
- [Architecture](#architecture)

### Feature Requirements
- [Decentralized Identity and Authentication](#decentralized-identity-and-authentication)
- [Blockchain-Anchored Messaging with Provable Integrity](#blockchain-anchored-messaging-with-provable-integrity)
- [Dynamic Trust Network and Social Verification](#dynamic-trust-network-and-social-verification)
- [Voice and Video Calls with Screen Sharing](#voice-and-video-calls-with-screen-sharing)
- [Large File Sharing and Cloud Storage Integration](#large-file-sharing-and-cloud-storage-integration)
- [Message Reactions, Polls, and Interactive Elements](#message-reactions-polls-and-interactive-elements)
- [Advanced Message Search and Archive System](#advanced-message-search-and-archive-system)
- [Hidden Folders with Biometric Protection](#hidden-folders-with-biometric-protection)
- [Silent and Scheduled Private Chats](#silent-and-scheduled-private-chats)
- [Disappearing Messages with Cryptographic Verification](#disappearing-messages-with-cryptographic-verification)
- [Public and Private Groups with Verified Status Display](#public-and-private-groups-with-verified-status-display)
- [Multiple Personas with Selective Visibility](#multiple-personas-with-selective-visibility)
- [Broadcast Channels and Community Features](#broadcast-channels-and-community-features)
- [Enterprise Organization Profiles with Verified Status](#enterprise-organization-profiles-with-verified-status)
- [Verified Financial Institution Integration](#verified-financial-institution-integration)
- [User Rewards Tracker on Profile](#user-rewards-tracker-on-profile)
- [Streamlined Onboarding with Verifiable Credentials and Passkeys](#streamlined-onboarding-with-verifiable-credentials-and-passkeys)
- [In-App High-Assurance Identity Verification and Reward](#in-app-high-assurance-identity-verification-and-reward)
- [Decentralized Bot Framework and Automation](#decentralized-bot-framework-and-automation)
- [Platform Roadmap and Future Vision](#platform-roadmap-and-future-vision)
- [Universal Onboarding and Identity Creation](#universal-onboarding-and-identity-creation)
- [Privacy-Preserving Contact Discovery](#privacy-preserving-contact-discovery)
- [Privacy Architecture and Secure Data Handling](#privacy-architecture-and-secure-data-handling)
  - [Secure Enclave Key Management](#secure-enclave-key-management)
  - [End-to-End Message Encryption and Commitment](#end-to-end-message-encryption-and-commitment)
  - [Privacy-Preserving Blockchain Data Model](#privacy-preserving-blockchain-data-model)
  - [Zero-Knowledge Proofs and Midnight Integration](#zero-knowledge-proofs-and-midnight-integration)
- [ECHO Tokenomics, Founder Allocation, and Token Launch](#echo-tokenomics-founder-allocation-and-token-launch)
- [Production Launch, Infrastructure, and Deployment](#production-launch-infrastructure-and-deployment)

---

# Overview Documents

## Business Problem

# Echo - Product Requirements Document (v2.5)

## Changelog

| Version | Date | Changes |
| --- | --- | --- |
| 2.5.1 | March 31, 2026 | Consolidated document: removed v2.4 duplicate section, resolved tokenomics conflict (auto-scaling model adopted, daily caps removed), fixed P2P references in feature specs to reflect relay architecture, wrote Secure Enclave Key Management spec, wrote Privacy Architecture overview spec, wrote Privacy-Preserving Blockchain Data Model spec, wrote ZK Proofs and Midnight Integration spec, removed stale v1.0 content. Revised onboarding: username + passkey (zero PII) replaces phone-first model; phone verification becomes optional Tier 2 upgrade. Added Contact Discovery and Enterprise Fraud Prevention features. |
| 2.5 | March 26, 2026 | Revised founder allocation: CEO 10%, co-founders 2% each (18% total), treasury reduced to 22%. Added trust-tier weighted governance model (single-token, no separate governance token). Added Midnight blockchain evaluation roadmap (Cardano now, Midnight Phase 3+). |
| 2.4 | March 7, 2026 | Finalized tokenomics: 1B total supply (40% community, 25% treasury, 15% founders, 10% future team, 10% ecosystem). Founder allocation model: visionary-led split (40/25/15/10/10), 4-year vesting with 1-year cliff via on-chain TokenLock. Resolved wallet question: native ECHO Wallet built on Stargazer SDK (replaces "rewards page" concept). Added Wallet tab architecture. |
| 2.3 | March 7, 2026 | Aligned with Constellation ecosystem: Tessellation v3 transaction primitives, Digital Evidence for enterprise compliance, PacaSwap DEX integration, cross-chain bridges (Base, Ink), validator slashing, L0 token standard. Updated technical architecture table. Updated token economy with v3 staking model. Added Digital Evidence to Organization plans. Updated roadmap phases 2–5 with ecosystem integration milestones. |
| 2.2 | March 1, 2026 | Added long-term vision: community-owned Network State. Added Phase 5 (AI-managed treasury, ECHO burn + BTC reserve, VIP/Org revenue model) and Phase 6 (Network State formation, real-world asset acquisition, legal entity structure). Added governance structure (5 founders + 5 elected board). Added revenue model section. |
| 2.1 | March 1, 2026 | Resolved Constellation deployment strategy: public Hypergraph mainnet metagraph with permissioned L1 validators (hybrid model). Added DAG staking requirements (750K DAG minimum), Scala/JVM consideration, metagraph cost model, snapshot fee economics, and phased testnet → mainnet → delegation → permissionless roadmap. Updated budget estimate. |
| 2.0 | February 23, 2026 | Resolved messaging transport architecture: client-server relay with decentralized anchoring. Removed libp2p/P2P references. Updated technical architecture, key decisions, privacy section, and roadmap. Added messaging architecture rationale and metadata protection roadmap. |
| 1.0 | February 2026 | Initial PRD |

---

## Business Problem

Current messaging platforms force users to choose between convenience and privacy, creating significant security vulnerabilities that affect billions of people daily. WhatsApp and Telegram rely on centralized servers that create single points of failure and government surveillance risks, while Signal's focus on privacy comes at the cost of advanced features and user adoption. These platforms cannot provide cryptographic proof of message authenticity, leaving users vulnerable to deepfakes, impersonation attacks, and fraud that costs consumers and businesses billions annually.

The problem extends beyond individual privacy concerns. Financial institutions lose approximately $5 billion yearly to SMS-based phishing attacks because customers cannot verify authentic communications from their banks. Businesses struggle with compliance and audit requirements when using traditional messaging for sensitive communications, as these platforms cannot provide immutable proof of conversations or participant identity verification.

## Vision

A messaging platform with decentralized identity, blockchain-anchored security, and end-to-end encryption that combines:

* **WhatsApp's** seamless UX and media sharing
* **Telegram's** extensibility via bots/channels
* **Signal's** gold-standard privacy and E2E encryption model
* **X.com-inspired** trust mechanics (verification badges, trusted circles)
* **IRON SPIDR's** blockchain-anchored security (adapted from federal use cases)

**What makes ECHO decentralized:** ECHO's decentralization comes from three layers that traditional messengers lack entirely. Your identity is self-sovereign (Cardano DIDs — no company owns your account). Your data integrity is blockchain-verified (metagraph consensus — no company can silently alter records). Your message content is mathematically private (E2E encryption — relay servers see only opaque encrypted blobs). The message relay layer uses a client-server model for reliability, but the servers are stateless pipes with no ability to read, alter, or forge message content, and no authority over your identity or data.

### Long-Term Vision: Community-Owned Network State

ECHO's endgame is not a messaging company — it is a **community-owned digital nation**. The messaging platform is the foundation that creates daily engagement, shared identity, and collective economic power. Over time, ECHO evolves from a product users consume into an organization all users co-own:

**Phase 1–4 (Product**): Build a world-class encrypted messaging platform with 1M+ daily active users. Every user earns ECHO tokens through participation. Token holders govern the protocol through stake-weighted voting.

**Phase 5 (Economy):** Launch revenue streams (VIP subscriptions, organization plans, payment rail fees). All revenue flows to a community treasury managed by AI agents — no human executives skimming overhead. The treasury executes two annual programs: ECHO token burns (deflationary pressure) and Bitcoin reserve accumulation (hard-asset backing).

**Phase 6 (Network State):** The community, now economically self-sustaining, begins acquiring real-world assets — land, buildings, companies, infrastructure — for members to use. A governance board (5 founders + 5 community-elected members) oversees strategic decisions. ECHO becomes a digitally-native society with physical territory, shared resources, and democratic governance.

This is the Balaji Srinivasan Network State thesis applied to a messaging platform: start with a highly engaged digital community, build collective economic power, and progressively acquire sovereignty in the physical world.

## Technical Architecture

### Core Components

| Layer | Technology | Decentralization Level | Notes |
| --- | --- | --- | --- |
| Identity & Auth | Cardano (Veridian/Atala PRISM), KERI | High | Self-sovereign DIDs; no central auth server |
| Identity Privacy (Phase 3+) | Midnight (ZK-SNARKs, Compact) | High | ZK credential verification: prove trust tier / KYC without revealing data. Cardano partner chain. |
| Message Relay | Go backend WebSocket relay, APNs | Medium | Stateless relay; E2E encrypted; sees only ciphertext |
| Message Integrity | Constellation Hypergraph (Data L1) | High | Merkle roots of message commitments anchored on public Hypergraph |
| Storage | IPFS/Storj | Medium-High | Encrypted audit logs; no plaintext stored |
| Trust Engine | Cardano Smart Contracts, Metagraph Data L1 | High | Trust tier on-chain; raw scores off-chain |
| Token Economy | Constellation Metagraph (Currency L1) | High | ECHO as L0 token on public Hypergraph; v3 primitives (TokenLock, StakeDelegation, AtomicAction) |
| Token Primitives | Tessellation v3 (TokenLock, StakeDelegation, AtomicAction, AllowSpend, WithdrawLock, FeeTransaction) | High | Native Hypergraph types for staking, delegation, swaps, payments; interoperable with Stargazer/PacaSwap |
| Metagraph Validation | Scala/JVM (Euclid SDK, Tessellation) | High | Custom L1 validation logic; validator slashing in Phase 4; permissioned → permissionless |
| Enterprise Evidence | Constellation Digital Evidence (managed API) | High | SHA-256 fingerprinting with public verification explorer; Smart Checkmark; court-admissible compliance |
| DeFi / Liquidity | PacaSwap DEX, Base bridge, Ink bridge | High | AMM liquidity pools (ECHO/DAG, ECHO/USDC); atomic cross-metagraph swaps; CEX access via bridges |
| Frontend | Swift (iOS native), SwiftUI | N/A | Secure Enclave integration |

### Key Technical Decisions

* **DIDs on Cardano**: Self-sovereign identity without central auth servers. Users own their identity across applications.
* **E2E Encrypted Relay**: Messages are end-to-end encrypted on the sender's device before transmission. Relay servers transport opaque ciphertext and cannot read, modify, or forge message content. This follows the same model proven by Signal at scale.
* **Blockchain Anchoring**: Message integrity commitments (Merkle roots of hash commitments, never content) are recorded on the Constellation metagraph, providing cryptographic proof of message authenticity and tamper detection.
* **Zero-Knowledge Proofs**: Privacy-preserving authentication and verification. Prove you meet a trust threshold without revealing your exact score. Prove your age without revealing your birthdate.
* **Stateless Backend**: The Go backend is an operational coordinator and hot cache, not an authority. All persistent state lives on-chain (metagraph for app data/rewards, Cardano for identity). PostgreSQL and Redis serve as performance caches only.

### Constellation Metagraph Deployment Strategy

**Decision: Public Hypergraph Mainnet with Permissioned L1 Validators (Hybrid Model)**

ECHO deploys as a public metagraph on Constellation's Hypergraph mainnet, not a private chain. This is a deliberate choice:

* **Public verifiability is ECHO's value proposition.** ECHO token supply, distribution, and reward claims are publicly auditable by anyone with a block explorer. A private chain would mean "trust us, the balances are real" — the exact problem ECHO exists to solve.
* **IRON SPIDR precedent.** ECHO's PRD cites IRON SPIDR as inspiration. IRON SPIDR started as a private permissioned chain and deliberately transitioned to public. Constellation's own leadership states the future is public networks.
* **Ecosystem network effects.** Public metagraph means ECHO token is visible in Stargazer wallet, tradeable on PacaSwap DEX, eligible for DAG delegation programs, and interoperable with other metagraphs on the Hypergraph. Private chain requires building all tooling from scratch.
* **Privacy is handled at the application layer, not the chain layer.** ECHO's metagraph only stores Merkle roots (hashes), trust commitments (H(score||nonce)), and token transactions. No PII, no message content. The public Hypergraph sees only opaque hashes — privacy is already preserved by design.

**What "hybrid" means in practice:** L1 validators are permissioned (project-operated) during Phases 1–3, controlling who validates ECHO-specific business logic (reward caps, anti-gaming, Merkle structure). L0 nodes submit snapshots to the public Global L0 for immutable recording. Phase 4 opens L1 validation to community operators with ECHO token staking requirements.

**Node Requirements:**

| Node Type | Count | DAG Staking | Role |
| --- | --- | --- | --- |
| L0 Hybrid Nodes | 3 minimum | 250K DAG each (750K total) | Run both Global L0 and Metagraph L0; submit snapshots to Hypergraph |
| Currency L1 Validators | 3–5 initially | Set by ECHO (e.g., minimum ECHO token stake) | Validate ECHO token transactions, rewards, staking |
| Data L1 Validators | 3–5 initially | Set by ECHO (e.g., minimum ECHO token stake) | Validate Merkle roots, trust commitments, governance |

**Cost Model:**

| Cost Item | Estimate | Notes |
| --- | --- | --- |
| DAG staking (3 L0 nodes) | 750K DAG (not spent — staked, recoverable) | Capital lockup; L0 nodes earn DAG validator rewards |
| Snapshot fees | Variable; offset by DAG delegation | \~288 snapshots/day at 100K users; fees burned in DAG |
| Node infrastructure | 3 servers minimum (Ubuntu 22.04, 8+ cores, 32GB RAM) | Hetzner dedicated (primary) or bare metal |
| Scala developer | 1 developer for L1 validation logic | Euclid SDK is Scala/JVM; Go backend and iOS unchanged |

**Snapshot Fee Economics:** End users pay zero fees. ECHO as a project pays snapshot fees in DAG to the Hypergraph for each snapshot submitted by the L0 nodes. More delegated DAG staked to ECHO's validators = lower net snapshot fees (delegators subsidize). At scale, fees can potentially be fully rebated through sufficient delegation.

**Technology Stack Note:** All metagraph L1 validation logic (custom consensus, business rules) must be written in Scala using the Euclid SDK / Tessellation framework. This is the code that enforces ECHO-specific rules: daily reward caps, trust-tier multiplier validation, Merkle root structure checks, anti-gaming rules. The Go backend submits data to the metagraph via its REST API. The iOS app is unaffected.

### Messaging Architecture Rationale

ECHO uses a client-server relay model rather than peer-to-peer (libp2p) for the following reasons:

**iOS platform constraints make pure P2P unviable.** Apple suspends background network connections within \~30 seconds of app backgrounding. There is no sanctioned workaround. A P2P messenger on iOS cannot receive messages when the app is closed, which is a non-starter for a consumer messaging product. Every P2P messaging project that ships on iOS has had to add server-side relay infrastructure to achieve basic reliability.

**Offline delivery requires store-and-forward infrastructure.** When a recipient's device is offline, someone must hold the encrypted message until the device comes back. In a P2P network, this is handled by relay nodes — which are functionally servers that you operate. The complexity is the same; the reliability is worse.

**Group messaging at scale requires server-side fan-out.** The PRD targets groups of up to 1M members. Coordinating fan-out across a million peers is an unsolved problem in P2P networking. Server-side fan-out is well-understood and scales linearly.

**Push notifications require a server.** APNs (Apple Push Notification service) can only be called from a server-side component. Any iOS messaging app must have server infrastructure for notifications.

**ECHO's decentralization value comes from identity and data integrity, not transport.** The relay server is a "dumb pipe" — it transports encrypted blobs it cannot read. It does not own user identities (Cardano does). It does not control token balances (the metagraph does). It cannot forge messages (clients verify signatures). Removing the server would not meaningfully improve security; it would meaningfully degrade reliability.

**Metadata protection is addressed separately and incrementally:**

| Phase | Metadata Protection | Method |
| --- | --- | --- |
| Phase 1–2 | TLS 1.3 transport encryption; server sees sender/recipient DIDs and timestamps | Baseline (comparable to Signal) |
| Phase 3 | Sealed sender | Server knows recipient but not sender (Signal's proven approach) |
| Phase 4 | Federated relay nodes | Multiple independent relay operators; no single operator sees all traffic |
| Phase 4+ | Optional direct P2P for both-online users | Optimization: when both parties are online, relay directly via WebSocket without server hop |

## Key Features

### Core Messaging

* 1:1 and group chats (up to 1M members)
* Voice/video calls with screen sharing
* Voice notes, reactions, stickers
* End-to-end encryption by default (X25519 key agreement + ChaCha20-Poly1305)
* Offline message queuing with encrypted store-and-forward
* Multi-device sync (Phase 3)

### Trust & Verification

* Progressive trust scoring (5 tiers: Unverified → Trusted)
* Verification badges (blue/gold via on-chain Cardano credentials)
* Trusted Circles: Inner Circle, Trusted, Acquaintance
* On-chain evidence for reports/blocks
* Trust tier commitments on-chain (H(score || nonce)); raw scores never on-chain

### Privacy & Security

* Disappearing messages with cryptographic verification of deletion
* Hidden folders with biometric protection (Secure Enclave)
* Silent and scheduled messages
* Minimal metadata collection; sealed sender roadmap (Phase 3)
* Zero PII on any blockchain (enforced by T0–T7 data classification)
* Device-local secrets: passkeys and private keys never leave iOS Secure Enclave

### ECHO Token Economy

**Total Supply: 1,000,000,000 ECHO (fixed, deflationary via Phase 5 burns)**

| Allocation | % | Tokens | Purpose |
| --- | --- | --- | --- |
| Community Rewards | 40% | 400M | Messaging rewards, referrals, staking APY, governance — emitted over 10 years via declining curve |
| Treasury | 22% | 220M | PacaSwap liquidity, DAG staking, Digital Evidence subscriptions, operations, Phase 5–6 |
| Founders (5) | 18% | 180M | 4-year vesting, 1-year cliff, on-chain TokenLock (see Tokenomics doc) |
| Future Team & Advisors | 10% | 100M | Reserved for recruits; same vesting terms |
| Ecosystem & Partnerships | 10% | 100M | PacaSwap LP incentives, DAG delegator rewards, Constellation grants, exchange listings |

**Founder Allocation (18% = 180M ECHO):**

| Founder | Role | % Supply | ECHO |
| --- | --- | --- | --- |
| Founder 1 | CEO / Visionary / Product | 10.0% | 100M |
| Founder 2 | CTO / Lead iOS Engineer | 2.0% | 20M |
| Founder 3 | Scala / Blockchain Lead | 2.0% | 20M |
| Founder 4 | Head of Growth / Community | 2.0% | 20M |
| Founder 5 | Head of Design / UX | 2.0% | 20M |

All founder tokens are held in on-chain TokenLock positions with 1-year cliff + 36-month monthly vesting, publicly visible on DAG Explorer. No founder can sell any tokens for the first 12 months. After cliff, 1/36th of remaining allocation vests monthly. The blockchain is the cap table.

**Reward Mechanics:**

* ECHO is an L0 token on the Constellation Hypergraph, conforming to the Tessellation v3 L0 token standard
* Messaging rewards: 0.1 ECHO per message (daily cap per trust tier, claimed via AtomicAction bundling tier verification + claim + cap update)
* Payment rail rewards: 1-5 ECHO per transaction
* Referral program: 50 ECHO per verified referral
* Staking via native v3 primitives: users lock ECHO in their own Stargazer wallet (TokenLock), delegate to L1 validators (StakeDelegation), earn 5-15% APY by tier, 14-day withdrawal cooldown (WithdrawLock)
* Anti-gaming: trust-multiplied rewards, daily caps, economic micro-fees at scale; all enforced atomically via AtomicAction
* Year 1 emission: 80M ECHO (20% of community pool); declining annually over 10 years
* PacaSwap DEX: ECHO/DAG and ECHO/USDC liquidity pools for trading and treasury operations
* Phase 5 marketplace payments: time-limited AllowSpend approvals (no unlimited token approvals)

### ECHO Wallet (Stargazer SDK)

ECHO includes a native decentralized wallet built on the Constellation Stargazer Wallet SDK, replacing the concept of a "rewards page" with true asset ownership. The wallet is a primary tab in the iOS app alongside Messaging and Profile.

**Why a wallet, not a rewards page:** A rewards page implies gamification points inside someone else's app. A wallet implies real assets the user owns, controls, and can use across the Constellation ecosystem. For a project whose core value proposition is "all users are owners," the wallet framing is essential.

**Wallet Features:**

* Balance display: available, staked (TokenLock), delegated, pending rewards, USD equivalent
* Staking: lock ECHO via TokenLock, choose tier (Bronze 30d/5%, Silver 90d/8%, Gold 180d/12%, Platinum 365d/15%)
* Delegation: browse validators (uptime, commission, delegated stake), delegate via StakeDelegation, switch instantly
* Rewards: claim pending rewards via AtomicAction, daily cap progress bar, trust tier multiplier display
* Swap (Phase 3+): ECHO ↔ DAG and ECHO ↔ USDC via PacaSwap integration
* Bridge (Phase 3+): ECHO → Base, ECHO → Ink for broader DeFi and exchange access
* Founder vesting display (founders only): allocated, vested, locked, next unlock date, cliff status, "View on DAG Explorer" link
* Transaction history: all staking, delegation, reward, swap, and bridge activity

**External wallet compatibility:** Users can also view and manage ECHO in standalone Stargazer wallet or D'Cent hardware wallet. The ECHO iOS wallet and Stargazer share the same underlying Constellation keypair.

### Revenue Model (Phase 5+)

ECHO is free for all users. Revenue comes from premium tiers and payment rails. All revenue flows to the community treasury — not to a corporation.

| Revenue Stream | Source | Estimated Unit | Treasury Allocation |
| --- | --- | --- | --- |
| **VIP Subscriptions** | Individual users opting for premium features (larger groups, priority relay, enhanced storage, custom themes, advanced bots, extended disappearing message options) | $4.99–$9.99/month | 100% to treasury |
| **Organization Plans** | Businesses and teams needing compliance audit trails, branded channels, SLAs, admin controls, SSO integration, **Digital Evidence Smart Checkmark on messages, court-admissible audit fingerprinting via Constellation Digital Evidence API, compliance dashboard with public verification URLs** | $10–$50/seat/month | 100% to treasury |
| **Payment Rail Fees** | Small percentage on in-app fiat-to-ECHO conversions, ECHO-to-fiat off-ramps, and merchant payment processing | 0.5–1.5% per transaction | 100% to treasury |
| **Marketplace/Bot Platform** | Revenue share from third-party bots, integrations, and marketplace transactions | 15–30% platform fee | 100% to treasury |

**Key principle:** The platform itself never extracts value. There are no shareholders, no dividends to a parent company, no executive compensation beyond what governance approves. Every dollar of revenue enters the treasury and is allocated by community governance.

### Treasury Management (Phase 5+)

The treasury is managed by AI agents operating under policies set by community governance votes. Human oversight comes from the governance board (see below).

**AI Agent Responsibilities:**

| Agent Role | Operations | Human Override |
| --- | --- | --- |
| **Treasury CFO Agent** | Cash flow monitoring, budget tracking, financial reporting dashboards, surplus calculation | Board reviews quarterly reports |
| **ECHO Burn Agent** | Executes scheduled ECHO token buybacks and burns per governance-approved ratio | Board can pause in emergency |
| **BTC Reserve Agent** | Dollar-cost-averages treasury surplus into Bitcoin per governance-approved allocation | Board approves annual allocation % |
| **Stablecoin Manager** | Manages operational reserves in stablecoins (USDC/DAI), handles yield optimization on idle reserves | Board sets risk parameters |
| **Compliance Agent** | Monitors regulatory requirements, flags transactions needing review, generates audit reports | Board + legal counsel review flags |
| **Reporting Agent** | Generates public real-time treasury dashboards, monthly reports, annual audit preparation | All reports public by default |

**Annual Treasury Allocation (governance-decided, example starting ratios):**

| Allocation | % of Annual Surplus | Purpose |
| --- | --- | --- |
| ECHO Token Burn | 30% | Reduce circulating supply; deflationary pressure |
| Bitcoin Reserve | 30% | Hard-asset backing; long-term store of value |
| Operational Reserve | 20% | Infrastructure costs, node operations, security audits, development grants |
| Real-World Asset Fund | 15% | Phase 6: land, buildings, companies (Network State assets) |
| Emergency Fund | 5% | Minimum 12-month operating runway in stablecoins |

These ratios are set by annual governance vote and can be adjusted. AI agents execute within the approved ratios; deviations require board approval.

### Governance Structure

**Single-Token Governance:** ECHO is the sole token for utility and governance. No separate governance token. Plutocracy is prevented through trust-tier weighted voting (see below), not token splitting.

**Ownership:** All ECHO token holders are owners. Governance votes are weighted by staked ECHO × trust tier multiplier. This ensures governance power reflects both economic commitment and verified community participation.

**Governance Weight Formula:**

```
Weight = StakedECHO × TrustTierMultiplier

Tier 1 (Unverified):  ×0.0  (no governance)
Tier 2 (Newcomer):    ×0.5
Tier 3 (Member):      ×1.0
Tier 4 (Verified):    ×1.5
Tier 5 (Trusted):     ×2.0
```

A whale who buys 50M ECHO but never verifies (Tier 1) gets zero governance power. The CEO's 100M staked ECHO at Tier 5 gives 200M effective weight — but 10,000 Tier 5 community members each staking 10K ECHO also produce 200M effective weight. The community can always outvote any individual at scale.

**Requirements to vote:** Must have staked ECHO (TokenLock), must be Tier 2+, one vote per DID per proposal, staked tokens (including founder vesting locks) are eligible.

**Governance Board (10 seats):**

| Seats | Selection | Term | Role |
| --- | --- | --- | --- |
| 5 Founders | Permanent (years 1–5); advisory with veto only on existential matters after year 5 | Permanent → advisory | Strategic direction, protocol safety, veto on existential changes (e.g., abandoning E2E encryption) |
| 5 Community Board Members | Elected annually by token-weighted vote (Trust Tier 3+ eligible to stand) | 1 year, re-electable | Oversee treasury AI agents, approve RWA acquisitions, set annual treasury allocation ratios, represent community interests |

**Decision Authority:**

| Decision Type | Who Decides | Threshold |
| --- | --- | --- |
| Protocol upgrades (metagraph schema, encryption changes) | All token holders (governance vote) | 67% supermajority |
| Annual treasury allocation ratios | All token holders (governance vote) | Simple majority |
| ECHO burn / BTC buy execution | AI agents (within approved ratios) | Automatic |
| Real-world asset acquisition &gt; $100K | Board (10 members) + governance ratification | Board 7/10 + 60% governance vote |
| Real-world asset acquisition < $100K | Board (10 members) | Board 6/10 majority |
| Emergency protocol changes | Founders (3/5 multi-sig) | 3-of-5 founders |
| Board member removal (misconduct) | All token holders (governance vote) | 75% supermajority |

**Legal Structure (Phase 6):** The DAO requires a legal entity to hold real-world assets (a DAO cannot directly own land in most jurisdictions). Recommended structure: DAO → Wyoming DAO LLC or Marshall Islands DAO LLC → Real-World Asset Holdings. The legal entity is controlled by the governance board, which is controlled by the DAO. All asset titles are held by the legal entity on behalf of the community. Structure to be finalized with legal counsel before first RWA acquisition.

**Open Source:** The entire ECHO codebase — iOS app, Go backend, Scala metagraph validation logic — is open-sourced under a permissive license (MIT or Apache 2.0) once the core product reaches stability (target: Phase 3). Open source ensures no single entity can capture the platform, and allows the community to fork if governance fails.

## Success Metrics

| Metric | Year 1 Target | Year 2 Target | Year 3+ Target |
| --- | --- | --- | --- |
| Monthly Active Users | 100,000 | 1,000,000 | 5,000,000 |
| Daily Messages/User | 50+ | 75+ | 100+ |
| 30-Day Retention | 60% | 70% | 75% |
| Verified Users (Tier 3+) | 30% | 50% | 60% |
| Enterprise Pilots | 5 | 25 | 100 |
| Message Delivery Rate | 99.9% | 99.95% | 99.99% |
| End-to-End Finality (on-chain) | < 10s | < 15s | < 15s |
| VIP Conversion Rate | — | 5% | 10% |
| Treasury AUM | — | $1M+ | $10M+ |
| Annual ECHO Burned | — | Governance-set | Governance-set |
| BTC Reserve | — | First accumulation | $1M+ BTC |
| Governance Participation Rate | — | 15% of token holders | 25% of token holders |
| Real-World Assets Held | — | — | First acquisition |

## Development Roadmap

### Phase 1: Research & Prototype (1-2 months)

* Validate IRON SPIDR parallels
* Build PoC for Cardano DID + E2E encrypted chat via WebSocket relay
* Security whitepaper covering encryption model, relay trust assumptions, and on-chain anchoring
* **Constellation metagraph testnet deployment** using Euclid SDK (Scala)
* Develop and test Data L1 + Currency L1 custom validation logic on testnet (no real DAG required)
* **Implement Tessellation v3 transaction types** in Currency L1 Scala code: TokenLock, StakeDelegation, WithdrawLock, AtomicAction for reward claims, FeeTransaction for snapshot fees
* Acquire or plan acquisition of 750K+ DAG for mainnet L0 node staking
* Evaluate PacaSwap liquidity bootstrapping requirements for ECHO token launch

### Phase 2: Core Build (3-5 months)

* Implement E2E encrypted messaging stack (Kinnami: X25519 + ChaCha20-Poly1305)
* Go backend relay services with WebSocket + APNs push notifications
* iOS native app with Secure Enclave integration and SwiftUI
* Trust scoring: Cardano credential issuance, metagraph trust commitments
* Finalize Data L1 + Currency L1 validation logic in Scala (reward caps via AtomicAction, anti-gaming, Merkle validation, TokenLock/StakeDelegation staking)
* **Deploy metagraph to Constellation Hypergraph mainnet** — 3 L0 hybrid nodes (750K DAG staked), project-operated L1 validators
* ECHO token goes live on public Hypergraph as L0 token; visible in Stargazer wallet and DAG Explorer
* **Seed ECHO/DAG liquidity pool on PacaSwap** — liquidity bootstrapping event for price discovery and initial trading
* **Automate snapshot fee payment** via FeeTransaction from treasury DAG reserves
* Offline message queuing (encrypted store-and-forward)
* Confirm D'Cent hardware wallet compatibility for ECHO cold storage
* Alpha release (100 beta users)

### Phase 3: Feature Polish & Launch (2-3 months)

* Sealed sender implementation (metadata protection)
* Bots/channels framework
* Multi-device sync with device-linked key management
* Group messaging optimization (server-side fan-out, on-chain group metadata)
* Client-side verification: Merkle proofs for message anchoring, snapshot hash verification
* **Begin DAG delegation campaign** — attract DAG holders to delegate to ECHO validators for lower snapshot fees; offer ECHO token incentives to delegators
* **Create ECHO/USDC liquidity pool on PacaSwap** — stablecoin on/off ramp for users and treasury
* **ECHO ↔ Base brid**ge — coordinate with 3A DAO to add ECHO as bridgeable L0 token; enables Aerodrome DeFi and treasury BTC accumulation path
* **Digital Evidence integration** — Go backend submits media fingerprints for optional user-initiated image/video verification; prepare enterprise API client for Phase 5 Org tier
* **Midnight evaluation** — assess stability after 6+ months of mainnet; proof-of-concept ZK trust tier verification ("Prove I'm Tier 3+ without revealing my credential")
* Cardano mainnet deployment (identity layer)
* App Store submission

### Phase 4: Scale & Integrate (Ongoing)

* **Open L1 validators to community** — any operator meeting minimum ECHO TokenLock stake can run a Currency L1 or Data L1 validator (L0 nodes still require 250K DAG)
* **Activate validator slashing** — fraudulent validation, double-signing, extended downtime; slashed ECHO to treasury
* **ECHO ↔ Ink brid**ge — connect to Kraken exchange via Ink L2; pursue Kraken listing for ECHO
* Federated relay nodes (multiple independent operators, registered on Data L1 with TokenLock stake)
* Optional direct P2P for both-online users
* Optimize for 1M+ users (additional L1 validator nodes, relay scaling)
* Bank pilots with compliance audit trail (IPFS/Storj encrypted logs + Digital Evidence fingerprinting)
* Governance DAO (trust-tier weighted voting on protocol upgrades, metagraph schema changes, slashing thresholds — governance weight = StakedECHO × TrustTierMultiplier)
* **Midnight integration** — ZK trust tier verification live on Midnight mainnet; Org-tier clients get private KYC proofs, group membership proofs, compliance verification without data exposure
* ZK proof system for privacy-preserving verification (via Midnight Compact contracts)
* Android support (StrongBox equivalent of Secure Enclave)
* Explore cross-metagraph interoperability via Hypergraph
* Optional in-app PacaSwap swap interface (ECHO ↔ DAG, ECHO ↔ USDC without leaving ECHO app)

### Phase 5: Community Economy (Year 2–3)

*Prerequisite: 500K+ MAU, stable governance DAO operational*

* **Launch VIP subscriptions and Organization plans** — premium features, compliance tools, enterprise SLAs
* **Organization tier includes Digital Evidence** — Smart Checkmark on messages, automated audit fingerprinting via Constellation Digital Evidence API, compliance dashboard with public verification URLs, data retention proof
* **Deploy AI treasury agents** — CFO agent, burn agent, BTC reserve agent, stablecoin manager, compliance agent, reporting agent
* **AI Burn Agent uses PacaSwap** — buys ECHO from ECHO/DAG pool via atomic swaps, then burns (reduces circulating supply)
* **AI BTC Reserve Agent uses cross-chain bridges** — ECHO → Base bridge → Aerodrome (USDC) → CEX → BTC → cold storage multi-sig
* **AI Stablecoin Manager** uses ECHO/USDC PacaSwap pool and Base bridge for operational reserve management
* **FeeTransaction automation** — AI CFO Agent maintains DAG reserves and pays snapshot fees automatically
* All revenue flows to on-chain community treasury (transparent, auditable on DAG Explorer)
* Community votes on first annual treasury allocation ratios (burn %, BTC %, operational %, RWA fund %, emergency %)
* Launch public real-time treasury dashboard (AI-generated, on-chain verifiable via DAG Explorer)
* Payment rail integration — AllowSpend + SpendTransaction for subscription auto-renewals, bot payments, marketplace escrow (time-limited approvals only, never unlimited)
* Bot/integration marketplace with revenue share to treasury
* Elect first 5 community board members (annual election, Trust Tier 3+ eligible)
* Open-source entire codebase (iOS, Go, Scala) under permissive license
* Engage legal counsel for DAO LLC formation (Wyoming or Marshall Islands)

### Phase 6: Network State Formation (Year 3+)

*Prerequisite: 1M+ MAU, self-sustaining treasury, legal entity established*

* **Establish legal entity** (DAO LLC) to hold real-world assets on behalf of the community
* **First real-world asset acquisition** — community votes on target (co-working space, community housing, or similar high-utility asset for members)
* Expand RWA portfolio based on community governance: land, buildings, companies, infrastructure
* Network State membership tiers — physical access tied to ECHO token staking levels
* Partnerships with existing Network State projects and digital nomad communities
* Cross-metagraph alliances — interoperability agreements with complementary Hypergraph metagraphs
* Explore sovereign recognition pathways (special economic zones, free zones, charter cities)
* Scale AI agent layer: property management agent, investment analysis agent, member services agent
* Annual board elections become a signature community event
* Long-term goal: ECHO community as a recognized digital jurisdiction with physical presence across multiple geographies

## Budget Estimate

**Phase 1–4 (Product Build): $500K - $**2M

* Development team (5-10 blockchain + mobile experts, including at least 1 Scala/JVM developer for metagraph L1 validation logic)
* Security audits (E2E encryption, Secure Enclave integration, metagraph validation logic, Scala L1 code review)
* **750K DAG staking** for 3 L0 hybrid nodes on Constellation Hypergraph mainnet (capital lockup, not expenditure — recoverable; nodes earn DAG validator rewards)
* Constellation metagraph node infrastructure (3+ servers, \~$300-500/month)
* Constellation snapshot fees in DAG (offset by delegation; estimated low at launch volumes)
* Cardano transaction fees (credential issuance from platform treasury, \~15,000 ADA/month at 100K users)
* IPFS/Storj pinning costs (\~$70/month at 100K users)
* Marketing and launch

**Phase 5 (Community Economy): Self-Funding**

Once VIP subscriptions, Organization plans, and payment rail fees are generating revenue, ECHO becomes self-sustaining. All operational costs (infrastructure, security audits, development grants) are funded from treasury. The development team transitions from founder-funded to treasury-funded via governance-approved budgets.

**Phase 6 (Network State): Treasury-Funded**

Real-world asset acquisitions are funded from the RWA allocation of treasury surplus. Scale depends entirely on revenue growth and community governance decisions. No external fundraising required — the community funds its own expansion.

**Note on external funding:** ECHO is designed to *not require* venture capital. VC funding creates misaligned incentives — investors want returns, which means extracting value from users. ECHO's model is the opposite: all value stays in the community. If early-stage funding is needed before revenue, it should come from token presale to aligned community members, Constellation ecosystem grants, or founder capital — not from VCs who would expect equity or governance control.

## Current State

The current messaging landscape is dominated by centralized platforms that require users to trust corporate intermediaries with their communications and personal data. WhatsApp, owned by Meta, serves over 2 billion users but stores message metadata on centralized servers and has faced criticism for data sharing practices with parent company Facebook. While WhatsApp offers end-to-end encryption, users cannot verify message authenticity or prove conversations occurred without relying on Meta's infrastructure.

Telegram positions itself as a privacy-focused alternative but only encrypts "secret chats" by default, leaving most communications vulnerable on centralized servers. Its large group capabilities and bot ecosystem attract users seeking advanced features, but the platform's centralized architecture creates single points of failure and government pressure points, as demonstrated by various national bans and content restrictions.

Signal represents the gold standard for privacy-focused messaging, implementing robust end-to-end encryption protocols that have been adopted industry-wide. However, Signal's user base remains limited due to its austere feature set, lack of advanced functionality like large groups or bots, and dependence on centralized servers for message routing and user discovery.

IRON SPIDR, developed by Constellation Network for U.S. federal agencies, demonstrates the potential for blockchain-anchored secure communications but remains restricted to government use cases. Its architecture provides cryptographic proof of message integrity and participant identity through distributed ledger technology, but lacks the consumer-friendly interface and feature richness needed for mainstream adoption.

Current solutions force users to choose between security, features, and usability. No existing platform combines blockchain-verified identity, cryptographic message provability, decentralized infrastructure, and the rich feature set that modern users expect from messaging applications.


## Personas

ECHO serves multiple user personas, each with distinct needs, motivations, and use cases for the platform.

### Privacy-Conscious Individual User

Early adopters who prioritize digital privacy and security. Frustrated with mainstream messaging apps that monetize user data or lack cryptographic guarantees. Values self-sovereign identity, E2E encryption, and provable message integrity. Willing to learn new technologies for superior privacy. Motivated by ECHO token rewards and governance participation. Likely to become Inner Circle members and advocates for the platform.

### Institutional Financial Services User

Employees at banks and financial institutions who need fraud-proof customer communications. Currently loses money to SMS phishing attacks and lacks cryptographic proof of authentic communications. Requires compliance audit trails, selective disclosure capabilities, and integration with existing banking systems. Values ECHO's verification system and blockchain-anchored message integrity. Phase 5 Organization plan customer with Digital Evidence requirements.

### Enterprise Organization Administrator

IT leaders and compliance officers managing secure communications for businesses. Needs SSO integration, admin controls, SLAs, and court-admissible audit trails. Frustrated with consumer messaging apps that lack enterprise features and compliance tooling. Values ECHO's Digital Evidence Smart Checkmark, public verification URLs, and data retention proof. Willing to pay per-seat pricing for Organization tier.

### DeFi Power User

Sophisticated crypto users managing cross-chain positions and high-value transactions. Needs confidential execution to avoid MEV extraction, frontrunning, and strategy copying. Values optional privacy for financial messaging and payment rails. Interested in Phase 4+ Near Protocol Confidential Intents integration for institutional-grade confidential execution. Active in governance and staking.

### Developer and Bot Creator

Third-party developers building integrations, bots, and applications on ECHO's platform. Motivated by marketplace revenue share, treasury grants, and open-source contribution rewards. Values ECHO's extensibility framework, API access, and decentralized bot infrastructure. Contributes to ecosystem growth through innovative features and integrations.

### Validator Node Operator

Technical operators running L0 hybrid nodes and L1 validators for network security. Requires capital for DAG staking (250K DAG per L0 node) and ECHO token staking for L1 validation. Earns DAG validator rewards, ECHO staking yields (5-15% APY), and snapshot fee rebates through delegation. Motivated by sustainable passive income and supporting network decentralization. Transitions from project-operated (Phase 1-3) to community-operated (Phase 4+).

### Network State Founding Member

Long-term token holders committed to ECHO's evolution into a community-owned digital nation. Participates in governance votes, serves on elected board positions, and contributes to real-world asset acquisition decisions. Values collective economic power, shared resources, and digital sovereignty. Sees ECHO as more than a product—a movement toward decentralized community ownership.

### Expanded Success Criteria

Success for this decentralized messaging platform will be measured through a combination of user adoption metrics, security effectiveness indicators, and ecosystem health measurements that demonstrate the product's ability to solve the core problems of centralized messaging while maintaining user satisfaction and network growth.

User adoption and engagement metrics will track the platform's ability to compete with established messaging apps while providing superior security and decentralization benefits. We will monitor monthly active users with a target of reaching 100,000 users within the first year and 1 million users by year two, focusing on organic growth through word-of-mouth and privacy-conscious user communities. Daily message volume per user should exceed 50 messages to indicate the platform serves as a primary communication tool rather than a secondary privacy option. User retention rates must demonstrate that the enhanced security features do not compromise usability, with 30-day retention rates targeting 60% and 90-day retention rates of 40%, comparable to successful messaging platforms.

The effectiveness of the trust and verification system will be measured through fraud prevention and spam reduction metrics that validate the core value proposition. We will track the percentage of users who complete progressive verification steps, targeting 70% of active users achieving basic verification and 30% completing advanced credentials within six months of registration. Spam and fraud incident rates should remain below 0.1% of total messages, significantly lower than traditional platforms, while maintaining user satisfaction scores above 4.2/5.0 for the verification process. The trust scoring system's effectiveness will be evaluated through correlation analysis between trust scores and user behavior, ensuring scores accurately predict trustworthy interactions.

Network decentralization and resilience metrics will demonstrate the platform's ability to operate without central points of failure while maintaining performance standards. We will monitor the distribution of relay nodes across geographic regions and operators, targeting at least 100 community-operated nodes within the first year to ensure no single entity controls more than 10% of network capacity. Message delivery success rates must exceed 99.5% even during peak usage periods, with average message latency remaining under 500 milliseconds for relay-routed connections. Blockchain integration effectiveness will be measured through successful hash anchoring rates above 99% and zero-knowledge proof generation times under 2 seconds for provable messaging features.

Business impact and ecosystem development metrics will track the platform's progress toward enterprise adoption and financial integration opportunities. We will measure the number of enterprise pilot programs initiated, targeting partnerships with at least 5 financial institutions for fraud-proof customer communications within 18 months of launch. Developer ecosystem growth will be tracked through third-party bot and application integrations, with a goal of 50 verified bots and 10 enterprise integrations by the end of year one. Revenue metrics will focus on premium feature adoption rates and enterprise licensing, targeting 15% of users upgrading to premium tiers and generating $1M in annual recurring revenue by year two while maintaining the free tier's core functionality.

Community economy and Network State metrics measure ECHO's evolution from product to community-owned organization. VIP conversion rate tracks premium subscriptions, targeting 5% of users by year two and 10% by year three. Treasury AUM (assets under management) should reach $1M+ by year two and $10M+ by year three, demonstrating sustainable revenue generation. Annual ECHO burned and BTC reserve accumulation are governance-set targets that demonstrate economic sustainability and hard-asset backing. Governance participation rate measures active community engagement, targeting 15% of token holders voting by year two and 25% by year three. Real-world assets held tracks Network State progression, with first acquisition targeted for year three once treasury sustainability is established.

## VIP Subscription Pricing Strategy

### Recommended Approach: Single Tier ($9.99/month)

**Rationale:** A single VIP tier at $9.99/month is recommended over a two-tier model for the following strategic reasons:

**Simplicity and Focus:**

* Reduces decision paralysis for users (no "which tier?" question)
* Clear value proposition: free core features vs. premium VIP package
* Easier to communicate and market ("Try ECHO VIP" vs. explaining tier differences)
* Simpler product management and roadmap planning

**Revenue Optimization:**

* Most users willing to pay will accept $9.99 for comprehensive premium features
* Two-tier model risks anchoring users at lower $4.99 price point
* Single higher tier generates 2x revenue per subscriber vs. split model
* At 5% conversion (Year 2 target), single tier yields higher ARR than mixed tiers

**Community Alignment:**

* ECHO's value proposition is "all users are owners" — premium should feel premium
* $9.99 positions VIP as aspirational status symbol, not basic upgrade
* Token rewards can offset cost for engaged users (earn while you chat)
* Aligns with Telegram Premium ($4.99) and Discord Nitro ($9.99) positioning

**Conversion Scenarios:**

| Model | Conversion Rate | Avg Price | Year 2 Revenue (1M users) |
| --- | --- | --- | --- |
| Single Tier $9.99 | 5% | $9.99 | $5.99M ARR |
| Two Tier (60/40 split) | 5% | $7.40 | $4.44M ARR |
| Single Tier $4.99 | 8% | $4.99 | $4.79M ARR |

Single tier at $9.99 maximizes revenue even with lower conversion than a $4.99 option, and the price point is defensible given the unique value (E2E encryption + blockchain + token rewards + community ownership).

### VIP Tier Feature Set ($9.99/month)

**Core Principle:** Free tier provides complete messaging functionality. VIP adds convenience, capacity, customization, and status — never security or core messaging.

**Included in VIP:**

**Capacity Upgrades:**

* Create and manage groups up to 100K members (vs. 10K free tier limit)
* Voice/video calls up to 50 participants (vs. 10 free)
* 20GB cloud storage for media/files (vs. 2GB free)
* Message history unlimited (vs. 6 months free)
* 20 custom chat folders with unlimited chats per folder (vs. 5 folders free)

**Priority Performance:**

* Priority relay routing for faster message delivery
* Priority customer support (24-hour response vs. community forum)
* Early access to beta features and Phase 3-4 updates
* Increased daily reward cap: 150 ECHO/day (vs. 100 ECHO/day free)

**Customization & Expression:**

* Custom app themes and color schemes (10+ premium themes)
* Animated avatar borders and profile effects
* Custom emoji reactions (upload your own)
* VIP badge on profile (visible to other users)
* Profile customization: extended bio (500 chars vs. 150), custom fonts
* Message formatting: markdown support, code blocks, rich text

**Advanced Features:**

* Schedule messages up to 1 year in advance (vs. 1 week free)
* Disappearing message options: 5 seconds to 1 year (vs. 24 hours max free)
* Translate messages in 100+ languages (powered by on-device ML)
* Advanced bot management: run up to 10 personal bots (vs. 2 free)
* Broadcast channels: create channels with unlimited subscribers (vs. 1K subscriber limit free)

**Exclusive Governance:**

* VIP status grants +10% governance weight multiplier (on top of trust tier)
* Access to VIP-only governance proposals (e.g., feature prioritization votes)
* Monthly VIP community calls with founders
* Early voting access on major protocol decisions

**Status & Social:**

* VIP badge displayed on profile and in chats
* Access to VIP-only broadcast channel (founder updates, alpha features)
* Exclusive VIP sticker packs and emoji sets
* Profile appears higher in search results and discovery

### Features NOT Included (Always Free)

**Never Gate Security or Core Messaging:**

* End-to-end encryption (always free, always default)
* Self-sovereign identity (DIDs on Cardano)
* Blockchain-anchored message integrity (Merkle proofs)
* Trust scoring and verification badges (based on trust tier, not payment)
* Basic messaging: 1:1 chats, voice/video calls up to 10 people
* Groups up to 10K members (covers 99% of use cases)
* File sharing up to 2GB per file
* Disappearing messages (24-hour window)
* Hidden folders with biometric protection
* Staking and token rewards (all users earn ECHO)
* Governance voting rights (based on staked ECHO × trust tier, not VIP status)

**Why This Matters:** ECHO's mission is community ownership and privacy for all. VIP must enhance convenience and status without creating a "pay-to-play" system for security or core functionality. The free tier must be genuinely useful for 90%+ of users.

### Alternative: Two-Tier Model (If Needed)

If data shows significant demand for a lower-priced tier, consider this fallback:

**Standard Tier ($4.99/month):**

* Groups up to 50K members
* 10GB cloud storage
* Voice/video calls up to 25 participants
* Basic custom themes (5 themes)
* Standard badge
* Daily reward cap: 120 ECHO/day
* Early access to beta features

**VIP Tier ($9.99/month):**

* All Standard features plus:
* Groups up to 100K members
* 20GB cloud storage
* Voice/video calls up to 50 participants
* Full custom themes (10+ themes) + animated effects
* VIP badge with animation
* Daily reward cap: 150 ECHO/day
* Priority support
* +10% governance weight
* VIP-only calls with founders
* Advanced bot management (10 bots)

**Launch Strategy:** Start with single tier ($9.99), monitor feedback, add lower tier only if conversion < 3% after 6 months. Better to start premium and add down-market tier than start cheap and raise prices.

### Success Metrics for VIP Subscriptions

| Metric | Year 1 | Year 2 | Year 3 |
| --- | --- | --- | --- |
| VIP Conversion Rate | — | 5% | 10% |
| VIP MRR | — | $499K | $4.99M |
| VIP ARR | — | $5.99M | $59.9M |
| VIP Churn Rate (monthly) | — | <5% | <3% |
| VIP LTV (Lifetime Value) | — | $200 | $330 |
| VIP CAC (Customer Acquisition Cost) | — | <$50 | <$30 |
| Free-to-VIP Time (median days) | — | 90 days | 60 days |
| VIP Renewal Rate (annual) | — | 70% | 80% |

**Key Principle:** VIP revenue flows 100% to community treasury. Track these metrics transparently on public dashboard to demonstrate sustainable economics for Phase 5-6 Network State funding.

## Go To Market Strategy

### \1. Open Source Strategy

**Recommendation: Delayed Open Source at Phase 3 Launch**

**Rationale for Delaying (Not Immediate):**

**Competitive Protection (Phases 1-2):**

* Prevents large tech companies (Meta, Telegram) from cloning ECHO before launch
* Preserves first-mover advantage in blockchain-anchored messaging space
* Protects unique tokenomics model and trust-tier governance implementation
* Allows time to establish brand identity and user base before copycats emerge

**Product Quality (Phases 1-2):**

* Early code quality may not represent final vision (technical debt, iteration)
* Security audits not yet complete (don't expose vulnerabilities)
* Metagraph validation logic still evolving (reward caps, anti-gaming rules)
* Better to open source polished, audited code than messy v0.1

**Business Development (Phases 1-2):**

* Easier to negotiate enterprise pilots without open code review
* Financial institutions prefer closed-source during security evaluation period
* Constellation ecosystem partnerships easier with proprietary code
* Token presale/launch benefits from scarcity perception

**Why Open Source at Phase 3 (Launch):**

**Trust and Transparency:**

* ECHO's value proposition is "no company owns your account" — open source proves it
* Community can verify: no backdoors, no hidden data collection, E2E encryption is real
* Blockchain transactions are already public; code should match
* Network State vision requires community ownership of everything, including code

**Developer Ecosystem:**

* Open source enables third-party bot developers to build integrations
* Community can contribute features, translations, platform ports
* Security researchers can audit and report vulnerabilities (bug bounty program)
* Android developers can fork and build Android app from iOS codebase

**Decentralization:**

* Community can fork if governance fails or founders abandon project
* Prevents vendor lock-in (users can self-host relay nodes)
* Enables community-run L1 validators (Phase 4) with full code transparency
* Aligns with "all users are owners" — you can't own what you can't see

**Network Effects:**

* Open source projects attract passionate contributors (Signal, Matrix, Mastodon)
* Creates legitimacy in crypto/privacy communities (closed source = red flag)
* Generates free marketing from developer advocacy
* Positions ECHO as public infrastructure, not just another app

**Open Source Timeline:**

| Phase | Code Status | Rationale |
| --- | --- | --- |
| Phase 1 (Prototype) | Closed | Early iteration, unaudited, no competitive advantage |
| Phase 2 (Core Build) | Closed | Security audits in progress, enterprise pilots require NDA |
| Phase 3 (Launch) | **Open Source** | App Store launch + mainnet = code is public anyway, maximize trust |
| Phase 4+ | Open + Community PRs | Accept community contributions, community forks encouraged |

**License:** MIT or Apache 2.0 (permissive) — allows commercial use, forks, and modifications. This aligns with "Network State" vision where community can build on ECHO infrastructure.

**What to Open Source:**

* iOS app (Swift/SwiftUI)
* Go backend relay services
* Scala metagraph L1 validation logic
* Documentation, API specs, deployment guides

**What Stays Private:**

* Founder private keys and treasury multi-sig setup
* Production infrastructure credentials (Hetzner API tokens, Vault secrets, CI/CD tokens)
* Security vulnerability reports (until patched)
* Financial institution partnership agreements

### \2. Beta User Targets Before Launch

**Recommended Beta Progression:**

**Phase 2 Alpha (Closed Beta): 100-500 Users**

* **Target:** 100 users minimum, 500 users maximum
* **Duration:** 2-3 months during Core Build phase
* **Recruitment:** Invite-only from crypto/privacy communities, Constellation ecosystem, personal networks
* **Purpose:** Stress test relay infrastructure, debug E2E encryption edge cases, validate token reward mechanics
* **Success Criteria:** 60+ daily messages per user, <1% crash rate, 99%+ message delivery, zero security incidents

**Phase 3 Public Beta (Testnet): 1,000-10,000 Users**

* **Target:** 1,000 users minimum, 10,000 users stretch goal
* **Duration:** 1-2 months before App Store submission
* **Recruitment:** TestFlight (iOS), public announcement on Twitter/Reddit/Product Hunt "Beta"
* **Purpose:** Validate app performance at scale, test metagraph mainnet under load, gather UX feedback, build waitlist
* **Success Criteria:** 50+ daily messages per user, 30-day retention &gt;50%, NPS score &gt;40, App Store review-ready quality

**Phase 3 Soft Launch (Mainnet): 10,000-100,000 Users**

* **Target:** 10,000 users first month, 100,000 users by end of Phase 3
* **Duration:** 2-3 months post-App Store approval
* **Recruitment:** App Store listing (limited marketing), community referrals (50 ECHO per referral), crypto influencer partnerships
* **Purpose:** Organic growth validation, onboard early community governance participants, seed PacaSwap liquidity pools
* **Success Criteria:** 30-day retention &gt;60%, VIP conversion &gt;2%, 99.9% uptime, treasury &gt;$100K ARR

**Why This Progression:**

**100 users (Phase 2 Alpha):**

* Small enough to maintain personal relationships (feedback loops, bug reports)
* Large enough to test group chats, relay load balancing, reward distribution
* Matches Signal's early beta size (100-500 users for months before public launch)

**1,000-10,000 users (Phase 3 Public Beta):**

* Catches edge cases that 100 users won't (network effects, spam, abuse)
* Builds waitlist and FOMO ("Join the beta before launch")
* Tests metagraph at realistic transaction volumes (1K users = \~50K messages/day)
* Creates initial liquidity in PacaSwap pools (users buy ECHO to stake)

**10,000-100,000 users (Phase 3 Soft Launch):**

* Validates product-market fit before scaling marketing spend
* Builds treasury revenue to fund Phase 4-5 (VIP subscriptions, org pilots)
* Establishes trust tier distribution (30% Tier 3+) before governance votes matter
* Proves sustainability before Network State claims

**Red Flags to Pause Launch:**

* Alpha retention <40% (product not sticky enough)
* Beta NPS score <30 (users actively unhappy)
* Security audit finds critical vulnerabilities (delay until patched)
* Metagraph transaction finality &gt;30s (infrastructure not ready)

### \3. Enterprise Sales Timing

**Recommendation: Pilot Phase 4, Scale Phase 5**

**Phase 4 (6-18 months post-launch): Enterprise Pilots (5-25 Organizations)**

**When to Start:**

* **Prerequisite:** 100K+ MAU, 99.9% uptime for 3+ consecutive months, zero major security incidents
* **Timing:** Month 9-12 post-launch (after proving consumer product stability)
* **Why Wait:** Enterprises won't adopt unstable product; need proof of reliability first

**Pilot Program Structure:**

* **Target:** 5 pilot customers in Phase 4 (Year 1), expand to 25 in early Phase 5 (Year 2)
* **Industries:** Financial institutions (banks, credit unions), healthcare (HIPAA compliance), legal (attorney-client privilege)
* **Pricing:** $10-20/seat/month pilot pricing (50% discount), commit to 1-year minimum
* **Pilot Duration:** 3-6 months with defined success metrics (fraud reduction, customer satisfaction)

**Pilot Customer Profile:**

* **Problem:** SMS/email fraud costing $100K+ annually, customer complaints about phishing
* **Size:** 50-500 customer service reps (small enough to pilot, large enough to matter)
* **Tech Savvy:** Early adopters willing to try blockchain solutions (not legacy holdouts)
* **Budget Authority:** Can approve $50K-250K annual contract without board approval

**Pilot Value Proposition:**

**For Banks/Financial Institutions:**

* **Fraud Reduction:** Blockchain-anchored messages prove authenticity (customer can verify bank didn't spoof)
* **Compliance:** Digital Evidence Smart Checkmark creates court-admissible audit trail of all customer communications
* **Cost Savings:** Reduce $5B annual SMS phishing losses; ECHO costs < $100K/year for mid-size bank
* **Customer Trust:** Verification badges + trust scoring reduce customer service call volume (fewer "was this you?" calls)

**Pilot Success Metrics:**

* **Fraud Reduction:** 50%+ reduction in successful phishing attacks vs. SMS baseline
* **Customer Satisfaction:** NPS increase of 10+ points for customers using ECHO vs. SMS
* **Adoption:** 60%+ of customer service reps actively using ECHO within 90 days
* **ROI:** Positive ROI within 12 months (savings &gt; cost)

**Phase 5 (Year 2-3): Enterprise Scale (25-100 Organizations)**

**When to Scale:**

* **Prerequisite:** 500K+ MAU, 5+ successful pilot case studies, Organization tier launched, treasury sustaining operations
* **Timing:** Month 18-24 post-launch
* **Why Scale Now:** Product proven, case studies available, revenue funds sales team

**Enterprise GTM Motion:**

**Inbound:**

* Case study content marketing (bank pilot results: "50% fraud reduction in 90 days")
* Conference presence (Money20/20, Sibos, SXSW) with pilot customer speakers
* Gartner/Forrester analyst relations (get in "Cool Vendors" report)
* Digital Evidence API as wedge (start with compliance, expand to messaging)

**Outbound:**

* Hire VP Enterprise Sales + 2-3 AEs (funded by treasury, governance-approved budget)
* Target accounts: Top 100 US banks, Top 50 healthcare systems, Fortune 1000 legal departments
* Sales cycle: 3-6 months (RFP, security review, pilot, procurement)
* Deal size: $100K-500K ARR (500-5000 seats at $10-50/seat/month)

**Enterprise Tier Pricing (Phase 5):**

| Plan | Price/Seat/Month | Min Seats | Features |
| --- | --- | --- | --- |
| Organization | $10-20 | 50 | SSO, admin controls, branded channels, SLAs, basic Digital Evidence |
| Enterprise | $30-50 | 500 | All Org features + priority support, custom integrations, dedicated success manager, advanced Digital Evidence (API access, custom retention policies) |

**Enterprise Success Metrics:**

| Metric | Phase 4 (Year 1) | Phase 5 (Year 2) | Phase 5+ (Year 3) |
| --- | --- | --- | --- |
| Enterprise Pilots/Customers | 5 | 25 | 100 |
| Enterprise ARR | $250K | $2.5M | $10M |
| Avg Deal Size | $50K | $100K | $100K |
| Sales Cycle (days) | 180 | 120 | 90 |
| Win Rate (Pilot → Paid) | 80% | 70% | 70% |
| Customer Churn (annual) | <10% | <10% | <5% |
| NPS (Enterprise) | 50+ | 60+ | 70+ |

**Why This Timing Works:**

**Too Early (Phase 1-2):** Product not stable, no case studies, burns credibility\
**Just Right (Phase 4):** Consumer traction proves product, pilots low-risk, builds case studies\
**Scale (Phase 5):** Proven ROI, sales team funded by treasury, revenue accelerates Network State funding

**Go-To-Market Summary:**

| Timing | Action | Target | Purpose |
| --- | --- | --- | --- |
| Phase 2 | Alpha (closed beta) | 100-500 users | Stress test, debug, validate core mechanics |
| Phase 3 | Public Beta + Soft Launch | 1K-100K users | Build waitlist, gather feedback, seed liquidity |
| Phase 3 | **Open Source Code** | Developer community | Trust, transparency, ecosystem growth |
| Phase 4 | Enterprise Pilots | 5-25 orgs | Prove ROI, build case studies, learn enterprise needs |
| Phase 5 | Enterprise Scale | 25-100 orgs | Revenue acceleration, treasury funding, sales team |
| Phase 6 | Network State | 1M+ users, self-sustaining treasury | Real-world asset acquisition, digital jurisdiction |

**Key Principle:** ECHO is a community-owned platform. Go-to-market must balance growth (users, revenue) with decentralization (open source, governance). Every decision optimizes for long-term community ownership, not short-term extraction.

## Technical Requirements

## Technical Requirements

### Core Technology Stack

**Identity & Authentication**

* Cardano blockchain for self-sovereign DIDs (Veridian/Atala PRISM)
* KERI standards for DID interoperability
* Zero-knowledge proofs for privacy-preserving verification
* No email/phone required for registration

**Messaging Infrastructure**

* Go backend with WebSocket relay for message routing
* X25519 key agreement + ChaCha20-Poly1305 for E2E encryption
* APNs for iOS push notifications
* Encrypted store-and-forward for offline message queuing
* TLS 1.3 transport encryption

**Blockchain Integration**

* Constellation Hypergraph (Data L1) for message integrity anchoring via Merkle roots
* Constellation Metagraph (Currency L1) for ECHO token operations
* Tessellation v3 transaction primitives: TokenLock, StakeDelegation, AtomicAction, WithdrawLock, FeeTransaction
* Scala/JVM for L1 validation logic using Euclid SDK
* Cardano smart contracts for trust tier commitments and verification badges

**Storage & Media**

* IPFS/Storj for distributed encrypted file storage
* PostgreSQL and Redis as performance caches only (not source of truth)
* All persistent state on-chain (metagraph for app data, Cardano for identity)

**Frontend**

* Swift native iOS app with SwiftUI
* Secure Enclave integration for private key storage and biometrics
* Cross-platform design system for future Android support

### Infrastructure Requirements

**Node Operations**

* Minimum 3 L0 hybrid nodes with 250K DAG staking each (750K DAG total)
* 3-5 Currency L1 validators for token transactions
* 3-5 Data L1 validators for message integrity commitments
* Scala/JVM development capability for custom L1 validation logic
* Ubuntu servers with 8+ cores, 32GB RAM minimum

**Network Performance**

* Message delivery rate: 99.9%+ target
* Message latency: <500ms for relay routing
* Blockchain finality: <15s for on-chain commitments
* Uptime: 99.9%+ target with redundant relay nodes

**Security Requirements**

* End-to-end encryption for all message content
* Device-local private key storage (Secure Enclave)
* Zero PII on blockchain (enforced by T0-T7 data classification)
* Annual security audits (E2E encryption, Secure Enclave, metagraph validation logic)
* Penetration testing before major releases

### Scalability Targets

**Phase 1-2 (MVP)**

* 100-1,000 users
* Single region deployment
* Basic relay infrastructure

**Phase 3 (Launch)**

* 100,000 users
* Multi-region relay nodes
* Sealed sender metadata protection

**Phase 4 (Scale)**

* 1M+ users
* Federated relay operators
* Community-run L1 validators
* Metagraph sharding if needed

**Phase 5-6 (Maturity)**

* 5M+ users
* Global relay distribution
* Full decentralization of validation
* Enterprise-grade SLAs

### Compliance & Legal

* GDPR compliance for EU users (right to be forgotten via key deletion)
* Data retention policies aligned with legal requirements
* Constellation Digital Evidence API integration for enterprise compliance
* Court-admissible audit trails for Organization tier customers
* DAO LLC legal structure for Network State phase (Wyoming or Marshall Islands)


# Feature Requirements

## Decentralized Identity and Authentication

### Decentralized Identity and Authentication

This feature provides users with self-sovereign identity management that eliminates dependence on centralized authentication servers while enabling progressive verification and trust building. Users create blockchain-anchored identities through Decentralized Identifiers (DIDs) that they fully control, removing the need for traditional email or phone number registration while supporting verifiable credentials for enhanced trust.

Users begin by connecting a compatible wallet or generating a new DID through the app's On-Boarding flow. The system creates their unique identifier on the Cardano blockchain using Atala PRISM or Veridian infrastructure, establishing an immutable identity anchor. Users can then progressively add verifiable credentials such as proof of humanity, KYC-lite verification, or professional credentials through zero-knowledge proof mechanisms that preserve privacy while establishing authenticity. The trust scoring algorithm evaluates on-chain behavior, interaction history, and verification levels to assign dynamic trust scores from 0-100 that unlock additional features and privileges within the app.

The feature requires integration with Cardano's SSI infrastructure, specifically Atala PRISM for credential management and KERI standards for DID interoperability. Users must have access to a compatible wallet or allow the app to generate and manage their DID securely. The system depends on Cardano blockchain availability for identity verification and smart contract execution for trust score calculations.

This approach ensures users maintain complete control over their identity data while enabling the trust mechanisms necessary for secure communications and fraud prevention. The progressive verification model allows users to start with basic functionality and unlock advanced features as they build reputation and complete additional verification steps.

## Blockchain-Anchored Messaging with Provable Integrity

### Blockchain-Anchored Messaging with Provable Integrity

This feature provides end-to-end encrypted messaging with cryptographic proof of message authenticity and conversation integrity, eliminating the possibility of message tampering or impersonation attacks. Users can engage in private conversations while maintaining the ability to prove that specific communications occurred when needed for legal, business, or security purposes.

Users send messages through the Echo relay infrastructure using WebSocket connections. Each message is encrypted on the sender's device using X25519 key agreement and ChaCha20-Poly1305 before transmission, ensuring only intended recipients can decrypt the content. The relay servers transport opaque ciphertext and cannot read, modify, or forge message content. Simultaneously, message hashes are anchored to the Constellation metagraph through Merkle root commitments, creating immutable timestamps and integrity proofs without exposing message content. Users can enable "provable mode" for sensitive conversations, which generates cryptographic receipts that can later verify message authenticity, sender identity, and delivery confirmation through zero-knowledge proofs.

The messaging layer handles metadata protection through progressive phases: TLS 1.3 transport encryption in Phase 1-2, sealed sender implementation in Phase 3, and federated relay nodes in Phase 4. The system requires active blockchain connectivity for hash anchoring and proof generation, though offline messaging is supported through encrypted store-and-forward with delayed blockchain confirmation once connectivity is restored.

This architecture enables users to maintain complete privacy for everyday communications while providing the cryptographic guarantees necessary for high-stakes conversations involving financial transactions, legal agreements, or sensitive business communications. The provable integrity feature addresses the growing need for verifiable digital communications in an era of increasing deepfake and impersonation threats.

## Dynamic Trust Network and Social Verification

### Dynamic Trust Network and Social Verification

This feature creates a decentralized reputation system that enables users to build trust relationships and verify authenticity without relying on centralized authorities or exposing personal information. The system combines blockchain-anchored verification badges, progressive trust circles, and community-driven reputation scoring to create a self-regulating network that reduces fraud and spam while preserving user privacy.

Users establish trust relationships through multiple pathways that mirror real-world social verification patterns. They can earn verification badges by completing on-chain credential verification such as proof of humanity, professional credentials, or KYC-lite processes that use zero-knowledge proofs to confirm authenticity without revealing personal details. The trusted circles feature allows users to categorize their contacts into different trust levels based on interaction history and mutual connections, with "Inner Circle" members receiving unlimited message history and priority routing, while "Acquaintances" operate under standard limitations. The dynamic trust scoring algorithm continuously evaluates user behavior including verification completions, interaction volume, report history, and on-chain activity to assign scores from 0-100 that unlock progressive features and privileges.

The trust network operates through smart contracts deployed on Cardano that maintain reputation scores and verification states while preserving user privacy through cryptographic commitments. Users can report suspicious behavior or spam through an on-chain evidence system that creates immutable records for community review without exposing reporter identity. The system integrates with the DID infrastructure to ensure trust scores and verification badges are portable across applications and resistant to Sybil attacks through blockchain-anchored identity requirements.

This approach creates natural spam resistance and fraud prevention through economic incentives and social proof mechanisms. Users with higher trust scores gain access to premium features like enhanced group capabilities, priority message routing, and potential integration with financial services, while maintaining the ability to interact pseudonymously with appropriate privacy protections. The decentralized nature ensures no single entity can manipulate trust scores or verification status, creating a transparent and auditable reputation system.

## Voice and Video Calls with Screen Sharing

# Voice and Video Calls with Screen Sharing

This feature provides high-quality voice and video calling capabilities with advanced screen sharing functionality, enabling users to conduct business meetings, technical support sessions, and collaborative work directly within the secure messaging environment. The system maintains end-to-end encryption for all audio, video, and screen content while leveraging the platform's trust infrastructure to verify participant identities and prevent unauthorized access to sensitive shared content.

Users initiate voice or video calls through the standard chat interface, with the system establishing encrypted relay connections for optimal quality and privacy. The calling infrastructure supports up to 50 participants in group calls, with automatic quality adjustment based on network conditions and device capabilities. Screen sharing allows users to broadcast their entire screen, specific application windows, or selected desktop areas to call participants, with granular permission controls that prevent unauthorized recording or screenshot capture. The system includes advanced features like virtual backgrounds, noise cancellation, and real-time transcription for accessibility, all processed locally on user devices to maintain privacy.

The technical implementation uses WebRTC protocols enhanced with the platform's Noise Protocol encryption to ensure all call data remains private and tamper-proof. Screen sharing content is encrypted before transmission and includes blockchain-anchored integrity proofs that can verify shared content authenticity for business or legal purposes. The system integrates with the trust scoring infrastructure to provide verified caller identification, reducing the risk of voice phishing attacks and impersonation during important business calls. Call metadata including participant lists, duration, and quality metrics are recorded on the blockchain for audit purposes while maintaining participant privacy through zero-knowledge proofs.

Integration with the existing DID system enables seamless identity verification during calls, with visual indicators showing each participant's verification status and trust score. The feature supports scheduled calls with calendar integration, automatic recording with participant consent, and post-call summaries that include action items and shared files. Enterprise users can configure additional security measures including mandatory verification requirements for call participants, automatic call recording for compliance purposes, and integration with corporate communication policies that govern screen sharing permissions and content restrictions.

## Large File Sharing and Cloud Storage Integration

### Large File Sharing and Cloud Storage Integration

This feature enables users to share files up to 2GB in size while maintaining end-to-end encryption and decentralized storage principles, addressing the need for secure document exchange in both personal and professional communications. The system combines IPFS distributed storage with blockchain anchoring to ensure file integrity and availability while providing seamless integration with popular cloud storage services for user convenience.

Users can share files through drag-and-drop functionality or direct upload from their device, with automatic encryption occurring before the file leaves their device. Large files are automatically chunked and distributed across the IPFS network, with each chunk encrypted using unique keys derived from the conversation's encryption context. The system generates blockchain-anchored hashes for all shared files, creating immutable proof of file integrity and enabling recipients to verify that files haven't been tampered with during transmission. File sharing permissions are controlled through the same trust mechanisms used for messaging, with higher trust scores unlocking larger file size limits and priority storage allocation.

The feature integrates with popular cloud storage services including Google Drive, Dropbox, and OneDrive through secure API connections that maintain end-to-end encryption while enabling users to share files directly from their existing cloud storage. Shared files include automatic virus scanning through decentralized security oracles, with malicious content blocked before distribution to recipients. The system supports collaborative document editing through integration with decentralized office suites, enabling real-time collaboration on shared documents while maintaining the platform's privacy and security standards.

File storage utilizes a hybrid approach where frequently accessed files are cached on IPFS nodes for quick retrieval, while long-term storage is handled through Filecoin's incentivized storage network to ensure permanent availability. Users can configure automatic file expiration for sensitive documents, with cryptographic deletion ensuring files become permanently inaccessible after specified timeframes. The system includes comprehensive file management features including search functionality, version control for collaborative documents, and automatic backup of important files to user-controlled storage locations.

## Message Reactions, Polls, and Interactive Elements

### Message Reactions, Polls, and Interactive Elements

This feature provides users with rich interactive communication tools including emoji reactions, polls, surveys, and interactive buttons that enhance group engagement while maintaining the platform's security and privacy standards. The system enables expressive communication and decision-making tools that rival traditional social media platforms while preserving end-to-end encryption and decentralized architecture.

Users can react to messages using a comprehensive emoji library that includes standard Unicode emojis, custom reactions, and blockchain-verified NFT emojis that users can collect and trade. The reaction system supports multiple reactions per user per message, with real-time synchronization across all participants' devices through the encrypted relay network. Poll creation allows users to pose questions with multiple choice answers, with voting results encrypted and tallied through zero-knowledge proofs that preserve voter privacy while ensuring result integrity. Advanced poll features include anonymous voting, time-limited polls, and weighted voting based on trust scores for community governance decisions.

Interactive elements extend beyond basic reactions to include action buttons that can trigger smart contract functions, payment requests, or external application integrations. Users can create interactive messages that include buttons for quick responses, calendar scheduling, or e-commerce transactions, with all interactions maintaining the platform's security standards. The system supports rich media reactions including voice note responses, photo reactions, and short video clips that are automatically compressed and encrypted for efficient transmission.

The feature integrates with the ECHO token system to enable reaction-based rewards, where popular content creators can earn tokens based on engagement metrics while maintaining user privacy through anonymous interaction tracking. Poll results and reaction data are anchored to the blockchain for transparency and audit purposes, enabling community-driven decision making for group governance and platform development priorities. The system includes comprehensive analytics for group administrators to understand engagement patterns and optimize community management strategies while respecting individual user privacy through aggregated, anonymized reporting.

## Advanced Message Search and Archive System

### Advanced Message Search and Archive System

This feature provides users with powerful search capabilities across their entire message history while maintaining end-to-end encryption and privacy protection through client-side indexing and zero-knowledge search techniques. Users can quickly locate specific conversations, files, or information across years of communication history without compromising the security principles that protect their private communications.

The search system operates through local indexing where message content is processed and indexed on each user's device using privacy-preserving techniques that create searchable metadata without exposing message content to external systems. Users can search by keywords, date ranges, file types, sender identity, or conversation context, with results ranked by relevance and recency. Advanced search filters enable users to locate specific types of content such as shared files, links, images, or messages containing payment information, with all search operations performed locally to maintain privacy.

Archive functionality allows users to organize their message history into custom categories and folders while maintaining the ability to search across archived content. The system supports automatic archiving based on user-defined rules such as conversation inactivity, trust score thresholds, or content type classifications. Archived messages remain fully encrypted and accessible through the search interface, with options for secure backup to user-controlled storage locations including hardware devices or decentralized storage networks.

The feature includes advanced search capabilities such as semantic search that can locate messages based on meaning rather than exact keyword matches, utilizing locally-processed natural language understanding that never exposes message content to external AI services. Cross-device search synchronization occurs through encrypted index sharing that allows users to search their complete message history from any device while maintaining end-to-end encryption. The system supports search result sharing where users can create secure links to specific messages or conversations that can be shared with verified contacts while maintaining access controls and expiration settings.

## Hidden Folders with Biometric Protection

### Hidden Folders with Biometric Protection

This feature provides users with secure, biometrically-protected folders for sensitive one-on-one conversations that require additional privacy layers beyond standard end-to-end encryption. Hidden folders remain completely invisible in the main chat interface and can only be accessed through successful biometric authentication, creating a secure vault for confidential communications that protects against unauthorized access even if the device is compromised.

Users create hidden folders by selecting specific one-on-one conversations and moving them to a protected space that requires Face ID, Touch ID, or other biometric verification methods supported by their device. The folder creation process generates additional encryption keys that are bound to the user's biometric template, ensuring that even if someone gains access to the device or the user's primary authentication credentials, they cannot access hidden conversations without the correct biometric signature. The system supports multiple hidden folders with different access requirements, allowing users to categorize sensitive conversations by security level or relationship type.

The hidden folder interface operates as a completely separate chat environment that maintains its own message history, notification settings, and backup protocols. Messages within hidden folders use enhanced encryption that combines the standard Noise Protocol implementation with biometric-derived key material, creating multi-layered security that protects against both network-level and device-level attacks. Users can configure custom notification behaviors for hidden conversations, including silent notifications that appear only when the folder is unlocked, or complete notification suppression that prevents any indication of incoming messages from appearing on the device.

The feature integrates with the device's secure enclave or trusted execution environment to ensure biometric templates and derived encryption keys never leave the hardware security module. Hidden folder metadata is encrypted and stored locally on the device rather than synchronized across multiple devices, maintaining the principle that sensitive conversations remain isolated to the specific device where they were created. Users can optionally enable secure backup of hidden folders through additional biometric verification combined with a recovery phrase that allows restoration on new devices.

This approach addresses the need for ultra-secure communication channels for sensitive personal, professional, or financial discussions that require protection beyond what standard messaging provides. The biometric binding ensures that even sophisticated attackers who compromise the device or intercept network traffic cannot access the most sensitive conversations without physical access to the authorized user.

## Silent and Scheduled Private Chats

### Silent and Scheduled Private Chats

This feature enables users to send messages that generate no notifications or visible indicators on the recipient's device, while also supporting scheduled message delivery for time-sensitive communications across different time zones or planned conversations. The system provides granular control over message visibility and timing while maintaining end-to-end encryption and blockchain anchoring for all communications.

Users can activate silent mode for individual conversations or specific messages, which suppresses all notification behaviors including push notifications, badge counts, typing indicators, and read receipts on the recipient's device. Silent messages appear in the conversation thread only when the recipient actively opens the chat, creating a non-intrusive communication channel for sensitive or low-priority messages. The scheduling functionality allows users to compose messages that are delivered at predetermined times, with messages encrypted and stored locally on the sender's device until the scheduled delivery time when they are transmitted through the encrypted relay infrastructure.

The silent messaging system operates through enhanced metadata handling where notification suppression flags are embedded in the encrypted message payload, ensuring that even relay nodes cannot determine which messages should generate notifications. Scheduled messages use time-locked encryption where the message content is encrypted with keys that are only released at the specified delivery time through smart contract automation on the Constellation network. Users can schedule messages up to 30 days in advance, with the system supporting recurring message patterns for regular communications like daily check-ins or weekly reports.

The feature integrates with the existing trust scoring system to prevent abuse, where users with low trust scores face limitations on silent messaging frequency to prevent spam or harassment. Scheduled messages maintain full blockchain anchoring and provable integrity features, with delivery timestamps cryptographically verified to ensure messages were sent at the intended time. The system supports cross-timezone scheduling with automatic conversion based on recipient location preferences while maintaining privacy through zero-knowledge proofs that confirm delivery timing without exposing user location data.

This functionality addresses the need for respectful communication patterns that don't interrupt recipients during sensitive times while enabling users to maintain consistent communication schedules across global time zones. The silent messaging capability is particularly valuable for professional communications, emergency contact protocols, and personal relationships where immediate notification may be inappropriate or disruptive.

## Disappearing Messages with Cryptographic Verification

### Disappearing Messages with Cryptographic Verification

This feature provides users with the ability to send messages that automatically delete from all devices after predetermined time periods while maintaining cryptographic proof that the messages existed and were delivered, addressing privacy needs without compromising the platform's provable integrity capabilities. The system ensures that sensitive communications can be ephemeral while preserving audit trails for compliance and dispute resolution when necessary.

Users can enable disappearing messages for individual conversations or specific messages by selecting from preset time intervals ranging from 10 seconds to 7 days, with custom timing options available for premium users. When activated, messages display countdown timers that show remaining visibility time to all participants, creating transparency about message lifecycle. The deletion process occurs simultaneously across all devices through cryptographic coordination, ensuring that messages cannot persist on any participant's device beyond the specified timeframe. However, the system maintains blockchain-anchored hashes of deleted messages that can prove conversations occurred without revealing content, enabling users to demonstrate communication history for legal or business purposes.

The technical implementation uses client-side timers coordinated through the relay infrastructure. When expiration times are reached, the client app securely wipes the plaintext and ciphertext from local storage, and the relay server deletes any cached encrypted blobs. Messages are encrypted with time-sensitive keys that become invalid after the specified period, making recovery impossible even if encrypted data fragments remain on devices. The system supports different deletion policies for different trust levels, where verified users can set longer retention periods and access advanced features like selective message preservation for important communications. Screenshots and forwarding are technically prevented through device-level security measures, though users are notified when these protections may be bypassed.

The feature integrates with the existing trust scoring system to prevent abuse, where users with low trust scores face restrictions on very short disappearing timeframes to prevent harassment or evidence destruction. Blockchain anchoring continues to record message metadata including timestamps, participant identities, and delivery confirmations while the actual content becomes permanently inaccessible. The system maintains compliance with legal discovery requirements by preserving cryptographic evidence of communications while respecting user privacy through content deletion.

This approach addresses growing privacy concerns about permanent digital records while maintaining the platform's core value proposition of provable communications. The feature is particularly valuable for sensitive personal conversations, confidential business discussions, and situations where users want to communicate freely without creating permanent digital footprints that could be compromised or misused in the future.

## Public and Private Groups with Verified Status Display

### Public and Private Groups with Verified Status Display

This feature enables users to create and participate in both public and private group conversations while displaying transparent verification status for all participants, creating trust-based community spaces that reduce spam and impersonation while maintaining appropriate privacy controls. Groups leverage the platform's trust infrastructure to create self-moderating communities where verification levels determine participation privileges and administrative capabilities.

Users can create public groups that are discoverable through the platform's search functionality and allow anyone to join based on configurable trust score requirements, or private groups that require invitation links or direct invitations from existing members. Group creators establish verification requirements during setup, such as minimum trust scores, specific credential types, or manual approval processes that filter participants based on their blockchain-anchored identity verification. Each group displays a verification badge indicating the collective trust level of its members, with color-coded indicators showing the percentage of verified participants and the group's overall security rating based on member credentials and interaction history.

The group interface prominently displays each participant's verification status through visual indicators next to their usernames, including verification badges earned through credential completion, trust scores represented through progressive visual elements, and administrative roles that require enhanced verification levels. Group administrators with high trust scores can configure advanced moderation settings including automatic message filtering based on trust scores, temporary muting of unverified users during sensitive discussions, and evidence-based reporting systems that create blockchain-anchored records of policy violations. The system supports nested permission structures where different verification levels unlock specific capabilities such as file sharing, voice chat participation, or the ability to invite new members.

The feature integrates with the existing DID infrastructure to ensure verification status remains portable and tamper-resistant while supporting privacy-preserving group discovery that allows users to find relevant communities without exposing their personal interests. Group metadata including member counts, verification statistics, and activity levels are anchored to the blockchain to prevent manipulation while maintaining participant privacy through zero-knowledge proofs that confirm group membership without revealing individual identities to external observers.

This approach creates natural community curation where high-quality groups attract verified users while spam-prone groups become self-evident through low verification rates. The transparent trust display enables users to make informed decisions about group participation while providing group administrators with the tools necessary to maintain productive community spaces without relying on centralized moderation systems.

## Multiple Personas with Selective Visibility

### Multiple Personas with Selective Visibility

This feature enables users to create multiple distinct personas under their main profile, allowing them to compartmentalize their identity and interactions across different social circles while maintaining complete control over which contacts can see each persona. Users can present different aspects of their identity to different groups without compromising their privacy or creating separate accounts, addressing the need for contextual identity management in both personal and professional communications.

Users create additional personas through their main profile settings, with each persona having its own display name, avatar, bio, and verification status while sharing the underlying DID and trust score from the master identity. The system supports up to five personas per user, with categories like "Professional," "Personal," "Family," "Gaming," or custom labels that help users organize their different social contexts. Each persona can have distinct privacy settings, notification preferences, and feature access levels, allowing users to maintain professional boundaries while engaging in casual conversations through different identity presentations.

The selective visibility system operates through cryptographic access controls where users explicitly grant specific contacts permission to see particular personas. When initiating conversations or joining groups, users choose which persona to present, and only contacts who have been granted access to that persona can see the associated profile information and interaction history. The system maintains separate conversation threads for each persona, ensuring that messages sent as "Professional John" remain completely isolated from conversations conducted as "Gaming John," even when communicating with overlapping contact lists.

The feature integrates with the existing trust scoring system where the master identity's trust score applies to all personas, but individual personas can earn additional verification badges specific to their context, such as professional credentials for work personas or gaming achievements for entertainment personas. Contact management becomes persona-aware, allowing users to categorize their contacts based on which personas they know about, with automatic suggestions for appropriate persona selection based on conversation context and contact relationships. The blockchain anchoring system maintains provable integrity for all personas while using zero-knowledge proofs to ensure that contacts cannot discover the existence of personas they haven't been granted access to.

This approach addresses the growing need for contextual identity management in digital communications, where users want to maintain professional relationships without exposing personal interests, or engage in hobby communities without revealing work affiliations. The feature prevents the social awkwardness and privacy concerns that arise when all aspects of a user's digital identity are visible to all contacts, while maintaining the platform's core principles of verifiable identity and trustworthy communications through the shared underlying DID infrastructure.

## Broadcast Channels and Community Features

### Broadcast Channels and Community Features

This feature enables users to create one-to-many communication channels for broadcasting information to large audiences while maintaining the platform's decentralized architecture and privacy protections. Channels support various content types and engagement models, from simple announcement channels to interactive community spaces that foster discussion and collaboration around shared interests.

Channel creators can establish broadcast channels that support unlimited subscribers, with content distributed through the platform's encrypted relay network to ensure resilience and prevent censorship. Channels can be configured as public (discoverable through search), private (invitation-only), or semi-private (discoverable but requiring approval to join). Content types include text messages, images, videos, files, polls, and interactive elements, with all content encrypted and distributed through the same E2E encryption and relay infrastructure used for private messaging. Channel administrators can configure moderation settings, subscriber permissions, and content policies while maintaining transparency through blockchain-anchored governance records.

The system supports various channel types including news channels for media organizations, announcement channels for businesses and projects, educational channels for course content and tutorials, and community channels that enable subscriber interaction and discussion. Advanced features include scheduled posting, content categorization, subscriber segmentation for targeted messaging, and integration with external content management systems. Channel analytics provide creators with insights into subscriber engagement, content performance, and growth metrics while maintaining subscriber privacy through anonymized reporting.

Monetization options for channel creators include subscription fees paid in ECHO tokens, premium content tiers, sponsored content with transparent disclosure, and direct donations from subscribers. The system includes discovery mechanisms that help users find relevant channels based on their interests, trust network connections, and engagement history while preventing spam and low-quality content through community-driven curation. Channel content is archived and searchable, with subscribers able to access historical content and receive notifications for new posts based on their preferences and the channel's trust score.

## Enterprise Organization Profiles with Verified Status

### Enterprise Organization Profiles with Verified Status

This feature enables organizations including banks, corporations, government agencies, and non-profits to establish verified enterprise profiles that display authenticated organizational credentials and provide enhanced communication capabilities for official business interactions. Enterprise profiles receive distinctive verification checkmarks that differentiate legitimate organizations from impersonators while providing customers and stakeholders with trusted communication channels for official business.

Organizations begin the verification process by submitting comprehensive documentation including business registration certificates, regulatory licenses, executive authorization letters, and compliance certifications through a dedicated enterprise onboarding portal. The verification process involves multi-stage authentication where legal entities must provide proof of incorporation, regulatory standing with relevant authorities, and multi-signature authorization from C-level executives or board members. Financial institutions undergo additional scrutiny including FDIC registration verification, banking license validation, and compliance with anti-money laundering regulations. The system supports different verification tiers including Basic Enterprise (standard business registration), Regulated Entity (financial services, healthcare, legal), and Government Agency (federal, state, local authorities) with corresponding visual indicators and privilege levels.

Enterprise profiles display prominent verification badges that indicate the organization's verified status, regulatory compliance level, and industry classification. The interface shows organizational hierarchy with verified employee accounts linked to the main enterprise profile, enabling customers to distinguish between official representatives and potential impersonators. Organizations can configure branded communication channels with custom themes, official logos, and standardized message templates that maintain consistent corporate identity across all customer interactions. The system supports role-based access controls where different employee verification levels unlock specific communication privileges, from basic customer service to executive-level secure channels.

The feature integrates with existing regulatory databases and compliance systems to maintain real-time verification status, automatically flagging organizations that lose regulatory standing or face compliance violations. Enterprise profiles can establish verified communication policies that require cryptographic signatures for official announcements, financial disclosures, or legal notifications, creating immutable audit trails for regulatory compliance. The system supports integration with corporate identity management systems including Active Directory, SAML authentication, and enterprise single sign-on solutions to streamline employee verification and access management.

Organizations benefit from enhanced trust signals that reduce customer skepticism about official communications, while customers gain confidence in distinguishing legitimate business communications from phishing attempts and fraud. The verification system creates natural barriers against impersonation attacks while providing organizations with the tools necessary to maintain professional communication standards and regulatory compliance in a decentralized messaging environment. Enterprise profiles can leverage the platform's blockchain anchoring capabilities to create legally admissible records of customer communications, policy notifications, and compliance disclosures that satisfy regulatory examination requirements.

## Verified Financial Institution Integration

### Verified Financial Institution Integration

This feature transforms the messaging platform into a secure communication channel for financial institutions to conduct fraud prevention, customer service, and compliance activities with cryptographic proof and enhanced security compared to traditional SMS and email channels. Banks and credit unions can establish verified channels that leverage the platform's trust infrastructure to reduce phishing attacks and improve customer authentication while maintaining regulatory compliance.

Financial institutions begin integration by establishing institutional DIDs through the same Cardano-based identity system used by individual users, but with enhanced verification requirements including regulatory compliance documentation and multi-signature authorization from institution executives. Once verified, banks can create dedicated communication channels with their customers who have opted into institutional messaging. The system supports four primary interaction modes: automated fraud alerts that require cryptographic confirmation from customers using their DID-based authentication; dedicated customer service channels staffed by verified bank representatives with trust scores visible to customers; secure document exchange for sensitive financial communications that require immutable audit trails; and high-assurance, end-to-end encrypted video calls for face-to-face consultations, such as for wealth management or complex issue resolution. These video calls leverage the platform's existing WebRTC infrastructure, but with added identity verification overlays to confirm both the customer and the bank representative are who they claim to be, which should of happened during on boarding of a user or entity.

Customer interactions flow through a structured verification process where banks send transaction alerts or service requests through the platform's API integration, which generates cryptographically signed messages that customers can verify originated from their actual financial institution. Customers respond using biometric authentication combined with their DID signatures, creating immutable proof of authorization that prevents later disputes about transaction approvals. The trust scoring system prioritizes customers with higher verification levels for premium support channels, while maintaining privacy through zero-knowledge proofs that confirm customer identity without exposing personal financial information to the platform operators.

The feature requires integration with existing banking core systems through secure API endpoints that comply with PCI DSS and SOC 2 Type II standards. Banks must complete regulatory compliance reviews including FDIC communication guidelines and implement multi-factor authentication for their institutional accounts. The system depends on real-time blockchain connectivity for transaction verification and maintains encrypted audit logs that satisfy regulatory examination requirements while preserving customer privacy through cryptographic commitments.



**Enterprise Fraud Prevention Enhancements:**

The financial institution integration includes dedicated fraud analytics and prevention capabilities designed to replace SMS-based verification as the primary customer authentication channel. Real-Time Transaction Verification Alerts allow banks to send cryptographically signed transaction alerts through ECHO, where customers see the bank's verified identity badge, transaction details (amount, merchant, timestamp), and a one-tap "Confirm" or "Report Fraud" button. The customer's confirmation is signed with their DID, creating a court-admissible authorization record that cannot be forged or repudiated.

Enterprise clients receive access to a Fraud Analytics Dashboard showing: fraud attempt volume (phishing attempts blocked by verified channel versus SMS baseline), customer response times to fraud alerts, verification adoption rate (percentage of customers using ECHO versus SMS for alerts), and an ROI calculator demonstrating cost savings versus SMS fraud losses. This dashboard is critical for enterprise sales — CISOs need data to justify the investment to their boards.

Cross-Organization Fraud Intelligence (Phase 5+) enables participating institutions to share fraud pattern data using zero-knowledge proofs without revealing customer information. For example, institutions can query whether a specific DID has been flagged by multiple institutions within a time period, without revealing which institutions flagged it or what the specific fraud pattern was. This creates a decentralized fraud detection network that no centralized platform currently offers, directly leveraging ECHO's Midnight ZK integration for privacy-preserving inter-institutional data sharing.

This integration addresses the critical security gap in current banking communications where 70% of financial fraud originates from SMS phishing attacks. By providing cryptographically verifiable communication channels, banks can reduce fraud response times by up to 50% while creating immutable audit trails that satisfy regulatory compliance requirements and improve customer trust through transparent verification mechanisms.

## User Rewards Tracker on Profile

### User Rewards Tracker on Profile

This feature provides users with a summary view of their ECHO token activity and achievements within their profile interface, complementing the primary ECHO Wallet tab. While the Wallet tab (built on Stargazer SDK) serves as the primary interface for managing balances, staking, delegation, and transactions, the profile rewards tracker focuses on gamification elements, achievement milestones, and social status indicators that encourage continued platform engagement.

The profile dashboard displays a high-level summary of ECHO token activity with quick stats including total earned, current balance (linked to Wallet tab), trust tier multiplier, and active earning streak. Users can view achievement milestones such as "First 1000 Messages," "Trusted Verifier," or "Super Referrer" that unlock special badges and bonus multipliers, creating progression pathways that encourage long-term platform adoption. The interface showcases reputation metrics like trust score history, verification badges earned, and community contributions, creating social proof of user authenticity and network value.

The tracker integrates with the trust scoring system to display how verification levels and trust improvements directly impact earning potential, showing users their current trust tier multiplier (1x to 3x) and progress toward the next tier. Users can see personalized recommendations for increasing their rewards through activities like completing additional verification steps (Cardano credentials), participating in community governance votes, or referring high-quality users. The system includes social comparison features that allow users to see anonymized leaderboards of top contributors in their region or trust tier, fostering healthy competition while maintaining privacy through zero-knowledge ranking proofs.

The profile tracker serves as a "trophy case" and social signaling tool while the ECHO Wallet tab handles actual financial operations. Users can display achievement badges, earning streaks, and trust tier status on their profile for other users to see, creating reputation-based network effects. The tracker includes quick-action buttons that deep-link to relevant Wallet features: "View Wallet" (opens Wallet tab), "Stake ECHO" (opens staking interface), "Claim Rewards" (AtomicAction claim), "Invite Friends" (referral program).

Integration with the blockchain infrastructure ensures all displayed statistics are cryptographically verified and pulled from on-chain data (metagraph snapshots). The profile view is optimized for social sharing and status display, while detailed financial management happens in the dedicated Wallet tab. This separation ensures the profile remains focused on reputation and gamification while the Wallet provides professional asset management built on Stargazer SDK.

This rewards tracking system transforms the token economy from an abstract concept into tangible social status that demonstrates the value of platform participation. The gamification elements create positive feedback loops that encourage users to increase their engagement while building trust and verification levels that benefit the entire network ecosystem. For actual financial operations, users are directed to the ECHO Wallet tab with full Constellation ecosystem compatibility.

## Streamlined Onboarding with Verifiable Credentials and Passkeys

### Streamlined Onboarding with Verifiable Credentials and Passkeys

This feature streamlines the user enrollment and registration process by enabling new users to onboard instantly using industry-standard Verifiable Credentials, compliant with the OpenID Connect for Verifiable Credentials (OIDC4VC) specification. This method allows users to establish a high-trust identity from the moment they join the platform by presenting pre-existing, cryptographically verified credentials from trusted issuers like governments or financial institutions. The process also incorporates passkey creation, providing secure, passwordless access for subsequent logins.

The onboarding flow begins when a new user selects the "Register with Verifiable Credential" option. The application initiates an OIDC4VC-compliant request, prompting the user to connect their existing digital wallet. The user then selects a relevant Verifiable Credential—such as a digital driver's license, a bank-issued identity credential, or a proof of humanity certificate—to present to the application. The system verifies the credential's cryptographic signature, checks the issuer's status against a distributed trust registry, and confirms the credential has not been revoked. Upon successful verification, the user's profile is automatically created and populated with the verified information, and they are immediately granted a high initial trust score and a corresponding verification badge. As the final step, the user is prompted to create a passkey, which links their account to their device's biometric security (e.g., Face ID, fingerprint) for future passwordless authentication.

This feature's functionality depends on the user possessing a digital wallet that supports the OIDC4VC protocol and holds Verifiable Credentials from an issuer recognized by the platform's trust registry. The platform must maintain and regularly update a decentralized trust registry of approved issuers to prevent fraudulent credentials. Implementation requires integration with device-native WebAuthn/FIDO2 APIs to enable passkey creation and management, binding the user's identity to their device's hardware security module. This creates a dependency on the underlying operating system's support for these standards.

By adopting the OIDC4VC standard, the platform significantly reduces friction during onboarding, eliminating the need for manual data entry or multi-step email/SMS verification. It immediately establishes a high-trust environment by ensuring new users are authenticated against reliable, pre-vetted sources. This approach mitigates the risk of Sybil attacks and fraudulent account creation from the outset. For users, it offers a fast, secure, and privacy-preserving way to join the platform while retaining full control over their identity data. For the platform, it accelerates the growth of a verified user base, which is critical for the trust-based features and financial integrations.

## In-App High-Assurance Identity Verification and Reward

### In-App High-Assurance Identity Verification and Reward

This feature provides an optional, in-app workflow for users to generate a high-assurance Verifiable Credential by verifying their government-issued photo ID. This process enables the highest level of trust on the platform, unlocks advanced financial features, and rewards users with ECHO tokens for their participation.

Users can initiate this verification flow from their profile as a way to maximize their trust score and unlock payment capabilities. The user is prompted to either scan a government-issued photo ID, such as a driver's license, and complete a selfie-based liveness check, or, on compatible iOS devices, share their verified Apple Digital ID. Upon successful verification by a certified identity proofing service, a new high-assurance Verifiable Credential is issued directly to the user's wallet. This automatically elevates their trust score to the highest tier, grants them a premium "Identity Verified" badge, and enables access to regulated financial services within the app.

As a direct incentive for strengthening the network's trust layer, users who successfully complete this verification process are automatically rewarded with a significant amount of ECHO coin, such as 100 ECHO, credited to their account. This flow requires the user to provide their government ID and a live selfie, or consent to share their Apple Digital ID. The raw identity data is processed exclusively by a third-party identity verification partner and is not stored by the application, ensuring user privacy is maintained.

This feature is dependent on integration with a third-party identity verification service that is compliant with standards like NIST 800-63-3 IAL2 and is capable of issuing Verifiable Credentials. The Apple Digital ID pathway is specific to the iOS ecosystem and depends on the user having it pre-configured. The entire process must adhere to strict data privacy regulations for handling PII, and the reward mechanism depends on the ECHO token smart contract for automated distribution.

## Decentralized Bot Framework and Automation

### Decentralized Bot Framework and Automation

This feature enables developers to create and deploy autonomous bots that can interact with users and provide services within the messaging platform while operating on decentralized infrastructure and maintaining the platform's security and privacy standards. The bot framework supports a wide range of applications from simple utility bots to complex AI assistants and business automation tools.

Developers create bots using a comprehensive SDK that provides access to messaging APIs, payment processing, file sharing, and blockchain integration capabilities while enforcing strict security and privacy requirements. Bots operate as smart contracts deployed on the Constellation network, ensuring they cannot access user data beyond what is explicitly authorized and cannot be shut down by centralized authorities. The framework supports both simple rule-based bots and advanced AI-powered assistants that can process natural language requests while maintaining user privacy through local processing and zero-knowledge techniques.

Bot interactions are governed by the same trust and verification systems used for human users, with bots earning trust scores based on user feedback, functionality reliability, and security audit results. Users can discover bots through a decentralized marketplace where bot capabilities, trust scores, and user reviews are displayed transparently. Bot permissions are granular and user-controlled, allowing individuals to specify exactly what data and capabilities each bot can access, with all permissions revocable at any time.

The framework includes specialized bot types for common use cases including customer service bots for enterprise users, trading bots that can execute cryptocurrency transactions with user authorization, productivity bots that integrate with external services while maintaining privacy, and entertainment bots that provide games and interactive content. Revenue sharing mechanisms allow bot developers to monetize their creations through ECHO token payments, subscription models, or transaction fees, with all payments processed through the platform's secure payment infrastructure. The system includes comprehensive bot analytics and monitoring tools that help developers optimize their bots while respecting user privacy through anonymized usage statistics.

## Universal Onboarding and Identity Creation

### Universal Onboarding and Identity Creation

This feature provides a zero-PII registration flow for new users, starting with a username and passkey to create a self-sovereign decentralized identity (DID) in under 5 seconds. No phone number, email, or real name is collected at signup. The goal is to deliver an onboarding experience faster than every competitor while upholding ECHO's core promise: privacy from everyone, including ECHO.

The user journey begins upon first opening the app, where they are prompted to enter a desired username (3–24 characters, alphanumeric + underscore). The app checks username availability against the backend, then guides the user through passkey creation using Face ID or Touch ID. In the background, the app generates a P-256 key pair in the iOS Secure Enclave and sends only the public key and username to the backend. The backend submits a DID registration transaction to Cardano (fee paid from platform treasury) and returns the DID document. The user lands in the app at Trust Tier 1 — able to message immediately with a 1.0x reward multiplier. No SMS wait, no email confirmation, no centralized verification gate.

This onboarding process serves as the entry point to the platform's progressive trust system. After initial setup, the home screen displays contextual trust-tier upgrade cards encouraging users to voluntarily strengthen their profile. Phone verification (Tier 2) unlocks contact discovery and 1.2x rewards. Third-party identity verification (Tier 3) unlocks full rewards and group creation. Government ID verification (Tier 4) unlocks payment rails and financial features. Each upgrade is an informed, voluntary choice — the user sees what they gain before providing any information.

Phone verification is explicitly NOT part of onboarding. It is a separate opt-in flow triggered by the user tapping a trust-tier upgrade card. When initiated, the app collects the phone number, sends an SMS OTP for verification, and on success: (a) hashes the phone number on-device using Argon2id with a per-user salt, (b) sends only the hash to the server's contact discovery index, (c) discards the raw phone number immediately. The backend never stores raw phone numbers. This architecture ensures that phone verification adds value (contact discovery + higher rewards) without compromising the zero-PII signup promise.

Sybil defense is provided by Secure Enclave hardware binding (one account per physical device), the lowest reward multiplier for unverified accounts (1.0x — farming unprofitable at scale), auto-scaling reward rates that drop as the network grows, L1 anti-gaming validators (velocity checks, suspicious pattern detection), and Tier 3+ requirements for governance power and financial features (requires government ID + liveness check). These defenses are structurally stronger than phone number verification, which only prevents the laziest attack vectors.



## Privacy-Preserving Contact Discovery

### Privacy-Preserving Contact Discovery

This feature enables users to find friends and contacts who are already on ECHO without uploading their address book to a server or exposing their social graph. Contact discovery is the #1 adoption driver for any messaging platform, and ECHO must provide it without compromising the privacy guarantees that differentiate it from WhatsApp, Telegram, and Signal.

The system supports four discovery mechanisms, each preserving privacy through different approaches. First, phone number hashing: users can opt-in to match their existing contacts against ECHO's user base. The app hashes each contact's phone number using Argon2id with a per-user salt on-device before sending the hashed values to the server. The server matches hashed entries against its index of registered users (also stored as salted hashes) and returns encrypted DID references for matches — the server never sees raw phone numbers. Second, QR code DID exchange: users can share their DID directly via a QR code displayed in-app, enabling in-person contact sharing with zero server involvement. Third, username search: users who create a public handle (optional) can be discovered through the app's search interface. Handles are not linked to real names on-chain. Fourth, invite links: users can generate unique referral links that, when opened, establish a contact connection and trigger the 50 ECHO referral reward upon the new user's verification milestone.

Contact discovery requires careful privacy engineering because the phone-to-DID mapping is inherently sensitive. The server-side contact index stores only: Argon2id hashed phone numbers (not reversible without the per-user salt which stays on-device) linked to encrypted DID references. The index is not stored on any public blockchain. Users who decline phone-based discovery are completely invisible to the matching system and can only be found via QR code, username, or direct DID share.

This approach balances the practical need for users to find each other with ECHO's commitment to minimizing server-side knowledge of user relationships. The system is designed so that even a complete server breach reveals no usable phone numbers or social graph information — only irreversible hashes linked to encrypted pointers.

## Privacy Architecture and Secure Data Handling

## Overview

Privacy Architecture and Secure Data Handling defines the system-wide framework that ensures user data is protected at every layer — on-device, in-transit, at-rest on servers, and on public blockchains. This is not a single feature but the foundational security model that all other features depend on. It establishes the data classification hierarchy, encryption boundaries, biometric gates, metadata protection roadmap, and GDPR compliance mechanisms that collectively deliver ECHO's core promise: privacy from everyone, including ECHO itself.

The architecture is built on three principles. First, the device is the trust boundary — private keys, plaintext messages, and biometric templates never leave the user's hardware security module (Secure Enclave on iOS). Second, servers are untrusted by design — the Go backend relay handles only encrypted blobs and never possesses decryption keys. Third, blockchains are public by design — so ECHO stores only pseudonymous identifiers, hash commitments, and opaque references on-chain, never any data that could identify a person or reveal message content.

## Terminology

* **T0–T7 Data Classification**: ECHO's data sensitivity hierarchy. T0 (biometric templates, Secure Enclave private keys) is the most sensitive — never leaves hardware. T7 (public blockchain records like Merkle roots) is the least sensitive — designed to be publicly readable. Each tier has strict rules governing storage location, encryption requirements, and access controls.
* **Trust Boundary**: The security perimeter beyond which data is considered exposed to potential adversaries. In ECHO's model, the primary trust boundary is the iOS Secure Enclave. Data that crosses this boundary must be encrypted before transmission.
* **Sealed Sender**: A metadata protection technique (Phase 3) where the relay server knows the recipient but not the sender of a message, preventing traffic analysis of communication patterns.
* **Federated Relay**: A Phase 4 architecture where multiple independent relay operators handle message routing, ensuring no single operator sees all traffic metadata.

## Requirements

### REQ-PRIV-001: Data Classification Enforcement

**User Story:** As a user, I want ECHO to enforce strict data classification rules at every system boundary, so that my most sensitive data never appears where it shouldn't — not in server logs, not on blockchains, not in analytics.

**Acceptance Criteria:**

* AC-PRIV-001.1: The system shall enforce an 8-tier data classification hierarchy (T0 through T7) at every service boundary. Data shall not cross a boundary unless it meets the destination tier's requirements.
* AC-PRIV-001.2: T0 data (biometric templates, Secure Enclave private keys) shall never leave the iOS Secure Enclave hardware under any condition.
* AC-PRIV-001.3: T1 data (real name, government ID) shall exist only on the user's device and at the third-party IDV provider during verification — never on ECHO servers or blockchains.
* AC-PRIV-001.4: T2 data (phone number, email) shall be stored on-device only. If used for contact discovery, it shall be hashed with Argon2id and a per-user salt before any server transmission.
* AC-PRIV-001.5: T3 data (message plaintext, media) shall be encrypted on-device before transmission. The relay server shall see only T6/T7 data (ciphertext blobs, encrypted metadata).
* AC-PRIV-001.6: T4 data (contact lists, group membership) shall be encrypted locally and never transmitted to the server in plaintext form.
* AC-PRIV-001.7: T5 data (trust scores, interaction counts) shall be stored as hash commitments on-chain (H(score||nonce)), with raw values kept only on-device.
* AC-PRIV-001.8: T6 data (encrypted blobs in transit) may be temporarily held by the relay server for offline delivery but shall be deleted upon successful delivery confirmation.
* AC-PRIV-001.9: T7 data (DIDs, Merkle roots, token balances, trust tier brackets) is the only data class permitted on public blockchains.

### REQ-PRIV-002: No Plaintext on Servers

**User Story:** As a user, I want mathematical certainty that ECHO's servers cannot read my messages, so that even a compromised server reveals nothing about my conversations.

**Acceptance Criteria:**

* AC-PRIV-002.1: The Go backend relay shall never possess any decryption key for any user's messages. The relay processes only ciphertext and routing metadata.
* AC-PRIV-002.2: Server-side logs shall contain only: DID identifiers, timestamps, encrypted payload sizes, and delivery status. Message content, sender-recipient pairs (Phase 3+), and trust scores shall never appear in logs.
* AC-PRIV-002.3: PostgreSQL and Redis shall store only: encrypted message queue blobs (pending delivery), hashed contact discovery index entries, and cached DID-to-public-key mappings. No plaintext PII or message content.
* AC-PRIV-002.4: A compromised server backup or database dump shall yield zero usable personal data — only encrypted blobs and pseudonymous identifiers.

## Feature Behavior and Rules

### Data Tier Hierarchy

Data tiers are strictly ordered: T0 is the most sensitive and T7 is the least sensitive. A violation at any tier — such as T1 (name) data appearing in a server log or T3 (message) content appearing in a database — constitutes a privacy breach regardless of whether that data was encrypted in transit. The enforcement system shall apply tier checks at the service boundary, before data leaves the device or before the backend persists anything.

### Blockchain Privacy by Design

The public nature of the Constellation Hypergraph and Cardano blockchains is not a privacy risk for ECHO because no recoverable personal data is ever submitted. An adversary with full read access to both blockchains can determine: that a DID exists, what its public key is, what trust tier commitment is on record, what token balance it holds, and what Merkle roots have been anchored. They cannot determine: the real-world identity behind the DID, what messages were sent, who communicates with whom, what credentials are held, or what the exact trust score is. This property holds by construction, not by obscurity.

Privacy is not a feature toggle in ECHO — it is the foundation every other feature is built on. A user's real name, phone number, message content, biometrics, and private keys never reach any server or blockchain in any recoverable form. This architecture satisfies GDPR, CCPA, and HIPAA requirements by design, not by policy.

## Terminology

* **Secure Enclave**: A hardware-isolated security subsystem on iOS devices that stores private keys and requires biometric authentication (Face ID / Touch ID) for cryptographic operations. Keys generated in the Secure Enclave are never extractable.
* **Data Tier (T0–T7)**: An 8-level classification system that governs where each category of data may be stored. T0 (biometrics, private keys) may only exist in the Secure Enclave. T7 (public usernames) may be published on-chain if the user chooses.
* **Hash Commitment**: A one-way cryptographic construct of the form H(H(data) || nonce) that proves data existed at a point in time without revealing the data itself.
* **Merkle Root**: A single hash that cryptographically summarizes a batch of individual message commitments. Only the root is stored on-chain; individual messages are never exposed.
* **Zero-Knowledge Proof (ZKP)**: A cryptographic proof that demonstrates a statement is true (e.g., "I am Trust Tier 3+") without revealing the underlying data (e.g., the actual credential or score).
* **Reference ID**: An opaque UUID stored on-chain in place of a credential. It has no semantic meaning and cannot be reversed to reveal credential content or holder identity.
* **Forward Secrecy**: The property that compromise of a current session key does not expose past communications, because each session uses a freshly generated ephemeral key.
* **Blind Index**: A deterministic but unlinkable hash used for contact discovery. The server can match hashed phone numbers without ever learning the actual phone numbers.

## Requirements

### REQ-PRIV-001: Data Classification Enforcement

**User Story:** As a user, I want my personal information to be classified and handled according to strict privacy tiers, so that sensitive data never reaches servers or blockchains in readable form.

**Acceptance Criteria:**

* AC-PRIV-001.1: When any data is processed by the platform, it shall be assigned to one of eight tiers (T0–T7) that determine permissible storage locations.
* AC-PRIV-001.2: T0 data (biometrics, private keys) shall never be stored outside the device's Secure Enclave.
* AC-PRIV-001.3: T1 data (real names, DOB, SSN, addresses) shall never be transmitted to any server or blockchain in plaintext or recoverable form.
* AC-PRIV-001.4: T3 data (message content, files) shall only reach relay servers in end-to-end encrypted form; servers shall see only opaque ciphertext.
* AC-PRIV-001.5: T4 data (phone numbers, email addresses) shall only be stored as salted Argon2id hashes on servers; the raw values shall never be persisted server-side.
* AC-PRIV-001.6: T7 data (usernames, public keys) may be published on-chain only if the user explicitly chooses to do so.
* AC-PRIV-001.7: The system shall enforce data tier rules at the service layer, rejecting any operation that would violate tier constraints.

### REQ-PRIV-002: Device-Local Key Management

**User Story:** As a user, I want my private keys to be secured by my biometrics on my device, so that only I can authorize cryptographic operations and no one — including ECHO — can access my keys.

**Acceptance Criteria:**

* AC-PRIV-002.1: When a user creates their identity, the system shall generate a key pair inside the device's Secure Enclave (iOS) or StrongBox-backed KeyStore (Android), ensuring the private key is never extractable.
* AC-PRIV-002.2: When a cryptographic signing operation is required, the system shall present a biometric prompt (Face ID / Touch ID) before the Secure Enclave performs the operation.
* AC-PRIV-002.3: The system shall maintain a 3-tier key hierarchy: Device Root Key → Biometric-Protected Key → User Identity Key, with all application keys derived via HKDF from the User Identity Key.
* AC-PRIV-002.4: Derived application keys (message key, storage key, token key) shall be held in memory only and cleared when the app backgrounds.
* AC-PRIV-002.5: The system shall support key rotation without disrupting active sessions or requiring re-verification of credentials.
* AC-PRIV-002.6: When a user exports their public key (e.g., for DID registration), the private key shall not be included under any circumstances.

### REQ-PRIV-003: End-to-End Message Encryption

**User Story:** As a user, I want every message I send to be encrypted on my device before transmission, so that relay servers and third parties see only ciphertext and can never read my conversations.

**Acceptance Criteria:**

* AC-PRIV-003.1: When a user sends a message, the system shall encrypt it on-device using X25519 key agreement and ChaCha20-Poly1305 before the message leaves the device.
* AC-PRIV-003.2: The relay server shall receive only the encrypted payload, sender DID (pseudonymous), recipient DID (pseudonymous), and a timestamp — no plaintext content.
* AC-PRIV-003.3: For each message, the system shall generate a hash commitment H(H(plaintext) || nonce) that allows integrity verification without exposing content.
* AC-PRIV-003.4: Message commitments shall be batched into Merkle trees and only the Merkle root shall be anchored on-chain; no individual message data shall reach the blockchain.
* AC-PRIV-003.5: The system shall use ephemeral key pairs for each session to ensure forward secrecy — compromise of one session key shall not expose any previous communications.
* AC-PRIV-003.6: All local message storage shall be encrypted at rest using AES-GCM keys derived from the Secure Enclave, requiring biometric unlock to access.

### REQ-PRIV-004: Privacy-Preserving Blockchain Data

**User Story:** As a user, I want any data stored on the public blockchain to be unlinkable to my real identity, so that public blockchain access reveals nothing about who I am or what I communicate.

**Acceptance Criteria:**

* AC-PRIV-004.1: When the system stores identity data on-chain, it shall store only the user's DID and public key — no name, email, phone, address, or any other PII.
* AC-PRIV-004.2: When the system stores trust score data on-chain, it shall store only a commitment H(score || nonce) and the trust tier — not the exact score.
* AC-PRIV-004.3: When the system stores credential data on-chain, it shall store only an opaque reference UUID, the issuer DID, the credential type, and a revocation status bit — not the credential content or holder identity.
* AC-PRIV-004.4: When the system stores token balance data on-chain, balances shall be linked to pseudonymous DIDs only — not real-world identities.
* AC-PRIV-004.5: Contact discovery shall use a blind index approach: the server shall match hashed phone numbers between users without ever learning the actual phone numbers. The hash shall use Argon2id with a per-user salt known only to the user's device.
* AC-PRIV-004.6: The system shall use opaque UUID reference IDs for any on-chain data that maps to off-chain records; the mapping shall exist only on the user's device.

### REQ-PRIV-005: Zero-Knowledge Verification (Phase 3+)

**User Story:** As a user, I want to prove attributes about myself (my age, trust tier, or credential validity) without revealing the underlying data, so that I can satisfy verification requirements while preserving my privacy.

**Acceptance Criteria:**

* AC-PRIV-005.1: When an age verification is required, the system shall generate a ZK proof that the user is over the required threshold (18/21) without revealing the user's actual birthdate.
* AC-PRIV-005.2: When a trust tier check is required (e.g., for governance voting), the system shall generate a ZK proof that the user meets the minimum tier without revealing their exact score.
* AC-PRIV-005.3: When a credential validity check is required, the system shall generate a ZK proof that the credential is valid and issued by the claimed issuer without revealing the credential content.
* AC-PRIV-005.4: When a balance threshold check is required (e.g., for staking eligibility), the system shall generate a ZK proof that the user holds at least the required amount without revealing the exact balance.
* AC-PRIV-005.5: ZK proofs shall be verified on-device or via the Midnight partner chain (Phase 4+) before any transaction is submitted on-chain.
* AC-PRIV-005.6: The Midnight integration shall enable Organization-tier clients to obtain private KYC proofs and compliance verification without exposing customer data to the public Hypergraph.

### REQ-PRIV-006: Identity Verification Without PII Exposure

**User Story:** As a user, I want to complete identity verification without ECHO ever seeing my government-issued ID or personal information, so that I gain trust tier benefits without surrendering my privacy to the platform.

**Acceptance Criteria:**

* AC-PRIV-006.1: When a user initiates identity verification, the system shall direct the verification session to a third-party IDV provider via a direct TLS connection — the ECHO platform backend shall never receive ID document images, selfies, or extracted PII.
* AC-PRIV-006.2: The IDV provider shall return to the ECHO backend only: pass/fail result, confidence score, document type, issuing country, and age-over-threshold boolean — no names, DOB, document numbers, or addresses.
* AC-PRIV-006.3: The system shall store the IDV result on-chain as an opaque reference ID with credential type and assurance level only — not any PII returned by the IDV provider.
* AC-PRIV-006.4: The IDV provider shall delete all captured images immediately after processing and shall not retain any PII beyond the verification session.

## Feature Behavior and Rules

### Data Tier Hierarchy

Data tiers are strictly ordered: T0 is the most sensitive and T7 is the least sensitive. A violation at any tier — such as T1 (name) data appearing in a server log or T3 (message) content appearing in a database — constitutes a privacy breach regardless of whether that data was encrypted in transit. The enforcement system shall apply tier checks at the service boundary, before data leaves the device or before the backend persists anything.

### Blockchain Privacy by Design

The public nature of the Constellation Hypergraph and Cardano blockchains is not a privacy risk for ECHO because no recoverable personal data is ever submitted. An adversary with full read access to both blockchains can determine: that a DID exists, what its public key is, what trust tier commitment is on record, what token balance it holds, and what Merkle roots have been anchored. They cannot determine: the real-world identity behind the DID, what messages were sent, who communicates with whom, what credentials are held, or what the exact trust score is. This property holds by construction, not by obscurity.

### Biometric Requirement Scope

Biometric authentication is required for: generating new keys, signing DID operations, decrypting local message storage, performing staking or wallet transactions, and accessing hidden folders. Biometric authentication is not required for: reading cached plaintext messages already decrypted in an active session, browsing the public feed, or viewing non-sensitive profile information. This scope ensures security without friction for everyday use.

### Metadata Protection Phases

Even with content encrypted, communication metadata (who talks to whom, when, how often) can reveal sensitive information. ECHO addresses this progressively: Phase 1-2 uses TLS 1.3 transport, meaning the relay server knows sender and recipient DIDs and timestamps. Phase 3 introduces sealed sender, so the server knows the recipient but not the sender. Phase 4 deploys federated relay nodes, ensuring no single operator sees all traffic. Phase 4+ enables optional direct P2P for both-online users, eliminating the relay hop entirely. Each phase materially reduces the metadata surface area.

### GDPR Right to Erasure

Because all PII is stored on the user's device, GDPR erasure ("right to be forgotten") is implemented by deleting the user's Secure Enclave keys. Once the keys are deleted, all locally encrypted data becomes unrecoverable. Off-chain server data (hashed phone index entries, encrypted message queue) is deleted upon account deletion request. On-chain data (DIDs, commitments, token balances) is pseudonymous and contains no PII; however, the user's DID can be deactivated on Cardano, rendering it inactive while the historical record remains (consistent with blockchain immutability).

### Secure Enclave Key Management

## Overview

Secure Enclave Key Management governs how ECHO generates, stores, uses, and protects all cryptographic keys on the user's device. The iOS Secure Enclave is a hardware-isolated coprocessor that stores private keys in a way that makes them inaccessible to the application processor, the operating system, or any software — including ECHO itself. This feature ensures that the user's DID private key, message encryption keys, and wallet signing keys never exist outside the Secure Enclave hardware, providing protection against device compromise, app sandbox escapes, and even physical device attacks.

This is the foundational security feature that all other cryptographic operations depend on. Without correct Secure Enclave key management, E2E encryption, DID signing, wallet transactions, and biometric gating cannot function securely.

## Terminology

* **Secure Enclave**: Apple's hardware security module embedded in iOS devices (A7 chip and later). It runs its own microkernel, has its own encrypted memory, and performs cryptographic operations internally — private keys never leave the hardware.
* **Key Reference**: A handle (not the key itself) that the application holds to request the Secure Enclave to perform operations (sign, decrypt) using a stored key. The application never sees the raw private key material.
* **Biometric Gate**: A requirement that the user authenticate via Face ID or Touch ID before the Secure Enclave will perform a cryptographic operation. This binds key usage to physical user presence.
* **Key Derivation Function (KDF)**: A function that derives purpose-specific keys from a master key. ECHO uses HKDF-SHA256 to derive separate keys for message encryption, local storage encryption, and DID operations from the Secure Enclave master key.
* **Forward Secrecy**: The property that compromise of a long-term key does not compromise past session keys. Achieved through ephemeral key pairs generated per session.

## Requirements

### REQ-SE-001: Key Generation in Hardware

**User Story:** As a user, I want my private keys generated inside the Secure Enclave hardware so that no software — including ECHO — can ever extract or copy my keys.

**Acceptance Criteria:**

* AC-SE-001.1: When a user creates an ECHO account, the system shall generate an asymmetric key pair (P-256 or Curve25519) inside the Secure Enclave using the `SecKeyCreateRandomKey` API with the `kSecAttrTokenIDSecureEnclave` attribute.
* AC-SE-001.2: The private key shall have the `.privateKeyUsage` access control flag set, ensuring it can only be used for signing and key agreement operations inside the Secure Enclave — never exported.
* AC-SE-001.3: The public key shall be extracted from the Secure Enclave and used as the basis for the user's DID document registration on Cardano.
* AC-SE-001.4: The system shall verify Secure Enclave availability at app launch. If the device does not support Secure Enclave (pre-A7 devices), the app shall refuse to create an account and display an explanation.

### REQ-SE-002: Biometric-Gated Key Operations

**User Story:** As a user, I want all sensitive cryptographic operations to require my biometric authentication, so that even someone with physical access to my unlocked phone cannot sign transactions or decrypt messages without my face or fingerprint.

**Acceptance Criteria:**

* AC-SE-002.1: The Secure Enclave key shall be created with `.biometryCurrentSet` access control, requiring Face ID or Touch ID authentication before any private key operation.
* AC-SE-002.2: The following operations shall require biometric authentication: DID document signing, message decryption key derivation (when app returns from background), wallet transaction signing (staking, delegation, transfers), hidden folder decryption, and account recovery initiation.
* AC-SE-002.3: The following operations shall NOT require biometric authentication: reading already-decrypted cached messages in an active session, viewing public profile information, and browsing the app's public feed.
* AC-SE-002.4: If biometric authentication fails 5 consecutive times, the system shall require the device passcode as fallback. After 10 total failures, the app shall lock for 15 minutes.

### REQ-SE-003: Key Hierarchy and Derivation

**User Story:** As a user, I want separate keys for different purposes so that compromising one key type does not compromise all my data.

**Acceptance Criteria:**

* AC-SE-003.1: The system shall maintain a key hierarchy with the Secure Enclave master key at the root, deriving purpose-specific keys using HKDF-SHA256 with unique context strings: "echo-did-signing" for DID operations, "echo-msg-encryption" for message key agreement, "echo-storage-encryption" for local database encryption, and "echo-wallet-signing" for token transactions.
* AC-SE-003.2: Ephemeral session keys for message encryption shall be generated per-session using X25519 key agreement between the sender's ephemeral key and the recipient's long-term public key. The ephemeral private key shall be discarded after the shared secret is derived.
* AC-SE-003.3: The local storage encryption key shall be derived on-demand when the user authenticates biometrically. It shall not persist in application memory when the app is backgrounded.
* AC-SE-003.4: Key derivation shall use unique salts per purpose to ensure derived keys are cryptographically independent — compromise of the storage key reveals nothing about the signing key.

### REQ-SE-004: Key Lifecycle and Recovery

**User Story:** As a user, I want a way to recover my account if I lose my device, without compromising the security of the Secure Enclave model.

**Acceptance Criteria:**

* AC-SE-004.1: When a user initiates account recovery, the system shall generate a 24-word BIP-39 mnemonic recovery phrase derived from the Secure Enclave master key's public parameters and a user-provided passphrase.
* AC-SE-004.2: The recovery phrase shall be displayed once during initial account setup. The user must confirm they have saved it by re-entering selected words. The phrase is never stored on any server.
* AC-SE-004.3: On a new device, entering the recovery phrase shall generate a new Secure Enclave key pair and submit a DID document update to Cardano, rotating the public key while maintaining the same DID identifier.
* AC-SE-004.4: The old device's keys become invalid after key rotation. Messages encrypted with the old key cannot be decrypted on the new device unless the user has enabled encrypted backup (opt-in, biometric-protected backup to iCloud Keychain).
* AC-SE-004.5: Multi-device sync (Phase 3) shall use per-device Secure Enclave keys with cross-device key agreement — each device has its own hardware-bound key, and messages are encrypted separately for each device's public key.

## Feature Behavior and Rules

### Key Never Leaves Hardware

The Secure Enclave's security guarantee is absolute: the private key material cannot be read by any software, including the iOS kernel. The application holds a key reference (a handle) that allows it to request the Secure Enclave to perform operations, but the key itself is inaccessible. This means there is no "export key" function, no backup of the raw key, and no way to transfer the exact key to another device. Recovery requires key rotation, not key transfer.

### Biometric Template Isolation

Face ID and Touch ID biometric templates are stored inside the Secure Enclave's own encrypted memory, separate from the application. ECHO never receives biometric data — it receives only a success/failure result from the Secure Enclave after biometric matching. This means ECHO cannot collect, store, or transmit biometric information under any circumstances.

### Background Key Purging

When the iOS app transitions to the background state, the system shall clear all derived key material from application memory. When the user returns to the app, biometric authentication is required to re-derive the storage decryption key and resume access to message history. This ensures that a memory dump of a backgrounded app reveals no usable key material.




## Overview

End-to-End Message Encryption and Commitment ensures that message content is encrypted on the sender's device before transmission and that relay servers see only opaque ciphertext. In addition to confidentiality, this feature provides cryptographic integrity guarantees: each message generates a hash commitment that can prove the message existed and was unaltered, without ever revealing its content to the blockchain or any third party.

Phase 1 cryptographic primitives (commitment generation, key derivation, signing) are complete. Phase 2 implements the full E2EE message flow including ChaCha20-Poly1305 encryption, ephemeral key agreement, and local encrypted storage.

## Terminology

* **Ephemeral Key**: A freshly generated key pair created per-session for X25519 key agreement. Discarding it after the session ends provides forward secrecy.
* **X25519**: A Diffie-Hellman key agreement protocol over Curve25519 used to derive a shared secret between sender and recipient without transmitting the secret.
* **ChaCha20-Poly1305**: An authenticated encryption scheme. ChaCha20 provides confidentiality; Poly1305 provides integrity and authenticity. The combination is AEAD (Authenticated Encryption with Associated Data).
* **Hash Commitment**: H(H(plaintext) || nonce || timestamp) — a one-way transformation that proves a message existed and was unaltered without revealing its content.
* **Merkle Root**: The single root hash of a binary tree built from individual message commitments. Only this root is anchored on-chain.
* **Merkle Proof**: A set of sibling hashes that allows a recipient to verify that a specific message commitment is included in a published Merkle root.

## Requirements

### REQ-E2E-001: On-Device Encryption Before Transmission

**User Story:** As a user, I want my messages encrypted on my device before they are sent, so that the relay server and any network observer see only ciphertext and can never read my conversations.

**Acceptance Criteria:**

* AC-E2E-001.1: When a user sends a message, the system shall encrypt the plaintext on-device using ChaCha20-Poly1305 with a shared secret derived via X25519 key agreement between the sender's ephemeral key and the recipient's long-term public key.
* AC-E2E-001.2: The system shall generate a new ephemeral key pair for each session; the ephemeral private key shall be discarded after the shared secret is derived, ensuring forward secrecy.
* AC-E2E-001.3: The encrypted payload transmitted to the relay server shall contain: ephemeral public key, ciphertext, encryption nonce, hash commitment, and sender signature — no plaintext content.
* AC-E2E-001.4: The relay server shall be structurally incapable of decrypting message content, as it never receives a private key or shared secret.
* AC-E2E-001.5: Group messages shall use the same per-session ephemeral key agreement model, with the sender encrypting independently for each recipient's public key, or via a group key ratchet mechanism.

### REQ-E2E-002: Hash Commitments for Message Integrity

**User Story:** As a user, I want a cryptographic record that proves my messages existed and were unaltered, so that I can verify integrity or present proof in disputes without exposing message content.

**Acceptance Criteria:**

* AC-E2E-002.1: When a message is sent, the system shall generate a hash commitment using the formula H(H(plaintext) || nonce || timestamp), where the nonce is a 32-byte cryptographically random value generated per message.
* AC-E2E-002.2: The nonce and timestamp shall be stored locally alongside the message; without them, the commitment cannot be verified — providing an additional layer of protection against rainbow table attacks.
* AC-E2E-002.3: When a user needs to prove a message occurred (e.g., for a legal dispute), they shall be able to provide the plaintext, nonce, and timestamp, which the verifier can use to recompute and check against the on-chain commitment.
* AC-E2E-002.4: Commitment verification shall use constant-time comparison to prevent timing attacks.

### REQ-E2E-003: Merkle Tree Batching for On-Chain Anchoring

**User Story:** As a user, I want my message integrity proofs anchored to the blockchain without flooding it with individual transactions, so that provability is maintained at scale and at low cost.

**Acceptance Criteria:**

* AC-E2E-003.1: The system shall batch message commitments into Merkle trees at regular intervals (target: hourly at launch, adjustable by governance).
* AC-E2E-003.2: Only the Merkle root shall be submitted to the Constellation metagraph (Data L1) — individual message commitments shall not appear on-chain.
* AC-E2E-003.3: For each batch, the system shall generate and store Merkle proofs for every included message, allowing any individual message to be proven against the published root.
* AC-E2E-003.4: The on-chain batch record shall contain: Merkle root, batch timestamp, and message count — no message content, no sender or recipient identifiers.
* AC-E2E-003.5: The system shall support Merkle proof verification on the client, enabling a recipient to confirm their message is included in a given on-chain root without querying a server.

### REQ-E2E-004: Encrypted Local Storage

**User Story:** As a user, I want my messages stored on my device in an encrypted form, so that physical device access or a compromised app sandbox cannot expose my conversation history.

**Acceptance Criteria:**

* AC-E2E-004.1: When messages are persisted locally, the system shall encrypt them using AES-GCM with a storage key derived from the Secure Enclave (see Secure Enclave Key Management feature).
* AC-E2E-004.2: The storage key shall not be held in memory when the app is backgrounded; accessing local message history after backgrounding shall require a biometric unlock to re-derive the key.
* AC-E2E-004.3: Each stored record shall use a unique random nonce; deterministic nonces shall be prohibited to prevent ciphertext reuse.
* AC-E2E-004.4: When a user deletes a message, the plaintext and ciphertext shall be securely wiped from local storage; the on-chain commitment remains as an immutable integrity anchor but reveals nothing about content.

## Feature Behavior and Rules

### Offline Message Queue

When the recipient is offline, the relay server temporarily holds the encrypted message blob. The server never decrypts it — it holds ciphertext only. Once the recipient device comes online, the blob is delivered and the server deletes its copy. The server's possession of the encrypted blob does not constitute a privacy violation because decryption requires the recipient's private key, which never leaves their Secure Enclave.

### Provable Mode

Users can enable "provable mode" for a conversation, which activates explicit commitment generation and Merkle anchoring for every message in that thread. In standard mode, commitments are still generated but batching is optimized for cost. In provable mode, batching is more frequent and the user receives an in-app receipt linking each message to its on-chain commitment. Provable mode is recommended for legal, financial, and compliance-sensitive conversations.


## Overview

The Privacy-Preserving Blockchain Data Model defines what ECHO stores on public blockchains and, critically, what it never stores. Because the Constellation Hypergraph and Cardano are public ledgers readable by anyone, ECHO must ensure that on-chain data reveals nothing about real-world identities, message content, or communication patterns — even to an adversary with full blockchain access. This is achieved through a combination of pseudonymous identifiers, hash commitments, opaque reference IDs, and blind index hashing.

This feature defines the data boundaries at the product level. The corresponding blueprint specifies the cryptographic implementation in detail.

## Terminology

* **DID (Decentralized Identifier)**: A pseudonymous identifier anchored on Cardano. It is linked to a public key, not a real-world identity. A user controls their DID through their Secure Enclave private key.
* **Opaque Reference ID**: A UUID v4 with no embedded information, stored on-chain as a pointer to off-chain data. Without the off-chain mapping (which exists only on the user's device), the UUID is meaningless.
* **Salted Hash**: A hash of the form H(salt || data) where the salt is known only to the user's device. This prevents brute-force reversal of phone numbers or email addresses via rainbow tables.
* **Trust Tier Commitment**: H(score || nonce) stored on-chain. Reveals only the tier bracket (1-5), not the exact score. The nonce prevents reverse-engineering of the score.
* **Revocation Bit**: A single bit in a compressed status list that indicates whether a credential has been revoked. Checking revocation requires only the bit position, not any credential content.

## Requirements

### REQ-BPM-001: Identity Data On-Chain

**User Story:** As a user, I want my blockchain identity to contain no personal information, so that anyone who reads the blockchain cannot determine who I am in the real world.

**Acceptance Criteria:**

* AC-BPM-001.1: When a DID document is registered on Cardano, it shall contain only: the DID identifier, the user's public key, verification method references, and creation/update timestamps.
* AC-BPM-001.2: The DID document shall not contain any of: real name, email, phone number, address, date of birth, government ID number, or any other PII.
* AC-BPM-001.3: The DID identifier shall be derived from the user's public key using a one-way function, making it computationally infeasible to map a DID back to a real-world identity without the private key.
* AC-BPM-001.4: When a user changes their display name or avatar, those changes shall not trigger a DID document update; DID documents contain only cryptographic material, not profile data.

### REQ-BPM-002: Trust Score Privacy On-Chain

**User Story:** As a user, I want my trust score stored as a cryptographic commitment rather than a raw value, so that no one can determine my exact score from the blockchain.

**Acceptance Criteria:**

* AC-BPM-002.1: When a trust score update is committed to the metagraph, the system shall store H(score || nonce) and the tier bracket (1–5) only — not the raw numeric score.
* AC-BPM-002.2: The nonce shall be unique per update and stored only on the user's device, ensuring the commitment cannot be reversed even by an adversary who knows the approximate score range.
* AC-BPM-002.3: For governance purposes, the system shall use the trust tier (1–5) for weight calculation, not the exact score — ensuring on-chain governance operations never expose score precision.
* AC-BPM-002.4: Trust tier history shall not accumulate on-chain; only the current tier commitment shall be stored, preventing behavioral pattern analysis from historical tier changes.

### REQ-BPM-003: Credential References On-Chain

**User Story:** As a user, I want my credentials represented on-chain by opaque references rather than content, so that my ID type, issuer, and verification status are not exposed to public scrutiny.

**Acceptance Criteria:**

* AC-BPM-003.1: When a verifiable credential is linked to a DID, the system shall store on-chain only: an opaque UUID reference, the issuer DID, the credential type string (e.g., "DriversLicense"), issuance timestamp, and a revocation status index.
* AC-BPM-003.2: The on-chain credential reference shall contain no holder name, date of birth, document number, address, or any field from the credential content.
* AC-BPM-003.3: Credential revocation shall be implemented via a compressed bit vector status list; checking revocation requires only reading the bit at the credential's status index, not accessing any credential content.
* AC-BPM-003.4: The mapping between an on-chain credential reference UUID and the actual credential shall exist only on the user's device; loss of the device without backup means the credential must be re-issued.

### REQ-BPM-004: Contact Discovery Without PII Exposure

**User Story:** As a user, I want to find my contacts on ECHO without uploading my address book to a server, so that ECHO never learns who is in my contact list.

**Acceptance Criteria:**

* AC-BPM-004.1: When a user enables contact discovery, the system shall hash their contacts' phone numbers using Argon2id with a per-user salt known only to the user's device before transmitting any data to the server.
* AC-BPM-004.2: The server shall match hashed phone numbers from multiple users to identify mutual contacts, without ever receiving or storing any raw phone numbers.
* AC-BPM-004.3: The contact discovery index on the server shall store only: salted hashes (not reversible without the per-user salt) linked to encrypted DID references.
* AC-BPM-004.4: Contact discovery shall be opt-in; users who decline shall be discoverable only via direct DID share or QR code.
* AC-BPM-004.5: The contact discovery index shall not be stored on any public blockchain.

### REQ-BPM-005: Token Balances Pseudonymous

**User Story:** As a user, I want my ECHO token balance linked to my pseudonymous DID only, so that my holdings cannot be traced back to my real-world identity.

**Acceptance Criteria:**

* AC-BPM-005.1: Token balances, staking positions, and delegation records on the Constellation metagraph shall be keyed by DID only — no real-world identity metadata shall be included in any token transaction.
* AC-BPM-005.2: The system shall not associate a user's token address with any PII in any on-chain data structure.
* AC-BPM-005.3: Governance votes recorded on-chain shall include only the voter's DID, their effective weight (staked ECHO × tier multiplier), and their vote — not their name, tier history, or other identifying information.

## Feature Behavior and Rules

### What an Adversary Can Learn

With full read access to both the Constellation Hypergraph and Cardano blockchain, an adversary can determine: that a DID exists, what its public key is, what trust tier it currently holds, what token balance it has, what Merkle roots have been anchored, and what credential types are associated with it. They cannot determine: the real-world person behind the DID, what messages were exchanged, who communicates with whom, what credentials contain, or what the precise trust score is. This property holds structurally, not by policy.

### On-Chain Data Is Permanent

Any data submitted to the public blockchain is permanent and immutable. For this reason, the classification system is strictly enforced before any data reaches the submission layer. A privacy boundary violation — submitting T1-T5 data to the blockchain — cannot be undone. The enforcement gate at the service layer is therefore a hard stop, not a warning.


## Overview

Zero-Knowledge Proofs and Midnight Integration enables ECHO users and organizations to verify attributes about themselves — age, trust tier, credential validity, token balance — without revealing the underlying data. This is the most advanced layer of ECHO's privacy stack. Rather than sharing a credential (which exposes its content), users share a cryptographic proof that the credential satisfies a condition.

Midnight is Cardano's privacy-focused partner chain, using ZK-SNARKs and the Compact smart contract language. ECHO evaluates Midnight in Phase 3 and integrates it in Phase 4, providing a production-grade ZK verification environment — particularly valuable for Organization-tier enterprise clients who need compliance verification without public data exposure.

## Terminology

* **ZK-SNARK**: Zero-Knowledge Succinct Non-Interactive Argument of Knowledge. A proof system that allows one party to prove they know something without revealing what they know. "Succinct" means the proof is small and fast to verify.
* **Circuit**: The logical description of what a ZK proof must demonstrate (e.g., "birthdate is before threshold date"). Circuits are compiled into proving and verification keys.
* **Proving Key / Verification Key**: A paired key set generated from a circuit. The prover uses the proving key to generate a proof; the verifier uses the verification key to check it. Neither key reveals the private inputs.
* **Public Signals**: The outputs of a ZK proof visible to the verifier (e.g., "isOverThreshold: true"). The private inputs (e.g., the actual birthdate) are not visible.
* **Midnight**: A Cardano partner chain built specifically for privacy-preserving smart contracts using ZK-SNARKs. It uses the Compact language for contract authoring and supports selective disclosure of on-chain state.
* **Compact**: The smart contract language for Midnight, designed for privacy-preserving computations.
* **Groth16**: The specific ZK proof system used by the initial circuits — compact proofs, fast verification, trusted setup required.

## Requirements

### REQ-ZK-001: Age Verification Without Birthdate Disclosure (Phase 3+)

**User Story:** As a user, I want to prove I am over the required age threshold without revealing my actual birthdate, so that I can access age-restricted features while preserving my privacy.

**Acceptance Criteria:**

* AC-ZK-001.1: When age verification is required (e.g., for financial institution integration), the system shall generate a ZK proof demonstrating the user's age exceeds the threshold (18 or 21) using the user's device-stored credential.
* AC-ZK-001.2: The public signals presented to the verifier shall contain only: the threshold value and a boolean `meetsThreshold: true` — not the user's birthdate, exact age, or any other PII.
* AC-ZK-001.3: The verifier shall be able to confirm the proof using only the verification key and public signals, without access to any private input.
* AC-ZK-001.4: Proof generation shall occur on-device; the private inputs (birthdate) shall not leave the device at any point during the proving process.

### REQ-ZK-002: Trust Tier Verification Without Score Disclosure (Phase 3+)

**User Story:** As a user, I want to prove I meet a minimum trust tier for governance or feature access without revealing my exact score, so that my reputation data remains private.

**Acceptance Criteria:**

* AC-ZK-002.1: When a trust tier threshold check is required (e.g., Tier 3+ for governance eligibility or Tier 4+ for financial institution features), the system shall generate a ZK proof that the user's score commitment corresponds to a score meeting the threshold.
* AC-ZK-002.2: The public signals shall contain only: the minimum tier required and `meetsThreshold: true` — not the exact score or score commitment nonce.
* AC-ZK-002.3: The proof shall be verifiable against the on-chain trust tier commitment (H(score || nonce)) without the verifier ever learning the score.
* AC-ZK-002.4: The system shall support proof generation for governance voting weight calculation, allowing the smart contract to apply the correct tier multiplier without a raw score being broadcast.

### REQ-ZK-003: Credential Validity Without Content Disclosure (Phase 3+)

**User Story:** As a user, I want to prove a credential I hold is valid and issued by a trusted authority without revealing what the credential contains, so that I can satisfy verification requirements without surrendering private information.

**Acceptance Criteria:**

* AC-ZK-003.1: When credential verification is required, the system shall generate a ZK proof demonstrating: the credential was issued by the claimed issuer, the credential has not been revoked, and the credential holder is the current user (via DID binding).
* AC-ZK-003.2: The public signals shall contain only: the issuer DID, the credential type, and `isValid: true` — not the credential subject's name, DOB, document number, or any field from the credential body.
* AC-ZK-003.3: The proof shall include a response to the verifier's challenge (a random nonce) to prevent replay attacks.
* AC-ZK-003.4: Credential proofs shall be verified before any access grant; access shall not be granted on the basis of an unverified credential reference alone.

### REQ-ZK-004: Balance Threshold Proof Without Balance Disclosure (Phase 3+)

**User Story:** As a user, I want to prove I hold a sufficient ECHO token balance for staking eligibility or marketplace access without revealing my exact holdings, so that my financial position remains private.

**Acceptance Criteria:**

* AC-ZK-004.1: When a balance threshold check is required (e.g., minimum stake for validator eligibility or VIP feature access), the system shall generate a ZK proof that the user's balance meets or exceeds the threshold.
* AC-ZK-004.2: The public signals shall contain only: the threshold and `meetsThreshold: true` — not the exact balance.
* AC-ZK-004.3: The proof shall be bound to the user's current balance commitment on-chain, making it infeasible to generate a valid proof for a balance the user does not hold.

### REQ-ZK-005: Midnight Evaluation and Integration (Phase 3 Evaluate, Phase 4 Integrate)

**User Story:** As an enterprise Organization-tier customer, I want ZK credential verification through Midnight so that my compliance proofs and KYC verifications are privacy-preserving and legally admissible without exposing customer data on a public chain.

**Acceptance Criteria:**

* AC-ZK-005.1: In Phase 3, the team shall evaluate Midnight mainnet stability, assess Compact contract tooling maturity, and deliver a proof-of-concept demonstrating "Prove I'm Trust Tier 3+ without revealing my credential" using Midnight's ZK infrastructure.
* AC-ZK-005.2: In Phase 4, the system shall integrate Midnight for production ZK trust tier verification, with proofs generated on the Midnight partner chain and commitments verifiable against Cardano.
* AC-ZK-005.3: Organization-tier enterprise clients shall be able to obtain private KYC proofs through Midnight: proving a customer is verified without the verification details appearing on the public Constellation or Cardano chains.
* AC-ZK-005.4: Group membership proofs (prove a user is a member of an enterprise group without revealing the group roster) shall be supported via Midnight Compact contracts in Phase 4.
* AC-ZK-005.5: The Midnight integration shall not break or replace Phase 1-3 privacy guarantees; it shall extend them with a higher-assurance ZK layer for enterprise use cases.

## Feature Behavior and Rules

### Phase Gating

ZK proof features are explicitly Phase 3+ because they require: stable ZK circuit libraries, Midnight mainnet availability, and completion of Phases 1-2 cryptographic foundations. Attempting to ship ZK features before the encryption and commitment layer is complete would create a false sense of security. The Phase 3 evaluation milestone is a deliberate risk gate — if Midnight is not production-stable, the ZK layer falls back to the existing commitment scheme until Phase 4.

### On-Device Proof Generation

ZK proof generation occurs on the user's device, not on a server. Private inputs (birthdate, score, credential content) are never transmitted during proving. This is a hard requirement — server-side proving would require the server to receive private inputs, which defeats the purpose of the ZK proof entirely. Proof generation latency is acceptable (target: under 5 seconds on a modern iPhone) because ZK operations are infrequent and user-initiated.

### Enterprise Midnight Use Cases

The primary enterprise use cases for Midnight are: (1) banks proving a customer completed KYC to a compliance auditor without exposing the customer's identity documents; (2) enterprises proving group membership for access control without broadcasting the access control list; (3) compliance verification for regulated messaging where proof of verification is required but the verification data itself is subject to privacy law. These use cases are the reason the Midnight evaluation is prioritized at Phase 3 rather than left to Phase 5+.

## ECHO Tokenomics, Founder Allocation, and Token Launch

## Overview

ECHO Tokenomics defines the complete supply, distribution, emission, vesting, and governance model for the ECHO token. The design is built around one principle: all users are owners. The fixed 1 billion supply, transparent on-chain founder vesting, trust-tier weighted governance, and community-first emission curve are designed to create a platform where value flows to participants, not extractors. This document covers the token genesis mechanics, rate of issuance to users, founder allocation and vesting, and the token launch sequence.

## Terminology

* **Genesis**: The single event at Phase 2 mainnet launch where all 1,000,000,000 ECHO tokens are minted. No tokens are minted after genesis.
* **Emission Curve**: The schedule by which community reward tokens are released from the protocol-controlled pool to users. Front-loaded toward early adopters; declining annually over 10 years.
* **TokenLock**: A Tessellation v3 primitive that locks ECHO for a defined period. Used for founder vesting, user staking, and validator requirements. Enforced by Currency L1 Scala validation — cannot be bypassed.
* **WithdrawLock**: A Tessellation v3 primitive creating a 14-day cooldown before locked tokens become transferable. Prevents immediate dumping of newly vested or unstaked tokens.
* **AtomicAction**: A Tessellation v3 primitive bundling multiple operations (verify tier + claim reward + update daily cap) into a single indivisible transaction. Prevents reward gaming.
* **Cliff**: The 12-month period from genesis during which no founder vesting occurs. All founder tokens remain locked regardless of time elapsed until the cliff date passes.
* **FDV (Fully Diluted Valuation)**: Market cap if all 1B tokens were in circulation at the current price. Reference metric for evaluating allocation sizes.
* **Deflationary Pressure**: Phase 5 burns 30% of annual treasury surplus to permanently remove ECHO from circulation, reducing supply as revenue grows.

## Requirements

### REQ-TOK-001: Fixed Supply and Genesis Allocation

**User Story:** As a token holder, I want the total ECHO supply fixed at genesis and publicly verifiable, so that I can trust no additional tokens will ever be minted to dilute my holdings.

**Acceptance Criteria:**

* AC-TOK-001.1: At Phase 2 mainnet launch, the Currency L1 Scala genesis block shall mint exactly 1,000,000,000 ECHO tokens and allocate them to five protocol-controlled pools: Community Rewards 40% (400M), Treasury 22% (220M), Founders 18% (180M), Future Team & Advisors 10% (100M), Ecosystem & Partnerships 10% (100M).
* AC-TOK-001.2: No additional minting shall be possible after genesis. The Currency L1 validation logic shall reject any transaction attempting to increase total supply.
* AC-TOK-001.3: The genesis block and all five allocation pools shall be publicly visible on DAG Explorer from the moment of mainnet launch.
* AC-TOK-001.4: After Phase 5 burns begin, total circulating supply shall decrease over time. The 1B genesis supply is a ceiling, not a floor.

### REQ-TOK-002: Community Reward Emission Budget

**User Story:** As a user, I want to earn ECHO tokens for every message I send with no daily limit, so that I am always incentivized to communicate — while knowing the total annual budget keeps the economy sustainable.

**Acceptance Criteria:**

* AC-TOK-002.1: The 400M community reward pool shall emit over 10 years per a declining annual budget: Year 1 = 80M (20%), Year 2 = 64M (16%), Year 3 = 52M (13%), Year 4 = 44M (11%), Year 5 = 36M (9%), Years 6–10 = 24M each (6%).
* AC-TOK-002.2: There shall be no per-user daily earning cap. Every message a user sends earns a reward regardless of how many messages they have already sent that day.
* AC-TOK-002.3: The per-message reward rate shall auto-scale based on total daily network activity. The actual rate = Daily Budget ÷ Total Daily Network Activity Weight, where each message contributes 1 × the sender's trust tier multiplier. As the network grows, the per-message rate declines — but every message always earns something.
* AC-TOK-002.4: The daily budget shall equal Annual Emission ÷ 365 (Year 1 ≈ 219,178 ECHO/day). Unused daily budget from low-activity days rolls forward within the same calendar year.
* AC-TOK-002.5: After Year 10, no new ECHO shall be emitted. Staking APY from Year 11 onward is funded from transaction fees and platform revenue — not new emission.
* AC-TOK-002.6: The current year's emission budget, total distributed year-to-date, current auto-scaled per-message rate, and remaining pool balance shall be publicly queryable via DAG Explorer and the ECHO backend API in real time.

### REQ-TOK-003: Per-Action Reward Rates

**User Story:** As a user, I want to know exactly how many ECHO tokens each of my actions earns, so that I can understand the reward model and verify it is applied correctly.

**Acceptance Criteria:**

* AC-TOK-003.1: **Messaging**: Rate = auto-scaled daily rate × trust tier multiplier (Tier 1: 1.0x, Tier 2: 1.2x, Tier 3: 1.5x, Tier 4: 2.0x, Tier 5: 3.0x). The 0.1 ECHO/message figure is the target rate when network activity exactly matches the daily budget; actual rate scales up when activity is low and down when activity is high. There is no per-user cap — every message always earns.
* AC-TOK-003.2: **Referrals**: 50 ECHO each to referrer and new user when the new user completes DID-verified identity and sends their first 100 messages. Referral rewards are fixed payments exempt from auto-scaling, drawn directly from the community pool. Capped at 3 referral tiers to prevent pyramid gaming.
* AC-TOK-003.3: **Payment Rails**: 1–5 ECHO per payment transaction based on transaction value and verification level. Tier 5 × Tier 5 transactions earn the maximum rate.
* AC-TOK-003.4: **Staking APY**: Bronze 5% (30d), Silver 8% (90d), Gold 12% (180d), Platinum 15% (365d) annually on staked amount, distributed continuously and claimable via AtomicAction.
* AC-TOK-003.5: All messaging reward claims shall be AtomicActions that simultaneously verify the trust tier, apply the correct multiplier, record the claim against the network daily total, and update the auto-scale rate — preventing any partial-state gaming.

### REQ-TOK-004: Founder Allocation and Vesting

**User Story:** As a community member, I want founder token allocations locked on-chain with verified vesting, so that I can confirm founders cannot dump tokens and can hold them accountable to the same transparency ECHO promises users.

**Acceptance Criteria:**

* AC-TOK-004.1: At genesis, the system shall create five founder TokenLock positions: Founder 1 (CEO/Visionary/Product) 100M ECHO (10% supply); Founders 2–5 (co-founders) 20M ECHO each (2% each).
* AC-TOK-004.2: All founder TokenLocks shall enforce a 12-month cliff — no tokens are withdrawable before the cliff date regardless of any other condition.
* AC-TOK-004.3: After the cliff, each founder's remaining allocation vests at 1/36th per month over 36 months (48-month total vesting period from genesis).
* AC-TOK-004.4: Vested tokens are subject to a 14-day WithdrawLock cooldown before becoming transferable.
* AC-TOK-004.5: Pre-cliff departure: the entire TokenLock balance is returned to the Future Team pool via 3-of-5 founder multi-sig revocation.
* AC-TOK-004.6: Post-cliff departure: vested tokens are released; unvested balance is returned to the Future Team pool via multi-sig.
* AC-TOK-004.7: All founder TokenLock positions (allocated, cliff date, vested, locked, monthly vest, all WithdrawLock transactions) shall be publicly visible on DAG Explorer from genesis.
* AC-TOK-004.8: The ECHO Wallet shall display a founder vesting panel (visible only to the DID holding a founder TokenLock) showing: allocated, vested, locked, next unlock date, cliff status, and a "View on DAG Explorer" link.
* AC-TOK-004.9: DAO transition acceleration: 50% of unvested founder tokens accelerate when ECHO transitions to full DAO governance (Phase 5–6), triggered by governance vote in L1 code.

### REQ-TOK-005: Treasury Allocation and Controls

**User Story:** As a community member, I want the treasury allocation clearly defined with spend controls, so that I know funds cannot be misappropriated before DAO governance is operational.

**Acceptance Criteria:**

* AC-TOK-005.1: The 220M treasury at genesis shall be subdivided as: 80M to PacaSwap liquidity seeding, 50M to operational reserve (bridged to stablecoins), and 90M locked in a 3-of-5 founder multi-sig for Phase 5–6 operations.
* AC-TOK-005.2: During Phases 1–3, treasury disbursements require 3-of-5 founder multi-sig. From Phase 4 onward, disbursements require a governance vote.
* AC-TOK-005.2b: Future Team & Advisors pool disbursements (allocating tokens to new hires, advisors, or contractors) shall require Governance Board approval. The CEO may propose allocations but cannot unilaterally approve them. Unallocated tokens after 3 years from genesis revert to treasury via governance vote.
* AC-TOK-005.3: The treasury multi-sig address and all disbursement transactions shall be publicly visible on DAG Explorer.
* AC-TOK-005.4: From Phase 5, 30% of annual treasury surplus from VIP, Organization, and payment rail revenue shall be used by the AI Burn Agent to buy back and permanently destroy ECHO via PacaSwap.

### REQ-TOK-006: Token Launch Sequence

**User Story:** As an early user or ecosystem participant, I want to understand the token launch sequence so that I know when ECHO becomes tradeable, how liquidity is seeded, and how to participate from day one.

**Acceptance Criteria:**

* AC-TOK-006.1: **Phase 1 (Pre-Launch):** No ECHO tokens exist. No presale, no private round, no VC allocation, and no community token sale of any kind. ECHO shall never be sold before it is earned or traded on the open market. Community awareness is built through waitlist, beta signup, and Constellation ecosystem participation only. If pre-launch capital is needed, the sources are founder capital and Constellation ecosystem grants — never token sales.
* AC-TOK-006.2: **Phase 2 (Genesis):** 1B ECHO minted. Founder TokenLocks created. Community reward emission begins. ECHO Wallet tab goes live in iOS app so alpha users immediately see their accumulated rewards.
* AC-TOK-006.3: **Phase 2 (DEX Launch):** Treasury seeds ECHO/DAG and ECHO/USDC liquidity pools on PacaSwap within 7 days of mainnet launch — the first moment ECHO is tradeable.
* AC-TOK-006.4: **Phase 2 (First Holders):** The 100–500 alpha beta users receive their accumulated messaging rewards at genesis, creating the first authentic ECHO holders — people who earned tokens through product usage, not purchase.
* AC-TOK-006.5: **Phase 3 (DAG Delegation Campaign):** Community is invited to delegate DAG to ECHO validators in exchange for ECHO token incentives from the Ecosystem pool, bootstrapping validator decentralization and liquidity.
* AC-TOK-006.6: **Phase 3 (Base Bridge):** ECHO becomes bridgeable to Base via the 3A DAO bridge, opening Aerodrome liquidity and broader on-ramp paths.
* AC-TOK-006.7: **Phase 4 (CEX Listing):** ECHO bridges to Ink (Kraken L2) to pursue a Kraken listing, expanding to a mainstream trading audience.
* AC-TOK-006.8: ECHO shall not conduct a presale, private round, or VC allocation at any phase. Early access to ECHO is earned through product usage and ecosystem participation, not financial investment.

### REQ-TOK-007: Single-Token Governance Model

**User Story:** As a token holder, I want ECHO to serve as the only governance token, and whale attacks prevented through trust-tier weighting, so that community participation — not capital concentration — determines governance outcomes.

**Acceptance Criteria:**

* AC-TOK-007.1: ECHO shall be the sole token for all utility (rewards, staking, payments) and all governance. No separate governance token shall ever be created.
* AC-TOK-007.2: Governance votes shall use the formula: Governance Weight = StakedECHO × TrustTierMultiplier (Tier 1 = 0.0, Tier 2 = 0.5, Tier 3 = 1.0, Tier 4 = 1.5, Tier 5 = 2.0).
* AC-TOK-007.3: Tier 1 (Unverified) users shall have zero governance weight regardless of token holdings. Governance participation requires Trust Tier 2 minimum.
* AC-TOK-007.4: Only staked (TokenLock) ECHO counts toward governance weight. Unstaked tokens confer no voting power.
* AC-TOK-007.5: Founder TokenLock positions shall be eligible for governance voting, giving founders participation from day one proportional to their staked allocation and trust tier.
* AC-TOK-007.6: Governance weight shall be calculated and enforced by Data L1 Scala validation — not the Go backend — ensuring it cannot be manipulated at the application layer.

## Feature Behavior and Rules

### No Caps: Why It Works and How Supply Stays Controlled

Removing per-user daily caps keeps the incentive to message alive every minute of the day. Caps create a frustrating cliff — users hit their limit, stop earning, and reduce engagement at exactly the wrong moment. Without caps, every message always earns something.

Supply is controlled by the annual emission budget through the auto-scaling mechanism. The per-message rate adjusts in real time based on total network activity. When the network is small and active, each message earns more than the 0.1 ECHO target. When the network is large and active, each message earns less. The annual pool is never exceeded — the math enforces it structurally:

| Scenario | Daily Budget | Network Msgs/Day | Auto-Scaled Rate (Tier 3) | User Earnings/Day (50 msgs) |
| --- | --- | --- | --- | --- |
| Year 1, 10K users | 219K ECHO | 500K | \~0.66 ECHO | \~32.8 ECHO |
| Year 1, 100K users | 219K ECHO | 5M | \~0.066 ECHO | \~3.3 ECHO |
| Year 1, 1M users | 219K ECHO | 50M | \~0.0066 ECHO | \~0.33 ECHO |
| Year 3, 1M users | 142K ECHO | 50M | \~0.0043 ECHO | \~0.21 ECHO |

This creates a powerful early-adopter effect: the earlier you join, the more each message is worth. Early users earn significantly more per message than later joiners — a natural reward for building the network. Every message earns something regardless of network size. The annual pool is always fully distributed.

### No Presale: Why and the Commitment

ECHO will not conduct a presale, private round, community token sale, or VC allocation at any phase. This is absolute.

A community presale sounds fair but creates the same problem at smaller scale: early buyers get tokens at a discount, establishing a class of holders with financial exposure rather than earned ownership. The moment you sell tokens before the product exists, you attract speculators, not users.

ECHO's model is cleaner: the first ECHO holders are alpha users who earned tokens by using the product. First price discovery happens on PacaSwap at mainnet launch with treasury-seeded liquidity. Anyone who wants ECHO after launch can buy it on PacaSwap or earn it by messaging. No early access. No discount tier.

If pre-launch capital is needed, the sources are founder capital and Constellation ecosystem grants. These preserve the "no early investors" story without creating a two-tier holder structure.

### The Blockchain Is the Cap Table

All founder vesting, treasury balances, emission distributions, and token holdings are on-chain and publicly verifiable on DAG Explorer. There is no private cap table, no off-chain vesting agreement that can be altered, and no backdoor token releases. Any user, journalist, investor, or regulator can verify the exact token distribution at any moment.

### Founder Allocation Rationale

The CEO's 10% reflects the totality of pre-team contributions: full architecture, 5+ PRD versions, backend/iOS/API architecture documents, tokenomics model, governance structure, and all strategic decisions produced before any co-founder joined. The co-founder 2% equal split provides a clean, competitive offer that avoids internal allocation disputes. The insider total (founders 18% + future team 10% = 28%) stays below the industry average of 35–45% inclusive of VC allocation. Community + ecosystem retains 50% — the majority.

### REQ-TOK-001: Fixed Supply and Genesis Allocation

**User Story:** As a token holder, I want the total ECHO supply fixed at genesis and publicly verifiable, so that I can trust no additional tokens will be minted to dilute my holdings.

**Acceptance Criteria:**

* AC-TOK-001.1: At Phase 2 mainnet launch, the Currency L1 Scala genesis block shall mint exactly 1,000,000,000 ECHO tokens — no more, no fewer — and allocate them to the five protocol-controlled pools defined below.
* AC-TOK-001.2: The genesis allocation shall be: Community Rewards 40% (400M), Treasury 22% (220M), Founders 18% (180M), Future Team & Advisors 10% (100M), Ecosystem & Partnerships 10% (100M).
* AC-TOK-001.3: No additional minting shall be possible after genesis. The Currency L1 validation logic shall reject any transaction attempting to increase total supply.
* AC-TOK-001.4: The genesis block and all five allocation pools shall be publicly visible on DAG Explorer from the moment of mainnet launch.
* AC-TOK-001.5: After Phase 5 burns begin, total circulating supply shall decrease over time; the genesis supply of 1B is a ceiling, not a floor.

### REQ-TOK-002: Community Reward Emission Rate

**User Story:** As a user, I want to understand exactly how many ECHO tokens I can earn and how the emission rate changes over time, so that I can plan my participation and understand my earning potential.

**Acceptance Criteria:**

* AC-TOK-002.1: The 400M community reward pool shall be emitted over 10 years per a declining curve: Year 1 = 80M (20%), Year 2 = 64M (16%), Year 3 = 52M (13%), Year 4 = 44M (11%), Year 5 = 36M (9%), Years 6-10 = 24M each (6% each).
* AC-TOK-002.2: After Year 10, no new ECHO shall be emitted. Staking APY from Year 11 onward shall be funded exclusively from transaction fees, AllowSpend fees, and VIP subscription revenue — not new emission.
* AC-TOK-002.3: The annual emission cap shall be enforced in Currency L1 Scala validation logic. Validators shall reject reward claims that would cause total Year-N distributions to exceed the Year-N emission cap.
* AC-TOK-002.5: The current year's emission rate, total distributed-to-date, and remaining emission pool balance shall be publicly queryable via DAG Explorer and the ECHO backend API.

### REQ-TOK-003: Per-Action Reward Rates

**User Story:** As a user, I want to know exactly how many ECHO tokens each of my actions earns, so that I can understand the reward model and trust it is being applied correctly.

**Acceptance Criteria:**

* AC-TOK-003.1: **Messaging rewards** shall pay 0.1 ECHO per message sent or received, multiplied by the sender's trust tier multiplier (Tier 1: 1.0x, Tier 2: 1.2x, Tier 3: 1.5x, Tier 4: 2.0x, Tier 5: 3.0x), subject to the daily cap for the user's trust tier.
* AC-TOK-003.2: **Referral rewards** shall pay 50 ECHO each to both the referrer and the new user when the new user: (a) completes DID-verified identity, and (b) sends their first 100 messages. Multi-level referral bonuses shall be capped at 3 tiers to prevent pyramid gaming.
* AC-TOK-003.3: **Payment rail rewards** shall pay 1–5 ECHO per payment transaction based on transaction value and verification level of the participants. Tier 5 × Tier 5 transactions earn the maximum rate.
* AC-TOK-003.4: **Staking APY** shall pay 5% (Bronze/30d), 8% (Silver/90d), 12% (Gold/180d), or 15% (Platinum/365d) annually on the staked amount, distributed continuously and claimable via AtomicAction.
* AC-TOK-003.5: All reward claims shall be bundled as AtomicActions that simultaneously: verify the user's current trust tier on-chain, apply the correct multiplier, record the claim against the daily cap, and update the cap counter — preventing any partial-state gaming.

### REQ-TOK-004: Founder Allocation and Vesting

**User Story:** As a community member, I want founder token allocations locked on-chain with verified vesting schedules, so that I can confirm founders cannot dump tokens and I can hold them accountable to the same blockchain transparency ECHO promises users.

**Acceptance Criteria:**

* AC-TOK-004.1: At genesis, the system shall create five founder TokenLock positions with the following allocations: Founder 1 (CEO/Visionary/Product) 100M ECHO (10% supply), Founders 2–5 (co-founders) 20M ECHO each (2% supply each).
* AC-TOK-004.2: All founder TokenLock positions shall enforce a 12-month cliff — no tokens shall be withdrawable before the cliff date, regardless of any other condition.
* AC-TOK-004.3: After the cliff, each founder's remaining allocation shall vest at 1/36th per month over 36 months (total vesting period: 48 months from genesis).
* AC-TOK-004.4: Vested tokens shall be subject to a 14-day WithdrawLock cooldown before becoming transferable.
* AC-TOK-004.5: If a founder departs before the cliff, their entire TokenLock balance shall be returned to the Future Team pool via a 3-of-5 founder multi-sig revocation transaction.
* AC-TOK-004.6: If a founder departs after the cliff, their vested tokens shall be released and their unvested balance returned to the Future Team pool via the same multi-sig mechanism.
* AC-TOK-004.7: All five founder TokenLock positions — allocated amount, cliff date, vested amount, locked amount, monthly vest amount, and all WithdrawLock transactions — shall be publicly visible on DAG Explorer from genesis. This transparency is non-negotiable.
* AC-TOK-004.8: The ECHO Wallet shall display a founder vesting panel (visible only to the DID holding a founder TokenLock) showing: allocated, vested, locked, next unlock date, cliff status, and a "View on DAG Explorer" link.
* AC-TOK-004.9: Governance acceleration shall apply: 50% of unvested founder tokens accelerate when ECHO formally transitions to full DAO governance (Phase 5–6), triggered by a governance vote in L1 code.

### REQ-TOK-005: Treasury Allocation and Controls

**User Story:** As a community member, I want the treasury allocation clearly defined with spend controls, so that funds are not misappropriated before DAO governance is operational.

**Acceptance Criteria:**

* AC-TOK-005.1: The 220M treasury allocation at genesis shall be subdivided as: 80M to PacaSwap liquidity seeding (ECHO/DAG and ECHO/USDC pools), 50M to operational reserve (bridged to stablecoins), and 90M locked in a 3-of-5 founder multi-sig for Phase 5–6 operations.
* AC-TOK-005.2: During Phases 1–3, treasury disbursements shall require 3-of-5 founder multi-sig authorization. From Phase 4 onward, treasury disbursements shall require a governance vote.
* AC-TOK-005.3: The treasury multi-sig address and all disbursement transactions shall be publicly visible on DAG Explorer.
* AC-TOK-005.4: Starting Phase 5, 30% of annual treasury surplus from revenue shall be used by the AI Burn Agent to buy back and permanently destroy ECHO tokens via PacaSwap, reducing total circulating supply.

### REQ-TOK-006: Token Launch Sequence

**User Story:** As an investor or early user, I want to understand the token launch sequence so that I know when ECHO becomes tradeable, how liquidity is established, and how I can participate from day one.

**Acceptance Criteria:**

* AC-TOK-006.1: **Phase 1 (Pre-Launch):** No ECHO tokens exist. Community awareness is built through waitlist, beta signup, and Constellation ecosystem partnerships. No presale, no private round.
* AC-TOK-006.2: **Phase 2 (Genesis Launch):** Token genesis mints 1B ECHO. Founder TokenLocks are created. Community reward emission begins. The ECHO Wallet tab goes live in the iOS app, allowing users to see their earned rewards accumulate.
* AC-TOK-006.3: **Phase 2 (PacaSwap Liquidity):** The treasury seeds ECHO/DAG and ECHO/USDC liquidity pools on PacaSwap within 7 days of mainnet launch. This is the first moment ECHO becomes tradeable on a DEX.
* AC-TOK-006.4: **Phase 2 (Alpha Rewards):** The 100-500 alpha beta users from Phase 1 receive their accumulated messaging rewards as the first ECHO distributions at genesis, creating authentic early holders.
* AC-TOK-006.5: **Phase 3 (DAG Delegation Campaign):** Community is invited to delegate DAG to ECHO validators, earning ECHO token incentives from the Ecosystem pool in return. This bootstraps validator decentralization and increases liquidity depth.
* AC-TOK-006.6: **Phase 3 (Base Bridge):** ECHO becomes bridgeable to Base via the 3A DAO bridge, enabling access to Aerodrome liquidity and CEX on-ramp paths for a broader audience.
* AC-TOK-006.7: **Phase 4 (CEX Listing):** ECHO bridges to Ink (Kraken L2) to pursue a Kraken exchange listing, expanding ECHO to a mainstream trading audience.
* AC-TOK-006.8: The system shall not conduct a presale, private round, or VC allocation. Early access to ECHO shall be through product usage (messaging rewards) and ecosystem participation (DAG delegation), not financial investment.

### REQ-TOK-007: Single-Token Governance

**User Story:** As a token holder, I want ECHO to be the only token used for both utility and governance, and I want whale attacks prevented through trust-tier weighting rather than a separate governance token.

**Acceptance Criteria:**

* AC-TOK-007.1: ECHO shall be the sole token for all utility (rewards, staking, payments) and all governance (voting, board elections, treasury allocation). No separate governance token shall be created.
* AC-TOK-007.2: Governance votes shall be weighted by the formula: `Governance Weight = StakedECHO × TrustTierMultiplier`, where multipliers are: Tier 1 = 0.0, Tier 2 = 0.5, Tier 3 = 1.0, Tier 4 = 1.5, Tier 5 = 2.0.
* AC-TOK-007.3: Tier 1 (Unverified) users shall have zero governance weight regardless of their token holdings. Participation in governance shall require at minimum Trust Tier 2.
* AC-TOK-007.4: Staking (TokenLock) shall be required to vote. Unstaked tokens confer no governance weight, incentivizing long-term commitment over short-term speculation.
* AC-TOK-007.5: Founder TokenLock vesting positions shall be eligible for governance voting, giving founders governance participation from day one proportional to their staked allocation and trust tier.
* AC-TOK-007.6: The governance weight formula shall be calculated and enforced by Data L1 Scala validation logic, not by the Go backend — ensuring it cannot be manipulated at the application layer.

## Feature Behavior and Rules

### No Presale, No VCs

ECHO is explicitly designed to not require venture capital. The launch sequence generates token holders through product usage (messaging rewards from alpha), ecosystem participation (DAG delegation incentives), and open market trading (PacaSwap from day one of mainnet). There is no presale, no private round, no SAFT agreements, and no VC allocation. Early access to ECHO is earned, not bought.

This is a strategic choice: VC funding creates misaligned incentives (investors expect returns, which means extracting value from users). ECHO's model keeps all value in the community. If pre-launch capital is needed, it comes from founder capital or Constellation ecosystem grants — not from investors who would expect equity or governance control.

### Rate of Issuance vs. Anti-Inflation Controls

At 100K daily active users, each sending 50 messages/day at 0.1 ECHO with a Tier 2 multiplier (1.2x), the raw daily issuance is approximately: 100,000 users × 50 messages × 0.1 ECHO × 1.2 = 600,000 ECHO/day = 219M ECHO/year. This would exceed the Year 1 emission cap of 80M.

The daily cap system resolves this: with Tier 2 users capped at 25 ECHO/day, 100K active users hit at most 2.5M ECHO/day, or 912M ECHO/year — still too high. This means at 100K users, most users will hit their daily cap early in the day, and the emission rate will be controlled by the per-user daily cap rather than raw activity volume. The Year 1 cap of 80M effectively limits the average active user to \~800 ECHO/year from messaging alone at 100K users. As the user base grows, the per-user daily cap becomes the binding constraint, not the annual pool.

### Founder Allocation Rationale

The CEO's 10% of total supply is high by typical startup standards (3-5%) but is justified by the totality of pre-team work: full product architecture, 5 versions of the PRD, backend/iOS/API architecture documents, tokenomics design, governance model, and all strategic decisions. The co-founder 2% equal split is a clean, competitive, and fair offer that avoids internal politics. The insider total (founders 18% + future team 10% = 28%) remains below the industry average of 35-45% that typically includes VC allocation. Community + ecosystem retains 50% — the majority.

### The Blockchain Is the Cap Table

All founder vesting, treasury balances, and token distributions are on-chain and publicly verifiable. ECHO has no private cap table spreadsheet, no off-chain vesting agreements that can be altered, and no backdoor token releases. Any ECHO user, journalist, investor, or regulator can verify the exact token distribution at any moment by querying DAG Explorer. This transparency is a product feature, not a legal obligation.

## Production Launch, Infrastructure, and Deployment

## Overview

This document defines the requirements and step-by-step process for launching ECHO to production. It covers the four pillars of a successful production launch: User Acceptance Testing (UAT) across all system layers, cloud infrastructure setup for the Go relay backend, Apple App Store submission and approval, and the CI/CD pipeline that automates deployment across all environments.

ECHO's production launch is more complex than a typical iOS app because it involves three independent technology layers that must all be live and verified before users can be onboarded: the Constellation metagraph on public Hypergraph mainnet, the Cardano identity layer, and the Go relay backend on cloud infrastructure. These layers must launch in sequence, not simultaneously. This document defines that sequence, the tests that gate each stage, and the infrastructure that supports it.

## Terminology

* **UAT (User Acceptance Testing)**: Structured testing by the team and trusted beta users that validates all system layers behave correctly under real conditions before public launch.
* **Staging Environment**: A production-identical environment used for final integration testing. All infrastructure, blockchain connections, and API keys must match production exactly.
* **TestFlight**: Apple's platform for distributing pre-release iOS apps to beta testers. Required step before App Store submission.
* **APNs (Apple Push Notification Service)**: Apple's infrastructure for delivering push notifications to iOS devices. Requires a server-side certificate and endpoint.
* **Hypergraph Mainnet**: The Constellation public network where the ECHO metagraph runs in production. Distinct from testnet — real DAG staking required.
* **Euclid SDK**: The Scala/JVM framework for building Constellation metagraph L1 validation logic. Produces the JAR files deployed to validator nodes.
* **CI/CD (Continuous Integration/Continuous Deployment)**: Automated pipelines that build, test, and deploy code changes across all environments without manual intervention.
* **IaC (Infrastructure as Code)**: Cloud infrastructure defined in version-controlled configuration files (Terraform, Pulumi) rather than configured manually through a UI.
* **Blue/Green Deployment**: A release strategy that runs two identical production environments (Blue = current live, Green = new version). Traffic switches to Green after validation; Blue remains on standby for instant rollback.

## Requirements

### REQ-PROD-001: User Acceptance Testing (UAT)

**User Story:** As a launch team member, I want a structured UAT checklist covering every system layer, so that we can verify the complete system is production-ready before any public user touches it.

**Acceptance Criteria:**

**iOS App Testing:**

* AC-PROD-001.1: The iOS app shall pass testing on a minimum device matrix of: iPhone 14 (iOS 17), iPhone 15 Pro (iOS 17), iPhone 16 (iOS 18), iPad Pro (latest iOS). Both physical devices and simulators must be tested; Secure Enclave features require physical devices only.
* AC-PROD-001.2: All biometric flows shall be tested on physical devices: Face ID key generation, Face ID signing for DID operations, Face ID wallet transactions, Touch ID fallback. Simulator testing is not sufficient for Secure Enclave validation.
* AC-PROD-001.3: The following user journeys shall be executed end-to-end on physical devices before TestFlight distribution: new user onboarding (DID creation → wallet setup → first message), send and receive message (verify E2EE), stake ECHO tokens (TokenLock), claim messaging reward (AtomicAction), referral flow (invite + reward distribution), and disappearing message with cryptographic deletion.
* AC-PROD-001.4: The app shall be tested under degraded network conditions: 3G simulation, intermittent connectivity, relay server unreachable (verify graceful offline mode), and push notification delivery after backgrounded state.
* AC-PROD-001.5: All App Store Review Guidelines shall be verified before TestFlight upload, specifically: no private API usage, proper permission strings for camera/microphone/biometrics, no references to other payment systems (Apple IAP compliance), privacy manifest included.

**Go Backend Testing:**

* AC-PROD-001.6: Load testing shall simulate 10x expected Day 1 traffic (target: 10,000 concurrent WebSocket connections) using a tool such as k6 or Locust. The relay service shall maintain <500ms p99 message latency under this load.
* AC-PROD-001.7: Integration tests shall cover all relay API endpoints: message send/receive, offline queue delivery, push notification trigger, DID resolution proxy, metagraph reward submission, and health check endpoint.
* AC-PROD-001.8: The backend shall be tested for graceful degradation: PostgreSQL unavailable (serve from Redis cache), Redis unavailable (degrade to DB-only mode), metagraph unreachable (queue reward submissions for retry), and APNs unreachable (queue notifications with exponential backoff).
* AC-PROD-001.9: Security testing shall verify: all endpoints require valid DID-signed authentication, no unauthenticated endpoints expose user data, message payloads are opaque ciphertext (backend cannot read content), rate limiting is enforced per DID, and SQL injection/SSRF protections are active.
* AC-PROD-001.10: The backend API contract (OpenAPI spec) shall be validated against the iOS app's API client. All request/response schemas shall match exactly before production deployment.

**Blockchain Testing (Constellation Metagraph):**

* AC-PROD-001.11: The metagraph shall complete a full testnet lifecycle before mainnet deployment: genesis block creation with all 5 allocation pools, 3 L0 hybrid nodes operational, Currency L1 and Data L1 validators processing transactions, TokenLock creation and cliff enforcement, AtomicAction reward claim validation, and snapshot submission to Hypergraph testnet.
* AC-PROD-001.12: The following metagraph transactions shall be tested on testnet with real Scala L1 validation: token genesis (1B ECHO mint), founder TokenLock creation (5 positions), messaging reward claim (AtomicAction with tier verification), staking stake + withdrawal (TokenLock + WithdrawLock with 14-day enforcement), Merkle root submission (message integrity anchor), and governance vote weight calculation.
* AC-PROD-001.13: The 750K DAG staking requirement for 3 L0 nodes shall be confirmed in the staging/mainnet wallet before deployment. Mainnet deployment shall be blocked if DAG balance is insufficient.
* AC-PROD-001.14: The Currency L1 emission cap enforcement shall be tested by attempting to submit reward claims that would exceed the Year 1 daily budget — the validator shall reject these transactions with a specific error code.

**Blockchain Testing (Cardano):**

* AC-PROD-001.15: DID registration, credential issuance, and trust tier commitment transactions shall be tested on Cardano preprod testnet before mainnet deployment.
* AC-PROD-001.16: The estimated ADA transaction cost per operation (DID registration \~0.2 ADA, credential issuance \~0.3 ADA) shall be measured on testnet and confirmed against the treasury budget. Monthly ADA cost at 10K users shall be projected and a funded treasury wallet prepared.

**API Keys and Secrets Validation:**

* AC-PROD-001.17: Before production deployment, all of the following shall be provisioned, tested end-to-end, and stored in the secrets manager: APNs production certificate (distinct from sandbox), Stargazer SDK API credentials, IPFS/Storj storage API key, IDV provider API key (Stripe Identity or Sumsub), Constellation metagraph REST API endpoint, Cardano node API endpoint, PacaSwap contract addresses (mainnet), and Base/Ink bridge contract addresses.
* AC-PROD-001.18: Each API key shall be tested via a dedicated integration test that calls the live endpoint with a minimal valid request and asserts a 2xx response. No API key shall be deployed to production untested.
* AC-PROD-001.19: API keys shall never be stored in source code or environment variable files committed to the repository. All secrets shall be stored in HashiCorp Vault (self-hosted on Hetzner) and injected at runtime via Kubernetes secrets.

**UAT Sign-off Criteria:**

* AC-PROD-001.20: Production launch is gated by sign-off from: iOS engineer (app testing complete), backend engineer (load + integration tests passing), blockchain engineer (metagraph testnet lifecycle complete), and security review (no critical or high vulnerabilities open). All four sign-offs must be recorded before the production deployment checklist is initiated.

### REQ-PROD-002: Cloud Infrastructure Setup

**User Story:** As the engineering team, I want the Go relay backend deployed on scalable, cost-effective cloud infrastructure with full observability, so that ECHO can serve users reliably from day one and scale to 1M+ users without re-architecting.

**Acceptance Criteria:**

**Cloud Provider and Region:**

* AC-PROD-002.1: The initial production deployment shall use Hetzner Cloud (Falkenstein, Germany) as the primary region. German jurisdiction provides the strongest EU privacy protection (GDPR + BDSG). A secondary region (Hetzner Helsinki, Finland) shall be configured for failover from Phase 3 onward. OVHcloud (France) serves as a third provider for multi-cloud resilience.
* AC-PROD-002.2: All infrastructure shall be defined as Infrastructure as Code using Terraform (Hetzner provider) or Pulumi. No production resources shall be created manually through the Hetzner console. The Terraform state shall be stored in an encrypted Hetzner Object Storage bucket with state locking.

**Compute (Go Relay Backend):**

* AC-PROD-002.3: The Go relay service shall run on k3s (lightweight Kubernetes) on Hetzner Cloud or Hetzner dedicated servers. k3s is recommended for all phases — it provides full Kubernetes API compatibility with lower overhead than managed K8s services. For Phase 3+, Hetzner Managed Kubernetes is also acceptable.
* AC-PROD-002.4: Initial production sizing shall be: 2 tasks × (2 vCPU, 4GB RAM) for the relay service. Auto-scaling shall be configured to add tasks when CPU &gt; 70% or WebSocket connection count &gt; 5,000 per task. Maximum auto-scale to 10 tasks before manual review is required.
* AC-PROD-002.5: The WebSocket relay shall be fronted by a Hetzner Cloud Load Balancer with WebSocket support enabled, or an nginx/HAProxy ingress controller on k3s. Sticky sessions shall NOT be used — the relay is stateless and any pod can serve any connection.

**Estimated Monthly Cloud Costs:**

* AC-PROD-002.6: The team shall budget for the following estimated monthly costs at launch (100K users):

| Service | Specification | Est. Monthly Cost |
| --- | --- | --- |
| k3s relay pods (Hetzner) | 2 pods × 2vCPU/4GB on Hetzner Cloud | \~$30 |
| PostgreSQL | Self-managed on Hetzner (Phase 1-2) | \~$15 |
| Redis | Self-managed on Hetzner with AOF persistence | \~$10 |
| Hetzner Load Balancer | 1 LB + nginx ingress | \~$7 |
| S3 (media/audit logs) | 100GB + data transfer | \~$15 |
| Prometheus + Grafana + Loki | Self-hosted on Hetzner | \~$15 |
| ACM SSL certificates | Free | $0 |
| Route 53 (DNS) | Hosted zone + queries | \~$5 |
| HashiCorp Vault | Self-hosted on Hetzner | \~$5 |
| **Phase 1-2 Total** | **\~$280/month** |  |

At 1M users (Phase 3+), estimated costs rise to $1,500–3,000/month depending on traffic patterns. This is covered by VIP subscription revenue well before reaching that scale.

**Database Setup:**

* AC-PROD-002.7: PostgreSQL shall be deployed on Hetzner dedicated servers with: automated daily backups to Hetzner Object Storage with 7-day retention, encryption at rest (LUKS + AES-256), private network isolation (Hetzner vSwitch — no public internet access), synchronous streaming replica from Phase 3 onward. The database stores only non-sensitive relay metadata — all PII is on-device per the privacy architecture.
* AC-PROD-002.8: Redis shall be deployed on Hetzner (self-managed with AOF persistence) for: WebSocket connection state, message delivery queue (encrypted offline message blobs), session token cache, and reward claim rate limiting. Redis data is ephemeral — loss of Redis does not cause data loss; the system falls back to PostgreSQL.

**Networking and Security:**

* AC-PROD-002.9: All backend services shall run within a Hetzner private network (vSwitch). Only the load balancer / nginx ingress shall be publicly accessible. The k3s pods, PostgreSQL, and Redis shall have no public IP addresses.
* AC-PROD-002.10: Hetzner firewall rules shall enforce: ingress load balancer accepts only 443 (HTTPS/WSS); k3s pods accept only traffic from the ingress controller; PostgreSQL accepts only traffic from backend pods; Redis accepts only traffic from backend pods.
* AC-PROD-002.11: A WAF layer (ModSecurity with OWASP Core Rule Set on nginx ingress, or Hetzner DDoS protection) shall be configured with rules for: rate limiting (100 req/sec per IP), common attack patterns (OWASP Top 10), and geographic blocking if required by compliance.

**Observability:**

* AC-PROD-002.12: The following dashboards shall be live before launch: relay service (request rate, WebSocket connections, p50/p95/p99 latency, error rate), database (connections, query time, disk usage), Redis (memory usage, hit rate, eviction rate), and metagraph submission queue (pending, success rate, retry rate).
* AC-PROD-002.13: PagerDuty or equivalent alerting shall be configured for: relay error rate &gt;1% (page immediately), WebSocket connection count &gt;80% of capacity (warn), p99 latency &gt;2s (warn), PostgreSQL disk usage &gt;80% (warn), and any 5xx error spike &gt;10 in 1 minute (page immediately).
* AC-PROD-002.14: All application logs shall be shipped to Grafana Loki (self-hosted on Hetzner) with a 30-day retention policy. Sensitive data (DIDs, message hashes) shall appear in logs; plaintext message content shall never appear in any log.

**Admin Access:**

* AC-PROD-002.15: HashiCorp Vault shall be configured with least-privilege policies: a deploy policy (used by CI/CD, can update k3s deployments and push images), a read-only policy (for monitoring/debugging), and a break-glass admin policy (MFA-required, full access, usage audited). Root Vault tokens shall be revoked after initial setup.
* AC-PROD-002.16: Hetzner Cloud shall be configured with separate projects for dev, staging, and production environments. Vault policies and k3s namespaces shall enforce environment isolation. CI/CD pipelines deploy to all three environments from a single GitHub Actions configuration.

### REQ-PROD-003: Apple App Store Submission and Launch

**User Story:** As the iOS engineer, I want a complete App Store submission checklist so that the app is approved on the first review attempt and launches without delay.

**Acceptance Criteria:**

**Pre-Submission Requirements:**

* AC-PROD-003.1: App Store Connect account shall be set up under a legal entity (LLC or corporation) rather than a personal account. The $99/year Apple Developer Program membership shall be active.
* AC-PROD-003.2: The following App Store Connect metadata shall be prepared before submission: app name ("ECHO"), subtitle (max 30 chars), description (max 4000 chars), keywords (max 100 chars), support URL, privacy policy URL (required — must be live on a publicly accessible URL), and copyright string.
* AC-PROD-003.3: App screenshots shall be prepared for required device sizes: iPhone 6.9" (iPhone 16 Pro Max), iPhone 6.5" (iPhone 14 Plus/15 Plus), and iPad Pro 12.9" (if iPad supported). Minimum 3 screenshots per size, maximum 10. Screenshots must show actual app UI — no mockups or marketing images as primary screenshots.
* AC-PROD-003.4: An App Preview video (optional but strongly recommended) shall demonstrate the core messaging flow: onboarding → first message → ECHO wallet. Max 30 seconds, must start with the actual app.
* AC-PROD-003.5: A Privacy Nutrition Label shall be completed in App Store Connect disclosing all data types collected: usage data (message count for reward calculation), identifiers (DID, device identifier), and diagnostic data. The label must accurately reflect what data is collected and linked to the user's identity.

**App Review Compliance (Critical Risk Areas):**

* AC-PROD-003.6: **Cryptocurrency/Wallet (Guideline 3.1.1)**: The ECHO Wallet allows users to manage ECHO tokens earned through the app. Apple requires that apps facilitating crypto transactions must comply with local laws and may not use Apple's IAP for digital currency purchases. The submission must clearly state in the review notes that: (a) ECHO tokens are earned through app usage, not purchased directly in-app, (b) any ECHO purchases happen via PacaSwap (external DEX), and (c) the app complies with applicable regulations.
* AC-PROD-003.7: **In-App Purchases (Guideline 3.1.1)**: VIP subscriptions ($9.99/month) must be implemented as Apple In-App Purchases (IAP) — not direct payment via ECHO tokens or credit card. The App Store takes 15-30% commission on IAP. Budget for this in the revenue model (net revenue = $7.00-8.50 per VIP subscriber per month after Apple's cut).
* AC-PROD-003.8: **Sign In with Apple (Guideline 4.8)**: If any third-party login is offered (Cardano wallet connect), Sign In with Apple must also be offered as an option. Alternatively, DID-only login (no social login at all) avoids this requirement.
* AC-PROD-003.9: **Biometric Authentication (Guideline 5.1.1)**: The Face ID usage string in Info.plist must accurately describe why Face ID is used: "ECHO uses Face ID to protect your private keys and authorize transactions." Vague strings like "for security" will cause rejection.
* AC-PROD-003.10: **Export Compliance**: The app uses encryption (E2EE with ChaCha20-Poly1305) which requires an Export Compliance declaration. Select "Yes, this app uses encryption" and "Yes, qualifies for exemption" (standard encryption algorithms, no custom crypto). Include an encryption exemption justification if asked.

**TestFlight Beta Process:**

* AC-PROD-003.11: Internal TestFlight testing (up to 100 Apple IDs in the developer account) shall run for a minimum of 2 weeks before external beta or App Store submission. All critical bugs found in UAT (REQ-PROD-001) must be resolved before TestFlight distribution.
* AC-PROD-003.12: External TestFlight testing (up to 10,000 testers, requires Beta App Review) shall run for a minimum of 2 weeks. Beta App Review typically takes 1-3 days. This is the closest simulation to the actual App Store review.
* AC-PROD-003.13: TestFlight crash rate shall be <0.5% of sessions before App Store submission. Apple monitors crash rates and may limit distribution for crash-prone apps.

**App Store Submission Process:**

* AC-PROD-003.14: The App Store submission shall follow this exact sequence:

  1. Archive the app in Xcode using the Production provisioning profile and Distribution certificate
  2. Upload to App Store Connect via Xcode Organizer (not Transporter for initial submission)
  3. Complete all metadata, screenshots, and privacy labels in App Store Connect
  4. Answer all compliance questions (encryption, content rights, advertising identifier)
  5. Write detailed App Review Notes explaining: the blockchain technology, why Face ID is required, what ECHO tokens are, and how VIP subscriptions work via IAP
  6. Submit for review — standard review time is 24-48 hours; expedited review available for critical issues
* AC-PROD-003.15: A dedicated App Store review account (test user DID) shall be created and provided in the App Review Notes. The review account shall have pre-populated data (sent messages, earned ECHO, active stake) so reviewers can evaluate all features without needing to complete the full onboarding flow.
* AC-PROD-003.16: If the app is rejected, the team shall respond within 24 hours. Common rejection reasons for crypto apps: missing IAP for subscriptions, unclear cryptocurrency compliance statement, Face ID string too vague. Responses to reviewers that are polite and specific (citing the exact guideline) have a higher approval rate on appeal.

**Launch Day Checklist:**

* AC-PROD-003.17: On approval, the release shall NOT be set to automatic. Manual release control shall be enabled so the backend infrastructure and metagraph can be verified live before users download the app.
* AC-PROD-003.18: The following shall be verified as live before releasing the app: relay backend health check returns 200, Constellation metagraph L0 nodes are synced and processing snapshots, Cardano mainnet DID registration is working, PacaSwap liquidity pools are seeded, and APNs push notification test delivers successfully.
* AC-PROD-003.19: The app release shall be staged using App Store phased release (1% → 2% → 5% → 10% → 20% → 50% → 100% over 7 days). This allows the team to detect infrastructure scaling issues before full traffic hits.

### REQ-PROD-004: CI/CD Pipeline

**User Story:** As the engineering team, I want fully automated CI/CD pipelines for all three codebases (iOS, Go backend, Scala metagraph), so that every code change is tested automatically and deployments to staging and production are repeatable, auditable, and rollback-capable.

**Acceptance Criteria:**

**Pipeline Architecture:**

* AC-PROD-004.1: GitHub Actions shall be the CI/CD platform (integrated with the existing GitHub repository). Separate workflow files shall exist for: iOS app, Go backend, Scala metagraph L1, and infrastructure (Terraform).
* AC-PROD-004.2: Three deployment environments shall be maintained: `dev` (auto-deploys on every merge to `main`), `staging` (auto-deploys on every merge to `release/*` branch; mirrors production exactly), and `production` (manual approval gate required after staging validation).
* AC-PROD-004.3: No code shall reach production without: all automated tests passing, a staging deployment succeeding, and a named engineer approving the production deployment in GitHub. Approvals shall be logged and auditable.

**iOS CI/CD Pipeline:**

* AC-PROD-004.4: The iOS CI/CD pipeline shall execute on every pull request: `xcodebuild test` (unit and integration tests), SwiftLint (code style), and a build verification that the app compiles without warnings or errors.
* AC-PROD-004.5: On merge to `main`, Xcode Cloud (Apple's native CI) or Fastlane shall automatically build and distribute a new TestFlight build to internal testers. The build number shall be auto-incremented on each CI run.
* AC-PROD-004.6: On merge to a `release/*` branch, the pipeline shall: run the full test suite, build a release-signed IPA, upload to TestFlight for external beta, and post a Slack notification with the build number and TestFlight link.
* AC-PROD-004.7: Production App Store releases shall be triggered manually via a GitHub release tag (e.g., `v1.0.0`). The pipeline shall upload the signed IPA to App Store Connect but hold for manual submission — a human must press "Submit for Review" in App Store Connect.

**Go Backend CI/CD Pipeline:**

* AC-PROD-004.8: The Go backend pipeline shall execute on every pull request: `go test ./...` (all unit and integration tests), `golangci-lint` (static analysis), and a Docker image build verification.
* AC-PROD-004.9: On merge to `main`, the pipeline shall: build a Docker image, tag it with the Git SHA, push to GitHub Container Registry (GHCR), and deploy to the `dev` k3s namespace via Argo CD automatically.
* AC-PROD-004.10: Staging deployments shall use a canary strategy via Argo Rollouts. The new pod set receives 10% of traffic for 5 minutes; if error rate remains <0.1%, traffic shifts to 100%. If error rate exceeds 0.1%, automatic rollback to the previous pod set.
* AC-PROD-004.11: Production deployments shall follow the same canary pattern with an additional 15-minute canary phase (10% traffic) requiring explicit approval to proceed to 100%. Rollback shall be achievable in under 60 seconds via Argo Rollouts revision rollback.
* AC-PROD-004.12: Database migrations shall run as a separate pipeline step before the new container version is deployed. Migrations must be backward compatible (the old code version must work with the new schema) to support zero-downtime Blue/Green deploys.

**Scala Metagraph CI/CD Pipeline:**

* AC-PROD-004.13: The Scala CI/CD pipeline shall execute on every pull request: `sbt test` (unit tests for all L1 validation logic), `scalafmt` (formatting), and a JAR build verification.
* AC-PROD-004.14: Metagraph deployments are higher-risk than backend deployments because L1 validation logic changes affect on-chain behavior. A staging metagraph (on Constellation testnet) shall receive every merge to `main`. Production metagraph updates shall require: all tests passing, 48 hours of staging validation, and 3-of-5 founder multi-sig approval (separate from GitHub approvals).
* AC-PROD-004.15: The metagraph deployment pipeline shall: build the JAR, copy to all validator nodes via SSH/SFTP (or S3 + node pull), restart the L1 validator services with a rolling restart (one node at a time to maintain consensus), and verify the new JAR is processing snapshots correctly before proceeding to the next node.
* AC-PROD-004.16: L1 validator node health shall be monitored continuously. If a node goes unhealthy during a rolling deploy, the pipeline shall halt and alert. Rollback restores the previous JAR from S3.

**Secrets and Configuration Management:**

* AC-PROD-004.17: All secrets (API keys, database passwords, private keys for deployment) shall be stored in HashiCorp Vault and referenced by path in GitHub Actions workflows via the Vault GitHub Action. No secrets shall appear in workflow YAML files, `.env` files committed to the repo, or CI logs.
* AC-PROD-004.18: Environment-specific configuration (relay endpoint URLs, PacaSwap contract addresses, Cardano network) shall be stored in Kubernetes ConfigMaps per namespace and injected into pods as environment variables at deploy time.
* AC-PROD-004.19: GitHub repository secrets shall store only: Hetzner API token, Vault token (for secret injection), Apple Developer team credentials (for Xcode Cloud/Fastlane), and Slack webhook URL for notifications. All application secrets live in HashiCorp Vault, not GitHub.

## Feature Behavior and Rules

### Launch Sequence: What Goes Live in What Order

The three technology layers must launch in this sequence. Deploying out of order creates dependencies that cannot be satisfied:

**Step 1 — Constellation Metagraph Mainnet (2 weeks before app launch)**

* Stake 750K DAG across 3 L0 hybrid nodes
* Deploy metagraph L1 validators (Currency + Data)
* Execute token genesis (1B ECHO minted, all 5 pools created)
* Create founder TokenLock positions (5 founders)
* Seed PacaSwap ECHO/DAG liquidity pool
* Verify all transactions on DAG Explorer
* Run 2 weeks of mainnet health monitoring before proceeding

**Step 2 — Cardano Mainnet (1 week before app launch)**

* Deploy ECHO DID registry schema to Cardano mainnet
* Fund platform treasury wallet with ADA (\~15,000 ADA/month estimated)
* Test DID registration end-to-end with a real user flow on mainnet
* Verify credential issuance and trust tier commitment transactions

**Step 3 — Go Backend Production (3 days before app launch)**

* Apply all database migrations
* Deploy relay service to production k3s via Argo Rollouts canary
* Verify WebSocket connections, APNs push delivery, and offline queue
* Run production load test at 1,000 concurrent connections
* Verify all API keys are live (APNs, Stargazer, IPFS/Storj, IDV)
* Enable monitoring and alerting

**Step 4 — App Store Launch**

* Submit to App Store review (expect 24-48 hours)
* Hold manual release after approval
* Verify all three backend layers are healthy
* Release via phased rollout (1% → 100% over 7 days)
* Monitor crash rate and error rate in real time

### Rollback Plans

Each layer has an independent rollback:

* **Go backend**: Argo Rollouts rollback — revert to previous pod revision in <60 seconds
* **Metagraph JAR**: Rolling node restart with previous JAR from S3 — \~10 minutes
* **iOS app**: Use App Store phased release pause (stops new downloads; existing users unaffected)
* **Blockchain state**: Cannot be rolled back (immutable). This is why testnet validation is mandatory before mainnet deployment. The only recovery for a bad blockchain state is a governance vote to patch the L1 logic forward.

### Cost Summary

| Environment | Monthly Cost | Notes |
| --- | --- | --- |
| Production cloud (Phase 1-2) | \~$280/month | Scales with users |
| Constellation L0 nodes (3) | \~$135–165/month | Hetzner dedicated servers (AX41-NVMe ~€45/month each) |
| Cardano ADA fees | \~$500/month at 10K users | \~0.3 ADA/credential issuance |
| Apple Developer Program | $99/year | One-time annual |
| Apple IAP Commission | 15–30% of VIP revenue | Net $7.00–8.50 per subscriber |
| DAG staking (750K DAG) | Capital lockup, not expense | Recoverable, nodes earn rewards |
| **Total Monthly (pre-revenue)** | **\~$1,100–1,280/month** | Drops after delegation subsidizes DAG fees |

