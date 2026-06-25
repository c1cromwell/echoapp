import XCTest
@testable import Echo

/// Foundations for the BLE-mesh feature: the IDV-verification gate flag and the two-lane
/// entitlement math. These are the pieces the verified/anonymous mesh lanes gate on.
final class MeshEntitlementsTests: XCTestCase {

    func testVerifiedLaneIsFullFeaturedEvenWithoutVIP() {
        let free = MeshEntitlements.limits(lane: .verified, isVIP: false)
        XCTAssertEqual(free.maxHops, MeshEntitlements.protocolMaxHops)
        XCTAssertTrue(free.persistence, "Verified lane keeps message history for free")
    }

    func testAnonymousLaneIsLimitedButFreeAndEphemeral() {
        let free = MeshEntitlements.limits(lane: .anonymous, isVIP: false)
        XCTAssertLessThan(free.maxHops, MeshEntitlements.protocolMaxHops, "Anon reach is limited")
        XCTAssertLessThan(free.maxGroupSize, 32)
        XCTAssertFalse(free.persistence, "Anon lane is ephemeral")
        XCTAssertFalse(free.priorityRelay)
    }

    func testVIPLiftsPerksOnBothLanes() {
        let v = MeshEntitlements.limits(lane: .verified, isVIP: true)
        let a = MeshEntitlements.limits(lane: .anonymous, isVIP: true)
        XCTAssertGreaterThan(v.maxGroupSize, MeshEntitlements.limits(lane: .verified, isVIP: false).maxGroupSize)
        XCTAssertTrue(v.priorityRelay)
        XCTAssertGreaterThan(a.maxHops, MeshEntitlements.limits(lane: .anonymous, isVIP: false).maxHops)
        XCTAssertTrue(a.priorityRelay)
    }

    func testVIPNeverGrantsAnonymousPersistence() {
        // Anonymity guarantee: the anonymous lane stays ephemeral even for paying users.
        XCTAssertFalse(MeshEntitlements.limits(lane: .anonymous, isVIP: true).persistence)
    }

    func testNoLaneExceedsProtocolMaxHops() {
        for lane in MeshLane.allCases {
            for vip in [true, false] {
                XCTAssertLessThanOrEqual(
                    MeshEntitlements.limits(lane: lane, isVIP: vip).maxHops,
                    MeshEntitlements.protocolMaxHops
                )
            }
        }
    }
}

#if os(iOS)
/// The verified mesh lane gates on a real IDV+selfie, recorded as `echo.idvVerified` — distinct
/// from a paid VIP tier, which raises `echo.trustTier` but must NOT unlock verified mesh.
final class IdentityVerificationGateTests: XCTestCase {
    private let idvKey = "echo.idvVerified"
    private let evidenceKey = "echo.evidenceType"
    private let tierKey = "echo.trustTier"

    override func setUp() {
        super.setUp()
        UserDefaults.standard.removeObject(forKey: idvKey)
        UserDefaults.standard.removeObject(forKey: evidenceKey)
        UserDefaults.standard.removeObject(forKey: tierKey)
    }
    override func tearDown() {
        UserDefaults.standard.removeObject(forKey: idvKey)
        UserDefaults.standard.removeObject(forKey: evidenceKey)
        UserDefaults.standard.removeObject(forKey: tierKey)
        super.tearDown()
    }

    func testUnverifiedByDefault() {
        XCTAssertFalse(CurrentUserSession.isIdentityVerified())
        XCTAssertNil(CurrentUserSession.identityEvidenceType())
    }

    func testVIPTierAloneDoesNotCountAsIdentityVerified() {
        // Simulate a VIP subscription (raises tier, never sets the IDV flag).
        UserDefaults.standard.set(3, forKey: tierKey)
        XCTAssertGreaterThanOrEqual(CurrentUserSession.trustTier(), 2)
        XCTAssertFalse(CurrentUserSession.isIdentityVerified(), "VIP tier must not unlock verified mesh")
    }

    func testIDVFlagUnlocksAndReportsEvidence() {
        UserDefaults.standard.set(true, forKey: idvKey)
        UserDefaults.standard.set("standard_idv", forKey: evidenceKey)
        XCTAssertTrue(CurrentUserSession.isIdentityVerified())
        XCTAssertEqual(CurrentUserSession.identityEvidenceType(), "standard_idv")
    }
}
#endif
