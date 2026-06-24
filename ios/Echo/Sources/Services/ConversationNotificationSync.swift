#if os(iOS)
import Foundation

struct ConversationNotificationPrefsResponse: Codable, Sendable {
    let conversationId: String
    let muted: Bool

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case muted
    }
}

struct ConversationNotificationPrefsRequest: Codable, Sendable {
    let muted: Bool
}

enum ConversationNotificationEndpoint: APIEndpoint {
    case prefs(conversationId: String)

    private static func encode(_ id: String) -> String {
        id.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? id
    }

    var path: String {
        switch self {
        case .prefs(let conversationId):
            return "/v3/conversations/\(Self.encode(conversationId))/notifications"
        }
    }
}

/// Syncs per-conversation mute prefs to the relay for silent push (WO-56).
enum ConversationNotificationSync {
    static func syncMuted(conversationId: String, muted: Bool) {
        guard !conversationId.isEmpty else { return }
        Task {
            guard let client = await DIContainer.shared.resolveAPIClient() else { return }
            _ = try? await client.put(
                endpoint: ConversationNotificationEndpoint.prefs(conversationId: conversationId),
                body: ConversationNotificationPrefsRequest(muted: muted)
            ) as ConversationNotificationPrefsResponse
        }
    }
}
#endif
