#if os(iOS)
import Foundation

/// Shared DM thread + display helpers after add-contact flows (QR, PSI, username).
enum ContactThreadHelper {
    /// Stable `dm:` thread id and `ConversationStore` upsert for messaging + Phase 3 signals.
    @MainActor
    static func upsertDirectThread(peerDID: String, displayName: String) async -> StoredConversation? {
        guard let localDID = await CurrentUserSession.currentDID(), !localDID.isEmpty else {
            return nil
        }
        let threadId = ConversationID.direct(localDID: localDID, peerDID: peerDID)
        let conversation = StoredConversation(
            id: threadId,
            contactName: displayName,
            peerDID: peerDID,
            personaId: PersonaSessionStore.activePersonaId
        )
        ConversationStore.shared.upsert(conversation)
        return conversation
    }

    static func truncatedDID(_ did: String) -> String {
        guard did.count > 22 else { return did }
        return String(did.prefix(14)) + "…" + String(did.suffix(8))
    }
}
#endif
