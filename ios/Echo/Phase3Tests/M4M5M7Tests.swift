import XCTest
@testable import Echo

#if os(iOS)
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
    func testSmartReplies_question() async {
        PrivacyAIConsentStore.save(.init(smartRepliesEnabled: true, summariesEnabled: false, translationEnabled: false))
        let replies = await OnDeviceAIService.shared.smartReplies(from: ["Are you free tomorrow?"])
        XCTAssertFalse(replies.isEmpty)
    }

    func testSummarizeThread() async {
        PrivacyAIConsentStore.save(.init(smartRepliesEnabled: false, summariesEnabled: true, translationEnabled: false))
        let summary = await OnDeviceAIService.shared.summarizeThread(messages: [
            "We should meet at the lighthouse.",
            "I'll bring the documents we discussed last week.",
            "Let me know if 3pm works for you.",
        ])
        XCTAssertNotNil(summary)
    }
}
#endif
