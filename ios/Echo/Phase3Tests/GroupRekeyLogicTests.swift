import XCTest
@testable import Echo

#if os(iOS)
final class GroupRekeyLogicTests: XCTestCase {
    func testGroupMemberWireAdminRoles() {
        XCTAssertTrue(GroupMemberWire(memberId: "a", groupId: "g", displayName: nil, role: "owner").isAdmin)
        XCTAssertTrue(GroupMemberWire(memberId: "a", groupId: "g", displayName: nil, role: "admin").isAdmin)
        XCTAssertFalse(GroupMemberWire(memberId: "a", groupId: "g", displayName: nil, role: "member").isAdmin)
    }
}
#endif
