#if os(iOS)
import Foundation

/// Rotating opaque queue aliases per conversation (WO-SX3 metadata minimization).
enum ConversationQueueAliasStore {
    private static func key(_ conversationId: String) -> String {
        "echo.queue.alias.\(conversationId)"
    }

    static func alias(for conversationId: String) -> String {
        let k = key(conversationId)
        if let existing = UserDefaults.standard.string(forKey: k), !existing.isEmpty {
            return existing
        }
        let created = UUID().uuidString.lowercased()
        UserDefaults.standard.set(created, forKey: k)
        return created
    }

    static func rotate(conversationId: String) -> String {
        let created = UUID().uuidString.lowercased()
        UserDefaults.standard.set(created, forKey: key(conversationId))
        return created
    }
}
#endif
