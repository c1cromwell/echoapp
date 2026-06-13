#if os(iOS)
import Foundation

/// Canonical history export for WO-CA3 sync and WO-64 backup (M3b/M3c shared format).
struct HistorySyncBundle: Codable, Sendable, Equatable {
    /// Bundle schema version (increment on breaking changes).
    static let currentVersion = 1

    let version: Int
    let bundleId: String
    let exportedAt: String
    let conversations: [StoredConversation]
    let threads: [String: [StoredThreadMessage]]

    init(
        version: Int = HistorySyncBundle.currentVersion,
        bundleId: String = UUID().uuidString,
        exportedAt: String = HistorySyncBundle.isoNow(),
        conversations: [StoredConversation],
        threads: [String: [StoredThreadMessage]]
    ) {
        self.version = version
        self.bundleId = bundleId
        self.exportedAt = exportedAt
        self.conversations = conversations
        self.threads = threads
    }

    func encoded() throws -> Data {
        try JSONEncoder().encode(self)
    }

    static func decode(from data: Data) throws -> HistorySyncBundle {
        try JSONDecoder().decode(HistorySyncBundle.self, from: data)
    }

    static func isoNow() -> String {
        ISO8601DateFormatter().string(from: Date())
    }
}

// MARK: - Build + merge

@MainActor
enum HistorySyncBundleBuilder {
    /// Snapshot local conversations + thread history for sync/backup.
    static func build(from store: ConversationStore = .shared) -> HistorySyncBundle {
        let conversations = store.conversations
        var threads: [String: [StoredThreadMessage]] = [:]

        let threadIds = Set(conversations.map(\.id)).union(Set(ConversationThreadStore.allStoredConversationIds()))
        for conversationId in threadIds {
            let messages = ConversationThreadStore.exportMessages(conversationId: conversationId)
            if !messages.isEmpty {
                threads[conversationId] = messages
            }
        }

        return HistorySyncBundle(conversations: conversations, threads: threads)
    }
}

@MainActor
enum HistorySyncBundleMerger {
    /// Idempotent apply: upsert conversations, merge messages by id.
    static func apply(_ bundle: HistorySyncBundle, to store: ConversationStore = .shared) {
        for conversation in bundle.conversations {
            store.upsert(conversation)
        }
        for (conversationId, messages) in bundle.threads {
            ConversationThreadStore.mergeMessages(conversationId: conversationId, incoming: messages)
        }
    }
}
#endif
