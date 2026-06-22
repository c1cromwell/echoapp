// SecurityTests/MessageSenderAuthTests.swift
//
// Sender authentication round trip. The sender signs (messageId, ciphertext) with the
// messaging key; the receiver verifies against the sender's public key. Proves a valid
// signature passes, and that tampering or a wrong/forged key is rejected — closing the
// "ECIES gives confidentiality but not authenticity" gap on the 1:1 path.
//
// MessageSenderAuth is UIKit-free, so this runs on the macOS host via `swift test`.

import XCTest
import CryptoKit
@testable import Echo

final class MessageSenderAuthTests: XCTestCase {

    private func makeEncrypted(_ plaintext: String, to recipient: P256.KeyAgreement.PrivateKey) async throws -> EncryptedMessageWithPublicKey {
        try await KinnamiEncryption().encryptWithKeyAgreement(
            plaintext: plaintext,
            recipientPublicKeyData: recipient.publicKey.rawRepresentation
        )
    }

    func testValidSignatureVerifies() async throws {
        let sender = P256.KeyAgreement.PrivateKey()        // messaging key (dual-use)
        let recipient = P256.KeyAgreement.PrivateKey()
        let encrypted = try await makeEncrypted("hello", to: recipient)
        let messageId = "msg-1"

        let signature = try MessageSenderAuth.sign(
            messageId: messageId, encrypted: encrypted, signingKeyRaw: sender.rawRepresentation
        )
        // Receiver verifies with the sender's public key (x963, as the registry returns it).
        XCTAssertTrue(MessageSenderAuth.verify(
            messageId: messageId, encrypted: encrypted,
            signature: signature, signerPublicKey: sender.publicKey.x963Representation
        ))
    }

    func testWrongSignerKeyIsRejected() async throws {
        let sender = P256.KeyAgreement.PrivateKey()
        let attacker = P256.KeyAgreement.PrivateKey()
        let recipient = P256.KeyAgreement.PrivateKey()
        let encrypted = try await makeEncrypted("hello", to: recipient)

        let signature = try MessageSenderAuth.sign(
            messageId: "msg-1", encrypted: encrypted, signingKeyRaw: sender.rawRepresentation
        )
        // Verifying against a different (attacker / spoofed sender) key must fail.
        XCTAssertFalse(MessageSenderAuth.verify(
            messageId: "msg-1", encrypted: encrypted,
            signature: signature, signerPublicKey: attacker.publicKey.x963Representation
        ))
    }

    func testTamperedMessageIdIsRejected() async throws {
        let sender = P256.KeyAgreement.PrivateKey()
        let recipient = P256.KeyAgreement.PrivateKey()
        let encrypted = try await makeEncrypted("hello", to: recipient)

        let signature = try MessageSenderAuth.sign(
            messageId: "msg-1", encrypted: encrypted, signingKeyRaw: sender.rawRepresentation
        )
        // Same signature, different messageId → canonical bytes change → rejected.
        XCTAssertFalse(MessageSenderAuth.verify(
            messageId: "msg-2", encrypted: encrypted,
            signature: signature, signerPublicKey: sender.publicKey.rawRepresentation
        ))
    }
}
