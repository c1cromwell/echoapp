#if os(iOS)
import XCTest
@testable import Echo

final class PQHybridPreferencesTests: XCTestCase {
    func testActiveWhenEnabledAndSupported() {
        PQHybridPreferences.usePQBootstrap = true
        XCTAssertTrue(PQHybridBootstrap.isPlatformSupported)
        XCTAssertTrue(PQHybridBootstrap.isActive)
        PQHybridPreferences.usePQBootstrap = false
        XCTAssertFalse(PQHybridBootstrap.isActive)
    }

    func testCachesPeerHybridBundle() {
        let peer = "did:key:zPQPeer"
        let bundle = HybridPublicBundleWire(ec: "ecB64", pq: "pqB64")
        PQHybridBootstrap.cachePeerHybridBundle(peerDID: peer, bundle: bundle)
        XCTAssertEqual(PQHybridBootstrap.cachedPeerHybridBundle(peerDID: peer), bundle)
        PQHybridBootstrap.cachePeerHybridBundle(peerDID: peer, bundle: nil)
        XCTAssertNil(PQHybridBootstrap.cachedPeerHybridBundle(peerDID: peer))
    }

    func testRatchetPreKeyPayloadEncodesHybridFields() throws {
        let wire = try ConversationSignalCodec.encodeRatchetPreKey(
            to: "did:key:zBob",
            conversationId: "conv-1",
            ratchetPublicKeyB64: "cmF0Y2hldA==",
            hybridPublicBundle: HybridPublicBundleWire(ec: "ec", pq: "pq"),
            hybridCiphertext: HybridCiphertextWire(ephemeralEC: "eec", pq: "pqct")
        )
        XCTAssertTrue(wire.contains("hybrid_public_bundle"))
        XCTAssertTrue(wire.contains("hybrid_ciphertext"))
        XCTAssertTrue(wire.contains("ratchet_public_key"))
    }

    func testHybridEncapsulateDecapsulateAgree() throws {
        let generated = try PQHybridCrypto.generateKeyPair()
        let (ct, secretA) = try PQHybridCrypto.encapsulate(remote: generated.bundle)
        let secretB = try PQHybridCrypto.decapsulate(
            ec: generated.ec,
            pq: generated.pq,
            ciphertext: ct
        )
        XCTAssertEqual(secretA, secretB)
        XCTAssertEqual(secretA.count, 32)
    }

    /// Device E2E (two clients): export `ECHO_PQ_ENABLED=true` on backend, confirm
    /// `GET /v3/pq/status` → `available: true`, then enable PQ in Privacy Hub on both devices
    /// and exchange a ratchet pre-key (signal path uses `PQHybridBootstrap`).
    func testPQDeviceE2EPrerequisites() {
        XCTAssertTrue(PQHybridBootstrap.isPlatformSupported)
        let active = PQHybridPreferences.usePQBootstrap && PQHybridBootstrap.isPlatformSupported
        XCTAssertEqual(PQHybridBootstrap.isActive, active)
    }
}
#endif
