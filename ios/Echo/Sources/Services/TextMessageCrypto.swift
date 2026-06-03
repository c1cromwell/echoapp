#if os(iOS)
import Foundation
import CryptoKit

/// Encrypts/decrypts chat bodies for WebSocket relay (Kinnami P-256 + identity registry keys).
actor TextMessageCrypto {
    private let identityResolve: IdentityResolveClient
    private var peerKeyCache: [String: String] = [:]

    init(identityResolve: IdentityResolveClient) {
        self.identityResolve = identityResolve
    }

    func encryptPayload(plaintext: String, peerDID: String, messageId: String) async throws -> TextMessagePayload {
        let hex = try await cachedPeerKeyHex(peerDID: peerDID)
        let pubData = try Self.dataFromPublicKeyHex(hex)
        let kinnami = KinnamiEncryption()
        let encrypted = try await kinnami.encryptWithKeyAgreement(
            plaintext: plaintext,
            recipientPublicKeyData: pubData
        )
        return TextMessagePayload(messageId: messageId, text: nil, encrypted: encrypted)
    }

    func decryptPayload(_ payload: TextMessagePayload) async throws -> String {
        if let plain = payload.text, !plain.isEmpty, payload.encrypted == nil {
            return plain
        }
        guard let encrypted = payload.encrypted else {
            throw TextMessageCryptoError.missingCiphertext
        }
        let privateKey = try await Self.loadAgreementPrivateKey()
        let kinnami = KinnamiEncryption()
        return try await kinnami.decryptWithKeyAgreement(
            encryptedMessage: encrypted,
            ourPrivateKey: privateKey
        )
    }

    private func cachedPeerKeyHex(peerDID: String) async throws -> String {
        if let cached = peerKeyCache[peerDID] { return cached }
        let hex = try await identityResolve.primaryPublicKeyHex(peerDID: peerDID)
        peerKeyCache[peerDID] = hex
        return hex
    }

    static func dataFromPublicKeyHex(_ hex: String) throws -> Data {
        var cleaned = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        if cleaned.hasPrefix("0x") { cleaned = String(cleaned.dropFirst(2)) }
        guard cleaned.count % 2 == 0, let data = dataFromHex(cleaned), !data.isEmpty else {
            throw IdentityResolveError.invalidKeyMaterial
        }
        if let agreementKey = try? P256.KeyAgreement.PublicKey(rawRepresentation: data) {
            return agreementKey.rawRepresentation
        }
        if let signingKey = try? P256.Signing.PublicKey(x963Representation: data) {
            return signingKey.rawRepresentation
        }
        throw IdentityResolveError.invalidKeyMaterial
    }

    private static func dataFromHex(_ hex: String) -> Data? {
        var data = Data(capacity: hex.count / 2)
        var index = hex.startIndex
        while index < hex.endIndex {
            let next = hex.index(index, offsetBy: 2, limitedBy: hex.endIndex) ?? hex.endIndex
            guard next <= hex.endIndex else { return nil }
            let byte = hex[index..<next]
            guard let value = UInt8(byte, radix: 16) else { return nil }
            data.append(value)
            index = next
        }
        return data
    }

    /// Uses the same P-256 identity key material as `echo-identity-signing` (simulator + device).
    private static func loadAgreementPrivateKey() async throws -> P256.KeyAgreement.PrivateKey {
        #if targetEnvironment(simulator)
        if let data = try? KeychainManager.shared.retrieveData(key: "sim_privkey_echo-identity-signing"),
           let agreement = try? P256.KeyAgreement.PrivateKey(rawRepresentation: data) {
            return agreement
        }
        if let data = try? KeychainManager.shared.retrieveData(key: "sim_privkey_echo-identity-signing"),
           let signing = try? P256.Signing.PrivateKey(rawRepresentation: data),
           let agreement = try? P256.KeyAgreement.PrivateKey(rawRepresentation: signing.rawRepresentation) {
            return agreement
        }
        #endif
        throw TextMessageCryptoError.noLocalKey
    }
}

enum TextMessageCryptoError: LocalizedError {
    case missingCiphertext
    case noLocalKey

    var errorDescription: String? {
        switch self {
        case .missingCiphertext: return "Message payload could not be decrypted."
        case .noLocalKey: return "Unlock with Face ID to read encrypted messages on this device."
        }
    }
}
#endif
