import XCTest
@testable import Echo

final class InviteHandleTests: XCTestCase {
    func testNormalizeStripsAtAndWhitespace() {
        XCTAssertEqual(InviteHandle.normalize(" @Alice "), "Alice")
        XCTAssertEqual(InviteHandle.display(" alice "), "@alice")
        XCTAssertTrue(InviteHandle.isUsernameToken("@alice"))
        XCTAssertFalse(InviteHandle.isUsernameToken("ABC123"))
    }

    func testShareURLUsesUsernameQuery() {
        XCTAssertEqual(InviteHandle.shareURL(username: "@Alice")?.absoluteString, "echo://invite?u=Alice")
        XCTAssertNil(InviteHandle.shareURL(username: "   "))
    }

    func testEchoDeepLinkParsesOpaqueInviteCode() {
        let url = URL(string: "echo://invite?code=ABC123")!
        guard case .invite(let code)? = EchoDeepLink.parse(url) else {
            return XCTFail("expected invite deep link")
        }
        XCTAssertEqual(code, "ABC123")
        XCTAssertFalse(InviteHandle.isUsernameToken(code))
    }

    func testEchoDeepLinkParsesUsernameInvite() {
        let url = URL(string: "echo://invite?u=alice")!
        guard case .invite(let token)? = EchoDeepLink.parse(url) else {
            return XCTFail("expected username invite")
        }
        XCTAssertEqual(token, "@alice")
        XCTAssertTrue(InviteHandle.isUsernameToken(token))
        XCTAssertEqual(InviteHandle.normalize(token), "alice")
    }

    func testEchoDeepLinkUsernameBeatsOpaqueCode() {
        let url = URL(string: "echo://invite?u=alice&code=ABC123")!
        guard case .invite(let token)? = EchoDeepLink.parse(url) else {
            return XCTFail("expected username invite")
        }
        XCTAssertEqual(token, "@alice")
    }

    func testEchoDeepLinkParsesAtPathInvite() {
        let url = URL(string: "echo://invite/@sam")!
        guard case .invite(let token)? = EchoDeepLink.parse(url) else {
            return XCTFail("expected username invite from path")
        }
        XCTAssertEqual(token, "@sam")
    }
}
