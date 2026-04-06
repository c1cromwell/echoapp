import Foundation
import CryptoKit

/// End-to-End Encryption service using Kinnami protocol
/// Provides message encryption and decryption with asymmetric key exchange
actor KinnamiEncryption {
    
    // MARK: - Properties
    
    private var recipientPublicKey: P256.KeyAgreement.PublicKey?
    private var sharedSecret: SymmetricKey?
    
    // MARK: - Key Agreement
    
    /// Generate a new ephemeral key pair for key agreement
    func generateEphemeralKeyPair() -> (privateKey: P256.KeyAgreement.PrivateKey, publicKey: Data) {
        let privateKey = P256.KeyAgreement.PrivateKey()
        let publicKey = privateKey.publicKey.rawRepresentation
        return (privateKey, publicKey)
    }
    
    /// Perform key agreement with recipient's public key
    /// - Parameters:
    ///   - privateKey: Our private key for this session
    ///   - recipientPublicKeyData: Recipient's public key data
    /// - Returns: Shared secret for symmetric encryption
    func performKeyAgreement(
        privateKey: P256.KeyAgreement.PrivateKey,
        recipientPublicKeyData: Data
    ) throws -> SymmetricKey {
        guard let recipientPublicKey = try? P256.KeyAgreement.PublicKey(rawRepresentation: recipientPublicKeyData) else {
            throw KinnamiError.invalidPublicKey
        }
        
        let sharedSecret = try privateKey.sharedSecretFromKeyAgreement(with: recipientPublicKey)
        
        // Derive a symmetric key from the shared secret
        let symmetricKey = sharedSecret.hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data("ECHO-E2E-KINNAMI".utf8),
            sharedInfo: Data("message-encryption".utf8),
            outputByteCount: 32
        )
        
        self.sharedSecret = symmetricKey
        return symmetricKey
    }
    
    // MARK: - Encryption
    
    /// Encrypt a message with the established shared secret
    /// - Parameters:
    ///   - plaintext: The message to encrypt
    ///   - aad: Additional authenticated data (optional)
    /// - Returns: Encrypted message with nonce
    func encrypt(
        plaintext: String,
        additionalAuthenticatedData: Data? = nil
    ) throws -> EncryptedMessage {
        guard let sharedSecret = sharedSecret else {
            throw KinnamiError.noSharedSecret
        }
        
        guard let data = plaintext.data(using: .utf8) else {
            throw KinnamiError.invalidInput
        }
        
        let sealedBox = try AES.GCM.seal(
            data,
            using: sharedSecret,
            authenticating: additionalAuthenticatedData
        )
        
        guard let nonce = sealedBox.nonce.withUnsafeBytes(Data.init) as Data? else {
            throw KinnamiError.encryptionFailed
        }
        
        let ciphertext = sealedBox.ciphertext
        let tag = sealedBox.tag
        
        return EncryptedMessage(
            nonce: nonce.base64EncodedString(),
            ciphertext: ciphertext.base64EncodedString(),
            tag: tag.base64EncodedString(),
            algorithm: "AES-256-GCM"
        )
    }
    
    /// Encrypt a message with key agreement in one operation
    /// - Parameters:
    ///   - plaintext: The message to encrypt
    ///   - recipientPublicKeyData: Recipient's public key
    /// - Returns: Encrypted message with ephemeral public key
    func encryptWithKeyAgreement(
        plaintext: String,
        recipientPublicKeyData: Data
    ) throws -> EncryptedMessageWithPublicKey {
        let (ephemeralPrivateKey, ephemeralPublicKey) = generateEphemeralKeyPair()
        
        let symmetricKey = try performKeyAgreement(
            privateKey: ephemeralPrivateKey,
            recipientPublicKeyData: recipientPublicKeyData
        )
        
        guard let data = plaintext.data(using: .utf8) else {
            throw KinnamiError.invalidInput
        }
        
        let sealedBox = try AES.GCM.seal(data, using: symmetricKey)
        
        guard let nonce = sealedBox.nonce.withUnsafeBytes(Data.init) as Data? else {
            throw KinnamiError.encryptionFailed
        }
        
        return EncryptedMessageWithPublicKey(
            ephemeralPublicKey: ephemeralPublicKey.base64EncodedString(),
            nonce: nonce.base64EncodedString(),
            ciphertext: sealedBox.ciphertext.base64EncodedString(),
            tag: sealedBox.tag.base64EncodedString(),
            algorithm: "AES-256-GCM-KINNAMI"
        )
    }
    
    // MARK: - Decryption
    
    /// Decrypt a message with the established shared secret
    /// - Parameters:
    ///   - encryptedMessage: The encrypted message
    ///   - aad: Additional authenticated data (optional)
    /// - Returns: Decrypted plaintext
    func decrypt(
        encryptedMessage: EncryptedMessage,
        additionalAuthenticatedData: Data? = nil
    ) throws -> String {
        guard let sharedSecret = sharedSecret else {
            throw KinnamiError.noSharedSecret
        }
        
        guard let nonce = try AES.GCM.Nonce(data: Data(base64Encoded: encryptedMessage.nonce) ?? Data()),
              let ciphertext = Data(base64Encoded: encryptedMessage.ciphertext),
              let tag = Data(base64Encoded: encryptedMessage.tag) else {
            throw KinnamiError.invalidEncryptedMessage
        }
        
        let sealedBox = try AES.GCM.SealedBox(nonce: nonce, ciphertext: ciphertext, tag: tag)
        let decryptedData = try AES.GCM.open(sealedBox, using: sharedSecret, authenticating: additionalAuthenticatedData)
        
        guard let plaintext = String(data: decryptedData, encoding: .utf8) else {
            throw KinnamiError.invalidDecryptedData
        }
        
        return plaintext
    }
    
    /// Decrypt a message with key agreement using sender's public key
    /// - Parameters:
    ///   - encryptedMessage: The encrypted message with ephemeral public key
    ///   - ourPrivateKey: Our private key for decryption
    /// - Returns: Decrypted plaintext
    func decryptWithKeyAgreement(
        encryptedMessage: EncryptedMessageWithPublicKey,
        ourPrivateKey: P256.KeyAgreement.PrivateKey
    ) throws -> String {
        guard let ephemeralPublicKeyData = Data(base64Encoded: encryptedMessage.ephemeralPublicKey),
              let ephemeralPublicKey = try? P256.KeyAgreement.PublicKey(rawRepresentation: ephemeralPublicKeyData) else {
            throw KinnamiError.invalidPublicKey
        }
        
        let sharedSecret = try ourPrivateKey.sharedSecretFromKeyAgreement(with: ephemeralPublicKey)
        
        let symmetricKey = sharedSecret.hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data("ECHO-E2E-KINNAMI".utf8),
            sharedInfo: Data("message-encryption".utf8),
            outputByteCount: 32
        )
        
        guard let nonce = try AES.GCM.Nonce(data: Data(base64Encoded: encryptedMessage.nonce) ?? Data()),
              let ciphertext = Data(base64Encoded: encryptedMessage.ciphertext),
              let tag = Data(base64Encoded: encryptedMessage.tag) else {
            throw KinnamiError.invalidEncryptedMessage
        }
        
        let sealedBox = try AES.GCM.SealedBox(nonce: nonce, ciphertext: ciphertext, tag: tag)
        let decryptedData = try AES.GCM.open(sealedBox, using: symmetricKey)
        
        guard let plaintext = String(data: decryptedData, encoding: .utf8) else {
            throw KinnamiError.invalidDecryptedData
        }
        
        return plaintext
    }
    
    // MARK: - Signature Operations
    
    /// Sign a message with a private key
    func sign(_ message: String, with privateKey: P256.Signing.PrivateKey) throws -> Data {
        guard let messageData = message.data(using: .utf8) else {
            throw KinnamiError.invalidInput
        }
        
        return try privateKey.signature(for: messageData).derRepresentation
    }
    
    /// Verify a message signature
    func verify(
        _ signature: Data,
        message: String,
        with publicKey: P256.Signing.PublicKey
    ) throws -> Bool {
        guard let messageData = message.data(using: .utf8) else {
            throw KinnamiError.invalidInput
        }
        
        guard let ecdsaSignature = try? P256.Signing.ECDSASignature(derRepresentation: signature) else {
            throw KinnamiError.invalidSignature
        }
        
        return publicKey.isValidSignature(ecdsaSignature, for: messageData)
    }
    
    // MARK: - Reset
    
    /// Clear the shared secret (call after message exchange)
    func resetSharedSecret() {
        sharedSecret = nil
    }
}

// MARK: - Data Models

struct EncryptedMessage: Codable {
    let nonce: String
    let ciphertext: String
    let tag: String
    let algorithm: String
}

struct EncryptedMessageWithPublicKey: Codable {
    let ephemeralPublicKey: String
    let nonce: String
    let ciphertext: String
    let tag: String
    let algorithm: String
}

// MARK: - Kinnami Errors

enum KinnamiError: LocalizedError {
    case invalidPublicKey
    case invalidPrivateKey
    case invalidSignature
    case invalidInput
    case invalidEncryptedMessage
    case invalidDecryptedData
    case noSharedSecret
    case encryptionFailed
    case decryptionFailed
    case keyAgreementFailed
    
    var errorDescription: String? {
        switch self {
        case .invalidPublicKey:
            return "Invalid public key format"
        case .invalidPrivateKey:
            return "Invalid private key format"
        case .invalidSignature:
            return "Invalid signature format"
        case .invalidInput:
            return "Invalid input data"
        case .invalidEncryptedMessage:
            return "Invalid encrypted message structure"
        case .invalidDecryptedData:
            return "Decrypted data is not valid UTF-8"
        case .noSharedSecret:
            return "No shared secret established for encryption"
        case .encryptionFailed:
            return "Encryption operation failed"
        case .decryptionFailed:
            return "Decryption operation failed"
        case .keyAgreementFailed:
            return "Key agreement operation failed"
        }
    }
}
