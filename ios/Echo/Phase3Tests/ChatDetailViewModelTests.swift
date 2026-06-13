import XCTest
@testable import Echo

@MainActor
final class ChatDetailViewModelTests: XCTestCase {

    private var transport: MockConversationSignalTransport!
    private var service: ConversationSignalService!
    private var reactions: MockReactionsAPIClient!
    private var vm: ChatDetailViewModel!

    override func setUp() async throws {
        try await super.setUp()
        transport = MockConversationSignalTransport()
        service = ConversationSignalService(transport: transport)
        reactions = MockReactionsAPIClient()
        vm = ChatDetailViewModel(signalService: service)
        vm.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            privacy: MessagingPrivacyPreferences(sendTypingIndicators: true, sendReadReceipts: true),
            reactionsAPI: reactions
        )
        vm.messages = [
            ChatDetailMessage(id: "out-1", senderDID: "did:key:me", currentUserDID: "did:key:me", deliveryStatus: .delivered),
            ChatDetailMessage(id: "in-1", senderDID: "did:key:peer", currentUserDID: "did:key:me"),
        ]
    }

    func testInboundTyping_setsPeerIsTyping() async {
        transport.simulateIncoming("""
        {"type":"typing","from":"did:key:peer","to":"did:key:me","conversation_id":"conv-1","payload":{"conversation_id":"conv-1","state":"start"}}
        """)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertTrue(vm.peerIsTyping)
    }

    func testInboundReadReceipt_advancesDeliveryStatus() async {
        transport.simulateIncoming("""
        {"type":"read_receipt","from":"did:key:peer","to":"did:key:me","conversation_id":"conv-1","payload":{"conversation_id":"conv-1","message_ids":["out-1"],"read_at":"2026-05-25T00:00:00Z"}}
        """)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(vm.messages.first(where: { $0.id == "out-1" })?.deliveryStatus, .read)
    }

    func testToggleReaction_callsRESTAndWS() async {
        await vm.toggleReaction(messageId: "in-1", emoji: "👍")
        XCTAssertEqual(reactions.reactCalls.count, 1)
        XCTAssertEqual(reactions.reactCalls.first?.emoji, "👍")
        XCTAssertTrue(transport.sentTexts.contains(where: { $0.contains("\"type\":\"reaction\"") }))
    }

    func testOnInputChanged_sendsTypingStart() async {
        vm.onInputChanged("hello")
        try? await Task.sleep(nanoseconds: 100_000_000)
        XCTAssertTrue(transport.sentTexts.contains(where: { $0.contains("\"state\":\"start\"") }))
    }

    func testSendMessage_appendsOptimisticMessageAndRoutesToHook() async {
        var sent: [String] = []
        vm.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            privacy: MessagingPrivacyPreferences(sendTypingIndicators: true, sendReadReceipts: true),
            reactionsAPI: reactions,
            onSend: { sent.append($0) }
        )
        let before = vm.messages.count
        let id = await vm.sendMessage("hi there")
        XCTAssertNotNil(id)
        XCTAssertEqual(vm.messages.count, before + 1)
        let appended = vm.messages.last
        XCTAssertEqual(appended?.content, "hi there")
        XCTAssertTrue(appended?.isFromCurrentUser == true)
        XCTAssertEqual(appended?.deliveryStatus, .sent)
        XCTAssertEqual(sent, ["hi there"])
        XCTAssertTrue(transport.sentTexts.contains(where: { $0.contains("\"type\":\"text\"") }))
    }

    func testConfigure_loadsPersistedThread() async {
        let stored = ChatDetailMessage(
            id: "stored-1",
            senderDID: "did:key:peer",
            currentUserDID: "did:key:me",
            content: "Earlier"
        )
        ConversationThreadStore.appendIfNew(conversationId: "conv-1", message: stored)
        defer { UserDefaults.standard.removeObject(forKey: "echo.thread.v1.conv-1") }

        let fresh = ChatDetailViewModel(signalService: service)
        fresh.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            reactionsAPI: reactions
        )
        XCTAssertEqual(fresh.messages.count, 1)
        XCTAssertEqual(fresh.messages.first?.content, "Earlier")
    }

    func testInboundTextMessage_appendsPeerMessage() async {
        transport.simulateIncoming("""
        {"type":"text","from":"did:key:peer","to":"did:key:me","conversation_id":"conv-1","payload":{"text":"Hey","message_id":"peer-msg-1"}}
        """)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertTrue(vm.messages.contains(where: { $0.id == "peer-msg-1" && $0.content == "Hey" }))
    }

    func testSendMessage_ignoresWhitespaceOnly() async {
        let before = vm.messages.count
        let id = await vm.sendMessage("   \n  ")
        XCTAssertNil(id)
        XCTAssertEqual(vm.messages.count, before)
    }

    func testDisconnect_emitsTypingStop() async {
        vm.onInputChanged("typing…")
        try? await Task.sleep(nanoseconds: 100_000_000)
        transport.sentTexts.removeAll()
        await vm.disconnect()
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertTrue(transport.sentTexts.contains(where: { $0.contains("\"state\":\"stop\"") }))
    }

    func testReadReceiptDisabled_doesNotSend() async {
        vm.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            privacy: MessagingPrivacyPreferences(sendTypingIndicators: true, sendReadReceipts: false),
            reactionsAPI: reactions
        )
        await vm.onMessageAppeared(messageId: "in-1", senderDID: "did:key:peer")
        XCTAssertFalse(transport.sentTexts.contains(where: { $0.contains("read_receipt") }))
    }

    // MARK: - Durable receipts (WO-192/48)

    func testOnMessageAppeared_persistsDurableReadReceipt() async {
        let receipts = MockMessageReceiptsAPIClient()
        vm.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            privacy: MessagingPrivacyPreferences(sendTypingIndicators: true, sendReadReceipts: true),
            reactionsAPI: reactions,
            receiptsAPI: receipts
        )
        await vm.onMessageAppeared(messageId: "in-1", senderDID: "did:key:peer")
        XCTAssertEqual(receipts.markReadCalls, ["in-1"], "read should be durably persisted via REST")
    }

    func testReconcileReceiptsOnOpen_advancesOwnMessageToRead() async {
        let receipts = MockMessageReceiptsAPIClient()
        receipts.statuses["out-1"] = MessageStatusResponse(
            messageId: "out-1", conversationId: "conv-1", status: "read",
            deliveredAt: "2026-05-25T00:00:00Z", readAt: "2026-05-25T00:00:01Z"
        )
        vm.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            reactionsAPI: reactions,
            receiptsAPI: receipts
        )
        vm.messages = [
            ChatDetailMessage(id: "out-1", senderDID: "did:key:me", currentUserDID: "did:key:me", deliveryStatus: .delivered),
            ChatDetailMessage(id: "in-1", senderDID: "did:key:peer", currentUserDID: "did:key:me"),
        ]
        await vm.reconcileReceiptsOnOpen()
        XCTAssertEqual(vm.messages.first(where: { $0.id == "out-1" })?.deliveryStatus, .read)
    }
}
