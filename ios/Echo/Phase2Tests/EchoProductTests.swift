import XCTest
@testable import Echo

#if os(iOS)
final class EchoProductTests: XCTestCase {
    func testProductRawValues() {
        XCTAssertEqual(EchoProduct.messaging.rawValue, "messaging")
        XCTAssertEqual(EchoProduct.comply.rawValue, "comply")
        XCTAssertEqual(EchoProduct.passport.rawValue, "passport")
    }

    func testBundleIdentifiers() {
        XCTAssertEqual(EchoProduct.messaging.bundleIdentifier, "com.echo.app")
        XCTAssertEqual(EchoProduct.comply.bundleIdentifier, "com.echo.comply")
        XCTAssertEqual(EchoProduct.passport.bundleIdentifier, "com.echo.passport")
    }
}
#endif
