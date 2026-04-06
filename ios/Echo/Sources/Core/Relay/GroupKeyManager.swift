import Foundation
import CryptoKit

/// Manages group encryption keys.
/// Group keys are symmetric (AES-256) and distributed to members
/// via individually encrypted E2E messages through the relay.
actor GroupKeyManager {

    private let encryption: KinnamiEncryption

    /// In-memory cache of group keys indexed by groupId -> version
    private var keyStore: [String: [Int: GroupKeyInfo]] = [:]

    struct GroupKeyInfo {
        let groupId: String
        let key: SymmetricKey
        let version: Int
        let receivedAt: Date
    }

    init(encryption: KinnamiEncryption) {
        self.encryption = encryption
    }

    // MARK: - Key Generation (Group Admin)

    /// Generate a new group key (called by group admin)
    func generateGroupKey(groupId: String) -> GroupKeyInfo {
        let key = SymmetricKey(size: .bits256)
        let version = (latestKeyVersion(groupId: groupId) ?? 0) + 1
        let info = GroupKeyInfo(
            groupId: groupId,
            key: key,
            version: version,
            receivedAt: Date()
        )
        storeGroupKey(info)
        return info
    }

    // MARK: - Group Message Encryption

    /// Encrypt a group message with the current group key (AES-256-GCM)
    func encryptForGroup(plaintext: Data, groupId: String) throws -> Data {
        guard let keyInfo = latestKey(groupId: groupId) else {
            throw GroupKeyError.noGroupKey
        }
        let nonce = AES.GCM.Nonce()
        let sealed = try AES.GCM.seal(plaintext, using: keyInfo.key, nonce: nonce)
        guard let combined = sealed.combined else {
            throw GroupKeyError.encryptionFailed
        }
        return combined
    }

    /// Decrypt a group message
    func decryptFromGroup(ciphertext: Data, groupId: String, keyVersion: Int) throws -> Data {
        guard let keyInfo = key(groupId: groupId, version: keyVersion) else {
            throw GroupKeyError.noGroupKey
        }
        let sealedBox = try AES.GCM.SealedBox(combined: ciphertext)
        return try AES.GCM.open(sealedBox, using: keyInfo.key)
    }

    // MARK: - Key Distribution

    /// Store a received group key (member receives via relay after admin distributes)
    func storeReceivedKey(groupId: String, version: Int, keyData: Data) {
        let key = SymmetricKey(data: keyData)
        let info = GroupKeyInfo(
            groupId: groupId,
            key: key,
            version: version,
            receivedAt: Date()
        )
        storeGroupKey(info)
    }

    /// Handle incoming group key distribution from WebSocket
    func handleKeyDistribution(_ payload: Data) throws {
        let distribution = try JSONDecoder().decode(WSGroupKeyPayload.self, from: payload)
        // In a real implementation, the encryptedKey would be decrypted with
        // the recipient's private key before storing. Here we store the raw key
        // data after the caller has already decrypted it.
        storeReceivedKey(
            groupId: distribution.groupId,
            version: distribution.version,
            keyData: distribution.encryptedKey
        )
    }

    // MARK: - Key Queries

    /// Get the latest key version for a group
    func latestKeyVersion(groupId: String) -> Int? {
        keyStore[groupId]?.keys.max()
    }

    /// Get the latest key for a group
    func latestKey(groupId: String) -> GroupKeyInfo? {
        guard let versions = keyStore[groupId],
              let maxVersion = versions.keys.max() else { return nil }
        return versions[maxVersion]
    }

    /// Get a specific key version for a group
    func key(groupId: String, version: Int) -> GroupKeyInfo? {
        keyStore[groupId]?[version]
    }

    /// Check if we have a key for a group
    func hasKey(groupId: String) -> Bool {
        keyStore[groupId] != nil && !(keyStore[groupId]?.isEmpty ?? true)
    }

    /// Get all stored group IDs
    func groupIds() -> [String] {
        Array(keyStore.keys)
    }

    // MARK: - Private Storage

    private func storeGroupKey(_ info: GroupKeyInfo) {
        if keyStore[info.groupId] == nil {
            keyStore[info.groupId] = [:]
        }
        keyStore[info.groupId]?[info.version] = info
    }
}

// MARK: - Errors

enum GroupKeyError: LocalizedError {
    case noGroupKey
    case encryptionFailed
    case decryptionFailed
    case keyRotationFailed
    case notGroupAdmin

    var errorDescription: String? {
        switch self {
        case .noGroupKey: return "No group key available"
        case .encryptionFailed: return "Group message encryption failed"
        case .decryptionFailed: return "Group message decryption failed"
        case .keyRotationFailed: return "Group key rotation failed"
        case .notGroupAdmin: return "Only group admins can rotate keys"
        }
    }
}
