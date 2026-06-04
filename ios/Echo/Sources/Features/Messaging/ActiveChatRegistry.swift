#if os(iOS)
import Foundation

/// Tracks which conversation is on-screen so inbound relay can skip inbox unread bumps.
@MainActor
enum ActiveChatRegistry {
    static var openConversationId: String?
}
#endif
