#if os(iOS)
import Foundation

/// Stable thread id both devices derive for a 1:1 chat. It is this device's private name for the
/// thread: a sealed message does not carry it (see `sealedWire`).
enum ConversationID {
    static func direct(localDID: String, peerDID: String) -> String {
        let pair = [localDID, peerDID].sorted()
        return "dm:\(pair[0]):\(pair[1])"
    }

    /// What a sealed message carries on the wire where a conversation id would go. `direct` names
    /// both people, so it stays on the device; the recipient works the thread out from the sender
    /// named inside the encryption.
    static let sealedWire = "dm:sealed"
}
#endif
