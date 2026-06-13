#if os(iOS)
import Foundation
import CryptoKit

/// WO-64 phrase-encrypted `HistorySyncBundle` archive (BIP-39 entropy + HKDF + AES-GCM).
enum BackupCrypto {
    static let currentVersion = 1
    private static let saltLength = 16

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
        let sealed = try AES.GCM.seal(plaintext, using: key)
        guard let combined = sealed.combined else { throw BackupCryptoError.sealFailed }
        let archive = EncryptedBackup(version: currentVersion, salt: salt, ciphertext: combined)
        return try JSONEncoder().encode(archive)
    }

    static func decrypt(archiveData: Data, phrase: RecoveryPhrase) throws -> HistorySyncBundle {
        guard let entropy = phrase.entropy else { throw BackupCryptoError.invalidPhrase }
        let archive = try JSONDecoder().decode(EncryptedBackup.self, from: archiveData)
        guard archive.version == currentVersion else { throw BackupCryptoError.unsupportedVersion }
        let key = deriveKey(entropy: entropy, salt: archive.salt)
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

    var errorDescription: String? {
        switch self {
        case .invalidPhrase: return "Recovery phrase is invalid."
        case .sealFailed: return "Could not seal backup archive."
        case .unsupportedVersion: return "This backup version is not supported."
        }
    }
}
#endif
