# Echo - Requirements

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
- [Privacy Architecture and Secure Data Handling](#privacy-architecture-and-secure-data-handling)
  - [Secure Enclave Key Management](#secure-enclave-key-management)
  - [End-to-End Message Encryption and Commitment](#end-to-end-message-encryption-and-commitment)
  - [Privacy-Preserving Blockchain Data Model](#privacy-preserving-blockchain-data-model)
  - [Zero-Knowledge Proofs and Midnight Integration](#zero-knowledge-proofs-and-midnight-integration)
- [ECHO Tokenomics, Founder Allocation, and Token Launch](#echo-tokenomics-founder-allocation-and-token-launch)
- [Production Launch, Infrastructure, and Deployment](#production-launch-infrastructure-and-deployment)
- [ECHO Comply — Enterprise Compliance Messaging](#echo-comply-enterprise-compliance-messaging)
  - [ECHO Comply — Healthcare (HIPAA)](#echo-comply-healthcare-hipaa)
  - [ECHO Comply — Local Government (FOIA)](#echo-comply-local-government-foia)
  - [ECHO Comply — Law Firms (Chain-of-Custody)](#echo-comply-law-firms-chain-of-custody)
- [ECHO Protocol Foundation and Corporate Structure](#echo-protocol-foundation-and-corporate-structure)
- [Portable Social Graph and Protocol Layer](#portable-social-graph-and-protocol-layer)
- [Post-Quantum Cryptography Mode](#post-quantum-cryptography-mode)
- [Privacy Commons Treasury](#privacy-commons-treasury)
- [Data Sovereignty Layer](#data-sovereignty-layer)

---

# Overview Documents

## Business Problem

# Echo — Product Requirements Document (v3.0)

## Changelog

| Version | Date | Changes |
| --- | --- | --- |
| 3.0 | April 16, 2026 | Complete strategic reorientation. Three-product architecture (ECHO Protocol, ECHO Comply, ECHO Message). B2B compliance-first go-to-market with healthcare, local government, and law firm segments. Token genesis deferred to Phase 3+ (conditional). No earn-by-chatting mechanic. Wyoming DUNA Foundation + commercial LLC corporate structure. Network State vision moved to long-term appendix. Consumer messenger preserved as Phase 2 fast-follow. |
| 2.5.1 | March 31, 2026 | Consolidated document: removed v2.4 duplicate section, resolved tokenomics conflict (auto-scaling model adopted, daily caps removed), fixed P2P references in feature specs to reflect relay architecture, wrote Secure Enclave Key Management spec, wrote Privacy Architecture overview spec, wrote Privacy-Preserving Blockchain Data Model spec, wrote ZK Proofs and Midnight Integration spec, removed stale v1.0 content. |
| 2.5 | March 26, 2026 | Revised founder allocation: CEO 10%, co-founders 2% each (18% total), treasury reduced to 22%. Added trust-tier weighted governance model (single-token, no separate governance token). Added Midnight blockchain evaluation roadmap (Cardano now, Midnight Phase 3+). |
| 2.4 | March 7, 2026 | Finalized tokenomics: 1B total supply. Founder allocation model via on-chain TokenLock. ECHO Wallet built on Stargazer SDK. |
| 2.0–2.3 | February–March 2026 | Relay architecture, Constellation deployment, Network State vision, tokenomics iterations. |
| 1.0 | February 2026 | Initial PRD |

---

## Business Problem

### The Authenticity Crisis

AI-generated deepfakes, synthetic voice cloning, and LLM-powered impersonation have made it impossible to trust digital communication at face value. In 2025–2026, organizations face a new category of threat: communications that look and sound authentic but are entirely fabricated. No existing messaging platform can cryptographically prove who sent a message, when, and that it has not been altered. WhatsApp, Signal, and Telegram provide encryption but not provability. Enterprise platforms like Microsoft Teams and Slack provide audit trails but not cryptographic integrity guarantees that survive outside their own infrastructure.

The cost is measurable. Financial institutions lose approximately $5 billion annually to SMS-based phishing and impersonation attacks. Healthcare organizations face average breach costs of $10.9 million per incident — the highest of any industry. Local governments are increasingly targeted by business email compromise schemes that impersonate officials. **The answer to deepfakes is not better detection — it is cryptographic proof of authenticity at the point of creation.**

### The Global Regulatory Pressure Wave

2026 marks an inflection point. Privacy regulation is no longer a European phenomenon — it is global, accelerating, and enforcement-focused:

* **20+ US states** have comprehensive privacy laws as of January 2026. California, Colorado, and Connecticut jointly enforced Global Privacy Control signals. The enforcement posture is shifting from "you should comply" to "you will comply."
* **EU AI Act** reaches full enforcement in August 2026. Any messaging platform that uses AI for content moderation must provide explainable, auditable decisions — a structural advantage for privacy-by-architecture apps like ECHO that process content only on-device.
* **2026 HIPAA Security Rule updates** mandate encryption for all ePHI, require MFA, and impose 24-hour breach reporting. Healthcare organizations are scrambling to replace personal phones and consumer apps used for clinical coordination.
* **India DPDP Act, Japan APPI amendments, Brazil LGPD enforcement, Vietnam Data Law** — the regulatory wave is global. 100+ jurisdictions now have active data privacy frameworks.
* **Children's data enforcement** is the fastest-growing enforcement priority globally, with age verification requirements cascading from the EU to US states to APAC markets.

### The Surveillance Ratchet

The most important long-horizon threat: every centralized platform is a potential instrument of state surveillance. Russia mandated the installation of "Max" — a state-controlled messaging app — on all smartphones sold in 2025. The UK's Investigatory Powers Act, EU Chat Control proposals, US EARN-IT-style legislation, and India's IT Rules all push toward compelled backdoors or content visibility.

Signal has publicly threatened to withdraw from jurisdictions that mandate backdoors. Any platform that cannot technically refuse a government backdoor request will eventually comply with one. **Decentralized, content-blind architecture is the only engineering defense.** The relay servers, the blockchain, and the protocol itself never see message content — there is nothing to compel.

### The "Harvest Now, Decrypt Later" Quantum Threat

Nation-state adversaries are collecting encrypted communications today, intending to decrypt them when quantum computers mature. NIST standardized post-quantum algorithms in 2024 (Kyber, Dilithium). CISA is actively mandating quantum migration timelines. Healthcare records retained for 7+ years and legal communications subject to long-term eDiscovery holds are already vulnerable to this attack model. **No mainstream messaging platform offers production-grade post-quantum cryptography today.** ECHO is building it.

### The Ownership Gap

Every major messaging platform extracts value from users and returns none. WhatsApp monetizes metadata for Meta's advertising business. Telegram sells Premium subscriptions and TON integration with no user equity. Even Signal asks for donations — users contribute but do not own. Meanwhile, the community-owned model is proving successful at scale: Audius has paid creators $30M+, Theta Network has 300,000+ active edge nodes earning bandwidth fees, Hive distributes 75% of platform inflation to content curators.

The gap between "using a platform" and "owning a stake in the infrastructure you depend on" is closing. The next generation of users — particularly in markets where surveillance is an existential concern — will not accept a platform that treats them as a resource to be extracted. **ECHO's model inverts this: users who contribute to the network own a proportional stake in its value.**

## Vision

ECHO is infrastructure for verifiable private communication — a privacy protocol, not just a messaging app. It ships as three products sharing a single protocol, builds a portable social graph that users own, and evolves into a community-governed organization where users share in the value they create.

**What makes ECHO structurally different from every other messaging platform:**

* **Provable communication** — Every message produces a cryptographic commitment anchored to an immutable ledger. You can prove a conversation happened, what was said, and who said it — without any third party holding the content. No other messenger does this.
* **Portable social graph** — Your identity, verified credentials, and trust relationships are anchored on Cardano. They belong to you, portable to any application. When ECHO succeeds, your reputation follows you. When you want to leave, you take your network with you.
* **Content-blind infrastructure** — The relay servers, the blockchain, and the protocol itself never see message content. Only Merkle roots reach the chain. Privacy is structural, not policy. There is nothing to compel.
* **Post-quantum readiness** — Hybrid X25519 + Kyber key exchange protects against the "harvest now, decrypt later" quantum threat. Built in now, before competitors begin evaluating the migration.
* **Dual-mode architecture** — The same protocol serves a hospital's HIPAA-compliant internal messaging and a journalist's anonymous secure communications. Enterprise and consumer products share the same cryptographic guarantees.
* **Community ownership** — Users who contribute to the network earn governance rights and economic participation. The Privacy Commons Treasury funds legal defense for users under surveillance pressure, subsidized access for journalists and activists, and privacy research that benefits the broader ecosystem.

**The three moats ECHO builds over 2–3 years:**

1. **Technical moat**: Provable communication + portable social graph + post-quantum cryptography. Each takes 12–18 months to implement correctly. The combination creates a capability gap no competitor can close quickly.

2. **Community moat**: Users whose trust reputation, verified credentials, and governance tokens are anchored on ECHO's protocol face enormous switching costs — but unlike centralized lock-in, these switching costs are in the user's favor. Their reputation is richest where they've invested most.

3. **Mission moat**: The Privacy Commons Treasury — legal defense fund, journalist program, privacy research grants — creates a user community that is passionately invested in ECHO's mission. These users don't churn; they evangelize. This is the Signal effect combined with financial ownership.

**What makes ECHO decentralized:** ECHO's decentralization comes from three layers that traditional messengers lack entirely. Your identity is self-sovereign (Cardano DIDs — no company owns your account). Your data integrity is blockchain-verified (metagraph consensus — no company can silently alter records). Your message content is mathematically private (E2E encryption — relay servers see only opaque encrypted blobs). The message relay layer uses a client-server model for reliability, but the servers are stateless pipes with no ability to read, alter, or forge message content, and no authority over your identity or data.

### Long-Term Vision: Community-Owned Network State

ECHO's endgame is not a messaging company — it is a **community-owned digital nation**. The messaging platform is the foundation that creates daily engagement, shared identity, and collective economic power. Over time, ECHO evolves from a product users consume into an organization all users co-own:

**Phase 1–4 (Produ**ct): Build a world-class encrypted messaging platform with 1M+ daily active users. Every user earns ECHO tokens through participation. Token holders govern the protocol through stake-weighted voting.

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
| Node infrastructure | 3 servers minimum (Ubuntu, 8+ cores, 32GB RAM) | AWS/DigitalOcean or bare metal |
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

```plaintext
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
* **ECHO ↔ Bas**e bridge — coordinate with 3A DAO to add ECHO as bridgeable L0 token; enables Aerodrome DeFi and treasury BTC accumulation path
* **Digital Evidence integration** — Go backend submits media fingerprints for optional user-initiated image/video verification; prepare enterprise API client for Phase 5 Org tier
* **Midnight evaluation** — assess stability after 6+ months of mainnet; proof-of-concept ZK trust tier verification ("Prove I'm Tier 3+ without revealing my credential")
* Cardano mainnet deployment (identity layer)
* App Store submission

### Phase 4: Scale & Integrate (Ongoing)

* **Open L1 validators to community** — any operator meeting minimum ECHO TokenLock stake can run a Currency L1 or Data L1 validator (L0 nodes still require 250K DAG)
* **Activate validator slashing** — fraudulent validation, double-signing, extended downtime; slashed ECHO to treasury
* **ECHO ↔ In**k bridge — connect to Kraken exchange via Ink L2; pursue Kraken listing for ECHO
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

**Phase 1–4 (Product Build): $50**0K - $2M

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

**Note on external funding:** ECHO is designed to *not require* venture capital. VC funding creates misaligned incentives — investors want returns, which means extracting value from users. ECHO's model is the opposite: all value stays in the community. If early-stage funding is needed before revenue, it should come from Constellation ecosystem grants or founder capital — not from VCs who would expect equity or governance control, and not from any form of token presale.\
\
\
\
\
\
\
\
\
\
\
\
\
\
\
\
\
\
\
\
\

\

\
\
\
\
\

## Current State

## Current State

The current messaging landscape is dominated by centralized platforms that require users to trust corporate intermediaries with their communications and personal data. WhatsApp, owned by Meta, serves over 2 billion users but stores message metadata on centralized servers and has faced criticism for data sharing practices with parent company Facebook. While WhatsApp offers end-to-end encryption, users cannot verify message authenticity or prove conversations occurred without relying on Meta's infrastructure.

Telegram positions itself as a privacy-focused alternative but only encrypts "secret chats" by default, leaving most communications vulnerable on centralized servers. Its large group capabilities and bot ecosystem attract users seeking advanced features, but the platform's centralized architecture creates single points of failure and government pressure points, as demonstrated by various national bans and content restrictions.

Signal represents the gold standard for privacy-focused messaging, implementing robust end-to-end encryption protocols that have been adopted industry-wide. However, Signal's user base remains limited due to its austere feature set, lack of advanced functionality like large groups or bots, and dependence on centralized servers for message routing and user discovery.

IRON SPIDR, developed by Constellation Network for U.S. federal agencies, demonstrates the potential for blockchain-anchored secure communications but remains restricted to government use cases. Its architecture provides cryptographic proof of message integrity and participant identity through distributed ledger technology, but lacks the consumer-friendly interface and feature richness needed for mainstream adoption.

Current solutions force users to choose between security, features, and usability. No existing platform combines blockchain-verified identity, cryptographic message provability, decentralized infrastructure, and the rich feature set that modern users expect from messaging applications.

## Product Description

## Product Description

ECHO is a privacy protocol, not just a messaging app. It is infrastructure for verifiable private communication — a base layer for trusted identity and provable conversation that ships as enterprise compliance software today and a community-owned communication network tomorrow.

The model that informs ECHO's design: Ethereum didn't succeed because it was a better Bitcoin. It succeeded because it created a base layer that developers could build anything on, and the value of that base layer grew with every application that used it. ECHO applies this logic to private communication — a protocol with a portable social graph, verifiable identity, and cryptographic integrity proofs, on top of which ECHO Comply, ECHO Message, and any third-party developer can build.

### Three Products, One Protocol

| Product | Audience | Phase | Revenue | Core Differentiator |
| --- | --- | --- | --- | --- |
| **ECHO Protocol** | Developers, ecosystem | Foundation-stewarded | Grants, protocol fees | Open-source base layer for trusted communication |
| **ECHO Comply** | Healthcare, government, law firms | Phase 1 (0–18 months) | $30–100/seat/month | HIPAA/FOIA/chain-of-custody compliance with blockchain-anchored integrity |
| **ECHO Message** | Privacy-conscious consumers, crypto-native | Phase 2 (12–30 months) | Freemium + $9.99/month | Portable identity, provable messages, post-quantum option, community ownership |

### The Four Structural Advantages

**\1. Verifiable Communication (No Competitor Has This)**

Every message produces a cryptographic commitment anchored to an immutable public ledger. Any third party can verify a conversation happened, who participated, and that the content hasn't been altered — without ECHO's cooperation. This is not an audit log. It is a mathematical proof that survives ECHO's hypothetical disappearance. Signal cannot provide this. Telegram cannot provide this. No mainstream platform can provide this.

**\2. The Portable Social Graph (Users Own Their Network)**

Your contacts, trust relationships, and verified credentials are anchored on Cardano — not in ECHO's database. They belong to you. Any DID-compatible application can read and verify your trust tier. When you accumulate reputation on ECHO, you are building an asset you own, not a number in someone else's spreadsheet. The social graph is the network effect — and on ECHO, the network effect belongs to users.

Third-party developers who build on the ECHO Protocol access this identity layer immediately. A healthcare coordination app, a DAO governance tool, a secure document sharing system — all can authenticate users by their ECHO DID and inherit the trust verification ECHO has built. This is the protocol network effect: the more applications that build on ECHO, the more valuable every ECHO identity becomes.

**\3. Post-Quantum Cryptography (Built Before Competitors Begin)**

NIST standardized post-quantum algorithms in 2024. CISA is mandating quantum migration timelines. Healthcare records retained for 7+ years are already vulnerable to "harvest now, decrypt later" attacks. ECHO's hybrid X25519 + Kyber key exchange offers quantum-resistant encryption today — as a configurable feature for enterprises and premium consumers — while competitors haven't started their migration planning.

**\4. Community Ownership with Real Purpose**

Unlike "earn while you chat" token mechanics that attract mercenary users and create securities risk, ECHO's community ownership model is built around a mission: the right to private communication. The Privacy Commons Treasury funds legal defense for users under surveillance pressure, subsidized access for journalists and activists in restrictive regimes, and open-source privacy research that benefits the entire ecosystem.

Users who fund a journalist's legal defense are not passive stakers — they are co-owners of an organization fighting for something they believe in. This creates the most passionate, retention-proof user community possible.

### The Community Ownership Model: Learning from What Works

The research on community-owned platforms identifies what succeeds and what fails:

**What fails:** Earn-by-chatting mechanics (Brave/BAT: \~$0.01/hour earnings, poor retention impact). Tokens launched before network effects are established (Session/SESH: 85–95% decline from ATH despite solid architecture). Single-source fee pools with no diversified revenue.

**What succeeds:** Node operation rewards tied to real service provision (Theta: 300K+ nodes, 1B video hours). Creator revenue sharing with transparent on-chain pools (Audius: 5M MAU, $30M+ paid). Community treasury with mission-driven allocation (Nouns DAO: \~$50M treasury funded public goods with high member retention).

ECHO's model takes the lessons: enterprise revenue (ECHO Comply) funds development before tokens exist; tokens launch with demonstrated utility (node operation, governance, premium features) rather than speculative "chat-to-earn"; the Privacy Commons Treasury creates mission-driven purpose that retains users better than financial incentives alone.

### The Data Sovereignty Opportunity

The most radical element of ECHO's consumer model: users who choose to participate can contribute anonymized behavioral data (never content — never content) to a community data pool and receive direct ECHO token payments proportional to the value their data generates. This inverts the surveillance capitalism model:

* WhatsApp collects your data without consent and keeps 100% of the value
* ECHO lets you choose what to share, proves it's anonymous via ZK proof before it leaves your device, and returns 70% of query fees directly to you

The addressable market includes academic researchers studying information spread, public health agencies, enterprise analytics firms, and AI training organizations — all of whom currently cannot access privacy-respecting communication data at scale.

### Technology Foundation

ECHO is built on the Constellation Hypergraph (metagraph for message integrity and token operations), Cardano (self-sovereign DIDs, verifiable credentials), a Go backend WebSocket relay (stateless, content-blind), and Swift/SwiftUI iOS with Secure Enclave key management. The protocol's content-blind architecture ensures there is nothing to compel, no ePHI on servers to breach, and no platform-controlled backdoor to mandate.

## Personas

## Personas

ECHO serves two distinct product tracks — ECHO Comply (enterprise) and ECHO Message (consumer) — each with specific personas. Phase 1 personas are enterprise-focused. Consumer personas become active at Phase 2.

---

## ECHO Comply Personas (Phase 1 Priority)

### Healthcare IT Administrator

Maria, 42, Director of Information Security at a 400-bed regional hospital. Responsible for HIPAA compliance across all clinical communications. Currently struggles with staff using personal WhatsApp for patient coordination — a clear HIPAA violation but difficult to stop without a better alternative. Has budget for a compliant solution but faces a difficult procurement process. Needs: signed BAA, MFA enforcement, configurable retention (minimum 6 years), audit log exportable for OCR investigations. Success measure: zero HIPAA violations related to messaging in the next audit cycle. Does not care about blockchain or tokens — just needs compliance documentation.

### Clinical End User (Healthcare)

Dr. James, 38, attending cardiologist. Uses his personal iPhone to coordinate with nurses about patient care because it's faster than the hospital's pager system. Doesn't want to carry two phones or log into a complicated enterprise system. Needs: a fast, familiar messaging interface that works exactly like iMessage but has a green compliance checkmark. Adoption dies if the app feels like "compliance software." Will not read documentation. Success measure: never has to think about compliance — it just works.

### Local Government Records Officer

Patricia, 55, City Clerk for a municipality of 80,000 residents. Responsible for responding to public records (FOIA) requests for official communications. Currently receives requests for messages sent on personal phones and WhatsApp by council members — cannot produce them because they were not preserved. Faces personal liability for failure to respond to records requests. Needs: automatic preservation of all official communications with tamper-evident integrity proofs, exportable in standard formats for FOIA responses. Success measure: every FOIA request related to digital communications is answerable within the statutory deadline.

### Law Firm IT Director

David, 48, Director of IT at a 200-attorney litigation firm. Responsible for implementing litigation hold policies and ensuring chain-of-custody for all digital evidence. Currently uses Microsoft Teams but lacks blockchain-anchored integrity proofs that survive outside Microsoft's infrastructure — a liability in court. Needs: matter-based message organization, litigation hold enforcement, ethical wall management, and Digital Evidence-compatible export. Will sign a long-term contract if the solution covers eDiscovery needs. Success measure: first successful use of ECHO Comply communications as court-admitted evidence.

---

## ECHO Message Personas (Phase 2)

### Privacy-Conscious Professional

Sarah, 34, attorney in private practice. Uses Signal for personal messaging but needs provable communications for client interactions that don't go through her firm's enterprise system. Wants to prove conversations happened without relying on screenshots. Values self-sovereign identity. Not interested in tokens initially but appreciates governance voting on platform decisions. Will tell colleagues about ECHO if it helps her with client documentation.

### Crypto-Native Community Builder

Marcus, 28, DAO contributor and community manager. Manages three Discord servers and a Telegram group. Frustrated by platform risk — Telegram bans, Discord data breaches, arbitrary account deletions. Wants messaging infrastructure his community collectively controls. Interested in staking, governance, and the long-term ownership thesis when the token launches. Will advocate strongly if ECHO Message delivers on the ownership promise.

### International User in Restrictive Environment

Aisha, 31, journalist in a country with communications surveillance. Needs metadata-minimized messaging with verified identity to confirm sources are who they claim. Interested in Hidden Folders and Duress PIN for protecting source identities. Not interested in tokens. Needs the app to not look like "a crypto app" on her phone. Values the Duress PIN feature specifically — it protects her sources even under physical coercion.

### Developer and Protocol Builder

Wei, 26, open-source contributor and protocol developer. Interested in ECHO after the Phase 2 open-source release. Wants to build integrations, third-party clients, and automated workflows on the ECHO Protocol. Motivated by protocol grants and the long-term ecosystem opportunity. Values API access, documentation quality, and the Foundation governance model.

### Validator Node Operator (Phase 3+)

Technical operators who run L0 hybrid nodes and L1 validators once the token launches. Motivated by protocol fees, staking yields, and supporting network decentralization. Transitions from Foundation-operated (Phases 1–3) to community-operated (Phase 4+). Requires clear documentation on node requirements, staking mechanics, and slashing conditions.

## Success Metrics

## Success Metrics

ECHO operates on a dual-track model — enterprise and consumer — with distinct success metrics per phase. Enterprise revenue funds consumer development, so Phase 1 success is measured entirely by enterprise outcomes.

### Phase 1 Success Metrics (Months 0–18): Enterprise Foundation

| Metric | Target |
| --- | --- |
| Foundation + LLC incorporated | Month 2 |
| Core team size | 3–5 people |
| ECHO Comply paying customers | 5–10 |
| Annual Recurring Revenue (ARR) | $500K–$2M |
| Average contract value | $50K/year |
| Metagraph uptime | 99.5%+ |
| Message delivery success rate | 99.9%+ |
| SOC 2 Type I certification | Achieved |
| HIPAA BAAs signed | 5+ |
| Enterprise NPS | 40+ |

### Phase 2 Success Metrics (Months 12–30): Consumer Launch + Enterprise Expansion

| Metric | Target |
| --- | --- |
| ECHO Comply paying customers | 25–50 |
| ECHO Comply ARR | $3–5M |
| ECHO Message registered users | 50,000+ |
| ECHO Message 30-day retention | 40%+ |
| ECHO Message daily messages/user | 30+ |
| ECHO Message Premium conversion ($9.99/mo) | 5% |
| Combined ARR | $3.5–6M |
| Protocol open-sourced | Yes |
| Android clients launched | Yes |
| SOC 2 Type II certification | Achieved |

### Phase 3 Success Metrics (Months 30–48): Scale + Token Evaluation

| Metric | Target |
| --- | --- |
| ECHO Comply paying customers | 100+ |
| ECHO Message MAU | 500,000+ |
| ECHO Message 30-day retention | 55%+ |
| Combined ARR | $10M+ |
| Token genesis conditions met | All 5 evaluated |
| Community governance participation (if token launched) | 15%+ of token holders |

### Enterprise-Specific Metrics

| Metric | Phase 1 | Phase 2 |
| --- | --- | --- |
| Sales cycle length | <90 days (mid-size) | <60 days |
| Customer churn (annual) | <10% | <5% |
| Integrity proof verification success rate | 99.9%+ | 99.9%+ |
| eDiscovery export success rate | 100% | 100% |
| Compliance audit findings | Zero critical | Zero critical |
| HIPAA incident rate | Zero breaches | Zero breaches |

### Protocol Reliability Metrics

| Metric | Target |
| --- | --- |
| Metagraph snapshot finality | <15 seconds |
| Message delivery rate | 99.9%+ |
| Relay uptime | 99.9%+ |
| Merkle root anchoring success | 99%+ |
| DID resolution latency | <500ms |

### Revenue Model

| Source | Phase 1 | Phase 2 | Phase 3 |
| --- | --- | --- | --- |
| ECHO Comply subscriptions | $500K–$2M | $3–5M | $10M+ |
| ECHO Message Premium | — | $500K–$1M | $2–5M |
| Token ecosystem (if launched) | — | — | Protocol fees |

**Revenue allocation (Foundation structure):** 60% operations (salaries, infrastructure, compliance), 25% Foundation development grant, 15% Foundation reserve.

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
* Production infrastructure credentials (AWS keys, API tokens)
* Security vulnerability reports (until patched)
* Financial institution partnership agreements

### \2. Beta User Targets Before Launch

**Recommended Beta Progression:**

**Phase 2 Alpha (Closed Beta): 100-500 Users**

* **Target:** 100 users minimum, 500 users maximum
* **Duration:** 2-3 months during Core Build phase
* **Recruitment:** Invite-only from crypto/privacy communities, Constellation ecosystem, personal networks
* **Purpose:** Stress test P2P relay infrastructure, debug E2E encryption edge cases, validate token reward mechanics
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

* **did:key** method — W3C-standard DID derived from the user's Secure Enclave key pair. Zero cost, device-sovereign, permanent. No blockchain required for DID creation.
* **Constellation Identity Metagraph** (dedicated, separate from Data L1 and Currency L1) — issues W3C VC 2.0 Verifiable Credentials, anchors trust tier commitments H(tier || nonce), manages EchoOrgRoleCredential org membership, and publishes StatusList2021 revocation vectors.
* **Feeless for users** — ECHO treasury pays Constellation snapshot fees in DAG. No identity transaction fees to end users.
* **ZK proofs (Phase 2+)** — Constellation-native zero-knowledge circuits for privacy-preserving credential verification. Midnight (Cardano ZK sidechain) evaluated as Phase 3 option if production-ready.
* **No Cardano in Phase 1–2** (unlikely to be needed at all — may never be used)

**Messaging Infrastructure**

* Go backend with WebSocket relay for message routing
* X25519 key agreement + ChaCha20-Poly1305 for E2E encryption
* APNs for iOS push notifications
* Encrypted store-and-forward for offline message queuing
* TLS 1.3 transport encryption

**Blockchain Integration**

* Constellation Hypergraph (Data L1) for message integrity anchoring via Merkle roots
* Constellation Identity Metagraph (separate L1) for VC issuance, trust tier commitments, StatusList2021 revocation
* Constellation Metagraph (Currency L1) for ECHO token operations (Phase 3+)
* Tessellation v3 transaction primitives: TokenLock, StakeDelegation, AtomicAction, WithdrawLock, FeeTransaction
* Scala/JVM for L1 validation logic using Euclid SDK across all three metagraphs
* No Cardano in Phase 1–2 (evaluate Midnight at Phase 3 for ZK proof circuits)

**Storage & Media**

* IPFS/Storj for distributed encrypted file storage
* PostgreSQL and Redis as performance caches only (not source of truth)
* All persistent state on-chain (Constellation metagraphs for message integrity, identity credentials, and token operations)

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

## Product Features

## Product Architecture — Three Products, One Protocol

ECHO ships as three products sharing a single protocol. Enterprise and consumer products share the same cryptographic guarantees, the same metagraph, and the same relay infrastructure. They differ only in the policy layer on top.

| Product | Audience | Phase | Revenue |
| --- | --- | --- | --- |
| **ECHO Protocol** | Developers, ecosystem | Foundation-stewarded | Grants |
| **ECHO Comply** | Healthcare, government, law firms | Phase 1 (0–18 months) | $30–100/seat/month |
| **ECHO Message** | Privacy-conscious consumers, crypto-native | Phase 2 (12–30 months) | Freemium + $9.99/month premium |

---

## ECHO Comply Features (Phase 1)

**Target segments:** Healthcare (HIPAA), Local Government (FOIA), Law Firms (chain-of-custody)

* End-to-end encrypted messaging with no server-side decryption under any circumstances
* Tamper-evident message integrity proofs anchored to Constellation metagraph (Merkle commitments)
* Admin console: user provisioning, retention policies, compliance dashboard, eDiscovery export
* SSO/SAML/OIDC integration with Azure AD, Okta, Google Workspace; SCIM provisioning
* Configurable retention policies (permanent, time-limited, litigation hold)
* Role-based routing for healthcare (on-call cardiologist, charge nurse) with escalation workflows
* FOIA/public records export with independently verifiable integrity proofs for government
* Litigation hold, ethical wall enforcement, matter-based organization for law firms
* HIPAA Business Associate Agreement (BAA) included for healthcare customers
* Multi-factor authentication enforced (Secure Enclave biometric + device credential)
* All compliance features invisible to end users — the messaging experience is a familiar consumer-grade interface

**Pricing:** Comply Starter $30/seat/month (min 10), Comply Professional $50 (min 50), Comply Enterprise $80–100 (min 500)

---

## ECHO Message Features (Phase 2)

**Core Messaging:**

* 1:1 and group chats (up to 1,000 members Phase 2, scaling to 100,000+ Phase 3)
* Voice and video calls: 1:1 (E2E encrypted), group voice (up to 32), group video (up to 8)
* Text, images, video, voice notes, documents, reactions, stickers
* All content E2E encrypted by default (X25519 + XChaCha20-Poly1305). No opt-in required.
* Offline message queuing with encrypted store-and-forward
* Disappearing messages (5 minutes to 7 days, configurable)
* Screen security (screenshot detection + optional blocking)

**Identity and Trust:**

* Self-sovereign identity via Cardano DIDs — no phone number required
* Username-based discovery + QR code + invite links
* Progressive trust tier system (Tier 1–5) based on identity verification depth
* Verification badges displayed on profiles; underlying score kept private
* Trust tier unlocks: governance voting weight, premium features, staking tiers (Phase 3+)

**Provable Messaging:**

* Every message produces a cryptographic commitment anchored to the Constellation metagraph
* User-initiated "proof of conversation" — verifiable by any third party without ECHO access
* Proof generation is user-initiated, not automatic

**Privacy Features:**

* Hidden Folders with biometric protection (Secure Enclave)
* Duress PIN: enters decoy empty folder under coercion — indistinguishable from normal access
* No admin visibility — user is sole authority over their account

**Gamification (Phase 2 — badge only, pre-token):**

* Quest system: structured activities that onboard users into ECHO features and build trust tier
* Streak system: daily messaging streaks with progress tracking and milestone badges
* Leaderboards: opt-in weekly/monthly rankings for messaging volume and community contributions
* Pre-token quests award badges only (no token rewards until Phase 3+)

**Premium Features ($9.99/month via Apple IAP):**

* Increased file size limits (2GB vs. 100MB free)
* Custom themes and app icons
* Extended disappearing message timers (up to 90 days)
* Priority message delivery
* Advanced privacy controls (granular read receipts, typing indicators)
* Premium badge on profile

**Wallet and Governance (Phase 3+ — conditional on token launch):**

* Native ECHO Wallet built on Stargazer SDK: staking, delegation, governance voting
* Staking tiers: Bronze 5% APY (30d), Silver 8% (90d), Gold 12% (180d), Platinum 15% (365d)
* Token-holder governance: Governance Weight = StakedECHO × TrustTierMultiplier
* PacaSwap swap integration (ECHO ↔ DAG, ECHO ↔ USDC)
* Note: No earn-by-chatting. Tokens are not earned for sending messages.

**Security Principle:** Free tier is complete. Premium adds convenience, customization, and status. Security and privacy are never gated behind payment.

Signal (protocol gold standard).

ZK-proofs for auth without revealing data; audit logs on-chain for disputes.

**Extensibility**

Bots for automation (e.g., polls, games), channels for broadcasts, unlimited file sharing (up to 2GB).

Telegram (ecosystem).

Bots run on-chain via smart contracts; decentralized storage for channels.

**Trust & Social**

\- Verified badges (blue/gold via on-chain creds). 

\- "Trusted Circles": Mutual follows evolve to "Inner Circle" (unlimited history) or "Acquaintance" (limited). 

\- Trust Score: 0-100 based on verifs, interaction volume, no reports (increases with time). 

\- Report/Block with on-chain evidence. (verification, mutuals).

\- Scores computed via smart contracts; portable across apps via DIDs.

**Extras**

Status updates, polls, location sharing (ephemeral), multi-device sync.

WhatsApp/Telegram.

Sync via encrypted P2P; locations via ZK (prove proximity without coords).

This creates a "trust gradient": New users start basic (acquaintance mode), gain features/privileges as trust builds—e.g., higher scores unlock premium calls or bank links.

## Development Roadmap

**Development Roadmap**

Phased approach, assuming a 6-12 month MVP timeline with a 5-10 dev team (blockchain + mobile experts). Budget estimate: $500K-$2M (dev salaries, audits, marketing).

**1: Research & Prototype**

Validate IRON SPIDR parallels; build PoC for Cardano DID + P2P chat. Audit E2EE.

1-2 months

Wireframes, DID wallet integration demo, security whitepaper.

**2: Core Build**

Implement messaging stack, basic UI, trust scoring contracts. Testnet deployment.

3-5 months

Alpha app (iOS/Android/web); 100 beta users for P2P stress tests.

**3: Feature Polish & Launch**

Add bots/channels, multi-device, full trust UI. Community node incentives.

2-3 months

Mainnet launch; App Store/Google Play submission; Marketing (target privacy enthusiasts).

**4: Scale & Integrate**

Optimize for 1M+ users; Bank pilots with compliance audit trail. Governance DAO for protocol updates.

Ongoing (post-launch)

Open L1 validators to community; activate validator slashing; federated relay nodes; ZK proof system; Android support; API for third-party integrations; Quarterly audits.

**5: Community Economy (Year 2–3)**

*Prerequisite: 500K+ MAU, stable governance DAO operational*

Launch VIP subscriptions and Organization plans with Digital Evidence (Smart Checkmark, audit fingerprinting, compliance dashboard). Deploy AI treasury agents (CFO, burn, BTC reserve, stablecoin manager, compliance, reporting). All revenue flows to on-chain community treasury. Community votes on annual treasury allocation ratios (burn %, BTC %, operational %, RWA fund %, emergency %). Elect first 5 community board members. Open-source entire codebase under permissive license. Engage legal counsel for DAO LLC formation.

Year 2+

Self-sustaining revenue model; AI-managed treasury operations; first ECHO token burns and BTC reserve accumulation.

**6: Network State Formation (Year 3+)**

*Prerequisite: 1M+ MAU, self-sustaining treasury, legal entity established*

Establish legal entity (DAO LLC) to hold real-world assets on behalf of community. First real-world asset acquisition via community governance vote. Expand RWA portfolio (land, buildings, companies, infrastructure). Network State membership tiers tied to ECHO token staking. Cross-metagraph alliances. Long-term goal: ECHO community as recognized digital jurisdiction with physical presence.

Year 3+

Community-owned physical assets; Network State governance operational; treasury managing real-world investments.

Tools: Scala for Constellation L1 validation logic (Euclid SDK); Go for backend relay; Swift for iOS. Open-source everything on GitHub for community buy-in.

## Architecture

**High-Level Architecture**

ECHO uses a single-chain Constellation-native architecture for all blockchain operations: identity (Constellation Identity Metagraph), message integrity (Constellation Data L1), and token operations (Constellation Currency L1, Phase 3+). This eliminates the two-chain Cardano + Constellation complexity from v3.0 and below.

**Identity & Auth**\

User-owned did:key DIDs — W3C-standard identifiers derived from the user's Secure Enclave key pair. Zero cost, device-sovereign, no chain required for DID creation. Verifiable Credentials (trust tier commitments, org membership, professional credentials) are issued and anchored to the Constellation Identity Metagraph. No email, phone, or third-party wallet required.

Technology: did:key (W3C standard); Constellation Identity Metagraph (Scala, Euclid SDK) for VC issuance, H(tier || nonce) trust commitments, and StatusList2021 revocation. Feeless for users — ECHO treasury pays DAG snapshot fees.

High Decentralization Level: Device-sovereign identity, on-chain credential anchoring, no central auth server, no Cardano dependency in Phase 1–2.

Note: Cardano / Midnight are not in scope. ECHO may never use Cardano. The did:key standard keeps identities interoperable with any DID ecosystem without a Cardano dependency.

**Messaging Layer**

End-to-end encrypted messaging, groups, and calls using client-server relay architecture. Messages encrypted on-device with X25519 + ChaCha20-Poly1305 before transmission. Relay servers transport opaque ciphertext and cannot read, modify, or forge content. Message hashes anchored to blockchain for provability through Merkle root commitments on Constellation metagraph.

Go backend WebSocket relay with APNs push notifications; Constellation Hypergraph (Data L1) for message integrity commitments; X25519 + ChaCha20-Poly1305 for E2E encryption.

Medium-High Decentralization Level: Stateless relay servers; identity and message integrity on-chain; progressive metadata protection (sealed sender Phase 3, federated relays Phase 4).

**Storage & Media**

Encrypted blobs for files/photos; temporary off-chain caching for speed.

IPFS/Storj for distributed encrypted storage.

Medium-High Decentralization Level: Blockchain anchors hashes for integrity.

**Trust Engine**

Dynamic scoring based on interactions, verifications, and on-chain credential history. Trust tier commitments published to Constellation Identity Metagraph as H(tier || nonce) — proves tier without revealing score.

Constellation Identity Metagraph (Scala, Euclid SDK); off-chain trust score calculation with on-chain commitment anchoring.

High Decentralization Level: Transparent, auditable via Constellation; raw scores never on-chain.

**Frontend/UX**

iOS-first (Swift, SwiftUI, Secure Enclave); Android Phase 2+. ECHO Message and ECHO Comply share the same cryptographic primitives via shared Swift library.

**Backend/Relays**

Go backend relay services for message routing, offline queuing, and push notifications. Stateless coordinator — all persistent state on-chain. Snapshot fees paid in DAG to Constellation Hypergraph (offset by delegation).

Go with WebSocket + APNs; PostgreSQL and Redis as performance caches only; three Constellation metagraphs (Identity, Data L1, Currency L1).

Medium Decentralization Level: Community-operated relay nodes in Phase 4; validator staking requirements.

# Feature Requirements

## Decentralized Identity and Authentication

### Decentralized Identity and Authentication

ECHO uses a Constellation-native identity architecture. Every user has a **did:key** — a W3C-standard Decentralized Identifier derived directly from their Secure Enclave cryptographic key pair. The DID is permanent, device-sovereign, and costs nothing to create: no blockchain transaction, no ADA required, no dependency on any third-party chain. The private key never leaves the Secure Enclave; the DID is the public key's fingerprint.

Verifiable Credentials (trust tier commitments, organization membership, professional licenses) are issued and anchored on the **Constellation Identity Metagraph** — a dedicated metagraph separate from the message integrity and token layers. This collapses the architecture to a single blockchain dependency (Constellation) for all identity, integrity, and token operations, eliminating the cost volatility and engineering complexity of a two-chain setup.

Users begin onboarding by generating a new did:key pair in their device's Secure Enclave. No email, phone number, or third-party wallet is required. The DID is immediately usable for messaging. Users then progressively earn Verifiable Credentials — each trust tier upgrade, completed identity verification, and organization membership credential is issued as a W3C VC 2.0, signed by the ECHO Protocol Foundation's Identity Metagraph, with revocation managed via StatusList2021 anchored to Constellation. The trust scoring algorithm evaluates on-chain credential history, interaction volume, and verification depth to place users in Trust Tiers 1–5, each unlocking additional features, governance weight, and (at Phase 3+) token multipliers.

**Architecture decisions (ratified in PRD v3.1 Q&A, April 2026):**

* **did:key** as base DID method: pure cryptographic identity derived from the Secure Enclave key. Zero cost. Permanent. No chain required for DID itself.
* **Constellation Identity Metagraph** (separate from Data L1 and Currency L1): handles VC issuance, trust tier commitments H(tier || nonce), organization membership VCs (EchoOrgRoleCredential), and StatusList2021 revocation vectors.
* **Feeless for users**: ECHO treasury pays Constellation snapshot fees in DAG. Users never pay identity transaction fees.
* **On-chain events (Phase 1–**2): trust tier upgrades, org membership VCs, StatusList2021 revocation. Governance votes deferred to Phase 3.
* **Cardano:** Not in scope. The did:key + Constellation Identity Metagraph architecture fully satisfies ECHO's identity and credential requirements without Cardano. Midnight (Cardano's ZK sidechain) may be evaluated at Phase 3+ only if Constellation-native ZK circuits prove insufficient for a specific use case — this is considered unlikely given Constellation's metagraph flexibility. ECHO may never use Cardano.

**Cost analysis at scale:**

The did:key + Constellation architecture achieves near-zero identity cost. A 1M-user deployment generates roughly 2–3 identity events per user per year (tier upgrade, org VC issuance, occasional revocation). Constellation snapshot fees are paid in DAG by the ECHO treasury; at current DAG pricing and Tessellation v3 throughput, identity infrastructure costs are well under $10K/year at 1M users — compared to \~$98K/year for Cardano Hydra DID writes at equivalent scale. More importantly, costs are denominated in DAG (already staked and held for metagraph operations) rather than ADA, eliminating the second-token cost volatility risk.

**ZK proofs (Phase 2+):** Zero-knowledge proofs for privacy-preserving verification (prove Trust Tier 3+ without revealing exact score, prove age ≥ 18 without revealing birthdate) are planned as Constellation-native circuits written against the Euclid SDK. Midnight (Cardano's ZK-focused partner chain) is evaluated at Phase 3 as an optional integration if its production-ready ZK tooling provides materially better developer experience or privacy guarantees than Constellation-native ZK.

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

File storage utilizes a hybrid approach where frequently accessed files are cached on IPFS nodes for quick retrieval, while long-term storage is handled through Storj's distributed storage network to ensure availability with strong encryption guarantees. Users can configure automatic file expiration for sensitive documents, with cryptographic deletion ensuring files become permanently inaccessible after specified timeframes. The system includes comprehensive file management features including search functionality, version control for collaborative documents, and automatic backup of important files to user-controlled storage locations.

## Message Reactions, Polls, and Interactive Elements

### Message Reactions, Polls, and Interactive Elements

This feature provides users with rich interactive communication tools including emoji reactions, polls, surveys, and interactive buttons that enhance group engagement while maintaining the platform's security and privacy standards. The system enables expressive communication and decision-making tools that rival traditional social media platforms while preserving end-to-end encryption and decentralized architecture.

Users can react to messages using a comprehensive emoji library that includes standard Unicode emojis, custom reactions, and blockchain-verified NFT emojis that users can collect and trade. The reaction system supports multiple reactions per user per message, with real-time synchronization across all participants' devices through the peer-to-peer network. Poll creation allows users to pose questions with multiple choice answers, with voting results encrypted and tallied through zero-knowledge proofs that preserve voter privacy while ensuring result integrity. Advanced poll features include anonymous voting, time-limited polls, and weighted voting based on trust scores for community governance decisions.

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

This feature provides users with secure, biometrically-protected folders for sensitive conversations that require additional privacy layers beyond standard end-to-end encryption. Hidden folders remain completely invisible in the main chat interface and can only be accessed through successful biometric authentication, creating a secure vault for confidential communications that protects against unauthorized access even if the device is compromised. A Duress PIN feature protects users under physical coercion by opening a decoy folder that reveals nothing about the actual hidden content.

Users create hidden folders by selecting conversations and moving them to a protected space that requires Face ID, Touch ID, or other biometric verification methods supported by their device. The folder creation process generates additional encryption keys bound to the user's biometric template, ensuring that even if someone gains access to the device or the user's primary authentication credentials, they cannot access hidden conversations without the correct biometric. The system supports multiple hidden folders with different access requirements, allowing users to categorize sensitive conversations by security level or relationship type.

The hidden folder interface operates as a completely separate chat environment that maintains its own message history, notification settings, and backup protocols. Messages within hidden folders use enhanced encryption that combines the standard E2E encryption implementation with biometric-derived key material. Users can configure custom notification behaviors: silent notifications appearing only when the folder is unlocked, or complete notification suppression preventing any indication of incoming messages.

**Duress PIN:** Users can configure a secondary PIN that — when entered instead of the normal access method — opens a decoy Hidden Folder environment. The decoy folder is empty by default (or optionally seeded with innocuous conversations). When the Duress PIN is entered, the system behaves identically to normal access: no warning, no flag, no difference in timing or behavior. An observer watching over the user's shoulder cannot tell whether the real or decoy folder was opened. This protects journalists, activists, and any user whose device might be examined under physical coercion.

The feature integrates with the device's Secure Enclave to ensure biometric templates and derived encryption keys never leave the hardware security module. Hidden folder metadata is encrypted and stored locally rather than synchronized across multiple devices. Users can optionally enable secure backup via additional biometric verification combined with a recovery phrase.

## Silent and Scheduled Private Chats

### Silent and Scheduled Private Chats

This feature enables users to send messages that generate no notifications or visible indicators on the recipient's device, while also supporting scheduled message delivery for time-sensitive communications across different time zones or planned conversations. The system provides granular control over message visibility and timing while maintaining end-to-end encryption and blockchain anchoring for all communications.

Users can activate silent mode for individual conversations or specific messages, which suppresses all notification behaviors including push notifications, badge counts, typing indicators, and read receipts on the recipient's device. Silent messages appear in the conversation thread only when the recipient actively opens the chat, creating a non-intrusive communication channel for sensitive or low-priority messages. The scheduling functionality allows users to compose messages that are delivered at predetermined times, with messages encrypted and stored locally on the sender's device until the scheduled delivery time when they are transmitted through the normal P2P messaging infrastructure.

The silent messaging system operates through enhanced metadata handling where notification suppression flags are embedded in the encrypted message payload, ensuring that even relay nodes cannot determine which messages should generate notifications. Scheduled messages use time-locked encryption where the message content is encrypted with keys that are only released at the specified delivery time through smart contract automation on the Constellation network. Users can schedule messages up to 30 days in advance, with the system supporting recurring message patterns for regular communications like daily check-ins or weekly reports.

The feature integrates with the existing trust scoring system to prevent abuse, where users with low trust scores face limitations on silent messaging frequency to prevent spam or harassment. Scheduled messages maintain full blockchain anchoring and provable integrity features, with delivery timestamps cryptographically verified to ensure messages were sent at the intended time. The system supports cross-timezone scheduling with automatic conversion based on recipient location preferences while maintaining privacy through zero-knowledge proofs that confirm delivery timing without exposing user location data.

This functionality addresses the need for respectful communication patterns that don't interrupt recipients during sensitive times while enabling users to maintain consistent communication schedules across global time zones. The silent messaging capability is particularly valuable for professional communications, emergency contact protocols, and personal relationships where immediate notification may be inappropriate or disruptive.

## Disappearing Messages with Cryptographic Verification

### Disappearing Messages with Cryptographic Verification

This feature provides users with the ability to send messages that automatically delete from all devices after predetermined time periods while maintaining cryptographic proof that the messages existed and were delivered, addressing privacy needs without compromising the platform's provable integrity capabilities. The system ensures that sensitive communications can be ephemeral while preserving audit trails for compliance and dispute resolution when necessary.

Users can enable disappearing messages for individual conversations or specific messages by selecting from preset time intervals ranging from 10 seconds to 7 days, with custom timing options available for premium users. When activated, messages display countdown timers that show remaining visibility time to all participants, creating transparency about message lifecycle. The deletion process occurs simultaneously across all devices through cryptographic coordination, ensuring that messages cannot persist on any participant's device beyond the specified timeframe. However, the system maintains blockchain-anchored hashes of deleted messages that can prove conversations occurred without revealing content, enabling users to demonstrate communication history for legal or business purposes.

The technical implementation uses time-locked smart contracts on the Constellation network that automatically trigger deletion commands across the peer-to-peer network when expiration times are reached. Messages are encrypted with time-sensitive keys that become invalid after the specified period, making recovery impossible even if encrypted data fragments remain on devices. The system supports different deletion policies for different trust levels, where verified users can set longer retention periods and access advanced features like selective message preservation for important communications. Screenshots and forwarding are technically prevented through device-level security measures, though users are notified when these protections may be bypassed.

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

Channel creators can establish broadcast channels that support unlimited subscribers, with content distributed through the platform's peer-to-peer network to ensure resilience and prevent censorship. Channels can be configured as public (discoverable through search), private (invitation-only), or semi-private (discoverable but requiring approval to join). Content types include text messages, images, videos, files, polls, and interactive elements, with all content encrypted and distributed through the same security infrastructure used for private messaging. Channel administrators can configure moderation settings, subscriber permissions, and content policies while maintaining transparency through blockchain-anchored governance records.

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

## Platform Roadmap and Future Vision

# Platform Roadmap and Future Vision

## Development Roadmap

Phased approach, assuming a 6-12 month MVP timeline with a 5-10 dev team (blockchain + mobile experts). Budget estimate: $500K-$2M (dev salaries, audits, marketing).

**1: Research & Prototype**

Validate IRON SPIDR parallels; build PoC for Cardano DID + P2P chat. Audit E2EE.

1-2 months

Wireframes, DID wallet integration demo, security whitepaper.

**2: Core Build**

Implement messaging stack, basic UI, trust scoring contracts. Testnet deployment.

3-5 months

Alpha app (iOS/Android/web); 100 beta users for P2P stress tests.

**3: Feature Polish & Launch**

Add bots/channels, multi-device, full trust UI. Community node incentives.

2-3 months

Mainnet launch; App Store/Google Play submission; Marketing (target privacy enthusiasts).

**4: Scale & Integrate**

Optimize for 1M+ users; Bank pilots (see below). Governance DAO for updates.

Ongoing (post-launch)

API for third-party integrations; Quarterly audits.

Tools: Use Solidity/Rust for contracts (Cardano/Constellation compatible); Test with Ganache for Ethereum sims if needed. Open-source everything on GitHub for community buy-in.

## Success Metrics

Success for this decentralized messaging platform will be measured through a combination of user adoption metrics, security effectiveness indicators, and ecosystem health measurements that demonstrate the product's ability to solve the core problems of centralized messaging while maintaining user satisfaction and network growth.

User adoption and engagement metrics will track the platform's ability to compete with established messaging apps while providing superior security and decentralization benefits. We will monitor monthly active users with a target of reaching 100,000 users within the first year and 1 million users by year two, focusing on organic growth through word-of-mouth and privacy-conscious user communities. Daily message volume per user should exceed 50 messages to indicate the platform serves as a primary communication tool rather than a secondary privacy option. User retention rates must demonstrate that the enhanced security features do not compromise usability, with 30-day retention rates targeting 60% and 90-day retention rates of 40%, comparable to successful messaging platforms.

The effectiveness of the trust and verification system will be measured through fraud prevention and spam reduction metrics that validate the core value proposition. We will track the percentage of users who complete progressive verification steps, targeting 70% of active users achieving basic verification and 30% completing advanced credentials within six months of registration. Spam and fraud incident rates should remain below 0.1% of total messages, significantly lower than traditional platforms, while maintaining user satisfaction scores above 4.2/5.0 for the verification process. The trust scoring system's effectiveness will be evaluated through correlation analysis between trust scores and user behavior, ensuring scores accurately predict trustworthy interactions.

Network decentralization and resilience metrics will demonstrate the platform's ability to operate without central points of failure while maintaining performance standards. We will monitor the distribution of relay nodes across geographic regions and operators, targeting at least 100 community-operated nodes within the first year to ensure no single entity controls more than 10% of network capacity. Message delivery success rates must exceed 99.5% even during peak usage periods, with average message latency remaining under 500 milliseconds for direct peer-to-peer connections. Blockchain integration effectiveness will be measured through successful hash anchoring rates above 99% and zero-knowledge proof generation times under 2 seconds for provable messaging features.

Business impact and ecosystem development metrics will track the platform's progress toward enterprise adoption and financial integration opportunities. We will measure the number of enterprise pilot programs initiated, targeting partnerships with at least 5 financial institutions for fraud-proof customer communications within 18 months of launch. Developer ecosystem growth will be tracked through third-party bot and application integrations, with a goal of 50 verified bots and 10 enterprise integrations by the end of year one. Revenue metrics will focus on premium feature adoption rates and enterprise licensing, targeting 15% of users upgrading to premium tiers and generating $1M in annual recurring revenue by year two while maintaining the free tier's core functionality.

## Universal Onboarding and Identity Creation

### Universal Onboarding and Identity Creation

This feature provides a simple and familiar registration flow for new users, starting with a mobile number to bootstrap the account before seamlessly transitioning them to a secure, self-sovereign decentralized identity (DID). The goal is to lower the barrier to entry for mainstream users while immediately establishing the high-trust foundation of the platform.

The user journey begins upon first opening the app, where they are prompted to enter their mobile number. The system sends a one-time verification code via SMS to confirm ownership. Once verified, the app guides the user through the creation of their core digital identity. In the background, the system generates a new DID on the Cardano blockchain, effectively linking their familiar identifier (phone number) to their new sovereign identity. This initial step also includes the creation of a passkey, binding their account to their device's biometric security for future passwordless logins. The mobile number is then decoupled and can be optionally discarded by the user to enhance privacy, with the DID and passkey becoming the primary account credentials.

This onboarding process serves as the entry point to the platform's progressive identity system. Immediately after this initial setup, the user is encouraged to further strengthen their profile by engaging with the "Streamlined Onboarding with Verifiable Credentials" or the "In-App High-Assurance Identity Verification" flows to add credentials and increase their trust score.

This feature requires a temporary reliance on a centralized SMS gateway for the initial phone number verification, a trade-off made to ensure a frictionless onboarding experience for non-crypto-native users. The architecture is designed to immediately migrate the user to the decentralized identity model, minimizing long-term reliance on the centralized component. The flow is dependent on the successful and timely creation of a DID via the integrated Cardano infrastructure.

## Privacy Architecture and Secure Data Handling

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

ECHO's privacy architecture ensures that user information remains private even when leveraging public blockchains. The design follows a "privacy by architecture" approach: sensitive data never leaves the user's device unencrypted, biometrics are bound to cryptographic keys via the device's Secure Enclave, and only opaque hashes and reference IDs are ever stored on-chain — making blockchain discovery attacks structurally impossible.

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

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a \[role\], I want to \[perform action\], so that I can \[achieve outcome\].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user \[performs action\], the system shall \[respond with specific behavior\].
* **AC-XXX-001.2:** When \[condition exists\], the system shall \[handle appropriately\].
* **AC-XXX-001.N:** \[Continue for all acceptance criteria\]

### REQ-XXX-002: Requirement Name

**User Story:** As a \[role\], I want to \[perform action\], so that I can \[achieve outcome\].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When \[condition\], the system shall \[behavior\].
* **AC-XXX-002.2:** \[Continue for all acceptance criteria\]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

### End-to-End Message Encryption and Commitment

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

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

### Privacy-Preserving Blockchain Data Model

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

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

### Zero-Knowledge Proofs and Midnight Integration

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

Zero-Knowledge Proofs enables ECHO users and organizations to verify attributes about themselves — age, trust tier, credential validity, token balance — without revealing the underlying data. Rather than sharing a credential (which exposes its content), users share a cryptographic proof that the credential satisfies a condition.

**Midnight / Cardano:** Not in scope. **Constellation-native **ZK circuits are the primary and expected long-term solution. Midnight (Cardano's ZK partner chain) may be considered at Phase 3+ only if a specific ZK use case cannot be satisfied by Constellation's architecture — this is considered unlikely. **ECHO may never use Cardano or Midnight.** The did:key standard ensures ECHO identities remain interoperable with any ecosystem, including Cardano, without ECHO depending on it.

**Why Constellation-native ZK first:** Constellation's Identity Metagraph already holds the trust tier commitments and credential states that ZK proofs need to verify against. Building ZK circuits that reference on-chain Constellation state eliminates the need for a cross-chain bridge to Cardano, reduces latency, and keeps all identity and proof operations within a single Scala/Euclid codebase. Proof generation occurs on-device (private inputs never leave the device); the verification key and public signals are published to the Identity Metagraph.

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

ECHO Tokenomics defines the strategic framework, conditions, and reference design for the ECHO token. As of PRD v3.0, **token genesis is deferred to Phase 3+ (36+ months)** and is conditional on meeting all five launch criteria. This is a deliberate strategic decision — not a deferral of vision, but a sequencing of execution.

The core change from v2.5.1: earn-by-chatting is explicitly deprecated. Users do not earn tokens for sending messages. This mechanic attracts mercenary users who churn when rewards feel small (validated by Brave/BAT experience: \~$0.01/hour earnings with low retention impact) and creates securities classification risk. Token utility in v3.0 is governance, node operation, and premium feature access — following the Session/SESH model.

The community-ownership vision remains intact. The path to it runs through Phase 1–2 execution: enterprise revenue, real user base, demonstrated utility.

## Token Genesis Conditions (All Five Must Be Met)

Token genesis shall not occur until all five of the following conditions are confirmed:

1. **User base threshold** — Minimum 100,000 active users across ECHO Comply and ECHO Message combined
2. **Revenue sustainability** — ECHO Comply generating $5M+ ARR, demonstrating the platform is viable without token-dependent economics
3. **Regulatory clarity** — Securities counsel confirmation that the token structure does not constitute an unregistered securities offering under current applicable law
4. **Demonstrable utility** — Clear, specific use cases for the token beyond speculation: governance voting, node operation staking, premium feature access
5. **Foundation board unanimous approval** — All five founders approve the genesis event

Until all five conditions are met, the Currency L1 layer remains dormant. Data L1 (integrity proofs, trust commitments) operates from Phase 1.

## Terminology

* **Genesis**: The single event where all 1,000,000,000 ECHO tokens are minted. Not at mainnet launch — at Phase 3+ when all conditions are met.
* **TokenLock**: A Tessellation v3 primitive locking ECHO for a defined period. Used for team vesting, user staking, and validator requirements. Enforced by Currency L1 Scala validation.
* **WithdrawLock**: A Tessellation v3 primitive creating a 14-day cooldown before locked tokens become transferable.
* **StakeDelegation**: A Tessellation v3 primitive allowing staked ECHO to be delegated to L1 validators for additional rewards.
* **Deprecated: AtomicAction for messaging rewards** — The v2.5.1 AtomicAction mechanic (verify tier + claim reward + update cap) for per-message earnings is deprecated in v3.0. No per-message token rewards.

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

* AC-TOK-004.1: At genesis, the system shall create five founder TokenLock positions with equal allocations: each of the 5 founders receives 36M ECHO (3.6% of total supply), totalling 180M ECHO (18%) across all founders.
* AC-TOK-004.2: All founder TokenLocks shall enforce a 12-month cliff — no tokens are withdrawable before the cliff date regardless of any other condition.
* AC-TOK-004.3: After the cliff, each founder's remaining allocation vests at 1/36th per month over 36 months (48-month total vesting period from genesis).
* AC-TOK-004.4: Vested tokens are subject to a 14-day WithdrawLock cooldown before becoming transferable.
* AC-TOK-004.5: Pre-cliff departure: the entire TokenLock balance is returned to the Future Team pool via 5-of-5 founder unanimous approval.
* AC-TOK-004.6: Post-cliff departure: vested tokens are released; unvested balance is returned to the Future Team pool via 5-of-5 unanimous approval.
* AC-TOK-004.7: All founder TokenLock positions (allocated, cliff date, vested, locked, monthly vest, all WithdrawLock transactions) shall be publicly visible on DAG Explorer from genesis.
* AC-TOK-004.8: The ECHO Wallet shall display a founder vesting panel (visible only to the DID holding a founder TokenLock) showing: allocated, vested, locked, next unlock date, cliff status, and a "View on DAG Explorer" link.
* AC-TOK-004.9: DAO transition acceleration: 50% of unvested founder tokens accelerate when ECHO transitions to full DAO governance (Phase 5–6), triggered by governance vote in L1 code.

### REQ-TOK-005: Treasury Allocation and Controls

**User Story:** As a community member, I want the treasury allocation clearly defined with spend controls, so that I know funds cannot be misappropriated before DAO governance is operational.

**Acceptance Criteria:**

* AC-TOK-005.1: The 220M treasury at genesis shall be subdivided as: 80M to PacaSwap liquidity seeding, 50M to operational reserve (bridged to stablecoins), and 90M locked in a 5-of-5 founder unanimous approval multi-sig for Phase 5–6 operations.
* AC-TOK-005.2: During Phases 1–3, treasury disbursements require unanimous 5-of-5 founder approval. From Phase 4 onward, disbursements require a governance vote.
* AC-TOK-005.2b: Future Team & Advisors pool disbursements (allocating tokens to new hires, advisors, or contractors) shall require Governance Board approval. No founder can unilaterally approve an allocation. Unallocated tokens after 3 years from genesis revert to treasury via governance vote.
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

All 5 founders receive equal allocations of 36M ECHO (3.6% of total supply each). Equal allocation reflects a founding partnership where all members share the same economic stake and the same governance standing. There is no CEO premium — each founder's contribution is valued equally in the token structure. This eliminates internal disputes over differential allocations and signals to the community that no single founder can dominate governance or economic decisions.

For treasury and critical decisions, all 5 founders must approve unanimously. This ensures no single founder or bloc of founders can act unilaterally — every significant decision requires full alignment. The insider total (founders 18% + future team 10% = 28%) stays below the industry average of 35–45% inclusive of VC allocation. Community + ecosystem retains 50% — the majority.

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
* AC-TOK-002.4: Daily individual earning caps shall be enforced per trust tier via AtomicAction: Tier 1 (10 ECHO/day), Tier 2 (25 ECHO/day), Tier 3 (50 ECHO/day), Tier 4 (100 ECHO/day), Tier 5 (150 ECHO/day).
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
* AC-TOK-007.2: Governance votes shall be weighted by the formula: `Governance Weight = StakedECHO × TrustTierMultipli`er, where multipliers are: Tier 1 = 0.0, Tier 2 = 0.5, Tier 3 = 1.0, Tier 4 = 1.5, Tier 5 = 2.0.
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

---

## Token Marketing Strategy

### Core Positioning: "Chat. Earn. Own."

ECHO's token marketing must bridge two audiences simultaneously: crypto-native users who understand token economics, and mainstream messaging users who have never held a token. The winning approach — validated across Brave, Aave, Kyber, and Decentraland — is to **lead with the benefit, not the mechanism**. Never say "earn ERC-20 tokens via on-chain AtomicAction primitives." Say "every message you send earns you a share of ECHO."

The single most important framing decision: ECHO tokens are **ownership shares in a co-operative**, not speculative investments. Users who chat are not gambling — they are earning a stake in something they use daily.

### Message Architecture by Audience

**Non-crypto users (primary growth audience)**

The goal is to make tokens feel like a loyalty program they already understand, just with more power. Avoid all blockchain jargon in first contact.

* Lead message: "Chat and earn ECHO — the more you use the app, the more you own of it."
* Utility hook: Tokens unlock premium features, pay for VIP subscriptions, and can be exchanged for real value.
* Ownership hook: "ECHO has no shareholders. Every dollar the app makes goes to a community treasury that token holders govern. You are the shareholder."
* Simplest possible mental model: ECHO tokens are like airline miles — except they give you a vote on where the airline flies, and a cut of its profits.

**Privacy-conscious users**

* Lead message: "Your data stays yours. Your keys, your messages, your identity. And you get paid for building the network."
* Ownership hook: Unlike WhatsApp, which sells your data to advertisers, ECHO pays you to use it.

**Crypto-native / DeFi users**

* Lead message: "L0 token on Constellation Hypergraph. Trust-tier weighted governance. 5-15% APY on TokenLock. PacaSwap liquidity pairs live at mainnet."
* Ownership hook: Full on-chain vesting visibility for founders. No VC allocation. Community retains 50% of supply. Treasury governed by staked ECHO × trust tier.

**Enterprise / institutional users**

* Lead message: "Your organization's communications generate compliance value. ECHO's Digital Evidence integration means every verified message is court-admissible — and Organization plan fees flow to a community treasury, not a corporate profit center."

### Communicating the Three Pillars of Value

**Pillar 1: Earn by Using**

Users need to understand that messaging activity generates real economic value for them — not just app credits. The key messages:

* "Every message earns ECHO tokens, automatically credited to your wallet."
* "Trust tier multipliers mean the more you verify your identity, the more you earn per message. Tier 5 users earn 3× more per message than Tier 1."
* "Your daily earnings appear in real time in the ECHO Wallet tab. No claiming required for messaging rewards — they accumulate automatically."
* Referral hook: "Refer a friend who completes identity verification and sends 100 messages — you both receive 50 ECHO. No cap on referrals, but each referral is anti-abuse verified by DID uniqueness."

**Pillar 2: Build Wealth by Staking**

Staking turns casual earners into committed community members. The key messages:

* "Lock your ECHO tokens for 30 days or more and earn 5–15% APY — on top of your messaging rewards."
* Staking tier ladder: Bronze (30d/5%) → Silver (90d/8%) → Gold (180d/12%) → Platinum (365d/15%).
* "The longer you commit, the more you earn. Platinum stakers earn 3× the APY of Bronze stakers."
* "Your staked tokens also give you governance weight — so staking is not just earning, it's having a say."
* Early adopter angle: "Token emission is front-loaded. Year 1 users earn 20% of the total 10-year community pool — more per message than any future cohort."

**Pillar 3: Own the Platform by Governing**

This is ECHO's most differentiated value proposition versus any other messaging app. The key messages:

* "Every dollar ECHO earns — from VIP subscriptions, Organization plans, payment rail fees, marketplace commissions — goes to a community treasury. None to shareholders. None to a parent company."
* "You govern that treasury. Token holders vote on how much goes to token burns (reducing supply, increasing value), how much goes to BTC reserves (hard-asset backing), and how much funds operations."
* "ECHO is not a company with users. It is a co-operative with owners. You are an owner."
* Network State hook (Phase 6+): "As the community grows, the treasury accumulates real-world assets — land, buildings, infrastructure — accessible to token holders. ECHO is building a community you can actually live in."

**Revenue sharing transparency table** (use in website, onboarding, and whitepaper):

| Revenue Source | Annual Estimate (Year 2) | Flows To |
| --- | --- | --- |
| VIP Subscriptions ($9.99/mo × 5% of 1M users) | \~$5.99M | Community Treasury |
| Organization Plans | \~$2.5M | Community Treasury |
| Payment Rail Fees (0.5-1.5%) | \~$1M | Community Treasury |
| Marketplace/Bot Revenue Share | \~$500K | Community Treasury |
| **Total (est. Year 2)** | **\~$10M** | **100% to token holders via governance** |

Treasury allocation (governance-set example): 30% ECHO burns, 30% BTC reserve, 20% operations, 15% real-world assets, 5% emergency reserve.

### Sample Messaging Copy

**Website hero / app store description:**

> ECHO is the messaging app you own. Every message earns you ECHO tokens. Every dollar ECHO makes goes to a community treasury you govern. No shareholders. No ads. No data harvesting. Just communication that pays you back.

**In-app first token notification:**

> You just earned your first ECHO tokens. These tokens give you voting power over the platform, a share of all revenue it generates, and staking rewards of up to 15% APY. Welcome to ownership.

**Social thread format:**

> 1/ ECHO is not a messaging app. It's a messaging co-op.\
> 2/ Every message → earns ECHO tokens → auto-scaling rewards, no caps.\
> 3/ All revenue (VIP, enterprise, fees) → community treasury.\
> 4/ Token holders vote on treasury allocation. Burn, BTC, operations.\
> 5/ Stake tokens → 5-15% APY + governance weight.\
> 6/ Phase 6: treasury buys real-world assets for community use.\
> 7/ The more ECHO grows, the more every token holder benefits. Always.

---

## Gamification Strategy

### Design Principle: Reward Real Network Value

Every gamification mechanic must reward behavior that genuinely benefits the ECHO network — authentic communication, trusted identity building, quality referrals, and network growth. Anti-gaming enforcement (DID-verified identity, trust tier requirements, AtomicAction atomic claims) ensures the system rewards real participants, not bots.

Research shows apps with gamification features see 20–30% higher engagement and 47% higher 90-day retention (Deloitte 2024). The ECHO gamification stack is designed around five proven mechanics.

### Mechanic 1: Usage Leaderboards

Weekly and monthly leaderboards ranked by a composite activity score, with token pool prizes distributed to top performers.

**Leaderboard categories:**

| Leaderboard | Scoring Formula | Prize Pool | Reset |
| --- | --- | --- | --- |
| Chat Champions | Messages × trust tier multiplier | 500 ECHO/week from Ecosystem pool | Weekly |
| Top Referrers | Verified referrals × quality score | 1,000 ECHO/month | Monthly |
| Community Builders | Group creation + member growth + moderation actions | 750 ECHO/month | Monthly |
| Staking Leaders | Total ECHO staked × lock duration (tier weight) | Bonus APY boost (0.5%) | Monthly |
| Trust Climbers | Fastest trust tier advancement | 250 ECHO/week + exclusive badge | Weekly |

**Leaderboard rules:**

* Minimum Trust Tier 2 to appear on any public leaderboard (prevents unverified bot accounts).
* All leaderboard scores are computed on-chain and publicly verifiable.
* Prize pools come from the Ecosystem & Partnerships allocation (10% of supply, 100M ECHO), not from the Community Rewards emission pool — preserving the messaging reward rate.
* "Share my rank" button generates a social card with the user's badge, rank, and earnings — organic viral loop.

### Mechanic 2: Referral Program

The referral program is the primary user acquisition engine. It is built for quality referrals — users who actually engage — rather than quantity.

**Referral reward structure:**

| Milestone | Referrer Reward | New User Reward | Trigger |
| --- | --- | --- | --- |
| Friend installs + completes DID onboarding | 25 ECHO | 25 ECHO | DID creation confirmed on Cardano |
| Friend sends first 100 messages | +25 ECHO bonus | +25 ECHO bonus | 100th message confirmed |
| Friend reaches Trust Tier 3 | +50 ECHO bonus | — | Tier 3 credential issued |
| 5 active referrals in 30 days | Streak bonus: +100 ECHO | — | 5th qualifying referral |
| 10 active referrals in 30 days | "Connector" badge + +250 ECHO | — | 10th qualifying referral |

Total maximum per referral: 100 ECHO (referrer) + 50 ECHO (referee) = 150 ECHO per quality referral pair.

**Anti-abuse:**

* Each DID can only be referred once (blockchain-verified uniqueness).
* New user must send ≥100 messages within 30 days for full reward to vest — prevents install-farm abuse.
* Rewards claimed via AtomicAction — atomic tier verification + claim + anti-fraud check in a single transaction.

### Mechanic 3: Quest System

Quests are short-term, structured activities that onboard users into deeper ECHO features while rewarding them with tokens and badges.

**Starter Quest Catalog (Phase 2 Launch):**

| Quest | Action Required | Reward | Badge |
| --- | --- | --- | --- |
| First Contact | Send your first 10 messages | 5 ECHO | "Chat Starter" |
| Identity Builder | Complete Cardano DID verification | 20 ECHO | "Verified" |
| Community Joiner | Join or create a group with 5+ members | 10 ECHO | "Group Member" |
| Trusted Messenger | Reach Trust Tier 3 | 50 ECHO | "Trusted" badge |
| Stack & Earn | Stake ECHO for the first time | 15 ECHO | "Staker" |
| Invite & Grow | Complete first successful referral | 25 ECHO | "Connector" |
| Vault Keeper | Send a disappearing message | 5 ECHO | "Ghost" |
| Private Circle | Activate a Hidden Folder | 10 ECHO | "Vault" |
| VIP Experience | Upgrade to VIP for first month | 50 ECHO cashback | "VIP" |
| Governance Debut | Cast first governance vote | 25 ECHO | "Voter" |

**Advanced Quests (Phase 3+):**

| Quest | Action Required | Reward |
| --- | --- | --- |
| Network Validator | Delegate to a validator for 30 days | 100 ECHO + "Delegator" badge |
| Whale Staker | Stake Platinum tier (365 days) | 500 ECHO + "Platinum" animated badge |
| 1K Club | Send 1,000 total messages | 100 ECHO + "1K Club" badge |
| Network Builder | Refer 10 active users | 500 ECHO + "Builder" badge |
| Bot Creator | Publish a bot to the marketplace | 200 ECHO + revenue share activation |

### Mechanic 4: Streak System

Daily streaks create habit loops that keep users returning every day. Streaks operate independently from the auto-scaling message reward — they are a bonus on top of regular earnings.

**Streak mechanics:**

* A streak increments when a user sends at least 5 messages within any 24-hour window.
* Missing a day resets the streak counter to zero.
* Streak multiplier increases the base reward rate for all messages sent on that day.

**Streak multiplier table:**

| Streak Length | Multiplier Bonus | Notes |
| --- | --- | --- |
| 1–6 days | Base (no bonus) | Getting started |
| 7 days | +10% on all message rewards | "Week Warrior" notification |
| 14 days | +20% on all message rewards | "Fortnight Fighter" badge |
| 30 days | +40% on all message rewards | "Monthly Master" badge + 50 ECHO bonus |
| 60 days | +60% on all message rewards | "Two-Month Titan" badge + 150 ECHO bonus |
| 100 days | +100% (2× base rate) | "Century Club" badge + 500 ECHO bonus + leaderboard feature |

### Mechanic 5: Trust Tier Progression as Gamification

The trust tier system is itself the deepest gamification mechanic — a long-term progression arc that unlocks real economic benefits at each tier.

**Trust Tier Progression Map:**

| Tier | Name | How to Reach | Key Unlock |
| --- | --- | --- | --- |
| Tier 1 | Unverified | Default | Basic messaging |
| Tier 2 | Newcomer | Install + first message | Governance voting (0.5× weight) |
| Tier 3 | Member | DID verified + 100 messages | 1.5× message reward, governance 1.0× weight |
| Tier 4 | Verified | ID verification + 500 messages + no violations | 2.0× message reward, financial institution features |
| Tier 5 | Trusted | 1,000+ messages + high community score + credential | 3.0× message reward, governance 2.0× weight, Platinum staking |

**Marketing this progression:** "ECHO gets more valuable the more you use it. Every level you reach multiplies what you earn. Tier 5 users earn 3× more per message and 4× more governance weight than Tier 1. Your investment of time and trust pays compounding returns."

### Gamification Anti-Gaming Framework

All gamification mechanics must survive adversarial behavior. Key anti-gaming rules:

* **DID uniqueness**: Referral and quest rewards require unique on-chain DID. One phone ≠ infinite referrals.
* **Activity quality gates**: Rewards require sustained activity (100 messages, 30 days staking) — not just completion of a one-time action.
* **Trust tier gates**: Higher-value rewards require Trust Tier 2+ — bots are locked out of governance, leaderboards, and bonus rewards.
* **AtomicAction enforcement**: All reward claims are atomic on-chain operations. No partial claims, no double-claims.
* **Rate limiting**: Backend enforces per-minute and per-hour message rate limits to prevent message-spamming bots.
* **Sybil resistance**: Blockchain-anchored DIDs with Cardano credential backing make Sybil attacks economically expensive.

### Gamification Metrics to Track

| Metric | Target | Measurement |
| --- | --- | --- |
| Quest completion rate (new users, 7 days) | &gt;60% complete at least 3 quests | In-app event tracking |
| Leaderboard participation | &gt;40% of MAU earn a leaderboard rank monthly | On-chain score records |
| Referral conversion rate | &gt;25% of invited friends become 100-message users | Referral dashboard |
| 7-day streak retention | &gt;35% of users hit 7-day streak | Backend streak counters |
| Trust Tier 3+ conversion | 30% of MAU at 6 months | On-chain tier records |
| Gamification-attributed DAU | &gt;50% of DAU driven by quest/streak activity | Cohort analysis |

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
* AC-PROD-001.19: API keys shall never be stored in source code or environment variable files committed to the repository. All secrets shall be stored in AWS Secrets Manager (or equivalent) and injected at runtime.

**UAT Sign-off Criteria:**

* AC-PROD-001.20: Production launch is gated by sign-off from: iOS engineer (app testing complete), backend engineer (load + integration tests passing), blockchain engineer (metagraph testnet lifecycle complete), and security review (no critical or high vulnerabilities open). All four sign-offs must be recorded before the production deployment checklist is initiated.

### REQ-PROD-002: Cloud Infrastructure Setup

**User Story:** As the engineering team, I want the Go relay backend deployed on scalable, cost-effective cloud infrastructure with full observability, so that ECHO can serve users reliably from day one and scale to 1M+ users without re-architecting.

**Acceptance Criteria:**

**Cloud Provider and Region:**

* AC-PROD-002.1: The initial production deployment shall use AWS (recommended) with primary region us-east-1. A secondary region (us-west-2) shall be configured for failover from Phase 3 onward. DigitalOcean is an acceptable lower-cost alternative for Phase 1-2 when user base is under 100K.
* AC-PROD-002.2: All infrastructure shall be defined as Infrastructure as Code using Terraform. No production resources shall be created manually through the AWS console. The Terraform state shall be stored in an S3 bucket with DynamoDB locking.

**Compute (Go Relay Backend):**

* AC-PROD-002.3: The Go relay service shall run on ECS Fargate (serverless containers) or EKS (Kubernetes) depending on team familiarity. Fargate is recommended for Phase 1-2 (simpler ops); EKS for Phase 3+ (more control at scale).
* AC-PROD-002.4: Initial production sizing shall be: 2 tasks × (2 vCPU, 4GB RAM) for the relay service. Auto-scaling shall be configured to add tasks when CPU &gt; 70% or WebSocket connection count &gt; 5,000 per task. Maximum auto-scale to 10 tasks before manual review is required.
* AC-PROD-002.5: The WebSocket relay shall be fronted by an AWS Application Load Balancer (ALB) with WebSocket support enabled. Sticky sessions shall NOT be used — the relay is stateless and any task can serve any connection.

**Estimated Monthly Cloud Costs:**

* AC-PROD-002.6: The team shall budget for the following estimated monthly costs at launch (100K users):

| Service | Specification | Est. Monthly Cost |
| --- | --- | --- |
| ECS Fargate (relay) | 2 tasks × 2vCPU/4GB, \~730 hrs/month | \~$120 |
| RDS PostgreSQL | db.t3.medium, Multi-AZ off (Phase 1-2) | \~$60 |
| ElastiCache Redis | cache.t3.micro | \~$25 |
| ALB | 1 load balancer + data processing | \~$25 |
| S3 (media/audit logs) | 100GB + data transfer | \~$15 |
| CloudWatch logs/metrics | Standard retention | \~$20 |
| ACM SSL certificates | Free | $0 |
| Route 53 (DNS) | Hosted zone + queries | \~$5 |
| Secrets Manager | \~20 secrets | \~$8 |
| **Phase 1-2 Total** | **\~$280/month** |  |

At 1M users (Phase 3+), estimated costs rise to $1,500–3,000/month depending on traffic patterns. This is covered by VIP subscription revenue well before reaching that scale.

**Database Setup:**

* AC-PROD-002.7: PostgreSQL (AWS RDS) shall be deployed with: automated daily backups with 7-day retention, encryption at rest (AES-256), VPC-isolated (no public internet access), read replica from Phase 3 onward. The database stores only non-sensitive relay metadata — all PII is on-device per the privacy architecture.
* AC-PROD-002.8: Redis (AWS ElastiCache) shall be deployed for: WebSocket connection state, message delivery queue (encrypted offline message blobs), session token cache, and reward claim rate limiting. Redis data is ephemeral — loss of Redis does not cause data loss; the system falls back to PostgreSQL.

**Networking and Security:**

* AC-PROD-002.9: All backend services shall run within a VPC with private subnets. Only the ALB shall be publicly accessible. The ECS tasks, RDS, and ElastiCache shall have no public IP addresses.
* AC-PROD-002.10: Security groups shall enforce: ALB accepts only 443 (HTTPS/WSS); relay tasks accept only traffic from ALB security group; RDS accepts only traffic from relay task security group; Redis accepts only traffic from relay task security group.
* AC-PROD-002.11: AWS WAF shall be attached to the ALB with rules for: rate limiting (100 req/sec per IP), common attack patterns (OWASP Top 10 managed rules), and geographic blocking if required by compliance.

**Observability:**

* AC-PROD-002.12: The following dashboards shall be live before launch: relay service (request rate, WebSocket connections, p50/p95/p99 latency, error rate), database (connections, query time, disk usage), Redis (memory usage, hit rate, eviction rate), and metagraph submission queue (pending, success rate, retry rate).
* AC-PROD-002.13: PagerDuty or equivalent alerting shall be configured for: relay error rate &gt;1% (page immediately), WebSocket connection count &gt;80% of capacity (warn), p99 latency &gt;2s (warn), RDS disk usage &gt;80% (warn), and any 5xx error spike &gt;10 in 1 minute (page immediately).
* AC-PROD-002.14: All application logs shall be shipped to CloudWatch Logs with a 30-day retention policy. Sensitive data (DIDs, message hashes) shall appear in logs; plaintext message content shall never appear in any log.

**Admin Access:**

* AC-PROD-002.15: AWS IAM shall be configured with least-privilege roles: a deploy role (used by CI/CD, can update ECS services and push images), a read-only role (for monitoring/debugging), and a break-glass admin role (MFA-required, full access, usage audited). No root account credentials shall be used after initial setup.
* AC-PROD-002.16: AWS Organizations shall be set up with separate accounts for dev, staging, and production environments. Cross-account deployment roles shall allow the CI/CD pipeline to deploy to all three accounts from a single pipeline account.

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
* AC-PROD-004.9: On merge to `main`, the pipeline shall: build a Docker image, tag it with the Git SHA, push to AWS ECR (Elastic Container Registry), and deploy to the `dev` ECS environment automatically.
* AC-PROD-004.10: Staging deployments shall use a Blue/Green strategy via AWS CodeDeploy. The new task set receives 10% of traffic for 5 minutes; if error rate remains <0.1%, traffic shifts to 100%. If error rate exceeds 0.1%, automatic rollback to the previous task set.
* AC-PROD-004.11: Production deployments shall follow the same Blue/Green pattern with an additional 15-minute canary phase (10% traffic) requiring explicit approval to proceed to 100%. Rollback shall be achievable in under 60 seconds by re-routing the ALB target group.
* AC-PROD-004.12: Database migrations shall run as a separate pipeline step before the new container version is deployed. Migrations must be backward compatible (the old code version must work with the new schema) to support zero-downtime Blue/Green deploys.

**Scala Metagraph CI/CD Pipeline:**

* AC-PROD-004.13: The Scala CI/CD pipeline shall execute on every pull request: `sbt test` (unit tests for all L1 validation logic), `scalafmt` (formatting), and a JAR build verification.
* AC-PROD-004.14: Metagraph deployments are higher-risk than backend deployments because L1 validation logic changes affect on-chain behavior. A staging metagraph (on Constellation testnet) shall receive every merge to `main`. Production metagraph updates shall require: all tests passing, 48 hours of staging validation, and 3-of-5 founder multi-sig approval (separate from GitHub approvals).
* AC-PROD-004.15: The metagraph deployment pipeline shall: build the JAR, copy to all validator nodes via SSH/SFTP (or S3 + node pull), restart the L1 validator services with a rolling restart (one node at a time to maintain consensus), and verify the new JAR is processing snapshots correctly before proceeding to the next node.
* AC-PROD-004.16: L1 validator node health shall be monitored continuously. If a node goes unhealthy during a rolling deploy, the pipeline shall halt and alert. Rollback restores the previous JAR from S3.

**Secrets and Configuration Management:**

* AC-PROD-004.17: All secrets (API keys, database passwords, private keys for deployment) shall be stored in AWS Secrets Manager and referenced by ARN in GitHub Actions workflows. No secrets shall appear in workflow YAML files, `.env` files committed to the repo, or CI logs.
* AC-PROD-004.18: Environment-specific configuration (relay endpoint URLs, PacaSwap contract addresses, Cardano network) shall be stored in AWS Parameter Store (non-secret) and injected into ECS tasks as environment variables at deploy time.
* AC-PROD-004.19: GitHub repository secrets shall store only: AWS deployment role ARN, Apple Developer team credentials (for Xcode Cloud/Fastlane), and Slack webhook URL for notifications. All application secrets live in AWS Secrets Manager, not GitHub.

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
* Deploy relay service to production ECS via Blue/Green
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

* **Go backend**: Blue/Green rollback — re-route ALB to previous task set in <60 seconds
* **Metagraph JAR**: Rolling node restart with previous JAR from S3 — \~10 minutes
* **iOS app**: Use App Store phased release pause (stops new downloads; existing users unaffected)
* **Blockchain state**: Cannot be rolled back (immutable). This is why testnet validation is mandatory before mainnet deployment. The only recovery for a bad blockchain state is a governance vote to patch the L1 logic forward.

### Cost Summary

| Environment | Monthly Cost | Notes |
| --- | --- | --- |
| Production cloud (Phase 1-2) | \~$280/month | Scales with users |
| Constellation L0 nodes (3) | \~$300–500/month | AWS/DigitalOcean servers |
| Cardano ADA fees | \~$500/month at 10K users | \~0.3 ADA/credential issuance |
| Apple Developer Program | $99/year | One-time annual |
| Apple IAP Commission | 15–30% of VIP revenue | Net $7.00–8.50 per subscriber |
| DAG staking (750K DAG) | Capital lockup, not expense | Recoverable, nodes earn rewards |
| **Total Monthly (pre-revenue)** | **\~$1,100–1,280/month** | Drops after delegation subsidizes DAG fees |

## ECHO Comply — Enterprise Compliance Messaging

## Overview

ECHO Comply is a secure, cryptographically verifiable messaging platform for regulated industries. It provides end-to-end encrypted team communication with built-in compliance controls: configurable retention policies, tamper-evident audit trails, eDiscovery export, identity verification, and message integrity proofs anchored to an immutable public ledger — without exposing message content to any server, blockchain, or administrator.

ECHO Comply serves three priority segments: healthcare (HIPAA), local government (FOIA/public records), and law firms (chain-of-custody, litigation hold, ethical wall). All three segments share the same underlying ECHO Protocol — the policy layer applied on top determines which compliance controls are active.

**Core value proposition:** Every conversation in ECHO Comply produces a cryptographic proof of integrity — who said what, when, verified by blockchain consensus — without any server ever seeing the message content. Compliance teams get audit trails. Employees get privacy. Both are guaranteed by math, not policy.

**Intentional design principle:** ECHO Comply must feel like a polished messaging app, not a compliance tool. The compliance features are invisible to regular users — they exist in the admin console and export tools. If the app feels like compliance software, adoption dies. The messaging experience is clean, fast, and familiar.

## Terminology

* **ECHO Comply**: The enterprise product built on ECHO Protocol. Distinct from ECHO Message (consumer product) and ECHO Protocol (shared infrastructure).
* **BAA (Business Associate Agreement)**: A HIPAA-required contract between a covered entity (healthcare organization) and a business associate (ECHO Comply LLC) governing the handling of protected health information (ePHI).
* **ePHI (Electronic Protected Health Information)**: Any health information in electronic form that identifies or could identify a patient, subject to HIPAA privacy and security rules.
* **eDiscovery**: The process of identifying, preserving, collecting, and producing electronically stored information (ESI) in response to a legal request or regulatory inquiry.
* **Litigation Hold**: A directive to preserve all potentially relevant information once litigation is reasonably anticipated. Suspends normal document retention schedules.
* **Ethical Wall**: An information barrier preventing communication between designated user groups within the same organization — used in law firms to prevent conflicts of interest.
* **FOIA (Freedom of Information Act)**: Federal and state laws requiring government agencies to disclose records upon request, subject to specific exemptions. All 50 US states have analogous public records laws.
* **Merkle Proof**: A cryptographic proof that a specific message was included in a batch anchored to the blockchain, verifiable by any third party without ECHO Comply access.
* **Dual-Mode Architecture**: ECHO Protocol's design where the same relay infrastructure serves both ECHO Comply and ECHO Message, with behavior governed by a policy configuration layer.

## Requirements

### REQ-COMPLY-001: Core Encrypted Messaging

**User Story:** As an employee at a regulated organization, I want to communicate via a fast, familiar messaging app where compliance controls are automatic and invisible, so that I adopt secure channels without changing my behavior.

**Acceptance Criteria:**

* AC-COMPLY-001.1: All messages shall be end-to-end encrypted on the sender's device using X25519 key agreement and XChaCha20-Poly1305 before transmission. Relay servers shall see only opaque ciphertext under all circumstances and deployment modes.
* AC-COMPLY-001.2: The user experience shall support: 1:1 and group messaging, media sharing (images, documents, voice notes), reactions, read receipts, @mentions, message search, and delivery notifications — matching the feature set users expect from consumer messaging apps.
* AC-COMPLY-001.3: The ECHO Comply iOS app and Android app shall share the same messaging feature set. Web access (Phase 2+) shall be available via a browser-based client.
* AC-COMPLY-001.4: Multi-factor authentication shall be enforced for all users — biometric (Secure Enclave / StrongBox) plus device-bound credential. No single-factor login shall be permitted.
* AC-COMPLY-001.5: Messages shall be delivered via the stateless ECHO Protocol relay. The relay shall maintain no persistent message content — only encrypted blobs in the offline queue, deleted after delivery.

### REQ-COMPLY-002: Integrity Proofs and Audit Trail

**User Story:** As a compliance officer, I want every message to produce a tamper-evident integrity proof I can present to auditors, so that I can demonstrate our communications meet regulatory requirements without accessing message content.

**Acceptance Criteria:**

* AC-COMPLY-002.1: Every message shall produce a cryptographic commitment (Merkle root) anchored to the Constellation metagraph Data L1. The commitment proves the message existed at a specific time, was sent by a specific DID, and has not been altered — without revealing content.
* AC-COMPLY-002.2: Audit trail entries shall be immutable. No administrator, ECHO Comply operator, or third party shall be able to delete, modify, or suppress an audit trail entry once created.
* AC-COMPLY-002.3: Each audit trail entry shall contain: sender DID, recipient DID(s), timestamp (block time), message hash, Merkle path, metagraph snapshot reference, and channel ID. Message content shall never appear in the audit trail.
* AC-COMPLY-002.4: Any authorized party shall be able to independently verify an integrity proof against the public metagraph without requiring ECHO Comply access or cooperation.
* AC-COMPLY-002.5: The admin console shall display integrity verification status for all messages — showing the on-chain anchor confirmation, snapshot reference, and proof validity.

### REQ-COMPLY-003: Admin Console

**User Story:** As an organization administrator, I want a web-based console to manage users, channels, retention policies, and compliance exports, so that I can maintain regulatory compliance without requiring technical expertise.

**Acceptance Criteria:**

* AC-COMPLY-003.1: A web-based admin console (React) shall provide: user provisioning and deprovisioning, role assignment, trust tier management, channel creation and archiving, retention policy configuration, compliance dashboard, and export tools.
* AC-COMPLY-003.2: Admins shall see message metadata (sender, recipient, timestamp, channel, delivery status) and integrity proof status. Admins shall never see message content — content decryption requires a separate organizational key ceremony.
* AC-COMPLY-003.3: The compliance dashboard shall display: message volume by channel, retention policy compliance status, pending litigation holds, integrity verification errors, and export history.
* AC-COMPLY-003.4: User management shall support: bulk provisioning via CSV import, role-based access control (admin, compliance officer, standard user, read-only), and deprovisioning with configurable message retention rules.

### REQ-COMPLY-004: Retention Policies

**User Story:** As a compliance officer, I want configurable message retention policies per channel and message type, so that I meet regulatory retention minimums while not storing data beyond what is legally required.

**Acceptance Criteria:**

* AC-COMPLY-004.1: Retention policies shall be configurable per channel type (e.g., clinical, administrative, general) and per organization, with minimum retention periods enforced (not reducible below the organization's regulatory minimum).
* AC-COMPLY-004.2: Retention operates on the encrypted blob level — content remains encrypted and the integrity proof (Merkle commitment) is preserved on-chain indefinitely regardless of blob retention policy.
* AC-COMPLY-004.3: The system shall support: permanent retention, time-limited retention (e.g., 7 years for HIPAA, 3–7 years for government), and litigation hold (indefinite, superseding all other retention rules).
* AC-COMPLY-004.4: At retention expiry (where applicable), encrypted blobs shall be securely deleted from relay storage. The on-chain commitment and metadata remain permanently — this satisfies both the "data minimization" and "audit trail" requirements simultaneously.

### REQ-COMPLY-005: SSO and Identity Integration

**User Story:** As an IT administrator, I want ECHO Comply to integrate with our existing identity provider, so that employee onboarding and offboarding is automatic and we don't manage a separate credential system.

**Acceptance Criteria:**

* AC-COMPLY-005.1: ECHO Comply shall support SAML 2.0 and OIDC integration for enterprise SSO with major identity providers: Azure Active Directory, Okta, Google Workspace, and Ping Identity.
* AC-COMPLY-005.2: SCIM 2.0 provisioning shall support automated user lifecycle management — employees who join the organization are automatically provisioned; employees who leave are automatically deprovisioned with configurable message retention.
* AC-COMPLY-005.3: Upon SSO authentication, a Cardano DID shall be automatically provisioned for the employee — they do not manually create or manage a DID.
* AC-COMPLY-005.4: Organization credentials (role, department, professional license) shall be issuable from the admin console, attaching to the employee's DID as verifiable credentials.

### REQ-COMPLY-006: eDiscovery and Legal Export

**User Story:** As a compliance officer or legal counsel, I want to export a structured package of messages with their integrity proofs in response to legal, regulatory, or public records requests, so that I can satisfy discovery obligations efficiently.

**Acceptance Criteria:**

* AC-COMPLY-006.1: The admin console shall support eDiscovery exports filtered by: user, date range, channel, message type, and keyword search (over message metadata only — not content decryption).
* AC-COMPLY-006.2: An eDiscovery export package shall include: audit trail entries for the requested scope, Merkle proofs for all included messages, metagraph snapshot references, export request metadata (who requested, when, scope), and an integrity proof that the export package itself is complete and unmodified.
* AC-COMPLY-006.3: For organizations that manage their own organizational content decryption key, the export package may include encrypted content blobs. Content decryption requires the organizational key — ECHO Comply shall never hold or transmit organizational content keys.
* AC-COMPLY-006.4: Public records (FOIA) exports for government customers shall produce a structured package with message metadata and integrity proofs. Private channels (personnel matters, attorney-client) shall require additional authorization before inclusion.

## Feature Behavior and Rules

### Compliance is Invisible to End Users

The most important behavioral rule for ECHO Comply is that compliance controls are not visible to the employees sending messages. Retention policies, audit trails, eDiscovery exports, and integrity proofs operate at the infrastructure layer. A clinician sending a message to a colleague sees a familiar messaging interface — not a compliance interface. The admin console is where compliance officers manage all regulatory controls.

### Content-Blind Architecture Is Non-Negotiable

ECHO Comply administrators see message metadata and integrity proofs — never message content. This is both a privacy protection for employees and a security architecture decision: if ECHO Comply servers cannot decrypt content, a server compromise yields only encrypted blobs. Organizational content decryption (if needed for legal purposes) requires a separate key ceremony managed by the customer organization, not ECHO Comply.

### Pricing Tiers

| Plan | Price/Seat/Month | Minimum Seats | Target |
| --- | --- | --- | --- |
| Comply Starter | $30 | 10 | Small clinics, municipal offices, boutique law firms |
| Comply Professional | $50 | 50 | Mid-size healthcare systems, county government |
| Comply Enterprise | $80–100 | 500 | Hospital systems, state agencies, AmLaw 200 |

All plans include: E2E encryption, Merkle integrity proofs, admin console, audit trail, configurable retention, MFA, and signed BAA (healthcare) or compliance addendum (government/legal). Enterprise adds: SSO/SAML, SCIM provisioning, dedicated support, 99.9% uptime SLA, on-premises relay option.

---

## v3.1 Amendment — Organization, Identity, and Onboarding Requirements

The following requirement families were added in PRD v3.1 (April 24, 2026). They introduce the shared-DID architecture, self-serve admin signup, tiered seat enforcement, Verifiable Credential-based organization membership, and the iOS Personal-Business context-switching model.

**Core architectural decision (v3.1):** A user has exactly one personal Cardano DID. Organization membership is expressed as a Verifiable Credential (type: EchoOrgRoleCredential) attached to that DID, issued by the organization's own DID, and revocable via W3C StatusList2021 published to the Cardano blockchain. When a user is offboarded, their credential is revoked and Business context access disappears — their personal ECHO account is completely untouched. This is a genuinely novel architecture for a messaging product: Slack uses separate workspace accounts; Microsoft Teams conflates personal and work identity; ECHO gives every user one sovereign identity with organization access expressed as a portable, revocable credential.

### REQ-COMPLY-ORG-001: Admin Self-Serve Signup Flow (Starter Tier)

**User Story:** As a healthcare organization admin, I want to create an ECHO Comply organization, sign the BAA, and invite my team entirely through a self-serve web flow — with no sales call required — so that I can be operational in under 10 minutes.

**Acceptance Criteria:**

* AC-COMPLY-ORG-001.1: Screen 1 shall collect work email, full name, organization name, and role (dropdown). Email verification shall use a six-digit code (not a magic link) to reduce mobile drop-off.
* AC-COMPLY-ORG-001.2: Screen 2 shall generate the admin's Cardano DID, present the Emergency Recovery Kit for download, and require an acknowledgment checkbox before proceeding.
* AC-COMPLY-ORG-001.3: Screen 3 shall present plan selection with inline Stripe credit card capture, a six-bullet plain-English BAA summary above the full PDF scroll viewer, and a BAA acceptance checkbox distinct from the Terms of Service checkbox.
* AC-COMPLY-ORG-001.4: Screen 4 shall present three pre-selected compliance defaults — 90-day message retention, 6-year audit log retention (non-configurable), and required MFA for all users — with a "Use these defaults" primary CTA and a "Customize now" secondary link.
* AC-COMPLY-ORG-001.5: Screen 5 shall present a paste-multi invite box accepting comma or newline-separated emails with live validation and a collapsible role assignment panel defaulting to Member.
* AC-COMPLY-ORG-001.6: Screen 6 shall land the admin in the Compliance and Security Hub with a first-week checklist showing completed items (BAA signed, identity verified, teammates invited, compliance defaults confirmed) and remaining items (SSO configured, org logo, role-based routing, audit log export, sub-processor list reviewed).
* AC-COMPLY-ORG-001.7: Total elapsed time from "clicks Sign Up" to "first invitation sent" shall not exceed 8 minutes for a typical admin, measured by analytics timer events at each screen transition.

### REQ-COMPLY-ORG-002: Tiered Seat Enforcement with Enterprise Grace Period

**User Story:** As an ECHO Comply customer who is growing, I want clear seat-tier boundaries with a grace period that prevents my team from being locked out during Enterprise contract negotiations, so that my growth doesn't disrupt operations.

**Acceptance Criteria:**

* AC-COMPLY-ORG-002.1: Starter tier (1–10 seats): a yellow dismissible banner appears at 8/10 seats used; a blocking modal appears on the 11th-seat attempt offering one-click upgrade to Professional.
* AC-COMPLY-ORG-002.2: Professional tier (11–499 seats): no hard cap; a soft sales trigger fires at seat 50 (CRM entry, no user interruption); a dashboard nudge card appears at seat 200 suggesting Enterprise features.
* AC-COMPLY-ORG-002.3: Enterprise tier (500+ seats): a full-screen blocking interstitial appears on the 500th-seat attempt, but the customer's existing Professional access continues uninterrupted for 30 days while sales negotiation occurs, with a day-count indicator visible in the admin console.
* AC-COMPLY-ORG-002.4: The Enterprise handoff form shall collect at most 5 fields (name, work email, organization, estimated seat count, primary compliance framework) with "We'll email within 4 business hours" as an explicit SLA commitment.
* AC-COMPLY-ORG-002.5: BAA signing, basic audit log viewing, mandatory MFA policy, and end-to-end encryption shall be available at all tiers including Starter — never gated behind Professional or Enterprise.

### REQ-COMPLY-ORG-003: BAA Execution and Lifecycle

**User Story:** As a healthcare compliance officer, I want the Business Associate Agreement executed as a first-class signed artifact — not buried in Terms of Service — with its status persistently visible in the admin console, so that I can demonstrate HIPAA compliance to auditors at any time.

**Acceptance Criteria:**

* AC-COMPLY-ORG-003.1: The BAA screen shall present a six-bullet plain-English summary above the full PDF, covering: PHI scope, 6-year audit retention, breach notification SLA, sub-processor flow-down, termination mechanics, and covered-entity self-attestation.
* AC-COMPLY-ORG-003.2: The signed BAA shall be cryptographically timestamped with the admin's DID signature, stored indefinitely, and downloadable as PDF from the admin console at any point.
* AC-COMPLY-ORG-003.3: The Compliance and Security Hub shall display a persistent "BAA Active" status pill on every admin console page, showing signer name, signature date, and a download button.
* AC-COMPLY-ORG-003.4: BAA scope shall explicitly cover message content, attachments, metadata (sender-recipient-time), audit logs, and role credentials issued within the organization.
* AC-COMPLY-ORG-003.5: Subscription cancellation shall trigger a 30-day return-or-destroy window during which the admin can export all PHI, followed by cryptographic key shredding with a signed destruction attestation emailed to the admin and filed in the immutable audit log. Audit logs themselves shall be retained for 6 years post-termination per 45 CFR 164.316(b)(2)(i).

### REQ-COMPLY-ORG-004: SSO Tiered Availability

**User Story:** As an IT administrator adopting ECHO Comply, I want SSO available at the appropriate tier with sensible defaults, so that I can enforce identity governance without being blocked by licensing requirements I don't have.

**Acceptance Criteria:**

* AC-COMPLY-ORG-004.1: Starter tier: password + Google and Microsoft OIDC social login only. No SAML, no SCIM.
* AC-COMPLY-ORG-004.2: Professional tier: SAML 2.0 with Okta, Entra ID, Google Workspace, and generic SAML/OIDC connectors. SCIM available as opt-in recommended for 100+ seats.
* AC-COMPLY-ORG-004.3: Enterprise tier: SCIM provisioning configured during CSM-led onboarding in the first two weeks.
* AC-COMPLY-ORG-004.4: Professional customers who have not configured SSO after 7 days shall see a persistent amber banner: "Single sign-on isn't configured. Your team is signing in with passwords. Set up SSO to meet HIPAA administrative safeguards." The banner is dismissible per-session but reappears until SSO is configured or the admin explicitly opts out via a checkbox-confirmed modal.
* AC-COMPLY-ORG-004.5: Default provisioning mode shall be JIT via SAML (not hard-requiring SCIM) because many healthcare organizations on Azure AD lack Entra P1 / M365 E3 licensing required for SCIM.
* AC-COMPLY-ORG-004.6: Password login shall always remain as an escape hatch when SSO is erroring, preventing lockout from certificate expirations.

### REQ-COMPLY-ORG-005: First-Week Admin Checklist (Compliance and Security Hub)

**User Story:** As a new admin, I want a compliance-focused home screen with a guided setup checklist, so that I can reach a good security posture in my first week without missing critical steps.

**Acceptance Criteria:**

* AC-COMPLY-ORG-005.1: The Compliance and Security Hub's top card shall be the BAA status card (active, signer, date, download) on all admin pages, establishing compliance-first visual priority.
* AC-COMPLY-ORG-005.2: The setup checklist shall contain 9 items: BAA signed, identity verified, first teammates invited, compliance defaults confirmed (auto-complete at signup), then SSO configured, org logo and VC branding, role-based routing, audit log export destination, sub-processor list reviewed (persistent nudges until completed or dismissed).
* AC-COMPLY-ORG-005.3: No product functionality shall be gated on checklist completion. The checklist is a nudge, not a lock.
* AC-COMPLY-ORG-005.4: The Hub shall surface SOC 2 Type II and HIPAA attestation reports as downloadable documents with clickwrap-NDA protection and per-download watermarking.
* AC-COMPLY-ORG-005.5: The security posture card shall display a one-liner: "AES-256 at rest, Signal Protocol E2EE, TLS 1.3 in transit, US-East primary with US-West DR" in plain text so procurement reviewers can scan it without drilling into certifications.

---

### REQ-COMPLY-VC-001: Organization as Cardano DID Issuer

**User Story:** As a system architect, I want each ECHO Comply organization to have its own Cardano DID that acts as the issuer of all member credentials, so that organization membership is cryptographically verifiable by any third party without depending on ECHO's infrastructure.

**Acceptance Criteria:**

* AC-COMPLY-VC-001.1: When an admin commits the organization name during signup (Screen 1), the backend shall mint a new organization DID of the form `did:cardano:mainnet:org:{uuid}`, registered with the organization's legal name, headquarters address, and the admin's DID listed as the initial Owner.
* AC-COMPLY-VC-001.2: Domain verification via DNS TXT record shall be supported during SSO setup, attesting did-web linkage (e.g., `did:web:mercyhealth.org` resolves to the same organization DID).
* AC-COMPLY-VC-001.3: The organization DID's signing key shall be held in an HSM-backed key store operated by ECHO Comply on behalf of the organization, with key rotation possible via a quorum of Owner-role members.
* AC-COMPLY-VC-001.4: The organization DID document shall include branding metadata (primary color, logo URL, display name) used by iOS clients to theme the Business context when rendering credentials issued by this DID.

### REQ-COMPLY-VC-002: Organization Membership Credential Schema

**User Story:** As an organization admin, I want to issue standardized, cryptographically-signed membership credentials to my team, so that their access is portable, machine-verifiable, and revocable without affecting their personal ECHO identity.

**Acceptance Criteria:**

* AC-COMPLY-VC-002.1: Each member credential shall be of type EchoOrgRoleCredential with claims for: role (Admin, Member, Guest), department (optional), employee ID (optional), issuance timestamp, expiration timestamp (null for Admin/Member, 90-day default for Guest), and the revocation index in the organization's StatusList2021.
* AC-COMPLY-VC-002.2: Credentials shall be signed with the organization's Ed25519 key and serialized as W3C VC 2.0 with JSON-LD context.
* AC-COMPLY-VC-002.3: Credential display metadata shall include the organization's primary color and logo URL, enabling iOS to theme the credential presentation without an extra network lookup.
* AC-COMPLY-VC-002.4: Credential issuance shall be atomic: wallet storage on the member's device and StatusList2021 bit allocation on-chain both succeed or both roll back.
* AC-COMPLY-VC-002.5: Each credential shall include an org-scoped display name (e.g., "Dr. Jane Smith") that may differ from the user's personal ECHO display name, enabling healthcare "LastName, FirstName" naming conventions without affecting personal identity.

### REQ-COMPLY-VC-003: StatusList2021 Revocation

**User Story:** As an admin who needs to offboard a team member, I want their credential revoked immediately via a publicly verifiable on-chain mechanism, so that access is removed within seconds and any third party can verify the revocation independently.

**Acceptance Criteria:**

* AC-COMPLY-VC-003.1: Each organization shall maintain a W3C StatusList2021 bit vector of at least 131,072 bits (16KB) published to the Cardano blockchain, covering up to 131,072 credentials per list.
* AC-COMPLY-VC-003.2: The admin UI's Revoke action shall flip the target credential's bit in the StatusList2021 via a single Cardano transaction signed by the organization DID.
* AC-COMPLY-VC-003.3: Verification clients (iOS app, Go backend) shall check StatusList2021 freshness with a maximum 2-minute cache.
* AC-COMPLY-VC-003.4: Revocation propagation shall reach 99th-percentile client visibility within 5 seconds of the on-chain transaction confirming, measured from block timestamp to first status change observed by a sampled verifier.
* AC-COMPLY-VC-003.5: The admin UI shall present a confirmation modal before revocation with required reason selection, optional litigation-hold toggle, and a type-to-confirm "REVOKE" field to prevent accidental offboarding.

### REQ-COMPLY-VC-004: Member-Side Credential Card Presentation

**User Story:** As a member, I want to see my organization credential as a visual card in my profile, so that I understand what access I hold and can verify it is genuinely on-chain rather than a platform-asserted claim.

**Acceptance Criteria:**

* AC-COMPLY-VC-004.1: The credential card shall be 320×480 points with a gradient background using the organization's primary color, a status pill (Active, Suspended, Revoked, Expired, Unverified Issuer), the organization logo, and a three-claim display (role, department, employee ID).
* AC-COMPLY-VC-004.2: The card shall include metadata rows for: issuer DID with domain-verified indicator, issuance and expiration timestamps, credential type and version, StatusList2021 reference with last-checked timestamp and recheck button, and signature validity indicator.
* AC-COMPLY-VC-004.3: The card shall include a "View on Cardano explorer" deep link and a "View raw JSON" expander, enabling technical buyers (CISOs, security auditors) to verify the credential is genuinely on-chain.
* AC-COMPLY-VC-004.4: Revoked credentials shall render with strikethrough on the role claim, gray card background, and red status pill — remaining visible in the member's profile for audit purposes rather than disappearing.

### REQ-COMPLY-VC-005: Revocation User Experience

**User Story:** As a member who has been offboarded from an organization, I want the app to clearly separate the loss of my work access from my personal account, so that I am not confused and do not feel my personal identity has been compromised.

**Acceptance Criteria:**

* AC-COMPLY-VC-005.1: On next app foreground following revocation, the Business context pill in the iOS segmented control shall gray out with a lock icon, and a full-screen modal shall explain "Your access to {Organization} was revoked on {date}. You can no longer send or read {Organization} messages. Your personal ECHO account and contacts are not affected."
* AC-COMPLY-VC-005.2: Message history within the revoked organization's channels shall remain read-only on the member's device for 72 hours to allow personal export, after which the local cache evicts.
* AC-COMPLY-VC-005.3: The Personal context pill shall remain fully functional with no degradation — demonstrating the separation clearly to the revoked user as reassurance.
* AC-COMPLY-VC-005.4: The member's Cardano DID shall be completely untouched. Revocation is exclusively a credential-status change on the organization's StatusList2021 — no lifecycle event is emitted against the personal identity.

---

### REQ-COMPLY-INVITE-001: Invited-Member Onboarding

**User Story:** As an admin inviting a colleague, I want the invitation flow to smoothly handle both new users (who need to create an ECHO account first) and existing users (who just need to accept the credential), so that neither group is confused by an experience designed for the other.

**Acceptance Criteria:**

* AC-COMPLY-INVITE-001.1: A pre-check endpoint (`GET /v1/comply/invitations/{token}/resolve`) shall return whether the invited email resolves to an existing ECHO DID, returning `user_status: "new"` or `"existing"`.
* AC-COMPLY-INVITE-001.2: New users shall complete the Consumer first-run flow (display name, DID creation, silent provisioning) and then see a single "Join {Organization}" screen presenting the credential to be issued.
* AC-COMPLY-INVITE-001.3: Existing users shall authenticate with their existing passkey, then see a two-column "will see / will NOT see" transparency table before accepting. The left column lists what the organization can see (work display name, work role, messages in org channels, files in org channels). The right column lists what the organization cannot see (personal ECHO messages or DMs, personal contacts, phone number, activity outside the org).
* AC-COMPLY-INVITE-001.4: The transparency screen shall include an editable work display name (pre-filled from the invitation) and an optional work profile photo distinct from the user's personal avatar.
* AC-COMPLY-INVITE-001.5: Reassurance copy shall use absolute language where absolute is true ("Your employer cannot see...") rather than hedging language ("may not see...").

### REQ-COMPLY-COMPOSE-001: Misposting Guardrails

**User Story:** As a user who participates in both Personal and Business contexts, I want guardrails in the message composer that prevent me from accidentally sending a personal message to a work channel (or vice versa), so that I don't accidentally expose private information in professional communications.

**Acceptance Criteria:**

* AC-COMPLY-COMPOSE-001.1: A yellow informational banner above the composer reading "You're messaging as {name} in {context}" shall appear for exactly the first 10 sent messages following any context switch, then auto-dismiss permanently for that context until the user switches out and back.
* AC-COMPLY-COMPOSE-001.2: Cross-context contact autocomplete shall be blocked at the database query level — not the display layer — to prevent Personal contact names from leaking into Business context autocomplete results.
* AC-COMPLY-COMPOSE-001.3: iOS clipboard origin detection (iOS 14+ UIPasteboard) shall trigger a one-time modal "This content came from your personal context. Send anyway?" when the user pastes text longer than 200 characters or an image copied in a different context. The check shall use native UIPasteboard detection without reading clipboard contents.

---

### REQ-MSG-003 Amendment (v3.1) — iOS Personal-Business Context Switching

These three acceptance criteria are additions to REQ-MSG-003 (Self-Sovereign Identity), added in the v3.1 Comply Amendment to cover what happens when a Consumer user joins an organization and the iOS app must present two contexts distinctly.

* **AC-MSG-003.9:** The iOS app shall display a persistent identity header band (44pt height) at the top of the Messages tab, containing the user's avatar, display name, and a context glyph (person icon for Personal, shield icon with organization logo chip for Business). Immediately below, a 36pt segmented control shall present Personal and organization pills, with additional pills for users who belong to multiple organizations. Switching contexts shall trigger full-chrome color re-theming using the organization's branding metadata from the membership credential, with a 200ms crossfade transition and soft haptic feedback.
* **AC-MSG-003.10:** The iOS app shall implement three composer-level misposting guardrails active for all users holding at least one organization credential: (1) the post-switch banner per REQ-COMPLY-COMPOSE-001.1, (2) cross-context contact autocomplete blocking per REQ-COMPLY-COMPOSE-001.2, and (3) clipboard origin detection per REQ-COMPLY-COMPOSE-001.3.
* **AC-MSG-003.11:** When the user is viewing a Business context with no conversations, the empty state shall read: "Welcome to {Organization}, {name}. Your admin can help you get started, or tap + to start a new conversation with a teammate." — replacing the Consumer empty state for that context only.

---

## Architectural Notes (v3.1)

### Shared-DID Architecture

The v3.1 amendment ratifies the single most important architectural decision for the Comply-Consumer intersection: one Cardano DID per user. Organization membership is credential-based, not account-based. This means:

* A user moving from one organization to another keeps their DID, trust tier, and personal chat history
* An organization revoking a member's credential does not touch the member's personal identity in any way
* The iOS app presents one account with context-switching, not two separate accounts
* Trust tier in Personal context and Business context are independent: a Tier 1 personal user who holds an Admin-role credential is Tier 4 in Business context for the purposes of that organization's permissions

### Admin Read Access Posture

ECHO Comply maintains a strict no-admin-read-access posture. Admins see message metadata (sender, recipient, timestamp, channel, delivery status) and integrity proof status in audit logs — never message content. Content decryption in a legal context requires a separate organizational key ceremony managed by the customer organization. This is a feature, not a gap: a breach of ECHO Comply servers yields only encrypted blobs. This must be communicated explicitly to procurement reviewers because it differs from Slack and Microsoft Teams, where admins can read message content.

### Deferred Items (Tracked for v3.2)

Three items were documented in the v3.1 amendment but explicitly deferred:

1. **Multi-facility organization hierarchy** — large health systems with sub-organizations (e.g., Mercy Health nationally, Mercy Health St. Louis) need independent admin control per sub-org. Deferred to a future REQ-COMPLY-HIERARCHY spec, targeted for Phase 3 Enterprise deployments.
2. **VoCera and PagerDuty native integrations** — on-call workflow integration for Enterprise healthcare. Deferred to REQ-COMPLY-INTEGRATIONS, targeted when early Enterprise customers have this as a hard requirement.
3. **Break-glass emergency-access procedures** — HIPAA requires documented emergency-access procedures. For E2EE, true break-glass into message content is impossible without weakening encryption. A future REQ-COMPLY-BREAKGLASS spec will formalize the documentation admins can generate to satisfy the HIPAA requirement while honestly representing the E2EE constraint. Targeted for v3.2.

### ECHO Comply — Healthcare (HIPAA)

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

ECHO Comply — Healthcare provides HIPAA-compliant secure messaging for healthcare organizations. The 2026 HIPAA Security Rule updates mandate encryption for all ePHI at rest and in transit, require multi-factor authentication, and impose 24-hour breach reporting for business associates. Healthcare organizations are actively replacing legacy pager systems, SMS, and non-compliant consumer apps before enforcement begins in late 2026 / early 2027. ECHO Comply is designed to meet these requirements natively while delivering a clinical communication experience fast enough to replace personal phone usage for patient coordination.

## Terminology

* **ePHI**: Electronic Protected Health Information — any health information in electronic form that could identify a patient.
* **BAA**: Business Associate Agreement — HIPAA-required contract signed between ECHO Comply LLC and each healthcare customer.
* **OCR**: Office for Civil Rights — the HHS enforcement body for HIPAA violations. OCR investigations are triggered by breach reports and complaints.
* **Role-Based Routing**: The ability to address a message to a clinical role (e.g., "on-call cardiologist") rather than a specific named individual, enabling coverage-aware delivery.
* **Escalation Workflow**: An alert-until-read mechanism that escalates an unread urgent message to backup recipients after a configurable timeout.

## Requirements

### REQ-HC-001: HIPAA-Compliant Messaging

**User Story:** As a healthcare compliance officer, I want messaging that meets 2026 HIPAA Security Rule requirements natively, so that I can replace staff's use of personal phones for patient coordination without creating new compliance risk.

**Acceptance Criteria:**

* AC-HC-001.1: All messages containing or referencing ePHI shall be end-to-end encrypted using Secure Enclave key management. No server-side decryption capability shall exist under any deployment mode or administrative action.
* AC-HC-001.2: ECHO Comply LLC shall provide a signed Business Associate Agreement (BAA) to all healthcare customers before any data is processed. The BAA shall address breach notification, permitted uses, and data safeguards in compliance with 45 CFR Part 164.
* AC-HC-001.3: Multi-factor authentication shall be mandatory for all users — biometric (iOS Secure Enclave Face ID/Touch ID or Android StrongBox equivalent) plus device-bound credential. Single-factor login shall be blocked.
* AC-HC-001.4: All message metadata (sender DID, recipient DID(s), timestamp, channel, delivery status) shall be logged in an immutable audit trail exportable as evidence in OCR investigations.
* AC-HC-001.5: Message retention shall be configurable per channel: minimum 6 years for HIPAA compliance, extendable to 10 years at organization option. Retention operates on encrypted blob preservation; content remains encrypted throughout.
* AC-HC-001.6: Remote wipe capability shall allow admins to revoke device access and encrypted key material for departed or compromised employees without affecting the audit trail or integrity proofs.

### REQ-HC-002: Role-Based Clinical Communication

**User Story:** As a clinician, I want to send urgent messages to the on-call cardiologist — not to a specific named person I may not know — so that the right care team member receives the message regardless of schedule changes.

**Acceptance Criteria:**

* AC-HC-002.1: Messages shall be addressable to clinical roles (e.g., "on-call cardiologist," "charge nurse — Ward 4") that resolve to the currently scheduled individual. Role definitions are managed in the admin console.
* AC-HC-002.2: Escalation workflows shall support alert-until-read for urgent clinical messages: if unread after a configurable timeout (e.g., 5 minutes), the message escalates to the backup recipient defined for that role.
* AC-HC-002.3: Organization admins shall be able to define communication groups by department, care team, specialty, or custom grouping. Group membership shall be manageable via the admin console and SCIM provisioning.
* AC-HC-002.4: On-call schedule integration shall allow role-to-person mapping to update automatically based on scheduling system data (via API or manual admin update).

### REQ-HC-003: EHR Integration (Phase 2+)

**User Story:** As a clinician, I want to link a clinical message to a patient context in the EHR, so that care team communications are associated with the relevant patient record without creating a separate documentation burden.

**Acceptance Criteria:**

* AC-HC-003.1: ECHO Comply shall provide a REST API for integration with major EHR systems: Epic, Cerner/Oracle Health, and Athenahealth. Integration scope includes patient context linking and notification routing.
* AC-HC-003.2: Clinical messages linked to patient context shall generate a reference ID (not ePHI) that can be attached to the patient's EHR record. Message content is never transmitted to the EHR — only the reference ID and integrity proof.
* AC-HC-003.3: EHR integration shall be an optional add-on feature. ECHO Comply core features function without EHR connectivity.

## Feature Behavior and Rules

### Content Never Leaves Secure Infrastructure

ECHO Comply's architecture is content-blind. Even if an adversary compromises the relay servers, they receive only encrypted blobs — no ePHI, no identifiable content. This is how ECHO Comply achieves HIPAA's "encryption" requirement and simultaneously demonstrates zero-breach risk: there is no ePHI on the servers to breach.

### Breach Notification Obligations

HIPAA's 24-hour breach reporting requirement (2026 updates) applies to the discovery of a breach of unsecured ePHI. Because ECHO Comply's architecture stores only encrypted blobs (secured ePHI), a server-level compromise does not trigger HIPAA breach notification — the data is encrypted and unreadable. This is a direct selling point for healthcare customers: "a breach of our servers is not a HIPAA breach."

### Clinical UX Design Principle

Clinical staff will not adopt a messaging tool that feels like a compliance tool. ECHO Comply for healthcare must feel indistinguishable from a fast, modern consumer messaging app — with the compliance features operating silently in the background. The green checkmark indicating compliance verification should be subtle and non-disruptive. Speed and reliability are more important to clinical staff than feature richness.

### ECHO Comply — Local Government (FOIA)

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

ECHO Comply — Local Government provides FOIA-compliant secure messaging for state and local government organizations. Government agencies that allow staff to use personal phones or consumer apps (WhatsApp, SMS) for official business face serious legal liability: they cannot produce records in response to public records requests, and individual officials may face personal liability for destruction of public records. ECHO Comply gives government agencies automatic, tamper-evident preservation of official communications with the ability to produce independently verifiable records for any FOIA or public records request.

## Terminology

* **FOIA**: Freedom of Information Act — federal and state laws requiring government agencies to disclose records upon request, subject to specific exemptions. All 50 US states have analogous public records laws with varying requirements.
* **Public Record**: Any official communication made in the course of conducting government business, subject to preservation and disclosure requirements.
* **Exempt Communication**: Communications that qualify for FOIA exemption — attorney-client privileged communications, personnel matters, ongoing law enforcement investigations, and others defined by applicable state law.
* **Records Retention Schedule**: A government-defined policy specifying how long different categories of records must be preserved (e.g., permanent for council proceedings, 3–7 years for inter-agency).
* **Inter-Agency Communication**: Official communications between different government entities (e.g., city to county, agency to agency).

## Requirements

### REQ-GOV-001: Public Records Preservation

**User Story:** As a city clerk responsible for FOIA compliance, I want all official communications automatically preserved with tamper-evident proofs, so that I can respond to public records requests without depending on individual employees to retain their own messages.

**Acceptance Criteria:**

* AC-GOV-001.1: All messages sent within official government channels shall be automatically preserved with tamper-evident integrity proofs (Merkle root anchoring to Constellation metagraph) from the moment of transmission.
* AC-GOV-001.2: Retention periods shall be configurable per channel type: permanent for council proceedings and official policy communications, 3–7 years for inter-agency communications (configurable to match state law), and custom per organization.
* AC-GOV-001.3: FOIA/public records export shall produce a structured package including: message metadata (sender, recipients, timestamp), integrity proofs (Merkle path and snapshot reference), and — for public channels — encrypted content blobs accessible to authorized parties. Private channels (personnel, attorney-client) shall require additional legal authorization before inclusion.
* AC-GOV-001.4: The export package integrity proof shall demonstrate that the export is complete and unmodified — satisfying requirements for records completeness in FOIA responses.
* AC-GOV-001.5: Any third party receiving an exported package (court, journalist, citizen) shall be able to independently verify the integrity proofs against the public metagraph without requiring ECHO Comply access.

### REQ-GOV-002: Inter-Agency Secure Communication

**User Story:** As an emergency management coordinator, I want to securely communicate with officials from neighboring jurisdictions during an incident, so that sensitive operational communications are protected while remaining legally preserved.

**Acceptance Criteria:**

* AC-GOV-002.1: Government agencies shall be able to establish verified cross-organization channels with other agencies that have confirmed government DIDs. Cross-agency channel access requires mutual admin approval.
* AC-GOV-002.2: Emergency communication channels shall support alert-until-read with configurable escalation — critical incident notifications escalate automatically if unread after a defined timeout.
* AC-GOV-002.3: Official-to-constituent communications (e.g., city council member responding to a constituent) shall be distinguishable in the audit trail as "official record" vs. internal communications, enabling appropriate FOIA treatment per communication type.
* AC-GOV-002.4: Cross-agency communications shall generate integrity proofs that are independently verifiable by both participating agencies — neither agency depends on the other's cooperation to verify the record.

### REQ-GOV-003: Transparency and Accountability

**User Story:** As a government transparency officer, I want to generate compliance reports showing that official communications are being preserved correctly, so that I can demonstrate good-faith compliance with public records law.

**Acceptance Criteria:**

* AC-GOV-003.1: The admin console shall generate transparency reports showing: message volume by channel and time period, retention policy compliance status, channels with active legal holds, and export history — without exposing content.
* AC-GOV-003.2: The integrity proof for any official message shall be independently verifiable by any third party with the proof and access to the public metagraph — without requiring ECHO Comply assistance.
* AC-GOV-003.3: Officials who send messages on personal devices using the ECHO Comply app shall have those messages automatically preserved under the organization's retention policy — eliminating the "personal device" gap in public records compliance.

## Feature Behavior and Rules

### Automatic Preservation Eliminates Human Dependency

The most significant failure mode in government records compliance is reliance on individual employees to retain their own communications. ECHO Comply eliminates this dependency — preservation is automatic and tamper-evident from the moment of transmission, regardless of whether the sender subsequently deletes the message from their device. The encrypted blob and integrity proof are preserved on the server and anchored on-chain.

### Private Channels and Exemptions

Not all government communications are public records. Attorney-client communications, personnel matters, and ongoing law enforcement investigations typically qualify for FOIA exemptions. ECHO Comply supports "private channels" with restricted export access — these channels still preserve integrity proofs but require additional legal authorization before content (if any) can be produced in a FOIA response. The admin console clearly marks channels as "public record" or "restricted" during channel creation.

### ECHO Comply — Law Firms (Chain-of-Custody)

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

ECHO Comply — Law Firms provides secure messaging with chain-of-custody, litigation hold, ethical wall enforcement, and matter-based organization for law firms and legal departments. Law firms face a growing requirement to demonstrate that digital communications have not been tampered with when presented as evidence. No mainstream messaging platform provides cryptographic chain-of-custody natively. ECHO Comply's blockchain-anchored integrity proofs give legal teams the ability to present communications as evidence with the same rigor as traditional documentary evidence.

## Terminology

* **Chain-of-Custody**: The chronological documentation of the handling of evidence showing who obtained it, where it was stored, and who had access — establishing that evidence has not been tampered with.
* **Litigation Hold**: A directive to preserve all potentially relevant information (including communications) once litigation is reasonably anticipated, suspending normal retention schedules.
* **Ethical Wall (also "Chinese Wall")**: An information barrier preventing communication between attorneys working on conflicting matters — required by professional responsibility rules when a lawyer joins a firm with an existing conflict.
* **Matter**: A specific legal case or client engagement, used as the organizational unit for grouping related communications, documents, and activities.
* **Digital Evidence Export**: A structured package of communications with integrity proofs formatted for legal proceedings, showing authenticity and chain-of-custody.

## Requirements

### REQ-LAW-001: Chain-of-Custody for Digital Evidence

**User Story:** As a litigator, I want to produce client and team communications as evidence with cryptographic proof of authenticity, so that opposing counsel cannot challenge the integrity of the communications I present.

**Acceptance Criteria:**

* AC-LAW-001.1: Every message shall produce a Digital Evidence-compatible integrity proof containing: sender DID, timestamp (block time), message hash, Merkle path, and metagraph snapshot reference. This proof demonstrates that the message existed at the stated time and has not been modified.
* AC-LAW-001.2: Chain-of-custody metadata shall be exportable in standard forensic formats compatible with court e-filing requirements and eDiscovery platforms (Relativity, Everlaw, Logikcull).
* AC-LAW-001.3: The integrity proof shall be independently verifiable against the public Constellation metagraph by any third party — opposing counsel, a judge, or a forensic expert — without requiring ECHO Comply's cooperation.
* AC-LAW-001.4: Digital Evidence export packages shall include a cover sheet with: the scope of the export, the requesting administrator's DID, timestamp of export, and an integrity proof that the export package itself is complete.

### REQ-LAW-002: Litigation Hold

**User Story:** As a litigation support manager, I want to place a litigation hold on all communications related to a matter, so that I satisfy my preservation obligations the moment litigation becomes reasonably anticipated — even if users delete messages from their devices.

**Acceptance Criteria:**

* AC-LAW-002.1: Admins shall be able to place a litigation hold on specific channels, users, date ranges, or matter tags — suspending all retention schedule expiration for the held communications.
* AC-LAW-002.2: Litigation hold shall be enforceable even if the user's device is wiped, lost, or the user is terminated — the encrypted blob and integrity proof are preserved server-side and on-chain independently of device state.
* AC-LAW-002.3: Litigation hold activation and release events shall be logged in the audit trail with the admin's DID and timestamp, creating an immutable record of when the hold was placed.
* AC-LAW-002.4: Users subject to a litigation hold shall not receive notification of the hold — consistent with legal practice around preservation obligations.
* AC-LAW-002.5: The admin console shall display all active holds with their scope, duration, estimated data volume, and associated matter.

### REQ-LAW-003: Ethical Wall Enforcement

**User Story:** As a conflicts attorney, I want to configure information barriers between specified attorney groups, so that I comply with professional responsibility rules when we have a conflict situation.

**Acceptance Criteria:**

* AC-LAW-003.1: Admin console shall support ethical wall configuration — defining which user groups are prohibited from communicating with each other (e.g., attorneys on opposing sides of a matter, lateral hire with client conflict).
* AC-LAW-003.2: Ethical wall violations shall be blocked at the relay layer before delivery — the message never reaches the recipient. The sender receives an error message; the block is logged in the compliance audit trail.
* AC-LAW-003.3: Ethical wall configurations shall be enforceable retroactively — when a conflict arises, the wall activates immediately for all future communications; historical communications remain accessible only to authorized parties.
* AC-LAW-003.4: Ethical wall events (attempted violations, wall activation, wall removal) shall be logged immutably in the audit trail.

### REQ-LAW-004: Matter-Based Organization

**User Story:** As a partner managing multiple active matters, I want to organize team communications by matter, so that I can apply matter-specific retention and access controls and produce a clean eDiscovery package when a matter closes.

**Acceptance Criteria:**

* AC-LAW-004.1: Conversations shall be organizable by matter/case with matter-specific retention policies, access controls, and export capabilities. A single message may be associated with at most one matter.
* AC-LAW-004.2: Matters shall support: opening date, closing date, matter type (litigation, transactional, regulatory), responsible attorney, and team member access list.
* AC-LAW-004.3: When a matter closes, all associated communications shall be archived with full integrity proofs. The archive shall be exportable as a single Digital Evidence package for the client file.
* AC-LAW-004.4: Closed matter archives shall remain accessible for the firm's configured retention period (minimum 7 years from matter close). Access to closed matter archives requires admin authorization.

## Feature Behavior and Rules

### Evidence Admissibility Strategy

The key selling point for law firms is that ECHO Comply communications can be presented as evidence with stronger authenticity guarantees than screenshots, server logs, or platform-provided exports — all of which depend on trusting the platform's internal systems. ECHO Comply's Merkle proofs are anchored to a public blockchain verifiable by anyone, independent of ECHO Comply's continued existence or cooperation. This is a category-defining differentiator in legal technology.

### Ethical Wall vs. Admin Access

When an ethical wall is in place, not even organization admins should be able to view communications across the wall. The admin console shall display that an ethical wall exists but shall not permit admin-level bypass of the wall for content access. Admins can view the metadata (that messages were blocked) but not content that flows across conflicting matters.

## ECHO Protocol Foundation and Corporate Structure

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

ECHO operates as a two-entity structure designed to separate protocol stewardship from commercial operations. The ECHO Protocol Foundation (a Wyoming Decentralized Unincorporated Nonprofit Association — DUNA) stewards the open-source protocol, holds the metagraph, and issues the token at Phase 3+. ECHO Comply LLC (a Wyoming LLC) operates the commercial enterprise product, employs the team, signs customer contracts, and holds compliance certifications. Revenue flows from the LLC to the Foundation via a services agreement, funding protocol development without requiring token sales or venture capital.

This structure is deliberate: enterprise customers need to sign contracts with a commercial entity; the protocol needs to be governed by a nonprofit structure that cannot be captured by commercial interests. Both requirements are met simultaneously.

## Terminology

* **DUNA**: Decentralized Unincorporated Nonprofit Association — a Wyoming legal entity type (2024 statute) providing legal standing for decentralized organizations while preserving nonprofit, community-governance characteristics. A DUNA can hold assets, enter contracts, and sue/be sued.
* **ECHO Protocol Foundation**: The Wyoming DUNA that stewards the ECHO Protocol open-source code, metagraph, and (at Phase 3+) token issuance and governance.
* **ECHO Comply LLC**: The Wyoming LLC that sells ECHO Comply to enterprises, employs the team, holds SOC 2 and compliance certifications, and signs BAAs and enterprise contracts.
* **Services Agreement**: The contract between ECHO Comply LLC and the Foundation under which the LLC pays for protocol development and the Foundation grants development resources.
* **Wyoming DAPT**: Domestic Asset Protection Trust — an optional legal structure that can serve as managing member of the LLC to provide additional privacy protection for founders.
* **FinCEN / CTA**: Financial Crimes Enforcement Network / Corporate Transparency Act — federal requirements for beneficial ownership disclosure (private filing, not public disclosure).

## Requirements

### REQ-CORP-001: Foundation Establishment

**User Story:** As a founder, I want the ECHO Protocol stewarded by a nonprofit foundation, so that the protocol cannot be captured by any single commercial interest, and so that the community can trust the infrastructure as independent of corporate control.

**Acceptance Criteria:**

* AC-CORP-001.1: The ECHO Protocol Foundation shall be incorporated as a Wyoming DUNA. Incorporation shall be completed within the first 2 months of Phase 1.
* AC-CORP-001.2: The Foundation shall hold: the ECHO Protocol open-source codebase (Apache 2.0 upon Phase 2 open-source release), the metagraph node infrastructure, and (at Phase 3+) the token issuance and treasury.
* AC-CORP-001.3: The Foundation board shall consist of 3–5 members in Phase 1 (founders and independent advisors), expanding to 10 members at Phase 3+ (5 founders + 5 community-elected members).
* AC-CORP-001.4: The Foundation's DUNA structure shall support transition to on-chain token-holder governance at Phase 3+ without requiring entity restructuring. Decision-making can migrate from board-only to board + token-holder voting as the token ecosystem matures.

### REQ-CORP-002: Commercial LLC Operations

**User Story:** As an enterprise customer, I want to sign contracts with a commercial entity that holds compliance certifications and can be held legally accountable, so that I have appropriate legal recourse and vendor assurance.

**Acceptance Criteria:**

* AC-CORP-002.1: ECHO Comply LLC shall be incorporated as a Wyoming LLC. Incorporation shall be completed within the first 2 months of Phase 1.
* AC-CORP-002.2: ECHO Comply LLC shall employ the development team and sign all enterprise customer contracts (BAAs, SLAs, NDAs, data processing agreements).
* AC-CORP-002.3: ECHO Comply LLC shall hold all compliance certifications: SOC 2 Type I (Phase 1), SOC 2 Type II (Phase 2), HIPAA compliance documentation, and relevant state compliance certifications.
* AC-CORP-002.4: ECHO Comply LLC shall pay a Foundation services fee (percentage of net revenue) to fund protocol development. This creates sustainable protocol funding from commercial revenue without requiring token sales.

### REQ-CORP-003: Revenue Flow and Financial Structure

**User Story:** As a founder and Foundation board member, I want a clear financial structure where commercial revenue funds protocol development, so that the Foundation can operate sustainably without depending on token sales or donations.

**Acceptance Criteria:**

* AC-CORP-003.1: Revenue allocation shall be: 60% operations (LLC salaries, infrastructure, compliance certifications), 25% Foundation development grant, 15% Foundation reserve.
* AC-CORP-003.2: The Foundation reserve shall be used to: fund protocol security audits, pay metagraph node infrastructure costs (750K DAG staking + server costs), and maintain 12+ months of operating runway.
* AC-CORP-003.3: At Phase 3+ with token launch, the Foundation also receives protocol-level fees (metagraph transaction fees, premium feature ECHO burns) independent of LLC commercial revenue, creating dual revenue streams for the Foundation.

### REQ-CORP-004: Founder Privacy Structure

**User Story:** As a founder who requires identity protection for operational security or personal safety reasons, I want the corporate structure to support appropriate privacy layers, so that I can operate the company without unnecessary public exposure.

**Acceptance Criteria:**

* AC-CORP-004.1: Wyoming LLC privacy: state filings shall not require public disclosure of LLC member names. A professional registered agent shall shield the founder's address.
* AC-CORP-004.2: DUNA privacy: the DUNA statute does not require individual member identification in state filings. Governance participation can be pseudonymous via on-chain DID identity.
* AC-CORP-004.3: CTA/FinCEN compliance: beneficial ownership disclosure shall be filed with FinCEN as required (private federal filing, not public disclosure). Legal counsel shall ensure compliance before incorporation.
* AC-CORP-004.4: Optional additional layer: a Wyoming Domestic Asset Protection Trust (DAPT) may serve as managing member of the LLC, providing additional legal separation between founder identity and the commercial entity — a legitimate asset protection structure.
* AC-CORP-004.5: Healthcare and government customer due diligence: at least one named individual (co-founder or hired executive) shall be designated as the customer-facing representative for vendor due diligence purposes, satisfying enterprise procurement requirements without requiring all founders to be publicly disclosed.

### REQ-CORP-005: Phase 3+ Governance Transition

**User Story:** As a Foundation board member, I want governance to transition to token-holder participation when the token launches, so that the community-ownership thesis is fulfilled as the user base grows.

**Acceptance Criteria:**

* AC-CORP-005.1: At Phase 3+ token genesis, the Foundation board shall expand from Phase 1 composition (founders + advisors) to 10 members: 5 founders + 5 community-elected members.
* AC-CORP-005.2: Community board member elections shall occur annually, with Trust Tier 3+ token holders eligible to stand as candidates. Voting weight = staked ECHO × trust tier multiplier.
* AC-CORP-005.3: Protocol parameter changes (metagraph schema, encryption standards, fee schedules) shall require token-holder governance vote (67% supermajority for security-critical parameters, simple majority for others).
* AC-CORP-005.4: Founder board members retain veto rights on existential protocol decisions (e.g., abandoning E2E encryption, changing the open-source license) for years 1–5. After year 5, founders transition to advisory roles with no unilateral veto.
* AC-CORP-005.5: Emergency protocol changes (security patches) shall require 3-of-5 founder approval, enabling rapid response to critical vulnerabilities without the delay of a full governance vote.

## Feature Behavior and Rules

### Why Two Entities

The two-entity structure solves a fundamental tension: enterprise customers need commercial accountability, but the protocol must be community-owned. A single company that owns the protocol is always susceptible to acquisition, pivot, or value extraction. A foundation that holds the protocol cannot be acquired — it can only be dissolved by the community. The LLC provides commercial accountability; the Foundation provides permanent, independent protocol stewardship.

### Open Source Timeline

The protocol is closed during Phase 1 to protect competitive advantage during initial enterprise deployments. Healthcare and government enterprise sales are easier with a proprietary product — customers are not concerned about open-source forks. At Phase 2, the protocol is released under Apache 2.0. This timing is deliberate: open-source at launch to build community trust and developer ecosystem, after the enterprise foundation is established.

## Portable Social Graph and Protocol Layer

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

The Portable Social Graph is the most structurally differentiated capability ECHO can build. Every major messaging platform today — WhatsApp, Telegram, Discord, Signal — traps your social graph inside a company's private database. When you leave, you lose everything: your contacts, your conversation history, your group memberships, your trust reputation. This is the "platform lock-in" that keeps hostile platforms dominant despite abusing their users. The platform owns the network effect, not the users.

ECHO breaks this model by anchoring social relationships, trust reputation, and verified credentials on the Cardano blockchain and Constellation metagraph — owned by users, portable across any application, and verifiable by any third party. Your social graph on ECHO is not a record in ECHO's database. It is a set of cryptographic facts about your identity and relationships that you carry with you forever.

This creates two compounding advantages. First, users who accumulate Trust Tier reputation, verified credentials, and governance tokens on ECHO face enormous switching costs — not because ECHO locks them in, but because their verified identity and social capital are genuinely valuable and portable. Second, ECHO becomes a protocol layer that third-party developers can build on — a DAO chat client, a healthcare team coordinator, a secure journalist drop — all accessing the same identity layer and trust framework. The more applications that build on the ECHO Protocol, the more valuable every user's ECHO identity becomes.

This is the DeSoc (Decentralized Society) thesis applied to verifiable communication: a base-layer protocol that creates owned network effects where the value belongs to users, not the platform.

## Terminology

* **Social Graph**: The map of a user's relationships — contacts, followers, group memberships, trust connections. Currently owned by platforms; ECHO makes it user-owned and on-chain.
* **Portable Identity**: A user's DID, verifiable credentials, and trust tier are anchored on Cardano. Any application that supports Cardano DID standards can read and verify them.
* **Protocol Layer**: ECHO Protocol is the shared infrastructure beneath ECHO Comply and ECHO Message. Third-party developers can build new front-end applications on the protocol while accessing the same identity and trust layer.
* **Cross-App Composability**: A user's followers, trusted contacts, and governance tokens function across any ECHO Protocol-compatible application — not just the official ECHO apps.
* **Owned Network Effect**: Network effects that belong to users, not to a platform. When users own their social graph, the switching cost is in their favor — they carry their network anywhere, but their network history is richest where they've spent the most time.
* **Social Graph Edge**: A cryptographic representation of a relationship — "User A trusts User B at Tier 3" — stored on-chain as a verifiable credential or metagraph commitment.

## Requirements

### REQ-GRAPH-001: On-Chain Social Graph Anchoring

**User Story:** As an ECHO user, I want my contact relationships and trust connections anchored on-chain, so that my social network belongs to me and is portable to any application, not locked inside ECHO's database.

**Acceptance Criteria:**

* AC-GRAPH-001.1: When a user establishes a trusted contact relationship (Inner Circle, Trusted, Acquaintance), the relationship shall be recorded as a signed verifiable credential on Cardano, readable by any DID-compatible application.
* AC-GRAPH-001.2: When a user's trust tier changes, the updated tier commitment shall be recorded on the Constellation metagraph Data L1. Any third party with the user's DID can query their current trust tier from the public metagraph.
* AC-GRAPH-001.3: A user's complete social graph (contacts, trust tiers, group memberships) shall be exportable as a JSON-LD document signed by the user's DID — a machine-readable, cryptographically verifiable snapshot of their social connections.
* AC-GRAPH-001.4: Social graph data shall be structured using W3C DID and Verifiable Credential standards, enabling interoperability with any DID-compatible platform — not just ECHO.
* AC-GRAPH-001.5: The social graph export shall include: contact DIDs, relationship tier, relationship established date, and the user's own trust tier credentials. Message content shall never be included in the social graph export.

### REQ-GRAPH-002: Cross-Application Identity Portability

**User Story:** As a user who builds trust reputation on ECHO, I want my verified identity and trust tier to be recognized by other applications that check Cardano DIDs, so that my reputation is an asset I own, not a number locked inside ECHO.

**Acceptance Criteria:**

* AC-GRAPH-002.1: A user's ECHO Trust Tier credential shall be a standard W3C Verifiable Credential issued by the ECHO Protocol Foundation, verifiable by any relying party that can check Cardano DID status.
* AC-GRAPH-002.2: Third-party applications that integrate the ECHO Protocol DID resolver shall be able to verify a user's trust tier, credential status, and key public facts about their identity — without requiring ECHO company cooperation.
* AC-GRAPH-002.3: When a user switches to a third-party ECHO Protocol-compatible application, their trust tier, verified credentials, and contact list shall be importable via DID-based authentication — no re-verification required.
* AC-GRAPH-002.4: ECHO shall publish and maintain an open DID resolver SDK (TypeScript and Go) that any developer can use to verify ECHO identities.

### REQ-GRAPH-003: Third-Party Protocol Developer Access

**User Story:** As a developer building privacy-focused applications, I want to build on the ECHO Protocol's identity and trust infrastructure, so that I can launch applications with immediate access to verified users rather than bootstrapping from zero.

**Acceptance Criteria:**

* AC-GRAPH-003.1: Upon Phase 2 open-source release (Apache 2.0), ECHO Protocol shall publish a developer SDK with: DID resolver, trust tier verification, Merkle proof generation/verification, and relay client libraries.
* AC-GRAPH-003.2: Third-party applications that integrate the ECHO Protocol SDK shall be able to: authenticate users by their ECHO DID, read users' public trust tier, send and receive E2E encrypted messages via the ECHO relay network, and anchor integrity proofs to the Constellation metagraph.
* AC-GRAPH-003.3: Protocol usage by third-party applications shall contribute to the ECHO metagraph fee pool — a percentage of relay transactions from third-party apps flows to the Foundation treasury (subject to token launch, Phase 3+).
* AC-GRAPH-003.4: ECHO shall maintain a public developer registry of third-party applications built on the ECHO Protocol, giving users visibility into which apps can access their DID.

### REQ-GRAPH-004: Governance-Weighted Social Reputation

**User Story:** As a long-term ECHO user with a high trust tier and engaged governance history, I want my on-chain reputation to carry weight beyond ECHO — in other platforms, DAOs, and communities — so that my investment in building verified identity has compounding returns.

**Acceptance Criteria:**

* AC-GRAPH-004.1: A user's ECHO governance participation history (proposals voted on, proposals submitted, board elections) shall be anchored on Data L1 as an immutable record, readable by any third party.
* AC-GRAPH-004.2: The ECHO Protocol Foundation shall publish a "reputation standard" that defines how external applications can interpret ECHO trust tiers and governance history as signals of identity quality.
* AC-GRAPH-004.3: At Phase 3+, staked ECHO token positions shall be visible on-chain, enabling third-party applications (DeFi protocols, DAOs, compliance tools) to use staking history as a proof of long-term community commitment.

## Feature Behavior and Rules

### The Compounding Advantage

The portable social graph creates a compounding advantage for users who invest in ECHO early. Every verified credential earned, every trust tier advancement, every governance vote cast becomes part of an on-chain reputation that grows in value as more applications integrate the ECHO Protocol. Unlike centralized reputation systems (Twitter's follower count, LinkedIn's connections), ECHO reputation cannot be revoked by the platform, cannot be hidden by algorithm changes, and cannot be lost if the company shuts down.

### Switching Costs Favor Users, Not the Platform

When a centralized platform creates switching costs, it means users can't leave without losing their data — this is exploitation. When ECHO creates switching costs through the portable social graph, it means users have built genuine, verifiable social capital that is richest in the ECHO ecosystem. The user chooses to stay because their reputation is most valuable here — not because they're locked in.

### Developer Ecosystem as Network Effect Multiplier

Each third-party application built on the ECHO Protocol adds network effects to every ECHO user's identity. A healthcare app that uses ECHO DID verification makes a Tier 4 (government-ID verified) ECHO credential more valuable. A DAO governance tool that accepts ECHO trust tiers makes a Tier 5 credential more meaningful. The protocol layer is an open standard; the value of having ECHO credentials grows with the ecosystem that adopts it.

## Post-Quantum Cryptography Mode

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

Post-Quantum Cryptography (PQC) Mode is a forward-looking encryption upgrade that protects ECHO communications against a credible long-horizon threat: quantum computers capable of breaking current public-key cryptography. NIST finalized post-quantum algorithms in 2024 (CRYSTALS-Kyber for key exchange, CRYSTALS-Dilithium for digital signatures). Major governments — including CISA in the United States — are actively mandating quantum migration timelines for critical communications infrastructure.

The threat is not theoretical paranoia: communications intercepted today can be stored by adversaries and decrypted later when quantum computers mature ("harvest now, decrypt later"). Healthcare records retained for 7+ years, legal communications subject to long-term eDiscovery holds, and classified communications are all vulnerable to this attack model. Organizations that encrypt communications today with current algorithms are creating a future liability.

ECHO's dual-layer approach — running X25519 (current) and Kyber (post-quantum) key exchange in parallel — provides cryptographic agility. Users who need quantum-resistant guarantees can opt into PQC mode today. Users who don't need it face no performance overhead. This positions ECHO as the first messaging platform to offer production-grade PQC as a configurable feature, building a significant technical moat before competitors even begin evaluating the migration.

## Terminology

* **Post-Quantum Cryptography (PQC)**: Cryptographic algorithms designed to be secure against both classical and quantum computers. NIST standardized four PQC algorithms in 2024.
* **CRYSTALS-Kyber (now ML-KEM)**: The NIST-standardized post-quantum key encapsulation mechanism. Used for establishing shared secrets between parties, replacing X25519. Significantly larger key sizes than X25519 but computationally feasible on mobile devices.
* **CRYSTALS-Dilithium (now ML-DSA)**: The NIST-standardized post-quantum digital signature algorithm. Used for signing messages and DIDs, replacing ECDSA-P256.
* **Hybrid Mode**: Running X25519 + Kyber simultaneously for key exchange. Protects against both classical attacks (if Kyber has unknown weaknesses) and quantum attacks (if X25519 is broken by quantum computers). NIST recommends hybrid mode during the transition period.
* **"Harvest Now, Decrypt Later"**: An attack where adversaries collect encrypted communications today, intending to decrypt them when quantum computing matures. Communications with long retention requirements (healthcare, legal, government) are most at risk.
* **Cryptographic Agility**: An architecture that supports swapping or layering cryptographic algorithms without requiring a complete system rebuild. ECHO's encryption layer must be crypto-agile to support PQC migration.
* **CISA PQC Mandate**: The US Cybersecurity and Infrastructure Security Agency's mandate that federal agencies and critical infrastructure operators migrate to NIST PQC standards by 2030. State-level cascade mandates are following.

## Requirements

### REQ-PQC-001: Hybrid Key Exchange (X25519 + Kyber)

**User Story:** As a user in a high-risk environment or an enterprise customer with long retention requirements, I want my messages protected against future quantum attacks, so that communications I send today remain private even if quantum computers become capable of breaking current encryption within the retention period.

**Acceptance Criteria:**

* AC-PQC-001.1: When PQC mode is enabled, the system shall perform hybrid key exchange combining X25519 (current) and ML-KEM-768 (CRYSTALS-Kyber, NIST Level 3) simultaneously. The shared secret is derived from both exchanges, providing security if either algorithm is broken.
* AC-PQC-001.2: PQC mode shall be configurable per conversation, not globally forced. Users and enterprise administrators shall be able to require PQC for specific channels or conversations without applying it to all communications.
* AC-PQC-001.3: In ECHO Comply, enterprise administrators shall be able to enforce PQC mode as mandatory for specific channels (e.g., "all clinical communications require PQC") via the admin console.
* AC-PQC-001.4: Both parties in a conversation must have PQC mode enabled and support the hybrid key exchange for PQC mode to activate. The system shall gracefully fall back to standard X25519 if the other party's client does not support PQC, with a clear indicator showing whether PQC protection is active.
* AC-PQC-001.5: PQC mode activation shall be displayed as a visible indicator in the conversation UI — a shield icon or similar — so users know whether their current conversation is quantum-resistant.

### REQ-PQC-002: Post-Quantum DID Signatures

**User Story:** As a user, I want my Cardano DID operations (identity registration, credential issuance, trust tier updates) to be signed with post-quantum algorithms, so that my on-chain identity remains secure against quantum attacks on the Cardano network.

**Acceptance Criteria:**

* AC-PQC-002.1: At Phase 3+ PQC integration, new DID registrations shall use ML-DSA-65 (CRYSTALS-Dilithium) for signing DID documents in addition to (not replacing) the current ECDSA-P256 signatures.
* AC-PQC-002.2: Hybrid signatures (ECDSA-P256 + ML-DSA-65) shall be included in all DID operations to maintain backward compatibility with current Cardano verification infrastructure while adding quantum resistance.
* AC-PQC-002.3: Existing users' DIDs shall not be automatically migrated — PQC-enhanced DID signatures shall be an opt-in upgrade requiring the user's biometric confirmation in the Secure Enclave.

### REQ-PQC-003: Enterprise PQC Compliance Documentation

**User Story:** As a healthcare or government enterprise administrator, I want ECHO to produce documented evidence of our PQC encryption status, so that I can demonstrate compliance with CISA post-quantum migration mandates and healthcare data security regulations.

**Acceptance Criteria:**

* AC-PQC-003.1: The ECHO Comply admin console shall display: PQC mode status per channel, percentage of messages delivered with quantum-resistant encryption, the specific NIST-standardized algorithms in use, and the NIST security level of the current configuration.
* AC-PQC-003.2: ECHO Comply shall produce a PQC compliance report exportable as PDF and JSON, suitable for inclusion in HIPAA security documentation, FISMA compliance packages, and enterprise security audits.
* AC-PQC-003.3: ECHO shall maintain public documentation mapping its PQC implementation to CISA's Post-Quantum Cryptography Migration roadmap, enabling customers to reference ECHO in their own compliance filings.

### REQ-PQC-004: Consumer PQC Premium Feature

**User Story:** As a privacy-conscious consumer user who understands the long-horizon quantum threat, I want PQC mode available as a premium feature, so that I can protect my most sensitive communications against future quantum attacks.

**Acceptance Criteria:**

* AC-PQC-004.1: PQC mode shall be available to ECHO Message Premium subscribers ($9.99/month) as a configurable per-conversation feature.
* AC-PQC-004.2: The ECHO Message onboarding flow shall include a plain-language explanation of what PQC mode does and when it matters — targeted at privacy-conscious users who are not cryptography experts.
* AC-PQC-004.3: PQC mode shall have no visible latency impact on the messaging experience for messages under 10MB on devices released within the past 4 years. The additional key exchange overhead shall complete within 200ms.

## Feature Behavior and Rules

### Why Build This Now

The competitive window for PQC is narrow. Once CISA mandates cascade to enterprise customers, every competitor will scramble to implement PQC. Organizations that built PQC expertise early will complete the migration cleanly; late movers will face rushed, error-prone implementations. ECHO builds the PQC layer now — while the protocol is being architected — rather than retrofitting it later. This is the same reason X25519 was chosen over RSA: cryptographic agility is an architectural decision, not a feature addition.

### Phase Approach

**Phase 1 (ECHO Comply):** PQC mode available as a beta enterprise feature for healthcare and government customers who face CISA compliance timelines. Documentation focus — help customers include PQC status in their compliance filings.

**Phase 2 (ECHO Message):** PQC mode available to Premium subscribers. Plain-language explanation in onboarding.

**Phase 3+:** PQC-enhanced DID signatures for new registrations. PQC as default-on for high-trust-tier users who have opted into maximum security posture.

## Privacy Commons Treasury

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

The Privacy Commons Treasury transforms ECHO from a product users consume into an organization users govern for a shared purpose. Where v2.5.1 proposed a general-purpose community treasury focused on ECHO token burns and Bitcoin reserves, the Privacy Commons model focuses the treasury on an immediately compelling mission: funding the real-world fight for privacy rights.

The thesis is simple: people who care deeply about private communication don't just want an app that protects their messages — they want to be part of a movement that protects everyone's right to send them. The Privacy Commons Treasury gives users direct financial and governance control over initiatives that make privacy real for the most vulnerable people: journalists in restrictive regimes, activists facing government surveillance, defendants whose communications are used against them.

This is not charity. It's a mission that creates the most passionate user base imaginable. Signal's users are evangelical about it because they believe in what Signal stands for. ECHO's users will be evangelical because they literally fund what ECHO fights for. Every token holder who votes to fund a journalist's legal defense is more invested in ECHO's success than any user who merely earns staking yields.

The Privacy Commons Treasury is also the sustainable path to the Network State vision — not by acquiring land immediately, but by building the community infrastructure, legal precedent, and collective economic power that makes expanded ambitions achievable over time.

## Terminology

* **Privacy Commons**: The shared resources managed by the ECHO Protocol Foundation for the benefit of the user community and the broader cause of private communication.
* **Privacy Legal Defense Fund**: A treasury-funded initiative that provides legal representation for ECHO users who face government demands to reveal private communications, challenges to encryption law, or retaliation for using privacy tools.
* **Journalist and Activist Program**: A subsidized or fully funded access program for verified journalists, human rights workers, and activists operating in high-risk environments — funded by the community treasury.
* **Privacy Research Fund**: Treasury grants for independent security audits, academic privacy research, and development of open-source privacy infrastructure.
* **Governance Vote**: An on-chain proposal and vote (Token holders at Phase 3+, Foundation board before token launch) that determines how treasury funds are allocated across Privacy Commons initiatives.
* **Mission Moat**: The competitive advantage that comes from having a user community that is passionately invested in the organization's mission, not merely its product features.

## Requirements

### REQ-COMMONS-001: Privacy Legal Defense Fund

**User Story:** As an ECHO user facing a government demand to reveal my private communications, I want the community treasury to fund my legal defense, so that my choice to use private communications is backed by the collective resources of the ECHO community.

**Acceptance Criteria:**

* AC-COMMONS-001.1: The Privacy Legal Defense Fund shall be a designated allocation within the Foundation treasury, funded by a governance-approved percentage of annual treasury surplus (example starting ratio: 10–15%).
* AC-COMMONS-001.2: Users who face legal demands related to their ECHO communications (subpoenas, government warrants, compelled decryption orders) may apply for legal defense funding. Applications shall be reviewed by the Foundation's legal committee.
* AC-COMMONS-001.3: The Foundation shall maintain a roster of privacy-specialized law firms across key jurisdictions (US, EU, UK, Brazil, India) who are pre-vetted and available for rapid deployment when users face time-sensitive legal threats.
* AC-COMMONS-001.4: At Phase 3+ (post-token), the community may vote to expand the legal defense fund budget or to fund strategic impact litigation (cases that, if won, create precedent benefiting all privacy-focused platforms and users).
* AC-COMMONS-001.5: All legal defense grants shall be documented in the public treasury dashboard (preserving attorney-client privilege where required) — showing the number of cases supported, jurisdictions, and outcomes without revealing user identities.

### REQ-COMMONS-002: Journalist and Activist Program

**User Story:** As a journalist working in a country with communications surveillance, or as a human rights activist whose communications could put me at risk, I want free or subsidized access to ECHO's most secure features, so that I can communicate safely regardless of my ability to pay.

**Acceptance Criteria:**

* AC-COMMONS-002.1: The Journalist and Activist Program shall provide verified journalists and human rights workers with free access to ECHO Message Premium features, including Hidden Folders with Duress PIN, PQC mode (Phase 2+), and maximum privacy settings.
* AC-COMMONS-002.2: Verification for program eligibility shall be conducted by recognized press freedom organizations (Reporters Without Borders, CPJ, Freedom of the Press Foundation) whose confirmation is accepted in place of ECHO's standard identity verification — allowing program members to maintain lower on-chain identity disclosure.
* AC-COMMONS-002.3: Program participants shall receive enhanced onboarding that covers: operational security best practices, Duress PIN configuration, metadata minimization, and regional legal context for their jurisdiction.
* AC-COMMONS-002.4: The program shall have a governance-set annual budget (example: $200,000–$500,000 from treasury surplus) covering account subsidies, dedicated support, and onboarding resources.
* AC-COMMONS-002.5: At Phase 3+, the community may vote to expand program eligibility, increase the budget, or add new categories of at-risk users (domestic violence survivors, political dissidents, whistleblowers).

### REQ-COMMONS-003: Privacy Research and Infrastructure Grants

**User Story:** As a security researcher or open-source developer working on privacy tools, I want access to Foundation grants that fund important privacy infrastructure, so that the broader privacy ecosystem benefits from ECHO's success.

**Acceptance Criteria:**

* AC-COMMONS-003.1: The Privacy Research Fund shall allocate a governance-approved budget for: independent security audits of ECHO Protocol and third-party privacy tools, academic research grants for privacy-preserving cryptography and metadata protection, and open-source development grants for tools that integrate with or benefit the ECHO ecosystem.
* AC-COMMONS-003.2: Grant applications shall be submitted publicly and evaluated by a technical review committee (Foundation board during Phase 1–2; community-elected technical committee at Phase 3+).
* AC-COMMONS-003.3: All grant recipients shall publish their work as open-source (Apache 2.0 or equivalent) with a requirement to attribute the ECHO Protocol Foundation. Research papers funded by the Foundation shall be published in open-access venues.
* AC-COMMONS-003.4: The Foundation shall conduct an annual public security audit of the ECHO Protocol funded by the treasury, with the audit report published in full.

### REQ-COMMONS-004: Treasury Governance and Transparency

**User Story:** As a token holder or community member, I want full visibility into how Privacy Commons funds are allocated and spent, so that I can trust the treasury is being used for its stated mission.

**Acceptance Criteria:**

* AC-COMMONS-004.1: A public Privacy Commons dashboard shall display in real time: total treasury balance, allocation ratios across all programs, year-to-date spending per program, grant recipients (where public), and upcoming governance votes.
* AC-COMMONS-004.2: All treasury transactions shall be on-chain and independently verifiable. The dashboard shall link to on-chain transaction records for every disbursement.
* AC-COMMONS-004.3: Annual treasury reports shall be published summarizing: total funds received, total disbursed per program, impact metrics (cases supported, journalists protected, security vulnerabilities found), and the next year's proposed allocations.
* AC-COMMONS-004.4: Before the token launches (Phase 1–2), the Foundation board (5 unanimous votes required) shall approve all Privacy Commons disbursements above $10,000. After token launch (Phase 3+), all allocation ratios shall be governed by token holder votes.

## Feature Behavior and Rules

### Why This Creates the Strongest User Community

Research on community-owned platforms consistently shows that the most resilient communities are those built around a shared mission, not just shared economics. Audius users care about music creator rights. Hive users care about censorship resistance. ECHO users who fund the Privacy Legal Defense Fund and the Journalist Program care about something real and immediate — the right of human beings to communicate privately.

This is the "mission moat": the competitive advantage that comes from user passion. A user who votes to fund a journalist's legal defense in a restrictive country is not churning to a competitor with better sticker packs. They are an owner of an organization fighting for something they believe in.

### The Path to the Network State Vision

The Privacy Commons Treasury is the sustainable version of the Network State vision from PRD v2.5.1. Rather than projecting land acquisition and physical jurisdiction 10 years out, the Privacy Commons builds community infrastructure, legal precedent, and collective identity — the preconditions for anything more ambitious — starting from day one of the token launch.

The sequence: Privacy Commons (Year 3–5) → demonstrated community governance → treasury surplus → expanded missions (Year 5+). Each phase earns the next.

## Data Sovereignty Layer

## Overview

Provide a clear and concise summary of the feature, explaining what it does and the value it delivers to the user. Describe the core problem this feature solves and how it fits into the overall product.

## Terminology

* **Key Term 1**: Brief description that ensures shared understanding across the team.
* **Key Term 2**: Definition that clarifies any ambiguity in how this concept is used.

## Requirements

### REQ-XXX-001: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-001.1:** When the user [performs action], the system shall [respond with specific behavior].
* **AC-XXX-001.2:** When [condition exists], the system shall [handle appropriately].
* **AC-XXX-001.N:** [Continue for all acceptance criteria]

### REQ-XXX-002: Requirement Name

**User Story:** As a [role], I want to [perform action], so that I can [achieve outcome].

**Acceptance Criteria:**

* **AC-XXX-002.1:** When [condition], the system shall [behavior].
* **AC-XXX-002.2:** [Continue for all acceptance criteria]

## Feature Behavior & Rules

This section clarifies how the requirements behave in practice and how they interact. It explains cross-requirement interactions, defaults, constraints, and edge conditions without prescribing UI or user flows.

## Overview

The Data Sovereignty Layer is ECHO's most radical and forward-looking feature: an opt-in system that lets users monetize their own communication behavioral data — on their own terms, with zero content exposure — and receive direct payment in ECHO tokens.

Every major messaging platform already monetizes user behavior data. WhatsApp shares metadata with Meta's advertising machine. Google reads Gmail to target ads. The difference is that users receive nothing while platforms extract billions. The Data Sovereignty Layer inverts this model: users who choose to participate contribute anonymized behavioral patterns (not content — never content) to a community data pool, and receive a proportional share of the fees paid by researchers, enterprises, and public health organizations who query that pool.

This is not surveillance capitalism with extra steps. The architecture is fundamentally different:

1. **Zero content exposure** — message content never leaves the user's device. Only behavioral patterns (message frequency, network structure, topic metadata) are contributed, and only with explicit opt-in.
2. **ZK-proven anonymization** — contributions to the pool are proven anonymous via zero-knowledge proofs before any data leaves the device. Users can verify their contribution is unlinkable to their identity.
3. **User controlled** — users opt in per data category, can revoke consent at any time, and can see exactly what queries their data has satisfied.
4. **Direct payment** — 70% of all query fees flow directly to contributing users in proportion to the value of their data. No platform intermediary extracts the majority.

The addressable market is significant. Organizations that currently cannot access privacy-respecting communication behavioral data include: academic researchers studying misinformation spread, public health agencies tracking information access patterns during health crises, enterprise analytics firms benchmarking communication patterns, and AI training organizations that need consent-based behavioral datasets.

## Terminology

* **Behavioral Pattern Data**: Non-content metadata about how a user communicates — message frequency, network structure (how many contacts, trust tiers), response patterns — that does not reveal what was said or who specifically was involved.
* **ZK-Anonymization Proof**: A zero-knowledge proof generated on-device that proves a user's data contribution meets the anonymization requirements (e.g., unlinkable to their specific DID) without revealing the data itself.
* **Data Query**: A structured question submitted by a data consumer (researcher, enterprise, public health agency) that is answered using aggregated contributions from the data pool. Example: "What is the average daily message frequency for users who have achieved Trust Tier 4?"
* **Contribution Value Score**: An on-chain measure of how frequently a user's behavioral pattern data has been queried. Users with higher contribution value scores receive larger revenue shares.
* **Data Consumer**: An organization (academic institution, enterprise analytics firm, public health agency) that pays ECHO tokens to query the community data pool. Consumers must pass identity verification and agree to usage restrictions.
* **Consent Granularity**: The ability for users to opt in or out of specific data categories — they can contribute message frequency data without contributing network structure data, for example.

## Requirements

### REQ-DATA-001: Opt-In Consent Architecture

**User Story:** As a user, I want complete control over whether and what behavioral data I contribute to the community pool, with the ability to revoke consent at any time, so that I am never surprised by what data I'm sharing.

**Acceptance Criteria:**

* AC-DATA-001.1: The Data Sovereignty Layer shall be strictly opt-in. All users start with zero data contribution. No behavioral data is shared unless the user explicitly activates participation for a specific data category.
* AC-DATA-001.2: Consent shall be granular — users choose from specific categories: message frequency patterns, network size metrics, response time patterns, trust tier distribution of contacts. Each category is individually toggleable.
* AC-DATA-001.3: Users may revoke consent for any or all data categories at any time. Upon revocation, the system shall cease including the user's data in new query responses within 24 hours.
* AC-DATA-001.4: The consent interface shall display in plain language: what data is being shared, what it is used for, how much revenue it has generated for the user, and what queries have used their data type in the past 30 days.
* AC-DATA-001.5: Data Sovereignty features shall not be available to users below Trust Tier 2. Tier 1 (Unverified) users cannot participate, preventing bot-driven data pollution of the pool.

### REQ-DATA-002: Zero-Knowledge Anonymization

**User Story:** As a user participating in the data pool, I want mathematical proof that my contribution is anonymous — not just a promise — so that I can trust that participating does not compromise my privacy.

**Acceptance Criteria:**

* AC-DATA-002.1: Before any user's behavioral pattern data is contributed to the pool, the system shall generate a ZK proof on-device demonstrating: the data contribution is not linkable to the user's specific DID, the data meets the minimum anonymization thresholds (k-anonymity ≥ 50, differential privacy noise applied), and the data category matches what the user consented to.
* AC-DATA-002.2: The ZK proof shall be verifiable by any third party against the public metagraph, enabling independent verification that the data pool's anonymization guarantees hold.
* AC-DATA-002.3: Data contributions shall never include: message content, recipient DIDs, exact timestamps (only bucketed to day/week), geographic data (only country-level at most), or any data that was not explicitly consented to.
* AC-DATA-002.4: The ECHO Protocol Foundation shall publish the anonymization algorithm, noise parameters, and k-anonymity thresholds as open-source specifications auditable by independent researchers.

### REQ-DATA-003: Data Query Marketplace

**User Story:** As a public health researcher, I want to query aggregate communication behavioral patterns from opt-in ECHO users, so that I can study information spread dynamics using privacy-respecting, consent-based data rather than scraped or purchased data.

**Acceptance Criteria:**

* AC-DATA-003.1: Data consumers (researchers, enterprises, public health agencies) shall register with the Foundation, pass identity verification, and agree to usage restrictions before accessing the query marketplace.
* AC-DATA-003.2: Queries shall be submitted as structured data requests specifying: data category, aggregation level, time range, and user population filters (trust tier, region — never individual users).
* AC-DATA-003.3: All query results shall represent aggregated data from a minimum of 1,000 contributing users. No query shall be answerable from fewer than 1,000 users, preventing de-anonymization attacks.
* AC-DATA-003.4: Query pricing shall be set by governance vote (at Phase 3+) or Foundation board (Phase 1–2). Starting price: $0.10–$1.00 per query based on population size and data complexity.
* AC-DATA-003.5: Approved query types at launch: aggregate message frequency by trust tier, aggregate network size distributions, response time distributions, time-of-day activity patterns. Queries that could enable re-identification shall be rejected by the query validation layer.

### REQ-DATA-004: Direct Revenue Distribution to Users

**User Story:** As a user who contributes behavioral data to the community pool, I want to receive ECHO token payments proportional to the value my data generates, so that I am directly compensated for my contribution rather than the platform capturing all the value.

**Acceptance Criteria:**

* AC-DATA-004.1: 70% of all query fees shall flow to contributing users in proportion to their Contribution Value Score. 20% shall go to the Foundation treasury (privacy research fund allocation). 10% shall go to protocol operations.
* AC-DATA-004.2: Revenue distribution shall occur weekly, with ECHO tokens transferred automatically to contributing users' wallets without any claim action required.
* AC-DATA-004.3: The Contribution Value Score shall be calculated as: (frequency of data type queried) × (user data contribution period) × (data quality score). Users with rare or high-demand behavioral patterns receive proportionally more.
* AC-DATA-004.4: Users shall see in the ECHO Wallet tab (Phase 3+) or Profile tracker (Phase 2): total data sovereignty earnings, contribution value score, top data types queried, and estimated future earnings based on current query volume.
* AC-DATA-004.5: Data sovereignty earnings are distinct from token governance and staking. A user can earn from data contributions without holding staked tokens. This creates a second path to ECHO token ownership that rewards data participation, not just financial staking.

## Feature Behavior and Rules

### Why This Is Not Surveillance Capitalism

Surveillance capitalism (Meta, Google) has three characteristics: (1) data is collected without meaningful consent, (2) the platform extracts 100% of the value, and (3) users have no visibility into how their data is used. The Data Sovereignty Layer inverts all three: (1) strict opt-in with granular consent and instant revocation, (2) 70% of value flows directly to users, (3) users see every query type that used their data category. The architecture is fundamentally different — users are producers who set the terms, not resources being mined.

### Phase Gating

The Data Sovereignty Layer requires the token (Phase 3+) for revenue distribution. Before token launch, the infrastructure can be built and tested but revenue cannot be distributed. The Phase 2 consumer launch includes the consent UI and contribution architecture — but the marketplace and revenue distribution activate only when the token is live.

### Regulatory Considerations

The Data Sovereignty Layer engages directly with GDPR, CCPA, and the EU AI Act. The key regulatory principle is genuine consent — users are data controllers who choose to share specific patterns, can revoke at any time, and receive compensation. This is GDPR-compliant data contribution under Article 6(1)(a) (consent) rather than Article 6(1)(f) (legitimate interest), which is how most platforms justify their data collection. Legal review is required before launch in each jurisdiction.

