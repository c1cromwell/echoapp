import XCTest
@testable import Echo

final class MessageComposerLogicTests: XCTestCase {

    func testCanEdit_withinWindow() {
        let sent = Date().addingTimeInterval(-60)
        XCTAssertTrue(MessageComposerLogic.canEdit(sentAt: sent, isOwnMessage: true))
    }

    func testCanEdit_outsideWindow() {
        let sent = Date().addingTimeInterval(-(16 * 60))
        XCTAssertFalse(MessageComposerLogic.canEdit(sentAt: sent, isOwnMessage: true))
    }

    func testCanEdit_notOwn() {
        XCTAssertFalse(MessageComposerLogic.canEdit(sentAt: Date(), isOwnMessage: false))
    }

    func testForwardBody_prefixesArrow() {
        XCTAssertEqual(MessageComposerLogic.forwardBody("Hi"), "↪ Hi")
    }

    func testReplyPreview_includesAuthor() {
        let p = MessageComposerLogic.replyPreview(authorName: "Alex", content: "Hello")
        XCTAssertTrue(p.contains("Alex"))
        XCTAssertTrue(p.contains("Hello"))
    }

    func testChatMessageEnvelopeRoundTrip() throws {
        let env = ChatMessageEnvelope(body: "Hi", replyToMessageId: "m1", replyPreview: "Alex: Hi")
        let json = try env.serialized()
        let parsed = ChatMessageEnvelope.parseDecrypted(json)
        XCTAssertEqual(parsed.body, "Hi")
        XCTAssertEqual(parsed.envelope?.replyToMessageId, "m1")
        XCTAssertEqual(parsed.envelope?.replyPreview, "Alex: Hi")
    }

    func testChatMessageEnvelopeLegacyPlaintext() {
        let parsed = ChatMessageEnvelope.parseDecrypted("plain text")
        XCTAssertEqual(parsed.body, "plain text")
        XCTAssertNil(parsed.envelope)
    }
}
