import Foundation

/// Merged global + persona privacy for Phase 3 messaging signals.
struct MessagingPrivacyPreferences: Equatable, Sendable {
    var sendTypingIndicators: Bool
    var sendReadReceipts: Bool

    init(sendTypingIndicators: Bool = true, sendReadReceipts: Bool = true) {
        self.sendTypingIndicators = sendTypingIndicators
        self.sendReadReceipts = sendReadReceipts
    }

    static func merged(
        global: EnhancedPrivacySettings = EnhancedPrivacySettings(),
        persona: PersonaPrivacySettings? = nil
    ) -> MessagingPrivacyPreferences {
        var prefs = MessagingPrivacyPreferences(
            sendTypingIndicators: global.typingIndicators,
            sendReadReceipts: global.readReceipts
        )
        if let persona {
            prefs.sendTypingIndicators = prefs.sendTypingIndicators && persona.sendTypingIndicators
            prefs.sendReadReceipts = prefs.sendReadReceipts && persona.sendReadReceipts
        }
        return prefs
    }
}

enum DeliveryStatusAdvancement {
    /// Only advances `DeliveryStatus`; never regresses (WO Phase 3 read receipts).
    static func advanced(current: DeliveryStatus?, incoming: DeliveryStatus) -> DeliveryStatus {
        guard let current else { return incoming }
        return max(current, incoming)
    }
}

/// Decides which outgoing messages a relay `error` frame refers to.
///
/// The server's `RATE_LIMITED` frame names the rejected frame's type, recipient
/// and conversation but not a message id, so the dropped message is taken to be
/// the newest outgoing one still in flight. Each rejected frame produces its own
/// notice, and a message marked failed is no longer in flight, so a burst of
/// notices walks back through the burst newest-first.
enum RelayErrorLogic {
    struct OutgoingMessage: Equatable {
        let id: String
        let isFromCurrentUser: Bool
        let deliveryStatus: DeliveryStatus?
        let sentAt: Date?
    }

    /// Wire types that carry a chat message (see `routeInbound` in `internal/api/ws.go`).
    static let messageFrameTypes: Set<String> = ["text", "sealed_text"]

    /// A message older than this was not dropped by a notice arriving now.
    static let inFlightWindow: TimeInterval = 60

    /// Whether the notice is about a frame sent from this conversation. A frame
    /// that names neither conversation nor recipient could be from any chat.
    static func concerns(_ notice: RelayErrorNotice, conversationId: String, peerDID: String) -> Bool {
        if let rejected = notice.conversationId {
            return rejected == conversationId
        }
        if let to = notice.rejectedTo, !peerDID.isEmpty {
            return to == peerDID
        }
        return true
    }

    /// Ids of the outgoing messages to mark `.failed`; empty when the rejected
    /// frame was not a message (typing, receipts, reactions) or nothing is in flight.
    static func messageIDsToFail(
        notice: RelayErrorNotice,
        conversationId: String,
        peerDID: String,
        messages: [OutgoingMessage],
        now: Date = Date()
    ) -> [String] {
        guard notice.code == RelayErrorNotice.rateLimitedCode,
              concerns(notice, conversationId: conversationId, peerDID: peerDID) else {
            return []
        }
        if let type = notice.rejectedType, !messageFrameTypes.contains(type) {
            return []
        }
        let newestInFlight = messages.last { message in
            guard message.isFromCurrentUser,
                  message.deliveryStatus == .sending || message.deliveryStatus == .sent else {
                return false
            }
            guard let sentAt = message.sentAt else { return true }
            return now.timeIntervalSince(sentAt) <= inFlightWindow
        }
        return newestInFlight.map { [$0.id] } ?? []
    }
}

enum ReactionToggleLogic {
    /// Returns emoji to POST, or `nil` to remove (toggle off same emoji).
    static func nextEmoji(currentUserSelection: String?, tappedEmoji: String) -> String? {
        if currentUserSelection == tappedEmoji {
            return nil
        }
        return tappedEmoji
    }

    static func isSelected(reactors: [String], currentUserDID: String) -> Bool {
        reactors.contains(currentUserDID)
    }
}

enum TypingIndicatorLogic {
    static let debounceInterval: TimeInterval = 1.5
    static let idleStopInterval: TimeInterval = 4.0
    static let peerTypingSafetyTimeout: TimeInterval = 6.0

    static func shouldEmitStart(isBurstActive: Bool, hasText: Bool, privacy: MessagingPrivacyPreferences) -> Bool {
        privacy.sendTypingIndicators && hasText && !isBurstActive
    }

    static func shouldEmitStop(hasText: Bool, privacy: MessagingPrivacyPreferences) -> Bool {
        privacy.sendTypingIndicators && !hasText
    }
}

enum ReadReceiptLogic {
    /// Caps displayed outbound status at `.delivered` when read receipts are disabled.
    static func displayStatus(_ status: DeliveryStatus?, privacy: MessagingPrivacyPreferences) -> DeliveryStatus? {
        guard let status else { return nil }
        guard !privacy.sendReadReceipts, status == .read else { return status }
        return .delivered
    }

    static func pendingPeerMessageIDs(
        messages: [(id: String, senderDID: String, isRead: Bool)],
        currentUserDID: String,
        alreadySent: Set<String>
    ) -> [String] {
        messages
            .filter { $0.senderDID != currentUserDID && !$0.isRead && !alreadySent.contains($0.id) }
            .map(\.id)
    }

    static func shouldSendReceipts(privacy: MessagingPrivacyPreferences) -> Bool {
        privacy.sendReadReceipts
    }
}
