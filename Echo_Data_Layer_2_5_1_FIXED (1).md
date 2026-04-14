# Data Layer Architecture

## Overview

The data layer is built on a multi-blockchain architecture combining Constellation metagraph infrastructure for application data and ECHO rewards, Cardano for identity and verifiable credentials, and decentralized storage for logging and audit trails. A Go backend microservices layer acts as an operational coordinator, stateless message relay, and hot cache between client applications and on-chain state.

**Design Principles:**

* On-chain as source of truth; off-chain as performance optimization
* Zero PII on any blockchain (enforced by T0–T7 data classification)
* Each chain handles one concern: identity (Cardano), application state + rewards (metagraph on public Hypergraph), audit (IPFS/Storj)
* The Go backend is a relay and cache, not an authority—it cannot read message content, does not own user identities, and does not control token balances
* Message relay is client-server for reliability; decentralization comes from identity, data integrity, and encryption layers
* Public Hypergraph mainnet for token and data integrity verifiability; permissioned L1 validators for controlled business logic (progressive decentralization to permissionless in Phase 4)
* Use native Tessellation v3 transaction primitives (TokenLock, StakeDelegation, AtomicAction, AllowSpend, WithdrawLock, FeeTransaction) rather than custom implementations for Hypergraph interoperability
* Metagraph validation logic in Scala (Euclid SDK); Go backend and iOS app interact via REST API
* Proof of Reputable Observation (PRO) consensus—Constellation's DAG-based consensus model enables parallel transaction processing, near-zero fees for end users, and real-time data validation

## Core Components

### Constellation Metagraph (Application & Rewards Layer)

The primary data layer uses Constellation Network's metagraph architecture for high-throughput, decentralized consensus on application data and token transactions.

**Deployment Model: Public Hypergraph Mainnet with Permissioned L1 (Hybrid)**

ECHO deploys as a public metagraph on Constellation's Hypergraph mainnet. L0 nodes submit snapshots to the public Global L0 for immutable recording. L1 validators are permissioned (project-operated) initially, transitioning to community-operated in Phase 4.

This is not a private metagraph. Rationale:

* **Token credibility requires public verifiability.** ECHO token supply, distribution, and reward claims must be publicly auditable. A private chain would mean "trust us"—the exact problem ECHO solves.
* **Ecosystem network effects.** Public metagraph = ECHO token visible in Stargazer wallet, tradeable on PacaSwap DEX, eligible for DAG delegation, interoperable with other Hypergraph metagraphs.
* **Privacy is application-layer, not chain-layer.** The public Hypergraph sees only Merkle roots (opaque hashes), trust commitments, and token transactions. No PII, no message content ever reaches any chain.

**Node Infrastructure:**

| Node Type | Minimum Count | DAG Requirement | Operator | Role |
| --- | --- | --- | --- | --- |
| L0 Hybrid Nodes | 3 | 250K DAG staked per node (750K total) | Project-operated (all phases) | Run Global L0 + Metagraph L0 processes; submit snapshots to Hypergraph |
| Currency L1 Validators | 3 (launch), 5+ (scale) | ECHO token stake (governance-set) | Project-operated (Phase 1–3); community (Phase 4) | Validate ECHO token transfers, rewards, staking |
| Data L1 Validators | 3 (launch), 5+ (scale) | ECHO token stake (governance-set) | Project-operated (Phase 1–3); community (Phase 4) | Validate Merkle roots, trust commitments, governance votes |

**Server Requirements (per node):** Ubuntu 22.04, 8+ CPU cores, 32GB+ RAM, NVMe SSD storage, stable network. Recommended: Hetzner AX41-NVMe dedicated server (~€45/month) or Hetzner Cloud CCX33 (~€55/month). Bare metal preferred for L0 nodes (consistent consensus performance without virtualization overhead).

**Snapshot Fee Economics:**

Metagraphs pay snapshot fees in DAG to the Hypergraph for each snapshot submitted by L0 nodes. End users pay zero fees—ECHO as a project absorbs snapshot costs. Snapshot cap: 50KB per snapshot. ECHO snapshot frequency: \~1 per 5 seconds. Cost reduction via delegation: more DAG delegated to ECHO's validators = lower net snapshot fees.

**Metagraph Structure:**

| Layer | Role | Scaling Model |
| --- | --- | --- |
| **Currency L1** | Validates ECHO reward token transactions, manages balances, processes staking state changes | Horizontal (add validator nodes) |
| **Data L1** | Validates domain-specific application data with custom business logic | Horizontal (add validator nodes) |
| **Metagraph L0** | Aggregates validated L1 blocks into metagraph snapshots (finalized state) | Vertical (more powerful nodes) |
| **Global L0 (Hypergraph)** | Final consensus, immutable recording of all metagraph snapshots | Vertical |

**Metagraph L1 Validation Logic (Scala/JVM):**

All custom validation logic for ECHO's Data L1 and Currency L1 must be written in Scala using the Euclid SDK (built on Constellation's Tessellation framework). This enforces ECHO-specific business rules on-chain:

| Validation Rule | L1 Layer | Logic |
| --- | --- | --- |
| Annual emission enforcement | Currency L1 | Reject reward claims that would cause Year-N total distributions to exceed the Year-N emission cap. Per-message rate auto-scales based on total daily network activity weight. No per-user daily cap. |
| Trust-tier multiplier | Currency L1 | Apply correct multiplier based on cached trust tier; reject mismatched multipliers |
| Anti-gaming | Currency L1 | Detect and reject suspicious reward patterns (velocity checks, repeat claims) |
| Merkle root structure | Data L1 | Validate submitted Merkle roots have correct structure, authorized sender DID |
| Trust commitment format | Data L1 | Validate H(score |
| Governance vote rules | Data L1 | One-vote-per-DID, active proposal check, minimum stake requirement |
| Schema version check | Both | Reject submissions with unsupported schema versions |

The Go backend and iOS app interact with the metagraph through its REST API. Only the on-chain validation layer requires Scala.

**Data L1 Validated Data Types:**

| Data Type | On-Chain Content | Privacy | Validation Rules |
| --- | --- | --- | --- |
| Message integrity | Merkle root of batch commitments (never content) | T3 compliant | Valid Merkle structure, authorized sender DID |
| Trust commitments | H(trust_score | nonce) |  |
| Reward claims | Claim type, DID, amount, trust tier | T7 (public) | Annual budget enforcement, auto-scale rate validation, trust multiplier check, anti-gaming rules |
| Governance votes | Proposal ID, DID, vote, stake weight | T7 (public) | Active proposal, sufficient stake, one-vote-per-DID |
| Staking operations | DID, amount, tier, lock duration | T7 (public) | Minimum amounts per tier, valid lock period |
| Group metadata | Group ID, member count hash, admin DID | T7 (public) | Valid admin signature, member count bounds |
| Relay node registry | Node DID, endpoint, stake, uptime | T7 (public) | Minimum stake, valid endpoint (Phase 4) |

**Tessellation v3 Transaction Type Mapping:**

ECHO uses native Tessellation v3 transaction primitives for all token operations. Custom implementations are used only for ECHO-specific business logic that v3 primitives don't cover. Using native types ensures interoperability with Stargazer wallet, DAG Explorer, PacaSwap DEX, and cross-chain bridges.

| v3 Primitive | ECHO Operation | L1 Layer | Notes |
| --- | --- | --- | --- |
| **TokenLock** | ECHO staking (lock ECHO for 5–15% APY) | Currency L1 | Tokens locked but remain in user's wallet. Lock duration enforced by L1 validation. |
| **StakeDelegation** | Delegate locked ECHO to an L1 validator to earn rewards | Currency L1 | Users choose a validator. Delegation increases validator's consensus weight. |
| **WithdrawLock** | Initiate ECHO unstaking with 14-day cooldown | Currency L1 | 14-day cooldown (governance-adjustable). |
| **AtomicAction** | Bundle reward claim + trust tier verification + daily cap update as single all-or-nothing transaction | Both L1s | Eliminates race conditions in reward claims. |
| **AllowSpend** | Time-limited approval for bot/marketplace payments (Phase 5) | Currency L1 | Explicitly time-limited—avoids unlimited approval vulnerability. |
| **SpendTransaction** | Execute payment against an AllowSpend approval | Currency L1 | Paired with AllowSpend for Phase 5 payment rails. |
| **FeeTransaction** | Automated snapshot fee payment from ECHO treasury DAG reserves | Currency L1 | Eliminates manual snapshot fee management. |

**Performance Targets:**

| Metric | Launch (100K users) | Scale (1M users) |
| --- | --- | --- |
| Data L1 TPS | 50 | 500 |
| Currency L1 TPS | 100 | 1,000 |
| End-to-end finality | < 10 seconds | < 15 seconds |
| Metagraph snapshot interval | 5 seconds | 5 seconds |
| L1 validator nodes | 5 | 20+ |
| L0 nodes | 3 | 5 |

**Phased Deployment:**

| Phase | Metagraph State | Details |
| --- | --- | --- |
| Phase 1 | **Testnet** | Euclid SDK local development + Constellation testnet. Build and test all L1 validation logic. Acquire 750K+ DAG. |
| Phase 2 | **Mainnet — Permissioned L1** | Deploy 3 L0 hybrid nodes on Hypergraph mainnet (750K DAG staked). ECHO token live. Project operates all L1 validators. |
| Phase 3 | **Mainnet — DAG Delegation** | Launch delegation campaign: attract DAG holders to delegate to ECHO validators for lower snapshot fees. |
| Phase 4 | **Mainnet — Permissionless L1 + Federated Relay** | Any operator meeting minimum ECHO stake can run L1 validators. L0 nodes still require 250K DAG. DAO governance. Project relay nodes on Hetzner (DE) + OVHcloud (FR). Community relay operators encouraged to use Akash Network, Flux, or bare metal for maximum decentralization. No single provider > 60% of traffic. |

### Cardano Identity Layer

Cardano serves as the verifiable credential registry, fully separated from application logic and rewards.

**Network:** Cardano Mainnet (testnet during development phases 1–2)

**Credential Standard:** W3C Verifiable Credentials (VCs) using `did:prism` (Atala PRISM / Veridian) DID method

**On-Chain Data Structures:**

| Structure | Storage Method | Content |
| --- | --- | --- |
| DID Document | Cardano metadata (CIP-25 extended) | Public keys, service endpoints, controller |
| Credential Schema | Plutus reference script | Schema definition, version, issuer DID |
| Credential Status | Bit vector in UTXO datum | Revocation status per credential index |
| Trust Tier | UTXO datum | Tier level (1–5), issuer DID, expiry |
| Verification Record | Transaction metadata | Verifier DID, method used, timestamp |

**Trust Levels:**

| Tier | Method | On-Chain Record | Feature Access |
| --- | --- | --- | --- |
| 1 — Unverified | Self-registration | DID Document only | Basic messaging, no rewards |
| 2 — Newcomer | Email/phone verification | \+ Verification record | Messaging + basic rewards |
| 3 — Member | Third-party IDV (Stripe Identity / Sumsub) | \+ Credential status bit | Full rewards, group creation |
| 4 — Verified | Government ID (Apple Digital ID or IDV provider) | \+ Trust tier datum | Enhanced rewards multiplier, payment rails |
| 5 — Trusted | Peer attestations + sustained activity | \+ Web-of-trust attestation chain | Maximum multiplier, governance participation |

**Cost Model:**

* Credential issuance: \~0.3–0.5 ADA per transaction (paid from platform treasury)
* Revocation update: \~0.2 ADA (batch multiple revocations per transaction)
* Estimated monthly cost at 100K users with 30% verified: \~15,000 ADA
* Fee delegation: Platform submits transactions on behalf of users; users never need to hold ADA

**Revocation Mechanism:**

Bit vector stored in a Plutus UTXO datum. Each credential gets an index; setting the bit to 1 revokes. Verifiers check the bit vector before accepting a credential. Batch revocation: multiple credentials revoked in a single transaction.


### Midnight Privacy Verification Layer (Phase 3+)

Midnight is Cardano's privacy-focused partner chain using ZK-SNARKs for selective disclosure. It provides a production-grade ZK verification environment — particularly valuable for Organization-tier enterprise clients who need compliance verification without public data exposure. ECHO evaluates Midnight in Phase 3 after it has proven mainnet stability, and integrates it in Phase 4.

**Decision: Cardano for identity (Phases 1–2). Add Midnight for ZK credential verification (Phase 3+).**

**What stays on Cardano (always):**
- DID Document registration (public by design — contains public keys)
- Credential schema definitions (Plutus reference scripts)
- Credential issuance and revocation (bit vector in UTXO datum)
- Trust tier UTXO datums (backward compatible, works today)

**What moves to Midnight (Phase 3+):**

| Use Case | ZK Proof | Benefit | Phase |
| --- | --- | --- | --- |
| Trust tier verification | "Prove I am Tier 3+ without revealing my score or credential" | Eliminates hash-commitment workaround; native ZK verification | Phase 3-4 |
| KYC compliance proof | "Prove I passed KYC without revealing my passport data" | Organization tier: compliance without data exposure | Phase 4 |
| Private group membership | "Prove I am a member of Group X without revealing my groups" | Privacy for sensitive group affiliations | Phase 4 |
| Age/eligibility | "Prove I am 18+ without revealing my birthdate" | Minimal disclosure for age-gated features | Phase 4 |
| Balance threshold | "Prove I hold enough ECHO for staking without revealing exact balance" | Financial privacy for governance and feature access | Phase 4 |

**Architecture:**

```
ECHO Identity Stack (Phase 3+)
├── Cardano (Source of Truth)
│   ├── DID Documents (public)
│   ├── Credential issuance (Plutus)
│   ├── Revocation registry (bit vector)
│   └── Trust tier datums
│
├── Midnight (Privacy Verification Layer)
│   ├── Compact contracts (TypeScript-based DSL)
│   ├── ZK trust tier verifier (reads Cardano state via bridge)
│   ├── ZK KYC compliance prover
│   ├── ZK group membership prover
│   └── ZK balance threshold prover
│
├── Cardano ↔ Midnight Bridge (native, built by IOG)
│   └── Asset transfers + state queries between chains
│
└── Go Backend
    ├── Queries Cardano for credential issuance/revocation (existing)
    └── Queries Midnight for ZK verification results (new Phase 3+ service)
```

**Technical Notes:**
- Midnight uses Compact (TypeScript-based) — no Scala required. A web developer can write Midnight contracts.
- Midnight has a dual-token model: NIGHT (governance/staking) and DUST (renewable, non-tradable, pays for ZK computations). ECHO does not need to hold large NIGHT positions — ZK verification calls consume DUST which is generated from minimal NIGHT holdings.
- ZK proofs are generated locally on the user's device (iOS) and submitted to Midnight for on-chain verification. Private data never leaves the device.
- The native Cardano ↔ Midnight bridge allows Midnight contracts to read Cardano credential state for verification without duplicating data.
- Target proof generation time: under 5 seconds on modern iPhone hardware.

### Message Relay Layer

The message relay layer transports E2E encrypted messages between clients. It is a stateless, content-blind relay—it handles ciphertext blobs it cannot read, decrypt, or modify.

**Architecture Decision: Client-Server Relay (not P2P)**

ECHO uses a client-server WebSocket relay model. iOS platform constraints (aggressive background process killing), offline delivery requirements, group fan-out at scale (1M members), and push notification requirements make pure P2P unviable for a consumer iOS messaging product. ECHO's decentralization value comes from identity (Cardano DIDs), data integrity (metagraph consensus), and content privacy (E2E encryption)—not from the transport layer.

**Relay Server Capabilities and Limitations:**

| Relay Server CAN | Relay Server CANNOT |
| --- | --- |
| Transport encrypted blobs between clients | Read, decrypt, or modify message content |
| Queue encrypted messages for offline recipients | Forge messages (clients verify sender signatures) |
| Deliver push notifications via APNs | Access private keys or identity credentials |
| Track delivery status (sent, delivered, read receipt) | Override metagraph or Cardano state |
| Rate-limit abusive senders | Associate encrypted content with real-world identity |
| See sender DID, recipient DID, timestamp, blob size | See message plaintext, attachments, or reactions |

**Offline Message Queuing:**

When a recipient is offline, the relay server holds encrypted messages in a temporary queue:

| Property | Value |
| --- | --- |
| Storage | Redis (encrypted blobs in memory) or PostgreSQL (encrypted at rest) |
| Retention | Maximum 30 days for 1:1 chats; 7 days for large groups (100+ members) |
| Encryption | Messages are already E2E encrypted by sender; server stores opaque blobs |
| Delivery | On recipient reconnect (WebSocket), server drains queue in order |
| Size limit | 1000 queued messages per recipient; overflow backed up to IPFS (see below) |
| Overflow backup | When queue exceeds 1000 messages, overflow E2E encrypted blobs are pinned to IPFS/Storj. Relay stores only the CID in queue metadata. On reconnect, relay provides CIDs for the recipient to retrieve overflow messages directly from IPFS. Content-blind model preserved — backup is the same opaque encrypted blob. |
| Push notification | APNs notification sent immediately on queue insertion |

**Metadata Protection Roadmap:**

| Phase | Protection Level | Method | Server Sees |
| --- | --- | --- | --- |
| 1–2 | Baseline | TLS 1.3 transport; auth token per session | Sender DID, recipient DID, timestamp, blob size |
| 3 | Sealed sender | Sender identity encrypted inside E2E envelope; server routes by recipient only | Recipient DID, timestamp, blob size (sender hidden) |
| 4 | Federated relay | Traffic split across independent operators; no single operator sees all traffic | Each operator sees only its routed fraction |
| 4+ | Optional P2P | When both parties are online, establish direct WebSocket via relay-assisted signaling | Relay sees connection setup only |

### Go Backend Operational Layer

The Go backend is **not** a centralized authority—it is an operational coordinator and hot cache that sits between clients and on-chain state. It also serves as the message relay infrastructure (Phase 1–3), which transitions to federated relay operators in Phase 4.

**Role Clarification:**

| Function | Authoritative Source | Backend Role |
| --- | --- | --- |
| Token balances | Currency L1 (metagraph) | Read-through cache (TTL: 5s) |
| Trust scores | Cardano + metagraph commitment | Compute engine + cache (TTL: 60s) |
| Message content | Device-local (E2E encrypted) | Relay only (queues ciphertext for offline) |
| Message metadata | Data L1 (Merkle root) | Batch aggregator before submission |
| User identity | Cardano DID | Cache + credential proof validator |
| Reward eligibility | Data L1 validators | Pre-validator (reject obviously invalid claims) |
| Message relay | N/A (stateless transport) | WebSocket relay + APNs push + offline queue |

**Circuit Breakers:** Each downstream connection (metagraph, Cardano, IPFS/Storj) has an independent circuit breaker. Message relay continues even if all chains are unavailable—messages are transported as encrypted blobs regardless of on-chain status. On-chain operations (rewards, commitments, credential checks) degrade gracefully to cached state.


### Community Relay Node Economics (Phase 4)

Federated relay infrastructure requires economic incentives for community operators. The relay node program transitions message routing from project-operated (Phase 1–3) to community-operated (Phase 4+).

| Parameter | Value |
| --- | --- |
| Minimum ECHO stake | Governance-set (suggested: 50,000 ECHO via TokenLock) |
| Minimum DAG stake | Not required (DAG staking is for L0 nodes only) |
| Revenue model | Relay operators earn a share of snapshot fee rebates proportional to uptime and traffic served |
| Slashing conditions | Downtime > 1 hour in 24h period → warning; > 4 hours → 1% stake slashed; repeated violations → ejection from registry |
| Registration | Submit relay node DID + endpoint + cloud provider to Data L1 registry |
| Discovery | Clients query Data L1 for active relay nodes; rotate across 3 nodes per session |
| Load balancing | Client-side rotation with preference for low-latency, high-uptime nodes |
| Cloud diversity scoring | Relay registry tracks cloud provider per node; governance can set minimum diversity thresholds |
| Minimum nodes | 5 community-operated relay nodes before federated mode activates |

**Relay Node Data L1 Registry Entry:**

| Field | Content | Privacy |
| --- | --- | --- |
| Node DID | Operator's decentralized identifier | T7 (public) |
| Endpoint URL | WebSocket relay endpoint | T7 (public) |
| ECHO Stake | TokenLock amount and duration | T7 (public) |
| Cloud Provider | Hetzner, OVHcloud, Akash, Flux, bare metal, etc. | T7 (public) |
| Uptime (30d rolling) | Percentage based on heartbeat checks | T7 (public) |
| Traffic Served (30d) | Encrypted blob count (no content metadata) | T7 (public) |
| Registration Date | Timestamp of initial registration | T7 (public) |

### Decentralized Logging & Storage

**Storage Provider:** IPFS (with Pinata / web3.storage pinning) for immutable logs. Storj as fallback for large media audit trails.

**Log Lifecycle:**

| Phase | Action | Details |
| --- | --- | --- |
| **Collection** | Go backend batches API events, relay metadata, and metagraph transaction receipts | In-memory buffer, max 1000 events or 5 minutes |
| **Encryption** | Batch encrypted with AES-256-GCM using a rotating log encryption key | Key derived from platform master key via HKDF with date-based info string |
| **Submission** | Encrypted batch pushed to IPFS; CID recorded | Retry with exponential backoff on failure |
| **Pinning** | CID pinned via Pinata/web3.storage (primary) + self-hosted IPFS node (secondary) | Minimum 2 pin providers for redundancy |
| **Indexing** | CID + time range + batch hash submitted to Data L1 | Enables on-chain verifiable log index |
| **Retrieval** | Authorized auditors decrypt with log key (threshold scheme: 3-of-5 key holders) | Access logged on-chain as audit access event |
| **Retention** | Minimum 7 years for compliance; pins maintained by platform treasury | Unpinning only after retention period + governance vote |

**What relay metadata is logged (privacy-safe):**

* Message count per time window (no content, no DIDs unless required for compliance)
* Delivery success/failure rates
* Queue depth statistics
* Rate limit trigger events
* Circuit breaker state changes

**Key Management for Logs:**

* Log encryption keys are derived monthly from a platform master key
* Master key is split using Shamir's Secret Sharing (3-of-5 threshold)
* Key holders are designated platform operators (expandable to DAO members at Phase 4)

**Compression and Cost:**

* Batches compressed with zstd before encryption
* Estimated cost at 100K users: \~$50/month (IPFS pinning) + \~$20/month (Storj overflow)
* Batch size target: 1–5 MB compressed per batch

### DeFi and Liquidity Infrastructure

ECHO token's utility depends on liquid markets. Constellation's DeFi infrastructure provides this without ECHO building custom exchange infrastructure.

**PacaSwap DEX Integration:**

| Liquidity Pool | Purpose | Phase |
| --- | --- | --- |
| **ECHO/DAG** | Primary trading pair. Validators need DAG for L0 staking and ECHO for L1 staking. Treasury needs DAG for snapshot fees. | Phase 2 (at ECHO token launch) |
| **ECHO/USDC** | Stablecoin on/off ramp. Treasury needs stablecoins for operational reserves. | Phase 3 |

**Cross-Chain Bridges:**

Constellation has live bridges to Base (Coinbase L2) and Ink (Kraken L2). ECHO should be bridgeable for broader DeFi access and exchange liquidity.

| Bridge | Purpose | Phase |
| --- | --- | --- |
| **ECHO ↔ Base** | Access Aerodrome DEX on Base. Broader DeFi (lending, yield). Treasury BTC accumulation path. | Phase 3 |
| **ECHO ↔ Ink** | Access Kraken exchange. Major liquidity and credibility milestone. | Phase 4 |

**Stargazer Wallet:**

Stargazer is the official Constellation wallet supporting DAG, L0 tokens, delegation, and cross-chain bridging. ECHO token should be fully functional in Stargazer:

* Display ECHO balance and transaction history
* Stake and delegate ECHO to L1 validators via TokenLock + StakeDelegation
* Bridge ECHO to Base/Ink
* Execute PacaSwap swaps
* D'Cent hardware wallet support

## Data Flows

### Message Send (End-to-End)

```
1. iOS App (Sender)
   ├─ Compose message
   ├─ Encrypt with X25519 key agreement + ChaCha20-Poly1305
   ├─ Create commitment: H(H(plaintext) || nonce)
   ├─ Sign encrypted payload with Secure Enclave (P-256)
   └─ Send via WebSocket to relay server

2. Go Backend — Message Relay Service
   ├─ Validate auth token + delivery token
   ├─ Rate limit check (per-DID send rate)
   ├─ IF recipient online:
   │   └─ Forward encrypted blob via recipient's WebSocket
   ├─ IF recipient offline:
   │   ├─ Queue encrypted blob (Redis/PostgreSQL)
   │   └─ Send APNs push notification
   ├─ Add commitment to current Merkle batch
   └─ Log relay metadata (no content, no plaintext)

3. iOS App (Recipient)
   ├─ Receive encrypted blob via WebSocket
   ├─ Decrypt with own private key (X25519)
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
```


### Contact Discovery

```
1. iOS App
   ├─ User opts in to contact matching (Settings → Privacy)
   ├─ Hash each contact phone number with Argon2id + per-user salt (on-device)
   │   └─ Salt is generated once, stored in Secure Enclave, never transmitted
   └─ Send array of hashed entries to Contacts Service

2. Go Backend (Contacts Service)
   ├─ Match incoming hashed entries against server-side discovery index
   │   └─ Index stores: Argon2id salted hashes → encrypted DID references
   ├─ Return encrypted DID references for matches
   ├─ Rate limit: 1 discovery request per 24 hours per DID
   └─ Server never sees raw phone numbers; index not on any blockchain

3. iOS App
   ├─ Decrypt DID references
   ├─ Display matched contacts with trust tier badges
   └─ Offer to send connection request

Alternative Discovery Paths (no server involvement):
- QR Code: User displays DID QR code → other user scans → mutual connection
- Username: GET /contacts/search?handle={username} → DID + badge
- Invite Link: POST /contacts/invite → unique link with referral tracking
```

### Reward Claim

```
1. iOS App → POST /tokens/rewards/claim (type, evidence)

2. Go Backend (Rewards Service)
   ├─ Validate claim against annual emission budget and auto-scaled network rate
   ├─ Apply trust tier reward multiplier: Tier 1 (1.0x), Tier 2 (1.2x), Tier 3 (1.5x), Tier 4 (2.0x), Tier 5 (3.0x)
   ├─ Pre-validate against anti-gaming rules (velocity checks, repeat claims)
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

### Identity Verification

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
   └─ Notify metagraph Data L1 of new trust level
```

## Cross-Chain Consistency Model

The three chains operate under **eventual consistency** with the Go backend as the orchestrator. There is no distributed transaction across chains. The message relay layer operates independently of chain state—messages are delivered even if all chains are temporarily unavailable.

| Operation | Primary Chain | Secondary Chain | Failure Mode | Recovery |
| --- | --- | --- | --- | --- |
| Message relay | None (stateless transport) | Metagraph (commitment batch) | Metagraph failure → messages still delivered; commitments queue | Commitment retry with exponential backoff |
| Message anchoring | Metagraph (Data L1 Merkle root) | IPFS (log CID) | IPFS failure → retry queue; metagraph failure → batch queued | Retry; alert if &gt;5 min |
| Reward claim | Metagraph (Currency L1) | None | Claim queued in backend; retry on L1 recovery | Idempotent submission (claim ID dedup) |
| Identity verification | Cardano (credential) | Metagraph (trust level cache) | Cardano failure → credential pending; metagraph gets stale trust | Backend retries Cardano |
| ZK verification (Phase 3+) | Midnight (ZK proof verification) | Cardano (credential source) | Midnight failure → fall back to hash-commitment verification | Backend retries; graceful degradation |
| Staking | Metagraph (Currency L1) | None | Same as reward claim | Idempotent |

**Key insight:** Because message relay is decoupled from on-chain operations, a metagraph or Cardano outage does not prevent users from messaging. It only delays on-chain anchoring and reward distribution.

## Fault Tolerance

### Failure Scenarios

| Scenario | Impact on Messaging | Impact on Other Features | Mitigation |
| --- | --- | --- | --- |
| Metagraph L1 partition | **None** — messages relay normally | Rewards/anchoring queue | Backend queues submissions; drains on recovery |
| Metagraph L0 failure | **None** — messages relay normally | Snapshots halt; L1 blocks accumulate | L1 continues; L0 catches up |
| Global L0 unavailable | **None** — messages relay normally | Global finality delayed | Metagraph operates normally |
| Cardano congestion | **None** — messages relay normally | Credential operations slow | Backend uses cached credentials |
| Midnight unavailable (Phase 3+) | **None** — messages relay normally | ZK verification falls back to hash-commitment scheme | Graceful degradation; feature still works with lower assurance |
| IPFS/Storj outage | **None** — messages relay normally | Logs buffer locally | Flush on recovery |
| Go backend outage | **Message delivery stops** | All client operations blocked | Auto-scaling + multi-region; RTO < 60s |
| Redis failure | Offline queue degraded (falls back to PostgreSQL) | Cache miss → chain queries | Graceful degradation |
| PostgreSQL failure | Offline queue degraded | Operational data unavailable | Replicated (primary + 2 replicas); RTO < 30s |

### Recovery Targets

| Layer | RTO | RPO |
| --- | --- | --- |
| Go backend / relay | < 60 seconds | 0 (stateless; offline queue in Redis/PG) |
| Redis cache + queue | < 30 seconds | < 1 second (AOF persistence) |
| PostgreSQL | < 30 seconds | < 1 second (synchronous replication) |
| Metagraph L1 | Network-dependent | 0 (consensus ensures no data loss) |
| Cardano | Network-dependent | 0 (blockchain finality) |
| IPFS/Storj | < 1 hour | < 5 minutes (buffered in backend) |

## Security & Decentralization Summary

| Principle | Implementation |
| --- | --- |
| No centralized database as authority | PostgreSQL/Redis are caches; metagraph and Cardano are sources of truth |
| Public verifiability | ECHO metagraph on public Hypergraph mainnet; token supply, rewards, and integrity commitments auditable by anyone |
| PRO consensus | Proof of Reputable Observation — DAG-based parallel transaction processing; validators earn reputation through honest behavior |
| Content-blind relay | Relay servers transport E2E encrypted blobs; cannot read, modify, or forge messages |
| Client-verified authenticity | Recipients verify sender signatures and commitment hashes locally; no relay trust needed |
| Encrypted storage | Logs encrypted before IPFS/Storj; local data encrypted with derived keys; offline queue stores only ciphertext |
| Immutable audit trail | Merkle roots on Data L1; log CIDs on Data L1; Cardano credential history; all anchored on public Hypergraph |
| Separation of concerns | Identity (Cardano), app state + rewards (metagraph on Hypergraph), audit (IPFS/Storj), transport (relay) |
| Native token primitives | All token operations use Tessellation v3 types (TokenLock, StakeDelegation, AtomicAction) for Hypergraph-wide interoperability |
| Device-local secrets | Passkeys and private keys in iOS Secure Enclave; never extractable |
| Zero PII on-chain | Enforced by T0–T7 data classification; Data L1 validators reject prohibited types; public chain sees only hashes and token transactions |
| Forward secrecy | Ephemeral X25519 keys per message session |
| Anti-spam | Multi-layer: API rate limits, per-DID message rate, annual emission budget with auto-scaled per-message rate (no per-user daily cap), economic micro-fees at scale |
| Graceful degradation | Message relay operates independently of chain state; blockchain outages don't stop messaging |
| Progressive metadata protection | Phase 1: baseline; Phase 3: sealed sender; Phase 4: federated relay + optional P2P |
| ZK privacy layer | Phase 3: evaluate Midnight; Phase 4: integrate for trust tier, KYC, group membership, and balance threshold proofs |
| Progressive decentralization | Phase 1–3: permissioned L1 validators; Phase 4: permissionless with ECHO stake; L0 always public Hypergraph |

## Encryption Specification

All documents should reference this single canonical table:

| Purpose | Algorithm | Key Type | Library |
| --- | --- | --- | --- |
| Identity signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework |
| DID signing | ECDSA P-256 | Secure Enclave hardware key | Security.framework |
| Message key agreement | X25519 ECDH | Ephemeral Curve25519 | CryptoKit |
| Message encryption | ChaCha20-Poly1305 | Derived symmetric (256-bit) | CryptoKit |
| Sealed sender envelope | AES-256-GCM | Derived from recipient identity key | CryptoKit |
| Local storage encryption | AES-256-GCM | Derived from master key via HKDF | CryptoKit |
| Key derivation | HKDF-SHA256 | From Secure Enclave signature | CryptoKit |
| Hash commitments | SHA-256 | N/A | CryptoKit / Go crypto |
| Password/PII hashing | Argon2id | Per-user salt | [golang.org/x/crypto]() |
| Log encryption | AES-256-GCM | Monthly derived key (Shamir split) | Go crypto |
| Transport | TLS 1.3 | Certificate-based (with pinning) | URLSession / Go TLS |