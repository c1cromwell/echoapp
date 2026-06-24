import XCTest
@testable import Echo

#if os(iOS)
final class CallICEEndpointTests: XCTestCase {
    func testICEServersPath_matchesBackend() {
        XCTAssertEqual(CallICEEndpoint.iceServers.path, "/v3/calls/ice-servers")
    }
}

final class WebRTCSessionDescriptionTests: XCTestCase {
    func testOfferAnswerRoundTrip() throws {
        let offer = WebRTCSessionDescription.offer(callType: .voice, iceUfrag: "abc12345")
        let encoded = try offer.encoded()
        let decoded = try WebRTCSessionDescription.decode(from: encoded)
        XCTAssertEqual(decoded.type, "offer")
        let answer = WebRTCSessionDescription.answer(for: decoded)
        XCTAssertEqual(answer.type, "answer")
    }
}

final class OnDeviceAIServiceTests: XCTestCase {
    override func tearDown() {
        PrivacyAIConsentStore.save(.default)
        PrivacyAIAuditLog.clear()
        super.tearDown()
    }

    func testSmartReplies_question() async {
        PrivacyAIConsentStore.save(.init(smartRepliesEnabled: true, summariesEnabled: false, translationEnabled: false))
        let replies = await OnDeviceAIService.shared.smartReplies(from: ["Are you free tomorrow?"])
        XCTAssertFalse(replies.isEmpty)
        XCTAssertEqual(PrivacyAIAuditLog.load().first?.feature, .smartReplies)
    }

    func testSmartReplies_disabledByConsent() async {
        PrivacyAIConsentStore.save(.init(smartRepliesEnabled: false, summariesEnabled: false, translationEnabled: false))
        let replies = await OnDeviceAIService.shared.smartReplies(from: ["Are you free tomorrow?"])
        XCTAssertTrue(replies.isEmpty)
        XCTAssertTrue(PrivacyAIAuditLog.load().isEmpty)
    }

    func testSummarizeThread() async {
        PrivacyAIConsentStore.save(.init(smartRepliesEnabled: false, summariesEnabled: true, translationEnabled: false))
        let summary = await OnDeviceAIService.shared.summarizeThread(messages: [
            "We should meet at the lighthouse.",
            "I'll bring the documents we discussed last week.",
            "Let me know if 3pm works for you.",
        ])
        XCTAssertNotNil(summary)
        XCTAssertEqual(PrivacyAIAuditLog.load().first?.feature, .threadSummary)
    }

    func testAuditLog_neverStoresPlaintext() async {
        PrivacyAIConsentStore.save(.init(smartRepliesEnabled: true, summariesEnabled: false, translationEnabled: false))
        _ = await OnDeviceAIService.shared.smartReplies(
            from: ["Secret message body"],
            conversationId: "conv-secret"
        )
        let encoded = try? JSONEncoder().encode(PrivacyAIAuditLog.load())
        let json = String(data: encoded ?? Data(), encoding: .utf8) ?? ""
        XCTAssertFalse(json.contains("Secret message body"))
    }
}
#endif
