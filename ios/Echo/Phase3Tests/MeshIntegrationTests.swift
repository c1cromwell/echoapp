import XCTest
import CryptoKit
@testable import Echo

private struct DummyError: Error {}

private final class MockTransport: ConversationSignalTransport, @unchecked Sendable {
    var onTextMessage: (@Sendable (String) -> Void)?
    var sent: [String] = []
    var connected = false
    var failConnect = false
    var failSend = false
    func connect(accessToken: String) async throws { if failConnect { throw DummyError() }; connected = true }
    func disconnect() async { connected = false }
    func send(text: String) async throws { if failSend { throw DummyError() }; sent.append(text) }
}

private final class Box<T>: @unchecked Sendable { var value: T? }

final class TransportRouterTests: XCTestCase {
    func testSendFansOutToEveryPath() async throws {
        let a = MockTransport(); let b = MockTransport()
        let router = TransportRouter([a, b])
        try await router.send(text: "hi")
        XCTAssertEqual(a.sent, ["hi"]); XCTAssertEqual(b.sent, ["hi"])
    }

    func testSendSucceedsIfAnyPathWorks() async throws {
        let ok = MockTransport(); let broken = MockTransport(); broken.failSend = true
        let router = TransportRouter([broken, ok])
        try await router.send(text: "x")           // must not throw
        XCTAssertEqual(ok.sent, ["x"])
    }

    func testSendThrowsOnlyIfEveryPathFails() async {
        let a = MockTransport(); a.failSend = true
        let b = MockTransport(); b.failSend = true
        let router = TransportRouter([a, b])
        do { try await router.send(text: "x"); XCTFail("expected throw") } catch {}
    }

    func testConnectToleratesAFailingPath() async throws {
        let ok = MockTransport(); let broken = MockTransport(); broken.failConnect = true
        let router = TransportRouter([broken, ok])
        try await router.connect(accessToken: "t")  // offline-friendly: no throw
        XCTAssertTrue(ok.connected)
    }

    func testInboundFromAnyChildIsMerged() {
        let a = MockTransport(); let b = MockTransport()
        let router = TransportRouter([a, b])
        let box = Box<String>()
        router.onTextMessage = { box.value = $0 }
        b.onTextMessage?("from-mesh")               // router rewired the child in init
        XCTAssertEqual(box.value, "from-mesh")
    }
}

final class MeshKeyAnnounceTests: XCTestCase {
    private func validCert() throws -> MeshKeyCert {
        let identity = P256.Signing.PrivateKey()
        let did = try DidKeyDeriver.deriveFromPublicKeyBytes(identity.publicKey.x963Representation)
        let ka = P256.KeyAgreement.PrivateKey().publicKey.x963Representation
        return try MeshPeerDirectory.makeCert(did: did, identityKey: identity, kaPublicKey: ka)
    }

    func testEncodeDecodeRoundTrip() throws {
        let cert = try validCert()
        XCTAssertEqual(MeshKeyAnnounce.decode(MeshKeyAnnounce.encode(cert)), cert)
    }

    func testDecodeGarbageIsNil() {
        XCTAssertNil(MeshKeyAnnounce.decode(Data("not json".utf8)))
    }

    func testAnnouncePayloadPopulatesPeerCache() throws {
        let cert = try validCert()
        let wire = MeshKeyAnnounce.encode(cert)            // what a peer broadcasts as .keyAnnounce
        let cache = MeshPeerCache()
        let decoded = try XCTUnwrap(MeshKeyAnnounce.decode(wire))
        XCTAssertEqual(cache.ingest(decoded), cert.did)
        XCTAssertEqual(cache.kaPublicKey(forDID: cert.did), cert.kaPublicKey)
    }
}

private final class MemStore: OfflineQueueStore {
    var items: [QueuedMeshMessage] = []
    func load() -> [QueuedMeshMessage] { items }
    func save(_ i: [QueuedMeshMessage]) { items = i }
}

final class OfflineMessageQueueTests: XCTestCase {
    private func msg(_ id: String) -> QueuedMeshMessage {
        QueuedMeshMessage(id: id, conversationId: "c", peerDID: "did:key:p", wire: "{}")
    }

    func testEnqueueIsIdempotent() {
        let q = OfflineMessageQueue(store: MemStore())
        XCTAssertTrue(q.enqueue(msg("1")))
        XCTAssertFalse(q.enqueue(msg("1")))
        XCTAssertEqual(q.pending.count, 1)
    }

    func testFlushDrainsDeliveredAndPersists() async {
        let store = MemStore()
        let q = OfflineMessageQueue(store: store)
        q.enqueue(msg("1")); q.enqueue(msg("2"))
        await q.flush { _ in true }                 // both accepted
        XCTAssertTrue(q.pending.isEmpty)
        XCTAssertTrue(store.items.isEmpty)
    }

    func testFlushKeepsUndeliveredUntilMaxAttempts() async {
        let q = OfflineMessageQueue(store: MemStore(), maxAttempts: 3)
        q.enqueue(msg("1"))
        await q.flush { _ in false }; XCTAssertEqual(q.pending.first?.attempts, 1)
        await q.flush { _ in false }; XCTAssertEqual(q.pending.first?.attempts, 2)
        await q.flush { _ in false }; XCTAssertTrue(q.pending.isEmpty, "given up after maxAttempts")
    }

    func testMarkDeliveredRemoves() {
        let q = OfflineMessageQueue(store: MemStore())
        q.enqueue(msg("1")); q.markDelivered("1")
        XCTAssertTrue(q.pending.isEmpty)
    }
}
