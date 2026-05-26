#if os(iOS)
import XCTest
@testable import Echo

final class ContactDiscoveryLogicTests: XCTestCase {
    func testPhoneNormalizerStripsFormatting() {
        XCTAssertEqual(PhoneNormalizer.normalize("+1 (555) 010-00"), "+155501000")
    }

    func testE164AddsUSCountryCode() {
        XCTAssertEqual(PhoneNormalizer.e164(from: "5551234567"), "+15551234567")
    }

    func testMockOPRFBlindFinalizeRoundTrip() async throws {
        let client = MockOPRFClient()
        let blind = try await client.blind(phones: ["+15551234567"])
        XCTAssertEqual(blind.blinded.count, 1)
        let keys = try await client.finalize(sessionID: blind.sessionID, evaluated: ["eval"])
        XCTAssertEqual(keys.count, 1)
        XCTAssertFalse(keys[0].isEmpty)
    }
}
#endif
