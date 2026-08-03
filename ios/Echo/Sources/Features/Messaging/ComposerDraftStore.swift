import Foundation

/// Persists unsent composer text per conversation across leave/crash (Telegram-style drafts).
enum ComposerDraftStore {
    private static let keyPrefix = "echo.composerDraft.v1."

    static func load(conversationId: String) -> String {
        guard !conversationId.isEmpty else { return "" }
        return UserDefaults.standard.string(forKey: keyPrefix + conversationId) ?? ""
    }

    static func save(conversationId: String, text: String) {
        guard !conversationId.isEmpty else { return }
        let trimmed = text // keep whitespace while composing
        let key = keyPrefix + conversationId
        if trimmed.isEmpty {
            UserDefaults.standard.removeObject(forKey: key)
        } else {
            UserDefaults.standard.set(trimmed, forKey: key)
        }
    }

    static func clear(conversationId: String) {
        guard !conversationId.isEmpty else { return }
        UserDefaults.standard.removeObject(forKey: keyPrefix + conversationId)
    }
}
