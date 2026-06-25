import XCTest
import CryptoKit
@testable import Echo

final class AnonymousMeshCryptoTests: XCTestCase {
    func testSealOpenRoundTrip() throws {
        let recipient = AnonymousMeshCrypto.newSessionKey()
        let pt = Data("meet at the square at 6 🕕".utf8)
        let sealed = try AnonymousMeshCrypto.seal(pt, to: recipient.publicKey.rawRepresentation)
        XCTAssertEqual(try AnonymousMeshCrypto.open(sealed, recipient: recipient), pt)
    }

    func testWrongRecipientCannotOpen() throws {
        let recipient = AnonymousMeshCrypto.newSessionKey()
        let attacker = AnonymousMeshCrypto.newSessionKey()
        let sealed = try AnonymousMeshCrypto.seal(Data("secret".utf8), to: recipient.publicKey.rawRepresentation)
        XCTAssertThrowsError(try AnonymousMeshCrypto.open(sealed, recipient: attacker))
    }

    func testEphemeralSenderKeyMakesEachSealUnique() throws {
        let recipient = AnonymousMeshCrypto.newSessionKey()
        let a = try AnonymousMeshCrypto.seal(Data("x".utf8), to: recipient.publicKey.rawRepresentation)
        let b = try AnonymousMeshCrypto.seal(Data("x".utf8), to: recipient.publicKey.rawRepresentation)
        XCTAssertNotEqual(a, b, "Fresh ephemeral sender key per message")
    }

    func testMalformedOpenThrows() {
        let recipient = AnonymousMeshCrypto.newSessionKey()
        XCTAssertThrowsError(try AnonymousMeshCrypto.open(Data([1, 2, 3]), recipient: recipient))
    }
}

final class MeshRateLimiterTests: XCTestCase {
    func testAllowsUpToLimitThenBlocks() {
        let now = Date()
        let rl = MeshRateLimiter(limit: 3, window: 60)
        XCTAssertTrue(rl.allow(now: now))
        XCTAssertTrue(rl.allow(now: now))
        XCTAssertTrue(rl.allow(now: now))
        XCTAssertFalse(rl.allow(now: now), "4th in-window is blocked")
    }

    func testWindowSlidesAndResets() {
        let t0 = Date()
        let rl = MeshRateLimiter(limit: 1, window: 60)
        XCTAssertTrue(rl.allow(now: t0))
        XCTAssertFalse(rl.allow(now: t0.addingTimeInterval(30)))
        XCTAssertTrue(rl.allow(now: t0.addingTimeInterval(61)), "old event aged out of window")
    }
}

final class MeshLanePolicyTests: XCTestCase {
    func testAnonymousFreeIsLimitedEphemeralAndRateLimited() {
        let p = MeshLanePolicy(lane: .anonymous, isVIP: false)
        XCTAssertEqual(p.cappedTTL(7), 3, "anon free hop ceiling")
        XCTAssertFalse(p.persists)
        XCTAssertFalse(p.allowsGroup(size: 16))
    }

    func testVerifiedIsFullFeaturedAndUnthrottled() {
        let p = MeshLanePolicy(lane: .verified, isVIP: false)
        XCTAssertEqual(p.cappedTTL(7), 7)
        XCTAssertTrue(p.persists)
        for _ in 0..<1000 { XCTAssertTrue(p.allowsSend(), "verified lane is not rate-limited") }
    }

    func testVIPLiftsAnonymousReach() {
        let free = MeshLanePolicy(lane: .anonymous, isVIP: false)
        let vip = MeshLanePolicy(lane: .anonymous, isVIP: true)
        XCTAssertGreaterThan(vip.cappedTTL(7), free.cappedTTL(7))
    }
}

final class GeohashTests: XCTestCase {
    func testKnownVector() {
        // Canonical example: (42.6, -5.6) -> "ezs42".
        XCTAssertEqual(Geohash.encode(latitude: 42.6, longitude: -5.6, precision: 5), "ezs42")
    }

    func testDeterministicAndPrecisionLength() {
        let a = Geohash.encode(latitude: 37.7749, longitude: -122.4194, precision: 7)
        XCTAssertEqual(a.count, 7)
        XCTAssertEqual(a, Geohash.encode(latitude: 37.7749, longitude: -122.4194, precision: 7))
    }

    func testNearbyPointsShareACoarsePrefix() {
        let a = Geohash.encode(latitude: 37.7749, longitude: -122.4194, precision: 5)
        let b = Geohash.encode(latitude: 37.7750, longitude: -122.4195, precision: 5)
        XCTAssertEqual(a, b, "Same block-level channel")
    }

    func testChannelPeerIDIsDeterministic() {
        let ch = MeshChannel.channelID(latitude: 37.7749, longitude: -122.4194, precision: 6)
        XCTAssertEqual(MeshChannel.peerID(forChannel: ch), MeshChannel.peerID(forChannel: ch))
    }
}

final class MeshPanicWipeTests: XCTestCase {
    func testClearsOnlyThreadAndMeshKeys() {
        let d = UserDefaults(suiteName: "mesh.panic.test")!
        defer { d.removePersistentDomain(forName: "mesh.panic.test") }
        d.set("a", forKey: "echo.thread.v1.conv-1")
        d.set("b", forKey: "echo.mesh.offlineQueue.v1")
        d.set("keep", forKey: "echo.trustTier")

        let removed = MeshPanicWipe.clearPersistedData(d)
        XCTAssertEqual(removed, 2)
        XCTAssertNil(d.object(forKey: "echo.thread.v1.conv-1"))
        XCTAssertNil(d.object(forKey: "echo.mesh.offlineQueue.v1"))
        XCTAssertEqual(d.string(forKey: "echo.trustTier"), "keep", "unrelated keys survive")
    }
}
