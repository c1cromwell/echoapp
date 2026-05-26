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
}
