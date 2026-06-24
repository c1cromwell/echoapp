#if os(iOS)
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
        Self.clearPersistedThreads()
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

    override func tearDown() async throws {
        Self.clearPersistedThreads()
        try await super.tearDown()
    }

    /// `ConversationThreadStore` is UserDefaults-backed; clear its keys between tests so
    /// persisted threads from one test don't leak into another (these tests share a process).
    private static func clearPersistedThreads() {
        let defaults = UserDefaults.standard
        for key in defaults.dictionaryRepresentation().keys where key.hasPrefix("echo.thread.v1.") {
            defaults.removeObject(forKey: key)
        }
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
        XCTAssertEqual(sent, ["hi there"])
        // Fail-closed (2026-06 security fix): with no resolvable peer messaging key in this
        // unit context, encryption fails, so the message is marked .failed and is never
        // transmitted as plaintext. The optimistic UI hook still fires.
        XCTAssertEqual(appended?.deliveryStatus, .failed)
        XCTAssertFalse(transport.sentTexts.contains(where: { $0.contains("\"type\":\"text\"") }))
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

    // MARK: - Message ops (WO-25/84/59)

    private func configureWithOps(_ ops: MockMessageOpsAPIClient) {
        vm.configure(
            conversationId: "conv-1",
            peerDID: "did:key:peer",
            currentUserDID: "did:key:me",
            reactionsAPI: reactions,
            opsAPI: ops
        )
        vm.messages = [
            ChatDetailMessage(id: "out-1", senderDID: "did:key:me", currentUserDID: "did:key:me", content: "orig", deliveryStatus: .delivered, sentAt: Date()),
            ChatDetailMessage(id: "in-1", senderDID: "did:key:peer", currentUserDID: "did:key:me", content: "hi"),
        ]
    }

    func testApplyEdit_updatesLocallyAndCallsBackend() async {
        let ops = MockMessageOpsAPIClient()
        configureWithOps(ops)
        let ok = await vm.applyEdit(messageId: "out-1", newText: "edited!")
        XCTAssertTrue(ok)
        XCTAssertEqual(vm.messages.first(where: { $0.id == "out-1" })?.content, "edited!")
        // Fail-closed (2026-06 security fix): the local edit stands, but with no resolvable
        // peer messaging key the new ciphertext can't be produced, so the edit is never fanned
        // out to the backend (no plaintext edit transmission).
        XCTAssertTrue(ops.editCalls.isEmpty)
    }

    func testDeleteMessage_removesAndCallsBackend() async {
        let ops = MockMessageOpsAPIClient()
        configureWithOps(ops)
        let ok = await vm.deleteMessage("in-1")
        XCTAssertTrue(ok)
        XCTAssertFalse(vm.messages.contains { $0.id == "in-1" })
        XCTAssertEqual(ops.deleteCalls, ["in-1"])
    }

    func testTogglePin_pinsThenUnpins() async {
        let ops = MockMessageOpsAPIClient()
        configureWithOps(ops)
        let pinned = await vm.togglePin(messageId: "in-1")
        XCTAssertEqual(pinned, true)
        XCTAssertTrue(vm.pinnedMessageIDs.contains("in-1"))
        XCTAssertEqual(ops.pinCalls, ["in-1"])

        let unpinned = await vm.togglePin(messageId: "in-1")
        XCTAssertEqual(unpinned, false)
        XCTAssertFalse(vm.pinnedMessageIDs.contains("in-1"))
        XCTAssertEqual(ops.unpinCalls, ["in-1"])
    }

    func testTogglePin_rollsBackOnFailure() async {
        let ops = MockMessageOpsAPIClient()
        ops.pinShouldFail = true
        configureWithOps(ops)
        let result = await vm.togglePin(messageId: "in-1")
        XCTAssertNil(result)
        XCTAssertFalse(vm.pinnedMessageIDs.contains("in-1"), "optimistic pin must roll back on error")
    }

    func testInboundDelete_removesMessage() async {
        let ops = MockMessageOpsAPIClient()
        configureWithOps(ops)
        transport.simulateIncoming("""
        {"type":"delete","from":"did:key:peer","to":"did:key:me","conversation_id":"conv-1","payload":{"conversation_id":"conv-1","message_id":"in-1"}}
        """)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertFalse(vm.messages.contains { $0.id == "in-1" })
    }

    func testInboundPin_updatesPinnedSet() async {
        let ops = MockMessageOpsAPIClient()
        configureWithOps(ops)
        transport.simulateIncoming("""
        {"type":"pin","from":"did:key:peer","to":"did:key:me","conversation_id":"conv-1","payload":{"conversation_id":"conv-1","message_id":"in-1","pinned":true}}
        """)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertTrue(vm.pinnedMessageIDs.contains("in-1"))
    }

    func testInboundDisappearing_setsTTL() async {
        let ops = MockMessageOpsAPIClient()
        configureWithOps(ops)
        transport.simulateIncoming("""
        {"type":"disappearing_config","from":"did:key:peer","to":"did:key:me","conversation_id":"conv-1","payload":{"conversation_id":"conv-1","ttl_seconds":3600}}
        """)
        try? await Task.sleep(nanoseconds: 50_000_000)
        XCTAssertEqual(vm.disappearingTTLSeconds, 3600)
    }
}
#endif // os(iOS)
