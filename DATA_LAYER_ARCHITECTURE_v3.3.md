# Data Layer Architecture — v3.3

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 3.3 | March 7, 2026 | Finalized tokenomics: 1B supply, 5-pool allocation, on-chain founder vesting via TokenLock. Resolved wallet architecture: Stargazer SDK native wallet in iOS app. Added token genesis mechanics to data flows. Added emission curve enforcement to Currency L1 validation. Resolved open question #13 (Stargazer SDK wallet). |
| 3.2 | March 7, 2026 | Aligned with Constellation ecosystem (Tessellation v3 + Digital Evidence + DeFi). Added v3 transaction type mapping (TokenLock, StakeDelegation, AtomicAction, AllowSpend, WithdrawLock, FeeTransaction). Added validator slashing spec. Added L0 token standard compliance. Added Digital Evidence integration for enterprise audit, media fingerprinting, and Smart Checkmark. Added DeFi/liquidity section (PacaSwap, cross-chain bridges, Stargazer wallet). Added PRO consensus reference. Resolved open question #3. Tessellation v4 readiness notes. |
| 3.1 | March 1, 2026 | Resolved Constellation deployment model: public Hypergraph mainnet metagraph with permissioned L1 (hybrid). Added metagraph deployment specification (node requirements, DAG staking, snapshot fee economics, Scala/JVM stack, phased rollout). Added metagraph cost model. Updated governance with L1 validator admission detail. Resolved open question #2. |
| 3.0 | February 23, 2026 | Resolved messaging transport: client-server relay with decentralized anchoring. Added message relay layer as core component. Updated all data flows. Added offline message queuing. Added metadata protection roadmap. Added group messaging architecture. Removed P2P/libp2p from open questions. Updated open questions list. |
| 2.0 | February 23, 2026 | Cross-document review. Added cross-chain consistency, performance targets, Cardano spec, log lifecycle, fault tolerance, anti-spam, governance, encryption spec. |
| 1.0 | February 2026 | Initial blueprint |

---

## 1. Overview

The data layer is built on a multi-blockchain architecture combining Constellation metagraph infrastructure for application data and ECHO rewards, Cardano for identity and verifiable credentials, and decentralized storage for logging and audit trails. A Go backend microservices layer acts as an operational coordinator, stateless message relay, and hot cache between client applications and on-chain state.

**Design Principles:**

- On-chain as source of truth; off-chain as performance optimization
- Zero PII on any blockchain (enforced by T0–T7 data classification)
- Each chain handles one concern: identity (Cardano), application state + rewards (metagraph on public Hypergraph), audit (IPFS/Storj), evidence (Digital Evidence API)
- The Go backend is a relay and cache, not an authority — it cannot read message content, does not own user identities, and does not control token balances
- Message relay is client-server for reliability; decentralization comes from identity, data integrity, and encryption layers
- Public Hypergraph mainnet for token and data integrity verifiability; permissioned L1 validators for controlled business logic (progressive decentralization to permissionless in Phase 4)
- Use native Tessellation v3 transaction primitives (TokenLock, StakeDelegation, AtomicAction, AllowSpend, WithdrawLock, FeeTransaction) rather than custom implementations — interoperability with Stargazer, DAG Explorer, and PacaSwap requires standard types
- Metagraph validation logic in Scala (Euclid SDK); Go backend and iOS app interact via REST API
- Proof of Reputable Observation (PRO) consensus — Constellation's DAG-based consensus model enables parallel transaction processing, near-zero fees for end users, and real-time data validation

---

## 2. Core Components

### 2.1 Constellation Metagraph (Application & Rewards Layer)

The primary data layer uses Constellation Network's metagraph architecture for high-throughput, decentralized consensus on application data and token transactions.

**Deployment Model: Public Hypergraph Mainnet with Permissioned L1 (Hybrid)**

ECHO deploys as a public metagraph on Constellation's Hypergraph mainnet. L0 nodes submit snapshots to the public Global L0 for immutable recording. L1 validators are permissioned (project-operated) initially, transitioning to community-operated in Phase 4.

This is not a private metagraph. Rationale:

- **Token credibility requires public verifiability.** ECHO token supply, distribution, and reward claims must be publicly auditable. A private chain would mean "trust us" — the exact problem ECHO solves.
- **IRON SPIDR precedent.** ECHO cites IRON SPIDR, which deliberately transitioned from private to public. Constellation's leadership states the future is public networks, not private.
- **Ecosystem network effects.** Public metagraph = ECHO token visible in Stargazer wallet, tradeable on PacaSwap DEX, eligible for DAG delegation, interoperable with other Hypergraph metagraphs.
- **Privacy is application-layer, not chain-layer.** The public Hypergraph sees only Merkle roots (opaque hashes), trust commitments (H(score||nonce)), and token transactions. No PII, no message content ever reaches any chain.

**Node Infrastructure:**

| Node Type | Minimum Count | DAG Requirement | Operator | Role |
|-----------|--------------|-----------------|----------|------|
| L0 Hybrid Nodes | 3 | 250K DAG staked per node (750K total) | Project-operated (all phases) | Run Global L0 + Metagraph L0 processes on same server; submit snapshots to Hypergraph |
| Currency L1 Validators | 3 (launch), 5+ (scale) | ECHO token stake (set by project) | Project-operated (Phase 1–3); community (Phase 4) | Validate ECHO token transfers, rewards, staking |
| Data L1 Validators | 3 (launch), 5+ (scale) | ECHO token stake (set by project) | Project-operated (Phase 1–3); community (Phase 4) | Validate Merkle roots, trust commitments, governance votes |

**L0 Hybrid Node Detail:** Each L0 node must run both the Global L0 process and the Metagraph L0 process on the same server. This means ECHO's L0 nodes participate in the Hypergraph's global consensus while simultaneously managing ECHO's metagraph snapshots. L0 nodes earn DAG validator rewards from the Hypergraph in addition to any ECHO-specific rewards.

**Server Requirements (per node):** Ubuntu 20.04/22.04, 8+ CPU cores, 32GB+ RAM, SSD storage, stable network. Recommended: AWS m5.2xlarge, DigitalOcean Premium CPU-Optimized, or equivalent bare metal.

**Snapshot Fee Economics:**

Metagraphs pay snapshot fees in DAG to the Hypergraph for each snapshot submitted by L0 nodes. End users pay zero fees — ECHO as a project absorbs snapshot costs.

| Factor | Detail |
|--------|--------|
| Snapshot cap | 50KB per snapshot (each can contain many transactions) |
| ECHO snapshot frequency | ~1 per 5 seconds (metagraph default) |
| Cost reduction via delegation | More DAG delegated to ECHO's validators = lower net snapshot fees |
| Delegation incentive | ECHO can offer L0 token rewards or other incentives to DAG delegators |
| Fee pass-through | ECHO may optionally pass micro-fees to users at scale (0.001 ECHO per submission) but this is a product decision, not a protocol requirement |

At 100K users with ~288 message-integrity snapshots/day plus currency snapshots, the DAG snapshot fee cost is manageable and heavily offset by delegation. Constellation's metanomics model can even result in full fee rebates with sufficient delegation.

**Metagraph L1 Validation Logic (Scala/JVM):**

All custom validation logic for ECHO's Data L1 and Currency L1 must be written in Scala using the Euclid SDK (built on Constellation's Tessellation framework). This is the code that enforces ECHO-specific business rules on-chain:

| Validation Rule | L1 Layer | Logic |
|----------------|----------|-------|
| Daily reward cap | Currency L1 | Reject reward claims exceeding per-DID daily maximum for each reward type |
| Trust-tier multiplier | Currency L1 | Apply correct multiplier based on cached trust tier; reject mismatched multipliers |
| Anti-gaming | Currency L1 | Detect and reject suspicious reward patterns (velocity checks, repeat claims) |
| Merkle root structure | Data L1 | Validate submitted Merkle roots have correct structure, authorized sender DID |
| Trust commitment format | Data L1 | Validate H(score||nonce) has correct hash length and authorized issuer |
| Governance vote rules | Data L1 | One-vote-per-DID, active proposal check, minimum stake requirement |
| Schema version check | Both | Reject submissions with unsupported schema versions |

The Go backend and iOS app are unaffected — they interact with the metagraph through its REST API. Only the on-chain validation layer requires Scala.

**Development Environment:** The Euclid SDK provides Docker-based local development clusters via the Hydra CLI tool, including a developer dashboard and telemetry monitoring. Local development spins up all layers (Global L0, Metagraph L0, Currency L1, Data L1) in Docker containers.

**Phased Deployment:**

| Phase | Metagraph State | Details |
|-------|----------------|---------|
| Phase 1 | **Testnet** | Euclid SDK local development + Constellation testnet. Build and test all L1 validation logic. No real DAG required. Acquire 750K+ DAG for mainnet. |
| Phase 2 | **Mainnet — Permissioned L1** | Deploy 3 L0 hybrid nodes on Hypergraph mainnet (750K DAG staked). ECHO token live. Project operates all L1 validators. Snapshot fees begin. ECHO visible in Stargazer wallet. |
| Phase 3 | **Mainnet — DAG Delegation** | Launch delegation campaign: attract DAG holders to delegate to ECHO validators for lower snapshot fees. Offer ECHO token incentives. Open additional L1 validator slots selectively. |
| Phase 4 | **Mainnet — Permissionless L1** | Any operator meeting minimum ECHO stake can run L1 validators. L0 nodes still require 250K DAG. DAO governance over schema changes and validation rule updates. Relay node registry on Data L1. |

**Metagraph Structure:**

| Layer | Role | Scaling Model |
|-------|------|---------------|
| **Currency L1** | Validates ECHO reward token transactions, manages balances, processes staking state changes | Horizontal (add validator nodes) |
| **Data L1** | Validates domain-specific application data with custom business logic | Horizontal (add validator nodes) |
| **Metagraph L0** | Aggregates validated L1 blocks into metagraph snapshots (finalized state) | Vertical (more powerful nodes) |
| **Global L0 (Hypergraph)** | Final consensus, immutable recording of all metagraph snapshots | Vertical |

**Data L1 Validated Data Types:**

| Data Type | On-Chain Content | Privacy | Validation Rules |
|-----------|-----------------|---------|-----------------|
| Message integrity | Merkle root of batch commitments (never content) | T3 compliant | Valid Merkle structure, authorized sender DID |
| Trust commitments | H(trust_score \|\| nonce) | T6 compliant | Valid commitment format, authorized issuer |
| Reward claims | Claim type, DID, amount, trust tier | T7 (public) | Daily cap check, trust multiplier validation, anti-gaming rules |
| Governance votes | Proposal ID, DID, vote, stake weight | T7 (public) | Active proposal, sufficient stake, one-vote-per-DID |
| Staking operations | DID, amount, tier, lock duration (via TokenLock) | T7 (public) | Minimum amounts per tier, valid lock period, founder vesting cliff/schedule enforcement |
| Group metadata | Group ID, member count hash, admin DID | T7 (public) | Valid admin signature, member count bounds |
| Relay node registry | Node DID, endpoint, stake, uptime | T7 (public) | Minimum stake, valid endpoint (Phase 4) |

**Data L1 Schema Versioning:**

- Each data submission includes a `schema_version` field
- L1 validators support the current version and one prior version
- Schema upgrades are coordinated via governance proposal → vote → activation at a future snapshot height
- Breaking changes require a new Data L1 deployment (metagraph upgrade)

**Performance Targets:**

| Metric | Launch (100K users) | Scale (1M users) |
|--------|-------------------|------------------|
| Data L1 TPS | 50 | 500 |
| Currency L1 TPS | 100 | 1,000 |
| End-to-end finality | < 10 seconds | < 15 seconds |
| Metagraph snapshot interval | 5 seconds | 5 seconds |
| L1 validator nodes | 5 | 20+ |
| L0 nodes | 3 | 5 |

**Tessellation v3 Transaction Type Mapping:**

ECHO uses native Tessellation v3 transaction primitives for all token operations. Custom implementations are used only for ECHO-specific business logic (reward caps, anti-gaming) that v3 primitives don't cover. Using native types ensures interoperability with Stargazer wallet, DAG Explorer, PacaSwap DEX, and cross-chain bridges.

| v3 Primitive | ECHO Operation | L1 Layer | Notes |
|-------------|---------------|----------|-------|
| **TokenLock** | ECHO staking (lock ECHO in user's own Stargazer wallet for 5–15% APY) | Currency L1 | Replaces custom staking contract. Tokens locked but remain in user's wallet. Lock duration enforced by ECHO-specific L1 validation (Scala). |
| **StakeDelegation** | Delegate locked ECHO to an L1 validator to earn rewards | Currency L1 | Users choose a Currency L1 or Data L1 validator. Delegation increases validator's consensus weight. Delegators earn proportional share of emission rewards. |
| **WithdrawLock** | Initiate ECHO unstaking with 14-day cooldown | Currency L1 | 14-day cooldown (governance-adjustable). Shorter than DAG's 21-day period — ECHO is an application token, not a base-layer security asset. |
| **AtomicAction** | Bundle reward claim + trust tier verification + daily cap update as single all-or-nothing transaction | Both L1s | Eliminates race conditions in reward claims. Also used for: governance vote + stake verification, staking tier change + balance update. |
| **AllowSpend** | Time-limited approval for bot/marketplace payments (Phase 5) | Currency L1 | Explicitly time-limited — Constellation designed this to avoid Ethereum's unlimited approval vulnerability. Used for: subscription auto-renewals, bot payments, marketplace escrow. |
| **SpendTransaction** | Execute payment against an AllowSpend approval | Currency L1 | Paired with AllowSpend for Phase 5 payment rails. |
| **FeeTransaction** | Automated snapshot fee payment from ECHO treasury DAG reserves | Currency L1 | Eliminates manual snapshot fee management. AI Treasury CFO Agent triggers FeeTransaction to keep snapshot fees current. |
| **Atomic Swap** (AtomicAction) | Two-sided ECHO ↔ DAG or ECHO ↔ USDC swaps for treasury operations | Currency L1 | AI Burn Agent uses atomic swaps via PacaSwap for ECHO buyback. Guarantees both sides execute or neither does. |

**ECHO Staking Model (via v3 Primitives):**

| Parameter | Value | v3 Primitive |
|-----------|-------|-------------|
| Lock mechanism | ECHO tokens locked in user's own Stargazer wallet | TokenLock |
| Delegation | User delegates locked ECHO to L1 validator of choice | StakeDelegation |
| APR (base) | 5–15% by staking tier | Emission schedule in Currency L1 Scala validation |
| Reward distribution | Proportional to delegated stake weight per validator | Currency L1 validation |
| Validator switch | Instant re-delegation (no cooldown to switch validators) | New StakeDelegation tx |
| Withdrawal cooldown | 14 days (governance-adjustable) | WithdrawLock |
| Delegator slashing | None — delegators never lose staked ECHO | Protocol-level |
| Minimum stake | Governance-set per tier (e.g., 100 ECHO for Tier 1, 1000 for Tier 2, etc.) | TokenLock minimum enforced by L1 |
| Custody | User retains full custody at all times; ECHO never leaves their wallet | TokenLock design |

**L0 Token Standard Compliance:**

ECHO token is an L0 token on the Constellation Hypergraph, conforming to the Tessellation v3 L0 token standard. This is required for interoperability with:

- Stargazer wallet (display, send, receive, stake, delegate)
- DAG Explorer (supply visibility, transaction history, delegation tracking, validator rankings)
- PacaSwap DEX (ECHO/DAG and ECHO/USDC liquidity pools, atomic swaps)
- Cross-chain bridges (ECHO ↔ Base, ECHO ↔ Ink for CEX access)
- D'Cent hardware wallet (cold storage for ECHO)
- AllowSpend / SpendTransaction compatibility for Phase 5 marketplace and bot payments

**Validator Slashing (Phase 4 — Permissionless Validators):**

When ECHO L1 validators transition to permissionless in Phase 4, the following slashing conditions activate. Slashing is enforced by the Metagraph L0 layer based on cryptographic evidence from peer L1 validators.

| Offense | Penalty | Detection | Recovery |
|---------|---------|-----------|----------|
| Validating fraudulent reward claims (inflated amounts, exceeded daily caps) | 10% of staked ECHO slashed + 30-day suspension | Cross-validation by peer L1 nodes; conflicting validation proofs | Re-stake after suspension period |
| Submitting invalid Merkle roots (malformed structure, unauthorized sender DID) | 5% of staked ECHO slashed + warning | Data L1 consensus rejection | Automatic after fix |
| Extended downtime (>24h continuous offline) | 1% of staked ECHO slashed per 24h block | L0 heartbeat monitoring | Automatic on reconnect |
| Double-signing (conflicting blocks at same snapshot height) | 50% of staked ECHO slashed + permanent ban | Cryptographic proof of conflicting signatures | None — permanent |
| Colluding to bypass anti-gaming rules | 25% of staked ECHO slashed + permanent ban | On-chain pattern detection + governance review | Governance vote can reverse if false positive |

Slashed ECHO tokens are sent to the community treasury (not burned). Governance decides allocation of slashed funds (operational reserve, additional burns, or community grants). Delegators are never slashed — only the validator's own staked ECHO is at risk.

**Tessellation v4.0 Readiness:**

Tessellation v4.0 is in release candidate phase (RC6, February 2026). Full feature set is not yet public. ECHO must:

- Monitor v4 release notes and migration guides
- Test all Scala L1 validation logic against v4 APIs on testnet before mainnet migration
- Plan rolling upgrade of L0 and L1 nodes using nodectl utility
- Migrate snapshot data directories (Constellation provides migration scripts)
- Budget 1–2 weeks of engineering for metagraph upgrade testing

### 2.2 Cardano Identity Layer

Cardano serves as the verifiable credential registry, fully separated from application logic and rewards.

**Network:** Cardano Mainnet (testnet during development phases 1–2)

**Credential Standard:** W3C Verifiable Credentials (VCs) using `did:prism` (Atala PRISM / Veridian) DID method

**On-Chain Data Structures:**

| Structure | Storage Method | Content |
|-----------|---------------|---------|
| DID Document | Cardano metadata (CIP-25 extended) | Public keys, service endpoints, controller |
| Credential Schema | Plutus reference script | Schema definition, version, issuer DID |
| Credential Status | Bit vector in UTXO datum | Revocation status per credential index |
| Trust Tier | UTXO datum | Tier level (1–5), issuer DID, expiry |
| Verification Record | Transaction metadata | Verifier DID, method used, timestamp |

**Trust Levels:**

| Tier | Method | On-Chain Record | Feature Access |
|------|--------|----------------|----------------|
| 1 — Unverified | Self-registration | DID Document only | Basic messaging, no rewards |
| 2 — Newcomer | Email/phone verification | + Verification record | Messaging + basic rewards |
| 3 — Member | Third-party IDV (Stripe Identity / Sumsub) | + Credential status bit | Full rewards, group creation |
| 4 — Verified | Government ID (Apple Digital ID or IDV provider) | + Trust tier datum | Enhanced rewards multiplier, payment rails |
| 5 — Trusted | Peer attestations + sustained activity | + Web-of-trust attestation chain | Maximum multiplier, governance participation |

**Cost Model:**

- Credential issuance: ~0.3–0.5 ADA per transaction (paid from platform treasury)
- Revocation update: ~0.2 ADA (batch multiple revocations per transaction)
- Estimated monthly cost at 100K users with 30% verified: ~15,000 ADA
- Fee delegation: Platform submits transactions on behalf of users; users never need to hold ADA

**Revocation Mechanism:**

- Bit vector stored in a Plutus UTXO datum
- Each credential gets an index; setting the bit to 1 revokes
- Verifiers check the bit vector before accepting a credential
- Batch revocation: multiple credentials revoked in a single transaction

### 2.3 Message Relay Layer

The message relay layer transports E2E encrypted messages between clients. It is a stateless, content-blind relay — it handles ciphertext blobs it cannot read, decrypt, or modify.

**Architecture Decision: Client-Server Relay (not P2P)**

ECHO uses a client-server WebSocket relay model. See the PRD v2.0 "Messaging Architecture Rationale" for the full decision rationale. Summary: iOS platform constraints (aggressive background process killing), offline delivery requirements, group fan-out at scale (1M members), and push notification requirements make pure P2P unviable for a consumer iOS messaging product. ECHO's decentralization value comes from identity (Cardano DIDs), data integrity (metagraph consensus), and content privacy (E2E encryption) — not from the transport layer.

**Relay Server Capabilities and Limitations:**

| Relay Server CAN | Relay Server CANNOT |
|------------------|-------------------|
| Transport encrypted blobs between clients | Read, decrypt, or modify message content |
| Queue encrypted messages for offline recipients | Forge messages (clients verify sender signatures) |
| Deliver push notifications via APNs | Access private keys or identity credentials |
| Track delivery status (sent, delivered, read receipt) | Override metagraph or Cardano state |
| Rate-limit abusive senders | Associate encrypted content with real-world identity |
| See sender DID, recipient DID, timestamp, blob size | See message plaintext, attachments, or reactions |

**Relay Infrastructure:**

```
Phase 1–3: Centralized relay cluster
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Relay Pod 1 │     │  Relay Pod 2 │     │  Relay Pod N │
│  (Stateless)  │     │  (Stateless) │     │  (Stateless) │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       └────────────────────┼────────────────────┘
                            │
               ┌────────────▼────────────┐
               │      Load Balancer      │
               │    (Regional, TLS 1.3)  │
               └────────────┬────────────┘
                            │
                    iOS App (WebSocket)

Phase 4: Federated relay network
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Operator A       │  │  Operator B       │  │  Operator C       │
│  (Staked on       │  │  (Staked on       │  │  (Staked on       │
│   metagraph)      │  │   metagraph)      │  │   metagraph)      │
└────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
                    Relay Node Registry
                    (Data L1 on-chain)
```

**Offline Message Queuing:**

When a recipient is offline, the relay server holds encrypted messages in a temporary queue:

| Property | Value |
|----------|-------|
| Storage | Redis (encrypted blobs in memory) or PostgreSQL (encrypted at rest) |
| Retention | Maximum 30 days; configurable per user |
| Encryption | Messages are already E2E encrypted by sender; server stores opaque blobs |
| Delivery | On recipient reconnect (WebSocket), server drains queue in order |
| Size limit | 1000 queued messages per recipient; oldest evicted if exceeded |
| Push notification | APNs notification sent immediately on queue insertion |

**Metadata Protection Roadmap:**

| Phase | Protection Level | Method | Server Sees |
|-------|-----------------|--------|-------------|
| 1–2 | Baseline | TLS 1.3 transport; auth token per session | Sender DID, recipient DID, timestamp, blob size |
| 3 | Sealed sender | Sender identity encrypted inside the E2E envelope; server routes by recipient only | Recipient DID, timestamp, blob size (sender hidden) |
| 4 | Federated relay | Traffic split across independent operators; no single operator sees all traffic | Each operator sees only its routed fraction |
| 4+ | Optional P2P | When both parties are online, establish direct WebSocket via relay-assisted signaling | Relay sees connection setup only, not message traffic |

**Sealed Sender (Phase 3) Detail:**

Sealed sender follows Signal's proven approach. The sender encrypts their DID inside the E2E message envelope. The outer envelope contains only the recipient's DID and an encrypted delivery token. The relay server can route the message to the recipient but cannot determine who sent it.

```
Outer envelope (visible to relay):
  - Recipient DID
  - Encrypted delivery token (proves sender is registered, without revealing identity)
  - E2E ciphertext blob

Inner envelope (visible only to recipient after decryption):
  - Sender DID
  - Message content
  - Commitment hash
  - Signature
```

### 2.4 Go Backend Operational Layer

The Go backend is **not** a centralized authority — it is an operational coordinator and hot cache that sits between clients and on-chain state. With the relay model, the backend also serves as the message relay infrastructure (Phase 1–3), which transitions to federated relay operators in Phase 4.

**Role Clarification:**

| Function | Authoritative Source | Backend Role |
|----------|---------------------|-------------|
| Token balances | Currency L1 (metagraph) | Read-through cache (TTL: 5s) |
| Trust scores | Cardano + metagraph commitment | Compute engine + cache (TTL: 60s) |
| Message content | Device-local (E2E encrypted) | Relay only (never stores plaintext; queues ciphertext for offline) |
| Message metadata | Data L1 (Merkle root) | Batch aggregator before submission |
| User identity | Cardano DID | Cache + credential proof validator |
| Reward eligibility | Data L1 validators | Pre-validator (reject obviously invalid claims) |
| Message relay | N/A (stateless transport) | WebSocket relay + APNs push + offline queue |

**Scaling Architecture:**

```
                    ┌─────────────────────┐
                    │   Load Balancer      │
                    │   (Regional, TLS)    │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼──────┐ ┌──────▼────────┐ ┌─────▼───────┐
     │ Go Instance 1 │ │ Go Instance 2 │ │ Go Instance N│
     │ (Stateless)   │ │ (Stateless)   │ │ (Stateless)  │
     └───────┬───────┘ └──────┬────────┘ └──────┬───────┘
             │                │                  │
     ┌───────▼────────────────▼──────────────────▼───────┐
     │                 Shared Services                     │
     │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
     │  │  Redis    │  │  NATS    │  │  PostgreSQL      │  │
     │  │(Cache +   │  │  (Events)│  │  (Operational +  │  │
     │  │ Msg Queue)│  │         │  │   Offline Queue) │  │
     │  └──────────┘  └──────────┘  └──────────────────┘  │
     └───────┬────────────────┬──────────────────┬───────┘
             │                │                  │
     ┌───────▼───┐    ┌──────▼──────┐    ┌──────▼──────┐
     │ Metagraph │    │   Cardano   │    │  IPFS/Storj │
     │ (L1/L0)   │    │  (Identity) │    │  (Logging)  │
     └───────────┘    └─────────────┘    └─────────────┘
```

**Circuit Breakers:** Each downstream connection (metagraph, Cardano, IPFS/Storj) has an independent circuit breaker. Message relay continues even if all chains are unavailable — messages are transported as encrypted blobs regardless of on-chain status. On-chain operations (rewards, commitments, credential checks) degrade gracefully to cached state.

**Anti-Spam and Rate Limiting:**

| Level | Mechanism | Limit |
|-------|-----------|-------|
| API Gateway | Per-device token bucket | 100 requests/minute |
| Message relay | Per-DID send rate | 60 messages/minute (adjustable by trust tier) |
| Reward claims | Per-DID daily cap | Configurable per reward type (see tokenomics) |
| Data L1 submission | Per-DID submission rate | 10 data updates/minute |
| Metagraph write | Economic cost | Micro-fee (0.001 ECHO per submission at scale) |
| WebSocket | Per-connection message rate | 60 messages/minute |
| Offline queue | Per-recipient queue depth | 1000 messages max; oldest evicted |

### 2.5 Decentralized Logging & Storage

**Storage Provider:** IPFS (with Pinata / web3.storage pinning) for immutable logs. Storj as fallback for large media audit trails.

**Log Lifecycle:**

| Phase | Action | Details |
|-------|--------|---------|
| **Collection** | Go backend batches API events, relay metadata, and metagraph transaction receipts | In-memory buffer, max 1000 events or 5 minutes |
| **Encryption** | Batch encrypted with AES-256-GCM using a rotating log encryption key | Key derived from platform master key via HKDF with date-based info string |
| **Submission** | Encrypted batch pushed to IPFS; CID recorded | Retry with exponential backoff on failure |
| **Pinning** | CID pinned via Pinata/web3.storage (primary) + self-hosted IPFS node (secondary) | Minimum 2 pin providers for redundancy |
| **Indexing** | CID + time range + batch hash submitted to Data L1 | Enables on-chain verifiable log index |
| **Retrieval** | Authorized auditors decrypt with log key (threshold scheme: 3-of-5 key holders) | Access logged on-chain as audit access event |
| **Retention** | Minimum 7 years for compliance; pins maintained by platform treasury | Unpinning only after retention period + governance vote |

**What relay metadata is logged (privacy-safe):**

- Message count per time window (no content, no DIDs unless required for compliance)
- Delivery success/failure rates
- Queue depth statistics
- Rate limit trigger events
- Circuit breaker state changes

**Key Management for Logs:**

- Log encryption keys are derived monthly from a platform master key
- Master key is split using Shamir's Secret Sharing (3-of-5 threshold)
- Key holders are designated platform operators (expandable to DAO members at Phase 4)
- Key rotation does not affect retrieval of older batches (each batch records its key epoch)

**Compression and Cost:**

- Batches compressed with zstd before encryption
- Estimated cost at 100K users: ~$50/month (IPFS pinning) + ~$20/month (Storj overflow)
- Batch size target: 1–5 MB compressed per batch
- Monitoring: alert if batch sizes exceed 10 MB (indicates anomalous activity)

### 2.6 Digital Evidence Integration (Enterprise)

Constellation's Digital Evidence is a managed REST API service that anchors cryptographic fingerprints (SHA-256 hashes) on the Hypergraph. It provides a public verification explorer, "Smart Checkmark" certification, and enterprise compliance packaging — all without requiring ECHO to manage additional infrastructure.

**Relationship to ECHO's Custom Merkle Anchoring:**

ECHO's custom Merkle root pipeline (Section 3.1, step 4) remains the core message integrity system for all users. It is optimized for high-volume batch anchoring (1000+ commitments per 5-minute batch) and is deeply integrated with the relay architecture. Digital Evidence is layered on top as a premium enterprise feature for Organization tier clients who need individual-event verification, public proof URLs, and court-admissible evidence packaging.

| Feature | Core Merkle Pipeline (All Users) | Digital Evidence Layer (Org Tier) |
|---------|--------------------------------|----------------------------------|
| Purpose | Batch message integrity anchoring | Individual-event enterprise compliance |
| Granularity | 1 Merkle root per batch (~1000 commitments) | 1 fingerprint per auditable event |
| Cost | Included in ECHO metagraph snapshot fees | Digital Evidence subscription (passed through in Org pricing) |
| Verification | Client-side Merkle proof (Phase 3) | Public Explorer URL + Smart Checkmark (turnkey) |
| Infrastructure | ECHO metagraph (self-operated) | Constellation-managed |
| Compliance | Custom tooling needed | Designed for legal/compliance out of the box |

**ECHO Digital Evidence Use Cases:**

| Use Case | Data Fingerprinted | Phase | Integration Point |
|----------|-------------------|-------|-------------------|
| Enterprise audit trail | IPFS CID of encrypted conversation log batch + batch metadata hash | Phase 5 | Go backend submits fingerprint via Digital Evidence REST API after each IPFS batch push |
| Media authenticity | SHA-256 of image/video/document before E2E encryption | Phase 3+ | iOS client computes hash, submits via Digital Evidence API, embeds Event ID + verification URL in message metadata |
| Smart Checkmark messages | Hash of message content + timestamp + sender DID | Phase 5 | Backend submits fingerprint for Org-tier messages; iOS renders checkmark icon alongside `.anchored` status |
| Data retention proof | Hash of audit batch at retention boundary + deletion confirmation | Phase 5 | Backend submits fingerprint when regulatory retention period expires and data is deleted; proves data existed and was retained |

**Digital Evidence API Integration (Go Backend):**

```
Go Backend — Digital Evidence Client
├─ Configuration: API key + organization ID + tenant ID (from Constellation)
├─ Fingerprint submission: POST /evidence with signed SHA-256 hash payload
├─ Verification lookup: GET /evidence/{eventId} for compliance dashboards
├─ Smart Checkmark rendering: Event ID + Explorer URL stored in message metadata
└─ Rate limiting: Batch fingerprint submissions (1 per audit batch, not per message)
```

**Digital Evidence Subscription Tiers:**

| ECHO User Tier | Digital Evidence Access | Cost |
|---------------|----------------------|------|
| Free | None (standard Merkle anchoring only) | Included in metagraph fees |
| VIP | Optional user-initiated media fingerprinting | Included in VIP subscription |
| Organization | Automatic audit fingerprinting + Smart Checkmark + compliance dashboard + data retention proof | Factored into Org tier pricing |

### 2.7 DeFi and Liquidity Infrastructure

ECHO token's utility depends on liquid markets. Constellation's DeFi infrastructure (PacaSwap DEX, cross-chain bridges, Stargazer wallet) provides this without ECHO building custom exchange infrastructure.

**PacaSwap DEX Integration:**

PacaSwap is Constellation's native DEX using an AMM model with the SWAP governance token. ECHO requires PacaSwap integration for:

| Liquidity Pool | Purpose | Phase |
|---------------|---------|-------|
| **ECHO/DAG** | Primary trading pair. Validators need DAG for L0 staking and ECHO for L1 staking. Treasury needs DAG for snapshot fees. Users swap between tokens. | Phase 2 (at ECHO token launch) |
| **ECHO/USDC** | Stablecoin on/off ramp. Treasury needs stablecoins for operational reserves. Users convert to/from stable value. | Phase 3 |

| Operation | Mechanism | Phase |
|-----------|-----------|-------|
| Liquidity bootstrapping | PacaSwap liquidity bootstrapping event at ECHO mainnet launch (price discovery + early liquidity) | Phase 2 |
| Treasury ECHO burns | AI Burn Agent buys ECHO from ECHO/DAG pool via atomic swap, then burns | Phase 5 |
| Cross-metagraph swaps | ECHO ↔ DOR, ECHO ↔ PACA via PacaSwap atomic swaps | Phase 3+ |
| In-app swap (optional) | iOS client can execute PacaSwap swaps without leaving ECHO app | Phase 4+ |
| LP dual rewards | ECHO liquidity providers on PacaSwap earn SWAP + ECHO dual incentives | Phase 3+ (governance-approved) |

**Cross-Chain Bridges:**

Constellation has live bridges to Base (Coinbase L2) and Ink (Kraken L2). ECHO should be bridgeable for broader DeFi access and exchange liquidity.

| Bridge | Purpose | Phase |
|--------|---------|-------|
| **ECHO ↔ Base** | Access Aerodrome DEX on Base. Broader DeFi (lending, yield). Treasury BTC accumulation path (ECHO → Base → CEX → BTC). | Phase 3 |
| **ECHO ↔ Ink** | Access Kraken exchange. Major liquidity and credibility milestone. DAG already listed on Kraken via Ink. | Phase 4 |

Bridge integration requires coordination with 3A DAO (Constellation's bridge partner) to add ECHO as a bridgeable L0 token.

**Stargazer Wallet:**

Stargazer is the official Constellation wallet supporting DAG, L0 tokens, delegation, and cross-chain bridging. ECHO token should be fully functional in Stargazer:

- Display ECHO balance and transaction history
- Stake and delegate ECHO to L1 validators via TokenLock + StakeDelegation
- Bridge ECHO to Base/Ink
- Execute PacaSwap swaps
- D'Cent hardware wallet support (D'Cent added DAG support March 2025)

Decision: whether to build token management into the ECHO iOS app or direct users to Stargazer for all token operations (see Open Questions #13).

**Treasury Multi-Chain Fund Flow (Phase 5):**

```
AI Treasury Agents — Multi-Chain Operations
├─ ECHO metagraph (Constellation Hypergraph)
│   ├─ Receive revenue in ECHO (VIP/Org fees converted on PacaSwap)
│   ├─ Execute ECHO buyback + burn via PacaSwap ECHO/DAG pool
│   ├─ Pay snapshot fees via FeeTransaction (DAG)
│   └─ Hold ECHO reserves for staking rewards emission
├─ Base (Coinbase L2) — via ECHO ↔ Base bridge
│   ├─ Swap ECHO → USDC on Aerodrome for operational reserves
│   ├─ Hold stablecoin reserves (USDC/DAI)
│   └─ Route to CEX for BTC accumulation
├─ Ink (Kraken L2) — via ECHO ↔ Ink bridge
│   └─ Route to Kraken for BTC purchases
└─ Bitcoin
    └─ Transfer to cold storage (multi-sig treasury wallet)
```

---

## 3. Data Flows

### 3.1 Message Send (End-to-End)

```
1. iOS App (Sender)
   ├─ Compose message
   ├─ Encrypt with Kinnami (X25519 key agreement + ChaCha20-Poly1305)
   ├─ Create commitment: H(H(plaintext) || nonce)
   ├─ Sign encrypted payload with Secure Enclave (P256)
   └─ Send via WebSocket to relay server

2. Go Backend — Message Relay Service
   ├─ Validate auth token + delivery token
   ├─ Rate limit check (per-DID send rate)
   ├─ [Phase 1–2] Server can see sender DID + recipient DID
   │   [Phase 3+] Sealed sender: server sees only recipient DID
   ├─ IF recipient online:
   │   └─ Forward encrypted blob via recipient's WebSocket
   ├─ IF recipient offline:
   │   ├─ Queue encrypted blob (Redis/PostgreSQL)
   │   └─ Send APNs push notification
   ├─ Add commitment to current Merkle batch
   └─ Log relay metadata to in-memory buffer (no content, no plaintext)

3. iOS App (Recipient) — on reconnect or push wake
   ├─ Receive encrypted blob via WebSocket
   ├─ Decrypt with own private key (Kinnami)
   ├─ Verify sender signature
   ├─ Verify commitment integrity
   └─ Display plaintext message

4. Batch Processing (every 5 minutes or 1000 commitments)
   ├─ Build Merkle tree from commitment batch
   ├─ Submit Merkle root to Data L1
   ├─ Encrypt log batch, push to IPFS
   └─ Record IPFS CID + Merkle root on Data L1

5. Metagraph Consensus
   ├─ Data L1 validates Merkle root submission
   ├─ Metagraph L0 packages into snapshot
   └─ Global L0 finalizes snapshot

6. Confirmation
   ├─ Backend receives finality callback
   └─ Push confirmation to sender via WebSocket
       ("Message anchored on-chain at snapshot #N")
```

### 3.2 Group Message Send

```
1. iOS App (Sender)
   ├─ Encrypt message once per recipient using recipient's public key
   │   (For large groups: encrypt with group symmetric key,
   │    distribute group key individually to each member)
   ├─ Create single commitment for the group message
   ├─ Sign payload
   └─ Send to relay server with group ID

2. Go Backend — Message Relay Service
   ├─ Validate sender is group member (cached group membership)
   ├─ Fan-out: forward encrypted blob to each online member's WebSocket
   ├─ Queue for each offline member
   ├─ Send APNs push to offline members
   └─ Add single commitment to Merkle batch

3. Group Key Management
   ├─ Group symmetric key generated on group creation
   ├─ Key distributed via E2E encrypted 1:1 messages to each member
   ├─ On member add: new key generated and distributed to all members
   ├─ On member remove: new key generated and distributed to remaining members
   └─ Key rotation on admin-configurable schedule
```

**Scaling for Large Groups (10K+ members):**

| Concern | Approach |
|---------|----------|
| Fan-out latency | Backend uses NATS pub/sub for parallel delivery across relay pods |
| Offline queue explosion | Large-group messages have 7-day offline retention (vs. 30 days for 1:1) |
| Group key distribution | Distributed via message relay using sender trees (admin → sub-admins → members) |
| On-chain metadata | Group ID + member count hash only; member list never on-chain |
| Rate limits | Group messages consume 1 send-rate token regardless of member count |

### 3.3 Reward Claim

```
1. iOS App → POST /tokens/rewards/claim (type, evidence)

2. Go Backend (Rewards Service)
   ├─ Validate claim against daily caps
   ├─ Apply trust tier multiplier (cached from Cardano)
   ├─ Pre-validate against anti-gaming rules
   └─ Add to reward batch queue

3. Batch Processing (every 30 seconds)
   ├─ Construct reward batch transaction
   └─ Submit to Currency L1

4. Currency L1
   ├─ Validate each reward (caps, eligibility, signature)
   ├─ Update token balances
   └─ Package into L1 block

5. Metagraph L0 → Global L0 → Finality

6. Confirmation
   ├─ Backend cache updated with new balance
   └─ Push balance update to iOS via WebSocket
```

### 3.4 Identity Verification

```
1. iOS App
   ├─ Capture ID document (camera)
   └─ Send directly to IDV provider (Stripe Identity / Sumsub)
       ├─ Platform backend NEVER sees raw images or PII
       └─ IDV provider processes, verifies, DELETES images

2. IDV Provider → Callback to Go Backend
   ├─ pass/fail, confidence score, document type, age_over_18 flag
   └─ Reference UUID only (no PII)

3. Go Backend (Identity Service)
   ├─ Map reference UUID to user DID
   ├─ Determine trust tier based on verification result
   └─ Submit credential issuance transaction to Cardano

4. Cardano
   ├─ Record verifiable credential metadata
   ├─ Set trust tier datum
   └─ Update credential status bit vector

5. Backend Cache
   ├─ Update cached trust tier
   └─ Notify metagraph Data L1 of new trust level (for future validation)
```

---

## 4. Cross-Chain Consistency Model

### 4.1 Consistency Strategy

The three chains operate under **eventual consistency** with the Go backend as the orchestrator. There is no distributed transaction across chains. The message relay layer operates independently of chain state — messages are delivered even if all chains are temporarily unavailable.

| Operation | Primary Chain | Secondary Chain | Failure Mode | Recovery |
|-----------|--------------|----------------|--------------|----------|
| Message relay | None (stateless transport) | Metagraph (commitment batch) | Metagraph failure → messages still delivered; commitments queue | Commitment retry with exponential backoff |
| Message anchoring | Metagraph (Data L1 Merkle root) | IPFS (log CID) | IPFS failure → retry queue; metagraph failure → batch queued | Retry; alert if >5 min |
| Reward claim | Metagraph (Currency L1) | None | Claim queued in backend; retry on L1 recovery | Idempotent submission (claim ID dedup) |
| Identity verification | Cardano (credential) | Metagraph (trust level cache) | Cardano failure → credential pending; metagraph gets stale trust | Backend retries Cardano; metagraph uses cached trust |
| Staking | Metagraph (Currency L1) | None | Same as reward claim | Idempotent |

**Key insight:** Because message relay is decoupled from on-chain operations, a metagraph or Cardano outage does not prevent users from messaging. It only delays on-chain anchoring and reward distribution. This is a major resilience advantage of the client-server relay model.

### 4.2 Conflict Resolution

- **Metagraph vs. Backend cache:** Metagraph snapshot state always wins. Backend cache is invalidated on every new snapshot event.
- **Cardano vs. Backend cache:** Cardano on-chain state always wins. Backend polls Cardano every 60 seconds for credential status changes. Critical operations (e.g., credential revocation) trigger an immediate poll.
- **Stale reads:** Clients may see data up to TTL-old (5s for balances, 60s for trust tiers). The iOS app displays "pending" states for recently submitted transactions until on-chain confirmation arrives via WebSocket.

### 4.3 Saga Pattern for Multi-Chain Operations

For operations that touch multiple chains (e.g., identity verification → Cardano credential + metagraph trust update):

```
Step 1: Submit to Cardano (primary)
Step 2: On Cardano confirmation → submit trust update to metagraph Data L1
Step 3: On metagraph confirmation → update backend cache → notify client

Compensating actions:
- If Step 2 fails: Retry up to 5 times. If still failing, mark trust
  update as "pending_sync" in backend. Cron job retries hourly.
- If Step 1 fails: No compensating action needed (nothing committed).
  Return error to client. Client can retry.
```

---

## 5. Fault Tolerance

### 5.1 Failure Scenarios

| Scenario | Impact on Messaging | Impact on Other Features | Mitigation |
|----------|-------------------|------------------------|------------|
| Metagraph L1 partition | **None** — messages relay normally | Rewards/anchoring queue | Backend queues submissions; drains on recovery |
| Metagraph L0 failure | **None** — messages relay normally | Snapshots halt; L1 blocks accumulate | L1 continues; L0 catches up |
| Global L0 unavailable | **None** — messages relay normally | Global finality delayed | Metagraph operates normally |
| Cardano congestion | **None** — messages relay normally | Credential operations slow | Backend uses cached credentials |
| IPFS/Storj outage | **None** — messages relay normally | Logs buffer locally | Flush on recovery |
| Go backend outage | **Message delivery stops** | All client operations blocked | Auto-scaling + multi-region; RTO < 60s |
| Redis failure | Offline queue degraded (falls back to PostgreSQL) | Cache miss → chain queries | Graceful degradation |
| PostgreSQL failure | Offline queue degraded | Operational data unavailable | Replicated (primary + 2 replicas); RTO < 30s |

**Key observation:** Because the relay is decoupled from blockchain operations, the only scenario that stops messaging is a Go backend outage. This is mitigated by stateless horizontal scaling, multi-region deployment, and (in Phase 4) federated relay nodes operated by independent parties.

### 5.2 Recovery Targets

| Layer | RTO | RPO |
|-------|-----|-----|
| Go backend / relay | < 60 seconds | 0 (stateless; offline queue in Redis/PG) |
| Redis cache + queue | < 30 seconds | < 1 second (AOF persistence) |
| PostgreSQL | < 30 seconds | < 1 second (synchronous replication) |
| Metagraph L1 | Network-dependent | 0 (consensus ensures no data loss) |
| Cardano | Network-dependent | 0 (blockchain finality) |
| IPFS/Storj | < 1 hour | < 5 minutes (buffered in backend) |

---

## 6. Client Verification (Roadmap)

For Phase 1–2, the iOS app trusts the Go backend's responses. For Phase 3+, the following verification mechanisms reduce trust in the relay:

| Verification | Method | Phase |
|-------------|--------|-------|
| Token balance | Backend returns balance + metagraph snapshot hash; app verifies against known snapshot | Phase 3 |
| Message anchoring | Backend returns Merkle proof; app verifies commitment inclusion against on-chain root | Phase 3 |
| Sender authenticity | Recipient verifies sender's P256 signature directly (no relay trust needed) | Phase 1 (already designed) |
| Message integrity | Recipient verifies commitment hash matches decrypted content | Phase 1 (already designed) |
| Credential status | App queries Cardano directly via light client (Mithril) for credential bit vector | Phase 4 |
| Reward confirmation | Backend returns L1 block inclusion proof | Phase 3 |

**Note:** Message content verification (sender signature + commitment hash) is already fully client-side in Phase 1. The relay server is already untrusted for message content — clients verify authenticity and integrity themselves. The Phase 3–4 items add verification for *on-chain state* (balances, anchoring, credentials) so that the relay is also untrusted for those.

---

## 7. Governance and Upgrade Path

| Aspect | Mechanism | Timeline |
|--------|-----------|----------|
| Data L1 schema changes | Governance proposal → vote (stake-weighted via StakeDelegation) → activation at future snapshot height | Phase 4 |
| L0 node admission | Requires 250K DAG staked per node; project-operated in all phases; additional L0 nodes can be added by operators meeting DAG collateral | Phase 2+ |
| L1 validator admission (Currency + Data) | Phase 1–3: project-operated only (permissioned). Phase 4: permissionless with minimum ECHO TokenLock requirement set by governance | Phase 4 |
| Validator slashing thresholds | Governance proposal → vote → activation. Initial thresholds set in Section 2.1; adjustable by supermajority vote. Slashed ECHO sent to treasury. | Phase 4 |
| Relay node admission | Initially project-operated; Phase 4: any staked operator (TokenLock) can run a relay node registered on Data L1 | Phase 4 |
| DAG delegation governance | ECHO can incentivize DAG delegators with ECHO token rewards; delegators reduce snapshot fees and strengthen network security | Phase 3+ |
| ECHO staking parameters | Staking tiers, APR rates, lock durations, minimum stakes — all governance-adjustable. Enforced via TokenLock validation in Currency L1 Scala code. | Phase 4 |
| Cardano contract upgrades | Parameterized Plutus scripts with upgrade key (multi-sig); transition to DAO-controlled | Phase 4 |
| Protocol versioning | Semantic versioning; clients must support current and current-1; L1 validators support current and current-1 schema versions | Ongoing |
| Emergency upgrades | Multi-sig (3-of-5 core team) can push critical fixes to L1 validation logic without governance vote | All phases |
| Metagraph token issuance changes | Require governance proposal + supermajority vote; emission schedule enforced in Currency L1 Scala validation code | Phase 4 |
| Tessellation version upgrades | Follow Constellation's upgrade path (v3 → v4+). Test on integrationnet, migrate L0/L1 nodes using nodectl, coordinate hardfork ordinal. | As released |
| Digital Evidence subscription tier | Governance decides Digital Evidence allocation for Org tier pricing; subscription costs funded from treasury operational budget | Phase 5 |

---

## 8. Encryption Specification (Canonical Reference)

All documents should reference this single table:

| Purpose | Algorithm | Key Type | Library |
|---------|-----------|----------|---------|
| Identity signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework |
| DID signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework |
| Message key agreement | X25519 ECDH | Ephemeral Curve25519 | CryptoKit |
| Message encryption | ChaCha20-Poly1305 | Derived symmetric (256-bit) | CryptoKit |
| Sealed sender envelope | AES-256-GCM | Derived from recipient identity key | CryptoKit |
| Local storage encryption | AES-256-GCM | Derived from master key via HKDF | CryptoKit |
| Key derivation | HKDF-SHA256 | From Secure Enclave signature | CryptoKit |
| Hash commitments | SHA-256 | N/A | CryptoKit / Go crypto |
| Password/PII hashing | Argon2id | Per-user salt | golang.org/x/crypto |
| Log encryption | AES-256-GCM | Monthly derived key (Shamir split) | Go crypto |
| Transport | TLS 1.3 | Certificate-based (with pinning) | URLSession / Go TLS |

---

## 9. Security & Decentralization Summary

| Principle | Implementation |
|-----------|----------------|
| No centralized database as authority | PostgreSQL/Redis are caches; metagraph and Cardano are sources of truth |
| Public verifiability | ECHO metagraph on public Hypergraph mainnet; token supply, rewards, and integrity commitments auditable by anyone via DAG Explorer |
| PRO consensus | Proof of Reputable Observation — DAG-based parallel transaction processing; validators earn reputation through honest behavior; slashing for dishonest validators |
| Content-blind relay | Relay servers transport E2E encrypted blobs; cannot read, modify, or forge messages |
| Client-verified authenticity | Recipients verify sender signatures and commitment hashes locally; no relay trust needed |
| Encrypted storage | Logs encrypted before IPFS/Storj; local data encrypted with derived keys; offline queue stores only ciphertext |
| Immutable audit trail | Merkle roots on Data L1; log CIDs on Data L1; Digital Evidence fingerprints for enterprise compliance; Cardano credential history; all anchored on public Hypergraph |
| Separation of concerns | Identity (Cardano), app state + rewards (metagraph on Hypergraph), audit (IPFS/Storj), evidence (Digital Evidence), exchange (PacaSwap), transport (relay) |
| Native token primitives | All token operations use Tessellation v3 types (TokenLock, StakeDelegation, AtomicAction) for Hypergraph-wide interoperability |
| Device-local secrets | Passkeys and private keys in iOS Secure Enclave; never extractable |
| Zero PII on-chain | Enforced by T0–T7 data classification; Data L1 validators reject prohibited types; public chain sees only hashes and token transactions |
| Forward secrecy | Ephemeral X25519 keys per message session |
| Anti-spam | Multi-layer: API rate limits, per-DID message rate, per-DID daily reward caps, economic micro-fees at scale |
| Graceful degradation | Message relay operates independently of chain state; blockchain outages don't stop messaging |
| Progressive metadata protection | Phase 1: baseline; Phase 3: sealed sender; Phase 4: federated relay + optional P2P |
| Progressive decentralization | Phase 1–3: permissioned L1 validators; Phase 4: permissionless with ECHO stake + slashing; L0 always public Hypergraph |
| Validator accountability | Phase 4: slashing for fraudulent validation, double-signing, extended downtime; slashed funds to treasury |

---

## 10. Open Questions Requiring Product Decisions

| # | Question | Impact | Needed By |
|---|----------|--------|-----------|
| 1 | Which IDV provider? (Stripe Identity, Sumsub, Jumio) | Cardano credential issuance flow | Phase 2 |
| 2 | ~~Metagraph deployment: Constellation mainnet or dedicated metagraph?~~ **RESOLVED v3.1: Public Hypergraph mainnet with permissioned L1 validators (hybrid model).** | — | — |
| 3 | ~~Staking: on metagraph Currency L1 or separate staking contract?~~ **RESOLVED v3.2: Use native Tessellation v3 TokenLock + StakeDelegation primitives on Currency L1. No separate contract.** | — | — |
| 4 | Group messaging: group symmetric key vs. per-recipient encryption cutoff? (e.g., per-recipient for <100 members, group key for 100+) | Security/performance tradeoff | Phase 2 |
| 5 | Multi-device sync: device-linked keys or cloud key backup? | Security model, Secure Enclave assumptions | Phase 3 |
| 6 | ZK proof system: Groth16 (snarkjs) or alternative (Halo2, PLONK)? | Circuit complexity, verification cost | Phase 3 |
| 7 | DAO governance: on metagraph or Cardano? | Smart contract platform choice | Phase 4 |
| 8 | Federated relay: minimum stake requirement for relay operators? | Relay decentralization economics | Phase 4 |
| 9 | Disappearing messages: how to handle on-chain Merkle roots after message expiry? (Root persists but individual commitments become unverifiable after key deletion) | Privacy/auditability tradeoff | Phase 3 |
| 10 | DAG acquisition strategy: treasury purchase, OTC deal, or Stardust Collective grant? | 750K DAG needed before Phase 2 mainnet deployment | Phase 1 |
| 11 | L1 validator ECHO stake requirement: what minimum ECHO TokenLock amount for community validators in Phase 4? | Decentralization vs. security tradeoff | Phase 3 (decide), Phase 4 (implement) |
| 12 | DAG delegation incentive structure: what % of ECHO rewards allocated to DAG delegators? | Snapshot fee economics, DAG ecosystem relationship | Phase 3 |
| 13 | ~~Stargazer wallet vs. custom in-app token management~~ **RESOLVED v3.3: Build native ECHO Wallet inside iOS app using Stargazer SDK. Users manage balance, staking (TokenLock), delegation (StakeDelegation), rewards (AtomicAction), swaps (PacaSwap), and bridges in-app. External compatibility with standalone Stargazer and D'Cent.** | — | — |
| 14 | Digital Evidence subscription tier: which Constellation Digital Evidence tier for Organization plan? Free tier for testing, paid/enterprise for production? | Org tier pricing, compliance feature set | Phase 3 (evaluate), Phase 5 (deploy) |
| 15 | Tessellation v4 migration timeline: when v4 reaches stable release, what is the upgrade window for ECHO metagraph? Can ECHO remain on v3 temporarily? | Compatibility, potential downtime | When v4 released |
| 16 | PacaSwap liquidity bootstrapping: how much ECHO + DAG to seed initial ECHO/DAG pool? Treasury allocation or community fundraise? | Token launch liquidity, price discovery | Phase 2 |
| 17 | Cross-chain bridge priority: Base first or Ink first? Base has Aerodrome DeFi; Ink has Kraken exchange access. | Treasury operations path, exchange strategy | Phase 3 |

---

*Data Layer Architecture v3.3*
*Updated: March 7, 2026*
*Status: Constellation ecosystem fully integrated. Tokenomics finalized (1B supply, founder vesting via TokenLock). Wallet architecture resolved (Stargazer SDK). Ready for Phase 1 implementation.*
