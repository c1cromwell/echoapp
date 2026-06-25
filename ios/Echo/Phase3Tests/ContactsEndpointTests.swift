#if os(iOS)
import XCTest
@testable import Echo

final class ContactsEndpointTests: XCTestCase {
    func testListContactsPath_matchesBackend() {
        XCTAssertEqual(ContactsEndpoint.list.path, "/v3/contacts/list")
    }

    func testRelationshipPath_includesPeerDID() {
        let path = ContactsEndpoint.relationship(peerDID: "did:key:zPeer").path
        XCTAssertTrue(path.hasPrefix("/v3/contacts/relationship?peer_did="))
        XCTAssertTrue(path.contains("did:key"))
    }

    func testBlockedContactsPaths() {
        XCTAssertEqual(ContactsEndpoint.blocked.path, "/v3/contacts/blocked")
        XCTAssertEqual(ContactsEndpoint.unblock.path, "/v3/contacts/unblock")
        XCTAssertEqual(ContactsEndpoint.contactPrivacy.path, "/v3/contacts/privacy")
    }

    func testProfilePaths() {
        XCTAssertEqual(ProfileEndpoint.own.path, "/v3/profile")
        XCTAssertEqual(ProfileEndpoint.privacy.path, "/v3/profile/privacy")
        XCTAssertTrue(ProfileEndpoint.view(did: "did:key:zPeer").path.contains("/v3/profile/view?did="))
    }

    func testLogoutPath_matchesBackendRevoke() {
        XCTAssertEqual(AuthEndpoint.logout.path, "/v3/auth/revoke")
    }
}

final class ContactThreadHelperTests: XCTestCase {
    func testTruncatedDID_shortensLongKey() {
        let did = "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbZEBfALj"
        let short = ContactThreadHelper.truncatedDID(did)
        XCTAssertTrue(short.contains("…"))
        XCTAssertLessThan(short.count, did.count)
    }
}
#endif // os(iOS)
