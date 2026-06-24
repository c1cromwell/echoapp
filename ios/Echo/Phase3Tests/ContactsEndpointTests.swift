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
