import Foundation

/// First-contact message request gate (Signal Parity Wave S4).
/// Non-contacts must be accepted before their messages appear in the main hub.
/// UserDefaults-backed API is nonisolated so macOS SPM unit tests can call it synchronously.
enum MessageRequestStore {
    private static let pendingKey = "echo.messageRequests.pending"
    private static let acceptedKey = "echo.messageRequests.accepted"

    static func isAccepted(peerDID: String) -> Bool {
        if accepted().contains(peerDID) { return true }
        return isKnownContact(peerDID)
    }

    static func isPending(peerDID: String) -> Bool {
        pending().contains(peerDID)
    }

    static func enqueue(peerDID: String) {
        guard !peerDID.isEmpty, !isAccepted(peerDID: peerDID) else { return }
        var set = pending()
        set.insert(peerDID)
        savePending(set)
    }

    static func accept(peerDID: String) {
        var p = pending()
        p.remove(peerDID)
        savePending(p)
        var a = accepted()
        a.insert(peerDID)
        saveAccepted(a)
    }

    static func decline(peerDID: String) {
        var p = pending()
        p.remove(peerDID)
        savePending(p)
    }

    static func pendingDIDs() -> [String] { Array(pending()).sorted() }

    private static func pending() -> Set<String> {
        Set(UserDefaults.standard.stringArray(forKey: pendingKey) ?? [])
    }

    private static func accepted() -> Set<String> {
        Set(UserDefaults.standard.stringArray(forKey: acceptedKey) ?? [])
    }

    private static func savePending(_ set: Set<String>) {
        UserDefaults.standard.set(Array(set), forKey: pendingKey)
    }

    private static func saveAccepted(_ set: Set<String>) {
        UserDefaults.standard.set(Array(set), forKey: acceptedKey)
    }

    /// Contact lookup hops to the main actor when needed (ConversationStore is `@MainActor`).
    private static func isKnownContact(_ peerDID: String) -> Bool {
        if Thread.isMainThread {
            return MainActor.assumeIsolated { ContactStore.hasContact(did: peerDID) }
        }
        return false
    }
}

/// Minimal contact lookup used by the request gate (avoids hard dependency cycles).
@MainActor
enum ContactStore {
    static func hasContact(did: String) -> Bool {
        // Contacts are also mirrored into ConversationStore as dm: threads with known peers.
        ConversationStore.shared.conversations.contains { conv in
            conv.peerDID == did || conv.id.contains(did)
        }
    }
}

#if os(iOS)
import SwiftUI

struct MessageRequestBanner: View {
    let peerDID: String
    let displayName: String
    var onAccept: () -> Void
    var onDecline: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("Message request")
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(Color.echoInk)
            Text("\(displayName) isn’t in your contacts yet. Accept to keep chatting, or decline to hide this thread.")
                .font(.system(size: 13))
                .foregroundStyle(Color.echoInk55)
            HStack(spacing: 12) {
                Button("Decline", action: onDecline)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(Color.echoInk55)
                Spacer()
                Button("Accept", action: onAccept)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.white)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                    .background(Color.echoSignal)
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            }
        }
        .padding(14)
        .background(Color.echoPaperDim)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .padding(.horizontal, 16)
    }
}
#endif
