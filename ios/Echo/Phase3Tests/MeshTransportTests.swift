import XCTest
import CryptoKit
@testable import Echo

private final class Box<T>: @unchecked Sendable { var value: T? }

private final class MockMeshNode: MeshNode {
    var onMessage: ((Data, Data) -> Void)?
    var sent: [(payload: Data, recipient: Data)] = []
    var started = false
    func start() { started = true }
    func stop() { started = false }
    func send(_ payload: Data, to recipient: Data) { sent.append((payload, recipient)) }
}

final class MeshSignalTransportTests: XCTestCase {
    func testConnectStartsNodeAndSendFloodsAsBroadcast() async throws {
        let node = MockMeshNode()
        let transport = MeshSignalTransport(node: node)
        try await transport.connect(accessToken: "ignored-offline")
        XCTAssertTrue(node.started)

        try await transport.send(text: #"{"type":"text"}"#)
        XCTAssertEqual(node.sent.first?.payload, Data(#"{"type":"text"}"#.utf8))
        XCTAssertEqual(node.sent.first?.recipient, MeshPacket.broadcast)

        await transport.disconnect()
        XCTAssertFalse(node.started)
    }

    func testInboundMeshPayloadFiresOnTextMessage() {
        let node = MockMeshNode()
        let transport = MeshSignalTransport(node: node)
        let box = Box<String>()
        transport.onTextMessage = { box.value = $0 }
        node.onMessage?(Data("hello-wire".utf8), MeshPacket.broadcast)
        XCTAssertEqual(box.value, "hello-wire")
    }
}

final class MeshPeerDirectoryTests: XCTestCase {
    /// A valid cert whose DID is the did:key of the signing key.
    private func makeValidCert() throws -> (MeshKeyCert, did: String, ka: Data) {
        let identity = P256.Signing.PrivateKey()
        let did = try DidKeyDeriver.deriveFromPublicKeyBytes(identity.publicKey.x963Representation)
        let ka = P256.KeyAgreement.PrivateKey().publicKey.x963Representation
        let cert = try MeshPeerDirectory.makeCert(did: did, identityKey: identity, kaPublicKey: ka)
        return (cert, did, ka)
    }

    func testValidCertVerifiesOffline() throws {
        let (cert, did, ka) = try makeValidCert()
        let result = MeshPeerDirectory.verify(cert)
        XCTAssertEqual(result?.did, did)
        XCTAssertEqual(result?.kaPublicKey, ka)
    }

    func testForgedDIDIsRejected() throws {
        var (cert, _, _) = try makeValidCert()
        // Claim a different DID than the signing key derives to.
        let other = P256.Signing.PrivateKey()
        let otherDID = try DidKeyDeriver.deriveFromPublicKeyBytes(other.publicKey.x963Representation)
        cert = MeshKeyCert(did: otherDID, signingPublicKey: cert.signingPublicKey,
                           kaPublicKey: cert.kaPublicKey, signature: cert.signature)
        XCTAssertNil(MeshPeerDirectory.verify(cert), "DID must be bound to the signing key")
    }

    func testTamperedKAKeyIsRejected() throws {
        var (cert, _, _) = try makeValidCert()
        let attackerKA = P256.KeyAgreement.PrivateKey().publicKey.x963Representation
        cert = MeshKeyCert(did: cert.did, signingPublicKey: cert.signingPublicKey,
                           kaPublicKey: attackerKA, signature: cert.signature)
        XCTAssertNil(MeshPeerDirectory.verify(cert), "Swapping the KA key breaks the signature")
    }

    func testCacheStoresVerifiedAndRejectsInvalid() throws {
        let (cert, did, ka) = try makeValidCert()
        let cache = MeshPeerCache()
        XCTAssertEqual(cache.ingest(cert), did)
        XCTAssertEqual(cache.kaPublicKey(forDID: did), ka)

        let bad = MeshKeyCert(did: did, signingPublicKey: cert.signingPublicKey,
                              kaPublicKey: cert.kaPublicKey, signature: Data(repeating: 0, count: 64))
        XCTAssertNil(cache.ingest(bad))
        XCTAssertEqual(cache.count, 1)
    }
}
