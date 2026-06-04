import XCTest
@testable import Echo

@MainActor
final class ConversationThreadStoreTests: XCTestCase {

    override func tearDown() {
        UserDefaults.standard.removeObject(forKey: "echo.thread.v1.conv-test")
        super.tearDown()
    }

    func testAppendIfNew_dedupesByMessageId() {
        let msg = ChatDetailMessage(
            id: "m1",
            senderDID: "did:key:peer",
            currentUserDID: "did:key:me",
            content: "Hi"
        )
        ConversationThreadStore.appendIfNew(conversationId: "conv-test", message: msg)
        ConversationThreadStore.appendIfNew(conversationId: "conv-test", message: msg)
        let loaded = ConversationThreadStore.load(conversationId: "conv-test", currentUserDID: "did:key:me")
        XCTAssertEqual(loaded.count, 1)
        XCTAssertEqual(loaded.first?.content, "Hi")
    }

    func testReplace_roundTripsDeliveryStatus() {
        var msg = ChatDetailMessage(
            id: "out-1",
            senderDID: "did:key:me",
            currentUserDID: "did:key:me",
            content: "Sent",
            deliveryStatus: .sent
        )
        ConversationThreadStore.replace(conversationId: "conv-test", messages: [msg])
        msg.deliveryStatus = .read
        ConversationThreadStore.replace(conversationId: "conv-test", messages: [msg])
        let loaded = ConversationThreadStore.load(conversationId: "conv-test", currentUserDID: "did:key:me")
        XCTAssertEqual(loaded.first?.deliveryStatus, .read)
    }
}
