#if os(iOS)
import CryptoKit
import Foundation

/// Local encrypted backup of hidden conversations (WO-69).
enum HiddenFolderBackupManager {
    struct BackupPayload: Codable, Sendable {
        let exportedAt: Date
        let conversationIds: [String]
        let threads: [String: [StoredThreadMessage]]
        let preferences: [String: ConversationPreferences]
    }

    struct BackupFile: Codable, Sendable {
        let version: Int
        let salt: Data
        let ciphertext: Data
    }

    static let currentVersion = 1

    static func createBackup(phrase: RecoveryPhrase) async throws -> URL {
        guard let entropy = phrase.entropy else { throw BackupCryptoError.invalidPhrase }
        let payload = try await collectPayload()
        let plaintext = try JSONEncoder().encode(payload)
        var salt = [UInt8](repeating: 0, count: 16)
        _ = SecRandomCopyBytes(kSecRandomDefault, salt.count, &salt)
        let saltData = Data(salt)
        let key = BackupCrypto.deriveBackupKey(entropy: entropy, salt: saltData)
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw BackupCryptoError.sealFailed }
        let archive = BackupFile(version: currentVersion, salt: saltData, ciphertext: combined)
        let data = try JSONEncoder().encode(archive)
        let url = try backupURL()
        try data.write(to: url, options: .atomic)
        try (url as NSURL).setResourceValue(true, forKey: .isExcludedFromBackupKey)
        HiddenFolderAuditLog.record(.backupCreated)
        return url
    }

    static func restoreBackup(from url: URL, phrase: RecoveryPhrase) async throws {
        guard let entropy = phrase.entropy else { throw BackupCryptoError.invalidPhrase }
        let data = try Data(contentsOf: url)
        let archive = try JSONDecoder().decode(BackupFile.self, from: data)
        guard archive.version == currentVersion else { throw BackupCryptoError.unsupportedVersion }
        let key = BackupCrypto.deriveBackupKey(entropy: entropy, salt: archive.salt)
        let sealed = try AES.GCM.SealedBox(combined: archive.ciphertext)
        let plaintext = try AES.GCM.open(sealed, using: key)
        let payload = try JSONDecoder().decode(BackupPayload.self, from: plaintext)
        await MainActor.run {
            for (conversationId, prefs) in payload.preferences {
                ConversationPreferencesStore.shared.save(prefs, for: conversationId)
            }
            for (conversationId, messages) in payload.threads {
                ConversationThreadStore.replaceStored(conversationId: conversationId, messages: messages)
            }
        }
        HiddenFolderAuditLog.record(.backupRestored, detail: "\(payload.conversationIds.count) chats")
    }

    private static func collectPayload() async throws -> BackupPayload {
        await MainActor.run {
            let hiddenIds = ConversationPreferencesStore.shared.allHiddenConversationIds()
            var threads: [String: [StoredThreadMessage]] = [:]
            for id in hiddenIds {
                threads[id] = ConversationThreadStore.loadStoredOnly(conversationId: id)
            }
            var prefs: [String: ConversationPreferences] = [:]
            for id in hiddenIds {
                prefs[id] = ConversationPreferencesStore.shared.preferences(for: id)
            }
            return BackupPayload(
                exportedAt: Date(),
                conversationIds: hiddenIds,
                threads: threads,
                preferences: prefs
            )
        }
    }

    private static func backupURL() throws -> URL {
        let dir = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
            .appendingPathComponent("HiddenBackups", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let name = "hidden_backup_\(ISO8601DateFormatter().string(from: Date())).enc"
        return dir.appendingPathComponent(name)
    }
}
#endif
