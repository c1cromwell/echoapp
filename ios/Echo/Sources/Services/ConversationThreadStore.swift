#if os(iOS)
import Foundation

/// Persisted chat row for local thread history (Wave 0.1 / Phase 3 E2E).
struct StoredThreadMessage: Codable, Equatable, Sendable, Identifiable {
    let id: String
    let senderDID: String
    var content: String
    var timestamp: String
    var deliveryStatus: DeliveryStatus?
    var replyToMessageId: String?
    var replyPreview: String?
    var sentAtISO: String?

    func asChatDetailMessage(currentUserDID: String) -> ChatDetailMessage {
        ChatDetailMessage(
            id: id,
            senderDID: senderDID,
            currentUserDID: currentUserDID,
            content: content,
            timestamp: timestamp,
            deliveryStatus: deliveryStatus,
            replyToMessageId: replyToMessageId,
            replyPreview: replyPreview,
            sentAt: StoredThreadMessage.parseSentAt(sentAtISO)
        )
    }

    init(from message: ChatDetailMessage) {
        id = message.id
        senderDID = message.senderDID
        content = message.content
        timestamp = message.timestamp
        deliveryStatus = message.deliveryStatus
        replyToMessageId = message.replyToMessageId
        replyPreview = message.replyPreview
        sentAtISO = StoredThreadMessage.formatSentAt(message.sentAt)
    }

    private static let sentAtFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    static func formatSentAt(_ date: Date?) -> String? {
        guard let date else { return nil }
        return sentAtFormatter.string(from: date)
    }

    static func parseSentAt(_ iso: String?) -> Date? {
        guard let iso, !iso.isEmpty else { return nil }
        return sentAtFormatter.date(from: iso)
            ?? ISO8601DateFormatter().date(from: iso)
    }
}

/// UserDefaults-backed message history per conversation id.
@MainActor
enum ConversationThreadStore {
    private static let keyPrefix = "echo.thread.v1."

    static func load(conversationId: String, currentUserDID: String) -> [ChatDetailMessage] {
        guard !conversationId.isEmpty else { return [] }
        return loadStored(conversationId: conversationId)
            .map { $0.asChatDetailMessage(currentUserDID: currentUserDID) }
    }

    static func replace(conversationId: String, messages: [ChatDetailMessage]) {
        guard !conversationId.isEmpty else { return }
        let stored = messages.map(StoredThreadMessage.init)
        persist(conversationId: conversationId, messages: stored)
    }

    static func appendIfNew(conversationId: String, message: ChatDetailMessage) {
        guard !conversationId.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        guard !stored.contains(where: { $0.id == message.id }) else { return }
        stored.append(StoredThreadMessage(from: message))
        persist(conversationId: conversationId, messages: stored)
    }

    static func updateDeliveryStatus(
        conversationId: String,
        messageId: String,
        status: DeliveryStatus
    ) {
        guard !conversationId.isEmpty else { return }
        var stored = loadStored(conversationId: conversationId)
        guard let idx = stored.firstIndex(where: { $0.id == messageId }) else { return }
        let current = stored[idx].deliveryStatus
        stored[idx].deliveryStatus = DeliveryStatusAdvancement.advanced(current: current, incoming: status)
        persist(conversationId: conversationId, messages: stored)
    }

    private static func loadStored(conversationId: String) -> [StoredThreadMessage] {
        guard let data = UserDefaults.standard.data(forKey: keyPrefix + conversationId),
              let decoded = try? JSONDecoder().decode([StoredThreadMessage].self, from: data) else {
            return []
        }
        return decoded
    }

    private static func persist(conversationId: String, messages: [StoredThreadMessage]) {
        guard let data = try? JSONEncoder().encode(messages) else { return }
        UserDefaults.standard.set(data, forKey: keyPrefix + conversationId)
    }
}
#endif
