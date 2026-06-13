#if os(iOS)
import Foundation

/// WO-64 local encrypted backup + WO-CA2 optional cloud relay.
@MainActor
final class MessageBackupService {
    static let localFilename = "echo-message-backup.enc"
    private static let lastBackupKey = "echo.backup.lastAt"
    private static let lastSizeKey = "echo.backup.lastBytes"

    private let backupAPI: BackupAPIClient
    private let conversationStore: ConversationStore

    init(backupAPI: BackupAPIClient, conversationStore: ConversationStore = .shared) {
        self.backupAPI = backupAPI
        self.conversationStore = conversationStore
    }

    var localBackupURL: URL {
        FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
            .appendingPathComponent(Self.localFilename)
    }

    var lastBackupDate: Date? {
        UserDefaults.standard.object(forKey: Self.lastBackupKey) as? Date
    }

    var lastBackupByteCount: Int {
        UserDefaults.standard.integer(forKey: Self.lastSizeKey)
    }

    /// Export bundle, phrase-encrypt, write local file (excluded from iCloud by Documents policy).
    @discardableResult
    func createLocalBackup(phrase: RecoveryPhrase) async throws -> URL {
        let archive = try buildEncryptedArchive(phrase: phrase)
        let url = localBackupURL
        try archive.write(to: url, options: .atomic)
        recordBackupMetadata(byteCount: archive.count)
        return url
    }

    /// Decrypt local archive and merge into stores.
    @discardableResult
    func restoreLocalBackup(phrase: RecoveryPhrase) async throws -> Int {
        let data = try Data(contentsOf: localBackupURL)
        return try applyEncryptedArchive(data, phrase: phrase)
    }

    /// Phrase-encrypt and upload ciphertext to `/v3/backup/push`.
    func uploadCloudBackup(phrase: RecoveryPhrase) async throws {
        let archive = try buildEncryptedArchive(phrase: phrase)
        _ = try await backupAPI.push(ciphertext: archive)
        recordBackupMetadata(byteCount: archive.count)
        try archive.write(to: localBackupURL, options: .atomic)
    }

    /// Download cloud ciphertext, decrypt, merge.
    @discardableResult
    func restoreCloudBackup(phrase: RecoveryPhrase) async throws -> Int {
        let data = try await backupAPI.pull()
        try data.write(to: localBackupURL, options: .atomic)
        return try applyEncryptedArchive(data, phrase: phrase)
    }

    private func buildEncryptedArchive(phrase: RecoveryPhrase) throws -> Data {
        let bundle = HistorySyncBundleBuilder.build(from: conversationStore)
        return try BackupCrypto.encrypt(bundle: bundle, phrase: phrase)
    }

    @discardableResult
    private func applyEncryptedArchive(_ data: Data, phrase: RecoveryPhrase) throws -> Int {
        let bundle = try BackupCrypto.decrypt(archiveData: data, phrase: phrase)
        HistorySyncBundleMerger.apply(bundle, to: conversationStore)
        return bundle.conversations.count
    }

    private func recordBackupMetadata(byteCount: Int) {
        UserDefaults.standard.set(Date(), forKey: Self.lastBackupKey)
        UserDefaults.standard.set(byteCount, forKey: Self.lastSizeKey)
    }
}

enum MessageBackupError: LocalizedError {
    case invalidCloudPayload
    case missingLocalBackup

    var errorDescription: String? {
        switch self {
        case .invalidCloudPayload: return "Cloud backup payload is invalid."
        case .missingLocalBackup: return "No local backup file found."
        }
    }
}
#endif
