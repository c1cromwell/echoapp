#if os(iOS)
import Foundation
import CryptoKit

/// Dedicated P-256 key-agreement keypair for end-to-end message encryption (Option B).
///
/// Why this exists: the identity signing key (`echo-identity-signing`) lives in the
/// Secure Enclave on device, where its private scalar is non-extractable. CryptoKit's
/// `sharedSecretFromKeyAgreement` needs a `P256.KeyAgreement.PrivateKey` backed by raw
/// bytes, which an Enclave key can never provide — so message decryption could only
/// ever work on the Simulator.
///
/// This key is a software P-256 key persisted in the Keychain
/// (`kSecAttrAccessibleWhenUnlockedThisDeviceOnly`, non-synced), so ECDH works
/// identically on Simulator and hardware. Its public half is registered as a labelled
/// device (`MessagingAgreementKey.deviceLabel`) so peers encrypt to it — see
/// `MessagingKeyRegistrar` and `IdentityResolveClient`.
enum MessagingAgreementKey {
    /// Keychain account for the persisted raw private scalar.
    static let keychainKey = "echo-messaging-agreement.privkey"
    /// Device label under which the public key is registered + resolved.
    static let deviceLabel = "msg-agreement"

    /// Loads the persisted key-agreement private key, generating and storing one on
    /// first use. Stable across launches so the registered public key stays valid.
    static func loadOrCreate() async throws -> P256.KeyAgreement.PrivateKey {
        if let data = try? await KeychainManager.shared.retrieveData(key: keychainKey),
           let key = try? P256.KeyAgreement.PrivateKey(rawRepresentation: data) {
            return key
        }
        let key = P256.KeyAgreement.PrivateKey()
        try await KeychainManager.shared.store(data: key.rawRepresentation, key: keychainKey)
        return key
    }

    /// SEC1 / X9.63 uncompressed public-key hex (65 bytes, `04 ‖ X ‖ Y`) for
    /// registration and peer resolution. Matches the encoding the DID registry stores.
    static func publicKeyHex() async throws -> String {
        let key = try await loadOrCreate()
        return key.publicKey.x963Representation.map { String(format: "%02x", $0) }.joined()
    }
}
#endif
