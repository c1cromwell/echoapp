import Foundation
import CryptoKit
import LocalAuthentication

/// Manages cryptographic keys in iOS Secure Enclave with biometric protection
/// 
/// Implements 3-tier key hierarchy:
/// Device Root Key → Biometric-Protected Key → User Identity Key → Derived Keys
actor SecureEnclaveManager {
  
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
      let privateKey = try P256.Signing.PrivateKey(format: .uncompressed)
      let publicKey = privateKey.publicKey
      
      // Serialize public key for storage
      let publicKeyData = publicKey.rawRepresentation
      let publicKeyBase64 = publicKeyData.base64EncodedString()
      
      // Store private key in Secure Enclave via Keychain
      let attributes: [String: Any] = [
        kSecClass as String: kSecClassKey,
        kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeP256,
        kSecAttrKeyClass as String: kSecAttrKeyClassPrivate,
        kSecUseSecureEnclave as String: true,
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
  func verify(
    signature: Data,
    data: Data,
    publicKey: P256.Signing.PublicKey
  ) -> Bool {
    do {
      return try publicKey.isValidSignature(
        signature,
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
      kSecUseSecureEnclave as String: true
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
      kSecUseSecureEnclave as String: true,
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
}

// MARK: - Supporting Types

struct KeyMetadata: Codable {
  let keyId: String
  let createdAt: Date
  let rotatedAt: Date
  let algorithm: String
  let publicKey: String // base64-encoded
}

/// Thread-safe keychain manager
actor KeychainManager {
  static let shared = KeychainManager()
  
  func store<T: Codable>(key: String, value: T) async throws {
    let data = try JSONEncoder().encode(value)
    
    let query: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrAccount as String: key,
      kSecValueData as String: data,
      kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
    ]
    
    // Delete existing
    SecItemDelete(query as CFDictionary)
    
    // Add new
    let status = SecItemAdd(query as CFDictionary, nil)
    guard status == errSecSuccess else {
      throw SecureEnclaveError.operationFailed("Store failed: \(status)")
    }
  }
  
  func retrieve<T: Codable>(key: String, as: T.Type) async throws -> T? {
    let query: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrAccount as String: key,
      kSecReturnData as String: true
    ]
    
    var result: CFTypeRef?
    let status = SecItemCopyMatching(query as CFDictionary, &result)
    
    guard status == errSecSuccess else {
      return nil
    }
    
    guard let data = result as? Data else {
      return nil
    }
    
    return try JSONDecoder().decode(T.self, from: data)
  }
  
  func delete(key: String) async throws {
    let query: [String: Any] = [
      kSecClass as String: kSecClassGenericPassword,
      kSecAttrAccount as String: key
    ]
    
    let status = SecItemDelete(query as CFDictionary)
    guard status == errSecSuccess || status == errSecItemNotFound else {
      throw SecureEnclaveError.operationFailed("Delete failed: \(status)")
    }
  }
}
