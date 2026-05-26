#if os(iOS)
import XCTest
@testable import Echo

final class UsernameValidatorTests: XCTestCase {
    func testValidHandles() {
        XCTAssertTrue(UsernameValidator.isValid("alice"))
        XCTAssertTrue(UsernameValidator.isValid("user_42"))
        XCTAssertTrue(UsernameValidator.isValid("ABC123"))
    }

    func testRejectsTooShortOrLong() {
        XCTAssertFalse(UsernameValidator.isValid("ab"))
        XCTAssertFalse(UsernameValidator.isValid(String(repeating: "a", count: 31)))
    }

    func testRejectsSpacesAndSymbols() {
        XCTAssertFalse(UsernameValidator.isValid("bad name"))
        XCTAssertFalse(UsernameValidator.isValid("user-name"))
        XCTAssertFalse(UsernameValidator.isValid("@alice"))
    }

    func testTrimsWhitespace() {
        XCTAssertTrue(UsernameValidator.isValid("  alice  "))
    }
}
#endif
