#if os(iOS)
import Foundation

/// Sealed-sender policy (WO-SX3): default on; optionally limit to trusted contacts.
enum SealedSenderPreferences {
    private static let enabledKey = "echo.sealed_sender.enabled"
    private static let trustedOnlyKey = "echo.sealed_sender.trusted_only"

    static var isEnabled: Bool {
        get {
            if UserDefaults.standard.object(forKey: enabledKey) == nil { return true }
            return UserDefaults.standard.bool(forKey: enabledKey)
        }
        set { UserDefaults.standard.set(newValue, forKey: enabledKey) }
    }

    /// When true, sealed-sender applies only to contacts with trust tier ≥ 2 or favorites.
    static var trustedContactsOnly: Bool {
        get {
            if UserDefaults.standard.object(forKey: trustedOnlyKey) == nil { return true }
            return UserDefaults.standard.bool(forKey: trustedOnlyKey)
        }
        set { UserDefaults.standard.set(newValue, forKey: trustedOnlyKey) }
    }
}

enum SealedSenderPolicy {
    static func shouldUseSealed(peerDID: String, conversationId: String = "") -> Bool {
        guard SealedSenderPreferences.isEnabled, !peerDID.isEmpty else { return false }
        if !SealedSenderPreferences.trustedContactsOnly { return true }
        let tier = ContactTrustIndex.shared.tier(conversationId: conversationId, peerDID: peerDID)
        return tier >= 2 || ContactFavoritesStore.isFavorite(did: peerDID)
    }
}
#endif
