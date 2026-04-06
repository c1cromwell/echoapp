# Data Layer Architecture — v3.0

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 3.0 | February 23, 2026 | Resolved messaging transport: client-server relay with decentralized anchoring. Added message relay layer as core component. Updated all data flows. Added offline message queuing. Added metadata protection roadmap. Added group messaging architecture. Removed P2P/libp2p from open questions. Updated open questions list. |
| 2.0 | February 23, 2026 | Cross-document review. Added cross-chain consistency, performance targets, Cardano spec, log lifecycle, fault tolerance, anti-spam, governance, encryption spec. |
| 1.0 | February 2026 | Initial blueprint |

---

## 1. Overview

The data layer is built on a multi-blockchain architecture combining Constellation metagraph infrastructure for application data and ECHO rewards, Cardano for identity and verifiable credentials, and decentralized storage for logging and audit trails. A Go backend microservices layer acts as an operational coordinator, stateless message relay, and hot cache between client applications and on-chain state.

**Design Principles:**

- On-chain as source of truth; off-chain as performance optimization
- Zero PII on any blockchain (enforced by T0–T7 data classification)
- Each chain handles one concern: identity (Cardano), application state + rewards (metagraph), audit (IPFS/Storj)
- The Go backend is a relay and cache, not an authority — it cannot read message content, does not own user identities, and does not control token balances
- Message relay is client-server for reliability; decentralization comes from identity, data integrity, and encryption layers

---

## 2. Core Components

### 2.1 Constellation Metagraph (Application & Rewards Layer)

The primary data layer uses Constellation Network's metagraph architecture for high-throughput, decentralized consensus on application data and token transactions.

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
| Staking operations | DID, amount, tier, lock duration | T7 (public) | Minimum amounts per tier, valid lock period |
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
| Data L1 schema changes | Governance proposal → vote (stake-weighted) → activation at future snapshot height | Phase 4 |
| Metagraph node admission | Initially permissioned (project-operated nodes); transition to permissionless with stake requirement | Phase 4 |
| Relay node admission | Initially project-operated; Phase 4: any staked operator can run a relay node registered on Data L1 | Phase 4 |
| Cardano contract upgrades | Parameterized Plutus scripts with upgrade key (multi-sig); transition to DAO-controlled | Phase 4 |
| Protocol versioning | Semantic versioning; clients must support current and current-1 | Ongoing |
| Emergency upgrades | Multi-sig (3-of-5 core team) can push critical fixes without governance vote | All phases |

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
| Content-blind relay | Relay servers transport E2E encrypted blobs; cannot read, modify, or forge messages |
| Client-verified authenticity | Recipients verify sender signatures and commitment hashes locally; no relay trust needed |
| Encrypted storage | Logs encrypted before IPFS/Storj; local data encrypted with derived keys; offline queue stores only ciphertext |
| Immutable audit trail | Merkle roots on Data L1; log CIDs on Data L1; Cardano credential history |
| Separation of concerns | Identity (Cardano), app state + rewards (metagraph), audit (IPFS/Storj), transport (relay) |
| Device-local secrets | Passkeys and private keys in iOS Secure Enclave; never extractable |
| Zero PII on-chain | Enforced by T0–T7 data classification; Data L1 validators reject prohibited types |
| Forward secrecy | Ephemeral X25519 keys per message session |
| Anti-spam | Multi-layer: API rate limits, per-DID message rate, per-DID daily reward caps, economic micro-fees at scale |
| Graceful degradation | Message relay operates independently of chain state; blockchain outages don't stop messaging |
| Progressive metadata protection | Phase 1: baseline; Phase 3: sealed sender; Phase 4: federated relay + optional P2P |

---

## 10. Service Integrations & Dependencies

### 10.1 External Service Integrations

| Service | Purpose | Integration Point | Failure Handling | Phase |
|---------|---------|-------------------|-----------------|-------|
| **IDV Provider** (Stripe Identity / Sumsub / Jumio) | Identity verification for Tier 3+ | POST /identity/verify callback | Queue webhook; retry on recovery | 2 |
| **APNs** (Apple Push Notification) | Push notifications for offline messages | Go backend → APNs API | Silent fallback (relay queue + VoIP push); retry queue | 1 |
| **Constellation Network** | Metagraph consensus layer | Direct node connection via gRPC | Circuit breaker; batching continues; drains on recovery | 1 |
| **Cardano Mainnet** | Credential registry + trust tier issuance | Direct node connection via Kupo / Blockfrost | Cached credentials; retry on recovery; CDN-based indexing | 2 |
| **IPFS (Pinata / web3.storage)** | Log pinning + decentralized storage | HTTP API; encrypted blob submission | Local buffering; fallback to secondary pin provider; Storj overflow | 2 |
| **Storj DCS** | Fallback cold storage for large audit trails | S3-compatible API | Transparent failover from IPFS; separate cost tracking | 2 |
| **Twilio / Firebase** | SMS for multi-device verification (optional) | Callback for 2FA codes | Backend sends own codes via secure channel | 3 |
| **Sendgrid / Postmark** | Email notifications (Phase 3+) | Transactional email API | Queue; retry with exponential backoff | 3 |

### 10.2 Service Health Monitoring

**Backend Health Endpoints:**
```
GET /health/status → full service status
GET /health/chains → blockchain connectivity status
GET /health/cache → Redis + PostgreSQL status
GET /health/relay → message relay pod status (per region)
GET /health/storage → IPFS/Storj connectivity
```

**Monitoring Metrics (Prometheus):**
- Message relay latency (p50, p99)
- Offline queue depth (per recipient, global)
- Rate limit hit rate (by tier)
- Blockchain submission latency (per chain)
- Cache hit rate (balance, trust scores)
- Circuit breaker state transitions
- IPFS upload latency + success rate
- IDV verification latency + approval rate

### 10.3 API Version Management

**Client Request Header:**
```
Accept-Version: application/vnd.echo.v1+json
X-Client-Version: 1.0.0
X-Device-Id: UUID
```

**Backend Versioning Strategy:**
- Current version: `v1` (OpenAPI 3.1.0)
- Supported versions: `v1` only (Phase 1–2)
- Deprecated versions: None (Phase 1 is initial)
- Breaking change process: Add new version path (e.g., `/v2/...`) alongside `/v1/...` for 1 release cycle

---

## 11. Missing Features & Roadmap Gaps

### 11.1 Features in OpenAPI Spec Not Yet Documented

| Feature | Endpoint | Data Flow | Status |
|---------|----------|-----------|--------|
| **Group Management** | `/groups/{groupId}/members`, `/groups/{groupId}/members/{did}`, `/groups/{groupId}/key` | Create/add/remove members; rotate group key | SPEC ✅ / ARCH ⚠️ (outlined in 3.2, needs full detail) |
| **Message Reactions** | POST `/messages/{messageId}/reactions` | Add/remove emoji reactions to messages | SPEC ✅ / ARCH ⚠️ (implied but not detailed) |
| **Achievement System** | GET `/tokens/achievements` | Fetch user achievement badges (public metadata) | SPEC ✅ / ARCH ❌ (missing from architecture) |
| **Disappearing Messages** | Message model `expiresIn`, `expiresAt` | Auto-delete after TTL; optional feature | SPEC ✅ / ARCH ⚠️ (mentioned in question 9, not fully detailed) |
| **Sealed Sender** | Message model with nested encryption | Hide sender DID from relay server | SPEC ✅ / ARCH ⚠️ (Phase 3 roadmap item, not Phase 1) |
| **WebSocket Real-Time** | WebSocket `/ws` | Live message delivery + confirmation stream | SPEC ✅ / ARCH ⚠️ (referenced but not detailed) |
| **Referral System** | Implied in rewards | Earning via referral codes | SPEC ⚠️ / ARCH ❌ (missing from both) |
| **Staking Tiers** | `/tokens/stake`, `/tokens/unstake` | Variable APY by lock period | SPEC ✅ / ARCH ✅ |

### 11.2 New Features to Add to Architecture

**Achievement System:**
- Backend queries Data L1 for user activity metrics (messages sent, contacts, trust verified, etc.)
- Pre-defined achievement definitions (e.g., "Send 100 messages" → 10 ECHO badge)
- Achievements stored as Data L1 commitments (public metadata, no PII)
- Client fetches via GET /tokens/achievements, displays in profile

**Disappearing Messages (Phase 3):**
- Message `expiresIn` (seconds) or `expiresAt` (timestamp) sent with message
- Client retains plaintext for TTL; auto-deletes after expiry
- Relay server deletes ciphertext blob after TTL (or at first read if ephemeral)
- Merkle root on Data L1 is permanent but individual commitments become unverifiable after key deletion (per question 9)

**WebSocket Real-Time Delivery:**
- Connection: `wss://echo.api/ws?token=JWT`
- Auth: Bearer token validated per message; connection dies on token expiry
- Message events: `{type: 'message.new', payload: EncryptedMessage}`
- Delivery ACK: `{type: 'message.ack', messageId: '...', status: 'delivered'}`
- Confirmation: `{type: 'message.confirmed', messageId: '...', snapshotHeight: N}`
- User presence: `{type: 'presence.update', did: '...', isOnline: bool}` (opt-in per user)
- Connection limit: 1 active WebSocket per device (new connection closes prior)

**Referral System:**
- User generates unique referral code: `echo://ref/ABC123XYZ`
- Referred user enters code during registration
- On referred user reaching Tier 2 (email verified), referrer earns 5 ECHO bonus
- Capped at 50 referrals per user (mitigates gaming)
- Data L1 tracks referral chain for anti-fraud analysis

---

## 12. Comprehensive Test Coverage Specification

### 12.1 Unit Tests (Go Backend Services)

**AuthService:**
```go
TestRegisterNewUser()
  ✓ Valid registration creates DID + Cardano wallet
  ✓ Duplicate email rejected
  ✓ Invalid publicKey format rejected
  ✓ DeviceInfo recorded correctly

TestRequestChallenge()
  ✓ Valid DID generates nonce + hash
  ✓ Challenge expires after 5 minutes
  ✓ Unregistered DID rejected
  ✓ Challenge replay protection (per challenge ID)

TestVerifyChallenge()
  ✓ Valid signature + nonce accepted
  ✓ Invalid signature rejected
  ✓ Expired challenge rejected
  ✓ Returns JWT tokens + refresh token

TestRefreshToken()
  ✓ Valid refresh token extends session
  ✓ Expired refresh token rejected
  ✓ Access token regenerated correctly

TestLogout()
  ✓ Token added to revocation list
  ✓ Subsequent requests with token rejected
  ✓ Refresh token invalidated
```

**MessageService:**
```go
TestSendMessage()
  ✓ Valid message with signature accepted
  ✓ Message queued to offline recipient
  ✓ Rate limit enforced (60/min for Tier 1)
  ✓ Commitment added to batch
  ✓ APNs notification sent to offline recipient

TestFetchMessages()
  ✓ Cursor-based pagination works
  ✓ Limit parameter enforced (max 100)
  ✓ Messages ordered by timestamp DESC
  ✓ Unauthorized recipient returns 403

TestMarkAsRead()
  ✓ Message status updated to 'read'
  ✓ Timestamp recorded
  ✓ Batch submission to Data L1

TestAddReaction()
  ✓ Valid emoji accepted
  ✓ Duplicate emoji prevented
  ✓ Reaction count incremented
  ✓ Sender notified via WebSocket

TestDeleteMessage()
  ✓ Only sender can delete (within 5 minutes)
  ✓ Ciphertext removed from offline queue
  ✓ Data L1 commitment persists (immutable)
```

**RewardService:**
```go
TestClaimReward()
  ✓ Claim type validated
  ✓ Daily cap enforced per user + type
  ✓ Trust tier multiplier applied
  ✓ Claim batched to Currency L1
  ✓ Duplicate claim rejected (idempotent)

TestGetBalance()
  ✓ Cached balance returned within TTL
  ✓ Cache miss queries Currency L1
  ✓ Stale cache returned if L1 down

TestStakeTokens()
  ✓ Minimum stake amount enforced (100 ECHO)
  ✓ Lock period validated (30–1825 days)
  ✓ APY calculated correctly per tier
  ✓ Submitted to Currency L1
  ✓ Insufficient balance rejected

TestUnstakeTokens()
  ✓ Lock period enforced
  ✓ Unlock timestamp checked
  ✓ Rewards accrued calculated
  ✓ Balance updated on Currency L1
```

**IdentityService:**
```go
TestCreateDID()
  ✓ DID created on Cardano
  ✓ Public key published to DID Document
  ✓ Service endpoints set correctly
  ✓ Transaction fee paid from treasury

TestSubmitVerification()
  ✓ IDV callback triggers credential issuance
  ✓ Trust tier updated correctly
  ✓ Credential status bit vector updated
  ✓ Retry on IPFS failure

TestGetVerifications()
  ✓ User's verification records returned
  ✓ Status values correct (pending/approved/rejected)
  ✓ Timestamps accurate

TestVerifyCredential()
  ✓ Signature verified against issuer DID
  ✓ Bit vector status checked
  ✓ Expiry validated
  ✓ Revoked credential rejected
```

**ContactService:**
```go
TestAddContact()
  ✓ Valid DID added to contact list
  ✓ Duplicate prevented
  ✓ Self-contact rejected
  ✓ Blocked contact rejected (can unblock first)

TestRemoveContact()
  ✓ Contact deleted
  ✓ Conversations persist (soft delete)

TestBlockContact()
  ✓ Blocked contact hidden from list
  ✓ Messages from blocked contact filtered
  ✓ Block is mutual (both can unblock)

TestListContacts()
  ✓ Cursor pagination works
  ✓ Blocked contacts excluded
  ✓ Trust scores included
  ✓ Timestamp sorting
```

**MessageRelayService:**
```go
TestQueueOfflineMessage()
  ✓ Message queued when recipient offline
  ✓ Queue respects 30-day TTL
  ✓ Queue respects 1000-message limit
  ✓ Oldest message evicted on overflow
  ✓ APNs notification sent

TestFlushQueueOnConnect()
  ✓ Messages delivered in order (FIFO)
  ✓ Queue cleared after delivery
  ✓ Delivery status updated

TestRateLimiting()
  ✓ Per-DID send rate enforced (60 msgs/min)
  ✓ Tier 2+ get higher rate (120/min)
  ✓ HTTP 429 returned on limit exceeded
  ✓ Rate counter decays correctly

TestCircuitBreaker()
  ✓ Opens after 5 consecutive failures
  ✓ Rejects requests with 503 while open
  ✓ Closes after 60 seconds of success
  ✓ Half-open state allows single request
```

**CacheService:**
```go
TestBalanceCache()
  ✓ Balance cached for 5 seconds
  ✓ Cache miss queries Currency L1
  ✓ Cache invalidated on new snapshot
  ✓ Stale cache returned if L1 down

TestTrustScoreCache()
  ✓ Trust score cached for 60 seconds
  ✓ Tier updates invalidate cache
  ✓ Verification completion invalidates cache

TestCacheInvalidationOnChainUpdate()
  ✓ Metagraph snapshot event triggers invalidation
  ✓ Cardano confirmation event triggers invalidation
  ✓ Multiple cache keys invalidated correctly
```

### 12.2 Integration Tests (Service-to-Chain Communication)

**Metagraph Data L1 Integration:**
```go
TestMessageBatchSubmission()
  ✓ 1000 message commitments collected
  ✓ Merkle tree built correctly
  ✓ Merkle root submitted to Data L1
  ✓ L1 block finalized within 10 seconds
  ✓ Confirmation received and stored

TestTrustCommitmentSubmission()
  ✓ Trust tier update submitted
  ✓ Hash commitment calculated correctly
  ✓ Submitted with authoritative signature

TestDataL1ValidationRules()
  ✓ Invalid Merkle structure rejected
  ✓ Unauthorized sender rejected
  ✓ Duplicate commitments detected
  ✓ Schema version checked
```

**Cardano Credential Integration:**
```go
TestCredentialIssuance()
  ✓ Credential schema published to Cardano
  ✓ Credential issued with issuer signature
  ✓ Trust tier datum set on UTXO
  ✓ Expiry date recorded

TestCredentialRevocation()
  ✓ Bit vector updated
  ✓ Batch revocation batches multiple revocations
  ✓ Verifiers reject revoked credentials

TestDIDResolution()
  ✓ DID Document resolved from Cardano
  ✓ Public keys verified correctly
  ✓ Service endpoints parsed
  ✓ Caching works (TTL: 1 hour)

TestTrustTierQueryFromCardano()
  ✓ User's current tier fetched
  ✓ Expiry checked
  ✓ Multiple tier credentials handled
```

**IPFS/Storj Integration:**
```go
TestLogBatchEncryption()
  ✓ Batch encrypted with AES-256-GCM
  ✓ Encryption key derived from master key
  ✓ IV/nonce unique per batch

TestIPFSUpload()
  ✓ Encrypted batch uploaded
  ✓ CID returned and stored
  ✓ Retry on failure (exponential backoff)

TestStorjFallback()
  ✓ If IPFS unavailable, switches to Storj
  ✓ Both providers used simultaneously for redundancy

TestLogRetrieval()
  ✓ Authorized auditor can fetch batch
  ✓ Decryption succeeds with correct key
  ✓ Tampered batch fails HMAC verification
```

### 12.3 End-to-End Tests (iOS App ↔ Backend ↔ Blockchain)

**Authentication Flow (Real or Testnet):**
```
E2E_Auth_Registration()
  ✓ App generates P256 keypair in Secure Enclave
  ✓ Submits registration with public key
  ✓ Backend creates DID on Cardano
  ✓ App receives DID + service endpoints
  ✓ Local cache updated

E2E_Auth_Challenge_Response()
  ✓ App requests challenge for DID
  ✓ Receives nonce + hash
  ✓ Signs with Secure Enclave private key
  ✓ Submits signature
  ✓ Backend verifies signature
  ✓ Receives JWT access token + refresh token
  ✓ Token persists in Keychain

E2E_Auth_TokenRefresh()
  ✓ Access token expires (1 hour)
  ✓ App submits refresh token
  ✓ Backend validates refresh token
  ✓ Issues new access + refresh tokens
  ✓ Session continues without user re-auth
```

**Message Send & Delivery (Real or Testnet):**
```
E2E_Message_P2P_Online()
  ✓ Sender composes message
  ✓ Encrypts with Kinnami (X25519 + ChaCha20-Poly1305)
  ✓ Signs with Secure Enclave
  ✓ Sends via WebSocket to backend
  ✓ Backend validates + rate limits
  ✓ Recipient receives via WebSocket
  ✓ Decrypts + verifies signature
  ✓ Displays plaintext
  ✓ Sends read receipt
  ✓ Sender receives confirmation

E2E_Message_P2P_Offline()
  ✓ Sender sends message
  ✓ Recipient is offline
  ✓ Backend queues encrypted blob
  ✓ APNs notification sent
  ✓ Recipient launches app (or wakes via push)
  ✓ Establishes WebSocket connection
  ✓ Backend flushes queue
  ✓ Recipient receives 5 messages from queue (oldest first)
  ✓ Clears queue

E2E_Message_Group()
  ✓ Admin creates group with 50 members
  ✓ Group key distributed via encrypted 1:1 messages
  ✓ Admin sends group message
  ✓ 30 members online → receive immediately
  ✓ 20 members offline → queued
  ✓ Queued members receive on reconnect
  ✓ Message count metadata (not content) on Data L1
  ✓ No member list on-chain

E2E_Message_Disappearing()
  ✓ Sender sends message with expiresIn: 3600
  ✓ Recipient receives
  ✓ Message auto-deleted from recipient's device after 1 hour
  ✓ Relay server deletes ciphertext after 1 hour
  ✓ Merkle root persists on Data L1 (immutable)
```

**Reward Claim & Staking (Real or Testnet):**
```
E2E_Reward_Claim()
  ✓ User qualifies for "Send 10 messages" reward
  ✓ Backend verifies criteria met
  ✓ User clicks "Claim Reward"
  ✓ Backend submits to Currency L1
  ✓ Currency L1 validator processes
  ✓ Balance updated in metagraph snapshot
  ✓ Backend cache invalidated
  ✓ App receives balance update via WebSocket
  ✓ +5 ECHO displayed

E2E_Staking()
  ✓ User stakes 100 ECHO for 365 days
  ✓ Backend submits to Currency L1
  ✓ L1 validator confirms
  ✓ Staking UTXO locked with unlock_time = now + 365 days
  ✓ App shows "locked" status
  ✓ APY calculated: 100 ECHO × 8% / 365 = 0.022 ECHO/day
  ✓ After 365 days, unlock becomes available
  ✓ User unstakes, receives 108 ECHO (approx)
```

**Identity Verification (Testnet IDV):**
```
E2E_IDV_Stripe()
  ✓ User initiates identity verification
  ✓ App launches Stripe Identity flow (embedded)
  ✓ User captures ID document + selfie
  ✓ Stripe verifies & returns pass/fail
  ✓ Stripe sends webhook to backend
  ✓ Backend maps to user DID
  ✓ Cardano credential issued
  ✓ Trust tier updated to 3
  ✓ Backend cache invalidated
  ✓ App shows "Verified ✓"
  ✓ User can now create groups

E2E_TrustTierProgression()
  ✓ User registers (Tier 1)
  ✓ Verifies email (Tier 2)
  ✓ Completes IDV (Tier 3)
  ✓ Message reward rate: 0.01 ECHO/msg (Tier 1) → 0.02 (Tier 2) → 0.05 (Tier 3)
  ✓ Trust score multiplier: 1.0 → 1.2 → 1.5
```

### 12.4 Stress Tests (Load & Resilience)

**Message Throughput:**
```
StressTest_MessageRelay_1M_Users_100K_Online()
  Setup:
    - 1M total users registered
    - 100K concurrent WebSocket connections
    - 50 messages/sec from random senders
    - 10% offline recipients (queued)
    
  Assertions:
    ✓ P50 delivery latency: < 100 ms
    ✓ P99 delivery latency: < 500 ms
    ✓ Queue depth stable (not growing)
    ✓ No message loss
    ✓ Backend CPU: < 70%
    ✓ Redis memory: < 2 GB (10K messages queued)
    
  Cleanup:
    - Drain all queues
    - Verify all messages received
```

**Reward Submission Batching:**
```
StressTest_RewardBatching_10K_Claims_Per_Second()
  Setup:
    - 10,000 reward claims submitted simultaneously
    - Batch size: 100 claims per Data L1 transaction
    
  Assertions:
    ✓ All 10K claims batched into 100 transactions
    ✓ L1 block includes all 100 txns
    ✓ Metagraph snapshot finalizes within 10 seconds
    ✓ Backend cache invalidated correctly
    ✓ App receives confirmation for all claims
    
  Cost analysis:
    ✓ 100 txns × 0.001 ECHO = 0.1 ECHO per batch
    ✓ Cost-effective at scale
```

**Circuit Breaker Recovery:**
```
StressTest_CircuitBreaker_Chaindown_Recovery()
  Setup:
    - Message relay active (100 msg/sec)
    - Token balance queries active (50 req/sec)
    - Kill Metagraph connection
    
  During Outage:
    ✓ Message relay continues (independent of chain)
    ✓ Balance queries return cached data
    ✓ Circuit breaker opens after 5 failures
    ✓ New requests rejected with 503 (not hung)
    ✓ APNs push still works
    
  Recovery (30 seconds later):
    ✓ Metagraph reconnects
    ✓ Circuit breaker enters half-open state
    ✓ Single test request succeeds
    ✓ Circuit breaker closes
    ✓ All queued submissions drain
    ✓ No data loss
```

**Offline Queue Degradation:**
```
StressTest_OfflineQueueLimit()
  Setup:
    - Recipient offline for 1 hour
    - 100 msg/sec sent to recipient
    - Queue limit: 1000 messages
    
  Assertions:
    ✓ First 1000 messages queued
    ✓ Messages 1001+ rejected with 429 (quota exhausted)
    ✓ Sender receives rate limit error
    ✓ Oldest message evicted when new one added
    ✓ On recipient reconnect, receive latest 1000
    ✓ Missing older messages noted in UI
```

### 12.5 Security Tests

**Encryption Validation:**
```
SecurityTest_MessageEncryption()
  ✓ Plaintext never logged
  ✓ Ciphertext tampered → recipient rejects
  ✓ Sender signature forged → recipient rejects
  ✓ Ephemeral X25519 key unique per message
  ✓ Replay attack prevented (nonce unique)

SecurityTest_TokenSecurity()
  ✓ Access token expires (1 hour TTL)
  ✓ Refresh token has long TTL (30 days)
  ✓ Logout revokes both tokens
  ✓ Token stored in Keychain (not UserDefaults)
  ✓ Token header: "Authorization: Bearer <JWT>"

SecurityTest_AuthenticationBypass()
  ✓ Challenge reuse prevented (one-time nonce)
  ✓ Signature verification mandatory
  ✓ Missing auth header → 401
  ✓ Invalid auth header → 401
  ✓ Expired token → 401, client refreshes

SecurityTest_DataClassificationEnforcement()
  ✓ PII (name, email, phone) not logged to IPFS
  ✓ Only DIDs, timestamps, blob size logged
  ✓ Audit log encrypted at rest
  ✓ Audit log key split 3-of-5 (threshold crypto)
  ✓ Log retrieval logged as access event

SecurityTest_RateLimitBypass()
  ✓ Token bucket enforced per DID
  ✓ Burst capacity: 10 requests
  ✓ Refill rate: 100 requests/minute
  ✓ Distributed rate limiter (using Redis)
  ✓ Collusion across multiple clients detected
```

### 12.6 Test Coverage Goals

| Layer | Target Coverage | Status |
|-------|-----------------|--------|
| AuthService | 95% | Phase 2 |
| MessageService | 95% | Phase 2 |
| RewardService | 90% | Phase 2 |
| IdentityService | 90% | Phase 2 |
| MessageRelayService | 95% | Phase 2 |
| CacheService | 85% | Phase 2 |
| Integration tests (Chain ↔ Backend) | 80% | Phase 2 |
| E2E tests (App ↔ Backend ↔ Chain) | 70% | Phase 2 |
| Stress tests | Key scenarios | Phase 3 |
| Security tests | Critical paths | Phase 2 |

**CI/CD Pipeline:**
```yaml
On pull request:
  ✓ Unit tests (5 min)
  ✓ Integration tests (10 min)
  ✓ Lint + fmt (2 min)
  ✓ Code coverage (2 min) — must be > 80%

On merge to main:
  ✓ Full unit + integration test suite (20 min)
  ✓ E2E smoke tests against staging (15 min)
  ✓ Security scan (SAST) (5 min)
  ✓ Build + push Docker image (5 min)

Nightly:
  ✓ Stress tests (30 min)
  ✓ Full E2E test suite against staging (1 hour)
```

---

## 13. Missing Integration Checklist

### Phase 1–2 Critical Path

- [ ] **Go Backend Implementation**
  - [ ] AuthService (register, challenge, verify, refresh, logout)
  - [ ] MessageRelayService (send, queue offline, flush, rate limit)
  - [ ] RewardService (claim, batch to Currency L1)
  - [ ] IdentityService (DID creation, verify callback, credential issuance)
  - [ ] ContactService (add, remove, block, list)
  - [ ] WebSocket relay (connect, forward, ACK, confirmation)
  - [ ] Circuit breakers + health checks
  - [ ] Metrics + logging (OpenTelemetry)

- [ ] **Metagraph Integration**
  - [ ] Data L1 validator node (custom business logic)
  - [ ] Currency L1 validator (balance + staking)
  - [ ] Merkle root submission + finality handling
  - [ ] Snapshot event subscriptions
  - [ ] Offline queue fallback

- [ ] **Cardano Integration**
  - [ ] DID creation (did:prism via Atala PRISM or Veridian)
  - [ ] Credential schema publication
  - [ ] Credential issuance + revocation
  - [ ] Trust tier datum management
  - [ ] Transaction fee delegation (treasury pays)
  - [ ] Kupo indexer for credential queries

- [ ] **IPFS Integration**
  - [ ] Batch encryption + signing
  - [ ] Upload to Pinata + self-hosted node
  - [ ] CID recording + indexing on Data L1
  - [ ] Storj fallback configuration

- [ ] **APNs Integration**
  - [ ] Push notification sending (offline queue)
  - [ ] Silent notifications (delivery + confirmation)
  - [ ] VoIP push for message wake-up
  - [ ] Error handling + feedback loop

- [ ] **IDV Provider Integration** (Phase 2)
  - [ ] Stripe Identity SDK integration (server + client)
  - [ ] Webhook receiver + idempotency
  - [ ] Credential issuance on verification
  - [ ] Trust tier update + cache invalidation

### Phase 2–3 Enhancements

- [ ] **WebSocket Improvements**
  - [ ] User presence (online/offline)
  - [ ] Typing indicators (optional)
  - [ ] Voice/video call signaling (Phase 3)

- [ ] **Message Features**
  - [ ] Message reactions (emoji)
  - [ ] Disappearing messages (TTL)
  - [ ] Message search (backend indexing)
  - [ ] Message reactions (counts + list)

- [ ] **Achievement System**
  - [ ] Achievement definitions
  - [ ] Earned achievement tracking (Data L1)
  - [ ] Badge display in profile

- [ ] **Sealed Sender** (Phase 3)
  - [ ] Sender DID encryption
  - [ ] Delivery token obfuscation
  - [ ] Relay server metadata protection

### Phase 3–4 Decentralization

- [ ] **Federated Relay Operators**
  - [ ] Relay node registry (Data L1)
  - [ ] Staking requirement validation
  - [ ] Traffic distribution algorithm
  - [ ] Operator uptime monitoring

- [ ] **Direct P2P Messaging** (Optional, Phase 4)
  - [ ] Relay-assisted signaling
  - [ ] NAT traversal (STUN/TURN)
  - [ ] Direct WebSocket establishment
  - [ ] Fallback to relay if P2P fails

- [ ] **DAO Governance**
  - [ ] Proposal submission + voting
  - [ ] Stake-weighted voting
  - [ ] Treasury management
  - [ ] Upgrade activation

---

## 10. Open Questions Requiring Product Decisions

| # | Question | Impact | Needed By |
|---|----------|--------|-----------|
| 1 | Which IDV provider? (Stripe Identity, Sumsub, Jumio) | Cardano credential issuance flow | Phase 2 |
| 2 | Metagraph deployment: Constellation mainnet or dedicated metagraph? | Node infrastructure, cost model | Phase 2 |
| 3 | Staking: on metagraph Currency L1 or separate staking contract? | Tokenomics implementation | Phase 2 |
| 4 | Group messaging: group symmetric key vs. per-recipient encryption cutoff? (e.g., per-recipient for <100 members, group key for 100+) | Security/performance tradeoff | Phase 2 |
| 5 | Multi-device sync: device-linked keys or cloud key backup? | Security model, Secure Enclave assumptions | Phase 3 |
| 6 | ZK proof system: Groth16 (snarkjs) or alternative (Halo2, PLONK)? | Circuit complexity, verification cost | Phase 3 |
| 7 | DAO governance: on metagraph or Cardano? | Smart contract platform choice | Phase 4 |
| 8 | Federated relay: minimum stake requirement for relay operators? | Relay decentralization economics | Phase 4 |
| 9 | Disappearing messages: how to handle on-chain Merkle roots after message expiry? (Root persists but individual commitments become unverifiable after key deletion) | Privacy/auditability tradeoff | Phase 3 |
| 10 | Referral system: capped at N referrals or unlimited? | Fraud prevention vs. growth incentive | Phase 2 |

---

*Data Layer Architecture v3.1*
*Updated: February 23, 2026*
*Status: Messaging transport resolved (client-server relay). Service integrations mapped. Test coverage spec complete. Ready for Phase 1–2 implementation.*
