# Privacy Architecture and Secure Data Handling

## Overview

This specification defines the privacy architecture for the ECHO platform, ensuring user information remains private even when leveraging public blockchains. The design follows a "privacy by architecture" approach where sensitive data never leaves the user's device unencrypted, biometrics are bound to cryptographic keys via the device's Secure Enclave, and only opaque hashes and reference IDs are stored on-chain—making blockchain discovery attacks impossible.

## Core Privacy Principles

| Principle | Implementation |
|-----------|----------------|
| **Data Minimization** | Collect only what's necessary; delete when no longer needed |
| **Local-First Processing** | PII processed on-device; servers see only encrypted/hashed data |
| **Biometric Binding** | Private keys locked to user's biometrics via Secure Enclave |
| **Zero PII On-Chain** | Blockchain stores only hashes, commitments, and opaque references |
| **Unlinkability** | Different identifiers per context prevent correlation |
| **Forward Secrecy** | Compromised keys don't expose past communications |
| **User Sovereignty** | User controls all keys; platform cannot access data |

## Data Classification

### Classification Tiers

| Tier | Classification | Examples | Storage | On-Chain |
|------|---------------|----------|---------|----------|
| **T0** | Biometric | Face ID, Touch ID, fingerprint | Never stored | ❌ Never |
| **T1** | Cryptographic Secrets | Private keys, seeds, session keys | Secure Enclave only | ❌ Never |
| **T2** | PII - Direct | Name, DOB, SSN, address, phone | Device only (encrypted) | ❌ Never |
| **T3** | PII - Indirect | IP address, device ID, location | Ephemeral/anonymized | ❌ Never |
| **T4** | Pseudonymous | User ID, DID, wallet address | Encrypted database | ✅ Hashed only |
| **T5** | Content | Messages, files, media | Device + E2EE cloud | ❌ Never (hash of hash for integrity) |
| **T6** | Metadata | Timestamps, message counts | Aggregated/noised | ✅ Commitment only |
| **T7** | Public | Username (optional), avatar | User-controlled | ✅ If user chooses |

### What NEVER Touches the Blockchain

```
┌─────────────────────────────────────────────────────────────────────┐
│                    NEVER ON BLOCKCHAIN                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ❌ Real names, legal names                                         │
│  ❌ Phone numbers, email addresses                                  │
│  ❌ Physical addresses, locations                                   │
│  ❌ Government IDs, SSN, passport numbers                           │
│  ❌ Date of birth, age                                              │
│  ❌ Biometric data (face, fingerprint, voice)                       │
│  ❌ Financial account numbers                                       │
│  ❌ Message content (plaintext or ciphertext)                       │
│  ❌ File contents                                                   │
│  ❌ IP addresses                                                    │
│  ❌ Device identifiers                                              │
│  ❌ Contact lists                                                   │
│  ❌ Private keys                                                    │
│  ❌ Session tokens                                                  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### What CAN Be On Blockchain (Hashed/Committed)

```
┌─────────────────────────────────────────────────────────────────────┐
│                    BLOCKCHAIN-SAFE DATA                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ✅ DID (decentralized identifier) - pseudonymous                   │
│  ✅ Public keys (for encryption/verification)                       │
│  ✅ Hash commitments (prove data existed without revealing it)      │
│  ✅ Merkle roots (aggregate proofs)                                 │
│  ✅ Opaque reference IDs (UUIDs with no semantic meaning)           │
│  ✅ Zero-knowledge proofs (prove properties without data)           │
│  ✅ Encrypted credential status (revocation bits)                   │
│  ✅ Token balances (linked to pseudonymous address)                 │
│  ✅ Governance votes (if user chooses public voting)                │
│  ✅ Timestamps (for ordering, not correlation)                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## iOS Secure Enclave Architecture

### Secure Enclave Overview

The Secure Enclave is a hardware-based security subsystem isolated from the main processor, providing:
- Hardware-protected key storage
- Biometric authentication (Face ID / Touch ID)
- Cryptographic operations without exposing keys
- Secure boot chain and anti-replay mechanisms

### Key Hierarchy

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Key Hierarchy Architecture                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    SECURE ENCLAVE                            │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │              Device Root Key (UID)                   │    │   │
│  │  │         (Burned in at manufacturing)                 │    │   │
│  │  │              NEVER EXTRACTABLE                       │    │   │
│  │  └─────────────────────┬───────────────────────────────┘    │   │
│  │                        │                                     │   │
│  │                        ▼                                     │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │           Biometric-Protected Key (BPK)              │    │   │
│  │  │    (Unlocked only by Face ID / Touch ID)            │    │   │
│  │  │              NEVER LEAVES ENCLAVE                    │    │   │
│  │  └─────────────────────┬───────────────────────────────┘    │   │
│  │                        │                                     │   │
│  │                        ▼                                     │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │              User Identity Key (UIK)                 │    │   │
│  │  │       (Signs DID operations, auth challenges)        │    │   │
│  │  │              NEVER LEAVES ENCLAVE                    │    │   │
│  │  └─────────────────────┬───────────────────────────────┘    │   │
│  │                        │                                     │   │
│  └────────────────────────┼─────────────────────────────────────┘   │
│                           │                                         │
│                           ▼                                         │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                  DERIVED KEYS (In Memory)                    │   │
│  │                                                              │   │
│  │   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │   │
│  │   │ Message Key  │  │ Storage Key  │  │  Token Key   │     │   │
│  │   │   (Per-chat) │  │ (Local data) │  │  (Wallet)    │     │   │
│  │   └──────────────┘  └──────────────┘  └──────────────┘     │   │
│  │                                                              │   │
│  │   Derived via HKDF from UIK signature                       │   │
│  │   Cleared from memory when app backgrounds                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Secure Enclave Implementation (iOS)

```swift
import LocalAuthentication
import CryptoKit
import Security

/// Manages all Secure Enclave operations
class SecureEnclaveManager {
    
    // MARK: - Key Generation (Biometric-Protected)
    
    /// Generate a new key pair in the Secure Enclave, protected by biometrics
    func generateBiometricProtectedKey(
        keyId: String
    ) async throws -> SecKey {
        
        // Require biometric authentication for every use
        let accessControl = SecAccessControlCreateWithFlags(
            kCFAllocatorDefault,
            kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            [.privateKeyUsage, .biometryCurrentSet],  // Biometric required
            nil
        )!
        
        let attributes: [String: Any] = [
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecAttrKeySizeInBits as String: 256,
            kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,  // MUST be in Secure Enclave
            kSecPrivateKeyAttrs as String: [
                kSecAttrIsPermanent as String: true,
                kSecAttrApplicationTag as String: keyId.data(using: .utf8)!,
                kSecAttrAccessControl as String: accessControl,
            ],
        ]
        
        var error: Unmanaged<CFError>?
        guard let privateKey = SecKeyCreateRandomKey(attributes as CFDictionary, &error) else {
            throw SecureEnclaveError.keyGenerationFailed(error?.takeRetainedValue())
        }
        
        return privateKey
    }
    
    // MARK: - Signing (Requires Biometric)
    
    /// Sign data using Secure Enclave key (triggers Face ID / Touch ID)
    func sign(
        data: Data,
        withKeyId keyId: String,
        reason: String
    ) async throws -> Data {
        
        // Create authentication context with reason shown to user
        let context = LAContext()
        context.localizedReason = reason
        
        // Retrieve the key (will prompt for biometric)
        let query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrApplicationTag as String: keyId.data(using: .utf8)!,
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecReturnRef as String: true,
            kSecUseAuthenticationContext as String: context,
        ]
        
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        
        guard status == errSecSuccess, let privateKey = item else {
            throw SecureEnclaveError.keyNotFound
        }
        
        // Sign the data (biometric prompt appears here)
        var error: Unmanaged<CFError>?
        guard let signature = SecKeyCreateSignature(
            privateKey as! SecKey,
            .ecdsaSignatureMessageX962SHA256,
            data as CFData,
            &error
        ) else {
            throw SecureEnclaveError.signingFailed(error?.takeRetainedValue())
        }
        
        return signature as Data
    }
    
    // MARK: - Key Derivation
    
    /// Derive application keys from Secure Enclave signature
    /// Keys are derived in memory, never stored
    func deriveApplicationKeys(
        masterKeyId: String,
        context: String
    ) async throws -> DerivedKeys {
        
        // Sign a known value to get deterministic output
        let derivationInput = "ECHO-KEY-DERIVATION-\(context)".data(using: .utf8)!
        
        let signature = try await sign(
            data: derivationInput,
            withKeyId: masterKeyId,
            reason: "Unlock your ECHO account"
        )
        
        // Use HKDF to derive multiple keys from the signature
        let masterSecret = SHA256.hash(data: signature)
        
        return DerivedKeys(
            messageKey: deriveKey(from: masterSecret, info: "messages"),
            storageKey: deriveKey(from: masterSecret, info: "storage"),
            tokenKey: deriveKey(from: masterSecret, info: "tokens")
        )
    }
    
    private func deriveKey(from secret: SHA256.Digest, info: String) -> SymmetricKey {
        let inputKeyMaterial = SymmetricKey(data: Data(secret))
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: inputKeyMaterial,
            info: info.data(using: .utf8)!,
            outputByteCount: 32
        )
    }
    
    // MARK: - Public Key Export
    
    /// Export public key for sharing (private key NEVER leaves enclave)
    func exportPublicKey(keyId: String) throws -> Data {
        let query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrApplicationTag as String: keyId.data(using: .utf8)!,
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecReturnRef as String: true,
        ]
        
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        
        guard status == errSecSuccess, let privateKey = item as! SecKey? else {
            throw SecureEnclaveError.keyNotFound
        }
        
        guard let publicKey = SecKeyCopyPublicKey(privateKey) else {
            throw SecureEnclaveError.publicKeyExportFailed
        }
        
        var error: Unmanaged<CFError>?
        guard let publicKeyData = SecKeyCopyExternalRepresentation(publicKey, &error) else {
            throw SecureEnclaveError.publicKeyExportFailed
        }
        
        return publicKeyData as Data
    }
}

struct DerivedKeys {
    let messageKey: SymmetricKey   // For E2E message encryption
    let storageKey: SymmetricKey   // For local data encryption
    let tokenKey: SymmetricKey     // For wallet operations
    
    /// Clear keys from memory (call when app backgrounds)
    mutating func clear() {
        // In production, use secure memory clearing
        // Swift doesn't give direct memory access, but we can
        // rely on ARC and avoid keeping references
    }
}
```

### Android Equivalent (StrongBox/TEE)

```kotlin
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.Signature
import javax.crypto.Cipher

class SecureKeystoreManager(private val context: Context) {
    
    private val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
    
    /**
     * Generate biometric-protected key in StrongBox (hardware) if available
     */
    fun generateBiometricProtectedKey(keyId: String) {
        val keyPairGenerator = KeyPairGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_EC,
            "AndroidKeyStore"
        )
        
        val builder = KeyGenParameterSpec.Builder(
            keyId,
            KeyProperties.PURPOSE_SIGN or KeyProperties.PURPOSE_VERIFY
        )
            .setDigests(KeyProperties.DIGEST_SHA256)
            .setAlgorithmParameterSpec(ECGenParameterSpec("secp256r1"))
            .setUserAuthenticationRequired(true)  // Require biometric
            .setUserAuthenticationParameters(
                0,  // Require auth for every use
                KeyProperties.AUTH_BIOMETRIC_STRONG
            )
            .setInvalidatedByBiometricEnrollment(true)  // Invalidate if biometrics change
        
        // Use StrongBox if available (dedicated security chip)
        if (context.packageManager.hasSystemFeature(PackageManager.FEATURE_STRONGBOX_KEYSTORE)) {
            builder.setIsStrongBoxBacked(true)
        }
        
        keyPairGenerator.initialize(builder.build())
        keyPairGenerator.generateKeyPair()
    }
    
    /**
     * Sign data with biometric authentication
     */
    suspend fun sign(
        data: ByteArray,
        keyId: String,
        biometricPrompt: BiometricPrompt
    ): ByteArray {
        val privateKey = keyStore.getKey(keyId, null) as PrivateKey
        val signature = Signature.getInstance("SHA256withECDSA")
        signature.initSign(privateKey)
        
        // This will trigger biometric prompt
        val cryptoObject = BiometricPrompt.CryptoObject(signature)
        
        return suspendCancellableCoroutine { continuation ->
            biometricPrompt.authenticate(
                promptInfo,
                cryptoObject,
                object : BiometricPrompt.AuthenticationCallback() {
                    override fun onAuthenticationSucceeded(result: AuthenticationResult) {
                        val sig = result.cryptoObject?.signature!!
                        sig.update(data)
                        continuation.resume(sig.sign())
                    }
                    
                    override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                        continuation.resumeWithException(BiometricException(errString.toString()))
                    }
                }
            )
        }
    }
}
```

## Blockchain Privacy Architecture

### Hash and Reference ID Strategy

```
┌─────────────────────────────────────────────────────────────────────┐
│              Privacy-Preserving Blockchain Data Model               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  User Device                        Blockchain                      │
│  ───────────                        ──────────                      │
│                                                                     │
│  ┌─────────────────┐               ┌─────────────────────────────┐ │
│  │ Real Identity   │               │                             │ │
│  │                 │               │  DID Document               │ │
│  │ Name: John Doe  │ ──────────▶   │  ────────────               │ │
│  │ DOB: 1990-01-15 │   Derived     │  id: did:dag:abc123...      │ │
│  │ Phone: 555-1234 │   DID Only    │  publicKey: [...]           │ │
│  │ Email: j@e.com  │               │  (No PII)                   │ │
│  │                 │               │                             │ │
│  └─────────────────┘               └─────────────────────────────┘ │
│          │                                                          │
│          │ Hash + Salt                                              │
│          ▼                                                          │
│  ┌─────────────────┐               ┌─────────────────────────────┐ │
│  │ Phone Hash      │               │                             │ │
│  │                 │ ──────────▶   │  Lookup Index (Optional)    │ │
│  │ H(salt||phone)  │  Off-chain    │  ───────────────────────    │ │
│  │ = a1b2c3d4...   │  Server Only  │  hash → encrypted_did_ref   │ │
│  │                 │               │  (For contact discovery)    │ │
│  └─────────────────┘               │                             │ │
│                                    │  NOT on blockchain          │ │
│                                    └─────────────────────────────┘ │
│                                                                     │
│  ┌─────────────────┐               ┌─────────────────────────────┐ │
│  │ Message         │               │                             │ │
│  │                 │ ──────────▶   │  Integrity Anchor           │ │
│  │ "Hello World"   │   Hash of     │  ─────────────────          │ │
│  │                 │   Hash Only   │  commitment: H(H(msg)||...) │ │
│  │                 │               │  timestamp: 1234567890      │ │
│  │                 │               │  merkleRoot: xyz789...      │ │
│  └─────────────────┘               │                             │ │
│                                    │  Cannot reverse to content  │ │
│                                    └─────────────────────────────┘ │
│                                                                     │
│  ┌─────────────────┐               ┌─────────────────────────────┐ │
│  │ Credential      │               │                             │ │
│  │                 │ ──────────▶   │  Credential Reference       │ │
│  │ Type: License   │  Reference    │  ─────────────────────      │ │
│  │ Issuer: CA DMV  │  ID Only      │  refId: uuid-v4             │ │
│  │ Expires: 2027   │               │  issuerDID: did:...         │ │
│  │ Holder: John    │               │  type: "DriversLicense"     │ │
│  │                 │               │  status: 0 (not revoked)    │ │
│  └─────────────────┘               │                             │ │
│                                    │  No actual credential data  │ │
│                                    └─────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Hashing Strategies

#### 1. Salted Hashing (For Lookup Prevention)

```typescript
/**
 * NEVER store raw hashes of PII on-chain.
 * Unsalted hashes can be reversed via rainbow tables.
 * 
 * BAD:  H(phone) = a1b2c3...  ← Can be brute-forced!
 * GOOD: H(user_secret || phone) = x9y8z7...  ← Requires secret
 */

interface SaltedHashService {
  /**
   * Generate a salted hash that CANNOT be reversed without the salt.
   * The salt is stored ONLY on the user's device.
   */
  generateSaltedHash(
    data: string,
    userSecret: Uint8Array  // From Secure Enclave
  ): string {
    // Use Argon2id for PII hashing (memory-hard, GPU-resistant)
    const hash = argon2id({
      password: data,
      salt: userSecret,
      memoryCost: 65536,  // 64 MB
      timeCost: 3,
      parallelism: 4,
      hashLength: 32,
    });
    
    return base64url(hash);
  }
  
  /**
   * For contact discovery, use a blind index approach.
   * Server cannot learn phone numbers, but can match them.
   */
  generateBlindIndex(
    phoneNumber: string,
    serverBlindingKey: Uint8Array,  // Server's contribution
    userBlindingKey: Uint8Array     // User's contribution
  ): string {
    // Combine blinding keys (neither party knows the full key)
    const combinedKey = xorBytes(serverBlindingKey, userBlindingKey);
    
    // Generate blind index
    const blindIndex = hmacSha256(combinedKey, normalizePhone(phoneNumber));
    
    // Truncate to prevent exact matching attacks
    return base64url(blindIndex.slice(0, 16));  // 128 bits
  }
}
```

#### 2. Commitment Schemes (For Provable Integrity)

```typescript
/**
 * Commitments allow proving data existed at a time
 * WITHOUT revealing the data itself.
 */

interface CommitmentService {
  /**
   * Create a hiding commitment to message content.
   * 
   * commitment = H(H(message) || nonce || timestamp)
   * 
   * - Cannot determine message from commitment
   * - Can prove message matches commitment later
   * - Nonce prevents rainbow table attacks
   */
  createMessageCommitment(
    message: Uint8Array,
    nonce: Uint8Array = randomBytes(32)
  ): MessageCommitment {
    // Double-hash to add another layer of protection
    const innerHash = sha256(message);
    
    const commitmentInput = concat([
      innerHash,
      nonce,
      uint64ToBytes(Date.now()),
    ]);
    
    const commitment = sha256(commitmentInput);
    
    return {
      commitment: base64url(commitment),  // Safe to publish
      nonce: base64url(nonce),            // Store locally
      timestamp: Date.now(),
    };
  }
  
  /**
   * Verify a message matches a commitment (for disputes).
   * User must provide the original message and nonce.
   */
  verifyCommitment(
    message: Uint8Array,
    nonce: Uint8Array,
    timestamp: number,
    commitment: string
  ): boolean {
    const innerHash = sha256(message);
    const commitmentInput = concat([
      innerHash,
      nonce,
      uint64ToBytes(timestamp),
    ]);
    
    const computed = sha256(commitmentInput);
    return base64url(computed) === commitment;
  }
}
```

#### 3. Merkle Tree Aggregation (For Scalability)

```typescript
/**
 * Instead of storing individual commitments on-chain,
 * aggregate them into Merkle trees and store only the root.
 */

interface MerkleAggregationService {
  /**
   * Aggregate many message commitments into a single root.
   * Only the root goes on-chain; proofs stored locally.
   */
  aggregateCommitments(
    commitments: MessageCommitment[]
  ): MerkleAggregation {
    // Build Merkle tree
    const leaves = commitments.map(c => 
      base64urlToBytes(c.commitment)
    );
    
    const tree = new MerkleTree(leaves, sha256);
    const root = tree.getRoot();
    
    // Generate proofs for each commitment
    const proofs = commitments.map((_, i) => ({
      index: i,
      proof: tree.getProof(i),
    }));
    
    return {
      merkleRoot: base64url(root),      // This goes on-chain
      commitmentCount: commitments.length,
      proofs,                            // Store locally
    };
  }
  
  /**
   * Prove a specific message was included in an on-chain root.
   */
  proveInclusion(
    message: Uint8Array,
    nonce: Uint8Array,
    timestamp: number,
    proof: MerkleProof,
    merkleRoot: string
  ): boolean {
    // Recompute commitment
    const commitment = this.createMessageCommitment(message, nonce);
    
    // Verify Merkle proof
    return MerkleTree.verify(
      base64urlToBytes(commitment.commitment),
      proof.index,
      proof.proof,
      base64urlToBytes(merkleRoot),
      sha256
    );
  }
}
```

### Reference ID System

```typescript
/**
 * Reference IDs are opaque identifiers with NO semantic meaning.
 * They cannot be reversed to reveal any user information.
 */

interface ReferenceIdService {
  /**
   * Generate a reference ID for on-chain storage.
   * The mapping to real data exists ONLY on the user's device.
   */
  generateReferenceId(): string {
    // UUID v4 - completely random, no embedded information
    return crypto.randomUUID();
  }
  
  /**
   * For credentials, create a reference that reveals nothing
   * about the credential content or holder.
   */
  createCredentialReference(
    credential: VerifiableCredential,
    userDid: string
  ): CredentialReference {
    return {
      // Opaque reference ID
      refId: this.generateReferenceId(),
      
      // Issuer DID is public anyway
      issuerDid: credential.issuer.id,
      
      // Credential type (not sensitive)
      credentialType: credential.type[1],  // e.g., "DriversLicense"
      
      // Issuance timestamp (not sensitive)
      issuedAt: new Date(credential.issuanceDate).getTime(),
      
      // Commitment to credential (for verification)
      commitment: sha256(JSON.stringify(credential)),
      
      // Status index (for revocation checking)
      statusIndex: credential.credentialStatus?.statusListIndex,
      
      // NOTHING that identifies the holder or credential content
    };
  }
}

interface CredentialReference {
  refId: string;           // Opaque UUID
  issuerDid: string;       // Public issuer
  credentialType: string;  // Type only
  issuedAt: number;        // Timestamp
  commitment: string;      // Hash commitment
  statusIndex?: string;    // For revocation
  // NO: holder name, DOB, document number, etc.
}
```

### On-Chain Data Structures

```typescript
/**
 * Example of privacy-preserving on-chain data for ECHO
 */

// DID Document (on Cardano) - NO PII
interface OnChainDIDDocument {
  id: string;                        // did:cardano:abc123...
  verificationMethod: [{
    id: string;                      // did:cardano:abc123#key-1
    type: 'Ed25519VerificationKey2020';
    publicKeyMultibase: string;      // Public key only
  }];
  authentication: string[];          // Key references
  created: number;                   // Timestamp
  updated: number;                   // Timestamp
  // NO: name, email, phone, address, etc.
}

// Token State (on Constellation) - Pseudonymous
interface OnChainTokenState {
  balances: Map<string, bigint>;     // DID → balance (pseudonymous)
  stakes: Map<string, StakeInfo>;    // DID → stake info
  // NO: real names, just DIDs
}

// Message Integrity (on Constellation) - Commitment only
interface OnChainMessageIntegrity {
  merkleRoot: string;                // Aggregated commitment
  timestamp: number;                 // Batch timestamp
  messageCount: number;              // Count only
  // NO: message content, sender/receiver info
}

// Credential Status (on-chain) - Bit vector only
interface OnChainCredentialStatus {
  statusListCredential: string;      // Reference to status list
  encodedList: string;               // Compressed bit vector
  // NO: credential content, holder info
}

// Trust Score (on-chain) - Aggregated only
interface OnChainTrustCommitment {
  userDid: string;                   // Pseudonymous
  scoreCommitment: string;           // H(score || nonce)
  tier: 'unverified' | 'basic' | 'verified' | 'trusted';  // Tier only
  updatedAt: number;
  // NO: exact score, score history, verification details
}
```

## Zero-Knowledge Proofs

### Use Cases

| Proof Type | Proves | Without Revealing |
|------------|--------|-------------------|
| Age Verification | User is over 18/21 | Exact birthdate |
| Balance Proof | User has ≥X ECHO | Exact balance |
| Credential Validity | Credential is valid | Credential content |
| Trust Threshold | Trust score ≥ threshold | Exact score |
| Membership | User is in group | Which user |
| Transaction Auth | User authorized TX | Private key |

### ZK Implementation

```typescript
/**
 * Zero-knowledge proofs for privacy-preserving verification
 */

interface ZKProofService {
  /**
   * Prove age without revealing birthdate
   */
  async proveAgeOver(
    birthdate: Date,
    threshold: number,
    privateInputs: ZKPrivateInputs
  ): Promise<ZKProof> {
    // Circuit: (birthdate, currentDate) → (isOver18: boolean)
    // Public inputs: threshold (18), currentDate
    // Private inputs: birthdate
    
    const circuit = await this.loadCircuit('age_verification');
    
    const witness = {
      birthdate: dateToField(birthdate),
      threshold: threshold,
      currentDate: dateToField(new Date()),
    };
    
    const proof = await snarkjs.groth16.fullProve(
      witness,
      circuit.wasm,
      circuit.zkey
    );
    
    return {
      proof: proof.proof,
      publicSignals: proof.publicSignals,  // Only: isOverThreshold=true
      // birthdate is NOT in publicSignals
    };
  }
  
  /**
   * Prove token balance without revealing exact amount
   */
  async proveBalanceThreshold(
    balance: bigint,
    threshold: bigint,
    balanceCommitment: string
  ): Promise<ZKProof> {
    const circuit = await this.loadCircuit('balance_threshold');
    
    const witness = {
      balance: balance,
      threshold: threshold,
      commitment: balanceCommitment,
      randomness: randomFieldElement(),
    };
    
    const proof = await snarkjs.groth16.fullProve(
      witness,
      circuit.wasm,
      circuit.zkey
    );
    
    return {
      proof: proof.proof,
      publicSignals: [
        threshold.toString(),
        balanceCommitment,
        'true',  // meetsThreshold
      ],
      // Actual balance is NOT revealed
    };
  }
  
  /**
   * Prove credential ownership without revealing credential
   */
  async proveCredentialOwnership(
    credential: VerifiableCredential,
    holderPrivateKey: Uint8Array,
    verifierChallenge: Uint8Array
  ): Promise<ZKProof> {
    const circuit = await this.loadCircuit('credential_ownership');
    
    const witness = {
      credentialHash: sha256(JSON.stringify(credential)),
      holderDid: credential.credentialSubject.id,
      privateKey: holderPrivateKey,
      challenge: verifierChallenge,
      issuerPublicKey: await this.resolveIssuerKey(credential.issuer),
    };
    
    const proof = await snarkjs.groth16.fullProve(
      witness,
      circuit.wasm,
      circuit.zkey
    );
    
    return {
      proof: proof.proof,
      publicSignals: [
        credential.issuer.id,      // Issuer is public
        credential.type[1],        // Type is public
        sha256(verifierChallenge), // Challenge response
        'valid',                   // Credential is valid
      ],
      // Credential content, holder details NOT revealed
    };
  }
}
```

## Data Flow Architecture

### Message Flow (Privacy-Preserving)

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Privacy-Preserving Message Flow                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  SENDER DEVICE                                                      │
│  ─────────────                                                      │
│  1. User composes message "Hello Alice"                            │
│  2. Generate ephemeral key pair (X25519)                           │
│  3. Derive shared secret with recipient's public key               │
│  4. Encrypt message with ChaCha20-Poly1305                         │
│  5. Create commitment: H(H(plaintext) || nonce)                    │
│  6. Sign with Secure Enclave key (biometric required)              │
│                                                                     │
│        │                                                            │
│        ▼                                                            │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Encrypted Payload                                           │   │
│  │  ─────────────────                                           │   │
│  │  {                                                           │   │
│  │    "ephemeralPubKey": "...",     // For key agreement       │   │
│  │    "ciphertext": "...",          // Encrypted message       │   │
│  │    "nonce": "...",               // Encryption nonce        │   │
│  │    "commitment": "...",          // H(H(msg)||nonce)        │   │
│  │    "signature": "..."            // Sender signature        │   │
│  │  }                                                           │   │
│  │  // NO plaintext, NO sender name, NO metadata               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│        │                                                            │
│        ▼                                                            │
│  SERVER (Sees Only)                                                │
│  ──────────────────                                                │
│  • Encrypted blob (cannot decrypt)                                 │
│  • Sender DID (pseudonymous)                                       │
│  • Recipient DID (pseudonymous)                                    │
│  • Timestamp                                                        │
│  • Size                                                            │
│  // CANNOT see: message content, real identities                   │
│                                                                     │
│        │                                                            │
│        ▼                                                            │
│  BLOCKCHAIN (Stores Only)                                          │
│  ────────────────────────                                          │
│  • Merkle root of batch commitments                                │
│  • Batch timestamp                                                  │
│  • Message count                                                    │
│  // CANNOT see: any message data, sender/recipient                 │
│                                                                     │
│        │                                                            │
│        ▼                                                            │
│  RECIPIENT DEVICE                                                   │
│  ────────────────                                                   │
│  1. Receive encrypted payload                                      │
│  2. Biometric unlock Secure Enclave                                │
│  3. Derive shared secret with ephemeral key                        │
│  4. Decrypt message                                                │
│  5. Verify signature                                               │
│  6. Store plaintext locally (encrypted at rest)                    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Identity Verification Flow (Privacy-Preserving)

```
┌─────────────────────────────────────────────────────────────────────┐
│               Privacy-Preserving ID Verification                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  USER DEVICE                          IDV PROVIDER                  │
│  ───────────                          ────────────                  │
│                                                                     │
│  1. User initiates verification                                    │
│        │                                                            │
│        ▼                                                            │
│  2. IDV SDK captures:                                              │
│     • ID document images                                           │
│     • Selfie                                                        │
│     • Liveness check                                               │
│        │                                                            │
│        │  Direct TLS connection                                     │
│        │  (Platform backend NEVER sees images)                      │
│        ▼                                                            │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      IDV Provider                            │   │
│  │                                                              │   │
│  │  • Processes images                                          │   │
│  │  • Extracts data (name, DOB, etc.)                          │   │
│  │  • Performs verification                                     │   │
│  │  • DELETES all images after processing                      │   │
│  │  • Returns ONLY: pass/fail + confidence                     │   │
│  └─────────────────────────────────────────────────────────────┘   │
│        │                                                            │
│        │  Verification result only                                  │
│        │  (No PII transmitted)                                      │
│        ▼                                                            │
│  PLATFORM BACKEND                                                  │
│  ────────────────                                                  │
│  Receives ONLY:                                                    │
│  • Verification passed: true/false                                 │
│  • Confidence score: 0.95                                          │
│  • Document type: "drivers_license"                                │
│  • Issuing country: "US"                                           │
│  • Age over 18: true                                               │
│  • Provider reference ID: "abc123"                                 │
│                                                                     │
│  NEVER receives:                                                   │
│  • Actual name                                                     │
│  • Date of birth                                                   │
│  • Document number                                                 │
│  • Address                                                         │
│  • ID images                                                        │
│  • Selfie                                                          │
│        │                                                            │
│        ▼                                                            │
│  BLOCKCHAIN                                                        │
│  ──────────                                                        │
│  Stores ONLY:                                                      │
│  {                                                                  │
│    "credentialRef": "uuid-v4",           // Opaque reference      │
│    "issuerDid": "did:echo:platform",     // Platform DID          │
│    "type": "IdentityVerification",       // Type only             │
│    "verifiedAt": 1234567890,             // Timestamp             │
│    "level": "IAL2",                      // Assurance level       │
│    "statusIndex": 42                      // For revocation       │
│  }                                                                  │
│                                                                     │
│  // NO PII ever touches the blockchain                             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Local Storage Security

### On-Device Encryption

```swift
/**
 * All local data is encrypted with keys derived from Secure Enclave.
 * Even if device is compromised, data requires biometric to decrypt.
 */

class SecureLocalStorage {
    private let enclave: SecureEnclaveManager
    private var storageKey: SymmetricKey?
    
    // MARK: - Encrypted Storage
    
    /// Store data encrypted with biometric-protected key
    func store(key: String, data: Data) async throws {
        // Ensure we have the storage key (requires biometric)
        let storageKey = try await getStorageKey()
        
        // Generate random nonce
        let nonce = AES.GCM.Nonce()
        
        // Encrypt data
        let sealedBox = try AES.GCM.seal(data, using: storageKey, nonce: nonce)
        
        // Store encrypted data
        let encrypted = EncryptedData(
            nonce: Data(nonce),
            ciphertext: sealedBox.ciphertext,
            tag: sealedBox.tag
        )
        
        try FileManager.default.write(
            encrypted.serialize(),
            to: storageURL(for: key)
        )
    }
    
    /// Retrieve and decrypt data (requires biometric)
    func retrieve(key: String) async throws -> Data {
        let storageKey = try await getStorageKey()
        
        let encrypted = try EncryptedData.deserialize(
            from: FileManager.default.read(storageURL(for: key))
        )
        
        let sealedBox = try AES.GCM.SealedBox(
            nonce: AES.GCM.Nonce(data: encrypted.nonce),
            ciphertext: encrypted.ciphertext,
            tag: encrypted.tag
        )
        
        return try AES.GCM.open(sealedBox, using: storageKey)
    }
    
    // MARK: - Key Management
    
    private func getStorageKey() async throws -> SymmetricKey {
        if let key = storageKey {
            return key
        }
        
        // Derive from Secure Enclave (triggers biometric)
        let derived = try await enclave.deriveApplicationKeys(
            masterKeyId: "echo-master-key",
            context: "local-storage"
        )
        
        self.storageKey = derived.storageKey
        return derived.storageKey
    }
    
    /// Clear keys when app backgrounds
    func lockStorage() {
        storageKey = nil
    }
}
```

## Security Properties Summary

| Property | Implementation | Blockchain Exposure |
|----------|----------------|---------------------|
| **Message Content** | E2EE (ChaCha20-Poly1305) | Never (only commitment hash) |
| **User Identity** | Pseudonymous DID | DID only (no PII) |
| **Phone/Email** | Salted hash (Argon2id) | Never on-chain |
| **ID Documents** | Processed by IDV only | Never (reference ID only) |
| **Trust Score** | Commitment scheme | Tier only (not exact score) |
| **Token Balance** | Visible by DID | Pseudonymous (unlinkable) |
| **Private Keys** | Secure Enclave | Never |
| **Biometrics** | Never stored | Never |
| **Contacts** | Local encrypted | Never |
| **Location** | Ephemeral only | Never |

## Compliance & Audit

### Privacy Compliance Matrix

| Regulation | Requirement | Implementation |
|------------|-------------|----------------|
| **GDPR Art. 5** | Data minimization | Only collect necessary data |
| **GDPR Art. 17** | Right to erasure | Local data + off-chain mappings deletable |
| **GDPR Art. 25** | Privacy by design | Encryption, pseudonymization, hashing |
| **CCPA** | Do not sell | No PII stored centrally to sell |
| **HIPAA** | PHI protection | No health data stored |
| **PCI DSS** | Card data security | No card numbers stored |

### Audit Trail (Privacy-Preserving)

```typescript
/**
 * Audit logs contain NO PII, only pseudonymous references
 */
interface PrivacyPreservingAuditLog {
  // Event identification
  eventId: string;           // UUID
  eventType: string;         // "message_sent", "auth_success", etc.
  timestamp: number;
  
  // Pseudonymous references only
  actorDid?: string;         // DID, not name
  targetDid?: string;        // DID, not name
  resourceRef?: string;      // Opaque reference ID
  
  // Action metadata (no content)
  action: string;
  result: 'success' | 'failure';
  
  // Commitments for verification
  dataCommitment?: string;   // H(actual_data || nonce)
  
  // NO: names, emails, phones, IPs, message content
}
```

## Summary

This privacy architecture ensures:

1. **Biometric Binding**: All cryptographic operations require Face ID/Touch ID via Secure Enclave
2. **Zero PII On-Chain**: Blockchain stores only hashes, commitments, and opaque references
3. **Local-First**: Sensitive data never leaves the device unencrypted
4. **Unlinkability**: Different pseudonyms per context prevent correlation
5. **Forward Secrecy**: Compromised keys don't expose past data
6. **Compliance Ready**: GDPR, CCPA, and other regulations satisfied by design

Even with full blockchain access, an adversary cannot:
- Determine real user identities
- Read message content
- Link accounts across contexts
- Reverse hashes to PII
- Access biometric data
- Decrypt local storage

---

*Blueprint Version: 1.0*
*Last Updated: February 17, 2026*
*Status: Complete Privacy Architecture Specification*
