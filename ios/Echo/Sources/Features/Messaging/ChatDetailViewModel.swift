import Foundation
import Observation

/// Chat detail state for Phase 3: typing, read receipts, reactions.
/// Wire into SwiftUI `ChatView` from Xcode (see docs/PHASE3_IOS_UI_SPEC.md).
@Observable
@MainActor
final class ChatDetailViewModel {
    // MARK: - Published UI state

    var peerIsTyping = false
    var peerDisplayName = ""
    var messages: [ChatDetailMessage] = []
    var inputText = ""
    var errorMessage: String?
    var replyingTo: ChatDetailMessage?
    var editingMessageId: String?

    // MARK: - Configuration

    private(set) var conversationId: String = ""
    private(set) var peerDID: String = ""
    private(set) var currentUserDID: String = ""

    // MARK: - Dependencies

    private let signalService: ConversationSignalService
    private let textCrypto: TextMessageCrypto
    private var reactionsAPI: (any ReactionsAPIClient)?
    private var receiptsAPI: (any MessageReceiptsAPIClient)?
    private var privacy: MessagingPrivacyPreferences = .init()

    /// Outbound send hook. Wire to `MessagingService` / `MessageRelayManager` from Xcode;
    /// defaults to a no-op so the optimistic UI works headless and in previews.
    private var onSend: (String) -> Void = { _ in }

    // MARK: - Internal state

    private var typingBurstActive = false
    private var typingDebounceTask: Task<Void, Never>?
    private var typingIdleTask: Task<Void, Never>?
    private var peerTypingClearTask: Task<Void, Never>?
    private var sentReadReceiptIDs = Set<String>()
    private var userReactionByMessage: [String: String] = [:]

    init(signalService: ConversationSignalService, textCrypto: TextMessageCrypto? = nil) {
        self.signalService = signalService
        if let textCrypto {
            self.textCrypto = textCrypto
        } else {
            let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
            self.textCrypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: client))
        }
    }

    func displayedDeliveryStatus(_ status: DeliveryStatus?) -> DeliveryStatus? {
        ReadReceiptLogic.displayStatus(status, privacy: privacy)
    }

    var typingStatusText: String? {
        guard peerIsTyping, privacy.sendTypingIndicators else { return nil }
        let name = peerDisplayName.trimmingCharacters(in: .whitespacesAndNewlines)
        if name.isEmpty { return "typing…" }
        return "\(name) is typing…"
    }

    func configure(
        conversationId: String,
        peerDID: String,
        currentUserDID: String,
        peerDisplayName: String = "",
        privacy: MessagingPrivacyPreferences = .init(),
        reactionsAPI: (any ReactionsAPIClient)? = nil,
        receiptsAPI: (any MessageReceiptsAPIClient)? = nil,
        onSend: ((String) -> Void)? = nil
    ) {
        self.conversationId = conversationId
        self.peerDID = peerDID
        self.currentUserDID = currentUserDID
        self.peerDisplayName = peerDisplayName
        self.privacy = privacy
        self.reactionsAPI = reactionsAPI
        self.receiptsAPI = receiptsAPI
        if let onSend { self.onSend = onSend }

        messages = ConversationThreadStore.load(
            conversationId: conversationId,
            currentUserDID: currentUserDID
        )

        signalService.setConversationHandler(conversationId: conversationId) { [weak self] event in
            Task { @MainActor in
                self?.handleSignal(event)
            }
        }
    }

    /// Bulk GET for messages that may have reactions while offline.
    func reconcileReactionsOnOpen() async {
        guard reactionsAPI != nil else { return }
        let ids = messages.map(\.id)
        for id in ids {
            await reconcileReactions(messageId: id)
        }
    }

    func connect(accessToken: String) async {
        do {
            try await signalService.connect(accessToken: accessToken)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func disconnect() async {
        await emitTypingStopIfNeeded()
        signalService.setConversationHandler(conversationId: conversationId, handler: nil)
        cancelTypingTasks()
        peerTypingClearTask?.cancel()
        peerIsTyping = false
    }

    // MARK: - Typing

    func onInputChanged(_ text: String) {
        inputText = text
        let hasText = !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty

        if TypingIndicatorLogic.shouldEmitStart(
            isBurstActive: typingBurstActive,
            hasText: hasText,
            privacy: privacy
        ) {
            typingBurstActive = true
            scheduleTypingSend(state: .start)
        }

        typingDebounceTask?.cancel()
        typingDebounceTask = Task { [weak self] in
            try? await Task.sleep(nanoseconds: UInt64(TypingIndicatorLogic.debounceInterval * 1_000_000_000))
            guard !Task.isCancelled else { return }
            await self?.scheduleIdleStopIfNeeded(hasText: hasText)
        }
    }

    func onSendTapped() async {
        await emitTypingStopIfNeeded()
        inputText = ""
    }

    func cancelComposerMode() {
        replyingTo = nil
        editingMessageId = nil
    }

    func beginReply(to message: ChatDetailMessage) {
        editingMessageId = nil
        replyingTo = message
    }

    func beginEdit(message: ChatDetailMessage) {
        replyingTo = nil
        editingMessageId = message.id
        inputText = message.content
    }

    func applyEdit(messageId: String, newText: String) async -> Bool {
        let trimmed = newText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty,
              let idx = messages.firstIndex(where: { $0.id == messageId }),
              MessageComposerLogic.canEdit(
                  sentAt: messages[idx].sentAt,
                  isOwnMessage: messages[idx].isFromCurrentUser
              ) else { return false }
        messages[idx].content = trimmed
        ConversationThreadStore.replace(conversationId: conversationId, messages: messages)
        editingMessageId = nil
        inputText = ""
        return true
    }

    // MARK: - Outbound send

    /// Optimistically append an outbound message, route it to the send hook, and
    /// emit a typing-stop. Returns the optimistic message id so callers can later
    /// reconcile delivery status. Single source of truth for the send path —
    /// `ChatView` calls this instead of mutating `messages` directly.
    @discardableResult
    func sendMessage(_ text: String, replyTo: ChatDetailMessage? = nil) async -> String? {
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let replyTarget = replyTo ?? replyingTo
        let replyId = replyTarget?.id
        let replyPreview = replyTarget.map {
            MessageComposerLogic.replyPreview(authorName: peerDisplayNameForReply($0), content: $0.content)
        }

        let message = ChatDetailMessage(
            id: UUID().uuidString,
            senderDID: currentUserDID,
            currentUserDID: currentUserDID,
            content: trimmed,
            timestamp: "Now",
            deliveryStatus: .sending,
            replyToMessageId: replyId,
            replyPreview: replyPreview,
            sentAt: Date()
        )
        messages.append(message)
        ConversationThreadStore.appendIfNew(conversationId: conversationId, message: message)
        onSend(trimmed)

        do {
            let payload: TextMessagePayload
            do {
                payload = try await textCrypto.encryptPayload(
                    plaintext: trimmed,
                    peerDID: peerDID,
                    messageId: message.id
                )
            } catch {
                payload = TextMessagePayload(messageId: message.id, text: trimmed, encrypted: nil)
            }
            try await signalService.sendTextMessage(
                conversationId: conversationId,
                peerDID: peerDID,
                payload: payload
            )
            if let idx = messages.firstIndex(where: { $0.id == message.id }) {
                messages[idx].deliveryStatus = .sent
                syncThreadMessage(messages[idx])
            }
        } catch {
            errorMessage = error.localizedDescription
            if let idx = messages.firstIndex(where: { $0.id == message.id }) {
                messages[idx].deliveryStatus = .failed
                syncThreadMessage(messages[idx])
            }
        }

        cancelComposerMode()
        await onSendTapped()
        return message.id
    }

    private func peerDisplayNameForReply(_ message: ChatDetailMessage) -> String {
        message.isFromCurrentUser ? "You" : peerDisplayName
    }

    private func scheduleIdleStopIfNeeded(hasText: Bool) async {
        guard hasText else {
            await emitTypingStopIfNeeded()
            return
        }
        typingIdleTask?.cancel()
        typingIdleTask = Task { [weak self] in
            try? await Task.sleep(nanoseconds: UInt64(TypingIndicatorLogic.idleStopInterval * 1_000_000_000))
            guard !Task.isCancelled else { return }
            await self?.emitTypingStopIfNeeded()
        }
    }

    private func scheduleTypingSend(state: TypingState) {
        guard privacy.sendTypingIndicators else { return }
        Task {
            try? await signalService.sendTyping(
                conversationId: conversationId,
                peerDID: peerDID,
                state: state
            )
        }
    }

    private func emitTypingStopIfNeeded() async {
        guard typingBurstActive else { return }
        typingBurstActive = false
        typingIdleTask?.cancel()
        if privacy.sendTypingIndicators {
            try? await signalService.sendTyping(
                conversationId: conversationId,
                peerDID: peerDID,
                state: .stop
            )
        }
    }

    private func cancelTypingTasks() {
        typingDebounceTask?.cancel()
        typingIdleTask?.cancel()
    }

    // MARK: - Read receipts

    func onMessageAppeared(messageId: String, senderDID: String) async {
        guard ReadReceiptLogic.shouldSendReceipts(privacy: privacy) else { return }
        guard senderDID != currentUserDID else { return }
        guard !sentReadReceiptIDs.contains(messageId) else { return }

        sentReadReceiptIDs.insert(messageId)
        markMessageReadLocally(messageId)

        do {
            try await signalService.sendReadReceipt(
                conversationId: conversationId,
                peerDID: senderDID,
                messageIds: [messageId]
            )
        } catch {
            errorMessage = error.localizedDescription
        }
        // Durable receipt so read state survives the peer being offline; the server
        // fans out a live read_receipt to the sender (WO-192).
        await persistReadReceipts([messageId])
    }

    func onMessagesVisible(_ visiblePeerMessageIDs: [String]) async {
        guard ReadReceiptLogic.shouldSendReceipts(privacy: privacy) else { return }
        let batch = visiblePeerMessageIDs.filter { !sentReadReceiptIDs.contains($0) }
        guard !batch.isEmpty else { return }
        batch.forEach { sentReadReceiptIDs.insert($0); markMessageReadLocally($0) }

        do {
            try await signalService.sendReadReceipt(
                conversationId: conversationId,
                peerDID: peerDID,
                messageIds: batch
            )
        } catch {
            errorMessage = error.localizedDescription
        }
        await persistReadReceipts(batch)
    }

    /// Best-effort durable read receipts (WO-192). The WS signal is the live nudge;
    /// this REST call is what makes read state survive an offline peer.
    private func persistReadReceipts(_ messageIds: [String]) async {
        guard let receiptsAPI else { return }
        for id in messageIds {
            _ = try? await receiptsAPI.markRead(messageId: id)
        }
    }

    /// On open/reconnect, pull durable delivery state for our own sent messages so
    /// read receipts missed while offline are reconciled (WO-48). Never regresses.
    func reconcileReceiptsOnOpen() async {
        guard let receiptsAPI else { return }
        let ownIDs = messages.filter { $0.isFromCurrentUser }.map(\.id)
        for id in ownIDs {
            guard let status = try? await receiptsAPI.status(messageId: id),
                  let incoming = status.deliveryStatus,
                  let idx = messages.firstIndex(where: { $0.id == id }) else { continue }
            messages[idx].deliveryStatus = DeliveryStatusAdvancement.advanced(
                current: messages[idx].deliveryStatus,
                incoming: incoming
            )
            syncThreadMessage(messages[idx])
        }
    }

    private func markMessageReadLocally(_ messageId: String) {
        guard let idx = messages.firstIndex(where: { $0.id == messageId }) else { return }
        messages[idx].isRead = true
    }

    // MARK: - Reactions

    func toggleReaction(messageId: String, emoji: String) async {
        guard let reactionsAPI else { return }
        let previous = userReactionByMessage[messageId]
        let next = ReactionToggleLogic.nextEmoji(currentUserSelection: previous, tappedEmoji: emoji)

        do {
            let response: MessageReactionsResponse
            if let next {
                response = try await reactionsAPI.react(messageId: messageId, emoji: next)
                userReactionByMessage[messageId] = next
            } else {
                response = try await reactionsAPI.removeReaction(messageId: messageId)
                userReactionByMessage.removeValue(forKey: messageId)
            }
            applyReactionCounts(messageId: messageId, counts: response.reactions)
            let relayEmoji = next ?? ""
            try await signalService.sendReaction(
                conversationId: conversationId,
                peerDID: peerDID,
                messageId: messageId,
                emoji: relayEmoji
            )
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func reconcileReactions(messageId: String) async {
        guard let reactionsAPI else { return }
        do {
            let response = try await reactionsAPI.reactions(messageId: messageId)
            applyReactionCounts(messageId: messageId, counts: response.reactions)
            if let mine = response.reactions.first(where: { $0.reactors.contains(currentUserDID) }) {
                userReactionByMessage[messageId] = mine.emoji
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func applyReactionCounts(messageId: String, counts: [ReactionCount]) {
        guard let idx = messages.firstIndex(where: { $0.id == messageId }) else { return }
        messages[idx].reactions = counts
    }

    // MARK: - Inbound signals

    private func handleSignal(_ event: ConversationSignalEvent) {
        switch event {
        case .typing(let e):
            guard e.conversationId == conversationId else { return }
            switch e.state {
            case .start:
                peerIsTyping = true
                schedulePeerTypingSafetyClear()
            case .stop:
                peerIsTyping = false
                peerTypingClearTask?.cancel()
            }
        case .readReceipt(let e):
            guard e.conversationId == conversationId else { return }
            for id in e.messageIds {
                guard let idx = messages.firstIndex(where: { $0.id == id && $0.isFromCurrentUser }) else { continue }
                messages[idx].deliveryStatus = DeliveryStatusAdvancement.advanced(
                    current: messages[idx].deliveryStatus,
                    incoming: .read
                )
                syncThreadMessage(messages[idx])
            }
        case .reaction(let e):
            guard e.conversationId == conversationId else { return }
            Task { await reconcileReactions(messageId: e.messageId) }
        case .textMessage(let e):
            guard e.conversationId == conversationId else { return }
            guard e.peerDID == peerDID || !e.peerDID.isEmpty else { return }
            guard !messages.contains(where: { $0.id == e.messageId }) else { return }
            Task { await appendInboundText(e) }
        }
    }

    private func appendInboundText(_ e: TextMessageSignalEvent) async {
        let resolved = await InboundTextMessageResolver.resolveBody(for: e)
        let body = resolved.body
        let inbound = ChatDetailMessage(
            id: e.messageId,
            senderDID: e.peerDID,
            currentUserDID: currentUserDID,
            content: body,
            timestamp: "Now",
            deliveryStatus: .delivered
        )
        guard !messages.contains(where: { $0.id == e.messageId }) else { return }
        messages.append(inbound)
        ConversationThreadStore.appendIfNew(conversationId: conversationId, message: inbound)
        ConversationStore.shared.appendMessagePreview(conversationId: conversationId, preview: resolved.preview)
    }

    private func syncThreadMessage(_ message: ChatDetailMessage) {
        ConversationThreadStore.replace(conversationId: conversationId, messages: messages)
    }

    private func schedulePeerTypingSafetyClear() {
        peerTypingClearTask?.cancel()
        peerTypingClearTask = Task { [weak self] in
            try? await Task.sleep(nanoseconds: UInt64(TypingIndicatorLogic.peerTypingSafetyTimeout * 1_000_000_000))
            guard !Task.isCancelled else { return }
            self?.peerIsTyping = false
        }
    }
}

/// Lightweight message row for chat detail (maps to domain `Message` at UI boundary).
struct ChatDetailMessage: Identifiable, Equatable {
    let id: String
    let senderDID: String
    var isFromCurrentUser: Bool
    var content: String
    var timestamp: String
    var deliveryStatus: DeliveryStatus?
    var reactions: [ReactionCount]
    var isRead: Bool
    var replyToMessageId: String?
    var replyPreview: String?
    var sentAt: Date?

    init(
        id: String,
        senderDID: String,
        currentUserDID: String,
        content: String = "",
        timestamp: String = "",
        deliveryStatus: DeliveryStatus? = nil,
        reactions: [ReactionCount] = [],
        isRead: Bool = false,
        replyToMessageId: String? = nil,
        replyPreview: String? = nil,
        sentAt: Date? = nil
    ) {
        self.id = id
        self.senderDID = senderDID
        self.isFromCurrentUser = senderDID == currentUserDID
        self.content = content
        self.timestamp = timestamp
        self.deliveryStatus = deliveryStatus
        self.reactions = reactions
        self.isRead = isRead
        self.replyToMessageId = replyToMessageId
        self.replyPreview = replyPreview
        self.sentAt = sentAt
    }
}
