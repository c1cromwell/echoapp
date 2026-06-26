#if os(iOS)
import XCTest
@testable import Echo

final class SealedSenderPolicyTests: XCTestCase {
    func testTrustedOnlyRequiresTierOrFavorite() {
        SealedSenderPreferences.isEnabled = true
        SealedSenderPreferences.trustedContactsOnly = true
        let peer = "did:key:zUnknown"
        XCTAssertFalse(SealedSenderPolicy.shouldUseSealed(peerDID: peer))
        ContactFavoritesStore.toggle(did: peer)
        XCTAssertTrue(SealedSenderPolicy.shouldUseSealed(peerDID: peer))
        ContactFavoritesStore.toggle(did: peer)
    }

    func testDisabledGlobally() {
        SealedSenderPreferences.isEnabled = false
        XCTAssertFalse(SealedSenderPolicy.shouldUseSealed(peerDID: "did:key:zAny"))
        SealedSenderPreferences.isEnabled = true
    }
}
#endif
