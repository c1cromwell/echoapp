#if os(iOS)
import Foundation

/// Client-side disappearing-message enforcement (M1 / WO-59).
@MainActor
enum DisappearingMessageEnforcer {
    /// Computes expiry for a newly sent message when the conversation TTL is active.
    static func expiryDate(sentAt: Date?, ttlSeconds: Int) -> Date? {
        guard ttlSeconds > 0 else { return nil }
        let base = sentAt ?? Date()
        return base.addingTimeInterval(TimeInterval(ttlSeconds))
    }

    /// Removes expired rows from thread storage. Returns deleted message ids.
    @discardableResult
    static func purgeExpired(conversationId: String, ttlSeconds: Int, now: Date = Date()) -> [String] {
        guard ttlSeconds > 0, !conversationId.isEmpty else { return [] }
        var stored = ConversationThreadStore.exportMessages(conversationId: conversationId)
        let before = stored.count
        var removed: [String] = []
        stored.removeAll { row in
            guard let expires = StoredThreadMessage.parseSentAt(row.expiresAtISO) else {
                return false
            }
            if expires <= now {
                removed.append(row.id)
                return true
            }
            return false
        }
        guard stored.count != before else { return [] }
        ConversationThreadStore.replaceStored(conversationId: conversationId, messages: stored)
        return removed
    }

    /// Stamps `expiresAtISO` on outbound rows when TTL is active.
    static func stampExpiry(on message: inout StoredThreadMessage, ttlSeconds: Int) {
        guard ttlSeconds > 0 else { return }
        if let expires = expiryDate(sentAt: StoredThreadMessage.parseSentAt(message.sentAtISO), ttlSeconds: ttlSeconds) {
            message.expiresAtISO = StoredThreadMessage.formatSentAt(expires)
        }
    }
}
#endif
