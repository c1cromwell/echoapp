# Echo - Product Requirements Document (v2.5)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
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

- **WhatsApp's** seamless UX and media sharing
- **Telegram's** extensibility via bots/channels
- **Signal's** gold-standard privacy and E2E encryption model
- **X.com-inspired** trust mechanics (verification badges, trusted circles)
- **IRON SPIDR's** blockchain-anchored security (adapted from federal use cases)

**What makes ECHO decentralized:** ECHO's decentralization comes from three layers that traditional messengers lack entirely. Your identity is self-sovereign (Cardano DIDs — no company owns your account). Your data integrity is blockchain-verified (metagraph consensus — no company can silently alter records). Your message content is mathematically private (E2E encryption — relay servers see only opaque encrypted blobs). The message relay layer uses a client-server model for reliability, but the servers are stateless pipes with no ability to read, alter, or forge message content, and no authority over your identity or data.

### Long-Term Vision: Community-Owned Network State

ECHO's endgame is not a messaging company — it is a **community-owned digital nation**. The messaging platform is the foundation that creates daily engagement, shared identity, and collective economic power. Over time, ECHO evolves from a product users consume into an organization all users co-own:

**Phase 1–4 (Product):** Build a world-class encrypted messaging platform with 1M+ daily active users. Every user earns ECHO tokens through participation. Token holders govern the protocol through stake-weighted voting.

**Phase 5 (Economy):** Launch revenue streams (VIP subscriptions, organization plans, payment rail fees). All revenue flows to a community treasury managed by AI agents — no human executives skimming overhead. The treasury executes two annual programs: ECHO token burns (deflationary pressure) and Bitcoin reserve accumulation (hard-asset backing).

**Phase 6 (Network State):** The community, now economically self-sustaining, begins acquiring real-world assets — land, buildings, companies, infrastructure — for members to use. A governance board (5 founders + 5 community-elected members) oversees strategic decisions. ECHO becomes a digitally-native society with physical territory, shared resources, and democratic governance.

This is the Balaji Srinivasan Network State thesis applied to a messaging platform: start with a highly engaged digital community, build collective economic power, and progressively acquire sovereignty in the physical world.

## Technical Architecture

### Core Components

| Layer | Technology | Decentralization Level | Notes |
|-------|-----------|----------------------|-------|
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

- **DIDs on Cardano**: Self-sovereign identity without central auth servers. Users own their identity across applications.
- **E2E Encrypted Relay**: Messages are end-to-end encrypted on the sender's device before transmission. Relay servers transport opaque ciphertext and cannot read, modify, or forge message content. This follows the same model proven by Signal at scale.
- **Blockchain Anchoring**: Message integrity commitments (Merkle roots of hash commitments, never content) are recorded on the Constellation metagraph, providing cryptographic proof of message authenticity and tamper detection.
- **Zero-Knowledge Proofs**: Privacy-preserving authentication and verification. Prove you meet a trust threshold without revealing your exact score. Prove your age without revealing your birthdate.
- **Stateless Backend**: The Go backend is an operational coordinator and hot cache, not an authority. All persistent state lives on-chain (metagraph for app data/rewards, Cardano for identity). PostgreSQL and Redis serve as performance caches only.

### Constellation Metagraph Deployment Strategy

**Decision: Public Hypergraph Mainnet with Permissioned L1 Validators (Hybrid Model)**

ECHO deploys as a public metagraph on Constellation's Hypergraph mainnet, not a private chain. This is a deliberate choice:

- **Public verifiability is ECHO's value proposition.** ECHO token supply, distribution, and reward claims are publicly auditable by anyone with a block explorer. A private chain would mean "trust us, the balances are real" — the exact problem ECHO exists to solve.
- **IRON SPIDR precedent.** ECHO's PRD cites IRON SPIDR as inspiration. IRON SPIDR started as a private permissioned chain and deliberately transitioned to public. Constellation's own leadership states the future is public networks.
- **Ecosystem network effects.** Public metagraph means ECHO token is visible in Stargazer wallet, tradeable on PacaSwap DEX, eligible for DAG delegation programs, and interoperable with other metagraphs on the Hypergraph. Private chain requires building all tooling from scratch.
- **Privacy is handled at the application layer, not the chain layer.** ECHO's metagraph only stores Merkle roots (hashes), trust commitments (H(score||nonce)), and token transactions. No PII, no message content. The public Hypergraph sees only opaque hashes — privacy is already preserved by design.

**What "hybrid" means in practice:** L1 validators are permissioned (project-operated) during Phases 1–3, controlling who validates ECHO-specific business logic (reward caps, anti-gaming, Merkle structure). L0 nodes submit snapshots to the public Global L0 for immutable recording. Phase 4 opens L1 validation to community operators with ECHO token staking requirements.

**Node Requirements:**

| Node Type | Count | DAG Staking | Role |
|-----------|-------|-------------|------|
| L0 Hybrid Nodes | 3 minimum | 250K DAG each (750K total) | Run both Global L0 and Metagraph L0; submit snapshots to Hypergraph |
| Currency L1 Validators | 3–5 initially | Set by ECHO (e.g., minimum ECHO token stake) | Validate ECHO token transactions, rewards, staking |
| Data L1 Validators | 3–5 initially | Set by ECHO (e.g., minimum ECHO token stake) | Validate Merkle roots, trust commitments, governance |

**Cost Model:**

| Cost Item | Estimate | Notes |
|-----------|----------|-------|
| DAG staking (3 L0 nodes) | 750K DAG (not spent — staked, recoverable) | Capital lockup; L0 nodes earn DAG validator rewards |
| Snapshot fees | Variable; offset by DAG delegation | ~288 snapshots/day at 100K users; fees burned in DAG |
| Node infrastructure | 3 servers minimum (Ubuntu, 8+ cores, 32GB RAM) | AWS/DigitalOcean or bare metal |
| Scala developer | 1 developer for L1 validation logic | Euclid SDK is Scala/JVM; Go backend and iOS unchanged |

**Snapshot Fee Economics:** End users pay zero fees. ECHO as a project pays snapshot fees in DAG to the Hypergraph for each snapshot submitted by the L0 nodes. More delegated DAG staked to ECHO's validators = lower net snapshot fees (delegators subsidize). At scale, fees can potentially be fully rebated through sufficient delegation.

**Technology Stack Note:** All metagraph L1 validation logic (custom consensus, business rules) must be written in Scala using the Euclid SDK / Tessellation framework. This is the code that enforces ECHO-specific rules: daily reward caps, trust-tier multiplier validation, Merkle root structure checks, anti-gaming rules. The Go backend submits data to the metagraph via its REST API. The iOS app is unaffected.

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

**Total Supply: 1,000,000,000 ECHO (fixed, deflationary via Phase 5 burns)**

| Allocation | % | Tokens | Purpose |
|-----------|---|--------|---------|
| Community Rewards | 40% | 400M | Messaging rewards, referrals, staking APY, governance — emitted over 10 years via declining curve |
| Treasury | 22% | 220M | PacaSwap liquidity, DAG staking, Digital Evidence subscriptions, operations, Phase 5–6 |
| Founders (5) | 18% | 180M | 4-year vesting, 1-year cliff, on-chain TokenLock (see Tokenomics doc) |
| Future Team & Advisors | 10% | 100M | Reserved for recruits; same vesting terms |
| Ecosystem & Partnerships | 10% | 100M | PacaSwap LP incentives, DAG delegator rewards, Constellation grants, exchange listings |

**Founder Allocation (18% = 180M ECHO):**

| Founder | Role | % Supply | ECHO |
|---------|------|----------|------|
| Founder 1 | CEO / Visionary / Product | 10.0% | 100M |
| Founder 2 | CTO / Lead iOS Engineer | 2.0% | 20M |
| Founder 3 | Scala / Blockchain Lead | 2.0% | 20M |
| Founder 4 | Head of Growth / Community | 2.0% | 20M |
| Founder 5 | Head of Design / UX | 2.0% | 20M |

All founder tokens are held in on-chain TokenLock positions with 1-year cliff + 36-month monthly vesting, publicly visible on DAG Explorer. No founder can sell any tokens for the first 12 months. After cliff, 1/36th of remaining allocation vests monthly. The blockchain is the cap table.

**Reward Mechanics:**
- ECHO is an L0 token on the Constellation Hypergraph, conforming to the Tessellation v3 L0 token standard
- Messaging rewards: 0.1 ECHO per message (daily cap per trust tier, claimed via AtomicAction bundling tier verification + claim + cap update)
- Payment rail rewards: 1-5 ECHO per transaction
- Referral program: 50 ECHO per verified referral
- Staking via native v3 primitives: users lock ECHO in their own Stargazer wallet (TokenLock), delegate to L1 validators (StakeDelegation), earn 5-15% APY by tier, 14-day withdrawal cooldown (WithdrawLock)
- Anti-gaming: trust-multiplied rewards, daily caps, economic micro-fees at scale; all enforced atomically via AtomicAction
- Year 1 emission: 80M ECHO (20% of community pool); declining annually over 10 years
- PacaSwap DEX: ECHO/DAG and ECHO/USDC liquidity pools for trading and treasury operations
- Phase 5 marketplace payments: time-limited AllowSpend approvals (no unlimited token approvals)

### ECHO Wallet (Stargazer SDK)

ECHO includes a native decentralized wallet built on the Constellation Stargazer Wallet SDK, replacing the concept of a "rewards page" with true asset ownership. The wallet is a primary tab in the iOS app alongside Messaging and Profile.

**Why a wallet, not a rewards page:** A rewards page implies gamification points inside someone else's app. A wallet implies real assets the user owns, controls, and can use across the Constellation ecosystem. For a project whose core value proposition is "all users are owners," the wallet framing is essential.

**Wallet Features:**
- Balance display: available, staked (TokenLock), delegated, pending rewards, USD equivalent
- Staking: lock ECHO via TokenLock, choose tier (Bronze 30d/5%, Silver 90d/8%, Gold 180d/12%, Platinum 365d/15%)
- Delegation: browse validators (uptime, commission, delegated stake), delegate via StakeDelegation, switch instantly
- Rewards: claim pending rewards via AtomicAction, daily cap progress bar, trust tier multiplier display
- Swap (Phase 3+): ECHO ↔ DAG and ECHO ↔ USDC via PacaSwap integration
- Bridge (Phase 3+): ECHO → Base, ECHO → Ink for broader DeFi and exchange access
- Founder vesting display (founders only): allocated, vested, locked, next unlock date, cliff status, "View on DAG Explorer" link
- Transaction history: all staking, delegation, reward, swap, and bridge activity

**External wallet compatibility:** Users can also view and manage ECHO in standalone Stargazer wallet or D'Cent hardware wallet. The ECHO iOS wallet and Stargazer share the same underlying Constellation keypair.

### Revenue Model (Phase 5+)

ECHO is free for all users. Revenue comes from premium tiers and payment rails. All revenue flows to the community treasury — not to a corporation.

| Revenue Stream | Source | Estimated Unit | Treasury Allocation |
|---------------|--------|---------------|-------------------|
| **VIP Subscriptions** | Individual users opting for premium features (larger groups, priority relay, enhanced storage, custom themes, advanced bots, extended disappearing message options) | $4.99–$9.99/month | 100% to treasury |
| **Organization Plans** | Businesses and teams needing compliance audit trails, branded channels, SLAs, admin controls, SSO integration, **Digital Evidence Smart Checkmark on messages, court-admissible audit fingerprinting via Constellation Digital Evidence API, compliance dashboard with public verification URLs** | $10–$50/seat/month | 100% to treasury |
| **Payment Rail Fees** | Small percentage on in-app fiat-to-ECHO conversions, ECHO-to-fiat off-ramps, and merchant payment processing | 0.5–1.5% per transaction | 100% to treasury |
| **Marketplace/Bot Platform** | Revenue share from third-party bots, integrations, and marketplace transactions | 15–30% platform fee | 100% to treasury |

**Key principle:** The platform itself never extracts value. There are no shareholders, no dividends to a parent company, no executive compensation beyond what governance approves. Every dollar of revenue enters the treasury and is allocated by community governance.

### Treasury Management (Phase 5+)

The treasury is managed by AI agents operating under policies set by community governance votes. Human oversight comes from the governance board (see below).

**AI Agent Responsibilities:**

| Agent Role | Operations | Human Override |
|-----------|-----------|----------------|
| **Treasury CFO Agent** | Cash flow monitoring, budget tracking, financial reporting dashboards, surplus calculation | Board reviews quarterly reports |
| **ECHO Burn Agent** | Executes scheduled ECHO token buybacks and burns per governance-approved ratio | Board can pause in emergency |
| **BTC Reserve Agent** | Dollar-cost-averages treasury surplus into Bitcoin per governance-approved allocation | Board approves annual allocation % |
| **Stablecoin Manager** | Manages operational reserves in stablecoins (USDC/DAI), handles yield optimization on idle reserves | Board sets risk parameters |
| **Compliance Agent** | Monitors regulatory requirements, flags transactions needing review, generates audit reports | Board + legal counsel review flags |
| **Reporting Agent** | Generates public real-time treasury dashboards, monthly reports, annual audit preparation | All reports public by default |

**Annual Treasury Allocation (governance-decided, example starting ratios):**

| Allocation | % of Annual Surplus | Purpose |
|-----------|-------------------|---------|
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
|-------|----------|------|------|
| 5 Founders | Permanent (years 1–5); advisory with veto only on existential matters after year 5 | Permanent → advisory | Strategic direction, protocol safety, veto on existential changes (e.g., abandoning E2E encryption) |
| 5 Community Board Members | Elected annually by token-weighted vote (Trust Tier 3+ eligible to stand) | 1 year, re-electable | Oversee treasury AI agents, approve RWA acquisitions, set annual treasury allocation ratios, represent community interests |

**Decision Authority:**

| Decision Type | Who Decides | Threshold |
|--------------|------------|-----------|
| Protocol upgrades (metagraph schema, encryption changes) | All token holders (governance vote) | 67% supermajority |
| Annual treasury allocation ratios | All token holders (governance vote) | Simple majority |
| ECHO burn / BTC buy execution | AI agents (within approved ratios) | Automatic |
| Real-world asset acquisition > $100K | Board (10 members) + governance ratification | Board 7/10 + 60% governance vote |
| Real-world asset acquisition < $100K | Board (10 members) | Board 6/10 majority |
| Emergency protocol changes | Founders (3/5 multi-sig) | 3-of-5 founders |
| Board member removal (misconduct) | All token holders (governance vote) | 75% supermajority |

**Legal Structure (Phase 6):** The DAO requires a legal entity to hold real-world assets (a DAO cannot directly own land in most jurisdictions). Recommended structure: DAO → Wyoming DAO LLC or Marshall Islands DAO LLC → Real-World Asset Holdings. The legal entity is controlled by the governance board, which is controlled by the DAO. All asset titles are held by the legal entity on behalf of the community. Structure to be finalized with legal counsel before first RWA acquisition.

**Open Source:** The entire ECHO codebase — iOS app, Go backend, Scala metagraph validation logic — is open-sourced under a permissive license (MIT or Apache 2.0) once the core product reaches stability (target: Phase 3). Open source ensures no single entity can capture the platform, and allows the community to fork if governance fails.

## Success Metrics

| Metric | Year 1 Target | Year 2 Target | Year 3+ Target |
|--------|--------------|---------------|----------------|
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
- Validate IRON SPIDR parallels
- Build PoC for Cardano DID + E2E encrypted chat via WebSocket relay
- Security whitepaper covering encryption model, relay trust assumptions, and on-chain anchoring
- **Constellation metagraph testnet deployment** using Euclid SDK (Scala)
- Develop and test Data L1 + Currency L1 custom validation logic on testnet (no real DAG required)
- **Implement Tessellation v3 transaction types** in Currency L1 Scala code: TokenLock, StakeDelegation, WithdrawLock, AtomicAction for reward claims, FeeTransaction for snapshot fees
- Acquire or plan acquisition of 750K+ DAG for mainnet L0 node staking
- Evaluate PacaSwap liquidity bootstrapping requirements for ECHO token launch

### Phase 2: Core Build (3-5 months)
- Implement E2E encrypted messaging stack (Kinnami: X25519 + ChaCha20-Poly1305)
- Go backend relay services with WebSocket + APNs push notifications
- iOS native app with Secure Enclave integration and SwiftUI
- Trust scoring: Cardano credential issuance, metagraph trust commitments
- Finalize Data L1 + Currency L1 validation logic in Scala (reward caps via AtomicAction, anti-gaming, Merkle validation, TokenLock/StakeDelegation staking)
- **Deploy metagraph to Constellation Hypergraph mainnet** — 3 L0 hybrid nodes (750K DAG staked), project-operated L1 validators
- ECHO token goes live on public Hypergraph as L0 token; visible in Stargazer wallet and DAG Explorer
- **Seed ECHO/DAG liquidity pool on PacaSwap** — liquidity bootstrapping event for price discovery and initial trading
- **Automate snapshot fee payment** via FeeTransaction from treasury DAG reserves
- Offline message queuing (encrypted store-and-forward)
- Confirm D'Cent hardware wallet compatibility for ECHO cold storage
- Alpha release (100 beta users)

### Phase 3: Feature Polish & Launch (2-3 months)
- Sealed sender implementation (metadata protection)
- Bots/channels framework
- Multi-device sync with device-linked key management
- Group messaging optimization (server-side fan-out, on-chain group metadata)
- Client-side verification: Merkle proofs for message anchoring, snapshot hash verification
- **Begin DAG delegation campaign** — attract DAG holders to delegate to ECHO validators for lower snapshot fees; offer ECHO token incentives to delegators
- **Create ECHO/USDC liquidity pool on PacaSwap** — stablecoin on/off ramp for users and treasury
- **ECHO ↔ Base bridge** — coordinate with 3A DAO to add ECHO as bridgeable L0 token; enables Aerodrome DeFi and treasury BTC accumulation path
- **Digital Evidence integration** — Go backend submits media fingerprints for optional user-initiated image/video verification; prepare enterprise API client for Phase 5 Org tier
- **Midnight evaluation** — assess stability after 6+ months of mainnet; proof-of-concept ZK trust tier verification ("Prove I'm Tier 3+ without revealing my credential")
- Cardano mainnet deployment (identity layer)
- App Store submission

### Phase 4: Scale & Integrate (Ongoing)
- **Open L1 validators to community** — any operator meeting minimum ECHO TokenLock stake can run a Currency L1 or Data L1 validator (L0 nodes still require 250K DAG)
- **Activate validator slashing** — fraudulent validation, double-signing, extended downtime; slashed ECHO to treasury
- **ECHO ↔ Ink bridge** — connect to Kraken exchange via Ink L2; pursue Kraken listing for ECHO
- Federated relay nodes (multiple independent operators, registered on Data L1 with TokenLock stake)
- Optional direct P2P for both-online users
- Optimize for 1M+ users (additional L1 validator nodes, relay scaling)
- Bank pilots with compliance audit trail (IPFS/Storj encrypted logs + Digital Evidence fingerprinting)
- Governance DAO (trust-tier weighted voting on protocol upgrades, metagraph schema changes, slashing thresholds — governance weight = StakedECHO × TrustTierMultiplier)
- **Midnight integration** — ZK trust tier verification live on Midnight mainnet; Org-tier clients get private KYC proofs, group membership proofs, compliance verification without data exposure
- ZK proof system for privacy-preserving verification (via Midnight Compact contracts)
- Android support (StrongBox equivalent of Secure Enclave)
- Explore cross-metagraph interoperability via Hypergraph
- Optional in-app PacaSwap swap interface (ECHO ↔ DAG, ECHO ↔ USDC without leaving ECHO app)

### Phase 5: Community Economy (Year 2–3)
*Prerequisite: 500K+ MAU, stable governance DAO operational*

- **Launch VIP subscriptions and Organization plans** — premium features, compliance tools, enterprise SLAs
- **Organization tier includes Digital Evidence** — Smart Checkmark on messages, automated audit fingerprinting via Constellation Digital Evidence API, compliance dashboard with public verification URLs, data retention proof
- **Deploy AI treasury agents** — CFO agent, burn agent, BTC reserve agent, stablecoin manager, compliance agent, reporting agent
- **AI Burn Agent uses PacaSwap** — buys ECHO from ECHO/DAG pool via atomic swaps, then burns (reduces circulating supply)
- **AI BTC Reserve Agent uses cross-chain bridges** — ECHO → Base bridge → Aerodrome (USDC) → CEX → BTC → cold storage multi-sig
- **AI Stablecoin Manager** uses ECHO/USDC PacaSwap pool and Base bridge for operational reserve management
- **FeeTransaction automation** — AI CFO Agent maintains DAG reserves and pays snapshot fees automatically
- All revenue flows to on-chain community treasury (transparent, auditable on DAG Explorer)
- Community votes on first annual treasury allocation ratios (burn %, BTC %, operational %, RWA fund %, emergency %)
- Launch public real-time treasury dashboard (AI-generated, on-chain verifiable via DAG Explorer)
- Payment rail integration — AllowSpend + SpendTransaction for subscription auto-renewals, bot payments, marketplace escrow (time-limited approvals only, never unlimited)
- Bot/integration marketplace with revenue share to treasury
- Elect first 5 community board members (annual election, Trust Tier 3+ eligible)
- Open-source entire codebase (iOS, Go, Scala) under permissive license
- Engage legal counsel for DAO LLC formation (Wyoming or Marshall Islands)

### Phase 6: Network State Formation (Year 3+)
*Prerequisite: 1M+ MAU, self-sustaining treasury, legal entity established*

- **Establish legal entity** (DAO LLC) to hold real-world assets on behalf of the community
- **First real-world asset acquisition** — community votes on target (co-working space, community housing, or similar high-utility asset for members)
- Expand RWA portfolio based on community governance: land, buildings, companies, infrastructure
- Network State membership tiers — physical access tied to ECHO token staking levels
- Partnerships with existing Network State projects and digital nomad communities
- Cross-metagraph alliances — interoperability agreements with complementary Hypergraph metagraphs
- Explore sovereign recognition pathways (special economic zones, free zones, charter cities)
- Scale AI agent layer: property management agent, investment analysis agent, member services agent
- Annual board elections become a signature community event
- Long-term goal: ECHO community as a recognized digital jurisdiction with physical presence across multiple geographies

## Budget Estimate

**Phase 1–4 (Product Build): $500K - $2M**

- Development team (5-10 blockchain + mobile experts, including at least 1 Scala/JVM developer for metagraph L1 validation logic)
- Security audits (E2E encryption, Secure Enclave integration, metagraph validation logic, Scala L1 code review)
- **750K DAG staking** for 3 L0 hybrid nodes on Constellation Hypergraph mainnet (capital lockup, not expenditure — recoverable; nodes earn DAG validator rewards)
- Constellation metagraph node infrastructure (3+ servers, ~$300-500/month)
- Constellation snapshot fees in DAG (offset by delegation; estimated low at launch volumes)
- Cardano transaction fees (credential issuance from platform treasury, ~15,000 ADA/month at 100K users)
- IPFS/Storj pinning costs (~$70/month at 100K users)
- Marketing and launch

**Phase 5 (Community Economy): Self-Funding**

Once VIP subscriptions, Organization plans, and payment rail fees are generating revenue, ECHO becomes self-sustaining. All operational costs (infrastructure, security audits, development grants) are funded from treasury. The development team transitions from founder-funded to treasury-funded via governance-approved budgets.

**Phase 6 (Network State): Treasury-Funded**

Real-world asset acquisitions are funded from the RWA allocation of treasury surplus. Scale depends entirely on revenue growth and community governance decisions. No external fundraising required — the community funds its own expansion.

**Note on external funding:** ECHO is designed to *not require* venture capital. VC funding creates misaligned incentives — investors want returns, which means extracting value from users. ECHO's model is the opposite: all value stays in the community. If early-stage funding is needed before revenue, it should come from token presale to aligned community members, Constellation ecosystem grants, or founder capital — not from VCs who would expect equity or governance control.

---

*For detailed API specifications, see [docs/api/openapi.yaml](./api/openapi.yaml)*
*For data layer architecture, see [DATA_LAYER_ARCHITECTURE.md](./DATA_LAYER_ARCHITECTURE.md)*
*For iOS frontend architecture, see [ios-frontend-architecture-blueprint-v2.md](./ios-frontend-architecture-blueprint-v2.md)*
*For backend architecture, see [BACKEND_ARCHITECTURE_IMPLEMENTATION.md](./BACKEND_ARCHITECTURE_IMPLEMENTATION.md)*
