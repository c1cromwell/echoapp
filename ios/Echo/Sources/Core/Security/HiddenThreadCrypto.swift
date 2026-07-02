#if os(iOS)
import Foundation
import CryptoKit

/// AES-256-GCM encryption for hidden-conversation thread blobs at rest (WO-7/18).
@MainActor
enum HiddenThreadCrypto {
    private static let domain = Data("echo.hidden-thread.v1".utf8)
    private static let storageKeyId = "echo-hidden-threads"

    struct EncryptedThreadBlob: Codable, Sendable {
        let version: Int
        let ciphertext: Data
    }

    static let currentVersion = 1

    static func encrypt(messages: [StoredThreadMessage], conversationId: String) throws -> Data {
        let plaintext = try JSONEncoder().encode(messages)
        let key = deriveKey(conversationId: conversationId)
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw HiddenThreadCryptoError.sealFailed }
        let blob = EncryptedThreadBlob(version: currentVersion, ciphertext: combined)
        return try JSONEncoder().encode(blob)
    }

    static func decrypt(data: Data, conversationId: String) throws -> [StoredThreadMessage] {
        let blob = try JSONDecoder().decode(EncryptedThreadBlob.self, from: data)
        guard blob.version == currentVersion else { throw HiddenThreadCryptoError.unsupportedVersion }
        let key = deriveKey(conversationId: conversationId)
        let sealed = try AES.GCM.SealedBox(combined: blob.ciphertext)
        let plaintext = try AES.GCM.open(sealed, using: key)
        return try JSONDecoder().decode([StoredThreadMessage].self, from: plaintext)
    }

    private static func deriveKey(conversationId: String) -> SymmetricKey {
        if let sessionKey = HiddenChatsSession.shared.folderKey(for: conversationId) {
            return sessionKey
        }
        let base = SecureEnclaveManager.shared.deriveStorageKey(keyId: storageKeyId)
        let salt = Data(conversationId.utf8)
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: base,
            salt: salt,
            info: domain,
            outputByteCount: 32
        )
    }
}

enum HiddenThreadCryptoError: Error, Equatable {
    case sealFailed
    case unsupportedVersion
}
#endif
