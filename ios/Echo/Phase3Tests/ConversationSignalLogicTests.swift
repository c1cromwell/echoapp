import XCTest
@testable import Echo

final class WebSocketURLBuilderTests: XCTestCase {

    func testHTTPS_becomesWSSWithToken() throws {
        let api = URL(string: "https://echo.local:8000")!
        let ws = try XCTUnwrap(WebSocketURLBuilder.webSocketURL(apiBaseURL: api, accessToken: "jwt-abc"))
        XCTAssertEqual(ws.scheme, "wss")
        XCTAssertEqual(ws.host, "echo.local")
        XCTAssertEqual(ws.port, 8000)
        XCTAssertEqual(ws.path, "/ws")
        XCTAssertEqual(URLComponents(url: ws, resolvingAgainstBaseURL: false)?.queryItems?.first?.value, "jwt-abc")
    }

    func testHTTP_becomesWS() throws {
        let api = URL(string: "http://127.0.0.1:8000")!
        let ws = try XCTUnwrap(WebSocketURLBuilder.webSocketURL(apiBaseURL: api, accessToken: "t"))
        XCTAssertEqual(ws.scheme, "ws")
        XCTAssertEqual(ws.host, "127.0.0.1")
    }
}

final class ConversationSignalLogicTests: XCTestCase {

    func testReactionToggle_reselectRemoves() {
        XCTAssertNil(ReactionToggleLogic.nextEmoji(currentUserSelection: "👍", tappedEmoji: "👍"))
        XCTAssertEqual(ReactionToggleLogic.nextEmoji(currentUserSelection: nil, tappedEmoji: "❤️"), "❤️")
        XCTAssertEqual(ReactionToggleLogic.nextEmoji(currentUserSelection: "👍", tappedEmoji: "❤️"), "❤️")
    }

    func testDeliveryStatusAdvancement_neverRegresses() {
        XCTAssertEqual(
            DeliveryStatusAdvancement.advanced(current: .read, incoming: .delivered),
            .read
        )
        XCTAssertEqual(
            DeliveryStatusAdvancement.advanced(current: .delivered, incoming: .read),
            .read
        )
        XCTAssertEqual(
            DeliveryStatusAdvancement.advanced(current: nil, incoming: .sent),
            .sent
        )
    }

    func testPrivacy_mergedRequiresBothGlobalAndPersona() {
        var global = EnhancedPrivacySettings()
        global.typingIndicators = true
        global.readReceipts = true
        var persona = PersonaPrivacySettings()
        persona.sendTypingIndicators = false
        persona.sendReadReceipts = true
        let merged = MessagingPrivacyPreferences.merged(global: global, persona: persona)
        XCTAssertFalse(merged.sendTypingIndicators)
        XCTAssertTrue(merged.sendReadReceipts)
    }

    func testReadReceiptPendingIDs_excludesOwnAndAlreadySent() {
        let ids = ReadReceiptLogic.pendingPeerMessageIDs(
            messages: [
                (id: "m1", senderDID: "did:key:peer", isRead: false),
                (id: "m2", senderDID: "did:key:me", isRead: false),
                (id: "m3", senderDID: "did:key:peer", isRead: true),
            ],
            currentUserDID: "did:key:me",
            alreadySent: ["m1"]
        )
        XCTAssertEqual(ids, [])
    }

    func testTypingLogic_respectsPrivacyOff() {
        let off = MessagingPrivacyPreferences(sendTypingIndicators: false, sendReadReceipts: true)
        XCTAssertFalse(TypingIndicatorLogic.shouldEmitStart(isBurstActive: false, hasText: true, privacy: off))
    }

    func testReadReceiptDisplay_capsAtDeliveredWhenPrivacyOff() {
        let off = MessagingPrivacyPreferences(sendTypingIndicators: true, sendReadReceipts: false)
        XCTAssertEqual(ReadReceiptLogic.displayStatus(.read, privacy: off), .delivered)
        XCTAssertEqual(ReadReceiptLogic.displayStatus(.delivered, privacy: off), .delivered)
        let on = MessagingPrivacyPreferences(sendTypingIndicators: true, sendReadReceipts: true)
        XCTAssertEqual(ReadReceiptLogic.displayStatus(.read, privacy: on), .read)
    }
}
