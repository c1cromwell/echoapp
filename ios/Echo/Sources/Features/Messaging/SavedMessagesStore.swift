import Foundation

/// Telegram-style Saved Messages: a durable self-DID thread for notes and forwards.
enum SavedMessagesStore {
    static let displayName = "Saved Messages"

    /// Same algorithm as `ConversationID.direct` (kept local so macOS SPM tests can resolve it).
    static func conversationId(localDID: String) -> String {
        let pair = [localDID, localDID].sorted()
        return "dm:\(pair[0]):\(pair[1])"
    }

    static func isSavedMessages(_ conversation: StoredConversation, localDID: String) -> Bool {
        guard !localDID.isEmpty else { return false }
        return conversation.peerDID == localDID
            && conversation.id == conversationId(localDID: localDID)
    }

#if os(iOS)
    /// Ensures the Saved Messages conversation exists in the hub list.
    @MainActor
    @discardableResult
    static func ensureConversation() async -> StoredConversation? {
        guard let localDID = await CurrentUserSession.currentDID(), !localDID.isEmpty else {
            return nil
        }
        let id = conversationId(localDID: localDID)
        if var existing = ConversationStore.shared.conversation(id: id) {
            if existing.contactName != displayName {
                existing.contactName = displayName
                ConversationStore.shared.upsert(existing)
            }
            return existing
        }
        let conversation = StoredConversation(
            id: id,
            contactName: displayName,
            peerDID: localDID,
            personaId: PersonaSessionStore.activePersonaId,
            lastMessage: "Notes and forwards",
            timestamp: ""
        )
        ConversationStore.shared.upsert(conversation)
        return conversation
    }
#endif
}
