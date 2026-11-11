import XCTest
@testable import Echo

final class VIPSubscriptionPolicyTests: XCTestCase {
    func testMonthlyProductIsTheMessagingBundle() {
        XCTAssertTrue(VIPSubscriptionPolicy.accepts(
            productID: "com.echo.app.vip.monthly",
            bundleID: "com.echo.app"
        ))
        XCTAssertFalse(VIPSubscriptionPolicy.accepts(
            productID: "com.echo.app.vip.yearly",
            bundleID: "com.echo.app"
        ))
    }
}

final class RelayErrorNoticeTests: XCTestCase {
    func testRateLimitedFrameBecomesARetryMessage() {
        let frame = #"{"type":"error","from":"server","payload":{"code":"RATE_LIMITED","retry_after_seconds":4}}"#
        let notice = RelayErrorNotice.parse(frame)
        XCTAssertEqual(notice?.code, "RATE_LIMITED")
        XCTAssertEqual(notice?.message, "Echo is busy. Try again in 4 seconds.")
    }
}

final class BarSessionTests: XCTestCase {
    func testBarDiscoveryDoesNotAllow() {
        var session = BarSession()
        session.noteDiscovery([
            BarPerson(did: "did:key:sam", name: "Sam", allowed: true)
        ])
        XCTAssertFalse(session.canOpen("did:key:sam"))
        XCTAssertTrue(session.allowed.isEmpty)
        XCTAssertEqual(session.candidates.count, 1)
    }

    func testBarAllowThenRevokeKeepsSendClosed() {
        var session = BarSession()
        let sam = BarPerson(did: "did:key:sam", name: "Sam", allowed: false)
        session.allow(sam)
        XCTAssertTrue(session.canOpen(sam.did))
        session.select(did: sam.did)
        XCTAssertNotNil(session.send("Door is open."))
        session.revoke(did: sam.did)
        XCTAssertNil(session.selected)
        XCTAssertNil(session.send("again"))
        XCTAssertFalse(session.canOpen(sam.did))
    }

    func testBarEmailLookupIsNotLive() {
        XCTAssertEqual(BarSession.interpret("sam@example.com"), .emailNotLive)
        XCTAssertEqual(BarSession.interpret("@sam"), .username("sam"))
        XCTAssertEqual(BarSession.interpret("  "), .empty)
    }

    func testBarThreadSortsDIDs() {
        XCTAssertEqual(
            BarThread.id(a: "did:key:b", b: "did:key:a"),
            "dm:did:key:a:did:key:b"
        )
    }
}
