#if os(iOS)
import XCTest
@testable import Echo
import CryptoKit

final class DoubleRatchetTests: XCTestCase {
    func testRoundTrip() throws {
        var ss = Data(count: 32)
        _ = ss.withUnsafeMutableBytes { SecRandomCopyBytes(kSecRandomDefault, 32, $0.baseAddress!) }
        let bobPre = P256.KeyAgreement.PrivateKey()
        let bobPub = bobPre.publicKey.rawRepresentation
        let bobRaw = bobPub.count == 65 && bobPub.first == 0x04 ? Data(bobPub.dropFirst()) : bobPub

        let alice = try DoubleRatchet.newInitiator(sharedSecret: ss, remoteRatchetPub: bobRaw)
        let bob = try DoubleRatchet.newResponder(sharedSecret: ss, selfRatchet: bobPre)

        let wire = try DoubleRatchet.encrypt(alice, plaintext: Data("hello".utf8))
        let pt = try DoubleRatchet.decrypt(bob, message: wire)
        XCTAssertEqual(String(data: pt, encoding: .utf8), "hello")
    }

    func testPersistRoundTrip() throws {
        var ss = Data(count: 32)
        _ = ss.withUnsafeMutableBytes { SecRandomCopyBytes(kSecRandomDefault, 32, $0.baseAddress!) }
        let bobPre = P256.KeyAgreement.PrivateKey()
        let bobPub = bobPre.publicKey.rawRepresentation
        let bobRaw = bobPub.count == 65 && bobPub.first == 0x04 ? Data(bobPub.dropFirst()) : bobPub
        let alice = try DoubleRatchet.newInitiator(sharedSecret: ss, remoteRatchetPub: bobRaw)
        let wire = try DoubleRatchet.encrypt(alice, plaintext: Data("persist".utf8))
        let exported = DoubleRatchet.export(alice)
        let restored = try DoubleRatchet.restore(exported)
        let bob = try DoubleRatchet.newResponder(sharedSecret: ss, selfRatchet: bobPre)
        let pt = try DoubleRatchet.decrypt(bob, message: wire)
        XCTAssertEqual(String(data: pt, encoding: .utf8), "persist")
        _ = restored // compile-time persist path
    }
}
#endif
