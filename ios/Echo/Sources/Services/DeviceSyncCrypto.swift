#if os(iOS)
import Foundation
import CryptoKit

/// Pairwise P-256 ECDH wrap/unwrap for WO-CA3 sync ciphertext (Kinnami envelope).
actor DeviceSyncCrypto {
    private let encryption: KinnamiEncryption

    init(encryption: KinnamiEncryption = KinnamiEncryption()) {
        self.encryption = encryption
    }

    /// Seal a history bundle for the target device's P-256 key-agreement public key.
    func wrap(plaintext: Data, recipientPublicKey: Data) throws -> Data {
        guard !plaintext.isEmpty else { throw DeviceSyncCryptoError.emptyPlaintext }
        let pub = try Self.normalizedAgreementPublicKey(recipientPublicKey)
        let plaintextString = String(data: plaintext, encoding: .utf8)
            ?? plaintext.base64EncodedString()
        let envelope = try encryption.encryptWithKeyAgreement(
            plaintext: plaintextString,
            recipientPublicKeyData: pub
        )
        return try JSONEncoder().encode(envelope)
    }

    /// Open a sync entry with this device's P-256 key-agreement private key.
    func unwrap(ciphertext: Data, ourPrivateKey: P256.KeyAgreement.PrivateKey) throws -> Data {
        guard !ciphertext.isEmpty else { throw DeviceSyncCryptoError.emptyCiphertext }
        let envelope = try JSONDecoder().decode(EncryptedMessageWithPublicKey.self, from: ciphertext)
        let plain = try encryption.decryptWithKeyAgreement(
            encryptedMessage: envelope,
            ourPrivateKey: ourPrivateKey
        )
        if let data = Data(base64Encoded: plain), let _ = try? JSONDecoder().decode(HistorySyncBundle.self, from: data) {
            return data
        }
        guard let data = plain.data(using: .utf8) else {
            throw DeviceSyncCryptoError.invalidPlaintext
        }
        return data
    }

    /// Convenience: load local agreement key and unwrap.
    func unwrapWithLocalKey(ciphertext: Data) async throws -> Data {
        let privateKey = try await TextMessageCrypto.loadAgreementPrivateKey()
        return try unwrap(ciphertext: ciphertext, ourPrivateKey: privateKey)
    }

    /// Round-trip helper for tests: wrap to a key, unwrap with its private key.
    func roundTrip(plaintext: Data, recipientPublicKey: Data, recipientPrivateKey: P256.KeyAgreement.PrivateKey) throws -> Data {
        let wrapped = try wrap(plaintext: plaintext, recipientPublicKey: recipientPublicKey)
        return try unwrap(ciphertext: wrapped, ourPrivateKey: recipientPrivateKey)
    }

    private static func normalizedAgreementPublicKey(_ data: Data) throws -> Data {
        if let agreement = try? P256.KeyAgreement.PublicKey(rawRepresentation: data) {
            return agreement.rawRepresentation
        }
        let hex = data.map { String(format: "%02x", $0) }.joined()
        return try TextMessageCrypto.dataFromPublicKeyHex(hex)
    }
}

enum DeviceSyncCryptoError: LocalizedError {
    case emptyPlaintext
    case emptyCiphertext
    case invalidPlaintext

    var errorDescription: String? {
        switch self {
        case .emptyPlaintext: return "Nothing to encrypt for sync."
        case .emptyCiphertext: return "Sync ciphertext is empty."
        case .invalidPlaintext: return "Could not decode sync payload."
        }
    }
}
#endif
