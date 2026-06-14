#if os(iOS)
import Foundation
import CryptoKit

/// Device-local symmetric key for scheduled backups after the user opts in with their phrase.
enum BackupSessionKeyStore {
    private static let keychainKey = "echo.backup.sessionKey"

    static func hasStoredKey() -> Bool {
        (try? KeychainManager.shared.retrieve(key: keychainKey)) != nil
    }

    static func save(from phrase: RecoveryPhrase) throws {
        guard let entropy = phrase.entropy else { throw BackupCryptoError.invalidPhrase }
        let key = BackupCrypto.deriveBackupKey(entropy: entropy, salt: BackupCrypto.deviceSalt)
        let data = key.withUnsafeBytes { Data($0) }
        try KeychainManager.shared.store(value: data.base64EncodedString(), key: keychainKey)
    }

    static func loadSymmetricKey() throws -> SymmetricKey {
        guard let stored = try KeychainManager.shared.retrieve(key: keychainKey),
              let data = Data(base64Encoded: stored) else {
            throw BackupCryptoError.missingSessionKey
        }
        return SymmetricKey(data: data)
    }

    static func clear() {
        try? KeychainManager.shared.delete(key: keychainKey)
    }

    #if DEBUG
    static func resetForTesting() { clear() }
    #endif
}
#endif
