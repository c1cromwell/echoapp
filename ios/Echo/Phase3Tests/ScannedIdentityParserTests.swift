import XCTest
@testable import Echo

final class ScannedIdentityParserTests: XCTestCase {
    func testParseRawDidKey() {
        let id = ScannedIdentityParser.parse("did:key:z6Mkha")
        XCTAssertEqual(id?.did, "did:key:z6Mkha")
    }

    func testParseEchoProfileURL() {
        let raw = "echo://profile?did=did:key:z6Mkpeer&u=alice"
        let id = ScannedIdentityParser.parse(raw)
        XCTAssertEqual(id?.did, "did:key:z6Mkpeer")
        XCTAssertEqual(id?.username, "alice")
    }
}
