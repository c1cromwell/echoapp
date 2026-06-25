import XCTest
@testable import Echo

/// P0 BLE-mesh core: wire frame, fragmentation, and the TTL/dedup/relay router. These are the
/// transport-agnostic pieces; the CoreBluetooth driver is verified on real hardware
/// (docs/MESH_TWO_DEVICE_TEST.md).
final class MeshPacketTests: XCTestCase {
    func testRoundTrip() {
        let p = MeshPacket(
            type: .message, ttl: 7, flags: .fragmented,
            messageID: MeshPacket.randomMessageID(),
            sender: MeshPacket.peerID(from: Data("alice".utf8)),
            recipient: MeshPacket.peerID(from: Data("bob".utf8)),
            payload: Data("hello mesh 🛰️".utf8)
        )
        let decoded = MeshPacket.decode(p.encoded())
        XCTAssertEqual(decoded, p)
    }

    func testIDsAreFixedWidth() {
        let p = MeshPacket(type: .message, ttl: 1, messageID: Data([1, 2]),
                           sender: Data([9]), recipient: MeshPacket.broadcast, payload: Data())
        XCTAssertEqual(p.messageID.count, MeshPacket.idSize)
        XCTAssertEqual(p.sender.count, MeshPacket.peerIDSize)
        XCTAssertTrue(p.isBroadcast)
    }

    func testDecodeRejectsGarbageAndShort() {
        XCTAssertNil(MeshPacket.decode(Data([0, 0, 0])))
        var bad = MeshPacket(type: .message, ttl: 1, messageID: MeshPacket.randomMessageID(),
                             sender: Data(), recipient: MeshPacket.broadcast, payload: Data("x".utf8)).encoded()
        bad[0] = 0xFF // wrong version
        XCTAssertNil(MeshPacket.decode(bad))
    }

    func testPeerIDIsDeterministic() {
        XCTAssertEqual(MeshPacket.peerID(from: Data("k".utf8)), MeshPacket.peerID(from: Data("k".utf8)))
        XCTAssertNotEqual(MeshPacket.peerID(from: Data("a".utf8)), MeshPacket.peerID(from: Data("b".utf8)))
    }
}

final class MeshFragmenterTests: XCTestCase {
    func testFragmentReassembleRoundTrip() {
        let payload = Data((0..<5000).map { UInt8($0 & 0xFF) })
        let gid = MeshPacket.randomMessageID()
        let frags = MeshFragmenter.fragment(payload, groupID: gid, maxChunk: 180)
        XCTAssertGreaterThan(frags.count, 1)

        let asm = MeshReassembler()
        var result: Data?
        for frag in frags.shuffled() {        // out-of-order delivery must still reassemble
            if let full = asm.add(frag) { result = full }
        }
        XCTAssertEqual(result, payload)
    }

    func testSmallPayloadIsSingleFragment() {
        let frags = MeshFragmenter.fragment(Data("hi".utf8), groupID: MeshPacket.randomMessageID(), maxChunk: 180)
        XCTAssertEqual(frags.count, 1)
    }
}

private final class CapturingDelegate: MeshRouterDelegate {
    var delivered: [(payload: Data, sender: Data)] = []
    var broadcasts: [MeshPacket] = []
    func meshRouter(_ router: MeshRouter, didReceive payload: Data, from sender: Data, messageID: Data) {
        delivered.append((payload, sender))
    }
    func meshRouter(_ router: MeshRouter, broadcast packet: MeshPacket) { broadcasts.append(packet) }
}

final class MeshRouterTests: XCTestCase {
    private let me = MeshPacket.peerID(from: Data("me".utf8))
    private let peer = MeshPacket.peerID(from: Data("peer".utf8))

    private func makeRouter() -> (MeshRouter, CapturingDelegate) {
        let r = MeshRouter(localID: me, maxPacketSize: 512)
        let d = CapturingDelegate()
        r.delegate = d
        return (r, d)
    }

    func testDeliversUnicastAddressedToUs() {
        let (r, d) = makeRouter()
        let p = MeshPacket(type: .message, ttl: 7, messageID: MeshPacket.randomMessageID(),
                           sender: peer, recipient: me, payload: Data("hi".utf8))
        r.ingest(p)
        XCTAssertEqual(d.delivered.count, 1)
        XCTAssertEqual(d.delivered.first?.payload, Data("hi".utf8))
        XCTAssertTrue(d.broadcasts.isEmpty, "Final unicast destination must not relay")
    }

    func testDedupDropsDuplicateFloods() {
        let (r, d) = makeRouter()
        let p = MeshPacket(type: .message, ttl: 7, messageID: MeshPacket.randomMessageID(),
                           sender: peer, recipient: me, payload: Data("hi".utf8))
        r.ingest(p); r.ingest(p); r.ingest(p)
        XCTAssertEqual(d.delivered.count, 1, "Duplicate messageIDs are dropped")
    }

    func testRelaysBroadcastWithDecrementedTTL() {
        let (r, d) = makeRouter()
        let p = MeshPacket(type: .message, ttl: 5, messageID: MeshPacket.randomMessageID(),
                           sender: peer, recipient: MeshPacket.broadcast, payload: Data("yo".utf8))
        r.ingest(p)
        XCTAssertEqual(d.delivered.count, 1, "Broadcast is delivered locally")
        XCTAssertEqual(d.broadcasts.count, 1, "and relayed onward")
        XCTAssertEqual(d.broadcasts.first?.ttl, 4)
    }

    func testRelaysUnicastForOthersWithoutDelivering() {
        let (r, d) = makeRouter()
        let other = MeshPacket.peerID(from: Data("other".utf8))
        let p = MeshPacket(type: .message, ttl: 5, messageID: MeshPacket.randomMessageID(),
                           sender: peer, recipient: other, payload: Data("x".utf8))
        r.ingest(p)
        XCTAssertTrue(d.delivered.isEmpty, "Not addressed to us")
        XCTAssertEqual(d.broadcasts.first?.ttl, 4, "but relayed toward the destination")
    }

    func testTTLOneIsNotRelayed() {
        let (r, d) = makeRouter()
        let p = MeshPacket(type: .message, ttl: 1, messageID: MeshPacket.randomMessageID(),
                           sender: peer, recipient: MeshPacket.broadcast, payload: Data("x".utf8))
        r.ingest(p)
        XCTAssertEqual(d.delivered.count, 1)
        XCTAssertTrue(d.broadcasts.isEmpty, "TTL exhausted — stop flooding")
    }

    func testDropsOwnEcho() {
        let (r, d) = makeRouter()
        let p = MeshPacket(type: .message, ttl: 7, messageID: MeshPacket.randomMessageID(),
                           sender: me, recipient: MeshPacket.broadcast, payload: Data("loop".utf8))
        r.ingest(p)
        XCTAssertTrue(d.delivered.isEmpty)
        XCTAssertTrue(d.broadcasts.isEmpty)
    }

    func testEndToEndFragmentedSendDelivers() {
        // Sender originates a large payload; a second node ingests every frame and reassembles.
        let (sender, sd) = makeRouter()
        let recipientRouter = MeshRouter(localID: peer, maxPacketSize: 512)
        let rd = CapturingDelegate()
        recipientRouter.delegate = rd

        let big = Data((0..<3000).map { UInt8($0 & 0xFF) })
        sender.send(payload: big, to: peer)
        XCTAssertGreaterThan(sd.broadcasts.count, 1, "Large payload fragments into multiple frames")
        for frame in sd.broadcasts { recipientRouter.ingest(frame) }

        XCTAssertEqual(rd.delivered.count, 1)
        XCTAssertEqual(rd.delivered.first?.payload, big)
    }
}
