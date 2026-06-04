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
}
