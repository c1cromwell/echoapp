// Core/Mesh/AnonymousMeshCrypto.swift
//
// E2E for the ANONYMOUS mesh lane: no identity, no DID — ephemeral keys per session/area (like
// bitchat's per-geohash keys). Noise-style cipher suite (X25519 + HKDF-SHA256 + ChaChaPoly) using
// an ephemeral sender key per message (forward secrecy on the sender side). The verified lane keeps
// using ECHO's identity-bound TextMessageCrypto; this is only for the permissionless lane.

import Foundation
import CryptoKit

public enum AnonymousMeshError: Error { case malformed }

public enum AnonymousMeshCrypto {
    private static let salt = Data("echo-mesh-anon-v1".utf8)
    private static let ephemeralKeyLen = 32

    /// Seal `plaintext` to a recipient's X25519 public key. Output = ephemeralSenderPub(32) ‖ box.
    public static func seal(_ plaintext: Data, to recipientPublicKey: Data) throws -> Data {
        guard let recipient = try? Curve25519.KeyAgreement.PublicKey(rawRepresentation: recipientPublicKey)
        else { throw AnonymousMeshError.malformed }
        let ephemeral = Curve25519.KeyAgreement.PrivateKey()
        let shared = try ephemeral.sharedSecretFromKeyAgreement(with: recipient)
        let key = shared.hkdfDerivedSymmetricKey(using: SHA256.self, salt: salt,
                                                 sharedInfo: Data(), outputByteCount: 32)
        let box = try ChaChaPoly.seal(plaintext, using: key)
        return ephemeral.publicKey.rawRepresentation + box.combined
    }

    /// Open a sealed payload with the recipient's session private key.
    public static func open(_ data: Data, recipient: Curve25519.KeyAgreement.PrivateKey) throws -> Data {
        guard data.count > ephemeralKeyLen else { throw AnonymousMeshError.malformed }
        let ephPub = try Curve25519.KeyAgreement.PublicKey(rawRepresentation: Data(data.prefix(ephemeralKeyLen)))
        let shared = try recipient.sharedSecretFromKeyAgreement(with: ephPub)
        let key = shared.hkdfDerivedSymmetricKey(using: SHA256.self, salt: salt,
                                                 sharedInfo: Data(), outputByteCount: 32)
        let box = try ChaChaPoly.SealedBox(combined: Data(data.dropFirst(ephemeralKeyLen)))
        return try ChaChaPoly.open(box, using: key)
    }

    /// A fresh ephemeral session identity for the anonymous lane (no DID, regenerate per area).
    public static func newSessionKey() -> Curve25519.KeyAgreement.PrivateKey {
        Curve25519.KeyAgreement.PrivateKey()
    }
}
