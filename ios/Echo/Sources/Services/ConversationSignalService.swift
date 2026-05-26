import Foundation

/// Sends and receives Phase 3 ephemeral conversation signals over WebSocket.
final class ConversationSignalService: @unchecked Sendable {
    private let transport: ConversationSignalTransport
    private let lock = NSLock()
    private var eventHandler: (@Sendable (ConversationSignalEvent) -> Void)?

    init(transport: ConversationSignalTransport) {
        self.transport = transport
        transport.onTextMessage = { [weak self] text in
            self?.handleIncoming(text: text)
        }
    }

    #if os(iOS)
    convenience init(apiBaseURL: URL) {
        self.init(transport: WebSocketConversationSignalTransport(apiBaseURL: apiBaseURL))
    }
    #endif

    func setEventHandler(_ handler: (@Sendable (ConversationSignalEvent) -> Void)?) {
        lock.lock()
        eventHandler = handler
        lock.unlock()
    }

    func connect(accessToken: String) async throws {
        try await transport.connect(accessToken: accessToken)
    }

    func disconnect() async {
        await transport.disconnect()
    }

    func sendTyping(conversationId: String, peerDID: String, state: TypingState) async throws {
        let text = try ConversationSignalCodec.encodeTyping(
            to: peerDID,
            conversationId: conversationId,
            state: state
        )
        try await transport.send(text: text)
    }

    func sendReadReceipt(conversationId: String, peerDID: String, messageIds: [String]) async throws {
        guard !messageIds.isEmpty else { return }
        let text = try ConversationSignalCodec.encodeReadReceipt(
            to: peerDID,
            conversationId: conversationId,
            messageIds: messageIds,
            readAt: ConversationSignalCodec.isoTimestamp()
        )
        try await transport.send(text: text)
    }

    func sendReaction(conversationId: String, peerDID: String, messageId: String, emoji: String) async throws {
        let text = try ConversationSignalCodec.encodeReaction(
            to: peerDID,
            conversationId: conversationId,
            messageId: messageId,
            emoji: emoji
        )
        try await transport.send(text: text)
    }

    func handleIncoming(text: String) {
        guard let event = try? ConversationSignalCodec.decodeEvent(from: text) else { return }
        lock.lock()
        let handler = eventHandler
        lock.unlock()
        handler?(event)
    }
}
