#if os(iOS)
import Foundation
import CryptoKit

/// Encrypts/decrypts chat bodies for WebSocket relay (Kinnami P-256 + identity registry keys).
actor TextMessageCrypto {
    private let identityResolve: IdentityResolveClient
    private struct CachedKey { let hex: String; let fetchedAt: Date }
    private var peerKeyCache: [String: CachedKey] = [:]
    /// Bound the cache so peer key rotation is eventually re-fetched and memory
    /// can't grow without limit across many conversations.
    private let peerKeyTTL: TimeInterval = 60 * 60 // 1 hour
    private let peerKeyCacheMax = 256

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
        // Sender authentication: sign (messageId, ciphertext) with our messaging key so
        // the recipient can verify the message genuinely came from us and was not altered.
        let senderDID = await CurrentUserSession.currentDID()
        // Ensure our messaging key is registered before the signed message goes out, so the
        // recipient can resolve the matching public key and verification doesn't spuriously
        // fail on a registration race for a brand-new account. Idempotent + flag-gated.
        if let did = senderDID {
            await MessagingKeyRegistrar().ensureRegistered(did: did)
        }
        let signingKey = try await MessagingAgreementKey.loadOrCreate()
        let signature = try? MessageSenderAuth.sign(
            messageId: messageId, encrypted: encrypted, signingKeyRaw: signingKey.rawRepresentation
        )
        return TextMessagePayload(
            messageId: messageId, text: nil, encrypted: encrypted,
            senderDID: senderDID, signature: signature
        )
    }

    /// Result of checking an inbound payload's sender-authentication signature.
    enum SenderVerification {
        case verified   // signature present and valid
        case unsigned   // no signature (legacy sender) — caller decides policy
        case invalid    // signature present but does not verify → reject
    }

    /// Verifies the sender-authentication signature on an inbound 1:1 payload against the
    /// public key the registry holds for `expectedSenderDID` (the relay-claimed sender).
    /// A pass proves the message came from that DID and was not tampered with in transit.
    func senderVerification(for payload: TextMessagePayload, expectedSenderDID: String) async -> SenderVerification {
        guard let signature = payload.signature, let encrypted = payload.encrypted else {
            return .unsigned
        }
        guard !expectedSenderDID.isEmpty,
              let hex = try? await cachedPeerKeyHex(peerDID: expectedSenderDID),
              let signerKey = try? Self.dataFromPublicKeyHex(hex) else {
            return .invalid
        }
        let ok = MessageSenderAuth.verify(
            messageId: payload.messageId, encrypted: encrypted,
            signature: signature, signerPublicKey: signerKey
        )
        return ok ? .verified : .invalid
    }

    func decryptPayload(_ payload: TextMessagePayload) async throws -> String {
        // Fail closed on receive: a private 1:1 payload must be encrypted. A
        // plaintext-only payload (no ciphertext) is rejected rather than displayed,
        // so a malicious or misconfigured relay can't downgrade a chat to cleartext.
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
        if let cached = peerKeyCache[peerDID],
           Date().timeIntervalSince(cached.fetchedAt) < peerKeyTTL {
            return cached.hex
        }
        let hex = try await identityResolve.primaryPublicKeyHex(peerDID: peerDID)
        if peerKeyCache.count >= peerKeyCacheMax {
            // Evict the oldest entry to stay bounded.
            if let oldest = peerKeyCache.min(by: { $0.value.fetchedAt < $1.value.fetchedAt }) {
                peerKeyCache.removeValue(forKey: oldest.key)
            }
        }
        peerKeyCache[peerDID] = CachedKey(hex: hex, fetchedAt: Date())
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

    /// Loads the dedicated messaging key-agreement private key (Option B). Backed by a
    /// software P-256 key in the Keychain, so ECDH decryption works on hardware as well
    /// as the Simulator — unlike the non-extractable Secure Enclave identity key.
    /// The matching public key is registered for peers by `MessagingKeyRegistrar`.
    static func loadAgreementPrivateKey() async throws -> P256.KeyAgreement.PrivateKey {
        try await MessagingAgreementKey.loadOrCreate()
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
