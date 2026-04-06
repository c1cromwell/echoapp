# Phase 1–2 Engineering Kickoff Checklist

## 1. Pre-Implementation Planning

### 1.1 Architecture Review ✅
- [x] DATA_LAYER_ARCHITECTURE_v3.1 reviewed and enhanced
- [x] All 54 OpenAPI endpoints mapped to services
- [x] Service integration dependencies documented
- [x] Cross-chain consistency strategy defined
- [x] Failure scenarios and recovery targets established

**Action:** Confirm architecture with team before starting

### 1.2 Feature Completeness ✅
- [x] Missing features identified: 6 (achievements, disappearing messages, sealed sender, WebSocket, referral, reactions)
- [x] Full implementation specs provided for all features
- [x] Feature phasing: Phase 1 (core), Phase 2 (full launch), Phase 3+ (enhancements)
- [x] OpenAPI spec alignment verified (all endpoints covered)

**Action:** Prioritize features for each sprint

### 1.3 Test Strategy ✅
- [x] 37 unit test cases specified with assertions
- [x] 7 integration test cases specified
- [x] 6 E2E test cases specified
- [x] Coverage goals defined: 90%/80%/70% (service/integration/E2E)
- [x] Test environment setup documented (Docker Compose)
- [x] CI/CD pipeline defined (GitHub Actions)

**Action:** Set up test infrastructure before Phase 1 code

---

## 2. Phase 1 Implementation (Weeks 1–6)

### 2.1 Backend Go Services

#### AuthService ⏳
- [ ] User registration (DID creation on Cardano)
- [ ] Challenge generation + verification
- [ ] JWT token issuance + refresh
- [ ] Logout + token revocation
- [ ] Tests: 9 unit tests (register, challenge, verify, refresh, logout, edge cases)
- **Timeline:** Week 1–2
- **Dependencies:** Cardano integration, Redis cache

#### MessageService ⏳
- [ ] Message send + encryption validation
- [ ] Offline queue (Redis + PostgreSQL fallback)
- [ ] Message retrieval with cursor pagination
- [ ] Read receipts + status tracking
- [ ] Reaction management
- [ ] Tests: 8 unit tests + 1 E2E test (send/receive/offline/reconnect)
- **Timeline:** Week 2–3
- **Dependencies:** APNs integration, offline queue infrastructure

#### RewardService ⏳
- [ ] Reward claim submission + validation
- [ ] Trust tier multiplier application
- [ ] Daily cap enforcement
- [ ] Batching to Currency L1
- [ ] Balance cache management
- [ ] Tests: 5 unit tests + 1 E2E test (claim/staking/confirmation)
- **Timeline:** Week 3–4
- **Dependencies:** Metagraph Currency L1 integration

#### IdentityService ⏳
- [ ] DID creation on Cardano
- [ ] Credential issuance + revocation
- [ ] IDV provider callback handler
- [ ] Trust tier updates
- [ ] Verification status tracking
- [ ] Tests: 4 unit tests + 1 E2E test (verification flow)
- **Timeline:** Week 4–5
- **Dependencies:** Cardano integration, IDV provider (TBD)

#### ContactService ⏳
- [ ] Add/remove/block/unblock contacts
- [ ] Contact list with cursor pagination
- [ ] Contact filtering (blocked excluded)
- [ ] Tests: 4 unit tests
- **Timeline:** Week 5
- **Dependencies:** User service, database schema

#### WebSocket Relay Service ⏳
- [ ] WebSocket connection management
- [ ] Message forwarding (online recipients)
- [ ] Offline queue flushing
- [ ] ACK + confirmation stream
- [ ] Rate limiting per connection
- [ ] Tests: 4 unit tests + 1 E2E test (real-time delivery)
- **Timeline:** Week 5–6
- **Dependencies:** Redis, NATS pub/sub

#### Supporting Infrastructure ⏳
- [ ] Circuit breakers (Metagraph, Cardano, IPFS)
- [ ] Cache service (Redis TTL management)
- [ ] Health check endpoints
- [ ] Prometheus metrics + logging
- [ ] API rate limiting (per-device token bucket)
- [ ] Tests: 7 integration tests (circuit breaker, cache, health)
- **Timeline:** Week 1 (parallel) + Week 6 (polish)
- **Dependencies:** Redis, monitoring stack

### 2.2 Database + Cache Setup

#### PostgreSQL Schema ⏳
- [ ] Users table (DID, public key, trust tier)
- [ ] Message queue table (offline messages)
- [ ] Reward claims table (with status tracking)
- [ ] Contact relationships table
- [ ] Relay metrics table (for monitoring)
- [ ] Proper indexes (DID, timestamp, status)
- [ ] Migrations with versioning
- **Timeline:** Week 1
- **Reference:** TESTING_STRATEGY.md Section 1.2

#### Redis Schemas ⏳
- [ ] Offline message queue (sorted set by timestamp)
- [ ] Rate limit buckets (token bucket per DID)
- [ ] Cache keys (balances, trust scores, credentials)
- [ ] Session storage (auth tokens)
- [ ] Circuit breaker state
- **Timeline:** Week 1
- **Reference:** DATA_LAYER_ARCHITECTURE_v3.1 Section 2.4

#### Test Fixtures ⏳
- [ ] User fixtures with DID generation
- [ ] Message fixtures with encryption
- [ ] Reward claim fixtures
- [ ] Randomization utilities
- **Timeline:** Week 1
- **Reference:** TESTING_STRATEGY.md Section 1.3

### 2.3 Blockchain Integration

#### Cardano Integration ⏳
- [ ] DID creation (did:prism via Atala PRISM)
- [ ] Credential schema publication
- [ ] Credential issuance on verification
- [ ] Credential status bit vector
- [ ] Trust tier datum management
- [ ] Transaction fee delegation (platform pays)
- [ ] Kupo indexer queries for validation
- [ ] Tests: 2 integration tests (DID creation, credential issuance)
- **Timeline:** Week 2–4
- **Dependencies:** Cardano testnet setup, Kupo indexer
- **Decisions:** Which DID method? (Q: already answered: did:prism)

#### Metagraph Integration ⏳
- [ ] Data L1 node setup (custom validator)
- [ ] Currency L1 validator (balance + staking)
- [ ] Message batch submission (Merkle root)
- [ ] Snapshot event subscription
- [ ] Finality callback handling
- [ ] Offline queue failover
- [ ] Tests: 2 integration tests (message anchoring, reward submission)
- **Timeline:** Week 2–5
- **Dependencies:** Metagraph simulator (local) + testnet (later)
- **Decisions:** Mainnet vs. dedicated metagraph? (Q2: decide before Phase 2)

#### IPFS Integration ⏳
- [ ] Batch encryption (AES-256-GCM)
- [ ] Upload to Pinata + self-hosted node
- [ ] CID recording + indexing
- [ ] Storj fallback configuration
- [ ] Log retrieval with decryption
- [ ] Tests: 1 integration test (IPFS upload + fallback)
- **Timeline:** Week 4
- **Dependencies:** Pinata API key, Storj account

### 2.4 External Service Integration

#### APNs (Apple Push Notifications) ⏳
- [ ] APNs certificate setup
- [ ] Push notification sending (offline queue)
- [ ] Error handling + feedback loop
- [ ] Silent notifications (delivery + confirmation)
- [ ] VoIP push for message wake-up (Phase 2)
- [ ] Tests: Mocked in unit tests; real APNs in staging
- **Timeline:** Week 2–3
- **Dependencies:** Apple Developer account, certificates

#### IDV Provider (TBD - Stripe/Sumsub/Jumio) ⏳
- [ ] Provider SDK integration (client)
- [ ] Webhook receiver (server)
- [ ] Idempotency + deduplication
- [ ] Credential issuance trigger
- [ ] Trust tier update flow
- [ ] Tests: Mocked IDV in unit tests; real provider in staging
- **Timeline:** Week 4–5
- **Dependencies:** IDV provider decision (Q1: needed before Phase 2)
- **Status:** BLOCKED until Q1 answered

### 2.5 Deployment Infrastructure ⏳

#### Docker + Kubernetes ⏳
- [ ] Dockerfile for Go backend (multi-stage build)
- [ ] Docker Compose for local development
- [ ] Kubernetes manifests for staging/production
- [ ] Health checks + liveness probes
- [ ] Resource limits (CPU, memory)
- [ ] Horizontal pod autoscaling configuration
- **Timeline:** Week 1–2 (docker), Week 5–6 (k8s)
- **Reference:** DATA_LAYER_ARCHITECTURE_v3.1 Section 2.4 (Scaling Architecture)

#### Monitoring + Observability ⏳
- [ ] Prometheus metrics export
- [ ] Logging infrastructure (structured logs)
- [ ] OpenTelemetry tracing (optional Phase 2)
- [ ] Grafana dashboards for key metrics
- [ ] Alert rules (message latency, queue depth, error rate)
- **Timeline:** Week 4–5 (parallel)
- **Metrics reference:** TESTING_STRATEGY.md Section 1.2

---

## 3. Phase 2 Planning (Weeks 7–12)

### 3.1 Feature Completions

- [ ] Message reactions (emoji support)
- [ ] Disappearing messages (TTL + auto-delete)
- [ ] Achievement system (definition + tracking)
- [ ] Sealed sender (Phase 3 prep)
- [ ] Referral system (Phase 2 launch)
- [ ] Enhanced WebSocket (presence, typing indicators)

### 3.2 Security Hardening

- [ ] Token expiry enforcement
- [ ] Rate limit bypass prevention
- [ ] Data classification validation (T0–T7)
- [ ] Encryption validation (plaintext never logged)
- [ ] Security audit (internal + external)

### 3.3 Scale Testing

- [ ] Stress test: 1M users, 100K online, 50 msg/sec
- [ ] Reward batching: 10K claims/sec
- [ ] Circuit breaker recovery scenarios
- [ ] Offline queue limits + overflow

---

## 4. Decision Gate: Open Questions

| Q# | Question | Impact | Answer By | Status |
|----|----------|--------|-----------|--------|
| 1 | IDV Provider? (Stripe/Sumsub/Jumio) | Cardano credential flow | End of Week 1 | 🟡 NEEDED |
| 2 | Metagraph deployment? (mainnet vs. dedicated) | Node infrastructure | End of Week 2 | 🟡 NEEDED |
| 4 | Group key strategy? (<100: per-recipient, 100+: symmetric) | Performance tradeoff | End of Week 2 | 🟡 NEEDED |
| 10 | Referral cap? (unlimited vs. N) | Fraud prevention | End of Week 3 | 🟡 NEEDED |
| 3 | Staking on Currency L1 or separate? | Tokenomics | End of Week 4 | 🟢 LOWER PRIORITY |

**Recommendation:** Answer Q1, Q2, Q4, Q10 before starting Phase 1 code to avoid rework.

---

## 5. Resource Allocation

### 5.1 Team Structure (Recommended)

**Backend (Go):** 2–3 engineers
- 1 senior architect (services + blockchain integration)
- 1–2 full-stack (services + tests)

**Blockchain (Cardano + Metagraph):** 1–2 engineers
- 1 Cardano specialist (DID + credentials)
- 1 Metagraph specialist (L1 validator + snapshot integration)

**DevOps/Infrastructure:** 1 engineer
- Docker + Kubernetes setup
- Monitoring + alerting
- Database administration

**QA/Testing:** 1 engineer
- Test infrastructure setup
- Automated testing
- Stress testing + load profiles

**Total:** 5–6 full-time engineers

### 5.2 External Services

**Cardano Testnet:**
- Pre-provisioned, costs: free
- Kupo indexer: self-hosted or managed service

**Metagraph:**
- Local simulator: provided
- Testnet access: pending Q2 decision

**APNs:**
- Apple Developer account: required (team already has)
- Sandbox + production: free

**IDV Provider:**
- Stripe Identity: ≤$0.25/verification
- Sumsub/Jumio: ≤$1.00/verification
- Budget: ~$3K/month for Phase 2 (10K verifications)

**IPFS/Pinata:**
- Free tier: 100 GB storage
- Paid: $19/month (1 TB) or usage-based

---

## 6. Success Criteria (Phase 1 Completion)

### 6.1 Functional Criteria ✅
- [ ] All 54 API endpoints working (54/54)
- [ ] Message send/receive P2P + offline (test: WORKING)
- [ ] Reward claims batched to Metagraph (test: WORKING)
- [ ] Identity verification → Cardano credential (test: WORKING)
- [ ] WebSocket real-time delivery (<100ms P50)
- [ ] Rate limiting enforced (60 msg/min Tier 1)
- [ ] Offline queue working (1000 msg limit)
- [ ] Circuit breakers preventing cascade failures

### 6.2 Quality Criteria ✅
- [ ] Unit test coverage: ≥90% (service layer)
- [ ] Integration test coverage: ≥80% (chain interactions)
- [ ] E2E test coverage: ≥70% (critical flows)
- [ ] All tests automated in CI/CD
- [ ] Zero data loss scenarios
- [ ] No plaintext in logs/storage

### 6.3 Performance Criteria ✅
- [ ] Message relay P50 latency: <100ms
- [ ] Message relay P99 latency: <500ms
- [ ] Offline queue drain: full queue in <30 seconds
- [ ] Metagraph finality: <10 seconds
- [ ] Cache hit rate: ≥95% (balances, trust scores)

### 6.4 Security Criteria ✅
- [ ] End-to-end encryption enforced (Kinnami X25519 + ChaCha20)
- [ ] Sender signatures verified by recipient
- [ ] No PII on-chain (enforced by Data L1 validators)
- [ ] Token expiry enforced (access: 1hr, refresh: 30 days)
- [ ] Rate limits prevent abuse
- [ ] Audit trail immutable (IPFS + Data L1)

---

## 7. Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cardano testnet congestion | Medium | Message anchoring delayed | Use Kupo indexer; cache credentials |
| Metagraph validator crash | Low | Message batching halts (messages still relay) | Multi-validator setup (Phase 1: 3, Phase 2: 5) |
| APNs certificate expiry | Low | Notifications stop | Auto-renewal + monitoring alerts |
| IPFS pinning failure | Medium | Audit trail backup delayed | Storj fallback; local buffering (5-min retry) |
| Database corruption | Low | Operational data lost (no messages lost) | Backup + replication (primary + 2 replicas) |
| Rate limit bypass | Low | Spam/DoS possible | Distributed rate limiter (Redis); per-IP + per-DID |

---

## 8. Artifact Checklist

### 8.1 Documentation ✅
- [x] Architecture: DATA_LAYER_ARCHITECTURE_v3.1 (11 sections)
- [x] Testing: TESTING_STRATEGY.md (6 sections + code examples)
- [x] iOS Alignment: iOS_OPENAPI_IMPLEMENTATION_GUIDE.md
- [x] Review Summary: REVIEW_ENHANCEMENT_SUMMARY.md
- [x] OpenAPI Spec: openapi.yaml (54 endpoints)

### 8.2 Code Templates (Ready) ✅
- [x] Unit test examples (AuthService, MessageService, etc.)
- [x] Integration test examples (Cardano, Metagraph, IPFS)
- [x] E2E test examples (auth, messaging, rewards, IDV)
- [x] Docker Compose stack
- [x] Database migrations
- [x] Test fixtures

### 8.3 Infrastructure Code (Ready) ✅
- [x] GitHub Actions CI/CD pipeline template
- [x] Kubernetes manifests template
- [x] Prometheus alerts template
- [x] Grafana dashboard template (TBD: Phase 2)

---

## 9. Timeline Overview

```
Week 1–2: Foundation (Auth, Database, Docker)
  ├─ AuthService (register, challenge, verify)
  ├─ PostgreSQL + Redis setup
  ├─ Docker + Kubernetes templates
  └─ Unit tests for AuthService

Week 2–3: Messaging (Core Product)
  ├─ MessageService (send, queue, forward)
  ├─ WebSocket relay
  ├─ APNs integration
  └─ Unit + E2E tests for messaging

Week 3–4: Rewards (Blockchain Integration)
  ├─ RewardService (claim, batch, confirm)
  ├─ Metagraph integration (Currency L1)
  └─ Tests for reward flow

Week 4–5: Identity + Contacts
  ├─ IdentityService (DID, credentials)
  ├─ Cardano integration (credential issuance)
  ├─ ContactService
  └─ IDV provider callback handler

Week 5–6: Polish + Hardening
  ├─ Circuit breakers + health checks
  ├─ Monitoring + metrics
  ├─ Security audit (internal)
  └─ Stress testing + load profiles

Week 7–12: Phase 2 (Full Launch)
  ├─ Remaining features (achievements, disappearing, sealed sender)
  ├─ Enhanced testing (external security audit)
  └─ Staging deployment + QA
```

---

## 10. Next Steps

1. **Answer 4 critical decisions (by EOW1):**
   - [ ] IDV provider: Stripe Identity ✓ / Sumsub / Jumio
   - [ ] Metagraph: Constellation mainnet / dedicated deployment
   - [ ] Group key strategy: per-recipient cutoff point
   - [ ] Referral cap: unlimited / 50 / 100

2. **Set up development environment (Week 1):**
   - [ ] Clone monorepo + install Go deps
   - [ ] Start Docker Compose stack
   - [ ] Create PostgreSQL + Redis schemas (migrations)
   - [ ] Test fixtures + test data

3. **Create GitHub project + sprints:**
   - [ ] Phase 1 epic (6 weeks)
   - [ ] Sprint 1–2: Foundation (Weeks 1–2)
   - [ ] Sprint 3–4: Messaging (Weeks 3–4)
   - [ ] Sprint 5–6: Rewards + Identity (Weeks 5–6)

4. **Coordinate with iOS team:**
   - [ ] Share iOS_OPENAPI_IMPLEMENTATION_GUIDE.md
   - [ ] Align on API contract (POST/PUT/DELETE methods, paths, status codes)
   - [ ] Schedule bi-weekly sync on integration progress

5. **Begin Phase 1 implementation:**
   - [ ] Start Week 1: AuthService + database setup
   - [ ] Follow test-first approach (tests before code)
   - [ ] Daily standup + weekly progress review

---

*Phase 1–2 Engineering Kickoff Checklist*
*Created: February 23, 2026*
*Status: Ready for team onboarding*
