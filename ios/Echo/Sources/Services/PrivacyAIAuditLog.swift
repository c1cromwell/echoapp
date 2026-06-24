#if os(iOS)
import CryptoKit
import Foundation

/// Feature invoked by on-device AI (WO-CA1). No message plaintext is stored.
enum PrivacyAIFeature: String, Codable, Sendable, CaseIterable {
    case smartReplies
    case threadSummary
    case translation

    var label: String {
        switch self {
        case .smartReplies: return "Smart replies"
        case .threadSummary: return "Thread summary"
        case .translation: return "Translation"
        }
    }
}

struct PrivacyAIAuditEntry: Codable, Sendable, Identifiable, Equatable {
    let id: UUID
    let feature: PrivacyAIFeature
    /// Truncated SHA-256 of conversation id — not reversible to content.
    let conversationRef: String?
    let timestamp: Date
    let onDevice: Bool

    init(
        id: UUID = UUID(),
        feature: PrivacyAIFeature,
        conversationRef: String?,
        timestamp: Date = Date(),
        onDevice: Bool = true
    ) {
        self.id = id
        self.feature = feature
        self.conversationRef = conversationRef
        self.timestamp = timestamp
        self.onDevice = onDevice
    }
}

/// Local ring buffer of AI invocations for user transparency (T4 — no plaintext).
enum PrivacyAIAuditLog {
    private static let key = "echo.privacy.ai.audit.v1"
    private static let maxEntries = 100

    static func record(_ feature: PrivacyAIFeature, conversationId: String? = nil) {
        var entries = load()
        entries.insert(
            PrivacyAIAuditEntry(
                feature: feature,
                conversationRef: conversationId.map(conversationRefHash)
            ),
            at: 0
        )
        if entries.count > maxEntries {
            entries = Array(entries.prefix(maxEntries))
        }
        persist(entries)
    }

    static func load() -> [PrivacyAIAuditEntry] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let decoded = try? JSONDecoder().decode([PrivacyAIAuditEntry].self, from: data) else {
            return []
        }
        return decoded
    }

    static func clear() {
        UserDefaults.standard.removeObject(forKey: key)
    }

    private static func persist(_ entries: [PrivacyAIAuditEntry]) {
        guard let data = try? JSONEncoder().encode(entries) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }

    private static func conversationRefHash(_ conversationId: String) -> String {
        let digest = SHA256.hash(data: Data(conversationId.utf8))
        return digest.prefix(6).map { String(format: "%02x", $0) }.joined()
    }
}
#endif
