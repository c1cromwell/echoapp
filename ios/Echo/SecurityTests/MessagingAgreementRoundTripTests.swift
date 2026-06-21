// SecurityTests/MessagingAgreementRoundTripTests.swift
//
// Option B device-decrypt correctness. A peer encrypts to the holder's dedicated
// messaging key-agreement PUBLIC key (registered + resolved as SEC1/x963 hex), and the
// holder decrypts with the matching PRIVATE key via Kinnami. Software P-256 keys behave
// identically on Simulator and hardware, so a green round-trip here proves on-device
// decryption works — closing the prior Simulator-only gap in TextMessageCrypto.
//
// Kept in the clean SecurityTests target so pre-existing EchoTests issues don't block it.

import XCTest
import CryptoKit
@testable import Echo

final class MessagingAgreementRoundTripTests: XCTestCase {

    /// Mirror of `MessagingAgreementKey.publicKeyHex()`.
    private func x963Hex(_ key: P256.KeyAgreement.PrivateKey) -> String {
        key.publicKey.x963Representation.map { String(format: "%02x", $0) }.joined()
    }

    /// Mirror of the x963 → KeyAgreement raw normalisation used when a resolved peer key
    /// is fed into Kinnami (see `TextMessageCrypto.dataFromPublicKeyHex`).
    private func recipientKeyData(fromX963Hex hex: String) throws -> Data {
        let bytes = stride(from: 0, to: hex.count, by: 2).map { i -> UInt8 in
            let s = hex.index(hex.startIndex, offsetBy: i)
            let e = hex.index(s, offsetBy: 2)
            return UInt8(hex[s..<e], radix: 16)!
        }
        return try P256.Signing.PublicKey(x963Representation: Data(bytes)).rawRepresentation
    }

    func testRoundTripFromRegisteredPublicHex() async throws {
        let holder = P256.KeyAgreement.PrivateKey()           // the dedicated KA key
        let recipient = try recipientKeyData(fromX963Hex: x963Hex(holder))

        let kinnami = KinnamiEncryption()
        let plaintext = "the network is the asset 🛰️"
        let encrypted = try await kinnami.encryptWithKeyAgreement(
            plaintext: plaintext, recipientPublicKeyData: recipient
        )
        let decrypted = try await kinnami.decryptWithKeyAgreement(
            encryptedMessage: encrypted, ourPrivateKey: holder
        )
        XCTAssertEqual(decrypted, plaintext, "Holder must decrypt a message sent to its registered KA key")
    }

    func testWrongKeyCannotDecrypt() async throws {
        let holder = P256.KeyAgreement.PrivateKey()
        let attacker = P256.KeyAgreement.PrivateKey()
        let recipient = try recipientKeyData(fromX963Hex: x963Hex(holder))

        let kinnami = KinnamiEncryption()
        let encrypted = try await kinnami.encryptWithKeyAgreement(
            plaintext: "secret", recipientPublicKeyData: recipient
        )
        do {
            _ = try await kinnami.decryptWithKeyAgreement(encryptedMessage: encrypted, ourPrivateKey: attacker)
            XCTFail("A different KA key must not decrypt the message")
        } catch {
            // expected: decryption fails for the wrong key
        }
    }
}
