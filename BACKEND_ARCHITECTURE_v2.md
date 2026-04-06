# Backend Architecture & Implementation (v2.0)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | February 23, 2026 | Aligned with relay architecture decision. Added: Message Relay Service, offline message queue, circuit breakers per chain, metagraph submission pipeline, Cardano integration service, IPFS log publisher, group fan-out, sealed sender support (Phase 3), anchoring confirmation pipeline. Updated service registry. Clarified backend role as stateless relay + cache (not authority). |
| 1.0 | February 2026 | Foundation implementation |

---

## Architecture Overview

The Go backend is a **stateless operational coordinator** and **content-blind message relay** that sits between iOS clients and on-chain state. It is not an authority — it cannot read message content, does not own user identities, and does not control token balances.

**Backend Role:**

| Function | Authoritative Source | Backend Role |
|----------|---------------------|-------------|
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

---

## Service Registry

| Service | Port | Role | Downstream Dependencies |
|---------|------|------|------------------------|
| **Gateway** | 8000 | Load balancer, TLS termination, rate limiting | All services |
| **Identity Service** | 8001 | Registration, DID management, credential caching | Cardano, Redis |
| **Message Relay** | 8002 | WebSocket relay, offline queue, APNs push | Redis, PostgreSQL, NATS |
| **Trust Service** | 8003 | Trust score computation, tier caching | Cardano, Metagraph, Redis |
| **Rewards Service** | 8004 | Reward validation, batching, submission | Metagraph Currency L1, Redis |
| **Contacts Service** | 8005 | Contact list, block list, search | PostgreSQL, Redis |
| **Metagraph Gateway** | 8006 | L1/L0 submission, snapshot listening, anchoring | Metagraph nodes |
| **Notification Service** | 8007 | APNs push, in-app notifications | APNs, Redis |
| **Media Service** | 8008 | Encrypted media upload/download | Storj/S3, Redis |
| **Log Publisher** | 8009 | Batch encryption, IPFS submission, CID indexing | IPFS/Storj, Metagraph Data L1 |

---

## New Services (v2.0)

### Message Relay Service

The core messaging service. Operates as a content-blind relay — it transports E2E encrypted blobs between clients without the ability to read, decrypt, or modify content.

**Package:** `internal/services/relay/`

```go
// relay/relay.go — Message Relay Service

package relay

import (
    "context"
    "time"
)

// RelayService manages WebSocket connections and message transport.
// It CANNOT read message content (E2E encrypted blobs only).
type RelayService struct {
    connections  *ConnectionManager  // Active WebSocket connections by DID
    offlineQueue *OfflineQueue       // Redis/PG queue for offline recipients
    notifier     *APNsNotifier       // Push notifications
    anchoring    *AnchoringBatcher   // Merkle tree commitment batching
    rateLimiter  *RateLimiter        // Per-DID send rate limits
    nats         *NATSClient         // Pub/sub for cross-pod fan-out
}

// RelayMessage represents an E2E encrypted blob in transit.
// The server sees metadata only; content is opaque ciphertext.
type RelayMessage struct {
    MessageID      string    `json:"messageId"`
    ConversationID string    `json:"conversationId"`
    SenderDID      string    `json:"senderDID"`       // Hidden in Phase 3 (sealed sender)
    RecipientDIDs  []string  `json:"recipientDIDs"`
    ContentType    string    `json:"contentType"`
    EncryptedBlob  []byte    `json:"encryptedBlob"`    // Opaque E2E ciphertext
    Commitment     []byte    `json:"commitment"`        // H(H(plaintext) || nonce)
    Signature      []byte    `json:"signature"`         // Sender's P-256 signature
    Timestamp      time.Time `json:"timestamp"`
    ExpiresAt      *time.Time `json:"expiresAt,omitempty"` // Disappearing messages
}

// RelayResult indicates whether the message was delivered live or queued.
type RelayResult struct {
    MessageID string `json:"messageId"`
    Status    string `json:"status"` // "relayed" or "queued"
    Timestamp time.Time `json:"timestamp"`
}

// Relay processes an incoming message from a sender.
func (s *RelayService) Relay(ctx context.Context, msg RelayMessage) (*RelayResult, error) {
    // 1. Rate limit check
    if err := s.rateLimiter.Check(msg.SenderDID, "message_send"); err != nil {
        return nil, ErrRateLimitExceeded
    }

    // 2. For each recipient
    status := "relayed"
    for _, recipientDID := range msg.RecipientDIDs {
        conn, online := s.connections.Get(recipientDID)
        if online {
            // Deliver via WebSocket
            if err := conn.SendMessage(msg); err != nil {
                // Fallback to offline queue on send failure
                s.offlineQueue.Enqueue(recipientDID, msg)
                status = "queued"
            }
        } else {
            // Queue for offline delivery + send push notification
            s.offlineQueue.Enqueue(recipientDID, msg)
            s.notifier.SendPush(ctx, recipientDID, msg.ConversationID)
            status = "queued"
        }
    }

    // 3. Add commitment to anchoring batch
    s.anchoring.AddCommitment(msg.MessageID, msg.Commitment)

    return &RelayResult{
        MessageID: msg.MessageID,
        Status:    status,
        Timestamp: time.Now(),
    }, nil
}

// DrainOfflineQueue sends all queued messages to a reconnecting client.
func (s *RelayService) DrainOfflineQueue(ctx context.Context, did string, conn *WebSocketConn) error {
    messages, err := s.offlineQueue.DequeueAll(did)
    if err != nil {
        return err
    }
    for _, msg := range messages {
        if err := conn.SendMessage(msg); err != nil {
            // Re-queue on failure
            s.offlineQueue.Enqueue(did, msg)
            return err
        }
    }
    return nil
}
```

**Offline Queue:**

```go
// relay/offline_queue.go

package relay

import (
    "context"
    "time"
)

// OfflineQueue stores encrypted message blobs for offline recipients.
// Uses Redis (fast, in-memory) with PostgreSQL fallback (durable).
type OfflineQueue struct {
    redis    RedisClient
    postgres PostgresClient
}

const (
    MaxQueueDepth     = 1000          // Per recipient
    DefaultRetention  = 30 * 24 * time.Hour  // 30 days for 1:1
    GroupRetention    = 7 * 24 * time.Hour   // 7 days for large groups
)

// Enqueue adds an encrypted blob to the recipient's offline queue.
func (q *OfflineQueue) Enqueue(recipientDID string, msg RelayMessage) error {
    // Check queue depth; evict oldest if exceeded
    depth, _ := q.redis.LLen(ctx, queueKey(recipientDID))
    if depth >= MaxQueueDepth {
        q.redis.LPop(ctx, queueKey(recipientDID)) // Evict oldest
    }

    retention := DefaultRetention
    if len(msg.RecipientDIDs) > 100 {
        retention = GroupRetention
    }

    data, _ := json.Marshal(msg)
    return q.redis.RPush(ctx, queueKey(recipientDID), data, retention)
}

// DequeueAll retrieves and removes all queued messages for a recipient.
func (q *OfflineQueue) DequeueAll(recipientDID string) ([]RelayMessage, error) {
    // Atomic: get all + delete key
    data, err := q.redis.LRangeAndDelete(ctx, queueKey(recipientDID))
    if err != nil {
        return nil, err
    }
    
    messages := make([]RelayMessage, 0, len(data))
    for _, d := range data {
        var msg RelayMessage
        if err := json.Unmarshal(d, &msg); err == nil {
            messages = append(messages, msg)
        }
    }
    return messages, nil
}

func queueKey(did string) string {
    return "offline:" + did
}
```

### Anchoring Batcher

Batches message commitments into Merkle trees and submits roots to the metagraph Data L1:

```go
// metagraph/anchoring.go

package metagraph

import (
    "time"
)

// AnchoringBatcher collects message commitment hashes and periodically
// builds a Merkle tree, submitting the root to the Data L1.
type AnchoringBatcher struct {
    commitments  []Commitment
    metagraph    *MetagraphClient
    logPublisher *LogPublisher
    ticker       *time.Ticker
    maxBatch     int
}

type Commitment struct {
    MessageID string
    Hash      []byte
    Timestamp time.Time
}

const (
    BatchInterval = 5 * time.Minute
    MaxBatchSize  = 1000
)

// AddCommitment adds a message commitment to the current batch.
func (b *AnchoringBatcher) AddCommitment(messageID string, hash []byte) {
    b.commitments = append(b.commitments, Commitment{
        MessageID: messageID,
        Hash:      hash,
        Timestamp: time.Now(),
    })

    // Flush if batch size reached
    if len(b.commitments) >= b.maxBatch {
        go b.flush()
    }
}

// flush builds Merkle tree and submits to Data L1.
func (b *AnchoringBatcher) flush() {
    if len(b.commitments) == 0 {
        return
    }

    batch := b.commitments
    b.commitments = nil

    // 1. Build Merkle tree
    leaves := make([][]byte, len(batch))
    for i, c := range batch {
        leaves[i] = c.Hash
    }
    tree := BuildMerkleTree(leaves)
    root := tree.Root()

    // 2. Submit Merkle root to Data L1
    txHash, err := b.metagraph.SubmitDataL1(DataL1Submission{
        Type:          "message_integrity",
        MerkleRoot:    root,
        CommitmentCount: len(batch),
        TimeRange: TimeRange{
            From: batch[0].Timestamp,
            To:   batch[len(batch)-1].Timestamp,
        },
        SchemaVersion: 1,
    })
    if err != nil {
        // Retry with exponential backoff
        go b.retrySubmission(batch, root)
        return
    }

    // 3. Store tree for Merkle proof generation (Phase 3)
    b.storeTree(txHash, tree, batch)

    // 4. Push to log publisher
    b.logPublisher.AddBatch(batch, root, txHash)
}

// Run starts the periodic flush ticker.
func (b *AnchoringBatcher) Run() {
    b.ticker = time.NewTicker(BatchInterval)
    for range b.ticker.C {
        b.flush()
    }
}
```

### Metagraph Gateway Service

Handles all interactions with the Constellation metagraph (Data L1 + Currency L1):

```go
// metagraph/gateway.go

package metagraph

// MetagraphClient manages connections to Constellation metagraph nodes.
type MetagraphClient struct {
    dataL1Endpoint     string
    currencyL1Endpoint string
    l0Endpoint         string
    circuitBreaker     *CircuitBreaker
    ownedNodes         []NodeConfig    // For write operations
    thirdPartyAPIs     []string        // For read operations (fallback)
}

// SubmitDataL1 submits application data to the Data L1 for consensus.
func (c *MetagraphClient) SubmitDataL1(submission DataL1Submission) (string, error) {
    if !c.circuitBreaker.Allow("data_l1") {
        return "", ErrCircuitOpen
    }
    // Submit to owned node (write path)
    resp, err := c.postToOwnedNode(c.dataL1Endpoint, submission)
    if err != nil {
        c.circuitBreaker.RecordFailure("data_l1")
        return "", err
    }
    c.circuitBreaker.RecordSuccess("data_l1")
    return resp.TxHash, nil
}

// SubmitCurrencyL1 submits token transactions (rewards, transfers, staking).
func (c *MetagraphClient) SubmitCurrencyL1(tx CurrencyL1Transaction) (string, error) {
    if !c.circuitBreaker.Allow("currency_l1") {
        return "", ErrCircuitOpen
    }
    resp, err := c.postToOwnedNode(c.currencyL1Endpoint, tx)
    if err != nil {
        c.circuitBreaker.RecordFailure("currency_l1")
        return "", err
    }
    c.circuitBreaker.RecordSuccess("currency_l1")
    return resp.TxHash, nil
}

// ListenSnapshots subscribes to metagraph snapshot events.
// On each new snapshot, invalidates caches and pushes confirmations to clients.
func (c *MetagraphClient) ListenSnapshots(handler func(Snapshot)) {
    // Long-poll or WebSocket to L0 node for snapshot events
    for {
        snapshot, err := c.pollSnapshot()
        if err != nil {
            time.Sleep(time.Second)
            continue
        }
        handler(snapshot)
    }
}

// DataL1Submission types
type DataL1Submission struct {
    Type            string    `json:"type"`
    MerkleRoot      []byte    `json:"merkleRoot,omitempty"`
    CommitmentCount int       `json:"commitmentCount,omitempty"`
    TimeRange       TimeRange `json:"timeRange,omitempty"`
    SchemaVersion   int       `json:"schemaVersion"`
    // Additional fields per submission type (rewards, governance, trust, etc.)
}

type CurrencyL1Transaction struct {
    Type   string `json:"type"` // "reward_batch", "transfer", "stake", "unstake"
    Claims []RewardClaim `json:"claims,omitempty"`
    Transfer *TokenTransfer `json:"transfer,omitempty"`
    Stake    *StakeOperation `json:"stake,omitempty"`
}

type TimeRange struct {
    From time.Time `json:"from"`
    To   time.Time `json:"to"`
}
```

### Cardano Integration Service

Manages all Cardano interactions (DID operations, credentials, trust tiers):

```go
// cardano/cardano.go

package cardano

// CardanoClient manages connections to Cardano nodes for identity operations.
type CardanoClient struct {
    nodeEndpoint   string
    network        string            // "mainnet" or "preprod"
    circuitBreaker *CircuitBreaker
    treasury       *TreasuryWallet   // Pays transaction fees on behalf of users
    pollInterval   time.Duration     // Credential status polling (60s)
}

// RegisterDID creates a new DID document on Cardano.
func (c *CardanoClient) RegisterDID(publicKey []byte) (*DIDDocument, error) {
    if !c.circuitBreaker.Allow("cardano") {
        return nil, ErrCircuitOpen
    }
    // Build Cardano transaction with DID metadata
    // Fee paid by platform treasury
    tx := c.buildDIDRegistrationTx(publicKey)
    txHash, err := c.submitTx(tx)
    if err != nil {
        c.circuitBreaker.RecordFailure("cardano")
        return nil, err
    }
    c.circuitBreaker.RecordSuccess("cardano")
    
    return &DIDDocument{
        DID:           "did:prism:" + txHash[:32],
        PublicKeys:    []PublicKeyEntry{{Key: publicKey, Type: "EcdsaSecp256r1"}},
        CardanoTxHash: txHash,
        CreatedAt:     time.Now(),
    }, nil
}

// IssueCredential records a verifiable credential on Cardano.
func (c *CardanoClient) IssueCredential(req CredentialRequest) (*Credential, error) {
    // Build Plutus transaction with credential metadata
    // Update credential status bit vector
    // Fee paid by platform treasury (~0.3-0.5 ADA)
    return nil, nil // Implementation
}

// RevokeCredential sets the revocation bit for a credential.
func (c *CardanoClient) RevokeCredential(credentialID string) error {
    // Update bit vector in Plutus UTXO datum
    return nil // Implementation
}

// GetTrustTier queries Cardano for a user's trust tier.
// Results are cached (TTL 60s) to avoid per-request chain queries.
func (c *CardanoClient) GetTrustTier(did string) (int, error) {
    // Query UTXO datum for trust tier
    return 0, nil // Implementation
}

// PollCredentialStatus periodically checks for revocation changes.
func (c *CardanoClient) PollCredentialStatus(ctx context.Context) {
    ticker := time.NewTicker(c.pollInterval)
    for {
        select {
        case <-ticker.C:
            c.refreshCredentialCache()
        case <-ctx.Done():
            return
        }
    }
}
```

### Circuit Breaker

Independent circuit breakers per downstream chain:

```go
// infra/circuit_breaker.go

package infra

import (
    "sync"
    "time"
)

type CircuitBreaker struct {
    mu       sync.RWMutex
    circuits map[string]*Circuit
}

type Circuit struct {
    Name           string
    State          CircuitState  // Closed, Open, HalfOpen
    FailureCount   int
    FailureThreshold int        // e.g., 5 failures
    ResetTimeout   time.Duration // e.g., 30 seconds
    LastFailure    time.Time
}

type CircuitState int

const (
    CircuitClosed   CircuitState = iota // Normal operation
    CircuitOpen                         // Blocking requests
    CircuitHalfOpen                     // Testing recovery
)

// NewCircuitBreaker creates breakers for each chain.
func NewCircuitBreaker() *CircuitBreaker {
    return &CircuitBreaker{
        circuits: map[string]*Circuit{
            "data_l1":     {Name: "Metagraph Data L1", FailureThreshold: 5, ResetTimeout: 30 * time.Second},
            "currency_l1": {Name: "Metagraph Currency L1", FailureThreshold: 5, ResetTimeout: 30 * time.Second},
            "cardano":     {Name: "Cardano", FailureThreshold: 3, ResetTimeout: 60 * time.Second},
            "ipfs":        {Name: "IPFS/Storj", FailureThreshold: 5, ResetTimeout: 120 * time.Second},
        },
    }
}

// Allow checks if requests should be allowed through.
func (cb *CircuitBreaker) Allow(name string) bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    c := cb.circuits[name]
    if c.State == CircuitClosed {
        return true
    }
    if c.State == CircuitOpen && time.Since(c.LastFailure) > c.ResetTimeout {
        c.State = CircuitHalfOpen
        return true
    }
    return c.State == CircuitHalfOpen
}
```

### IPFS Log Publisher

Batches, encrypts, and publishes audit logs to IPFS:

```go
// logging/publisher.go

package logging

// LogPublisher batches operational events, encrypts them, and pushes to IPFS.
type LogPublisher struct {
    buffer        []LogEntry
    encryptionKey []byte           // Monthly rotating AES-256-GCM key
    ipfsClient    *IPFSClient
    metagraph     *MetagraphClient // For recording CID on Data L1
    maxBuffer     int              // 1000 entries
    flushInterval time.Duration    // 5 minutes
}

type LogEntry struct {
    EventType string    `json:"eventType"` // "relay_batch", "reward_claim", etc.
    Count     int       `json:"count"`
    Timestamp time.Time `json:"timestamp"`
    // No PII, no DIDs (unless required for compliance), no message content
}

// Flush encrypts the buffer, pushes to IPFS, and records the CID on-chain.
func (p *LogPublisher) Flush() error {
    if len(p.buffer) == 0 {
        return nil
    }

    batch := p.buffer
    p.buffer = nil

    // 1. Serialize + compress (zstd)
    data := serializeAndCompress(batch)

    // 2. Build Merkle tree for integrity
    root := merkleRootOfEntries(batch)

    // 3. Encrypt with AES-256-GCM
    encrypted, nonce, err := encryptAESGCM(data, p.encryptionKey)
    if err != nil {
        return err
    }

    // 4. Push to IPFS
    cid, err := p.ipfsClient.Add(encrypted)
    if err != nil {
        // Buffer locally; retry later
        p.bufferLocally(encrypted, root)
        return err
    }

    // 5. Pin on Pinata + self-hosted node
    p.ipfsClient.Pin(cid)

    // 6. Record CID + Merkle root on Data L1
    p.metagraph.SubmitDataL1(DataL1Submission{
        Type:       "audit_log",
        MerkleRoot: root,
        IPFSCID:    cid,
        EntryCount: len(batch),
        TimeRange:  timeRangeOf(batch),
    })

    return nil
}
```

---

## Service Communication

| Pattern | Use Case | Technology |
|---------|----------|------------|
| **Synchronous** | Client → Backend API requests | REST (JSON over TLS 1.3) |
| **Real-time** | Client ↔ Message relay | WebSocket (WSS) |
| **Async events** | Cross-service events, group fan-out | NATS pub/sub |
| **Chain writes** | Metagraph/Cardano submissions | Direct to owned nodes |
| **Chain reads** | Balance queries, credential checks | Third-party APIs (Infura/Alchemy) + cache |
| **Caching** | Hot data (balances, trust tiers, credentials) | Redis (TTL-based) |
| **Persistence** | Operational data, offline queues | PostgreSQL |
| **Logging** | Encrypted audit trail | IPFS/Storj |

---

## Scaling Architecture

```
                    ┌─────────────────────┐
                    │   Load Balancer      │
                    │   (Regional, TLS)    │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼──────┐ ┌──────▼────────┐ ┌─────▼───────┐
     │ Go Pod 1      │ │ Go Pod 2      │ │ Go Pod N    │
     │ (All services │ │ (All services │ │ (All svc    │
     │  stateless)   │ │  stateless)   │ │  stateless) │
     └───────┬───────┘ └──────┬────────┘ └──────┬───────┘
             │                │                  │
     ┌───────▼────────────────▼──────────────────▼───────┐
     │                 Shared Infrastructure               │
     │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
     │  │  Redis    │  │  NATS    │  │  PostgreSQL      │  │
     │  │(Cache +   │  │ (Events +│  │  (Operational +  │  │
     │  │ Msg Queue)│  │  Fanout) │  │   Offline Queue) │  │
     │  └──────────┘  └──────────┘  └──────────────────┘  │
     └───────┬────────────────┬──────────────────┬───────┘
             │                │                  │
     ┌───────▼───┐    ┌──────▼──────┐    ┌──────▼──────┐
     │ Metagraph │    │   Cardano   │    │  IPFS/Storj │
     │ (L1/L0)   │    │  (Identity) │    │  (Logging)  │
     │ Circuit: ✅│    │  Circuit: ✅│    │  Circuit: ✅│
     └───────────┘    └─────────────┘    └─────────────┘
```

**Key scaling properties:**

- **All Go pods are stateless.** WebSocket connections are sticky to a pod via load balancer, but any pod can handle any REST request.
- **Group fan-out uses NATS.** When a group message arrives at Pod 1 but recipients are connected to Pod 2 and Pod 3, NATS pub/sub distributes to all pods.
- **Chain failures don't stop messaging.** Circuit breakers isolate chain outages. Message relay continues with cached state. Anchoring and rewards queue for retry.
- **Auto-scaling triggers:** CPU > 70%, WebSocket connection count > 10K per pod, relay latency P99 > 200ms.

---

## Existing Implementation (Carried Forward from v1.0)

All v1.0 code remains valid:

**Tokenomics Package** (`internal/tokenomics/`):
- `models/token.go` — Token configuration, allocation, vesting
- `models/rewards.go` — Reward types, earning tracking, trust scores
- `emissions/schedule.go` — Emission calculations with halving
- `rewards/distributor.go` — Reward distribution and batching → now feeds into Metagraph Gateway
- `staking/staking.go` — Staking tiers and governance weight
- `governance/governance.go` — DAO proposals and voting

**iOS Swift Package** (`ios/Echo/`):
- `Sources/Models/` — Token.swift, Rewards.swift
- `Sources/Services/` — Service.swift, IdentityService.swift, MessagingService.swift
- `Tests/` — 34+ test cases across 5 suites

**Test Results:**
- Go unit tests: ✅ PASSING (50+ assertions)
- Swift unit tests: Ready (34+ test cases)

---

## Updated Project Structure

```
internal/
├── services/
│   ├── relay/                     # NEW: Message relay service
│   │   ├── relay.go              # Core relay logic
│   │   ├── offline_queue.go      # Redis/PG offline message queue
│   │   ├── connection_manager.go # WebSocket connection tracking
│   │   ├── sealed_sender.go      # Phase 3: sender-anonymous routing
│   │   └── relay_test.go
│   │
│   ├── identity/
│   │   └── identity.go           # User registration, DID management
│   │
│   ├── messaging/
│   │   └── messaging.go          # Conversation management (metadata only)
│   │
│   ├── trust/                     # NEW: Trust score service
│   │   └── trust.go              # Score computation, tier caching
│   │
│   ├── rewards/                   # NEW: Reward validation + batching
│   │   └── rewards.go            # Pre-validation, batch to Currency L1
│   │
│   ├── contacts/
│   │   └── contacts.go
│   │
│   └── registry.go               # Updated service registry (10 services)
│
├── metagraph/                     # NEW: Constellation integration
│   ├── gateway.go                # L1/L0 submission client
│   ├── anchoring.go              # Merkle tree batching + Data L1 submission
│   ├── snapshot_listener.go      # Snapshot event subscription
│   ├── merkle_tree.go            # Merkle tree implementation
│   └── metagraph_test.go
│
├── cardano/                       # NEW: Cardano integration
│   ├── cardano.go                # DID, credential, trust tier operations
│   ├── credential_cache.go       # Cached credential status (TTL 60s)
│   └── cardano_test.go
│
├── logging/                       # NEW: IPFS log publisher
│   ├── publisher.go              # Batch, encrypt, push to IPFS
│   └── publisher_test.go
│
├── infra/                         # NEW: Infrastructure utilities
│   ├── circuit_breaker.go        # Per-chain circuit breakers
│   ├── rate_limiter.go           # Per-DID rate limiting
│   └── infra_test.go
│
├── tokenomics/                    # Existing (unchanged)
│   ├── models/
│   ├── emissions/
│   ├── rewards/
│   ├── staking/
│   └── governance/
│
└── crypto/                        # NEW: Shared cryptographic utilities
    ├── commitments.go            # Hash commitment generation
    ├── hashing.go                # Argon2id, SHA-256
    └── crypto_test.go
```

---

## Next Steps (Updated)

### Immediate (Phase 1–2)
1. **Message Relay Service** — WebSocket relay, offline queue, APNs integration
2. **Metagraph Gateway** — Data L1 + Currency L1 submission pipeline
3. **Anchoring Batcher** — Merkle tree batching + submission
4. **Circuit Breakers** — Per-chain failure isolation
5. **Cardano Client** — DID registration, credential issuance

### Phase 3
6. **Sealed Sender** — Sender-anonymous routing
7. **Snapshot Listener** — Push on-chain confirmations to clients
8. **Merkle Proof Generation** — Client-verifiable anchoring proofs
9. **IPFS Log Publisher** — Encrypted audit trail

### Phase 4
10. **Federated Relay** — Multi-operator relay node registration on Data L1
11. **DAO Governance Service** — On-chain proposal/vote submission
12. **Optional P2P Signaling** — Direct WebSocket for both-online users

---

## Production Readiness Checklist

| Item | Status | Phase |
|------|--------|-------|
| Stateless Go pods behind load balancer | 🔲 | 1 |
| Redis with AOF persistence | 🔲 | 1 |
| PostgreSQL with synchronous replication (primary + 2 replicas) | 🔲 | 1 |
| NATS cluster for pub/sub | 🔲 | 1 |
| Circuit breakers per chain (metagraph, Cardano, IPFS) | 🔲 | 1 |
| Per-DID rate limiting | 🔲 | 1 |
| TLS 1.3 with certificate pinning | 🔲 | 1 |
| APNs push integration | 🔲 | 1 |
| Auto-scaling (CPU, connections, latency triggers) | 🔲 | 2 |
| Multi-region deployment | 🔲 | 2 |
| Kubernetes deployment specs | 🔲 | 2 |
| CI/CD pipeline (GitHub Actions) | 🔲 | 2 |
| Sealed sender | 🔲 | 3 |
| Merkle proof generation | 🔲 | 3 |
| Security audit (external) | 🔲 | 3 |
| Federated relay node support | 🔲 | 4 |

---

*Backend Architecture v2.0*
*Updated: February 23, 2026*
*Status: Aligned with relay architecture decision, Data Layer v3.0, and PRD v2.0*
