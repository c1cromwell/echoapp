import XCTest
@testable import Echo

final class ContactsEndpointTests: XCTestCase {
    func testListContactsPath_matchesBackend() {
        XCTAssertEqual(ContactsEndpoint.list.path, "/v3/contacts/list")
    }
}

#if os(iOS)
final class ContactThreadHelperTests: XCTestCase {
    func testTruncatedDID_shortensLongKey() {
        let did = "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbZEBfALj"
        let short = ContactThreadHelper.truncatedDID(did)
        XCTAssertTrue(short.contains("…"))
        XCTAssertLessThan(short.count, did.count)
    }
}
#endif
