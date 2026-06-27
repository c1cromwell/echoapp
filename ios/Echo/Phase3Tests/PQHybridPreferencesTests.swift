#if os(iOS)
import XCTest
@testable import Echo

final class PQHybridPreferencesTests: XCTestCase {
    func testInactiveUntilPlatformSupport() {
        PQHybridPreferences.usePQBootstrap = true
        XCTAssertFalse(PQHybridBootstrap.isPlatformSupported)
        XCTAssertFalse(PQHybridBootstrap.isActive)
        PQHybridPreferences.usePQBootstrap = false
    }

    func testCachesPeerHybridBundle() {
        let peer = "did:key:zPQPeer"
        let bundle = HybridPublicBundleWire(ec: "ecB64", pq: "pqB64")
        PQHybridBootstrap.cachePeerHybridBundle(peerDID: peer, bundle: bundle)
        XCTAssertEqual(PQHybridBootstrap.cachedPeerHybridBundle(peerDID: peer), bundle)
        PQHybridBootstrap.cachePeerHybridBundle(peerDID: peer, bundle: nil)
        XCTAssertNil(PQHybridBootstrap.cachedPeerHybridBundle(peerDID: peer))
    }

    func testRatchetPreKeyPayloadEncodesOptionalHybrid() throws {
        let wire = try ConversationSignalCodec.encodeRatchetPreKey(
            to: "did:key:zBob",
            conversationId: "conv-1",
            ratchetPublicKeyB64: "cmF0Y2hldA==",
            hybridPublicBundle: HybridPublicBundleWire(ec: "ec", pq: "pq")
        )
        XCTAssertTrue(wire.contains("hybrid_public_bundle"))
        XCTAssertTrue(wire.contains("ratchet_public_key"))
    }
}
#endif
