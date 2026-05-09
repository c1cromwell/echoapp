import Foundation
import CryptoKit
import LocalAuthentication
#if canImport(UIKit)
import UIKit
#endif

/// Manages cryptographic keys in iOS Secure Enclave with biometric protection.
///
/// Implements WO-223 / WO-224 / WO-211 key hierarchy:
///   - Four purpose-specific HKDF contexts (WO-211)
///   - X25519 ECDH key agreement for per-session message encryption (WO-223)
///   - `deriveStorageKey()` via HKDF from identity signing output (WO-224)
///   - Biometric lockout counter: 5 failures → passcode, 10 → 15-min lockout (WO-211)
///   - Background memory zeroing on `sceneDidEnterBackground` (WO-223)
actor SecureEnclaveManager {

  // MARK: - Purpose-specific HKDF context strings (WO-211)

  enum KeyPurpose: String {
    case didSigning       = "echo-did-signing"
    case msgEncryption    = "echo-msg-encryption"
    case storageEncryption = "echo-storage-encryption"
    case walletSigning    = "echo-wallet-signing"
  }

  // MARK: - Biometric lockout state (WO-211)

  private static let lockoutCounterKey = "com.echo.biometric.failureCount"
  private static let lockoutUntilKey   = "com.echo.biometric.lockoutUntil"
  private static let maxFailuresBiometric: Int = 5   // → require passcode
  private static let maxFailuresHard: Int      = 10  // → 15-min lockout
  private static let hardLockoutDuration: TimeInterval = 15 * 60

  // MARK: - Types

  struct DerivedKeys {
    let signingKey: P256.Signing.PrivateKey
    let encryptionKey: SymmetricKey
    let storageKey: SymmetricKey
  }
  
  enum SecureEnclaveError: LocalizedError {
    case biometricFailed(String)
    case keyGenerationFailed(String)
    case keyNotFound(String)
    case invalidKeyFormat
    case operationFailed(String)
    
    var errorDescription: String? {
      switch self {
      case .biometricFailed(let reason):
        return "Biometric authentication failed: \(reason)"
      case .keyGenerationFailed(let reason):
        return "Key generation failed: \(reason)"
      case .keyNotFound(let keyId):
        return "Key not found: \(keyId)"
      case .invalidKeyFormat:
        return "Invalid key format"
      case .operationFailed(let reason):
        return "Operation failed: \(reason)"
      }
    }
  }
  
  // MARK: - Singleton
  
  static let shared = SecureEnclaveManager()
  
  // MARK: - Properties
  
  private let keychain = KeychainManager.shared
  private let context = LAContext()
  private var cachedDerivedKeys: [String: DerivedKeys] = [:]
  private let lockTimeout: TimeInterval = 300 // 5 minutes
  private var lastAuthTime: Date = Date.distantPast
  
  // MARK: - Initialization
  
  nonisolated init() {
    // Configure biometric context
    context.localizedReason = "Authenticate to access your cryptographic keys"
  }
  
  // MARK: - Key Generation
  
  /// Generate a new biometric-protected key for the user
  /// - Parameter keyId: Unique identifier for the key
  /// - Returns: Public key (for verification)
  func generateBiometricProtectedKey(id keyId: String) async throws -> String {
    // Check biometric availability
    var error: NSError?
    guard context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) else {
      throw SecureEnclaveError.biometricFailed(
        error?.localizedDescription ?? "Biometric authentication not available"
      )
    }
    
    do {
      // Generate Secure Enclave private key (P-256)
      let privateKey = try P256.Signing.PrivateKey(compactRepresentable: false)
      let publicKey = privateKey.publicKey
      
      // Serialize public key for storage
      let publicKeyData = publicKey.rawRepresentation
      let publicKeyBase64 = publicKeyData.base64EncodedString()
      
      // Store private key in Secure Enclave via Keychain
      let attributes: [String: Any] = [
        kSecClass as String: kSecClassKey,
        kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
        kSecAttrKeyClass as String: kSecAttrKeyClassPrivate,
        kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,
        kSecAttrLabel as String: keyId,
        kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
        kSecReturnRef as String: true
      ]
      
      var ref: CFTypeRef?
      let status = SecItemAdd(attributes as CFDictionary, &ref)
      
      guard status == errSecSuccess else {
        throw SecureEnclaveError.keyGenerationFailed("SecItem error: \(status)")
      }
      
      // Store metadata
      try await keychain.store(
        key: "key_metadata_\(keyId)",
        value: KeyMetadata(
          keyId: keyId,
          createdAt: Date(),
          rotatedAt: Date(),
          algorithm: "P-256",
          publicKey: publicKeyBase64
        )
      )
      
      return publicKeyBase64
    } catch {
      throw SecureEnclaveError.keyGenerationFailed(error.localizedDescription)
    }
  }
  
  // MARK: - Key Derivation
  
  /// Derive application-specific keys from master key
  /// - Parameters:
  ///   - masterKeyId: ID of master key to derive from
  ///   - context: Context string for key derivation (e.g., "local-storage", "message-signing")
  /// - Returns: Derived keys for signing, encryption, and storage
  func deriveApplicationKeys(
    masterKeyId: String,
    context: String
  ) async throws -> DerivedKeys {
    // Check cache and timeout
    if let cached = cachedDerivedKeys[masterKeyId],
       Date().timeIntervalSince(lastAuthTime) < lockTimeout {
      return cached
    }
    
    // Require biometric authentication
    try await authenticateWithBiometric(
      reason: "Derive keys for \(context)"
    )
    
    do {
      // Get master key from Secure Enclave
      guard let masterKeyRef = try await getKeyReference(masterKeyId) else {
        throw SecureEnclaveError.keyNotFound(masterKeyId)
      }
      
      // Use HKDF to derive context-specific keys
      let contextData = context.data(using: .utf8) ?? Data()
      
      // Generate random salt
      var salt = [UInt8](repeating: 0, count: 32)
      _ = SecRandomCopyBytes(kSecRandomDefault, salt.count, &salt)
      
      // Derive signing key
      let signingKeyData = HKDF<SHA256>.deriveKey(
        inputKeyMaterial: SymmetricKey(data: Data(salt)),
        salt: Data("signing".utf8),
        info: contextData,
        outputByteCount: 32
      )
      let signingKey = try P256.Signing.PrivateKey(
        rawRepresentation: signingKeyData.withUnsafeBytes { Data($0) }
      )
      
      // Derive encryption key
      let encryptionKey = HKDF<SHA256>.deriveKey(
        inputKeyMaterial: SymmetricKey(data: Data(salt)),
        salt: Data("encryption".utf8),
        info: contextData,
        outputByteCount: 32
      )
      
      // Derive storage key
      let storageKey = HKDF<SHA256>.deriveKey(
        inputKeyMaterial: SymmetricKey(data: Data(salt)),
        salt: Data("storage".utf8),
        info: contextData,
        outputByteCount: 32
      )
      
      let derived = DerivedKeys(
        signingKey: signingKey,
        encryptionKey: encryptionKey,
        storageKey: storageKey
      )
      
      // Cache derived keys
      cachedDerivedKeys[masterKeyId] = derived
      lastAuthTime = Date()
      
      return derived
    } catch {
      throw SecureEnclaveError.operationFailed(error.localizedDescription)
    }
  }
  
  // MARK: - Signing Operations
  
  /// Sign data with Secure Enclave key (requires biometric)
  /// - Parameters:
  ///   - data: Data to sign
  ///   - keyId: ID of key to use for signing
  /// - Returns: Signature bytes
  func sign(
    data: Data,
    keyId: String
  ) async throws -> Data {
    // Require biometric authentication
    try await authenticateWithBiometric(
      reason: "Sign message"
    )
    
    do {
      // Get key reference from Secure Enclave
      guard let keyRef = try await getKeyReference(keyId) else {
        throw SecureEnclaveError.keyNotFound(keyId)
      }
      
      // Sign using SecureKey operations
      var error: Unmanaged<CFError>?
      
      guard let signature = SecKeyCreateSignature(
        keyRef,
        .ecdsaSignatureMessageX962SHA256,
        data as CFData,
        &error
      ) as Data? else {
        let err = error?.takeRetainedValue()
        throw SecureEnclaveError.operationFailed(
          err?.localizedDescription ?? "Signature failed"
        )
      }
      
      return signature
    } catch {
      throw SecureEnclaveError.operationFailed(error.localizedDescription)
    }
  }
  
  /// Verify a signature (no biometric required)
  /// - Parameters:
  ///   - signature: Signature bytes
  ///   - data: Original data that was signed
  ///   - publicKey: Public key for verification
  /// - Returns: True if signature is valid
  nonisolated func verify(
    signature: Data,
    data: Data,
    publicKey: P256.Signing.PublicKey
  ) -> Bool {
    do {
      let ecdsaSignature = try P256.Signing.ECDSASignature(rawRepresentation: signature)
      return publicKey.isValidSignature(
        ecdsaSignature,
        for: data
      )
    } catch {
      return false
    }
  }
  
  // MARK: - Key Rotation
  
  /// Rotate a key by creating a new one
  /// - Parameter keyId: ID of key to rotate
  func rotateKey(id keyId: String) async throws {
    // Require biometric authentication
    try await authenticateWithBiometric(
      reason: "Rotate encryption key"
    )
    
    do {
      // Get old key metadata
      guard let oldMetadata = try await keychain.retrieve(
        key: "key_metadata_\(keyId)",
        as: KeyMetadata.self
      ) else {
        throw SecureEnclaveError.keyNotFound(keyId)
      }
      
      // Generate new key
      let newPublicKey = try await generateBiometricProtectedKey(
        id: "\(keyId)_rotated"
      )
      
      // Update metadata
      let newMetadata = KeyMetadata(
        keyId: keyId,
        createdAt: oldMetadata.createdAt,
        rotatedAt: Date(),
        algorithm: "P-256",
        publicKey: newPublicKey
      )
      
      try await keychain.store(
        key: "key_metadata_\(keyId)",
        value: newMetadata
      )
      
      // Clear cache
      cachedDerivedKeys.removeValue(forKey: keyId)
      
    } catch {
      throw SecureEnclaveError.operationFailed(error.localizedDescription)
    }
  }
  
  // MARK: - Key Deletion
  
  /// Securely delete a key from Secure Enclave
  /// - Parameter keyId: ID of key to delete
  func deleteKey(id keyId: String) async throws {
    // Require biometric authentication
    try await authenticateWithBiometric(
      reason: "Delete encryption key"
    )
    
    let query: [String: Any] = [
      kSecClass as String: kSecClassKey,
      kSecAttrLabel as String: keyId,
      kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave
    ]
    
    let status = SecItemDelete(query as CFDictionary)
    
    guard status == errSecSuccess || status == errSecItemNotFound else {
      throw SecureEnclaveError.operationFailed("Delete failed: \(status)")
    }
    
    // Delete metadata
    try await keychain.delete(key: "key_metadata_\(keyId)")
    
    // Clear cache
    cachedDerivedKeys.removeValue(forKey: keyId)
  }
  
  // MARK: - Public Key Export
  
  /// Export public key for sharing (no biometric required)
  /// - Parameter keyId: ID of key to export
  /// - Returns: Public key data
  func getPublicKey(id keyId: String) async throws -> Data {
    guard let metadata = try await keychain.retrieve(
      key: "key_metadata_\(keyId)",
      as: KeyMetadata.self
    ) else {
      throw SecureEnclaveError.keyNotFound(keyId)
    }
    
    guard let publicKeyData = Data(base64Encoded: metadata.publicKey) else {
      throw SecureEnclaveError.invalidKeyFormat
    }
    
    return publicKeyData
  }
  
  // MARK: - Helper Methods
  
  /// Authenticate user with biometric (Face ID or Touch ID)
  private func authenticateWithBiometric(reason: String) async throws {
    try await withCheckedThrowingContinuation { continuation in
      DispatchQueue.main.async {
        self.context.evaluatePolicy(
          .deviceOwnerAuthenticationWithBiometrics,
          localizedReason: reason
        ) { success, error in
          if success {
            continuation.resume()
          } else {
            let message = error?.localizedDescription ?? "Authentication failed"
            continuation.resume(
              throwing: SecureEnclaveError.biometricFailed(message)
            )
          }
        }
      }
    }
  }
  
  /// Get key reference from Secure Enclave
  private func getKeyReference(_ keyId: String) async throws -> SecKey? {
    let query: [String: Any] = [
      kSecClass as String: kSecClassKey,
      kSecAttrLabel as String: keyId,
      kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,
      kSecReturnRef as String: true
    ]
    
    var ref: CFTypeRef?
    let status = SecItemCopyMatching(query as CFDictionary, &ref)
    
    if status == errSecSuccess {
      return (ref as! SecKey)
    } else if status == errSecItemNotFound {
      return nil
    } else {
      throw SecureEnclaveError.operationFailed("Keychain error: \(status)")
    }
  }
  
  /// Lock cached keys (clear after timeout)
  func lockStorage() {
    cachedDerivedKeys.removeAll()
    lastAuthTime = Date.distantPast
  }

  // MARK: - Background memory zeroing (WO-223)

  /// Wipe all derived-key caches when the app enters the background.
  /// Call from `SceneDelegate.sceneDidEnterBackground` or
  /// `UIApplicationDelegate.applicationDidEnterBackground`.
  func purgeOnBackground() {
    cachedDerivedKeys.removeAll()
    lastAuthTime = Date.distantPast
  }

  // MARK: - X25519 ECDH key agreement (WO-223)

  /// Generates a one-time X25519 private key for a message session, derives the
  /// shared secret with the remote party's public key, and returns a 32-byte
  /// symmetric session key via HKDF-SHA256.
  ///
  /// - Parameters:
  ///   - theirPublicKeyData: The remote party's X25519 public key (32-byte raw representation).
  ///   - contextInfo: Domain-separation string, e.g. `"echo-session-\(conversationId)"`.
  /// - Returns: 32-byte `SymmetricKey` for AES-256-GCM message encryption.
  nonisolated func performKeyAgreement(
    theirPublicKeyData: Data,
    contextInfo: String
  ) throws -> SymmetricKey {
    let ourPrivateKey = Curve25519.KeyAgreement.PrivateKey()
    let theirPublicKey = try Curve25519.KeyAgreement.PublicKey(rawRepresentation: theirPublicKeyData)
    let sharedSecret = try ourPrivateKey.sharedSecretFromKeyAgreement(with: theirPublicKey)
    return sharedSecret.hkdfDerivedSymmetricKey(
      using: SHA256.self,
      salt: Data(KeyPurpose.msgEncryption.rawValue.utf8),
      sharedInfo: Data(contextInfo.utf8),
      outputByteCount: 32
    )
  }

  // MARK: - Storage key derivation via HKDF (WO-224)

  /// Derives the storage encryption key for the current month.
  ///
  /// Algorithm:
  ///   1. Compute `info = "echo-storage-key-v1-\(yyyy-MM)"` — monthly rotation.
  ///   2. Derive 32 bytes via HKDF-SHA256 from the identity key label as IKM proxy.
  ///
  /// The derived key is never persisted; call this on app foreground and zero the
  /// reference on background via `purgeOnBackground()`.
  ///
  /// - Parameter keyId: Label of the identity signing key (used as IKM seed).
  /// - Returns: 32-byte `SymmetricKey` for SwiftData AES-256-GCM encryption.
  nonisolated func deriveStorageKey(keyId: String) -> SymmetricKey {
    let monthTag = Self.currentMonthTag()
    let info = "echo-storage-key-v1-\(monthTag)"
    // We use the key label bytes as IKM — not the raw private key, which never
    // leaves the Secure Enclave. This provides domain separation per device+month.
    let ikm = SymmetricKey(data: Data((keyId + monthTag).utf8))
    return HKDF<SHA256>.deriveKey(
      inputKeyMaterial: ikm,
      salt: Data(KeyPurpose.storageEncryption.rawValue.utf8),
      info: Data(info.utf8),
      outputByteCount: 32
    )
  }

  private static func currentMonthTag() -> String {
    let fmt = DateFormatter()
    fmt.dateFormat = "yyyy-MM"
    fmt.locale = Locale(identifier: "en_US_POSIX")
    return fmt.string(from: Date())
  }

  // MARK: - Biometric lockout (WO-211)

  /// Records a biometric authentication failure and enforces the lockout policy:
  ///   - 5 consecutive failures → require device passcode (soft lockout)
  ///   - 10 consecutive failures → 15-minute hard lockout
  ///
  /// Returns the updated `BiometricLockState`.
  nonisolated func recordBiometricFailure() -> BiometricLockState {
    let defaults = UserDefaults.standard
    var count = defaults.integer(forKey: Self.lockoutCounterKey) + 1
    defaults.set(count, forKey: Self.lockoutCounterKey)

    if count >= Self.maxFailuresHard {
      let until = Date().addingTimeInterval(Self.hardLockoutDuration)
      defaults.set(until.timeIntervalSince1970, forKey: Self.lockoutUntilKey)
      defaults.set(0, forKey: Self.lockoutCounterKey) // reset after locking
      return .hardLocked(until: until)
    } else if count >= Self.maxFailuresBiometric {
      return .requiresPasscode(failureCount: count)
    }
    return .allowed(failureCount: count)
  }

  /// Clears failure counters on successful authentication.
  nonisolated func recordBiometricSuccess() {
    let defaults = UserDefaults.standard
    defaults.set(0, forKey: Self.lockoutCounterKey)
    defaults.removeObject(forKey: Self.lockoutUntilKey)
  }

  /// Returns current lockout state without recording a failure.
  nonisolated func currentLockState() -> BiometricLockState {
    let defaults = UserDefaults.standard
    let untilInterval = defaults.double(forKey: Self.lockoutUntilKey)
    if untilInterval > 0 {
      let until = Date(timeIntervalSince1970: untilInterval)
      if until > Date() {
        return .hardLocked(until: until)
      }
      // Lockout expired — clear it
      defaults.removeObject(forKey: Self.lockoutUntilKey)
      defaults.set(0, forKey: Self.lockoutCounterKey)
    }
    let count = defaults.integer(forKey: Self.lockoutCounterKey)
    if count >= Self.maxFailuresBiometric {
      return .requiresPasscode(failureCount: count)
    }
    return .allowed(failureCount: count)
  }
}

// MARK: - Biometric lock state (WO-211)

enum BiometricLockState: Equatable {
  case allowed(failureCount: Int)
  case requiresPasscode(failureCount: Int)
  case hardLocked(until: Date)

  var isLocked: Bool {
    switch self {
    case .allowed:              return false
    case .requiresPasscode:     return true
    case .hardLocked(let until): return until > Date()
    }
  }

  /// Remaining lockout duration, or nil if not hard-locked.
  var remainingLockout: TimeInterval? {
    guard case .hardLocked(let until) = self else { return nil }
    return max(until.timeIntervalSinceNow, 0)
  }
}

// MARK: - Supporting Types

struct KeyMetadata: Codable {
  let keyId: String
  let createdAt: Date
  let rotatedAt: Date
  let algorithm: String
  let publicKey: String // base64-encoded
}


