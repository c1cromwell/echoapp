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

    // MARK: - Configuration

    private(set) var conversationId: String = ""
    private(set) var peerDID: String = ""
    private(set) var currentUserDID: String = ""

    // MARK: - Dependencies

    private let signalService: ConversationSignalService
    private let textCrypto: TextMessageCrypto
    private var reactionsAPI: (any ReactionsAPIClient)?
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
        onSend: ((String) -> Void)? = nil
    ) {
        self.conversationId = conversationId
        self.peerDID = peerDID
        self.currentUserDID = currentUserDID
        self.peerDisplayName = peerDisplayName
        self.privacy = privacy
        self.reactionsAPI = reactionsAPI
        if let onSend { self.onSend = onSend }

        signalService.setConversationHandler(conversationId: conversationId) { [weak self] event in
            Task { @MainActor in
                self?.handleSignal(event)
            }
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

    // MARK: - Outbound send

    /// Optimistically append an outbound message, route it to the send hook, and
    /// emit a typing-stop. Returns the optimistic message id so callers can later
    /// reconcile delivery status. Single source of truth for the send path —
    /// `ChatView` calls this instead of mutating `messages` directly.
    @discardableResult
    func sendMessage(_ text: String) async -> String? {
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let message = ChatDetailMessage(
            id: UUID().uuidString,
            senderDID: currentUserDID,
            currentUserDID: currentUserDID,
            content: trimmed,
            timestamp: "Now",
            deliveryStatus: .sending
        )
        messages.append(message)
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
            }
        } catch {
            errorMessage = error.localizedDescription
            if let idx = messages.firstIndex(where: { $0.id == message.id }) {
                messages[idx].deliveryStatus = .failed
            }
        }

        await onSendTapped()
        return message.id
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
        var body = e.text
        if let wire = e.wirePayload, wire.encrypted != nil {
            if let decrypted = try? await textCrypto.decryptPayload(wire) {
                body = decrypted
            }
        }
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
        let preview = body == TextMessagePayload.encryptedPlaceholder ? "Encrypted message" : body
        ConversationStore.shared.appendMessagePreview(conversationId: conversationId, preview: preview)
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

    init(
        id: String,
        senderDID: String,
        currentUserDID: String,
        content: String = "",
        timestamp: String = "",
        deliveryStatus: DeliveryStatus? = nil,
        reactions: [ReactionCount] = [],
        isRead: Bool = false
    ) {
        self.id = id
        self.senderDID = senderDID
        self.isFromCurrentUser = senderDID == currentUserDID
        self.content = content
        self.timestamp = timestamp
        self.deliveryStatus = deliveryStatus
        self.reactions = reactions
        self.isRead = isRead
    }
}
