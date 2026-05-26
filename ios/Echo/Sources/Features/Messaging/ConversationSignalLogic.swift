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
