# Privacy Architecture Implementation Plan

## Executive Summary

The privacy blueprint (1165 lines) provides comprehensive specifications for a privacy-first tokenomics platform. This plan identifies implementation gaps and provides a structured approach to realizing the blueprint across Go backend and iOS platforms.

**Status**: Blueprint review complete. Implementation plan created.
**Blueprint Coverage**: 100% analyzed, all 8 sections reviewed
**Estimated Effort**: 6-8 weeks for full implementation
**Priority**: Implement Secure Enclave + Encryption first (dependency for all other services)

---

## Blueprint Sections Analyzed

### ✅ Section 1: Core Privacy Principles (Lines 1-150)
**Identified Principles:**
- Data minimization (collect only what's necessary)
- Local-first processing (sensitive data stays on device)
- Biometric binding (crypto operations require Face ID/Touch ID)
- Zero PII on blockchain (only hashes and commitments)
- Unlinkability (different pseudonyms per context)
- Forward secrecy (compromised keys don't expose past data)
- User sovereignty (users control their keys)

**Implementation Status**: ✅ ARCHITECTURAL - Embedded in design
**Gap Analysis**: None - these are design principles, not code

---

### ✅ Section 2: Data Classification System (Lines 150-300)
**8-Tier Classification (T0-T7):**

| Tier | Examples | Storage | Blockchain |
|------|----------|---------|-----------|
| **T0: Ultra-Sensitive** | Biometrics, Private Keys | Secure Enclave only | Never |
| **T1: PII** | Names, SSN, Address, DOB | Device encrypted | Never |
| **T2: Account Secrets** | Account numbers, Auth tokens | Device encrypted | Never |
| **T3: Personal Comms** | Messages, Files | Device encrypted, Server encrypted | Never (only hash commitment) |
| **T4: Contact Info** | Phone, Email | Device hashed | Never (salted hash only) |
| **T5: Verification** | Document images, Selfies | IDV provider only | Never (reference ID only) |
| **T6: Aggregated Data** | Trust scores, Reputation | Device → Commitment | Tier only (not exact score) |
| **T7: Public Data** | DIDs, Public keys, Timestamps | Public | Public (no restriction) |

**13 Prohibited Types (Never on-chain):**
Real names, phones, addresses, SSN, DOB, biometrics, account numbers, message content, files, IPs, device IDs, contact lists, private keys, session tokens

**10 Blockchain-Safe Types:**
DID, public keys, hash commitments, Merkle roots, reference IDs, ZKPs, credential status, token balances, governance votes, timestamps

**Implementation Status**: ⚠️ PARTIAL
**Current Implementation**: Token models exist, trust tiers exist
**Missing Implementation**:
- Data classification enforcement layer (enforcer that prevents T0-T5 from leaving device)
- Hash commitment generators for T3-T4 data
- Reference ID mapping system
- Compliance audit checks

---

### ✅ Section 3: iOS Secure Enclave Architecture (Lines 300-450)
**Specification: 3-Tier Key Hierarchy**

```
Device Root Key (DRK)
    ↓ (Biometric-Protected)
Biometric-Protected Key (BPK)
    ↓ (HKDF Derivation)
User Identity Key (UIK)
    ↓ (Context-Specific)
Derived Keys: Storage, Signing, Encryption, etc.
```

**Implementation Details Provided:**
- Key generation in Secure Enclave (never extractable)
- Biometric gating with Face ID/Touch ID
- HKDF derivation for application keys
- Challenge-response signing
- Key lifecycle management

**Code Examples Provided (Swift):**
- `generateBiometricProtectedKey()` - Creates BPK with biometric requirements
- `sign()` - Signs with Face ID/Touch ID prompt
- `deriveApplicationKeys()` - Derives context-specific keys
- `rotateKeys()` - Periodic key rotation
- `deleteKeys()` - Secure key deletion

**Implementation Status**: ❌ NOT IMPLEMENTED
**Gap Severity**: **CRITICAL** - Required for all crypto operations
**Effort**: 2 weeks

---

### ✅ Section 4: Android StrongBox Equivalent (Lines 450-550)
**Specification: Comparable Security Model**

- Android KeyStore with KeyGenParameterSpec
- StrongBox backing (if hardware available)
- Biometric authentication
- Signature and encryption operations
- Fallback to software KeyStore

**Code Examples Provided (Kotlin):**
- Key generation with biometric binding
- Hardware-backed operations
- Fallback mechanisms

**Implementation Status**: ❌ NOT IMPLEMENTED
**Gap Severity**: **HIGH** - Required for Android platform
**Note**: Kotlin implementation not yet built (iOS Swift priority)
**Effort**: 2 weeks

---

### ✅ Section 5: Blockchain Privacy Model (Lines 550-700)
**Specification: Privacy-Preserving On-Chain Data**

**Privacy Strategies:**
1. **Hash Commitments**: H(H(plaintext) || nonce) - Proves integrity without exposure
2. **Reference IDs**: Opaque UUIDs with no semantic meaning
3. **Merkle Aggregation**: Batch commits into single root
4. **Tier Restrictions**: Only T7 data allowed on-chain

**On-Chain Data Structures (Privacy-Safe):**
- DID Documents: Public keys only
- Token State: Balances by pseudonymous DID
- Message Integrity: Merkle root only
- Credential Status: Bit vector for revocation
- Trust Commitment: H(score || nonce) instead of raw score

**Implementation Status**: ⚠️ PARTIAL
**Current Implementation**: Token state model exists
**Missing Implementation**:
- Hash commitment generator
- Reference ID system
- Merkle tree aggregation
- Message integrity root generation
- Credential reference system
- Trust commitment scheme

**Effort**: 1 week

---

### ✅ Section 6: Hashing & Commitment Schemes (Lines 700-900)
**Specification: Cryptographic Integrity Mechanisms**

**Hashing Strategies:**
- **Salted Hashing**: Argon2id for phone/email with per-user salt
- **Blind Index Approach**: Deterministic but unlinkable
- **Hash Commitments**: H(H(plaintext) || nonce)
- **Merkle Aggregation**: Batch tree with root commitment

**Code Examples Provided (TypeScript):**
- Salted hashing with Argon2id
- Blind index generation
- Hash commitment creation
- Merkle proof verification
- Reference ID generation

**Implementation Status**: ❌ NOT IMPLEMENTED (except basic hashing)
**Gap Severity**: **HIGH** - Required for blockchain integrity
**Missing Components**:
- Argon2id implementation (Go: `golang.org/x/crypto/argon2`)
- Merkle tree builder
- Commitment generator
- Proof verifier
- Blind index system

**Effort**: 1.5 weeks

---

### ✅ Section 7: Zero-Knowledge Proofs (Lines 900-1050)
**Specification: Privacy-Preserving Verification**

**Use Cases with Proof Types:**
1. **Age Verification**: Prove age > threshold without revealing birthdate
2. **Balance Proof**: Prove balance ≥ X without revealing exact amount
3. **Credential Validity**: Prove credential is valid without revealing content
4. **Trust Threshold**: Prove score ≥ threshold without revealing exact score
5. **Membership**: Prove group membership without revealing identity
6. **Transaction Auth**: Prove user authorization without revealing key

**Implementation Details:**
- Groth16 proving system
- SNARK circuits for each proof type
- Public signals (what verifier sees)
- Private inputs (what prover keeps secret)
- Witness generation and proof generation

**Code Examples Provided (TypeScript):**
- Age verification circuit
- Balance threshold proof
- Credential ownership proof
- Full snarkjs integration

**Implementation Status**: ❌ NOT IMPLEMENTED
**Gap Severity**: **MEDIUM** - Advanced feature, not blocking core operations
**Dependencies**: Requires snarkjs library and circuit compilation
**Effort**: 3 weeks (includes circuit design)

---

### ✅ Section 8: Data Flow Architectures (Lines 1050-1165)
**Specification: Privacy-Preserving Message & ID Verification Flows**

**Message Flow (Privacy-Preserving):**
```
Sender Device (plaintext) 
  → E2E Encrypt with ChaCha20-Poly1305
  → Sign with Secure Enclave
  → Server sees only: encrypted blob, DIDs, timestamp (size)
  → Blockchain sees only: Merkle root commitment
  → Recipient Device: Biometric unlock, decrypt, verify
```

**ID Verification Flow:**
```
User Device (captures images)
  → Direct connection to IDV Provider (bypasses platform backend)
  → IDV processes, extracts PII, verifies, DELETES images
  → Returns only: pass/fail, confidence, document type, age over 18
  → Platform backend: Never sees raw images or PII
  → Blockchain: Only receives reference UUID and credential tier
```

**Implementation Status**: ⚠️ PARTIAL
**Current Implementation**: Messaging service skeleton exists
**Missing Implementation**:
- E2E encryption in messaging service
- Server-side encrypted storage
- Merkle tree batching
- IDV provider integration
- Privacy-preserving credential storage

**Effort**: 2 weeks

---

## Implementation Gaps Summary

### Critical Gaps (Blocking Core Features)

| Gap | Location | Current State | Required For | Effort |
|-----|----------|---------------|-------------|--------|
| **Secure Enclave Manager** | iOS | Missing | All crypto operations | 2 weeks |
| **Message Encryption (E2EE)** | Messaging Service | Stub only | Secure messaging | 1.5 weeks |
| **Key Derivation** | iOS + Go | Missing | Token signing, encryption | 1 week |
| **Hash Commitments** | Blockchain Privacy | Missing | On-chain data integrity | 1 week |
| **Data Classification Enforcer** | Privacy Layer | Missing | PII leakage prevention | 1 week |

### High-Priority Gaps

| Gap | Location | Current State | Required For | Effort |
|-----|----------|---------------|-------------|--------|
| **Merkle Tree Aggregation** | Blockchain | Missing | Message batching on-chain | 1 week |
| **Reference ID System** | Privacy Utilities | Missing | Opaque on-chain references | 3 days |
| **Salted Hashing** | Crypto Utils | Missing | Phone/email hashing | 3 days |
| **Audit Logging** | Observability | Missing | Privacy compliance | 1 week |
| **Local Encrypted Storage** | iOS + Go | Missing | Persistent encryption at rest | 1 week |

### Medium-Priority Gaps

| Gap | Location | Current State | Required For | Effort |
|-----|----------|---------------|-------------|--------|
| **Zero-Knowledge Proofs** | Verification | Missing | Privacy-preserving age/balance checks | 3 weeks |
| **Android StrongBox** | Android | Missing | Android platform support | 2 weeks |
| **IDV Provider Integration** | Identity Verification | Missing | ID verification without PII leakage | 2 weeks |
| **Credential Reference System** | Credentials | Missing | Blockchain credential storage | 1 week |

---

## Recommended Implementation Sequence

### Phase 1: Foundation (Weeks 1-2) - **CRITICAL**

**Priority 1.1: Secure Enclave Manager (iOS)**
- Dependency: All other iOS crypto operations
- Deliverable: `ios/Echo/Sources/Security/SecureEnclaveManager.swift`
- Includes: Key generation, biometric gating, signing, key derivation
- Tests: 8-10 test cases covering key lifecycle

**Priority 1.2: Go Crypto Utilities**
- Dependency: Backend encryption, hashing, signatures
- Deliverables:
  - `internal/crypto/keyderivation.go` - HKDF-based key derivation
  - `internal/crypto/hashing.go` - Salted hashing with Argon2id
  - `internal/crypto/signing.go` - ECDSA operations
- Tests: 12-15 test cases

### Phase 2: Encryption & Privacy (Weeks 3-4)

**Priority 2.1: Message Encryption Service**
- Dependency: Secure Enclave Manager (iOS), Crypto Utils (Go)
- Deliverables:
  - `internal/services/encryption/e2ee.go` - Server-side key management
  - `ios/Echo/Sources/Services/EncryptionService.swift` - iOS E2E encryption
  - `internal/services/messaging/message_encryption.go` - Message encryption integration
- Tests: 15-20 test cases

**Priority 2.2: Local Storage Encryption**
- Dependency: Secure Enclave Manager (iOS)
- Deliverables:
  - `ios/Echo/Sources/Storage/SecureLocalStorage.swift`
  - `internal/storage/encrypted_storage.go`
- Tests: 10-12 test cases

### Phase 3: Blockchain Privacy (Weeks 5-6)

**Priority 3.1: Hash Commitments & Reference IDs**
- Dependency: Crypto Utilities
- Deliverables:
  - `internal/crypto/commitments.go` - Hash commitment generation
  - `internal/privacy/reference_id.go` - Opaque UUID system
  - `internal/privacy/data_classifier.go` - T0-T7 classification enforcer
- Tests: 12-15 test cases

**Priority 3.2: Merkle Aggregation**
- Dependency: Hash Commitments
- Deliverables:
  - `internal/blockchain/merkle_tree.go` - Merkle tree builder
  - `internal/blockchain/message_aggregator.go` - Batch aggregation
  - `internal/blockchain/proof_generator.go` - Merkle proof generation
- Tests: 12-15 test cases

### Phase 4: Advanced Features (Weeks 7-8)

**Priority 4.1: Zero-Knowledge Proofs**
- Dependency: snarkjs library, circuit files
- Deliverables:
  - `internal/zk/age_verifier.go` - Age verification proofs
  - `internal/zk/balance_prover.go` - Balance threshold proofs
  - `internal/zk/credential_prover.go` - Credential ownership proofs
  - Circuit files: `zk/circuits/*.circom`
- Tests: 18-20 test cases

**Priority 4.2: Audit Logging & Compliance**
- Dependency: Privacy Layer
- Deliverables:
  - `internal/audit/privacy_logger.go` - Privacy-preserving audit logs
  - `internal/audit/compliance_checker.go` - GDPR/CCPA compliance checks
- Tests: 10-12 test cases

---

## Implementation Details by Component

### 1. Secure Enclave Manager (iOS) - CRITICAL

**File**: `ios/Echo/Sources/Security/SecureEnclaveManager.swift`

**Interface**:
```swift
actor SecureEnclaveManager {
  // Key generation
  func generateBiometricProtectedKey(id: String) async throws -> String
  
  // Key derivation
  func deriveApplicationKeys(
    masterKeyId: String,
    context: String
  ) async throws -> DerivedKeys
  
  // Signing operations
  func sign(data: Data, keyId: String) async throws -> Data
  
  // Key rotation
  func rotateKeys(keyId: String) async throws
  
  // Key deletion
  func deleteKey(keyId: String) async throws
  
  // Public key export (only)
  func getPublicKey(keyId: String) throws -> Data
}
```

**Dependencies**:
- CryptoKit (Apple native)
- LocalAuthentication (Face ID/Touch ID)
- Security framework

**Test Coverage**:
- Key generation with biometric requirement
- Key derivation consistency
- Signing determinism
- Key rotation security
- Biometric gating enforcement
- Public key export correctness

---

### 2. Cryptographic Utilities (Go)

**File**: `internal/crypto/keyderivation.go`

**Interface**:
```go
// Key derivation using HKDF
func DeriveKey(masterKey []byte, salt []byte, info string) ([]byte, error)

// Biometric-protected key structure
type BiometricKey struct {
  KeyID       string
  PublicKey   *ecdsa.PublicKey
  PrivateKey  *ecdsa.PrivateKey
  CreatedAt   time.Time
  RotatedAt   time.Time
}

// Sign with ECDSA (P-256)
func Sign(message []byte, privateKey *ecdsa.PrivateKey) ([]byte, error)

// Verify signature
func Verify(message []byte, signature []byte, publicKey *ecdsa.PublicKey) bool
```

**Dependencies**:
- `crypto/ecdsa`, `crypto/sha256`, `crypto/rand`
- `golang.org/x/crypto/hkdf`
- `golang.org/x/crypto/argon2`

---

### 3. Hash Commitments

**File**: `internal/crypto/commitments.go`

**Interface**:
```go
// Create message commitment: H(H(plaintext) || nonce)
func CreateMessageCommitment(plaintext []byte) (Commitment, error)

type Commitment struct {
  Plaintext   string // base64url-encoded
  Nonce       string // base64url-encoded
  Commitment  string // base64url-encoded
  Timestamp   int64
}

// Salted hashing for PII
func HashWithSalt(data []byte, salt []byte) []byte

// Blind index for deterministic but unlinkable hashing
func CreateBlindIndex(data []byte, indexKey []byte) string

// Verify commitment integrity
func VerifyCommitment(commitment Commitment, plaintext []byte) bool
```

---

### 4. Message Encryption Service

**File**: `internal/services/encryption/e2ee.go`

**Interface**:
```go
type E2EEService struct {
  keyDerivation KeyDerivationService
  storage       EncryptedStorage
}

// Prepare message for transmission
func (svc *E2EEService) EncryptMessage(
  plaintext []byte,
  recipientPublicKey *ecdsa.PublicKey,
) (EncryptedMessage, error)

type EncryptedMessage struct {
  EphemeralPubKey string // base64url
  Ciphertext      string // base64url
  Nonce           string // base64url
  Tag             string // base64url
  Commitment      string // Hash commitment
  Signature       string // Sender signature
}

// Decrypt received message
func (svc *E2EEService) DecryptMessage(
  encryptedMsg EncryptedMessage,
  recipientPrivateKey *ecdsa.PrivateKey,
) ([]byte, error)
```

**Encryption Scheme**: ChaCha20-Poly1305 with ephemeral key agreement

---

### 5. Data Classification Enforcer

**File**: `internal/privacy/data_classifier.go`

**Interface**:
```go
type DataClassifier struct {
  rules map[DataTier]StorageRule
}

type DataTier int
const (
  T0_UltraSensitive DataTier = iota  // Biometrics, private keys
  T1_PII                              // Names, SSN, DOB
  T2_AccountSecrets                   // Passwords, auth tokens
  T3_PersonalComms                    // Messages, files
  T4_ContactInfo                      // Phone, email
  T5_Verification                     // Document images
  T6_AggregatedData                   // Trust scores
  T7_PublicData                       // DIDs, public keys
)

type StorageRule struct {
  AllowOnBlockchain bool
  RequiresEncryption bool
  MaxExposure StorageLocation
}

// Classify and validate data
func (dc *DataClassifier) Classify(data interface{}) (DataTier, error)
func (dc *DataClassifier) ValidateStorage(tier DataTier, location StorageLocation) error
```

---

### 6. Merkle Tree Aggregation

**File**: `internal/blockchain/merkle_tree.go`

**Interface**:
```go
type MerkleTree struct {
  leaves [][]byte
  tree   [][]byte
}

// Add message commitment to tree
func (mt *MerkleTree) AddLeaf(commitment []byte) error

// Generate tree and root
func (mt *MerkleTree) BuildTree() (root []byte, err error)

// Generate proof for specific leaf
func (mt *MerkleTree) GenerateProof(leafIndex int) (MerkleProof, error)

type MerkleProof struct {
  Index int
  Proof [][]byte // Path to root
}

// Verify proof
func (mt *MerkleTree) VerifyProof(
  leaf []byte,
  proof MerkleProof,
  root []byte,
) bool
```

---

### 7. Zero-Knowledge Proof System

**File**: `internal/zk/age_verifier.go` (example)

**Interface**:
```go
type ZKProofService struct {
  circuit CircuitLoader
  snarkjs SnarkJSWrapper
}

// Age verification ZK proof
func (zk *ZKProofService) ProveAgeOver(
  birthdate time.Time,
  threshold int,
) (ZKProof, error)

type ZKProof struct {
  Proof          string   // base64url proof
  PublicSignals  []string // Only: isOverThreshold=true
}

// Verify proof
func (zk *ZKProofService) VerifyProof(proof ZKProof, verificationKey string) bool
```

---

## Testing Strategy

### Unit Tests (Per Component)
- **Secure Enclave Manager**: 8-10 tests
- **Crypto Utilities**: 12-15 tests
- **Commitments**: 10-12 tests
- **Message Encryption**: 15-20 tests
- **Local Storage**: 10-12 tests
- **Data Classifier**: 12-15 tests
- **Merkle Tree**: 12-15 tests
- **ZK Proofs**: 18-20 tests
- **Audit Logging**: 10-12 tests

**Total**: ~115-140 test cases

### Integration Tests
- End-to-end message encryption and decryption
- Message → Blockchain commitment flow
- Biometric unlock → Storage access flow
- Privacy boundary enforcement

### Security Tests
- Key extraction prevention (Secure Enclave)
- Cryptographic randomness verification
- Collision resistance validation
- Signature verification
- Encryption determinism (correct), not always same output for same input with new nonce

---

## Deliverables Summary

### Phase 1 Deliverables (By End of Week 2)
```
✅ ios/Echo/Sources/Security/SecureEnclaveManager.swift
✅ ios/Echo/Sources/Security/SecureEnclaveManager.swift (tests)
✅ internal/crypto/keyderivation.go
✅ internal/crypto/hashing.go
✅ internal/crypto/signing.go
✅ internal/crypto/crypto_test.go
```

### Phase 2 Deliverables (By End of Week 4)
```
✅ internal/services/encryption/e2ee.go
✅ ios/Echo/Sources/Services/EncryptionService.swift
✅ internal/services/messaging/message_encryption.go
✅ ios/Echo/Sources/Storage/SecureLocalStorage.swift
✅ internal/storage/encrypted_storage.go
✅ [All test files]
```

### Phase 3 Deliverables (By End of Week 6)
```
✅ internal/crypto/commitments.go
✅ internal/privacy/reference_id.go
✅ internal/privacy/data_classifier.go
✅ internal/blockchain/merkle_tree.go
✅ internal/blockchain/message_aggregator.go
✅ internal/blockchain/proof_generator.go
✅ [All test files]
```

### Phase 4 Deliverables (By End of Week 8)
```
✅ internal/zk/age_verifier.go
✅ internal/zk/balance_prover.go
✅ internal/zk/credential_prover.go
✅ zk/circuits/age_verification.circom
✅ zk/circuits/balance_threshold.circom
✅ zk/circuits/credential_ownership.circom
✅ internal/audit/privacy_logger.go
✅ internal/audit/compliance_checker.go
✅ [All test files]
```

---

## Success Criteria

- [ ] All 115+ unit tests passing
- [ ] All 8 data tiers enforced (no T0-T5 data on blockchain)
- [ ] Message encryption provides semantic security
- [ ] Secure Enclave required for all key operations
- [ ] Zero-knowledge proofs verified off-chain before transaction
- [ ] Local storage encrypted at rest on all devices
- [ ] Audit logs contain zero PII
- [ ] GDPR/CCPA/HIPAA compliance verified
- [ ] Security audit (external) passes
- [ ] Performance: Key derivation < 500ms, encryption < 1s for 1MB message

---

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Secure Enclave key leakage | Use CryptoKit abstractions, never extract key material |
| Encryption key compromise | Perfect forward secrecy via ephemeral keys |
| Privacy boundary violation | Data classifier with enforcement gates |
| Merkle tree collisions | Use SHA256, BLAKE2b for proof robustness |
| ZK circuit vulnerabilities | Use proven snarkjs implementations, external audit |
| Performance degradation | Batch commitments, async operations |

---

## Next Steps

1. **Immediate (Today)**: Begin Secure Enclave Manager implementation
2. **Week 1**: Complete Phase 1 foundation
3. **Week 2-3**: Implement encryption and storage
4. **Week 4-6**: Build blockchain privacy layer
5. **Week 7-8**: Advanced features (ZK, audit)
6. **Week 9**: Integration testing and optimization
7. **Week 10**: Security audit and hardening

**Estimated Total Effort**: 10 weeks, 2-3 developers

---

## Questions Requiring Product Decision

1. **IDV Provider Selection**: Which third-party provider? (Stripe Identity, Sumsub, etc.)
2. **ZK Circuit Library**: Use snarkjs + circom, or alternatives?
3. **Blockchain Network**: Constellation testnet setup required?
4. **Performance SLA**: Encryption target latency?
5. **Audit Scope**: Internal or external security audit?

---

*Blueprint Implementation Plan v1.0*
*Created: February 17, 2026*
*Phase 1 Starting: Secure Enclave Integration*
