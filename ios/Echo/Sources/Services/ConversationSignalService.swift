#if os(iOS)
import Foundation

/// Sends and receives conversation signals and text chat over a shared WebSocket.
final class ConversationSignalService: @unchecked Sendable {
    private let transport: ConversationSignalTransport
    private let lock = NSLock()
    private var handlersByConversation: [String: @Sendable (ConversationSignalEvent) -> Void] = [:]
    private var onGroupKeyReceived: (@Sendable (GroupKeySignalEvent) -> Void)?
    /// Updates inbox preview when a text arrives and no chat handler is registered.
    private var onInboundTextMessage: (@Sendable (TextMessageSignalEvent) -> Void)?
    private var onCallSignal: (@Sendable (CallSignalEvent) -> Void)?
    private var isConnected = false

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

    func setInboundTextHandler(_ handler: (@Sendable (TextMessageSignalEvent) -> Void)?) {
        lock.lock()
        onInboundTextMessage = handler
        lock.unlock()
    }

    func setGroupKeyHandler(_ handler: (@Sendable (GroupKeySignalEvent) -> Void)?) {
        lock.lock()
        onGroupKeyReceived = handler
        lock.unlock()
    }

    func setCallSignalHandler(_ handler: (@Sendable (CallSignalEvent) -> Void)?) {
        lock.lock()
        onCallSignal = handler
        lock.unlock()
    }

    func sendRaw(wire: String) async throws {
        try await transport.send(text: wire)
    }

    func setConversationHandler(
        conversationId: String,
        handler: (@Sendable (ConversationSignalEvent) -> Void)?
    ) {
        lock.lock()
        if let handler {
            handlersByConversation[conversationId] = handler
        } else {
            handlersByConversation.removeValue(forKey: conversationId)
        }
        lock.unlock()
    }

    func connect(accessToken: String) async throws {
        lock.lock()
        let already = isConnected
        lock.unlock()
        guard !already else { return }
        try await transport.connect(accessToken: accessToken)
        lock.lock()
        isConnected = true
        lock.unlock()
    }

    func disconnect() async {
        await transport.disconnect()
        lock.lock()
        isConnected = false
        handlersByConversation.removeAll()
        lock.unlock()
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

    func sendPoll(conversationId: String, peerDID: String, payload: PollPayload) async throws {
        let text = try ConversationSignalCodec.encodePoll(
            to: peerDID,
            conversationId: conversationId,
            payload: payload
        )
        try await transport.send(text: text)
    }

    func sendGroupTextMessage(
        conversationId: String,
        payload: TextMessagePayload
    ) async throws {
        let wire = try ConversationSignalCodec.encodeGroupTextMessage(
            conversationId: conversationId,
            payload: payload
        )
        try await transport.send(text: wire)
    }

    func sendTextMessage(
        conversationId: String,
        peerDID: String,
        payload: TextMessagePayload,
        silent: Bool = false
    ) async throws {
        let wire = try ConversationSignalCodec.encodeTextMessage(
            to: peerDID,
            conversationId: conversationId,
            payload: payload,
            silent: silent
        )
        try await transport.send(text: wire)
    }

    func sendSealedTextMessage(
        conversationId: String,
        peerDID: String,
        payload: SealedTextPayload,
        silent: Bool = false
    ) async throws {
        let wire = try ConversationSignalCodec.encodeSealedTextMessage(
            to: peerDID,
            conversationId: conversationId,
            payload: payload,
            silent: silent
        )
        try await transport.send(text: wire)
    }

    func handleIncoming(text: String) {
        #if os(iOS)
        if OverflowManifestHandler.tryHandle(text: text, reprocess: { [weak self] replay in
            self?.handleIncoming(text: replay)
        }) {
            return
        }
        if let callEvent = try? CallSignalCodec.decode(from: text) {
            lock.lock()
            let handler = onCallSignal
            lock.unlock()
            handler?(callEvent)
            return
        }
        #endif

        guard let event = try? ConversationSignalCodec.decodeEvent(from: text) else { return }

        if case .groupKey(let keyEvent) = event {
            lock.lock()
            let handler = onGroupKeyReceived
            lock.unlock()
            handler?(keyEvent)
            return
        }

        if case .textMessage(let textEvent) = event {
            lock.lock()
            let global = onInboundTextMessage
            lock.unlock()
            global?(textEvent)
        }

        let conversationId: String
        switch event {
        case .typing(let e): conversationId = e.conversationId
        case .readReceipt(let e): conversationId = e.conversationId
        case .reaction(let e): conversationId = e.conversationId
        case .textMessage(let e): conversationId = e.conversationId
        case .edit(let e): conversationId = e.conversationId
        case .delete(let e): conversationId = e.conversationId
        case .pin(let e): conversationId = e.conversationId
        case .disappearing(let e): conversationId = e.conversationId
        case .poll(let e): conversationId = e.conversationId
        case .screenshotAlert(let e): conversationId = e.conversationId
        case .groupKey: return
        }

        lock.lock()
        let handler = handlersByConversation[conversationId]
        lock.unlock()
        handler?(event)
    }
}
#endif
