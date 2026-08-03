import Foundation

/// View-once media flag (Signal Parity Wave S4).
/// Wire as part of `MediaAttachmentRef` / message payload; clients burn local
/// plaintext after first successful open.
struct ViewOnceMarker: Codable, Sendable, Hashable {
    var isViewOnce: Bool
    var viewedAt: Date?
    var burned: Bool

    static let enabled = ViewOnceMarker(isViewOnce: true, viewedAt: nil, burned: false)

    mutating func markViewed() {
        guard isViewOnce, !burned else { return }
        viewedAt = Date()
        burned = true
    }
}

enum ViewOnceStore {
    private static func key(_ messageId: String) -> String { "echo.viewonce.\(messageId)" }

    static func isBurned(messageId: String) -> Bool {
        UserDefaults.standard.bool(forKey: key(messageId))
    }

    static func markBurned(messageId: String) {
        UserDefaults.standard.set(true, forKey: key(messageId))
    }
}

extension TextMessagePayload {
    /// Placeholder body for view-once media after burn.
    static var viewOnceBurnedPlaceholder: String { "••• View-once media" }
}
