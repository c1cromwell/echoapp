The backend is implemented as a **stateless operational coordinator** and **content-blind message relay** built in Go. It sits between iOS clients and on-chain state, but is not an authority—it cannot read message content, does not own user identities, and does not control token balances. The backend's role is to coordinate operations, relay encrypted messages, cache chain state, and validate submissions before forwarding to blockchain layers.

## Architecture Philosophy

| Function | Authoritative Source | Backend Role |
| --- | --- | --- |
| Token balances | Metagraph Currency L1 | Read-through cache (TTL: 5s) |
| Trust scores | Cardano + metagraph | Compute engine + cache (TTL: 60s) |
| Message content | Device-local (E2E encrypted) | Relay only (queues ciphertext for offline) |
| Message integrity | Metagraph Data L1 (Merkle root) | Batch aggregator before submission |
| User identity | Cardano DID | Cache + credential proof validator |
| Reward eligibility | Metagraph Data L1 validators | Pre-validator (reject obviously invalid) |

**3-Tier Architecture:**

```
Layer 1: iOS Clients (Swift, Secure Enclave, E2E encryption)
         ↓ WebSocket (real-time relay) + REST (operations)
Layer 2: Go Backend (Stateless relay + operational services)
         ↓ Chain APIs + submission pipelines
Layer 3: Constellation Metagraph (L1/L0) | Cardano | IPFS/Storj
```

## Technology Stack

* **Language & Framework**: Go with REST APIs and WebSocket for real-time messaging
* **Message Encryption**: End-to-end encryption on devices (Secure Enclave). Backend handles opaque encrypted blobs only
* **Caching**: Redis for hot data (balances, trust tiers, credentials) with TTL-based invalidation
* **Persistence**: PostgreSQL for operational data and offline message queues
* **Event Bus**: NATS for cross-service pub/sub and group message fan-out
* **Real-time Transport**: WebSocket (WSS) for message relay with sticky load balancing
* **Push Notifications**: APNs integration for offline message delivery alerts
* **Logging & Monitoring**: Batched, encrypted logs stored in decentralized storage (IPFS/Storj) with monthly key rotation
* **API Versioning**: Support multiple API versions (e.g., `/v1/`, `/v2/`) for backward compatibility
* **Transport Security**: TLS 1.3+ for all data in transit with certificate pinning

## Service Architecture

The backend consists of 10 independent microservices, each handling a specific domain:

| Service | Port | Role | Downstream Dependencies |
| --- | --- | --- | --- |
| **Gateway** | 8000 | Load balancer, TLS termination, rate limiting | All services |
| **Identity Service** | 8001 | Registration, DID management, credential caching | Cardano, Redis |
| **Message Relay** | 8002 | WebSocket relay, offline queue, APNs push | Redis, PostgreSQL, NATS |
| **Trust Service** | 8003 | Trust score computation, tier caching | Cardano, Metagraph, Redis |
| **Rewards Service** | 8004 | Reward validation (auto-scaling rate, annual emission enforcement), batching, submission | Metagraph Currency L1, Redis |
| **Contacts Service** | 8005 | Contact list, block list, search, privacy-preserving contact discovery (Argon2id hashed matching) | PostgreSQL, Redis |
| **Metagraph Gateway** | 8006 | L1/L0 submission, snapshot listening, anchoring | Metagraph nodes |
| **Notification Service** | 8007 | APNs push, in-app notifications | APNs, Redis |
| **Media Service** | 8008 | Encrypted media upload/download | Storj/S3, Redis |
| **Log Publisher** | 8009 | Batch encryption, IPFS submission, CID indexing | IPFS/Storj, Metagraph Data L1 |

## Core Responsibilities

**Message Relay (Content-Blind)**: The relay service transports end-to-end encrypted message blobs between clients without the ability to read, decrypt, or modify content. For online recipients, messages are delivered via WebSocket. For offline recipients, encrypted blobs are queued in Redis/PostgreSQL and delivered when the recipient reconnects. Push notifications alert offline users without exposing message content.

**Data Validation**: The backend pre-validates requests before submission to reduce unnecessary blockchain transactions. However, the metagraph L1 layers perform final authoritative validation. The backend can reject obviously invalid data but cannot override on-chain validation rules.

**Authentication Coordination**: The backend coordinates with the Cardano identity layer to verify user credentials and manage trust levels. It caches DID documents and credential status with TTL-based invalidation (60s). The backend validates passkey signatures but does not store private keys.

**Metagraph Integration**: The backend submits validated transactions to the Constellation metagraph using Tessellation v3 transaction primitives (TokenLock, StakeDelegation, AtomicAction, FeeTransaction). It uses owned nodes for critical write operations and third-party APIs for read operations. Circuit breakers per chain isolate failures—if the metagraph is down, message relay continues with cached state.

**Message Integrity Anchoring**: The backend batches message commitment hashes into Merkle trees every 5 minutes (or 1000 messages, whichever comes first) and submits the Merkle root to the Data L1 layer. This proves message integrity on-chain without exposing encrypted content. Clients can later request Merkle proofs to verify their messages were anchored.

**Rate Limiting & Throttling**: Implements tiered per-DID API rate limiting with base users limited to 100 requests/minute and VIP users (subscription: $9.99/month) receiving 2x-10x increases based on app scaling. Note: API rate limiting is for abuse prevention only — it does not cap token rewards. Reward distribution uses an auto-scaling model with no per-user daily caps (see Rewards Service).

**Logging & Monitoring**: Batches operational events (no PII, no message content) into encrypted logs with monthly rotating AES-256-GCM keys. Logs are pushed to IPFS/Storj every 5 minutes with the CID recorded on the Data L1 for auditability.

## API Design Principles

* All endpoints require authentication via passkey verification (ECDSA P-256 signature)
* Backend validates signatures but never stores private keys (keys remain in Secure Enclave)
* Responses include structured error codes for client-side handling
* CORS policy is strict, allowing requests only from verified iOS app origins
* Message content is never exposed to backend—only opaque encrypted blobs
* Metadata is minimized to what's necessary for routing (sender DID, recipient DIDs, timestamps)

## Authentication Endpoints

The Identity Service (port 8001) provides authentication endpoints:

**Registration & DID Creation (Zero PII)**: When a user creates an account, they enter a username and the iOS app generates a P-256 key pair in the Secure Enclave. The backend receives only the username and public key — no phone number, no email, no real name. The backend submits a DID registration transaction to Cardano (fees paid by platform treasury) and returns the DID document. The private key never leaves the device. Users start at Trust Tier 1 and can optionally verify their phone number (Tier 2) or government ID (Tier 3-4) to unlock additional features and higher reward multipliers.

**Passkey Verification**: All API requests include an ECDSA signature over the request payload. The backend verifies the signature against the user's DID public key cached from Cardano. Failed signature checks result in HTTP 401.

**Credential Caching**: The backend caches verifiable credentials and trust tier information from Cardano with a 60-second TTL. This avoids per-request blockchain queries while maintaining reasonable freshness for access control decisions.

**Phone Verification (Optional Tier 2 Upgrade)**: Users who want contact discovery and higher reward multipliers can optionally verify their phone number. The backend sends an SMS OTP, verifies the response, and upgrades the user to Trust Tier 2 (1.2x reward multiplier). The raw phone number is never stored — the iOS app hashes it on-device using Argon2id with a per-user salt and sends only the hash to the contact discovery index. This flow is triggered by the user tapping a trust-tier upgrade card, not during onboarding.

**Third-Party Verification Coordination**: For identity verification providers (Prove, Daon, 1Kosmos, Darwinium), the backend coordinates the verification flow and submits successful verifications as credentials to Cardano. The backend acts as an orchestrator but does not store verification results—they live on-chain as verifiable credentials.


## Contact Discovery Endpoints

The Contacts Service (port 8005) provides privacy-preserving contact discovery:

**Phone Number Matching (Opt-In)**: Users hash their contacts' phone numbers on-device using Argon2id with a per-user salt before sending to the server. The server matches hashed entries against its index of registered users (also stored as salted hashes) and returns encrypted DID references. The server never sees raw phone numbers. Rate limited to 1 discovery request per 24 hours per DID.

**QR Code Exchange**: Users generate a QR code containing their DID and public key. Scanning creates a mutual contact connection with zero server involvement for the QR generation itself. The backend records the connection for message routing purposes only.

**Username Search**: Users who create a public handle (optional) can be discovered via `GET /contacts/search?handle={username}`. Returns DID, display name, trust tier badge, and verification status. Handles are not linked to real names on-chain.

**Invite Links**: Users generate unique referral links via `POST /contacts/invite`. The backend tracks the referral chain (max 3 tiers) for the 50 ECHO referral reward, triggered when the new user completes DID verification and sends their first 100 messages.

**Contact Discovery Index**: The server-side index stores only Argon2id salted hashes linked to encrypted DID references. The index is NOT stored on any public blockchain. Even a complete server breach reveals no usable phone numbers — only irreversible hashes linked to encrypted pointers.


## Enterprise Fraud Prevention Endpoints

The Identity Service (port 8001) extends to support Organization-tier enterprise fraud prevention:

**Transaction Verification Alerts**: Banks send cryptographically signed transaction alerts through ECHO's verified channel via `POST /enterprise/fraud/alert`. The alert includes the bank's verified institutional DID signature, transaction details (amount, merchant, timestamp), and a unique verification request ID. The customer receives the alert with the bank's verification badge and responds with a DID-signed confirmation or fraud report via `POST /enterprise/fraud/confirm`, creating a court-admissible authorization record.

**Fraud Analytics Dashboard**: Organization-tier customers access fraud analytics via `GET /enterprise/fraud/dashboard`. Metrics include: fraud attempt volume (phishing attempts blocked by verified channel vs. SMS baseline), customer response times to fraud alerts, verification adoption rate (% of customers using ECHO vs. SMS), and ROI calculator (cost savings vs. SMS fraud losses). Dashboard data is computed from on-chain Digital Evidence records and relay metadata — no PII.

**Cross-Organization Fraud Intelligence (Phase 5+)**: Participating institutions query fraud patterns via `GET /enterprise/fraud/intelligence` using zero-knowledge proofs through the Midnight integration. Queries like "has DID xyz been flagged by 3+ institutions in 30 days" return a boolean result without revealing which institutions flagged it or the specific fraud type. This leverages ECHO's Midnight ZK infrastructure for privacy-preserving inter-institutional data sharing.

## Metagraph Integration Patterns

The Metagraph Gateway service (port 8006) handles all interactions with Constellation infrastructure:

**Data L1 Submissions**: The backend submits application data to the Data L1 layer for consensus. Primary use case is anchoring Merkle roots for message integrity verification. Submissions use this structure:

```go
type DataL1Submission struct {
    Type            string    // "message_integrity", "audit_log"
    MerkleRoot      []byte    // Root hash of Merkle tree
    CommitmentCount int       // Number of messages in batch
    TimeRange       TimeRange // From/To timestamps
    SchemaVersion   int       // Current: 1
}
```

**Currency L1 Submissions**: Token transactions use Tessellation v3 transaction primitives for Hypergraph interoperability:

```go
type CurrencyL1Transaction struct {
    Type            string
    TokenLock       *TokenLockData       // Lock ECHO for staking
    StakeDelegation *StakeDelegationData // Delegate to validator
    WithdrawLock    *WithdrawLockData    // Unstaking (14-day cooldown)
    AtomicBundle    *AtomicActionBundle  // Bundle multiple actions
    FeeTransaction  *FeeTransactionData  // Automated snapshot fee payment
    // Legacy types still supported:
    Claims          []RewardClaim
    Transfer        *TokenTransfer
}
```

**AtomicAction Bundles**: Multiple transactions can be bundled into all-or-nothing execution. Used for reward claims (verify tier + claim + update cap), staking tier changes, and governance operations.

**Snapshot Listening**: The backend subscribes to metagraph snapshot events. On each new snapshot, it invalidates caches and pushes confirmation events to clients via WebSocket. This ensures clients see finalized state within seconds of on-chain confirmation.

**Circuit Breakers**: Independent circuit breakers per downstream chain (Data L1, Currency L1, Cardano, IPFS). If a chain is unavailable, the circuit opens and the backend continues operating with cached state. Message relay never blocks on chain availability. Failed submissions queue for retry with exponential backoff.

**Error Handling**: Metagraph validation errors return structured error codes to clients. The backend distinguishes between retriable errors (network timeouts) and non-retriable errors (invalid transaction format) to avoid unnecessary retries.


## ZK Verification Endpoints (Phase 3+)

The Trust Service (port 8003) extends to coordinate zero-knowledge proof verification via Midnight:

**Trust Tier Verification**: `POST /zk/verify/trust-tier` accepts a ZK proof generated on-device demonstrating the user meets a minimum trust tier threshold. The proof is verified via the Midnight partner chain. Public signals contain only the threshold and a boolean result — not the exact score. Used for governance eligibility checks and feature access gating.

**Age Verification**: `POST /zk/verify/age` accepts a ZK proof that the user's age exceeds a threshold (18 or 21) without revealing their actual birthdate. Used for age-gated financial institution features.

**Credential Validity**: `POST /zk/verify/credential` accepts a ZK proof that a credential is valid and issued by a trusted authority, without revealing credential content. Includes a verifier challenge nonce to prevent replay attacks.

**Balance Threshold**: `POST /zk/verify/balance` accepts a ZK proof that the user's ECHO token balance meets a minimum threshold for staking eligibility or VIP access, without revealing the exact balance.

All ZK proofs are generated on the user's iOS device — private inputs (birthdate, score, credential content, exact balance) never leave the device. The backend forwards proofs to Midnight for on-chain verification and caches the verification result with a configurable TTL. Midnight uses the Compact language (TypeScript-based DSL) — no Scala dependency for ZK contracts.

## Encryption & Message Security

**End-to-End Encryption**: Message content is encrypted and decrypted exclusively on client devices using keys stored in the iOS Secure Enclave. The backend never has access to plaintext message content or encryption keys. Messages in transit through the backend are opaque encrypted blobs.

**Message Relay Flow**:

1. **Sender encrypts** message on-device using ephemeral X25519 key agreement with recipient's public key + ChaCha20-Poly1305 symmetric encryption (Kinnami)
2. **Sender signs** a commitment hash `H(H(plaintext) || nonce)` with their private key
3. **Backend receives** encrypted blob + commitment + signature (cannot read content)
4. **Backend validates** signature against sender's DID public key
5. **Backend relays** encrypted blob to recipient via WebSocket (if online) or queues (if offline)
6. **Recipient decrypts** on-device using their private key from Secure Enclave

**Offline Message Queue**: Encrypted blobs for offline recipients are stored in Redis (fast, in-memory) with PostgreSQL fallback (durable). Queue depth is limited to 1000 messages per recipient with retention: 30 days for 1:1 chats, 7 days for large groups (100+ members).

**Overflow Message Backup**: When a recipient's queue exceeds 1000 messages, overflow encrypted blobs are pinned to IPFS/Storj. The relay stores only the CID in the queue metadata. On reconnect, the relay provides CIDs for the recipient to retrieve overflow messages directly from IPFS. This prevents message loss during extended offline periods while maintaining the content-blind relay model — the backup is the same opaque E2E encrypted blob the relay would have queued.

**Commitment Anchoring**: The backend batches commitment hashes into Merkle trees and anchors the root on-chain every 5 minutes. This proves message integrity without exposing content. Clients can verify their messages were anchored by requesting Merkle proofs.

**Transport Security**: All client-backend communication uses TLS 1.3 with certificate pinning. WebSocket connections upgrade from HTTPS with strict same-origin policies. Push notifications never contain message content—only conversation IDs for wake-up signals.

**Backend Logging Keys**: Operational logs (no PII, no message content) are encrypted with AES-256-GCM using monthly rotating keys. Keys are managed by the platform and never shared with third parties. Encrypted logs are pushed to IPFS/Storj with CIDs recorded on-chain for auditability.

## Scaling & Reliability Architecture

**Stateless Pods**: All Go services are stateless and horizontally scalable. WebSocket connections are sticky to a pod via load balancer, but any pod can handle any REST request. This enables auto-scaling based on CPU, connection count, or relay latency metrics.

**Group Message Fan-Out**: When a group message arrives at Pod 1 but recipients are connected to Pods 2 and 3, NATS pub/sub distributes the encrypted blob to all pods for local delivery. This avoids cross-pod WebSocket proxying and reduces latency.

**Circuit Breakers Per Chain**: Independent circuit breakers for Data L1, Currency L1, Cardano, and IPFS. Thresholds: 5 failures for metagraph/IPFS, 3 failures for Cardano. Reset timeout: 30s for metagraph, 60s for Cardano, 120s for IPFS. When a circuit opens, the backend serves from cache and queues submissions for retry.

**Auto-Scaling Triggers**: Kubernetes horizontal pod autoscaler monitors: CPU &gt; 70%, WebSocket connections &gt; 10K per pod, relay latency P99 &gt; 200ms. Pods scale from minimum 3 to maximum based on demand.

**Multi-Region & Multi-Cloud Deployment**: Phase 1–3 deploys on Hetzner Cloud (Germany — strongest EU privacy jurisdiction, 60–70% cheaper than AWS, no Big Tech association). Phase 3 adds OVHcloud (France) as secondary for EU failover. Phase 4 introduces community-operated relay nodes on Akash Network, Flux, or bare metal for full decentralization. No single cloud provider may serve more than 60% of total relay traffic. Minimum 3 geographic regions for relay coverage. Redis and PostgreSQL use synchronous replication (primary + 2 replicas) to minimize data loss on failover. NATS cluster spans regions for event distribution redundancy.