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

final class RelayErrorLogicTests: XCTestCase {
    private let now = Date(timeIntervalSince1970: 1_800_000_000)

    private func outgoing(_ id: String, _ status: DeliveryStatus?, age: TimeInterval = 1, mine: Bool = true) -> RelayErrorLogic.OutgoingMessage {
        RelayErrorLogic.OutgoingMessage(id: id, isFromCurrentUser: mine, deliveryStatus: status, sentAt: now.addingTimeInterval(-age))
    }

    private func rateLimited(type: String? = "text", to: String? = "did:key:peer", conversation: String? = "conv-1") -> RelayErrorNotice {
        RelayErrorNotice(code: "RATE_LIMITED", message: "busy", rejectedType: type, rejectedTo: to, conversationId: conversation)
    }

    private func ids(_ notice: RelayErrorNotice, _ messages: [RelayErrorLogic.OutgoingMessage]) -> [String] {
        RelayErrorLogic.messageIDsToFail(notice: notice, conversationId: "conv-1", peerDID: "did:key:peer", messages: messages, now: now)
    }

    func testParseReadsTheRejectedFrame() {
        let frame = #"{"type":"error","from":"server","payload":{"code":"RATE_LIMITED","retry_after_seconds":2,"rejected_type":"sealed_text","to":"did:key:peer","conversation_id":"conv-1"}}"#
        let notice = RelayErrorNotice.parse(frame)
        XCTAssertEqual(notice?.rejectedType, "sealed_text")
        XCTAssertEqual(notice?.rejectedTo, "did:key:peer")
        XCTAssertEqual(notice?.conversationId, "conv-1")
    }

    func testNewestInFlightOutgoingMessageIsFailed() {
        let messages = [
            outgoing("delivered", .delivered, age: 5),
            outgoing("older", .sent, age: 3),
            outgoing("peer", nil, mine: false),
            outgoing("newest", .sent, age: 1),
        ]
        XCTAssertEqual(ids(rateLimited(), messages), ["newest"])
    }

    func testEachNoticeWalksBackOneMessage() {
        var messages = [outgoing("a", .sent, age: 3), outgoing("b", .sending, age: 1)]
        XCTAssertEqual(ids(rateLimited(), messages), ["b"])
        messages[1] = outgoing("b", .failed, age: 1)
        XCTAssertEqual(ids(rateLimited(), messages), ["a"])
        messages[0] = outgoing("a", .failed, age: 3)
        XCTAssertEqual(ids(rateLimited(), messages), [])
    }

    func testFrameWithoutDetailsStillFailsTheNewestInFlightMessage() {
        let bare = rateLimited(type: nil, to: nil, conversation: nil)
        XCTAssertEqual(ids(bare, [outgoing("a", .sent)]), ["a"])
    }

    func testDroppedTypingOrReceiptFrameFailsNothing() {
        let messages = [outgoing("a", .sent)]
        XCTAssertEqual(ids(rateLimited(type: "typing"), messages), [])
        XCTAssertEqual(ids(rateLimited(type: "read_receipt"), messages), [])
    }

    func testOtherConversationOrPeerIsIgnored() {
        let messages = [outgoing("a", .sent)]
        XCTAssertEqual(ids(rateLimited(conversation: "conv-2"), messages), [])
        XCTAssertEqual(ids(rateLimited(to: "did:key:other", conversation: nil), messages), [])
    }

    func testDeliveredReadAndOldMessagesAreLeftAlone() {
        XCTAssertEqual(ids(rateLimited(), [outgoing("a", .delivered), outgoing("b", .read), outgoing("c", .anchored)]), [])
        XCTAssertEqual(ids(rateLimited(), [outgoing("old", .sent, age: RelayErrorLogic.inFlightWindow + 1)]), [])
    }

    func testOnlyRateLimitedFailsAMessage() {
        let other = RelayErrorNotice(code: "BAD_FRAME", message: "nope", rejectedType: "text", rejectedTo: nil, conversationId: "conv-1")
        XCTAssertEqual(ids(other, [outgoing("a", .sent)]), [])
    }
}

final class VIPPriceDisplayTests: XCTestCase {
    func testNoPriceIsShownUntilTheAppStoreProvidesOne() {
        XCTAssertNil(VIPPriceDisplay.settingsRowValue(isActive: false, displayPrice: nil))
        XCTAssertNil(VIPPriceDisplay.settingsRowValue(isActive: false, displayPrice: "  "))
        XCTAssertEqual(VIPPriceDisplay.loadState(displayPrice: nil), .unavailable)
        XCTAssertEqual(VIPPriceDisplay.subscribeButtonTitle(state: .unavailable, isProcessing: false), "Subscribe")
        XCTAssertEqual(VIPPriceDisplay.subscribeButtonTitle(state: .loading, isProcessing: false), "Loading price…")
    }

    func testSubscribeIsDisabledWithoutAPrice() {
        XCTAssertFalse(VIPPriceDisplay.canSubscribe(state: .loading, isProcessing: false))
        XCTAssertFalse(VIPPriceDisplay.canSubscribe(state: .unavailable, isProcessing: false))
        XCTAssertTrue(VIPPriceDisplay.canSubscribe(state: .loaded(displayPrice: "€10,99"), isProcessing: false))
        XCTAssertFalse(VIPPriceDisplay.canSubscribe(state: .loaded(displayPrice: "€10,99"), isProcessing: true))
    }

    func testStoreKitPriceIsShownAsGiven() {
        XCTAssertEqual(VIPPriceDisplay.settingsRowValue(isActive: false, displayPrice: "€10,99"), "€10,99/mo")
        XCTAssertEqual(VIPPriceDisplay.settingsRowValue(isActive: true, displayPrice: nil), "Active")
        XCTAssertEqual(
            VIPPriceDisplay.subscribeButtonTitle(state: .loaded(displayPrice: "€10,99"), isProcessing: false),
            "Subscribe for €10,99/mo"
        )
    }
}

final class AppConfigResolutionTests: XCTestCase {
    func testEnvironmentWinsOverInfoPlist() {
        let result = EchoAPIBaseURL.resolve(
            environment: ["API_URL": "http://localhost:8000", "ECHO_API_URL": "https://staging.example.org"],
            infoPlistValue: "https://plist.example.org"
        )
        XCTAssertEqual(result.url.absoluteString, "https://staging.example.org")
        XCTAssertEqual(result.source, .environment)
    }

    func testInfoPlistValueIsUsed() {
        let result = EchoAPIBaseURL.resolve(environment: [:], infoPlistValue: " https://api.example.org ")
        XCTAssertEqual(result.url.absoluteString, "https://api.example.org")
        XCTAssertEqual(result.source, .infoPlist)
    }

    func testMissingOrUnsubstitutedSettingIsReportedAsUnconfigured() {
        for value in [nil, "", "$(API_URL)", "not a url"] as [String?] {
            let result = EchoAPIBaseURL.resolve(environment: [:], infoPlistValue: value)
            XCTAssertEqual(result.source, .unconfigured, "value: \(String(describing: value))")
            XCTAssertEqual(result.url, EchoAPIBaseURL.fallback)
        }
    }

    func testReleasePlaceholderIsRecognised() throws {
        XCTAssertTrue(EchoAPIBaseURL.isPlaceholder(try XCTUnwrap(URL(string: "https://REPLACE-WITH-PRODUCTION-API-HOST.invalid"))))
        XCTAssertTrue(EchoAPIBaseURL.isPlaceholder(EchoAPIBaseURL.fallback))
        XCTAssertFalse(EchoAPIBaseURL.isPlaceholder(try XCTUnwrap(URL(string: "https://api.example.org"))))
    }

    func testLegalLinksReadInfoPlistAndFallBackToPlaceholders() {
        let info: [String: Any] = [
            "EchoTermsURL": "https://echo.example.org/terms",
            "EchoPrivacyURL": "",
            "EchoSupportEmail": "help@echo.example.org",
        ]
        XCTAssertEqual(LegalLinks.url(forInfoKey: "EchoTermsURL", in: info).absoluteString, "https://echo.example.org/terms")
        XCTAssertEqual(LegalLinks.url(forInfoKey: "EchoPrivacyURL", in: info), LegalLinks.placeholderURL)
        XCTAssertEqual(LegalLinks.url(forInfoKey: "EchoTermsURL", in: [:]), LegalLinks.placeholderURL)
        XCTAssertEqual(LegalLinks.url(forInfoKey: "EchoTermsURL", in: ["EchoTermsURL": "javascript:alert(1)"]), LegalLinks.placeholderURL)
        XCTAssertEqual(LegalLinks.supportEmail(in: info), "help@echo.example.org")
        XCTAssertEqual(LegalLinks.supportEmail(in: [:]), LegalLinks.placeholderSupportEmail)
    }

    func testReportAProblemIsAMailtoLink() {
        let url = LegalLinks.mailtoURL(address: "help@echo.example.org", subject: "Echo: report a problem")
        XCTAssertEqual(url.scheme, "mailto")
        XCTAssertTrue(url.absoluteString.hasPrefix("mailto:help@echo.example.org?subject="))
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
