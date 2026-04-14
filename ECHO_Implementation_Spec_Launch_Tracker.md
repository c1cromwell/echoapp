# ECHO Implementation Specification & Launch Readiness Tracker
## Version 2.5.1 — Build-Ready Spec for VS Code Ingestion
### Generated: April 12, 2026 | Target Launch: June 1, 2026 (TestFlight Alpha)

---

# PART 1: LAUNCH READINESS REVIEW

## Feature Completeness Matrix

Every feature in the combined blueprint has been reviewed against the corrected PRD v2.5.1. Below is the build-readiness status for each.

### Foundation Layer — ALL READY TO BUILD ✅

| # | Foundation Component | Blueprint Lines | Status | Blocking Issues |
|---|---------------------|----------------|--------|-----------------|
| F1 | Go Backend (10 microservices) | 1–252 | ✅ Ready | None |
| F2 | iOS Frontend (SwiftUI/MVVM-C) | 253–1489 | ✅ Ready | None |
| F3 | Data Layer (Metagraph + Cardano + IPFS) | 1490–1954 | ✅ Ready | None |
| F4 | Secure Enclave Key Management | 1955–2163 | ✅ Ready | None |
| F5 | Privacy Architecture (T0–T7) | 2164–2316 | ✅ Ready | None |
| F6 | Tokenomics & Founder Allocation | 2317–2446 | ✅ Ready | None |
| F7 | Contact Discovery | 2447–2530 | ✅ Ready | None |

### Feature Layer — Phase Assignments

| # | Feature | Phase | Status | Notes |
|---|---------|-------|--------|-------|
| 1 | Decentralized Identity & Auth | 1–2 | ✅ Ready | Core onboarding flow |
| 2 | Blockchain-Anchored Messaging | 1–2 | ✅ Ready | E2E + Merkle anchoring |
| 3 | Dynamic Trust Network | 2 | ✅ Ready | Trust scoring + tiers |
| 4 | Voice/Video Calls | 2 | ✅ Ready | WebRTC via relay |
| 5 | File Sharing (2GB) | 2 | ✅ Ready | IPFS/Storj encrypted |
| 6 | Reactions, Polls, Interactive | 2 | ✅ Ready | Standard messaging UX |
| 7 | Advanced Search & Archive | 2–3 | ✅ Ready | Local encrypted search |
| 8 | Hidden Folders (Biometric) | 2 | ✅ Ready | Secure Enclave gated |
| 9 | Silent & Scheduled Messages | 2 | ✅ Ready | On-device timers |
| 10 | Disappearing Messages | 2 | ✅ Ready | Client-side deletion + relay cleanup |
| 11 | Public/Private Groups | 2 | ✅ Ready | Group key management |
| 12 | Multiple Personas | 3 | ✅ Ready | DID-linked personas |
| 13 | Broadcast Channels | 3 | ⚠️ Minor fix | One P2P ref → relay (line 5968) |
| 14 | Enterprise Org Profiles | 4–5 | ✅ Ready | Organization tier |
| 15 | Financial Institution Integration | 4–5 | ✅ Ready | Fraud prevention suite |
| 16 | Rewards Tracker on Profile | 2 | ✅ Ready | Gamification + wallet link |
| 17 | Streamlined Onboarding (VC + Passkeys) | 2 | ✅ Ready | OIDC4VC + WebAuthn |
| 18 | In-App ID Verification | 2 | ✅ Ready | IDV provider integration |
| 19 | Bot Framework | 4–5 | ✅ Ready | AllowSpend + trust scoring |
| 20 | Universal Onboarding (Username + Passkey → DID) | 1–2 | ✅ Ready | Zero-PII signup; phone is optional Tier 2 upgrade |
| 21 | E2E Encryption & Commitment | 1–2 | ✅ Ready | Kinnami + Merkle |
| 22 | Privacy-Preserving Blockchain Model | 1–2 | ✅ Ready | T0–T7 enforcement |
| 23 | ZK Proofs & Midnight | 3–4 | ✅ Ready | Phase-gated |
| 24 | Production Launch & Deployment | 2–3 | ✅ Ready | Full CI/CD spec |

### Remaining Fix (Apply During Build)

**Broadcast Channels (line 5968):** Replace "peer-to-peer network" with "encrypted relay network" in the channel content distribution description. This is a documentation-only fix; the architecture is already relay-based.

### Architecture Decision: Username + Passkey Onboarding (Zero PII)

The PRD's "Universal Onboarding" feature has been revised from phone-first to **username + passkey first**. Phone verification becomes an optional trust-tier upgrade, not a signup gate.

**Rationale:** ECHO's core promise is "privacy from everyone, including ECHO." Requiring a phone number at signup contradicts this — the user's first experience would be surrendering the strongest real-world identifier. Username + passkey onboarding collects zero PII, depends only on decentralized infrastructure (Cardano DID + iOS Secure Enclave), and is faster than every competitor (no SMS wait).

**Signup Flow (5 seconds, zero PII):**

```
1. User opens app → enters desired username
2. Backend checks username availability
3. App generates Secure Enclave key pair → creates passkey (Face ID binding)
4. Backend submits DID registration to Cardano (fee from platform treasury)
5. User lands in app at Tier 1 — can message immediately
6. Home screen shows progressive trust cards:
   "Verify phone → find friends + earn 1.2x"
   "Verify ID → earn 100 ECHO + unlock payments"
```

**Phone Verification (optional, Tier 1 → Tier 2 upgrade):**

```
1. User taps "Verify Phone" trust card
2. App collects phone number → sends to SMS gateway → OTP verification
3. On success:
   a. Phone number hashed on-device (Argon2id + per-user salt)
   b. Hash sent to backend → added to contact discovery index
   c. Raw phone number discarded from backend immediately
   d. Trust tier upgraded to Tier 2 (reward multiplier 1.0x → 1.2x)
4. Address book matching becomes available
5. User sees contacts already on ECHO
```

**Trust Tier Progression:**

| Tier | Method | PII Required | Reward Multiplier | Unlocks |
|------|--------|-------------|-------------------|---------|
| 1 — Unverified | Username + passkey | None | 1.0x | Basic messaging, rewards |
| 2 — Newcomer | Phone verification (opt-in) | Phone hash only (server) | 1.2x | Contact discovery, address book matching |
| 3 — Member | Third-party IDV (opt-in) | None to ECHO (IDV partner only) | 1.5x | Full rewards, group creation |
| 4 — Verified | Government ID + liveness | None to ECHO (IDV partner only) | 2.0x | Payment rails, financial features |
| 5 — Trusted | Peer attestations + sustained activity | None | 3.0x | Maximum multiplier, governance |

**Sybil Defense Without Phone Gates:**
- Secure Enclave hardware binding: 1 account per physical device (strongest defense)
- Tier 1 gets lowest multiplier (1.0x): farming unprofitable at scale
- Auto-scaling rate drops as network grows: mass fake accounts earn near-zero
- L1 anti-gaming validators: velocity checks, suspicious pattern detection
- Tier 3+ requires government ID: Sybil accounts can't pass liveness checks

**Impact on Blueprint Documents:**
- `identity/registration.go` — Accepts: username + public key. No phone required.
- `contacts/discovery.go` — Phone hash registration moves to separate `POST /identity/verify-phone` endpoint.
- iOS `Onboarding/` — Simplifies to: username → passkey → done (3 screens).
- New iOS `PhoneVerificationView` — Triggered from trust-tier card, not onboarding.
- Scala L1 validators — No change. Validators check DID + trust tier, not auth method.

---

## June 1, 2026 Scope Decision

**58 days remain.** The blueprint defines a 6–10 month full build. Here is the realistic scope:

| Milestone | Date | Scope |
|-----------|------|-------|
| **Phase 1 Complete** | April 18 | Testnet running, PoC demonstrated |
| **Phase 2 Alpha** | May 16 | Core messaging + DID + wallet on mainnet metagraph |
| **TestFlight Alpha** | June 1 | 100–500 users via TestFlight |
| **Phase 3 Beta** | August 2026 | 1K–10K users, App Store submission |
| **App Store Launch** | September 2026 | Public soft launch |

### Phase 1–2 Feature Scope (Build Now)

These features MUST be implemented for the June 1 TestFlight alpha:

| Priority | Feature | Reason |
|----------|---------|--------|
| P0 | Secure Enclave key generation + biometric auth | Foundation for everything |
| P0 | DID creation on Cardano testnet/mainnet | User identity |
| P0 | E2E encrypted messaging (Kinnami) | Core product |
| P0 | WebSocket relay + offline queue | Message delivery |
| P0 | Commitment hashing + Merkle batching | Integrity anchoring |
| P0 | ECHO token genesis + wallet display | Token economy |
| P0 | Trust scoring (tiers 1–3) | Reward multipliers |
| P0 | Username + passkey onboarding (zero PII) | User registration |
| P1 | Reward claiming (auto-scaled) | Token incentives |
| P1 | Group messaging (symmetric key) | Core messaging |
| P1 | Contact discovery (QR + username + invite links) | Find friends |
| P1 | Phone verification (optional Tier 2 upgrade) | Contact matching + higher rewards |
| P1 | Push notifications (APNs) | Offline alerts |
| P1 | Hidden folders (biometric) | Privacy feature |
| P1 | Disappearing messages | Privacy feature |
| P2 | Voice notes | Messaging UX |
| P2 | Reactions + replies | Messaging UX |
| P2 | File sharing (encrypted, up to 2GB) | Media sharing |
| P2 | Profile + rewards tracker | Gamification |
| P2 | Staking (TokenLock) | Token economy |
| P2 | Delegation (StakeDelegation) | Validator support |

### Deferred to Phase 3+ (Do NOT Build Now)

| Feature | Phase | Reason |
|---------|-------|--------|
| Voice/video calls | 3 | WebRTC complexity; messaging first |
| Sealed sender | 3 | Metadata protection enhancement |
| Midnight ZK proofs | 3–4 | Midnight mainnet maturity required |
| Multiple personas | 3 | UX complexity |
| Broadcast channels | 3 | Scale feature |
| PacaSwap swap/bridge | 3 | DEX integration |
| Enterprise org profiles | 4–5 | Requires consumer traction first |
| Financial institution integration | 4–5 | Requires enterprise pilots |
| Bot framework | 4–5 | Ecosystem maturity |
| VIP subscriptions | 5 | Revenue phase |
| AI treasury agents | 5 | Economy phase |
| Network State | 6 | Long-term vision |

---

# PART 2: GO BACKEND IMPLEMENTATION SPEC

## Directory Structure

```
echo-backend/
├── cmd/
│   ├── gateway/main.go          # Port 8000 — API gateway
│   ├── identity/main.go         # Port 8001 — Identity service
│   ├── relay/main.go            # Port 8002 — Message relay
│   ├── trust/main.go            # Port 8003 — Trust scoring
│   ├── rewards/main.go          # Port 8004 — Reward engine
│   ├── contacts/main.go         # Port 8005 — Contact discovery
│   ├── metagraph/main.go        # Port 8006 — Metagraph gateway
│   ├── notifications/main.go   # Port 8007 — Push notifications
│   ├── media/main.go            # Port 8008 — Media service
│   └── logpub/main.go           # Port 8009 — Log publisher
│
├── internal/
│   ├── auth/
│   │   ├── passkey.go           # ECDSA P-256 signature verification
│   │   ├── middleware.go        # Auth middleware for all services
│   │   └── did_cache.go         # DID document cache (TTL: 60s)
│   │
│   ├── relay/
│   │   ├── hub.go               # WebSocket connection manager
│   │   ├── handler.go           # Message routing logic
│   │   ├── offline_queue.go     # Redis/PG offline message store
│   │   ├── overflow_backup.go   # IPFS overflow for >1000 msgs
│   │   ├── fan_out.go           # NATS-based group fan-out
│   │   └── rate_limiter.go      # Per-DID API rate limiting
│   │
│   ├── identity/
│   │   ├── registration.go      # Username + passkey DID creation (zero PII)
│   │   ├── username_check.go    # Username availability check
│   │   ├── phone_verify.go      # Optional phone OTP → Tier 2 upgrade
│   │   ├── credentials.go       # Verifiable credential management
│   │   ├── idv_provider.go      # Third-party IDV coordination (Tier 3-4)
│   │   └── trust_tier.go        # Trust tier computation + upgrades
│   │
│   ├── contacts/
│   │   ├── discovery.go         # Argon2id hash matching
│   │   ├── qr_exchange.go       # QR code DID exchange
│   │   ├── username_search.go   # Public handle search
│   │   └── invite_links.go      # Referral link generation
│   │
│   ├── rewards/
│   │   ├── auto_scale.go        # Network-wide auto-scaling engine
│   │   ├── claim_validator.go   # Annual budget enforcement
│   │   ├── batch_processor.go   # AtomicAction batch construction
│   │   └── anti_gaming.go       # Velocity + pattern detection
│   │
│   ├── anchoring/
│   │   ├── merkle_builder.go    # Merkle tree construction
│   │   ├── batch_submitter.go   # Data L1 submission pipeline
│   │   └── confirmation.go      # Snapshot listener + WS push
│   │
│   ├── metagraph/
│   │   ├── client.go            # Metagraph REST API client
│   │   ├── currency_l1.go       # Token transaction submission
│   │   ├── data_l1.go           # Data submission pipeline
│   │   ├── snapshot_listener.go # Real-time snapshot events
│   │   └── circuit_breaker.go   # Per-chain circuit breakers
│   │
│   ├── cardano/
│   │   ├── client.go            # Cardano API client
│   │   ├── did_registry.go      # DID document operations
│   │   ├── credential_issuer.go # Credential issuance
│   │   └── trust_datum.go       # Trust tier UTXO management
│   │
│   ├── storage/
│   │   ├── ipfs_client.go       # IPFS upload/pin/retrieve
│   │   ├── storj_client.go      # Storj fallback
│   │   └── log_publisher.go     # Encrypted log batching
│   │
│   ├── enterprise/              # Phase 4-5
│   │   ├── fraud_alerts.go      # Transaction verification
│   │   ├── fraud_dashboard.go   # Analytics aggregation
│   │   └── fraud_zk.go          # Cross-org ZK intelligence
│   │
│   ├── zk/                      # Phase 3+
│   │   ├── midnight_client.go   # Midnight API integration
│   │   ├── proof_verifier.go    # ZK proof verification
│   │   └── cache.go             # Verification result cache
│   │
│   └── common/
│       ├── config.go            # Environment configuration
│       ├── logger.go            # Privacy-safe structured logging
│       ├── errors.go            # Error code definitions
│       ├── models.go            # Shared data models
│       └── crypto.go            # SHA-256, HKDF, AES-256-GCM
│
├── api/
│   ├── openapi.yaml             # OpenAPI 3.1 specification
│   └── proto/                   # gRPC proto files (internal)
│
├── deployments/
│   ├── docker-compose.yml       # Local development
│   ├── k8s/
│   │   ├── gateway.yaml
│   │   ├── identity.yaml
│   │   ├── relay.yaml
│   │   ├── trust.yaml
│   │   ├── rewards.yaml
│   │   ├── contacts.yaml
│   │   ├── metagraph.yaml
│   │   ├── notifications.yaml
│   │   ├── media.yaml
│   │   ├── logpub.yaml
│   │   └── hpa.yaml             # Horizontal pod autoscalers
│   ├── terraform/
│   │   ├── hetzner/             # Phase 1-3 (primary)
│   │   ├── ovhcloud/            # Phase 3+ (secondary/failover)
│   │   └── akash/               # Phase 4+ (community relay evaluation)
│   └── github-actions/
│       ├── ci.yml               # Test + lint on every PR
│       ├── staging.yml          # Auto-deploy to staging
│       └── production.yml       # Manual approval gate
│
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_offline_queue.sql
│   ├── 003_contact_discovery.sql
│   └── 004_reward_tracking.sql
│
├── scripts/
│   ├── seed_testnet.sh          # Seed metagraph testnet data
│   └── generate_genesis.sh      # Token genesis block tool
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Service Implementation Order (Build Sequence)

```
Week 1-2: Foundation
  1. common/ (config, logger, errors, models, crypto)
  2. auth/ (passkey verification, middleware)
  3. metagraph/client.go + circuit_breaker.go
  4. cardano/client.go + did_registry.go

Week 3-4: Core Messaging
  5. relay/hub.go (WebSocket connection manager)
  6. relay/handler.go (message routing)
  7. relay/offline_queue.go (Redis + PG fallback)
  8. relay/fan_out.go (NATS group distribution)
  9. notifications/ (APNs push)

Week 5-6: Identity + Trust
  10. identity/registration.go (username + passkey → DID creation)
  11. identity/username_check.go (availability check)
  12. identity/phone_verify.go (optional SMS OTP → Tier 2)
  13. identity/credentials.go (VC management)
  14. identity/idv_provider.go (third-party IDV for Tier 3-4)
  15. identity/trust_tier.go (score computation + tier upgrades)
  16. contacts/discovery.go (Argon2id matching — requires Tier 2)
  17. contacts/qr_exchange.go + username_search.go + invite_links.go

Week 7-8: Token Economy
  16. metagraph/currency_l1.go (token transactions)
  17. metagraph/data_l1.go (data submissions)
  18. rewards/auto_scale.go (auto-scaling engine)
  19. rewards/claim_validator.go (budget enforcement)
  20. rewards/batch_processor.go (AtomicAction batches)

Week 9-10: Integrity + Storage
  21. anchoring/merkle_builder.go (Merkle trees)
  22. anchoring/batch_submitter.go (Data L1 pipeline)
  23. anchoring/confirmation.go (snapshot → WS push)
  24. metagraph/snapshot_listener.go
  25. storage/ (IPFS + Storj + log publisher)
  26. media/ (encrypted media upload/download)

Week 11-12: Integration + Testing
  27. gateway/ (load balancer, TLS, rate limiting)
  28. End-to-end integration tests
  29. Load testing (1000 concurrent WebSocket connections)
  30. Security review preparation
```

## Core API Routes

### Gateway (port 8000)
All routes are prefixed with `/api/v1/` and require passkey auth header.

### Identity Service (port 8001)
```
POST   /identity/register          # Username + public key → create DID on Cardano (zero PII)
GET    /identity/username/:name    # Check username availability
POST   /identity/verify-phone      # Optional SMS OTP → Tier 2 upgrade + contact discovery
POST   /identity/verify            # Submit IDV verification result (Tier 3-4)
GET    /identity/did/:did          # Retrieve DID document (cached)
GET    /identity/credentials/:did  # List credentials for DID
POST   /identity/credentials       # Issue new credential
DELETE /identity/credentials/:id   # Revoke credential
GET    /identity/trust-tier/:did   # Get trust tier (cached 60s)
```

### Message Relay (port 8002)
```
WS     /relay/ws                   # WebSocket connection (sticky)
POST   /relay/send                 # REST fallback for message send
GET    /relay/queue/:did           # Check offline queue depth
GET    /relay/anchoring/:msgId     # Check anchoring status
```

### Trust Service (port 8003)
```
GET    /trust/score/:did           # Get trust score + tier
POST   /trust/report               # Report user (on-chain evidence)
POST   /trust/block                # Block user
GET    /trust/circles/:did         # Get trusted circles
```

### Rewards Service (port 8004)
```
POST   /rewards/claim              # Claim rewards (AtomicAction)
GET    /rewards/pending/:did       # Get pending rewards
GET    /rewards/history/:did       # Reward claim history
GET    /rewards/network-rate       # Current auto-scaled rate
GET    /rewards/emission-status    # Annual budget status
```

### Contacts Service (port 8005)
```
POST   /contacts/discover          # Phone hash matching (requires Tier 2+, rate: 1/24h)
POST   /contacts/register-phone    # Add verified phone hash to discovery index
POST   /contacts/qr-connect        # QR code DID exchange
GET    /contacts/search             # Username search (?handle=...)
POST   /contacts/invite            # Generate referral link
GET    /contacts/list/:did         # Get contact list
POST   /contacts/block             # Block contact
DELETE /contacts/block/:did        # Unblock contact
```

### Metagraph Gateway (port 8006)
```
POST   /metagraph/currency/submit  # Submit currency transaction
POST   /metagraph/data/submit      # Submit data transaction
GET    /metagraph/snapshot/latest   # Latest snapshot info
GET    /metagraph/balance/:did     # Token balance (cached 5s)
GET    /metagraph/staking/:did     # Staking positions
GET    /metagraph/delegation/:did  # Delegation info
POST   /metagraph/atomic           # Submit AtomicAction bundle
```

### Media Service (port 8008)
```
POST   /media/upload               # Upload encrypted media blob
GET    /media/:id                  # Download encrypted media blob
DELETE /media/:id                  # Delete media (owner only)
POST   /media/evidence             # Digital Evidence fingerprint
```

## Key Implementation: Auto-Scaling Reward Engine

```go
// internal/rewards/auto_scale.go

package rewards

import (
    "context"
    "math/big"
    "sync"
    "time"
)

// EmissionSchedule defines the 10-year declining emission curve
var EmissionSchedule = map[int]int64{
    1: 80_000_000,  // Year 1: 80M ECHO (20% of 400M community pool)
    2: 64_000_000,  // Year 2: 64M (16%)
    3: 52_000_000,  // Year 3: 52M (13%)
    4: 44_000_000,  // Year 4: 44M (11%)
    5: 36_000_000,  // Year 5: 36M (9%)
    6: 24_000_000,  // Years 6-10: 24M each (6%)
    7: 24_000_000, 8: 24_000_000, 9: 24_000_000, 10: 24_000_000,
}

// TrustTierRewardMultiplier maps trust tiers to reward multipliers
// NOTE: These are REWARD multipliers (1.0-3.0), NOT governance multipliers (0.0-2.0)
var TrustTierRewardMultiplier = map[int]float64{
    1: 1.0, // Unverified
    2: 1.2, // Newcomer
    3: 1.5, // Member
    4: 2.0, // Verified
    5: 3.0, // Trusted
}

type AutoScaleEngine struct {
    mu                  sync.RWMutex
    currentYear         int
    annualBudget        int64
    distributedToday    int64
    networkActivityToday float64 // Sum of all message weights today
    dailyBudget         int64
    lastResetDate       time.Time
}

func NewAutoScaleEngine(currentYear int) *AutoScaleEngine {
    annual := EmissionSchedule[currentYear]
    return &AutoScaleEngine{
        currentYear:  currentYear,
        annualBudget: annual,
        dailyBudget:  annual / 365,
    }
}

// CalculateReward computes the auto-scaled reward for a single message.
// Rate = DailyBudget / TotalDailyNetworkActivityWeight
// No per-user cap — every message earns something.
func (e *AutoScaleEngine) CalculateReward(
    senderTrustTier int,
) (rewardAmount int64, currentRate float64) {
    e.mu.RLock()
    defer e.mu.RUnlock()

    multiplier := TrustTierRewardMultiplier[senderTrustTier]
    activityWeight := 1.0 * multiplier // Each message = 1 × tier multiplier

    if e.networkActivityToday == 0 {
        currentRate = 0.1 // Target baseline rate when no activity yet
    } else {
        currentRate = float64(e.dailyBudget) / e.networkActivityToday
    }

    reward := currentRate * activityWeight
    return int64(reward * 1e8), currentRate // 8 decimal places
}

// RecordActivity tracks a message's contribution to daily network weight.
// Called by the Rewards Service after successful message relay confirmation.
func (e *AutoScaleEngine) RecordActivity(trustTier int) {
    e.mu.Lock()
    defer e.mu.Unlock()

    multiplier := TrustTierRewardMultiplier[trustTier]
    e.networkActivityToday += 1.0 * multiplier

    // Reset at midnight UTC
    now := time.Now().UTC()
    if now.Day() != e.lastResetDate.Day() {
        e.distributedToday = 0
        e.networkActivityToday = 0
        e.lastResetDate = now
    }
}

// GetNetworkStatus returns current emission status for public API
func (e *AutoScaleEngine) GetNetworkStatus() NetworkStatus {
    e.mu.RLock()
    defer e.mu.RUnlock()

    rate := float64(0.1)
    if e.networkActivityToday > 0 {
        rate = float64(e.dailyBudget) / e.networkActivityToday
    }

    return NetworkStatus{
        CurrentYear:          e.currentYear,
        AnnualBudget:         e.annualBudget,
        DailyBudget:          e.dailyBudget,
        DistributedToday:     e.distributedToday,
        NetworkActivityToday: e.networkActivityToday,
        CurrentPerMessageRate: rate,
    }
}

type NetworkStatus struct {
    CurrentYear          int     `json:"currentYear"`
    AnnualBudget         int64   `json:"annualBudget"`
    DailyBudget          int64   `json:"dailyBudget"`
    DistributedToday     int64   `json:"distributedToday"`
    NetworkActivityToday float64 `json:"networkActivityToday"`
    CurrentPerMessageRate float64 `json:"currentPerMessageRate"`
}
```

## Key Implementation: Contact Discovery

```go
// internal/contacts/discovery.go

package contacts

import (
    "context"
    "crypto/rand"
    "encoding/hex"

    "golang.org/x/crypto/argon2"
)

const (
    argon2Time    = 3
    argon2Memory  = 64 * 1024 // 64MB
    argon2Threads = 4
    argon2KeyLen  = 32
)

type DiscoveryService struct {
    db    *sql.DB
    redis *redis.Client
}

// HashPhoneNumber hashes a phone number with Argon2id using the user's salt.
// This function runs ON THE CLIENT — the server never sees raw numbers.
// Included here for reference and test parity only.
func HashPhoneNumber(phone string, salt []byte) string {
    hash := argon2.IDKey(
        []byte(phone), salt,
        argon2Time, argon2Memory, argon2Threads, argon2KeyLen,
    )
    return hex.EncodeToString(hash)
}

type DiscoveryRequest struct {
    HashedNumbers []string `json:"hashedNumbers"` // Argon2id hashes from client
    RequesterDID  string   `json:"requesterDID"`
}

type DiscoveryMatch struct {
    HashedNumber   string `json:"hashedNumber"`
    EncryptedDID   string `json:"encryptedDID"`   // Encrypted DID reference
    TrustTier      int    `json:"trustTier"`
    DisplayName    string `json:"displayName"`     // Public display name only
}

// MatchContacts checks hashed phone numbers against the server-side index.
// Requires Tier 2+ (phone-verified). Rate limited: 1 request per 24 hours per DID.
// Server index stores only: Argon2id hashes → encrypted DID references.
func (s *DiscoveryService) MatchContacts(
    ctx context.Context,
    req DiscoveryRequest,
) ([]DiscoveryMatch, error) {
    // 1. Rate limit check
    if !s.checkRateLimit(req.RequesterDID) {
        return nil, ErrRateLimited
    }

    // 2. Query server-side hash index
    matches := make([]DiscoveryMatch, 0)
    for _, hash := range req.HashedNumbers {
        var match DiscoveryMatch
        err := s.db.QueryRowContext(ctx,
            `SELECT encrypted_did, trust_tier, display_name 
             FROM contact_discovery_index 
             WHERE phone_hash = $1`, hash,
        ).Scan(&match.EncryptedDID, &match.TrustTier, &match.DisplayName)

        if err == nil {
            match.HashedNumber = hash
            matches = append(matches, match)
        }
    }

    return matches, nil
}

// RegisterForDiscovery adds a user's hashed phone to the discovery index.
// Called AFTER voluntary phone verification (Tier 2 upgrade), NOT during onboarding.
// The user must have completed POST /identity/verify-phone successfully.
func (s *DiscoveryService) RegisterForDiscovery(
    ctx context.Context,
    phoneHash string,
    encryptedDID string,
    displayName string,
) error {
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO contact_discovery_index 
         (phone_hash, encrypted_did, display_name, created_at)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (phone_hash) DO UPDATE SET encrypted_did = $2`,
        phoneHash, encryptedDID, displayName,
    )
    return err
}
```

## Key Implementation: Zero-PII Registration

```go
// internal/identity/registration.go

package identity

import (
    "context"
    "fmt"
    "regexp"
    "strings"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,24}$`)

type RegistrationRequest struct {
    Username  string `json:"username"`   // 3-24 chars, alphanumeric + underscore
    PublicKey []byte `json:"publicKey"`  // P-256 public key from Secure Enclave
}

type RegistrationResponse struct {
    DID       string `json:"did"`       // did:prism:<hash>
    Username  string `json:"username"`
    TrustTier int    `json:"trustTier"` // Always 1 at registration
}

// Register creates a new ECHO account with zero PII.
// Accepts: username + Secure Enclave public key. Nothing else.
// No phone number, no email, no real name collected.
func (s *IdentityService) Register(
    ctx context.Context,
    req RegistrationRequest,
) (*RegistrationResponse, error) {

    // 1. Validate username format
    if !usernameRegex.MatchString(req.Username) {
        return nil, ErrInvalidUsername
    }

    // 2. Check username availability
    taken, err := s.isUsernameTaken(ctx, req.Username)
    if err != nil {
        return nil, fmt.Errorf("username check failed: %w", err)
    }
    if taken {
        return nil, ErrUsernameTaken
    }

    // 3. Register DID on Cardano (fee from platform treasury)
    did, err := s.cardano.RegisterDID(ctx, req.PublicKey)
    if err != nil {
        return nil, fmt.Errorf("DID registration failed: %w", err)
    }

    // 4. Store username → DID mapping
    err = s.db.ExecContext(ctx,
        `INSERT INTO users (did, username, trust_tier, created_at)
         VALUES ($1, $2, 1, NOW())`,
        did, strings.ToLower(req.Username),
    )
    if err != nil {
        return nil, fmt.Errorf("user creation failed: %w", err)
    }

    // 5. Cache DID document
    s.didCache.Set(did, req.PublicKey, 60) // TTL 60s

    return &RegistrationResponse{
        DID:       did,
        Username:  req.Username,
        TrustTier: 1, // Unverified — can message, lowest reward multiplier
    }, nil
}
```

## Key Implementation: Optional Phone Verification (Tier 2 Upgrade)

```go
// internal/identity/phone_verify.go

package identity

import (
    "context"
    "crypto/rand"
    "fmt"
    "math/big"
    "time"
)

type PhoneVerifyRequest struct {
    PhoneNumber string `json:"phoneNumber"` // E.164 format
    DID         string `json:"did"`
}

type PhoneVerifyConfirm struct {
    DID  string `json:"did"`
    Code string `json:"code"` // 6-digit OTP
}

// InitiatePhoneVerification sends an SMS OTP to the user's phone.
// This is OPTIONAL — called only when user taps "Verify Phone" trust card.
// The phone number is used only for OTP delivery and is NOT stored by the backend.
func (s *IdentityService) InitiatePhoneVerification(
    ctx context.Context,
    req PhoneVerifyRequest,
) error {
    // 1. Generate 6-digit OTP
    code, err := generateOTP()
    if err != nil {
        return err
    }

    // 2. Store OTP in Redis with 5-minute TTL (keyed by DID, not phone)
    s.redis.Set(ctx, fmt.Sprintf("otp:%s", req.DID), code, 5*time.Minute)

    // 3. Send SMS via provider (Twilio, etc.)
    // Phone number is passed directly to SMS gateway — NOT stored in our DB
    err = s.smsGateway.SendOTP(req.PhoneNumber, code)
    if err != nil {
        return fmt.Errorf("SMS delivery failed: %w", err)
    }

    return nil
}

// ConfirmPhoneVerification validates the OTP and upgrades trust tier to 2.
// On success: phone number is hashed on-device by the iOS app, hash sent to
// POST /contacts/register-phone for discovery index. Raw number never stored.
func (s *IdentityService) ConfirmPhoneVerification(
    ctx context.Context,
    req PhoneVerifyConfirm,
) (*TrustTierUpgrade, error) {
    // 1. Check OTP
    stored, err := s.redis.Get(ctx, fmt.Sprintf("otp:%s", req.DID)).Result()
    if err != nil || stored != req.Code {
        return nil, ErrInvalidOTP
    }

    // 2. Delete OTP (one-time use)
    s.redis.Del(ctx, fmt.Sprintf("otp:%s", req.DID))

    // 3. Upgrade trust tier to 2
    err = s.upgradeTrustTier(ctx, req.DID, 2, "phone_verification")
    if err != nil {
        return nil, err
    }

    // 4. Issue phone-verified credential on Cardano
    err = s.cardano.IssueCredential(ctx, req.DID, "phone_verified", nil)
    if err != nil {
        // Non-fatal: tier upgrade succeeded, credential issuance can retry
        s.logger.Warn("credential issuance deferred", "did", req.DID, "err", err)
    }

    return &TrustTierUpgrade{
        DID:             req.DID,
        NewTier:         2,
        RewardMultiplier: 1.2,
        UnlockedFeatures: []string{
            "contact_discovery",
            "address_book_matching",
            "enhanced_rewards",
        },
    }, nil
}

type TrustTierUpgrade struct {
    DID              string   `json:"did"`
    NewTier          int      `json:"newTier"`
    RewardMultiplier float64  `json:"rewardMultiplier"`
    UnlockedFeatures []string `json:"unlockedFeatures"`
}

func generateOTP() (string, error) {
    n, err := rand.Int(rand.Reader, big.NewInt(999999))
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%06d", n.Int64()), nil
}
```

## Key Implementation: Merkle Batching

```go
// internal/anchoring/merkle_builder.go

package anchoring

import (
    "crypto/sha256"
    "encoding/hex"
    "sync"
    "time"
)

type MerkleBatcher struct {
    mu          sync.Mutex
    commitments [][]byte
    batchStart  time.Time
    maxBatchSize int           // 1000 commitments
    maxBatchAge  time.Duration // 5 minutes
    submitter   DataL1Submitter
}

func NewMerkleBatcher(submitter DataL1Submitter) *MerkleBatcher {
    b := &MerkleBatcher{
        commitments:  make([][]byte, 0, 1000),
        batchStart:   time.Now(),
        maxBatchSize: 1000,
        maxBatchAge:  5 * time.Minute,
        submitter:    submitter,
    }
    go b.flushLoop()
    return b
}

// AddCommitment adds a message commitment hash to the current batch.
// Called by relay/handler.go after message is accepted.
func (b *MerkleBatcher) AddCommitment(commitment []byte) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.commitments = append(b.commitments, commitment)

    if len(b.commitments) >= b.maxBatchSize {
        go b.flush()
    }
}

func (b *MerkleBatcher) flushLoop() {
    ticker := time.NewTicker(b.maxBatchAge)
    for range ticker.C {
        b.mu.Lock()
        if len(b.commitments) > 0 {
            b.flush()
        }
        b.mu.Unlock()
    }
}

func (b *MerkleBatcher) flush() {
    if len(b.commitments) == 0 {
        return
    }

    batch := make([][]byte, len(b.commitments))
    copy(batch, b.commitments)
    b.commitments = make([][]byte, 0, 1000)
    batchStart := b.batchStart
    b.batchStart = time.Now()

    // Build Merkle tree and submit root to Data L1
    root := BuildMerkleRoot(batch)
    b.submitter.SubmitMerkleRoot(DataL1Submission{
        Type:            "message_integrity",
        MerkleRoot:      root,
        CommitmentCount: len(batch),
        TimeRange: TimeRange{
            From: batchStart,
            To:   time.Now(),
        },
        SchemaVersion: 1,
    })
}

// BuildMerkleRoot constructs a binary Merkle tree and returns the root hash.
func BuildMerkleRoot(leaves [][]byte) []byte {
    if len(leaves) == 0 {
        return nil
    }
    if len(leaves) == 1 {
        return leaves[0]
    }

    // Pad to even number
    if len(leaves)%2 != 0 {
        leaves = append(leaves, leaves[len(leaves)-1])
    }

    var nextLevel [][]byte
    for i := 0; i < len(leaves); i += 2 {
        combined := append(leaves[i], leaves[i+1]...)
        hash := sha256.Sum256(combined)
        nextLevel = append(nextLevel, hash[:])
    }

    return BuildMerkleRoot(nextLevel)
}
```

## PostgreSQL Schema (Phase 1-2)

```sql
-- migrations/001_initial_schema.sql

-- Users (zero PII — only username + DID + trust tier)
CREATE TABLE users (
    did             VARCHAR(128) PRIMARY KEY,
    username        VARCHAR(24) UNIQUE NOT NULL,
    trust_tier      SMALLINT NOT NULL DEFAULT 1,  -- 1=Unverified, 2=Newcomer, 3=Member, 4=Verified, 5=Trusted
    phone_verified  BOOLEAN DEFAULT FALSE,
    idv_verified    BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_users_username ON users(LOWER(username));

-- Offline message queue (encrypted blobs only)
CREATE TABLE offline_queue (
    id              BIGSERIAL PRIMARY KEY,
    recipient_did   VARCHAR(128) NOT NULL,
    sender_did      VARCHAR(128) NOT NULL,
    encrypted_blob  BYTEA NOT NULL,
    commitment_hash BYTEA NOT NULL,
    signature       BYTEA NOT NULL,
    content_type    VARCHAR(32) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    delivered       BOOLEAN DEFAULT FALSE,
    overflow_cid    VARCHAR(128)  -- IPFS CID for overflow backup
);
CREATE INDEX idx_offline_queue_recipient ON offline_queue(recipient_did, delivered);
CREATE INDEX idx_offline_queue_expires ON offline_queue(expires_at);

-- Contact discovery index (hashes only — no raw phone numbers)
CREATE TABLE contact_discovery_index (
    phone_hash      VARCHAR(64) PRIMARY KEY,  -- Argon2id hash
    encrypted_did   VARCHAR(256) NOT NULL,     -- Encrypted DID reference
    display_name    VARCHAR(64),               -- Public display name
    trust_tier      SMALLINT DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Reward tracking (for auto-scaling engine)
CREATE TABLE reward_claims (
    id              BIGSERIAL PRIMARY KEY,
    claimer_did     VARCHAR(128) NOT NULL,
    claim_type      VARCHAR(32) NOT NULL,      -- 'messaging', 'referral', 'staking'
    amount          BIGINT NOT NULL,            -- In smallest unit (8 decimals)
    trust_tier      SMALLINT NOT NULL,
    multiplier      DECIMAL(4,2) NOT NULL,
    auto_scale_rate DECIMAL(20,8) NOT NULL,     -- Rate at time of claim
    batch_id        VARCHAR(64),                -- AtomicAction batch reference
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submitted_at    TIMESTAMPTZ,
    confirmed_at    TIMESTAMPTZ
);
CREATE INDEX idx_reward_claims_did ON reward_claims(claimer_did, created_at);
CREATE INDEX idx_reward_claims_date ON reward_claims(created_at);

-- Network activity tracking (for auto-scaling)
CREATE TABLE daily_network_activity (
    date            DATE PRIMARY KEY,
    total_messages  BIGINT DEFAULT 0,
    total_weight    DECIMAL(20,4) DEFAULT 0,    -- Sum of tier-weighted activity
    total_distributed BIGINT DEFAULT 0,
    daily_budget    BIGINT NOT NULL
);

-- Merkle batch tracking
CREATE TABLE merkle_batches (
    id              BIGSERIAL PRIMARY KEY,
    merkle_root     BYTEA NOT NULL,
    commitment_count INT NOT NULL,
    time_range_from TIMESTAMPTZ NOT NULL,
    time_range_to   TIMESTAMPTZ NOT NULL,
    submitted_at    TIMESTAMPTZ,
    confirmed_at    TIMESTAMPTZ,
    snapshot_hash   VARCHAR(128),
    snapshot_height BIGINT
);

-- Referral tracking
CREATE TABLE referrals (
    id              BIGSERIAL PRIMARY KEY,
    referrer_did    VARCHAR(128) NOT NULL,
    referee_did     VARCHAR(128) NOT NULL,
    invite_code     VARCHAR(32) UNIQUE NOT NULL,
    tier_depth      SMALLINT DEFAULT 1,         -- Max 3 tiers
    status          VARCHAR(16) DEFAULT 'pending', -- pending, verified, rewarded
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at     TIMESTAMPTZ,
    rewarded_at     TIMESTAMPTZ
);
CREATE INDEX idx_referrals_referrer ON referrals(referrer_did);
```

---

# PART 3: DATA LAYER INTEGRATION SPEC

## Scala Metagraph Implementation (Euclid SDK)

### Project Structure

```
echo-metagraph/
├── modules/
│   ├── l0/                          # Metagraph L0 (snapshot aggregation)
│   │   └── src/main/scala/
│   │       └── echo/l0/
│   │           └── EchoMetagraphL0.scala
│   │
│   ├── currency-l1/                  # Currency L1 (token validation)
│   │   └── src/main/scala/
│   │       └── echo/currency/
│   │           ├── EchoCurrencyL1.scala
│   │           ├── validators/
│   │           │   ├── RewardClaimValidator.scala
│   │           │   ├── StakingValidator.scala
│   │           │   ├── EmissionBudgetValidator.scala
│   │           │   └── AntiGamingValidator.scala
│   │           └── models/
│   │               ├── RewardClaim.scala
│   │               ├── StakingOperation.scala
│   │               └── EmissionState.scala
│   │
│   └── data-l1/                      # Data L1 (application data)
│       └── src/main/scala/
│           └── echo/data/
│               ├── EchoDataL1.scala
│               ├── validators/
│               │   ├── MerkleRootValidator.scala
│               │   ├── TrustCommitmentValidator.scala
│               │   ├── GovernanceVoteValidator.scala
│               │   └── SchemaVersionValidator.scala
│               └── models/
│                   ├── MerkleRootSubmission.scala
│                   ├── TrustCommitment.scala
│                   └── GovernanceVote.scala
│
├── build.sbt
├── project/plugins.sbt
└── docker-compose.yml               # Local Euclid SDK cluster
```

### Critical Validation: Emission Budget Enforcement

```scala
// modules/currency-l1/src/main/scala/echo/currency/validators/EmissionBudgetValidator.scala

package echo.currency.validators

import org.tessellation.currency.l1.domain.dataApplication._
import java.time.{LocalDate, Year}

object EmissionBudgetValidator {

  // 10-year declining emission schedule (community pool: 400M total)
  val annualBudget: Map[Int, Long] = Map(
    1 -> 80_000_000L,   // Year 1: 80M (20%)
    2 -> 64_000_000L,   // Year 2: 64M (16%)
    3 -> 52_000_000L,   // Year 3: 52M (13%)
    4 -> 44_000_000L,   // Year 4: 44M (11%)
    5 -> 36_000_000L,   // Year 5: 36M (9%)
    6 -> 24_000_000L, 7 -> 24_000_000L, 8 -> 24_000_000L,
    9 -> 24_000_000L, 10 -> 24_000_000L
  )

  // Trust tier REWARD multipliers (NOT governance multipliers)
  val rewardMultiplier: Map[Int, Double] = Map(
    1 -> 1.0, 2 -> 1.2, 3 -> 1.5, 4 -> 2.0, 5 -> 3.0
  )

  /**
   * Validates a reward claim against the annual emission budget.
   * 
   * KEY DESIGN: No per-user daily cap. The auto-scaled rate ensures
   * annual budget is never exceeded while every message always earns.
   * 
   * Validation rules:
   * 1. Reject if Year-N total distributions would exceed Year-N budget
   * 2. Verify trust tier multiplier matches claimed tier
   * 3. Verify auto-scale rate is within acceptable bounds
   * 4. Anti-gaming: reject suspicious velocity patterns
   */
  def validate(
    claim: RewardClaim,
    currentYearDistributed: Long,
    currentYear: Int,
    senderTrustTier: Int
  ): Either[ValidationError, ValidatedClaim] = {
    
    val yearBudget = annualBudget.getOrElse(currentYear, 0L)
    
    // Rule 1: Annual budget enforcement
    if (currentYearDistributed + claim.amount > yearBudget) {
      return Left(ValidationError("EMISSION_BUDGET_EXCEEDED",
        s"Year $currentYear budget: $yearBudget, distributed: $currentYearDistributed"))
    }
    
    // Rule 2: Trust tier multiplier validation
    val expectedMultiplier = rewardMultiplier.getOrElse(senderTrustTier, 1.0)
    if (math.abs(claim.appliedMultiplier - expectedMultiplier) > 0.001) {
      return Left(ValidationError("MULTIPLIER_MISMATCH",
        s"Expected ${expectedMultiplier}x for tier $senderTrustTier, got ${claim.appliedMultiplier}x"))
    }
    
    // Rule 3: No per-user daily cap enforcement (removed per PRD v2.5.1)
    // The auto-scaling rate naturally limits per-message rewards as network grows
    
    Right(ValidatedClaim(claim, expectedMultiplier))
  }
}
```

## Cardano Integration Sequence

```
Build Order:
1. Testnet DID registration (did:prism method)
2. Credential schema deployment (Plutus reference scripts)
3. Trust tier UTXO datum creation
4. Credential issuance flow (IDV → backend → Cardano)
5. Revocation bit vector management
6. Mainnet migration (same code, different network config)
```

---

# PART 4: iOS SPEC UPDATES

## Changes Required to Existing Frontend Blueprint

The iOS blueprint is already comprehensive. These updates align it with the build-ready spec:

### 1. Add ContactDiscoveryView to Presentation/Features

```swift
// Presentation/Features/Contacts/ContactDiscoveryView.swift
struct ContactDiscoveryView: View {
    @StateObject private var viewModel = ContactDiscoveryViewModel()
    
    var body: some View {
        NavigationStack {
            List {
                Section("Find Contacts") {
                    Button("Scan QR Code") { viewModel.showQRScanner = true }
                    Button("Search by Username") { viewModel.showSearch = true }
                    
                    if viewModel.isPhoneVerified {
                        Button("Match Phone Contacts") { viewModel.startDiscovery() }
                            .disabled(viewModel.isDiscovering)
                    } else {
                        // User hasn't verified phone yet — show upgrade prompt
                        Button(action: { viewModel.showPhoneVerification = true }) {
                            HStack {
                                Image(systemName: "phone.badge.checkmark")
                                VStack(alignment: .leading) {
                                    Text("Verify Phone to Find Contacts")
                                        .font(.body)
                                    Text("Upgrade to Tier 2 • Earn 1.2x rewards")
                                        .font(.caption).foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                }
                
                if !viewModel.matches.isEmpty {
                    Section("Contacts on ECHO") {
                        ForEach(viewModel.matches) { match in
                            ContactMatchRow(match: match)
                        }
                    }
                }
                
                Section("Invite Friends") {
                    ShareLink(item: viewModel.inviteLink) {
                        Label("Share Invite Link", systemImage: "link")
                    }
                    Text("Earn 50 ECHO when they verify")
                        .font(.caption).foregroundStyle(.secondary)
                }
            }
            .navigationTitle("Add Contacts")
            .sheet(isPresented: $viewModel.showPhoneVerification) {
                PhoneVerificationView()
            }
        }
    }
}
```

### 2. Add Zero-PII Onboarding Flow

```swift
// Presentation/Features/Onboarding/OnboardingView.swift

/// Zero-PII onboarding: username → passkey → DID → done
/// No phone number, no email, no real name collected at signup.
struct OnboardingView: View {
    @StateObject private var viewModel = OnboardingViewModel()
    
    var body: some View {
        NavigationStack {
            switch viewModel.step {
            case .username:
                UsernameEntryView(
                    username: $viewModel.username,
                    isAvailable: viewModel.isUsernameAvailable,
                    isChecking: viewModel.isCheckingUsername,
                    onSubmit: { viewModel.checkUsername() }
                )
                
            case .passkey:
                PasskeyCreationView(
                    username: viewModel.username,
                    onCreatePasskey: { await viewModel.createPasskey() }
                )
                // Triggers Face ID → Secure Enclave key generation
                // P-256 key pair created with .biometryCurrentSet access control
                
            case .creating:
                ProgressView("Creating your identity...")
                    .task { await viewModel.registerDID() }
                // Backend: POST /identity/register (username + publicKey)
                // Cardano: DID document registered (fee from platform treasury)
                
            case .complete:
                OnboardingCompleteView(
                    username: viewModel.username,
                    did: viewModel.did,
                    onContinue: { viewModel.finishOnboarding() }
                )
                // Shows: "Welcome @username — you're on ECHO"
                // Trust tier: 1 (Unverified) — can message immediately
            }
        }
    }
}

@MainActor
class OnboardingViewModel: ObservableObject {
    enum Step { case username, passkey, creating, complete }
    
    @Published var step: Step = .username
    @Published var username: String = ""
    @Published var isUsernameAvailable: Bool? = nil
    @Published var isCheckingUsername: Bool = false
    @Published var did: String = ""
    
    private let api: BackendAPIClient
    private let secureEnclave: SecureEnclaveManager
    
    func checkUsername() {
        isCheckingUsername = true
        Task {
            let available = try? await api.checkUsername(username)
            isUsernameAvailable = available
            isCheckingUsername = false
            if available == true { step = .passkey }
        }
    }
    
    func createPasskey() async {
        // Generate P-256 key pair in Secure Enclave
        // Requires Face ID / Touch ID authentication
        try? await secureEnclave.generateKeyPair(
            accessControl: .biometryCurrentSet
        )
        step = .creating
    }
    
    func registerDID() async {
        let publicKey = try? secureEnclave.getPublicKey()
        let result = try? await api.register(
            username: username,
            publicKey: publicKey!
        )
        did = result?.did ?? ""
        step = .complete
    }
}
```

### 3. Add Phone Verification View (Optional Tier 2 Upgrade)

```swift
// Presentation/Features/Identity/PhoneVerificationView.swift

/// Optional phone verification — NOT part of onboarding.
/// Triggered from trust-tier upgrade card on home screen or contact discovery.
/// Upgrades user from Tier 1 (1.0x rewards) to Tier 2 (1.2x rewards).
struct PhoneVerificationView: View {
    @StateObject private var viewModel = PhoneVerificationViewModel()
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                // Value proposition
                VStack(spacing: 8) {
                    Image(systemName: "phone.badge.checkmark")
                        .font(.system(size: 48))
                        .foregroundStyle(.blue)
                    Text("Verify Your Phone")
                        .font(.title2).bold()
                    Text("Find friends on ECHO and earn 1.2x rewards")
                        .font(.subheadline).foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }
                
                // Privacy assurance
                Label {
                    Text("Your number is hashed on-device. ECHO never stores raw phone numbers.")
                        .font(.caption)
                } icon: {
                    Image(systemName: "lock.shield")
                }
                .padding()
                .background(.regularMaterial)
                .clipShape(RoundedRectangle(cornerRadius: 8))
                
                switch viewModel.step {
                case .enterPhone:
                    TextField("Phone number", text: $viewModel.phoneNumber)
                        .keyboardType(.phonePad)
                        .textContentType(.telephoneNumber)
                    Button("Send Verification Code") {
                        await viewModel.sendOTP()
                    }
                    
                case .enterCode:
                    TextField("6-digit code", text: $viewModel.otpCode)
                        .keyboardType(.numberPad)
                    Button("Verify") {
                        await viewModel.confirmOTP()
                    }
                    
                case .verified:
                    VStack(spacing: 12) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 64)).foregroundStyle(.green)
                        Text("Phone Verified — Tier 2!")
                            .font(.title3).bold()
                        Text("Reward multiplier: 1.2x")
                        Button("Find Contacts Now") {
                            // Navigate to contact discovery
                            dismiss()
                        }
                    }
                }
            }
            .padding()
            .navigationTitle("Phone Verification")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}

@MainActor
class PhoneVerificationViewModel: ObservableObject {
    enum Step { case enterPhone, enterCode, verified }
    
    @Published var step: Step = .enterPhone
    @Published var phoneNumber: String = ""
    @Published var otpCode: String = ""
    
    private let api: BackendAPIClient
    private let contactsService: ContactDiscoveryUseCase
    
    func sendOTP() async {
        // POST /identity/verify-phone { phoneNumber, did }
        try? await api.initiatePhoneVerification(phoneNumber: phoneNumber)
        step = .enterCode
    }
    
    func confirmOTP() async {
        // POST /identity/verify-phone/confirm { did, code }
        let result = try? await api.confirmPhoneVerification(code: otpCode)
        
        if result != nil {
            // Hash phone number on-device with Argon2id + per-user salt
            // Then register hash with server for contact discovery
            let salt = try? SecureEnclaveManager.shared.getDiscoverySalt()
            let hash = Argon2id.hash(phoneNumber, salt: salt!)
            try? await contactsService.registerPhoneHash(hash)
            
            step = .verified
        }
    }
}
```

### 4. Add Trust Tier Upgrade Cards to Home Screen

```swift
// Presentation/Features/Conversations/TrustTierUpgradeCard.swift

/// Shows contextual trust-tier upgrade prompts on the conversation list.
/// These replace what would have been onboarding gates with voluntary upgrades.
struct TrustTierUpgradeCard: View {
    let currentTier: Int
    let onUpgrade: () -> Void
    
    var body: some View {
        if currentTier < 2 {
            UpgradeCard(
                icon: "phone.badge.checkmark",
                title: "Verify Phone",
                subtitle: "Find friends • Earn 1.2x rewards",
                action: onUpgrade
            )
        } else if currentTier < 3 {
            UpgradeCard(
                icon: "person.badge.shield.checkmark",
                title: "Verify Identity",
                subtitle: "Earn 100 ECHO • Unlock 1.5x rewards",
                action: onUpgrade
            )
        }
        // Tier 3+ users: no card shown (already verified)
    }
}
```

### 5. Add NetworkRewardStatus to WalletTab

```swift
// Replace DailyRewardsSection with NetworkRewardsSection
struct NetworkRewardsSection: View {
    let rewards: DailyRewards
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Today's Earnings")
                .font(.headline)
            
            HStack {
                VStack(alignment: .leading) {
                    Text("\(rewards.messaging, format: .number) ECHO")
                        .font(.title2).bold()
                    Text("from messaging")
                        .font(.caption).foregroundStyle(.secondary)
                }
                Spacer()
                VStack(alignment: .trailing) {
                    Text("Rate: \(rewards.currentAutoScaledRate, format: .number.precision(.fractionLength(4)))")
                        .font(.caption)
                    Text("× \(rewards.trustTierRewardMultiplier, format: .number.precision(.fractionLength(1)))x tier bonus")
                        .font(.caption).foregroundStyle(.secondary)
                }
            }
            
            // Network budget progress
            ProgressView(value: Double(rewards.networkDistributedToday),
                        total: Double(rewards.networkDailyBudget))
                .tint(.green)
            Text("Network: \(rewards.networkDistributedToday, format: .number) / \(rewards.networkDailyBudget, format: .number) ECHO today")
                .font(.caption2).foregroundStyle(.tertiary)
        }
        .padding()
        .background(.regularMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}
```

### 6. Key Hierarchy Constants

```swift
// Core/Security/SecureEnclaveManager.swift — Add these constants
extension SecureEnclaveManager {
    enum KeyContext: String {
        case didSigning = "echo-did-signing"
        case msgEncryption = "echo-msg-encryption"
        case storageEncryption = "echo-storage-encryption"
        case walletSigning = "echo-wallet-signing"
    }
    
    /// Derive a purpose-specific key using HKDF-SHA256
    func deriveKey(for context: KeyContext) throws -> SymmetricKey {
        let masterSignature = try sign(
            data: Data(context.rawValue.utf8),
            reason: "Key derivation"
        )
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: SymmetricKey(data: masterSignature),
            info: Data(context.rawValue.utf8),
            outputByteCount: 32
        )
    }
}
```

---

# PART 5: LAUNCH REQUIREMENT TRACKER

## Step-by-Step Checklist

### PRE-PHASE 1 (Now — April 11)

- [x] **DEV-001**: Set up GitHub monorepo with 3 directories: `echo-backend/`, `echo-ios/`, `echo-metagraph/`
- [x] **DEV-002**: Configure GitHub Actions CI: Go lint + test, Swift lint + test, Scala sbt test
- [x] **DEV-003**: Set up local Docker development environment (docker-compose with Redis, PostgreSQL, NATS)
- [x] **DEV-004**: Install Euclid SDK and verify local metagraph cluster starts
- [x] **DEV-005**: Create Cardano testnet wallet and fund with test ADA
- [x] **DEV-006**: Create Hetzner Cloud account + project; configure k3s cluster for dev/staging/prod
- [x] **DEV-007**: Provision 750K DAG (or confirm acquisition plan) for mainnet L0 staking
- [ ] **LEGAL-001**: Establish legal entity (LLC) for Apple Developer Program enrollment
- [ ] **LEGAL-002**: Enroll in Apple Developer Program ($99/year)
- [ ] **DESIGN-001**: Finalize app icon, splash screen, and core color palette

### PHASE 1: TESTNET + PROTOTYPE (April 12 — April 25)

**Metagraph (Scala):**
- [ ] **META-001**: Currency L1 genesis block logic (mint 1B ECHO, allocate 5 pools)
- [ ] **META-002**: Currency L1 reward claim validator (auto-scaling, tier multipliers)
- [ ] **META-003**: Currency L1 TokenLock validator (staking tiers)
- [ ] **META-004**: Data L1 Merkle root validator (structure, authorized sender)
- [ ] **META-005**: Data L1 trust commitment validator (H(score||nonce) format)
- [ ] **META-006**: All L1 validators passing unit tests on local Euclid cluster
- [ ] **META-007**: Deploy to Constellation testnet, verify snapshot finality < 30s

**Go Backend:**
- [ ] **GO-001**: Common package (config, logger, errors, crypto)
- [ ] **GO-002**: Auth middleware (passkey P-256 signature verification)
- [ ] **GO-003**: Metagraph client + circuit breaker
- [ ] **GO-004**: Cardano client + DID registry
- [ ] **GO-005**: WebSocket relay hub + handler (message routing)
- [ ] **GO-006**: Offline queue (Redis + PostgreSQL fallback)
- [ ] **GO-007**: Identity service (username + passkey → DID creation, zero PII)
- [ ] **GO-008**: Full message send/receive flow working: iOS → relay → iOS

**iOS:**
- [ ] **IOS-001**: Secure Enclave key generation (P-256, `.biometryCurrentSet`)
- [ ] **IOS-002**: HKDF key derivation (4 context keys)
- [ ] **IOS-003**: Kinnami encryption service (X25519 + ChaCha20-Poly1305)
- [ ] **IOS-004**: WebSocket relay client with reconnection
- [ ] **IOS-005**: MessageRelayManager (send + receive flow)
- [ ] **IOS-006**: Onboarding flow: username entry → passkey creation (Face ID) → DID registered → chat ready
- [ ] **IOS-007**: Basic chat UI (conversation list + chat view)
- [ ] **IOS-008**: End-to-end message flow: type → encrypt → relay → decrypt → display

**Phase 1 Go/No-Go Gate:**
- [ ] **GATE-P1-01**: iOS → Go backend → metagraph testnet full flow demonstrated
- [ ] **GATE-P1-02**: Metagraph testnet transaction finality < 30s
- [ ] **GATE-P1-03**: E2E encrypted message round-trip working
- [ ] **GATE-P1-04**: Security whitepaper draft complete

### PHASE 2: MAINNET CORE BUILD (April 26 — June 1 Alpha)

**Metagraph:**
- [ ] **META-101**: Deploy 3 L0 hybrid nodes on Hypergraph mainnet (750K DAG staked)
- [ ] **META-102**: Deploy Currency L1 + Data L1 validators (3 each)
- [ ] **META-103**: Execute token genesis (1B ECHO minted, 5 pools allocated)
- [ ] **META-104**: Create founder TokenLock positions (5 founders)
- [ ] **META-105**: ECHO token visible on Stargazer wallet + DAG Explorer
- [ ] **META-106**: Seed PacaSwap ECHO/DAG liquidity pool
- [ ] **META-107**: 2 weeks of mainnet health monitoring (no failed snapshots)

**Cardano:**
- [ ] **CARD-101**: Deploy DID registry schema to Cardano mainnet
- [ ] **CARD-102**: Fund platform treasury wallet with ADA (~15K ADA)
- [ ] **CARD-103**: Test DID registration end-to-end on mainnet
- [ ] **CARD-104**: Credential issuance + trust tier datum working

**Go Backend:**
- [ ] **GO-101**: Trust service (score computation, tier caching)
- [ ] **GO-102**: Rewards service (auto-scaling engine, claim validation)
- [ ] **GO-103**: Rewards batch processor (AtomicAction construction)
- [ ] **GO-104**: Merkle batcher (5-min/1000-msg batches)
- [ ] **GO-105**: Data L1 submission pipeline + snapshot listener
- [ ] **GO-106**: Contact discovery service (Argon2id matching — requires Tier 2+)
- [ ] **GO-107**: QR code + username search + invite links
- [ ] **GO-108**: Phone verification service (SMS OTP → Tier 2 upgrade, phone hash registration)
- [ ] **GO-109**: Push notification service (APNs)
- [ ] **GO-110**: Media service (encrypted upload/download to Storj)
- [ ] **GO-111**: Log publisher (encrypted IPFS batches)
- [ ] **GO-112**: Gateway (load balancer, TLS termination, rate limiting)
- [ ] **GO-113**: Deploy all 10 services to Hetzner Cloud k3s cluster (Argo CD GitOps)
- [ ] **GO-114**: Load test: 1,000 concurrent WebSocket connections
- [ ] **GO-115**: Integration tests: all API routes passing

**iOS:**
- [ ] **IOS-101**: DID creation on Cardano mainnet
- [ ] **IOS-102**: Trust scoring display (tier badge, multiplier)
- [ ] **IOS-103**: Group messaging (GroupKeyManager)
- [ ] **IOS-104**: AnchoringTracker + chain-link icon on messages
- [ ] **IOS-105**: ECHO Wallet tab (balance, staking, delegation)
- [ ] **IOS-106**: Reward claiming (AtomicAction)
- [ ] **IOS-107**: Contact discovery (QR + username search + invite links)
- [ ] **IOS-108**: Phone verification flow (optional Tier 2 upgrade → unlocks address book matching)
- [ ] **IOS-109**: Hidden folders (biometric-gated Secure Enclave)
- [ ] **IOS-110**: Disappearing messages (client-side timers)
- [ ] **IOS-111**: Profile + rewards tracker
- [ ] **IOS-112**: Push notification handling
- [ ] **IOS-113**: Offline queue management (outbound + inbound drain)
- [ ] **IOS-114**: Settings (privacy, notifications, security)
- [ ] **IOS-115**: Background key purging on app background
- [ ] **IOS-116**: TestFlight internal build (100 Apple IDs)

**Security Gates:**
- [ ] **SEC-201**: E2E encryption audit (external security firm)
- [ ] **SEC-202**: Secure Enclave audit (Apple platform review)
- [ ] **SEC-203**: Scala L1 code review (blockchain security firm)
- [ ] **SEC-204**: Go backend penetration test (OWASP scope)

**Phase 2 Go/No-Go → TestFlight Alpha (June 1):**
- [ ] **GATE-P2-01**: Metagraph mainnet finality < 10s (95th percentile)
- [ ] **GATE-P2-02**: Message delivery rate > 99.9% in load test
- [ ] **GATE-P2-03**: 100+ alpha users active for 2+ weeks, no security incidents
- [ ] **GATE-P2-04**: ECHO token visible on Stargazer + DAG Explorer
- [ ] **GATE-P2-05**: PacaSwap ECHO/DAG pool seeded
- [ ] **GATE-P2-06**: Crash rate < 1% of sessions
- [ ] **GATE-P2-07**: All 4 security gates passed or scheduled

### PHASE 3: APP STORE LAUNCH (June — August 2026)

- [ ] **P3-001**: TestFlight public beta (1K–10K users)
- [ ] **P3-002**: Sealed sender implementation
- [ ] **P3-003**: Client-side Merkle proof verification
- [ ] **P3-004**: Digital Evidence integration (Smart Checkmark)
- [ ] **P3-005**: Midnight ZK proof-of-concept (trust tier on testnet)
- [ ] **P3-006**: Open source entire codebase (Apache 2.0)
- [ ] **P3-007**: App Store submission (screenshots, privacy labels, review notes)
- [ ] **P3-008**: App Store approval + phased release (1% → 100% over 7 days)
- [ ] **P3-009**: 30-day retention > 60% in soft launch
- [ ] **P3-010**: 99.9% uptime for 3 consecutive months

### PHASE 4: SCALE + ENTERPRISE (Q4 2026+)

- [ ] **P4-001**: Multi-cloud relay deployment (Hetzner primary + OVHcloud secondary + community nodes on Akash/Flux/bare metal)
- [ ] **P4-002**: Community relay operator registration on Data L1
- [ ] **P4-003**: Community L1 validator onboarding + slashing activation
- [ ] **P4-004**: Midnight mainnet integration (trust tier + age proofs)
- [ ] **P4-005**: Governance DAO operational (first on-chain proposal)
- [ ] **P4-006**: Enterprise pilot program (5 organizations)
- [ ] **P4-007**: 100K+ MAU milestone

---

## Critical Path Summary

```
                            CRITICAL PATH TO JUNE 1
                            ========================

April 4 ──→ April 11: Environment Setup (DEV-001 through DESIGN-001)
   │
April 12 ──→ April 25: PHASE 1 — Testnet PoC
   │    ├── Scala metagraph L1 validators (META-001 → META-007)
   │    ├── Go backend core services (GO-001 → GO-008)
   │    └── iOS encryption + relay + basic chat (IOS-001 → IOS-008)
   │
April 25: ★ PHASE 1 GATE — Full flow demonstrated on testnet
   │
April 26 ──→ May 16: PHASE 2 — Mainnet Build
   │    ├── Metagraph mainnet deployment (META-101 → META-107)
   │    ├── Cardano mainnet deployment (CARD-101 → CARD-104)
   │    ├── Go backend all 10 services (GO-101 → GO-115)
   │    └── iOS full feature set (IOS-101 → IOS-116)
   │
May 16 ──→ May 31: Integration + Security + Load Testing
   │    └── Security gates (SEC-201 → SEC-204)
   │
June 1: ★ TESTFLIGHT ALPHA — 100-500 users
```

This spec is ready for VS Code ingestion. Each checklist item maps to a specific file or function in the directory structures above.
