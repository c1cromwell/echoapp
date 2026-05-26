# Phase 4: Blockchain & Trust Infrastructure

**Total Work Orders:** 31  
**Status Summary:** 28 Backlog, 3 Blocked  
**Last synced with Software Factory:** 2026-05-26

---

## Blocked (3)

### WO-12: Implement Cardano Smart Contract Trust Scoring System

**Blueprint:** Dynamic Trust Network and Social Verification

## ⚠ Blocked — Pending Blueprint Update

**Reason:** The Data Layer and Decentralized Identity blueprints were updated on 2026-04-25 to eliminate Cardano from Phase 1–2. Trust tier commitments are now anchored on the **Constellation Identity Metagraph** (not Cardano UTXO datums, not Cardano smart contracts). This WO references "Cardano Trust Tier Datum" and Cardano UTXO writes — all obsolete.

**Replaced by:**
- **WO-274 — Implement W3C VC 2.0 Issuance and StatusList2021 Revocation on Constellation Identity Metagraph** — covers trust tier commitment `H(tier || nonce)` submission to the Constellation Identity Metagraph
- **WO-272 — Deploy Constellation Identity Metagraph Scala L1 Validation Logic** — covers on-chain trust commitment format enforcement

**Do not implement until** this WO's description is rewritten to reflect Constellation Identity Metagraph trust tier management.

---

## Summary

*(Original content preserved for reference)*

Implement the Cardano-based trust tier and verification state management for the Dynamic Trust Network feature — anchoring trust commitments as privacy-preserving hashes on the Data L1, managing trust tier UTXO datums on Cardano for the verification-based trust levels, and providing the on-chain foundation that makes trust scores tamper-proof and portable.

---

### WO-20: Define Cardano On-Chain Schema Definitions for Identity and Credentials

**Blueprint:** Data Layer, Decentralized Identity and Authentication

## ⚠ Blocked — Pending Blueprint Update

**Reason:** The requirements document (v3.0, April 2026) eliminates Cardano from Phase 1–2 identity infrastructure. The on-chain schema layer is being rearchitected: Plutus reference scripts, UTXO datums, and CIP-25 DID metadata on Cardano are replaced by W3C VC 2.0 schemas on a dedicated **Constellation Identity Metagraph**. All Cardano-specific content in this WO (Plutus scripts, trust tier UTXO datums, bit vector format, CIP-25 DID metadata) is obsolete pending blueprint updates.

**Do not implement until:** The Data Layer and Decentralized Identity blueprints are updated to reflect the Constellation Identity Metagraph architecture, and this WO description is rewritten accordingly.

---

## Summary

Define and deploy the Cardano on-chain schema components for ECHO's identity layer — Plutus reference scripts for credential schema registry, UTXO datum structures for trust tier records and credential revocation, and the CIP-25 DID document metadata format. These are the on-chain building blocks that WO-37 (the Go Identity Service) calls at runtime. This WO owns what is deployed *on Cardano*; WO-37 owns the Go service code that calls it.

## In Scope

- Credential schema Plutus reference scripts
- Trust tier UTXO datum format: `TrustTierDatum` — `{holderDID, tier: uint8(1–5), issuerDID, issuedAt, expiresAt, commitment: H(score || nonce)}`
- Credential status bit vector format in Plutus UTXO
- CIP-25 DID document metadata format for Cardano anchoring
- Testnet deployment validation
- ADA cost verification (~0.3–0.5 ADA per issuance, ~0.2 ADA per revocation batch)

## Out of Scope

- Go Identity Service implementation (WO-37)
- Third-party IDV integration (WO-26, WO-120)

## Blueprints

- Data Layer — Specifies Cardano identity layer architecture, on-chain data structures, ADA cost model, and schema versioning
- Decentralized Identity and Authentication — Defines credential types, trust tier levels, and DID document structure

---

### WO-37: Implement Cardano Identity Layer Integration

**Blueprint:** Decentralized Identity and Authentication

## ⚠ Blocked — Pending Blueprint Update

**Reason:** The requirements document (v3.0, April 2026) eliminates Cardano from Phase 1–2 identity infrastructure. The identity layer is being rearchitected to use `did:key` (W3C DID derived from the Secure Enclave key pair) + a dedicated Constellation Identity Metagraph for VC issuance. All Atala PRISM, Veridian, and `did:prism:cardano:` references in this WO are obsolete pending blueprint updates in the Blueprints module.

**Do not implement until:** The Decentralized Identity and Authentication blueprint is updated to reflect `did:key` + Constellation Identity Metagraph, and this WO description is rewritten accordingly.

---

## Summary

Implement the Cardano identity layer integration in the Go backend Identity Service (port 8001) — including DID document creation and anchoring via Atala PRISM, verifiable credential issuance, trust tier management, credential revocation via bit vector, and credential caching with TTL-based invalidation. This is the authoritative source for user identity and trust in the Echo system. The on-chain Plutus schema definitions and UTXO datum formats are defined in WO-20; this WO owns the Go service code that calls them.

## In Scope

- DID creation via Atala PRISM/Veridian SDK: generate `did:prism:cardano:<id>`, anchor DID document on Cardano (using CIP-25 format from WO-20), include P-256 public key from Secure Enclave
- Trust tier UTXO datum management: write and read tier 1–5 assignments using the datum structure defined in WO-20 (issuer DID, expiry, commitment)
- Credential status bit vector updates in Plutus UTXO for revocation (using bit vector schema from WO-20)
- Credential issuance workflow: receive IDV callback (reference UUID only, no PII) → map to DID → determine tier → submit Cardano VC issuance transaction
- DID resolution and verification: query Cardano for DID document, extract public key, cache in Redis (60s TTL)
- Verification record storage in Cardano transaction metadata (verifier DID, method, timestamp)
- ADA fee management: platform treasury funds all Cardano transactions, users never hold ADA

## Out of Scope

- Cardano on-chain Plutus script definitions and UTXO datum formats (WO-20)
- Third-party IDV service API implementation (WO-26, WO-120)
- iOS Secure Enclave operations (Frontend work orders)
- ECHO token reward distribution (WO-184)
- ZK proof generation (ZK proof work orders)

## Requirements

Derived from the Decentralized Identity and Authentication and Data Layer blueprints.

**Trust Tier Feature Access:**

| Tier | Method | Feature Access |
|---|---|---|
| 1 (Unverified) | Self-registration | Basic messaging, no rewards |
| 2 (Newcomer) | Email/phone | Messaging + basic rewards |
| 3 (Member) | Third-party IDV | Full rewards, group creation (up to 500 members) |
| 4 (Verified) | Government ID / Apple Digital ID | Enhanced rewards multiplier, payment rails |
| 5 (Trusted) | Peer attestations + activity | Maximum multiplier, governance participation |

## Blueprints

- Decentralized Identity and Authentication — Defines DID creation, Cardano identity layer, trust tiers, credential types, and issuance process
- Data Layer — Specifies Cardano on-chain data structures, trust tier UTXO format, revocation mechanism, and ADA cost model

---

## Backlog (28)

### WO-8: Implement Constellation Metagraph Integration for Application Data and ECHO Rewards

**Blueprint:** Data Layer

## Summary

Set up the ECHO application's Constellation metagraph deployment on the public Hypergraph mainnet — including Data L1 validation logic (Scala/Euclid SDK), Currency L1 token transaction handling using Tessellation v3 primitives, metagraph node configuration, and the snapshot lifecycle from L1 validation through Global L0 finalization. This is the canonical on-chain source of truth for ECHO token balances, Merkle roots, and application data.

## In Scope

- Data L1 custom validation logic in Scala/Euclid SDK for: Merkle root structure, trust commitment format, reward claim caps, governance vote rules, schema version checks, authorized sender DID enforcement
- Currency L1 validation logic for: daily reward cap enforcement, trust-tier multiplier validation, anti-gaming velocity checks
- Metagraph node configuration for 3 L0 Hybrid Nodes (250K DAG staked per node = 750K total), 3 Currency L1 validators, 3 Data L1 validators
- Snapshot fee management: `FeeTransaction` primitive for automated DAG fee payment from treasury reserves
- Phased deployment: Phase 1 → Constellation testnet; Phase 2 → Public Hypergraph mainnet with permissioned L1 validators
- `DataL1Submission` and `CurrencyL1Transaction` type registration (defines the data schema the backend submits)
- Tessellation v3 type mapping: `TokenLock` (staking), `StakeDelegation` (validator delegation), `WithdrawLock` (unstaking, 14-day cooldown), `AtomicAction` (multi-step atomic bundles), `AllowSpend` (time-limited approvals, Phase 5), `FeeTransaction`

## Out of Scope

- Go backend Metagraph Gateway service (WO-27)
- Cardano DID/credential layer (WO-37)
- iOS client interaction (Frontend work orders)
- IPFS/Storj decentralized storage (WO-33)
- Node infrastructure provisioning (DevOps/deployment work orders)

## Requirements

Derived from the Data Layer blueprint.

**Metagraph Structure:**

| Layer | Role | Scaling |
|---|---|---|
| Currency L1 | ECHO token transactions, staking, rewards | Horizontal (add validators) |
| Data L1 | Merkle roots, trust commitments, governance | Horizontal (add validators) |
| Metagraph L0 | Aggregate L1 blocks into snapshots | Vertical (more powerful nodes) |
| Global L0 | Final consensus, immutable snapshot recording | Vertical |

**Data L1 Scala Validation (Euclid SDK):**
```scala
class MessageIntegrityValidator extends DataL1Validator {
  def validate(submission: DataL1Submission): ValidationResult = {
    submission.`type` match {
      case "message_integrity" =>
        if (submission.merkleRoot.length != 32) return Invalid("Invalid Merkle root length")
        if (submission.commitmentCount <= 0) return Invalid("Empty batch")
        if (submission.timeRange.from >= submission.timeRange.to) return Invalid("Invalid time range")
        if (submission.schemaVersion > CurrentSchemaVersion) return Invalid("Unsupported schema version")
        if (!isAuthorizedSender(submission.senderDID)) return Invalid("Unauthorized sender")
        Valid
      case _ => Invalid("Unknown submission type")
    }
  }
}
```

**Performance Targets (Phase 1 testnet → Phase 2 mainnet):**

| Metric | Launch (100K users) | Scale (1M users) |
|---|---|---|
| Data L1 TPS | 50 | 500 |
| Currency L1 TPS | 100 | 1,000 |
| End-to-end finality | < 10 seconds | < 15 seconds |
| Snapshot interval | 5 seconds | 5 seconds |

## Blueprints

- Data Layer — Defines the complete metagraph architecture, node infrastructure, Tessellation v3 type mapping, Scala validation rules, phased deployment model, and performance targets

---

### WO-15: Build Message Anchoring and Blockchain Verification System

**Blueprint:** Blockchain-Anchored Messaging with Provable Integrity

## Summary

Build the `AnchoringBatcher` in the Go backend that collects message commitment hashes, builds Merkle trees from batches of commitments, and submits Merkle roots to the Constellation Data L1 every 5 minutes or 1000 commitments (whichever comes first). Push WebSocket confirmations (with optional Merkle proof for Phase 3) to clients when snapshots finalize. The Scala `MessageIntegrityValidator` enforces submission structure at the L1 consensus layer.

## In Scope

- `AnchoringBatcher` service: collect commitment hashes from the relay, flush on 5-minute interval OR 1000 commitments
- Merkle tree construction from commitment hashes (SHA-256 leaves, double-hash `H(H(plaintext) || nonce)` scheme per E2E Encryption blueprint)
- `DataL1Submission` construction with type `"message_integrity"`, Merkle root, commitment count, time range, schema version 1
- Submission to Data L1 via Metagraph Gateway service (port 8006)
- Merkle tree storage for proof generation (in-memory + Redis with snapshot hash as key)
- WebSocket confirmation push to all affected message senders on metagraph finalization
- Confirmation payload: `{type: "confirmation", messageId, snapshotHash, snapshotHeight, merkleProof?}` — `merkleProof` populated in Phase 3+ for WO-227 client-side verification
- Snapshot reference storage (snapshot hash + height) per message batch for later proof queries
- `GET /v1/messages/{messageId}/merkle-proof` endpoint returning `{commitment, siblings, snapshotHash, snapshotHeight}` for WO-227

## Out of Scope

- Relay server transport and offline queue (WO-4)
- iOS `AnchoringTracker` (WO-28, Frontend)
- Phase 3 client-side Merkle proof verification — iOS side in WO-227; this WO provides the backend proof data
- Digital Evidence API integration for Organization tier

## Requirements

From the Blockchain-Anchored Messaging and End-to-End Message Encryption blueprints.

**AnchoringBatcher (Go):**
```go
const (
    BatchInterval = 5 * time.Minute
    MaxBatchSize  = 1000
)

func (b *AnchoringBatcher) flush() {
    batch := b.drain()
    tree := BuildMerkleTree(extractHashes(batch))
    root := tree.Root()

    txHash, err := b.metagraph.SubmitDataL1(DataL1Submission{
        Type:            "message_integrity",
        MerkleRoot:      root,
        CommitmentCount: len(batch),
        TimeRange:       TimeRange{From: batch[0].Timestamp, To: batch[len(batch)-1].Timestamp},
        SchemaVersion:   1,
    })

    if err == nil {
        b.storeTree(txHash, tree, batch)
        b.pushConfirmations(batch, root, txHash)  // Includes merkleProof in Phase 3+
    }
}
```

**Scala L1 Validator (from E2E Encryption blueprint — enforced at consensus layer):**
```scala
class MessageIntegrityValidator extends DataL1Validator {
  def validate(sub: DataL1Submission): ValidationResult = sub.`type` match {
    case "message_integrity" =>
      if (sub.merkleRoot.length != 32) Invalid("Invalid Merkle root length")
      else if (sub.commitmentCount <= 0) Invalid("Empty batch")
      else if (sub.timeRange.from >= sub.timeRange.to) Invalid("Invalid time range")
      else if (sub.schemaVersion > CurrentSchemaVersion) Invalid("Unsupported schema version")
      else if (!isAuthorizedSender(sub.senderDID)) Invalid("Unauthorized sender")
      else Valid
    case _ => Invalid("Unknown submission type")
  }
}
```

## Blueprints

- Blockchain-Anchored Messaging with Provable Integrity — Defines batching process, Merkle tree structure, `DataL1Submission` format, and WebSocket confirmation protocol
- End-to-End Message Encryption and Commitment — Defines commitment hash design (`H(H(plaintext) || nonce)`), Scala `MessageIntegrityValidator` code, and Phase 3 client-side verification flow (WO-227)
- Backend — Specifies Metagraph Gateway service integration, circuit breakers, and batch submission patterns

---

### WO-22: Build Trust Score Synchronization and Metagraph Integration

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the trust score synchronization layer between the Constellation Identity Metagraph, the metagraph trust cache, and user profiles. When a trust tier changes (new credential issued or revoked via the Constellation Identity Metagraph), the backend synchronizes state across all layers: invalidates Redis cache, updates the metagraph Data L1 trust commitment, and pushes the tier change to connected iOS clients.

## In Scope

- Trust tier change event handler: triggered by IDV callbacks and credential revocation events from the Constellation Identity Metagraph
- Redis cache invalidation: `DEL trust:{did}` on tier change
- Metagraph trust commitment update: submit new `H(trustScore || nonce)` to Data L1 after score recomputation
- WebSocket push to connected client: `{type: "trust_tier_changed", did, newTier, newScore}`
- iOS `TrustBadge` and `VerificationBadge` refresh on receiving tier change push
- Sync handling for metagraph confirmation delays: use eventual consistency, serve cached tier during finality window
- Batch sync for routine score changes (24-hour aggregated updates); immediate sync for critical events (verification, fraud reports)
- Profile update: update `trustTier` field in PostgreSQL user record

## Out of Scope

- VC issuance on Constellation Identity Metagraph (WO-274)
- Trust score computation (WO-181)
- Feature access enforcement (per-feature work orders)
- Behavioral metrics (WO-32)

## Requirements

Derived from the Dynamic Trust Network blueprint.

**Sync Flow:**
```
1. Verification completes → IDV callback processed
2. VC issued on Constellation Identity Metagraph (WO-274)
3. Trust Service: recompute score, assign new tier
4. Redis: DEL trust:{did}                     ← Cache invalidated
5. Data L1: submit H(newScore || nonce)       ← Metagraph commitment
6. PostgreSQL: UPDATE users SET trust_tier=4 WHERE did=...
7. WebSocket: push to iOS client {type: "trust_tier_changed", tier: 4}
8. iOS: refresh TrustBadge, VerificationBadge, unlock financial features
```

**Eventual Consistency Strategy:**
```
Metagraph finality: sub-second to a few seconds (Constellation DAG snapshots)
During finality gap: serve Redis cached tier (60s TTL)
On metagraph snapshot confirmation: invalidate cache, re-read metagraph state
Clients receive tier change push within ~60s of verification event
```

## Blueprints

- Dynamic Trust Network and Social Verification — Defines trust score synchronization to metagraph and profile update requirements
- Data Layer — Specifies cross-chain consistency model and eventual consistency approach

---

### WO-24: Build Smart Contract Deployment System for Constellation Network

**Blueprint:** Decentralized Bot Framework and Automation

## Summary

Build the smart contract deployment system for bots on the Constellation network. Bots are registered as smart contracts on the Constellation metagraph Data L1, ensuring their logic is immutable and their permission scope is verifiable on-chain. Supports contract deployment, verification, upgrades (proxy pattern), and versioning.

## In Scope

- Bot smart contract registration on Data L1: submit `{type: "bot_registration", botDID, botHash, permissionManifest, developerDID}` to anchor bot identity on-chain
- Contract verification: verify deployed contract matches registered source hash
- Contract upgrade via proxy pattern: deploy new version, update proxy pointer on Data L1 while preserving bot DID and permissions
- Contract versioning: maintain version history `{version, contractHash, deployedAt}` on Data L1
- Bot DID creation: each bot gets a unique `did:prism:cardano:bot-{id}` DID via Atala PRISM
- Security policy compliance check: automated scan before Data L1 submission
- Deployment confirmation within 30 seconds (Constellation metagraph finality)
- State management: bot configuration and persistent state stored in metagraph Data L1 validated storage

## Out of Scope

- Bot SDK (WO-11)
- Bot marketplace registration UI (WO-52)
- Trust scoring (WO-40)

## Requirements

Derived from the Decentralized Bot Framework blueprint.

**Bot Registration on-chain:**
```go
type BotRegistrationSubmission struct {
    Type           string   // "bot_registration"
    BotDID         string   // did:prism:cardano:bot-{id}
    DeveloperDID   string   // Bot developer's DID
    BotCodeHash    []byte   // SHA-256 of bot binary
    Permissions    []string // Declared permissions: ["messaging.read", "payment.initiate"]
    Version        string   // Semantic version
    SchemaVersion  int
}
// Anchored on Constellation Data L1
// Users can verify bot's permission scope by reading on-chain registration
```

## Blueprints

- Decentralized Bot Framework and Automation — Defines bots as smart contracts on Constellation network, authorization enforcement, censorship resistance, contract upgrades, and versioning

---

### WO-27: Implement Metagraph Integration Layer with Transaction Management

**Blueprint:** Backend, Data Layer

## Summary

Build the Metagraph Gateway service (port 8006) that handles all interactions with the Constellation metagraph infrastructure — Data L1 submissions (Merkle roots, audit log CIDs), Currency L1 submissions (token operations using Tessellation v3 primitives), snapshot listening for cache invalidation, and independent circuit breakers per chain. This is the single point of contact between the Go backend and all on-chain systems.

## In Scope

- Data L1 submission pipeline for Merkle roots (`message_integrity` type) and audit log CIDs (`audit_log` type)
- Currency L1 submission using Tessellation v3 transaction primitives: `TokenLock`, `StakeDelegation`, `WithdrawLock`, `AtomicAction`, `AllowSpend`, `SpendTransaction`, `FeeTransaction`
- `AtomicAction` bundle support for multi-step operations (reward claim + tier verify + cap update)
- Snapshot event listener: subscribe to metagraph snapshot events, invalidate Redis caches on each new snapshot, push WebSocket confirmations to clients
- Circuit breakers per downstream chain (Data L1, Currency L1, Constellation Identity Metagraph, IPFS) with independent thresholds
- Exponential backoff retry for failed submissions (queue for retry, not blocking)
- Read operations via third-party APIs; critical writes via owned nodes

## Out of Scope

- Metagraph node provisioning and configuration
- Scala/Euclid SDK L1 validation logic (separate metagraph codebase)
- Constellation Identity Metagraph DID/VC operations (WO-272, WO-273, WO-274)
- IPFS log submission (WO-53 Log Publisher)

## Requirements

Derived from the Backend and Data Layer blueprints.

**Data L1 Submission Structure:**
```go
type DataL1Submission struct {
    Type            string    // "message_integrity", "audit_log"
    MerkleRoot      []byte    // Root hash of Merkle tree (message_integrity only)
    CommitmentCount int       // Number of messages in batch
    TimeRange       TimeRange // From/To timestamps
    SchemaVersion   int       // Current: 1
    CID             string    // IPFS CID (audit_log only)
    BatchHash       []byte    // SHA-256 of encrypted log batch
}
```

**Currency L1 Transaction Primitives (Tessellation v3):**
```go
type CurrencyL1Transaction struct {
    Type            string
    TokenLock       *TokenLockData       // ECHO staking (30–365 day lock periods)
    StakeDelegation *StakeDelegationData // Delegate locked ECHO to L1 validator
    WithdrawLock    *WithdrawLockData    // Unstaking (14-day cooldown)
    AtomicBundle    *AtomicActionBundle  // Multi-step all-or-nothing execution
    AllowSpend      *AllowSpendData      // Time-limited payment approval (Phase 5)
    FeeTransaction  *FeeTransactionData  // Automated snapshot fee payment
}

type AtomicActionBundle struct {
    Actions []AtomicAction // All succeed or all fail
}

// Example: reward claim atomic bundle
// [verifyTrustTier(did), claimRewards(did, types), updateDailyCap(did)]
```

**Circuit Breaker Configuration:**
```go
type CircuitBreakerConfig struct {
    DataL1              BreakerThreshold{Failures: 5, ResetTimeout: 30 * time.Second}
    CurrencyL1          BreakerThreshold{Failures: 5, ResetTimeout: 30 * time.Second}
    IdentityMetagraph   BreakerThreshold{Failures: 5, ResetTimeout: 30 * time.Second}
    IPFS                BreakerThreshold{Failures: 5, ResetTimeout: 120 * time.Second}
}
// On circuit open: serve from Redis cache, queue submissions for retry
// Message relay NEVER blocks on chain availability
```

**Snapshot Listener:**
```go
func (g *MetagraphGateway) ListenForSnapshots(ctx context.Context) {
    for snapshot := range g.metagraphClient.SubscribeSnapshots(ctx) {
        // 1. Invalidate token balance cache (TTL: 5s)
        g.redis.Del(ctx, "balance:*")
        // 2. Invalidate trust score cache (TTL: 60s)
        g.redis.Del(ctx, "trust:*")
        // 3. Push WebSocket confirmations to anchored message clients
        for _, commitment := range snapshot.AnchoredCommitments {
            g.websocket.PushConfirmation(commitment)
        }
    }
}
```

## Blueprints

- Backend — Defines the Metagraph Gateway service, circuit breaker patterns, caching TTLs, and snapshot listener requirements
- Data Layer — Specifies Tessellation v3 transaction primitives, metagraph structure, Data L1 vs Currency L1 submission types, and phased deployment model

---

### WO-32: Develop Behavioral Trust Metrics and Scoring Engine

**Blueprint:** Dynamic Trust Network and Social Verification

## Summary

Build the behavioral trust metrics engine — the component that computes the interaction history (0–20 pts) and on-chain behavior (0–30 pts) sub-scores and applies report penalties (0 to −20 pts). These feed into the Trust Service's (WO-181) comprehensive score calculation. The engine processes behavioral data in batch (24-hour aggregations) and real-time (critical events like fraud reports).

## In Scope

- Account age scoring: 1 point per month of account age, capped at 5 points (5+ months)
- Message count scoring: 1 point per 100 messages sent, capped at 5 points (500+ messages)
- Unique contacts scoring: 1 point per 10 unique contacts, capped at 5 points (50+ contacts)
- Group participation scoring: 1 point per 5 groups joined, capped at 5 points (25+ groups)
- Payment transaction scoring: 1 point per transaction, capped at 10 points (10+ transactions) — pulled from Metagraph Currency L1 history
- Staking participation scoring: 1 point per 100 ECHO staked, capped at 10 points (1000+ ECHO) — pulled from TokenLock records
- Governance participation scoring: 1 point per vote, capped at 10 points (10+ votes) — pulled from Data L1 governance records
- Report penalty: spam reports −2 pts each (max −10), fraud reports −5 pts each (max −20), blocks −1 pt each (max −10)
- 24-hour batch job for routine metric updates; real-time updates for fraud/spam reports

## Out of Scope

- Verification-based score component (WO-37, WO-181)
- Trust tier assignment and API (WO-181, WO-49)
- Feature access enforcement (per-feature work orders)

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

**Scoring Formula:**
```go
type BehavioralScoreComponents struct {
    // Interaction History (0–20 pts)
    AccountAgePoints    int  // 1pt/month, max 5
    MessageCountPoints  int  // 1pt/100msgs, max 5
    UniqueContactPoints int  // 1pt/10contacts, max 5
    GroupJoinPoints     int  // 1pt/5groups, max 5

    // On-Chain Behavior (0–30 pts)
    PaymentTxPoints     int  // 1pt/txn, max 10 (from Currency L1)
    StakingPoints       int  // 1pt/100 ECHO staked, max 10 (from TokenLock)
    GovernancePoints    int  // 1pt/vote, max 10 (from Data L1 governance)

    // Report Penalties (0 to -20)
    SpamReportPenalty   int  // -2 per report, max -10
    FraudReportPenalty  int  // -5 per report, max -20
    BlockPenalty        int  // -1 per block, max -10
}

func (b BehavioralScoreComponents) Total() int {
    interaction := b.AccountAgePoints + b.MessageCountPoints + b.UniqueContactPoints + b.GroupJoinPoints
    behavior    := b.PaymentTxPoints + b.StakingPoints + b.GovernancePoints
    penalty     := b.SpamReportPenalty + b.FraudReportPenalty + b.BlockPenalty
    return interaction + behavior + penalty  // Combined with verification score in Trust Service
}
```

**Update Schedule:**
- Routine (24h batch): account age, message count, unique contacts, group joins, staking
- Real-time: fraud/spam reports (immediate −pts), governance votes (immediate +pts)

## Blueprints

- Decentralized Identity and Authentication — Defines all behavioral trust score components, weights, caps, and update schedules
- Dynamic Trust Network and Social Verification — Defines behavioral trust metrics and community-driven reputation

---

### WO-45: Implement Multi-Blockchain Coordination Logic for Data Layer Orchestration

**Blueprint:** Data Layer

## Summary

Implement the cross-chain coordination and eventual consistency layer in the Go backend that orchestrates data flow between the Constellation Identity Metagraph, Constellation Data L1, and IPFS/Storj storage. The backend is not a centralized authority — it is a stateless relay and cache coordinator. This work order establishes the cross-chain orchestration patterns, failure modes, and recovery targets.

## In Scope

- Cross-chain operation orchestration: identity verification flow (IDV provider → Constellation Identity Metagraph VC → metagraph trust cache), message anchoring flow, reward claim flow
- Eventual consistency coordinator: no distributed transactions across chains, operations queue and retry independently
- Chain-specific failure handling: each chain fails independently without blocking message relay
- Recovery queue for failed cross-chain operations (exponential backoff retry)
- Chain health monitoring and circuit breaker state management (per Data Layer blueprint circuit breaker specs)
- Backend role clarification enforcement: PostgreSQL/Redis as caches, metagraph as source of truth
- Operation idempotency for all cross-chain submissions (claim ID dedup, commitment dedup)

## Out of Scope

- Metagraph L1 validation logic (Scala/Euclid SDK, WO-8)
- Constellation node management
- iOS client interactions
- Individual service implementations (each covered by dedicated work orders)

## Requirements

Derived from the Data Layer blueprint.

**Cross-Chain Consistency Model:**

| Operation | Primary Chain | Failure Mode | Recovery |
|---|---|---|---|
| Message relay | None (stateless transport) | Metagraph fails → messages still delivered; commitments queue | Retry with backoff |
| Message anchoring | Metagraph Data L1 + IPFS | IPFS fails → retry queue; metagraph fails → batch queued | Retry; alert if >5 min |
| Reward claim | Metagraph Currency L1 | Queued in backend; retry on L1 recovery | Idempotent submission (claim ID dedup) |
| Identity verification | Constellation Identity Metagraph | VC pending; trust cache gets stale tier | Backend retries metagraph |
| Staking | Metagraph Currency L1 | Same as reward claim | Idempotent |

**Key Insight (enforce throughout):**
```
Message relay is DECOUPLED from on-chain operations.
A metagraph outage does NOT prevent users from messaging.
It only delays on-chain anchoring and reward distribution.
```

**Role Matrix:**
```go
// Backend is NOT the authority for these:
var AuthoritativeSources = map[string]string{
    "token_balances":     "Metagraph Currency L1 (TTL: 5s cache)",
    "trust_scores":       "Constellation Identity Metagraph + metagraph trust cache (TTL: 60s cache)",
    "message_content":    "Device-local (E2E encrypted) — backend never sees this",
    "message_metadata":   "Metagraph Data L1 Merkle root",
    "user_identity":      "did:key from PostgresDIDRegistry (TTL: 60s cache)",
    "reward_eligibility": "Data L1 validators",
}
```

**Recovery Targets:**

| Layer | RTO | RPO |
|---|---|---|
| Go backend / relay | < 60 seconds | 0 (stateless; offline queue in Redis/PG) |
| Redis cache + queue | < 30 seconds | < 1 second (AOF persistence) |
| PostgreSQL | < 30 seconds | < 1 second (synchronous replication) |

## Blueprints

- Data Layer — Defines the cross-chain consistency model, failure scenarios, recovery targets, and role clarification for the backend vs. on-chain authority
- Backend — Specifies the stateless coordinator architecture, circuit breakers, and caching TTLs

---

### WO-49: Build Trust Level Management System

**Blueprint:** Decentralized Identity and Authentication

## Summary

Build the Trust Service API layer (port 8003) that exposes trust level and tier data to other backend services. This is the query interface layer on top of the trust scoring engine (WO-181) — providing fast trust tier lookups, trust level cache management, synchronization with Constellation Identity Metagraph state, and tier downgrade handling when credentials expire or are revoked.

## In Scope

- `GET /v1/trust/{did}` — returns current trust score, tier (1–5), and tier multiplier within 100ms (Redis cache)
- Trust tier caching with 60s TTL in Redis; invalidate on snapshot confirmation from Metagraph Gateway
- Tier synchronization with Constellation Identity Metagraph: query metagraph trust record on cache miss or after revocation event
- Trust tier downgrade handling: when credential revocation detected, recompute score, update tier, clear cache
- Trust level history endpoint: `GET /v1/trust/{did}/history` returns verification events and tier changes
- Internal service-to-service trust tier query (used by: Rewards Service, Message Relay rate limiter, Group Service, bot permission checks)
- Tier change notification to connected clients via WebSocket: `{type: "trust_tier_changed", did, newTier}`

## Out of Scope

- Trust score computation formula (WO-181, Trust Service internal)
- Feature access enforcement (each feature work order enforces its own tier checks)
- VC storage on Constellation Identity Metagraph (WO-274)
- Identity verification processing (IDV work orders)

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

**Trust Tier API Response:**
```go
// GET /v1/trust/:did
type TrustLevelResponse struct {
    DID            string     `json:"did"`
    TrustScore     int        `json:"trust_score"`       // 0–100
    TrustTier      int        `json:"trust_tier"`        // 1–5
    Multiplier     float64    `json:"multiplier"`         // 0.0, 0.5, 1.0, 1.5, 2.0
    VerifiedAt     *time.Time `json:"verified_at,omitempty"`
    ExpiresAt      *time.Time `json:"expires_at,omitempty"`
    CachedAt       time.Time  `json:"cached_at"`
}
```

**Cache + Metagraph Fallback:**
```go
func (s *TrustService) GetTrustLevel(did string) (*TrustLevelResponse, error) {
    // 1. Check Redis cache (60s TTL)
    if cached := s.redis.Get("trust:" + did); cached != nil {
        return deserialize(cached), nil
    }
    // 2. Cache miss: query Constellation Identity Metagraph trust record
    tier, err := s.identityMetagraph.GetTrustTierRecord(did)
    if err != nil {
        return nil, err
    }
    // 3. Compute score + build response
    score := s.trustScorer.CurrentScore(did)
    resp := &TrustLevelResponse{DID: did, TrustScore: score, TrustTier: tier.Level, ...}
    s.redis.Set("trust:"+did, serialize(resp), 60*time.Second)
    return resp, nil
}
```

**Tier Downgrade on Revocation:**
```go
func (s *TrustService) HandleCredentialRevocation(did string, credentialType string) {
    s.redis.Del("trust:" + did)               // Invalidate cache
    newScore := s.trustScorer.Recompute(did)  // Recompute without revoked credential
    newTier := tierFromScore(newScore)
    s.websocket.PushTierChange(did, newTier)  // Notify connected client
}
```

## Blueprints

- Decentralized Identity and Authentication — Defines trust tier levels, feature access table, trust score components, and tier change handling
- Data Layer — Specifies metagraph trust tier record structure and revocation mechanism

---

### WO-94: Create Blockchain Hash Anchoring and Cryptographic Proof System

**Blueprint:** Disappearing Messages with Cryptographic Verification

## Summary

Implement the `DisappearingMessageProof` system — a cryptographic proof structure that demonstrates a message existed and was delivered at a specific timestamp, both before and after plaintext deletion. Proofs leverage the on-chain Merkle root (which persists indefinitely) and the local commitment hash. This is the backend endpoint and iOS proof generation utility.

## In Scope

- `DisappearingMessageProof` Swift struct for iOS with before/after deletion proof types
- Backend endpoint: `GET /v1/messages/{messageId}/proof` returning Merkle proof data
- Merkle proof retrieval from stored tree (post-anchoring)
- Proof serialization in JSON and PDF formats for export
- Shareable proof format for legal/compliance use cases
- Third-party verification support (anyone can verify Merkle root on DAG Explorer)
- Legal proof generation: `generateLegalProof(messageId)` returning existence proof with verification URL

## Out of Scope

- Phase 3 client-side Merkle proof verification (trustless verification, future work order)
- Legal hold infrastructure (separate compliance work order)
- Digital Evidence API (Organization tier, separate work order)
- Smart contract time-lock mechanisms (deletion is client-side, not smart contract — see WO-105)

## Requirements

Derived from the Disappearing Messages blueprint.

**Proof Capabilities:**

| Proof Type | Can Prove | Cannot Prove |
|---|---|---|
| Before Deletion | Message existed, exact content, Merkle inclusion | N/A |
| After Deletion | Message existed at timestamp, on-chain Merkle root | Message content, commitment verification |

**Proof Structure (iOS):**
```swift
struct DisappearingMessageProof {
    let messageId: String
    let conversationId: String
    let senderDID: String
    let recipientDID: String
    let timestamp: Date
    let expiresAt: Date
    let snapshotHash: String         // On-chain snapshot hash
    let snapshotHeight: Int          // On-chain block height
    let merkleRoot: Data             // On-chain Merkle root
    let merkleProof: [Data]?         // Sibling hashes (nil after deletion)
    let commitmentHash: Data?        // Local (nil after deletion)
    let proofType: ProofType
    var verificationURL: URL {
        URL(string: "https://dagexplorer.io/snapshot/\(snapshotHash)")!
    }

    enum ProofType { case beforeDeletion, afterDeletion }
}
```

**Legal Proof (Org tier):**
```swift
func generateLegalProof(messageId: String) -> LegalProof {
    return LegalProof(
        exists: true,
        timestamp: message.timestamp,
        snapshotHash: anchoringInfo.snapshotHash,
        merkleRoot: anchoringInfo.merkleRoot,
        contentAvailable: false,  // Deleted
        verificationURL: "https://dagexplorer.io/snapshot/\(snapshotHash)"
    )
}
```

## Blueprints

- Disappearing Messages with Cryptographic Verification — Defines `DisappearingMessageProof` structure, before/after deletion proof types, proof capabilities, and legal proof generation
- Blockchain-Anchored Messaging with Provable Integrity — Defines Merkle proof structure and on-chain verification

---

### WO-99: Implement Core Token Management and Distribution System

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the core ECHO token management infrastructure on the Constellation Metagraph Currency L1 — establishing the canonical five-pool allocation model, per-pool balance tracking, vesting schedule enforcement via `TokenLock`, and the Rewards Service distribution API. The actual genesis deployment (minting all pools and creating founder `TokenLock` positions) is handled by WO-214.

## In Scope

- Five-pool allocation tracking in metagraph state (per canonical ECHO Token Economics blueprint):
  - Community Rewards: 400,000,000 ECHO (emission account, 10-year declining curve)
  - Treasury: 220,000,000 ECHO (3-of-5 multi-sig → DAO Phase 4+)
  - Founders: 180,000,000 ECHO (via founder `TokenLock` positions, see WO-214)
  - Future Team & Advisors: 100,000,000 ECHO (multi-sig controlled)
  - Ecosystem & Partnerships: 100,000,000 ECHO (governance-controlled)
- Vesting schedule enforcement: founder allocations locked via `TokenLock` on Currency L1 — 1-year cliff, 1/36th monthly vest over 36 months; enforced by Scala validation
- `WithdrawLock` requests: 14-day cooldown on unlock (enforced by Currency L1)
- Distribution API (Rewards Service, port 8004): `POST /v1/rewards/distribute` — validate pool source and available balance before submitting `AtomicAction` or `SpendTransaction` to Currency L1
- Allocation balance query: `GET /v1/tokens/allocation` — real-time pool balances from metagraph state cache
- Currency L1 Scala validation: reject any mint transaction after genesis (fixed supply enforcement)

## Out of Scope

- Genesis deployment and founder TokenLock creation (WO-214)
- Treasury management and PacaSwap seeding (WO-215)
- Messaging reward volume-decay calculation (WO-213)
- Staking UI and flows (WO-127)
- Burn mechanisms (WO-155, WO-170)
- Cross-chain bridge (WO-174)

## Requirements

From the ECHO Token Economics and Founder Allocation blueprint (canonical source):

**Fixed Supply: 1,000,000,000 ECHO — no minting after genesis.**

| Allocation | % | Tokens | Vesting |
|---|---|---|---|
| Community Rewards | 40% | 400,000,000 | Emitted over 10 years via declining curve |
| Treasury | 22% | 220,000,000 | Multi-sig (founders) → DAO governance (Phase 4+) |
| Founders (5) | 18% | 180,000,000 | 4-year vest, 1-year cliff, Currency L1 `TokenLock` |
| Future Team & Advisors | 10% | 100,000,000 | Same vesting when allocated |
| Ecosystem & Partnerships | 10% | 100,000,000 | Governance-approved disbursement |

**Tessellation v3 Primitives:**
```
TokenLock:        Lock founder tokens (1-year cliff + 36-month monthly vest)
WithdrawLock:     Unlock vested tokens (14-day cooldown)
AtomicAction:     Reward claim bundles (verify tier + decay + credit + counter update)
SpendTransaction: Distributing rewards to user wallets
FeeTransaction:   Automated snapshot fee payment from treasury DAG reserves
AllowSpend:       VIP subscriptions; bot payments; marketplace escrow (time-limited)
```

## Blueprints

- ECHO Token Economics and Founder Allocation — Defines canonical five-pool allocation, founder vesting schedule, emission curve, Tessellation v3 primitive usage, and fixed supply enforcement
- Data Layer — Specifies Currency L1 architecture, vesting mechanics, and metagraph state management

---

### WO-105: Implement Time-Locked Smart Contract Integration

**Blueprint:** Disappearing Messages with Cryptographic Verification

## Summary

Implement the client-side disappearing message timer and deletion mechanism. Deletion is triggered entirely by client-side timers — there are no server-side smart contracts coordinating deletion. Each client device independently deletes the message plaintext and keys at expiration time. The on-chain Merkle root persists indefinitely as proof of existence.

## In Scope

- iOS `DisappearingMessageView` with live countdown timer display (`CountdownTimer` SwiftUI component)
- Timer expiry action: delete plaintext, encryption keys, and media from local SwiftData and Keychain
- `deleteMessageLocally(_:)` function sequence: delete from SwiftData → delete keys from Keychain → delete cached media → clear from memory
- Timer persistence across app restarts: store expiry timestamp in SwiftData, restart timer on app launch
- Per-message and per-conversation expiration settings (10s, 1m, 5m, 1h, 1d, 7d; custom for VIP)
- Expiry timestamp embedded in encrypted message metadata (recipient receives it with the message)
- Post-deletion state: commitment hash preserved (cannot verify without plaintext), on-chain Merkle root intact
- iOS `BGTaskScheduler` for background deletion when app is not in foreground

## Out of Scope

- Smart contract deployment on Constellation (deletion is client-side, not on-chain)
- Server-side deletion coordination (each device deletes independently)
- Proof generation (WO-94)
- Trust score restrictions (separate work order WO-143)
- Screenshot prevention (WO-125, platform limitation)

## Requirements

Derived from the Disappearing Messages blueprint.

**Critical Architecture Clarification:**
```
⚠️ Deletion is client-side ONLY — no smart contracts trigger deletion.
The "time-locked encryption" concept in early designs was superseded.
Each client's timer independently triggers local deletion.
The on-chain Merkle root is permanent proof of existence, not a deletion trigger.
```

**iOS Countdown Timer:**
```swift
struct DisappearingMessageView: View {
    let message: Message
    var body: some View {
        HStack {
            MessageBubble(message: message)
            if let expiresAt = message.expiresAt {
                CountdownTimer(expiresAt: expiresAt) { expired in
                    if expired { deleteMessageLocally(message.id) }
                }
            }
        }
    }
}
```

**Local Deletion:**
```swift
func deleteMessageLocally(_ messageId: String) async {
    try? await database.deleteMessage(messageId)  // SwiftData
    try? keychain.deleteKey(for: messageId)        // Encryption keys
    try? mediaCache.deleteMedia(for: messageId)    // Cached media
    messageCache.removeValue(forKey: messageId)    // In-memory
    // PRESERVE: commitment hash (for future proof generation)
    // PRESERVE: expiry timestamp (for audit log)
}
```

**What Gets Deleted vs. Preserved:**
```
DELETED:  plaintext, encryption keys (X25519/ChaCha20), media attachments, sender/timestamp metadata
PRESERVED: commitment hash (local, can't verify without plaintext), on-chain Merkle root (forever)
```

## Blueprints

- Disappearing Messages with Cryptographic Verification — Defines client-side deletion architecture, countdown timer implementation, deletion sequence, and what gets deleted vs. preserved

---

### WO-114: Build Reward Calculation Engine with Trust Score Integration

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Build the ECHO reward calculation engine for **payment rail and referral rewards**. The messaging reward calculation (volume-decay formula) is implemented in WO-213. This work order covers the remaining reward types plus the shared reward validation API used by all claim paths.

**Critical correction from combined Frontend blueprint:** Reward tier multipliers are `1.0×/1.2×/1.5×/2.0×/3.0×` (Tiers 1–5) — not `0.5×/1.0×/1.5×/2.0×`. These reward multipliers are **different from governance multipliers** (`0.0×/0.5×/1.0×/1.5×/2.0×`).

## In Scope

- **Payment rail reward**: 1–5 ECHO per transaction based on transaction value and verification level; Tier 3+ required; verification multiplier: Tier 3 = 1.5×, Tier 4 = 2.0×, Tier 5 = 3.0×
- **Referral reward**: 50 ECHO to referrer AND 50 ECHO to referee when referred user completes identity verification AND sends first 100 messages; Tier 2+ required for both parties
- Reward validation API: `POST /v1/rewards/calculate` — input `{userId, activityType, count, trustTier}` → output `{rewardAmount, breakdown, decayFactor, appliedMultiplier}`
- Anti-gaming pre-checks (before Currency L1 submission): duplicate nonce detection, trust tier mismatch, velocity limits per reward type
- Referral eligibility validation: verify referred user completed verification + 100-message threshold before distributing
- `AtomicAction` bundle construction using `.recordNetworkActivity` (not `.updateDailyCap`)

## Out of Scope

- Messaging reward volume-decay formula (WO-213)
- Token distribution execution (WO-99)
- Trust score computation (WO-181)
- Staking APY rewards (WO-127)

## Requirements

From the ECHO Token Economics blueprint and corrected by the combined Frontend blueprint:

**Reward Tier Multipliers (canonical — REWARDS scale, not governance):**
| Trust Tier | Reward Multiplier |
|---|---|
| Tier 1 (Unverified) | 1.0× (basic rewards) |
| Tier 2 (Newcomer) | 1.2× |
| Tier 3 (Member) | 1.5× |
| Tier 4 (Verified) | 2.0× |
| Tier 5 (Trusted) | 3.0× |

**Reward Types and Rates:**
| Reward | Base Rate | Trust Requirement | Mechanism |
|---|---|---|---|
| Messaging | 0.1 ECHO/msg with volume decay (→ WO-213) | Tier 2+ | AtomicAction |
| Payment rail | 1–5 ECHO/transaction | Tier 3+ | AtomicAction |
| Referral | 50 ECHO each (referrer + referee) | Tier 2+ both | AtomicAction on verification |
| Staking APY | 5–15% | Tier 2+ | TokenLock + StakeDelegation |

```go
type RewardCalculation struct {
    UserDID          string
    ActivityType     string    // "payment_rail", "referral"
    ActivityCount    int
    TrustTier        int       // 1–5
    RewardMultiplier float64   // 1.0, 1.2, 1.5, 2.0, or 3.0 (NOT 0.5–2.0)
    BaseReward       Decimal
    FinalReward      Decimal
}
```

## Blueprints

- ECHO Token Economics and Founder Allocation — Defines canonical reward rates, trust requirements, and AtomicAction enforcement
- Frontend — Clarifies reward multipliers (1.0×–3.0×) are distinct from governance multipliers (0.0×–2.0×)
- Data Layer — Specifies Currency L1 validation rules for reward claims

---

### WO-115: Build Trust Score Integration and Abuse Prevention System

**Blueprint:** Disappearing Messages with Cryptographic Verification

**Purpose**: Integrate disappearing messages with the existing trust scoring system to prevent abuse and harassment by restricting very short expiration timeframes for users with low trust scores, while providing transparent restriction policies and appeal mechanisms.

**Requirements**:
- Implement trust score-based restrictions that prevent users with low trust scores from setting expiration times shorter than defined thresholds
- Define and enforce minimum timeframe requirements based on user trust score levels (e.g., users below score X cannot set messages to expire in less than Y minutes)
- Provide restriction escalation system that increases limitations for repeated policy violations
- Create transparent restriction communication that informs users why certain expiration options are unavailable
- Implement restriction appeals process that allows users to request review of imposed limitations
- Support restriction monitoring that tracks usage patterns and identifies potential abuse
- Enable restriction enforcement that prevents API calls and UI interactions for prohibited expiration settings
- Provide restriction exception handling for verified users or special circumstances
- Maintain restriction audit trail that logs all enforcement actions and policy applications
- Ensure restriction system integrates seamlessly with existing trust score calculation and updates

**Out of Scope**:
- Trust score calculation algorithm modifications
- User verification process changes
- General abuse detection beyond disappearing message restrictions
- Trust score display in user interface

---

### WO-156: Implement Blockchain Integration and Metadata Anchoring System

**Blueprint:** Public and Private Groups with Verified Status Display

## Summary

Implement the blockchain anchoring of group metadata and governance decisions on the Constellation Data L1. Group membership counts (as hashes), governance vote results, and group creation events are anchored for tamper-proof record-keeping. Member identities are never on-chain.

## In Scope

- Group creation event on Data L1: `{type: "group_metadata", groupId, adminDID, memberCountHash, createdAt}` (same as WO-70 anchoring — this WO provides the updates pipeline)
- Periodic metadata updates: every 24 hours or on significant membership changes, update member count hash on-chain
- Governance vote anchoring: after each vote closes, submit `{type: "governance_vote", groupId, proposalId, result, participationPct, closedAt}` to Data L1
- Moderation decision anchoring: for Organization tier groups, submit event hash on significant actions (ban, major policy change)
- Public verification interface: `/v1/groups/{id}/blockchain-proof` — returns anchored metadata with Data L1 snapshot references
- Local cache of blockchain-confirmed group state with sync from Metagraph Gateway snapshot events

## Out of Scope

- Full message history on blockchain
- Token/currency functionality
- Custom blockchain development

## Requirements

Derived from the Public and Private Groups blueprint.

**Metadata Anchoring Schedule:**
```
Group creation → Immediate Data L1 submission (via Metagraph Gateway)
Membership change (join/leave) → Queue for next 24h batch update
Governance vote close → Immediate Data L1 submission
Moderation ban (Org tier) → Immediate hash submission
```

**On-Chain Data:**
```go
// Group metadata on Data L1 (submitted by backend Metagraph Gateway):
// ✅ groupId (UUID), adminDID, memberCountHash, timestamps
// ✅ governance vote outcomes (aggregated, not per-voter)
// ❌ NEVER: member DIDs, group name, description, messages
```

## Blueprints

- Public and Private Groups with Verified Status Display — Defines Data L1 group metadata anchoring, member count hash, zero-knowledge statistics, governance vote recording, and blockchain verification interface

---

### WO-164: Develop Constellation Metagraph Integration Layer

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the Constellation Metagraph integration layer for ECHO token reward economics — the Scala/Euclid SDK components that handle the Data L1 business logic validation for reward distribution, earning caps, anti-gaming checks, and state synchronization with Currency L1. This is the metagraph-side complement to the Go backend Metagraph Gateway (WO-27).

## In Scope

- Data L1 Scala validation rules (Euclid SDK): daily earning cap check per DID, trust tier multiplier validation, anti-gaming velocity checks, progressive decay calculation, reward claim format validation
- Currency L1 Scala validation rules: ECHO balance checks, `TokenLock`/`StakeDelegation`/`WithdrawLock` enforcement, burn address validation, validator stake minimums
- Metagraph L0 snapshot aggregation: package Currency L1 + Data L1 blocks into snapshots every 60 seconds
- Reward eligibility state management: per-DID daily counters maintained in Data L1 state tree
- State synchronization between Data L1 reward state and Currency L1 balance state
- Metagraph monitoring: block production rates, state health, consensus round tracking

## Out of Scope

- Go backend Metagraph Gateway (WO-27 — handles submissions from backend)
- Global L0 (Hypergraph public network — external)
- Validator node networking (infrastructure work)

## Requirements

Derived from the Data Layer blueprint.

**Data L1 Validation (Scala):**
```scala
class RewardClaimValidator extends DataL1Validator {
  def validate(claim: RewardClaim): ValidationResult = {
    val dailyUsed = state.getDailyUsed(claim.userDID, claim.rewardType)
    if (dailyUsed + claim.amount > dailyCap(claim.trustTier, claim.rewardType))
      return Invalid("Daily cap exceeded")
    if (claim.multiplier != expectedMultiplier(claim.trustTier))
      return Invalid("Invalid multiplier")
    if (isAntiGamingTriggered(claim.userDID))
      return Invalid("Anti-gaming restriction")
    Valid
  }
}
```

## Blueprints

- Data Layer — Defines the complete metagraph three-layer architecture, Data L1/Currency L1 validation rules, and Euclid SDK requirements

---

### WO-179: Implement Constellation Metagraph Three-Layer Architecture Integration

**Blueprint:** ECHO Token Reward System and Incentive Economy

## Summary

Implement the complete Constellation Metagraph three-layer architecture integration for ECHO tokens — establishing the full Currency L1 + Data L1 + Metagraph L0 pipeline with proper state synchronization, validator coordination, and snapshot production. This is the Phase 2 foundational integration that WO-164 (Phase 2) and WO-8 (Phase 2) both build toward; this work order provides the complete picture.

## In Scope

- Currency L1 full implementation: validate all 5+ transaction types, maintain authoritative token ledger, produce L1 blocks every ~5 seconds using Euclid SDK (Scala)
- Data L1 full implementation: validate reward distribution rules, maintain per-DID daily counters, enforce anti-gaming, produce Data L1 blocks
- Metagraph L0 snapshot production: aggregate Currency L1 + Data L1 blocks every 60 seconds; sign snapshots; submit to Global L0
- Cross-layer state synchronization: Data L1 reward eligibility decisions → Currency L1 execution ordering; consistent state merkle roots
- Atomic transaction processing: all-or-nothing execution for multi-step operations; rollback on failure
- Layer health monitoring: block production rates, consensus round latency, state tree size, snapshot submission status

## Out of Scope

- Global L0 Hypergraph infrastructure (Constellation public network)
- Validator node networking and P2P
- Go backend Metagraph Gateway integration (WO-27 handles that side)

## Requirements

Derived from the Data Layer blueprint.

**Architecture note:** This WO-179 is the comprehensive integration of the three-layer architecture. WO-8 (Phase 2) and WO-164 (Phase 2) build specific aspects; this WO ensures the full end-to-end pipeline works cohesively.

```
Currency L1 → validates token operations → produces blocks
Data L1     → validates business rules → produces blocks  
Metagraph L0 → aggregates → produces snapshots → submits to Global L0
```

## Blueprints

- Data Layer — Defines the complete three-layer metagraph architecture, validation rules, state synchronization, and snapshot lifecycle

---

### WO-181: Build Dynamic Trust Scoring Algorithm

**Assignee:** Chad Cromwell

**Blueprint:** Decentralized Identity and Authentication

## Summary

Implement the Trust Service (port 8003) — the trust scoring engine that computes dynamic trust scores from 0–100 for each user based on verification level, interaction history, on-chain behavior, and report penalties. Trust scores determine tier access, reward multipliers, and governance participation. The service maintains scored state with real-time updates for critical events and batch updates for routine activity.

## In Scope

- Trust score computation engine implementing the 4-component weighted formula
- Real-time updates for critical events: verification completion (+5 to +30 points), fraud/spam reports (−2 to −5 points)
- Batch updates for routine activity: message counts, contact interactions (every 24 hours)
- Trust tier assignment (Tier 1–5) and feature access table enforcement
- Redis cache for trust scores (60s TTL) and trust tiers
- Trust score API: `GET /v1/trust/{did}` returns current score + tier within 100ms
- Trust score history storage in PostgreSQL for audit and user transparency
- Trust score recalculation on credential revocation or verification status changes
- Trust tier governance multipliers for ECHO token governance (used by Rewards service)

## Out of Scope

- Feature access enforcement (each feature work order enforces its own access)
- Verification processing that produces trust score inputs (IDV work orders)
- ECHO reward multiplier application (Rewards Service, WO-184)

## Requirements

Derived from the Decentralized Identity and Authentication blueprint.

**Trust Score Formula:**
```go
type TrustScoreComponents struct {
    VerificationLevel int  // 0–30 pts: Unverified=0, Device=5, KYC-lite=15, High-Assurance=30
    InteractionScore  int  // 0–20 pts: account age + message count + contacts + group participation
    BehaviorScore     int  // 0–30 pts: payment txns + staking + governance participation
    ReportPenalty     int  // 0 to -20 pts: spam reports + fraud reports + blocks
}

func (t TrustScoreComponents) Total() int {
    return max(0, min(100, t.VerificationLevel + t.InteractionScore + t.BehaviorScore + t.ReportPenalty))
}

func tierFromScore(score int) int {
    switch {
    case score >= 81: return 5  // Trusted: max multiplier, governance
    case score >= 61: return 4  // Verified: enhanced rewards, payment rails
    case score >= 41: return 3  // Member: full rewards, group creation up to 500
    case score >= 21: return 2  // Newcomer: basic rewards
    default:          return 1  // Unverified: basic messaging, no rewards
    }
}
```

**Feature Access by Tier:**

| Tier | Max Group Size | Daily Messages | File Sharing | Features |
|---|---|---|---|---|
| 1 (Unverified) | 10 | 100/day | None | Basic messaging |
| 2 (Newcomer) | 50 | 500/day | Basic | + Basic rewards |
| 3 (Member) | 500 | 1000/day | Up to 100MB | + Full rewards, group creation |
| 4 (Verified) | 10,000 | 2000/day | Up to 500MB | + Payment rails, enhanced multiplier |
| 5 (Trusted) | 1,000,000 | Unlimited | Up to 2GB | + Governance voting, max multiplier |

**Governance Multipliers (for Rewards Service):**

| Tier | Multiplier |
|---|---|
| 1 | 0.0x (cannot vote or earn rewards) |
| 2 | 0.5x |
| 3 | 1.0x |
| 4 | 1.5x |
| 5 | 2.0x |

## Blueprints

- Decentralized Identity and Authentication — Defines trust score formula, tier definitions, component weights, update frequency, and feature access table
- Frontend — Specifies trust tier governance multipliers used in wallet voting power calculation

---

### WO-206: Implement ECHO Tokenomics, Founder Allocation, and Token Launch

**Type:** Build

**Blueprint:** ECHO Tokenomics, Founder Allocation, and Token Launch

## Summary

Implement the AllowSpend/SpendTransaction VIP subscription payment rails on Currency L1, the emission analytics API for annual budget monitoring, and the governance voting weight enforcement for locked founder positions. Most tokenomics work is covered by WO-213 (auto-scaling rewards), WO-214 (genesis deployment), and WO-215 (treasury/PacaSwap). This work order covers the remaining FRs from the ECHO Tokenomics blueprint.

## In Scope

- **VIP subscription AllowSpend (FR implicit):** `AllowSpend` authorization for $9.99/month VIP subscriptions — user approves time-limited recurring spend; `SpendTransaction` executes monthly renewal; Currency L1 enforces amount and expiry
- **Emission status API:** `GET /tokens/emission/status` → returns `{currentYear, distributedToDate, annualCap, remainingBudget}` from Currency L1 state cache (TTL: 5s)
- **FR10 — Governance voting from locked positions:** Ensure founder `TokenLock` positions are eligible to vote from genesis; `StakedECHO × TrustTierMultiplier` weight calculation applies to locked amounts; Scala L1 validation confirms locked amounts are included in voting weight
- **Annual emission cap monitoring UI:** iOS Wallet tab emission gauge showing Year N progress vs. cap; alert when > 90% consumed (governance notification per Production Launch blueprint)
- **VIP tier rate limit wiring:** After successful AllowSpend, backend updates user's rate limit tier in Redis from base → VIP (2×–10× increase)

## Out of Scope

- Auto-scaling reward formula (WO-213)
- Genesis block deployment (WO-214)
- PacaSwap liquidity seeding (WO-215)
- Founder departure revocation (separate work order)
- Founder vesting dashboard panel (separate work order)

## Requirements

From the ECHO Tokenomics, Founder Allocation, and Token Launch blueprint:

**FR10 — Governance voting from locked tokens:** Locked founder TokenLock positions are eligible for governance voting from genesis. Voting weight = locked amount × trust tier governance multiplier.

**AllowSpend for VIP:**
```go
// Currency L1 AllowSpend authorization
type AllowSpendData struct {
    SpenderDID  string    // Backend service DID
    MaxAmount   uint64    // 9.99 ECHO/month equivalent
    ExpiresAt   time.Time // 30 days from authorization
    Purpose     string    // "vip_subscription"
}
// After AllowSpend: backend executes SpendTransaction on renewal date
```

**Emission Status:**
```go
// GET /tokens/emission/status
type EmissionStatus struct {
    CurrentYear       int     `json:"current_year"`
    AnnualCap         uint64  `json:"annual_cap"`        // Year-N cap in ECHO units
    DistributedToDate uint64  `json:"distributed_to_date"`
    RemainingBudget   uint64  `json:"remaining_budget"`
    PercentConsumed   float64 `json:"percent_consumed"`
}
```

## Blueprints

- ECHO Tokenomics, Founder Allocation, and Token Launch — Defines FR10 governance voting from locked positions, AllowSpend primitive usage, and VIP subscription model
- Data Layer — Defines Tessellation v3 AllowSpend/SpendTransaction primitives and Currency L1 validation

---

### WO-209: Implement Privacy-Preserving Blockchain Data Model

**Type:** Build

**Blueprint:** Privacy-Preserving Blockchain Data Model

## Summary

Implement schema versioning enforcement in the Constellation metagraph Data L1 Scala validators (supporting current and one prior schema version, governance-gated upgrades), and the Scala pattern-matching guards that reject T0–T4 data (PII, IP addresses, message content) from on-chain submissions. The T0–T7 classification model and CI enforcement are covered by WO-217 and WO-35; this work order covers the Scala on-chain enforcement layer.

## In Scope

- **Schema version enforcement (Scala):** `SupportedSchemaVersions = Set(1, 2)` — current + one prior. `DataL1Submission` with unsupported schema version returns `ValidationResult.Invalid`. Schema version upgrades require governance proposal + vote + activation at a specific future snapshot height
- **Governance-gated schema upgrades:** Schema change proposal submitted to Data L1 governance; activated at `activationSnapshotHeight` in proposal. Validators begin accepting new version at that height; old version deprecated 6 months later
- **Zero PII Scala guards:** Pattern-matching validators in Euclid SDK reject submissions containing:
  - Email addresses (`@` + domain pattern)
  - E.164 phone numbers
  - IP addresses (IPv4/IPv6 patterns)
  - Raw trust scores (integer 0–100 not wrapped in commitment structure)
  - Device fingerprints
  - Message content (non-hash strings > 64 bytes in `message_integrity` submissions)
  - Full member lists (must use member count hash)
- **Trust commitment format validation:** Reject trust commitments missing required nonce structure `H(score || nonce)` — commitment must be exactly 32 bytes
- **Schema versioning governance proposal template:** Template for proposing schema changes (title, description, new schema version, activation height, backward compatibility notes)

## Out of Scope

- CI-level T0–T7 enforcement (WO-217)
- Backend pre-validation (WO-35)
- Relay node registry implementation (separate Phase 4 work order)

## Requirements

From the Privacy-Preserving Blockchain Data Model blueprint:

**Schema versioning Scala code:**
```scala
val SupportedSchemaVersions = Set(1, 2)  // current + one prior

def validate(sub: DataL1Submission): ValidationResult =
  if (!SupportedSchemaVersions.contains(sub.schemaVersion))
    ValidationResult.Invalid(s"Unsupported schema version: ${sub.schemaVersion}")
  else ValidationResult.Valid
```

**Rejected patterns (Scala guards):**
| Pattern | Reason |
|---|---|
| Email addresses | T0 PII |
| E.164 phone numbers | T0 PII |
| IP addresses | T0 network metadata |
| Raw trust scores (int 0–100 without commitment) | T6 — must be committed |
| Device fingerprints | T0 PII |
| Message content > 64 bytes in `message_integrity` | T5 — only hashes allowed |
| Full member lists | Must use member count hash (T5) |

## Blueprints

- Privacy-Preserving Blockchain Data Model — Defines schema versioning with governance-gated upgrades, Scala L1 PII guards, T0–T7 classification enforcement at consensus layer

---

### WO-210: Implement Production Launch, Infrastructure, and Deployment

**Type:** Build

**Blueprint:** Production Launch, Infrastructure, and Deployment

## Summary

Deploy Prometheus + Grafana monitoring, configure PagerDuty alerting with the thresholds from the Production Launch blueprint, set up Kubernetes liveness/readiness probes for all 10 backend services, and document the disaster recovery runbook with RTO/RPO targets. The infrastructure provisioning (Kubernetes cluster, node infrastructure) is covered by separate work orders; this work order covers the operational observability layer.

## In Scope

- **Prometheus scraping:** Configure scraping for all 10 Go microservices (ports 8000–8009), NATS JetStream, Redis, PostgreSQL, metagraph node metrics
- **Grafana dashboards:** Message delivery rate, relay latency (P50/P95/P99), metagraph finality time, WebSocket connection count per pod, offline queue depth, circuit breaker state, token emission budget consumption
- **PagerDuty alerting rules (from blueprint):**

| Metric | Warning | Critical | Channel |
|---|---|---|---|
| Message delivery rate | < 99.5% | < 99.0% | PagerDuty (on-call) |
| Relay latency P99 | > 300ms | > 1000ms | PagerDuty |
| Metagraph finality | > 15s | > 30s | PagerDuty |
| DAG snapshot fee reserves | < 30 days | < 7 days | Slack + PagerDuty |
| Redis queue depth | > 10K/recipient | > 50K | Slack |
| Circuit breaker opens | Any | 3+ simultaneous | PagerDuty |
| L0 node uptime | < 99% | < 95% | PagerDuty |
| Token emission budget | > 90% annual | > 99% annual | Governance notification |

- **Kubernetes probes:** Liveness (HTTP GET `/healthz`) and readiness (HTTP GET `/readyz`) probes for all services; restart policy on 3 consecutive liveness failures
- **Disaster recovery runbook:** Document recovery procedures for each failure scenario with RTO/RPO targets from blueprint; PostgreSQL replica promotion playbook; Redis AOF restore procedure; metagraph node replacement checklist

## Out of Scope

- Kubernetes cluster provisioning (Phase 2 mainnet deployment WO)
- Application-level metrics instrumentation (each service's own WO)
- Log analysis tooling (separate ELK/Loki stack)

## Requirements

From the Production Launch, Infrastructure, and Deployment blueprint — Monitoring and Alerting section and Disaster Recovery section.

**Recovery Targets:**
| Layer | RTO | RPO |
|---|---|---|
| Go backend / relay | < 60 seconds | 0 |
| Redis cache + queue | < 30 seconds | < 1 second |
| PostgreSQL | < 30 seconds | < 1 second |
| Metagraph L0 node | < 5 minutes | 0 (consensus) |

## Blueprints

- Production Launch, Infrastructure, and Deployment — Defines monitoring metrics, alerting thresholds, PagerDuty escalation, Kubernetes health check patterns, and disaster recovery targets

---

### WO-213: Implement Auto-Scaling Volume Decay Reward Engine

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation

## Summary

Replace the hard daily cap reward approach with the canonical auto-scaling volume-decay model defined in the ECHO Token Economics blueprint. The new model enforces smooth diminishing returns via an on-chain formula — making spam farming economically irrational while allowing genuine high-volume users to keep earning. Enforced on-chain via AtomicAction bundles on the Currency L1.

**Critical correction from combined Frontend blueprint:** Reward tier multipliers use a **different scale from governance multipliers**. Reward multipliers are `1.0×/1.2×/1.5×/2.0×/3.0×` (Tiers 1–5). Governance multipliers are `0.0×/0.5×/1.0×/1.5×/2.0×`. These are distinct values — do not confuse them.

## In Scope

- Implement the canonical reward formula in Currency L1 Scala validation (Euclid SDK):
  ```
  base_rate = 0.1 ECHO
  trust_multiplier = tier_reward_multiplier(trust_tier)  // 1.0× – 3.0×
  volume_decay = 1.0 - (0.01 × max(0, messages_today - 100))

  effective_rate = base_rate × trust_multiplier × max(0.01, volume_decay)
  ```
- First 100 messages/day: full rate; messages 101+: 1% decay per additional message; minimum rate 0.01× (never zero)
- VIP subscriber modifier: 50% higher effective rate at each volume level (`AllowSpend`-verified subscription flag)
- Network-level activity tracking in Data L1 state tree (no per-user daily cap); reset at midnight UTC
- AtomicAction bundle: `[verifyTrustTier(did), computeDecayRate(did, messagesCount), creditReward(did, amount), recordNetworkActivity(did)]`
- Go backend Rewards Service pre-calculation: estimate decay rate client-side before submitting AtomicAction (reduce failed submissions)
- Update `POST /v1/rewards/calculate` API to return decay-adjusted amounts with tier multiplier breakdown

## Out of Scope

- Hard daily cap enforcement (superseded by this work order)
- Referral and staking rewards (use separate mechanisms, unchanged)
- VIP subscription billing (handled by `AllowSpend` pipeline, WO-238)

## Requirements

From the ECHO Token Economics blueprint and corrected by the combined Frontend blueprint:

**Reward Tier Multipliers (REWARDS — distinct from governance):**
| Trust Tier | Reward Multiplier | Governance Multiplier |
|---|---|---|
| Tier 1 (Unverified) | 1.0× | 0.0× |
| Tier 2 (Newcomer) | 1.2× | 0.5× |
| Tier 3 (Member) | 1.5× | 1.0× |
| Tier 4 (Verified) | 2.0× | 1.5× |
| Tier 5 (Trusted) | 3.0× | 2.0× |

**Reward Formula:**
```plaintext
base_rate = 0.1 ECHO
trust_multiplier = tier_reward_multiplier(trust_tier)   // 1.0× – 3.0×
volume_decay = 1.0 - (0.01 × max(0, messages_today - 100))
effective_rate = base_rate × trust_multiplier × max(0.01, volume_decay)
```

**AtomicAction (updated — uses `recordNetworkActivity`, not `updateDailyCap`):**
```go
AtomicActionBundle{
    Actions: []AtomicAction{
        {Type: "verify_trust_tier", DID: did},
        {Type: "compute_decay_rate", DID: did, MessagesCount: count},
        {Type: "credit_reward", DID: did, Amount: rewardAmount},
        {Type: "record_network_activity", DID: did},  // NOT updateDailyCap
    },
}
```

**Reward Types and Rates:**
| Reward | Base Rate | Trust Requirement | Mechanism |
|---|---|---|---|
| Messaging | 0.1 ECHO/msg (with volume decay) | Tier 2+ | AtomicAction |
| Payment rail | 1–5 ECHO/transaction | Tier 3+ | AtomicAction |
| Referral | 50 ECHO each | Tier 2+ both | AtomicAction on verification |
| Staking APY | 5–15% | Tier 2+ | TokenLock + StakeDelegation |

## Blueprints

- ECHO Token Economics and Founder Allocation — Defines the auto-scaling reward model, volume decay formula, and AtomicAction enforcement on Currency L1
- Frontend — Clarifies that reward multipliers (1.0–3.0×) are distinct from governance multipliers (0.0–2.0×), and specifies `recordNetworkActivity` in AtomicAction reward claims

---

### WO-214: Deploy Genesis Block Token Allocation and Founder Vesting

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation

## Summary

Set up the ECHO genesis snapshot on the Constellation Hypergraph mainnet with all five allocation pools minted and founder `TokenLock` positions created at genesis. Total fixed supply: 1,000,000,000 ECHO — no minting after genesis. Founder vesting is enforced by Currency L1 Scala code, not legal agreement.

## In Scope

- Genesis snapshot creation with five allocation pools:
  - Community Rewards: 400,000,000 ECHO (emission account, 10-year declining curve)
  - Treasury: 220,000,000 ECHO (3-of-5 multi-sig → DAO Phase 4+)
  - Founders: 180,000,000 ECHO (via TokenLock positions)
  - Future Team & Advisors: 100,000,000 ECHO (multi-sig controlled)
  - Ecosystem & Partnerships: 100,000,000 ECHO (governance-controlled)
- Founder `TokenLock` positions (Currency L1 enforced): 1-year cliff, 1/36th monthly vest over 36 months
  - Founder 1 (CEO): `TokenLock(100,000,000 ECHO, cliff=12mo, vest=48mo)`
  - Founders 2–5: `TokenLock(20,000,000 ECHO each, cliff=12mo, vest=48mo)`
- Treasury sub-allocation at genesis: 80M → PacaSwap liquidity seeding, 50M → operational reserve (stablecoins), 90M → treasury multi-sig
- Emission curve enforcement: Community Rewards release on declining schedule (Year 1: 80M, Year 2: 64M, Year 3: 52M, Year 4: 44M, Year 5: 36M, Year 6: 28M, Years 7–10: 24M/yr)
- All founder TokenLock positions publicly auditable on DAG Explorer
- Currency L1 Scala validation: reject any minting transaction after genesis

## Out of Scope

- PacaSwap pool seeding execution (WO for Treasury Management)
- Day-to-day treasury operations (separate treasury management work order)
- Governance system for ecosystem pool disbursement (governance work orders)

## Requirements

From the ECHO Token Economics and Founder Allocation blueprint:

**Genesis Block Structure:**
```plaintext
Genesis Block (Snapshot #1)
├── Community Rewards Pool  (400,000,000 ECHO) — emission account
├── Treasury                (220,000,000 ECHO)
│   ├── 80M → PacaSwap liquidity seeding
│   ├── 50M → Operational reserve (stablecoins via bridge)
│   └── 90M → Treasury multi-sig (3-of-5 founders → DAO)
├── Founders                (180,000,000 ECHO)
│   ├── Founder 1 → TokenLock(100M, cliff=12mo, vest=48mo)
│   └── Founders 2–5 → TokenLock(20M each, cliff=12mo, vest=48mo)
├── Future Team Pool        (100,000,000 ECHO) — multi-sig controlled
└── Ecosystem Pool          (100,000,000 ECHO) — governance-controlled
```

**Security Principle:** Fixed supply — no admin key can mint tokens after genesis. Founder vesting enforced by Currency L1 Scala code.

## Blueprints

- ECHO Token Economics and Founder Allocation — Defines genesis block structure, allocation amounts, founder vesting schedule, emission curve, and fixed-supply enforcement

---

### WO-215: Build Treasury Management and PacaSwap Liquidity Pool Seeding

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation

## Summary

Implement the treasury multi-sig management system and execute the initial PacaSwap liquidity pool seeding at token launch. The treasury holds 220M ECHO with a 3-of-5 multi-sig (founders transitioning to DAO in Phase 4). At genesis, 80M ECHO seeds the ECHO/DAG pool; Phase 3 adds the ECHO/USDC pool.

## In Scope

- Treasury 3-of-5 multi-sig wallet setup on Constellation Currency L1
- Phase 2: Seed ECHO/DAG pool with 80M ECHO from treasury + equivalent DAG at genesis market rate
- Phase 3: Seed ECHO/USDC pool from operational reserve (stablecoin acquired via Base bridge)
- LP token receipt and management: treasury holds LP tokens from initial seeding
- Ecosystem pool governance integration: 20M ECHO for LP mining incentives over 3 years
- Treasury balance monitoring: track sub-allocations (liquidity: 80M, operational: 50M, multi-sig reserve: 90M)
- Multi-sig transaction approval workflow: propose → 3-of-5 founder signatures → execute
- DAG operational reserve management: maintain DAG for snapshot fees (`FeeTransaction`)

## Out of Scope

- PacaSwap DEX smart contract development (Constellation ecosystem project)
- AI Burn Agent (separate Phase 5 work order)
- Governance system for DAO transition (governance work orders)

## Requirements

From the ECHO Token Economics and Founder Allocation blueprint:

**DEX and Liquidity:**
- ECHO/DAG pool (Phase 2): Primary trading pair; seeded from treasury at genesis
- ECHO/USDC pool (Phase 3): Stablecoin on/off ramp for users and treasury
- Both use constant product AMM (x×y=k), 0.3% swap fees to liquidity providers
- 20M ECHO ecosystem pool funds LP mining incentives over 3 years
- Base bridge (Phase 3): Aerodrome DeFi, treasury BTC accumulation path
- Ink bridge (Phase 4): Kraken exchange access, CEX liquidity

## Blueprints

- ECHO Token Economics and Founder Allocation — Defines treasury allocation, PacaSwap liquidity seeding, DEX pool parameters, ecosystem pool LP mining, and multi-sig governance
- Data Layer — Specifies `AllowSpend` and `SpendTransaction` v3 primitives for treasury operations

---

### WO-225: Implement Founder Departure Revocation with 3-of-5 Multi-Sig

**Type:** Build

**Blueprint:** ECHO Tokenomics, Founder Allocation, and Token Launch

## Summary

Implement FR3 from the ECHO Tokenomics blueprint: a 3-of-5 founder multi-sig mechanism that can trigger partial or full revocation of any founder's TokenLock position. Revoked tokens flow to the Future Team pool (not to the revoking founders). Enforced by Currency L1 Scala validation.

## In Scope

- **3-of-5 multi-sig revocation transaction:** New Tessellation v3 `AtomicAction` bundle type: `[collectSignatures(3-of-5 founder DIDs), revokeTokenLock(targetFounderDID, amount), creditsPool(futureTeamPool, amount)]`
- **Currency L1 Scala validation:** Validate that revocation transaction contains exactly 3 valid founder DID signatures; verify `amount ≤ remainingLockedBalance(targetDID)`; verify revoked tokens credit Future Team pool (not any founder wallet)
- **Backend multi-sig coordinator:** `POST /admin/founder-revocation/initiate` — collects signatures from 3 founders over 24-hour window; submits AtomicAction when threshold reached; `GET /admin/founder-revocation/status/{id}` — tracks signature collection progress
- **Revocation event on-chain:** Immutable record of revocation: `{targetFounderDID, revokedAmount, revokerDIDs[3], timestamp, txHash}` — publicly auditable on DAG Explorer
- **iOS Wallet display update:** Revoked founders see updated vesting panel showing original allocation, current locked balance, and revocation history

## Out of Scope

- Founder departure legal process (not a code work order)
- Future Team pool disbursement (governed by multi-sig, separate release)
- WO-214 genesis deployment (must precede this WO)

## Requirements

From the ECHO Tokenomics, Founder Allocation, and Token Launch blueprint:

**FR3 — Departure revocation:** A 3-of-5 founder multi-sig may trigger partial or full TokenLock revocation. Revoked tokens return to the Future Team pool, not to revoking founders.

**Departure Revocation Invariants:**
- Requires 3-of-5 founder DID signatures (any 3 of the 5 founders)
- Revoked amount ≤ remaining locked balance of target founder
- Revoked tokens → Future Team Pool (not to triggering founders)
- Transaction is publicly auditable on DAG Explorer
- Revocation can be partial (e.g., revoke 50% of position)

## Blueprints

- ECHO Tokenomics, Founder Allocation, and Token Launch — Defines FR3 departure revocation, 3-of-5 multi-sig requirement, and Future Team pool routing of revoked tokens

---

### WO-226: Build Founder Vesting Dashboard Panel in iOS Wallet

**Type:** Build

**Blueprint:** ECHO Tokenomics, Founder Allocation, and Token Launch

## Summary

Build the iOS Wallet tab founder vesting panel (visible only to founder DIDs) per FR9 of the ECHO Tokenomics blueprint. Displays allocated, vested, locked, and withdrawable ECHO amounts; next unlock date; cliff completion status; monthly vest schedule; and a direct link to the DAG Explorer public position. This panel is the on-chain cap table made visible to founders in their wallet.

## In Scope

- **Founder detection:** On wallet load, check if authenticated DID matches any of the 5 founder `TokenLock` positions from genesis; show `FounderVestingSection` only if match found
- **Vesting panel data (from Stargazer SDK):**
  - `totalAllocated`: original TokenLock amount (100M or 20M)
  - `vested`: amount unlocked to date (after 12-month cliff)
  - `locked`: amount still in TokenLock
  - `nextUnlockAmount`: 1/36th of remaining locked (monthly)
  - `nextUnlockDate`: next monthly unlock date
  - `cliffCompleted`: Bool (true after 12 months from genesis)
  - `cliffDate`: genesis + 12 months
  - `withdrawable`: vested amount not yet withdrawn (requires 14-day `WithdrawLock` cooldown)
  - `vestingProgress`: Float 0.0–1.0 for progress bar
- **Withdraw action:** Tap "Withdraw Vested Tokens" → enter amount (≤ withdrawable) → biometric confirm → `stargazer.submitWithdrawLock(WithdrawLockRequest(amount: amount))` — 14-day cooldown shown in UI
- **DAG Explorer link:** "View on DAG Explorer" → deep link to `https://dagexplorer.io/address/{founderDID}` showing public TokenLock position
- **`GET /tokens/vesting`** backend endpoint: authenticated founders only; returns vesting position from Stargazer SDK cache (TTL: 30s)

## Out of Scope

- Genesis TokenLock creation (WO-214)
- Withdrawal cooldown enforcement (Currency L1, not UI)
- Non-founder wallet view (WO-95 covers standard balance display)

## Requirements

From the ECHO Tokenomics blueprint:

**FR9 — Public founder visibility:** All founder TokenLock positions (allocated, vested, locked, cliff status, withdrawal history) are publicly visible on DAG Explorer and queryable via the ECHO Wallet founder vesting display.

**`VestingInfo` struct (from Frontend blueprint):**
```swift
struct VestingInfo {
    let totalAllocated: Decimal
    let vested: Decimal
    let locked: Decimal
    let nextUnlockAmount: Decimal
    let nextUnlockDate: Date
    let cliffCompleted: Bool
    let cliffDate: Date
    let withdrawable: Decimal
    let vestingProgress: Float    // 0.0 to 1.0
}
```

## Blueprints

- ECHO Tokenomics, Founder Allocation, and Token Launch — Defines FR9 public founder visibility requirement and vesting schedule details
- Frontend — Defines `VestingInfo` struct, `FounderVestingSection` view, and Stargazer SDK `WithdrawLock` submission

---

### WO-231: Deploy Phase 2 Constellation Mainnet Node Infrastructure

**Type:** Build

**Blueprint:** Production Launch, Infrastructure, and Deployment

## Summary

Deploy the Phase 2 Constellation Hypergraph mainnet infrastructure: 3 L0 hybrid nodes (AWS m5.2xlarge, 250K DAG staked each), 3 Currency L1 validators, 3 Data L1 validators, ECHO token live on Hypergraph, and the PacaSwap ECHO/DAG liquidity pool seeded. This is the blockchain infrastructure gate that enables ECHO token transactions on mainnet.

## In Scope

- **3 L0 Hybrid Nodes (AWS us-east-1):**
  - Instance: `m5.2xlarge` (8 vCPU, 32GB RAM, 500GB SSD)
  - OS: Ubuntu 22.04 LTS
  - Processes: `global_l0` + `metagraph_l0` running on each node
  - DAG staking: 250,000 DAG per node = 750,000 DAG total
  - Node registration on public Hypergraph mainnet
- **3 Currency L1 Validators + 3 Data L1 Validators:**
  - Project-operated for Phase 1–3
  - ECHO stake requirement: governance-set before mainnet launch
  - Euclid SDK Scala validation code deployed: annual emission caps, auto-scaling reward rate, anti-gaming, Merkle root validation, trust commitments
- **DAG acquisition and staking:** Platform acquires 750,000 DAG and stakes to L0 nodes. Documentation of staking addresses for public auditability
- **ECHO token genesis:**
  - Execute WO-214 (genesis deployment) against mainnet
  - ECHO token visible in Stargazer wallet on mainnet
  - ECHO balance queryable on DAG Explorer
- **PacaSwap ECHO/DAG pool seeding:** Coordinate with WO-215 to execute pool seeding from treasury

## Out of Scope

- Application deployment to Kubernetes (separate WO)
- Community validator onboarding (Phase 4)
- Multi-region redundancy (Phase 4)

## Requirements

From the Production Launch, Infrastructure, and Deployment blueprint:

**Phase 2 node configuration:**
```yaml
l0_nodes:
  count: 3
  dag_staking: 250000  # per node (750K total)
  instance_type: m5.2xlarge
  storage: 500GB SSD
  os: Ubuntu 22.04 LTS
  processes_per_node: [global_l0, metagraph_l0]

currency_l1_validators:
  count: 3
  operator: project_operated_phase_1_3

data_l1_validators:
  count: 3
  operator: project_operated_phase_1_3
```

**Phase 2 Go/No-Go:** Metagraph mainnet finality < 10s (P95); ECHO token visible on Stargazer and DAG Explorer; PacaSwap ECHO/DAG pool seeded.

## Blueprints

- Production Launch, Infrastructure, and Deployment — Defines Phase 2 Constellation mainnet deployment requirements, node infrastructure specs, DAG staking, and go/no-go criteria

---

### WO-232: Configure Kubernetes Auto-Scaling, CI/CD Pipeline, and Production Database Setup

**Type:** Build

**Blueprint:** Production Launch, Infrastructure, and Deployment

## Summary

Configure Kubernetes on AWS EKS with Horizontal Pod Autoscalers for all 10 Go backend services, set up NATS JetStream cluster, deploy PostgreSQL + Redis with synchronous replication, and establish the full CI/CD deployment pipeline from GitHub to production. This is the application infrastructure layer on top of the mainnet node deployment (WO-231).

## In Scope

- **EKS cluster:** AWS us-east-1, minimum 3 nodes, auto-scaling node group (3–20 nodes), `m5.xlarge` instances
- **HPA for all 10 services:** Per blueprint configuration for Message Relay — `minReplicas: 3`, `maxReplicas: 50`, scale on CPU > 70%, WebSocket connections > 10K/pod, relay latency P99 > 200ms. Similar HPA for other services
- **NATS JetStream:** 3-node NATS cluster for cross-service pub/sub; group message fan-out
- **PostgreSQL 15+:** Primary + 2 replicas (synchronous replication); automated failover; `pgBouncer` connection pooling
- **Redis 7+ with AOF:** Primary + 2 replicas; AOF persistence enabled; `maxmemory-policy: allkeys-lru`
- **Secrets management:** AWS Secrets Manager for DB credentials, API keys, JWT secrets; mounted as Kubernetes secrets via ExternalSecrets operator
- **CI/CD pipeline:** GitHub Actions → build Docker images → push to ECR → rolling deployment to EKS; zero-downtime deployments with `RollingUpdate` strategy
- **Container images:** Multi-stage Dockerfiles for Go services (final stage: `gcr.io/distroless/base`); iOS builds via Xcode Cloud

## Out of Scope

- Prometheus/Grafana monitoring (WO-210)
- Metagraph node infrastructure (WO-231)
- Multi-region Phase 4 setup

## Requirements

From the Production Launch, Infrastructure, and Deployment blueprint:

**HPA configuration example (Message Relay):**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: message-relay-hpa
spec:
  minReplicas: 3
  maxReplicas: 50
  metrics:
  - type: Resource
    resource: { name: cpu, target: { averageUtilization: 70 } }
  - type: Pods
    pods: { metric: { name: websocket_connections }, target: { averageValue: "10000" } }
  - type: Pods
    pods: { metric: { name: relay_latency_p99_ms }, target: { averageValue: "200" } }
```

## Blueprints

- Production Launch, Infrastructure, and Deployment — Defines Kubernetes infrastructure sizing, HPA configuration, NATS/Redis/PostgreSQL specs, and CI/CD requirements

---

### WO-271: Implement Token Quest System with AtomicAction Reward Claiming

**Type:** Build

**Blueprint:** ECHO Token Economics and Founder Allocation, ECHO Tokenomics, Founder Allocation, and Token Launch

## Summary

Implement the ECHO quest system — short-term, structured onboarding activities that guide new users into ECHO's features while rewarding them with ECHO tokens and badges. Quests are feature-exploration milestones (identity setup, group joining, staking, referrals, feature discovery), not message-volume incentives. The quest catalog lives on the backend with completion state tracked on the Data L1. All reward claims go through AtomicAction. Quest tracking is pre-genesis; ECHO payouts are Phase 3+ conditional on token genesis.

## In Scope

- **Quest catalog API:**
  - `GET /gamification/quests` — returns full quest catalog with per-quest `{questId, title, description, action, requiredCount, reward_echo, badge, completedAt, rewardClaimed}` for the authenticated DID; completion status read from `Quest completion` Data L1 state
  - Response distinguishes `starter` vs `advanced` quest tiers
- **Quest completion tracking:**
  - `Quest completion` entity on Data L1: `{did, quest_id, completed_at, reward_claimed}` — on-chain proof of quest completion
  - Backend event hooks evaluate quest completion conditions on relevant user actions (identity verification, group join, staking, etc.)
- **Quest reward claiming:**
  - `POST /gamification/quests/:questId/claim` — submits AtomicAction bundle: `{verifyQuestCompletion(did, questId), claimQuestReward(did, echoAmount), awardBadge(did, badgeId), anchorDataL1(questCompletionEvent)}` — all-or-nothing
  - Idempotent: once `reward_claimed: true` is set on the Data L1 completion record, re-submission returns `HTTP 409: already_claimed`
  - Pre-genesis: badge is awarded and completion is anchored on-chain; ECHO reward is queued and credited at genesis
- **Starter quest catalog (Phase 2 launch):**

  | Quest ID | Action | Reward | Badge |
  |---|---|---|---|
  | identity_builder | Complete identity verification (earn first Verifiable Credential) | 20 ECHO | "Verified" |
  | community_joiner | Join or create a group with 5+ members | 10 ECHO | "Group Member" |
  | trusted_messenger | Reach Trust Tier 3 | 50 ECHO | "Trusted" |
  | stack_and_earn | Stake ECHO for the first time | 15 ECHO | "Staker" |
  | invite_and_grow | Complete first successful referral | 25 ECHO | "Connector" |
  | vault_keeper | Send a disappearing message | 5 ECHO | "Ghost" |
  | private_circle | Activate a Hidden Folder | 10 ECHO | "Vault" |
  | vip_experience | Upgrade to VIP for first month | 50 ECHO cashback | "VIP" |
  | governance_debut | Cast first governance vote | 25 ECHO | "Voter" |

- **Advanced quest catalog (Phase 3+):**

  | Quest ID | Action | Reward |
  |---|---|---|
  | network_validator | Delegate to a validator for 30 days | 100 ECHO + "Delegator" badge |
  | whale_staker | Stake Platinum tier (365 days) | 500 ECHO + animated badge |
  | network_builder | Refer 10 active users | 500 ECHO + "Builder" badge |
  | bot_creator | Publish a bot to the marketplace | 200 ECHO + revenue share activation |

- **Anti-gaming enforcement:**
  - DID uniqueness: each DID claims any quest reward exactly once
  - Trust tier gates: governance_debut requires Tier 2+ (validated by AtomicAction)
  - All claims are AtomicActions — no partial reward state
- **iOS gamification section** in the Profile tab: quest catalog with progress indicators, claimed badges, next available quest call-to-action, "Claim" button for completed unclaimed quests

## Out of Scope

- Message-volume rewards or earn-by-chatting mechanics (removed per requirements v3.0)
- Streak system (WO-270 deleted — streak multipliers on messaging rewards are removed)
- Achievement milestone system (WO-116 — long-term achievements)
- Base reward auto-scaling (WO-213)

## Requirements

From ECHO Tokenomics blueprint — Gamification Strategy, Mechanic 3: Quest System. Note: quests tied purely to message count (first_contact, club_1k) removed per requirements v3.0 "No earn-by-chatting mechanic."

**Quest definition:** Short-term, structured activities that onboard users into deeper ECHO features.

**Data model:**
`Quest completion | Data L1 state | did, quest_id, completed_at, reward_claimed`

**API:**
- `GET /gamification/quests` — Returns quest catalog with user completion status
- `POST /gamification/quests/:questId/claim` — Claims quest reward via AtomicAction

**Anti-gaming:**
- DID uniqueness: each DID receives quest rewards once only
- AtomicAction enforcement: all claims are atomic — no partial or double claims
- Trust tier gates for governance and staking quests

**Success metric:** >60% of new users complete at least 3 quests within 7 days

## Blueprints

- ECHO Tokenomics, Founder Allocation, and Token Launch — Defines quest catalog (Mechanic 3), Data L1 entity, API endpoints, and anti-gaming framework
- ECHO Token Economics and Founder Allocation (Foundation) — Defines AtomicAction primitive and on-chain reward claiming infrastructure

---
