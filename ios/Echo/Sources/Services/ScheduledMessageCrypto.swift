#if os(iOS)
import CryptoKit
import Foundation

/// Encrypts scheduled-message queue at rest (WO-76).
enum ScheduledMessageCrypto {
    private static let domain = Data("echo.scheduled.v1".utf8)
    private static let keyId = "echo-scheduled-messages"

    struct EncryptedQueue: Codable, Sendable {
        let version: Int
        let ciphertext: Data
    }

    static let currentVersion = 1

    static func encrypt(_ records: [ScheduledMessageRecord]) throws -> Data {
        let plaintext = try JSONEncoder().encode(records)
        let key = deriveKey()
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw ScheduledMessageCryptoError.sealFailed }
        let blob = EncryptedQueue(version: currentVersion, ciphertext: combined)
        return try JSONEncoder().encode(blob)
    }

    static func decrypt(_ data: Data) throws -> [ScheduledMessageRecord] {
        let blob = try JSONDecoder().decode(EncryptedQueue.self, from: data)
        guard blob.version == currentVersion else { throw ScheduledMessageCryptoError.unsupportedVersion }
        let key = deriveKey()
        let sealed = try AES.GCM.SealedBox(combined: blob.ciphertext)
        let plaintext = try AES.GCM.open(sealed, using: key)
        return try JSONDecoder().decode([ScheduledMessageRecord].self, from: plaintext)
    }

    /// Single opaque blob for `/v3/messages/schedule` (same device key as the local queue).
    static func encryptBlob(_ plaintext: Data) throws -> Data {
        let key = deriveKey()
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw ScheduledMessageCryptoError.sealFailed }
        return combined
    }

    static func decryptBlob(_ combined: Data) throws -> Data {
        let key = deriveKey()
        let sealed = try AES.GCM.SealedBox(combined: combined)
        return try AES.GCM.open(sealed, using: key)
    }

    private static func deriveKey() -> SymmetricKey {
        SecureEnclaveManager.shared.deriveStorageKey(keyId: keyId)
    }
}

enum ScheduledMessageCryptoError: Error {
    case sealFailed
    case unsupportedVersion
}
#endif
