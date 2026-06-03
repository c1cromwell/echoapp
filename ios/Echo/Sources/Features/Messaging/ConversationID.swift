#if os(iOS)
import Foundation

/// Stable thread id so two clients route to the same `conversation_id` on the relay.
enum ConversationID {
    static func direct(localDID: String, peerDID: String) -> String {
        let pair = [localDID, peerDID].sorted()
        return "dm:\(pair[0]):\(pair[1])"
    }
}
#endif
