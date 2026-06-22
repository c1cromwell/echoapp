import Foundation
import CryptoKit

/// Sender authentication for 1:1 messages.
///
/// Kinnami/ECIES gives *confidentiality* but not *authenticity*: anyone who knows the
/// recipient's public key can encrypt a message to them, and a malicious relay could
/// substitute a payload. The sender therefore signs a canonical binding of
/// `(messageId, ciphertext)` with the messaging key, and the receiver verifies it against
/// the sender's registered public key. A valid signature proves the message came from the
/// holder of that key (i.e. that DID) and was not altered in transit.
///
/// The signing key is the same P-256 key used for ECDH (`MessagingAgreementKey`): P-256
/// supports both ECDH and ECDSA, and reusing it avoids a second key + registration and a
/// per-message Secure-Enclave biometric prompt. (A dedicated signing key would be the
/// purist option; revisit if cross-protocol separation is required.)
///
/// Intentionally free of any `#if os(iOS)` / UIKit dependency so the round trip is
/// unit-testable on the macOS host.
enum MessageSenderAuth {

    /// Deterministic bytes covered by the signature: a version tag, the messageId, and the
    /// exact encrypted-envelope fields. Any tampering with the ciphertext, nonce, tag, or
    /// messageId invalidates the signature.
    static func canonicalBytes(messageId: String, encrypted: EncryptedMessageWithPublicKey) -> Data {
        let joined = [
            "echo-msg-auth-v1",
            messageId,
            encrypted.ephemeralPublicKey,
            encrypted.nonce,
            encrypted.ciphertext,
            encrypted.tag,
        ].joined(separator: "\n")
        return Data(joined.utf8)
    }

    /// Signs with the raw 32-byte P-256 scalar of the messaging key. Returns a P1363 (raw)
    /// ECDSA signature.
    static func sign(messageId: String, encrypted: EncryptedMessageWithPublicKey, signingKeyRaw: Data) throws -> Data {
        let key = try P256.Signing.PrivateKey(rawRepresentation: signingKeyRaw)
        let signature = try key.signature(for: canonicalBytes(messageId: messageId, encrypted: encrypted))
        return signature.rawRepresentation
    }

    /// Verifies `signature` over the canonical bytes using the sender's public key. Accepts
    /// either a 64-byte raw representation or a 65-byte X9.63/SEC1 uncompressed key, so it
    /// works directly with whatever the identity registry returns.
    static func verify(
        messageId: String,
        encrypted: EncryptedMessageWithPublicKey,
        signature: Data,
        signerPublicKey: Data
    ) -> Bool {
        guard let publicKey = signingPublicKey(from: signerPublicKey),
              let ecdsa = try? P256.Signing.ECDSASignature(rawRepresentation: signature) else {
            return false
        }
        return publicKey.isValidSignature(
            ecdsa,
            for: canonicalBytes(messageId: messageId, encrypted: encrypted)
        )
    }

    private static func signingPublicKey(from data: Data) -> P256.Signing.PublicKey? {
        if let key = try? P256.Signing.PublicKey(rawRepresentation: data) { return key }
        if let key = try? P256.Signing.PublicKey(x963Representation: data) { return key }
        return nil
    }
}
