#if os(iOS)
import Foundation
import CryptoKit

/// WO-64 phrase-encrypted `HistorySyncBundle` archive (BIP-39 entropy + HKDF + AES-GCM).
enum BackupCrypto {
    static let currentVersion = 1
    private static let saltLength = 16
    static let deviceSalt = Data("echo-backup-device-salt-v1".utf8)

    struct EncryptedBackup: Codable, Sendable {
        let version: Int
        let salt: Data
        let ciphertext: Data
    }

    static func encrypt(bundle: HistorySyncBundle, phrase: RecoveryPhrase) throws -> Data {
        guard let entropy = phrase.entropy else { throw BackupCryptoError.invalidPhrase }
        let plaintext = try bundle.encoded()
        let salt = randomSalt()
        let key = deriveKey(entropy: entropy, salt: salt)
        return try seal(plaintext: plaintext, salt: salt, key: key)
    }

    static func encrypt(bundle: HistorySyncBundle, key: SymmetricKey) throws -> Data {
        let plaintext = try bundle.encoded()
        return try seal(plaintext: plaintext, salt: deviceSalt, key: key)
    }

    static func decrypt(archiveData: Data, phrase: RecoveryPhrase) throws -> HistorySyncBundle {
        guard let entropy = phrase.entropy else { throw BackupCryptoError.invalidPhrase }
        let archive = try JSONDecoder().decode(EncryptedBackup.self, from: archiveData)
        guard archive.version == currentVersion else { throw BackupCryptoError.unsupportedVersion }
        let key = deriveKey(entropy: entropy, salt: archive.salt)
        return try open(archive: archive, key: key)
    }

    static func decrypt(archiveData: Data, key: SymmetricKey) throws -> HistorySyncBundle {
        let archive = try JSONDecoder().decode(EncryptedBackup.self, from: archiveData)
        guard archive.version == currentVersion else { throw BackupCryptoError.unsupportedVersion }
        return try open(archive: archive, key: key)
    }

    static func deriveBackupKey(entropy: Data, salt: Data) -> SymmetricKey {
        deriveKey(entropy: entropy, salt: salt)
    }

    private static func seal(plaintext: Data, salt: Data, key: SymmetricKey) throws -> Data {
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw BackupCryptoError.sealFailed }
        let archive = EncryptedBackup(version: currentVersion, salt: salt, ciphertext: combined)
        return try JSONEncoder().encode(archive)
    }

    private static func open(archive: EncryptedBackup, key: SymmetricKey) throws -> HistorySyncBundle {
        let sealed = try AES.GCM.SealedBox(combined: archive.ciphertext)
        let plaintext = try AES.GCM.open(sealed, using: key)
        return try HistorySyncBundle.decode(from: plaintext)
    }

    private static func deriveKey(entropy: Data, salt: Data) -> SymmetricKey {
        let input = SymmetricKey(data: entropy)
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: input,
            salt: salt,
            info: Data("echo-message-backup-v1".utf8),
            outputByteCount: 32
        )
    }

    private static func randomSalt() -> Data {
        var bytes = [UInt8](repeating: 0, count: saltLength)
        _ = SecRandomCopyBytes(kSecRandomDefault, saltLength, &bytes)
        return Data(bytes)
    }
}

enum BackupCryptoError: LocalizedError {
    case invalidPhrase
    case sealFailed
    case unsupportedVersion
    case missingSessionKey

    var errorDescription: String? {
        switch self {
        case .invalidPhrase: return "Recovery phrase is invalid."
        case .sealFailed: return "Could not seal backup archive."
        case .unsupportedVersion: return "This backup version is not supported."
        case .missingSessionKey: return "Automatic backup is not configured."
        }
    }
}
#endif
