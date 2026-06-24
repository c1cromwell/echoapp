#if os(iOS)
import Foundation

/// Canonical history export for WO-CA3 sync and WO-64 backup (M3b/M3c shared format).
struct HistorySyncBundle: Codable, Sendable, Equatable {
    /// Bundle schema version (increment on breaking changes).
    static let currentVersion = 2

    let version: Int
    let bundleId: String
    let exportedAt: String
    let conversations: [StoredConversation]
    let threads: [String: [StoredThreadMessage]]
    /// Poll state keyed by conversation id → poll id (S4 durability).
    let pollsByConversation: [String: [String: ChatPoll]]

    init(
        version: Int = HistorySyncBundle.currentVersion,
        bundleId: String = UUID().uuidString,
        exportedAt: String = HistorySyncBundle.isoNow(),
        conversations: [StoredConversation],
        threads: [String: [StoredThreadMessage]],
        pollsByConversation: [String: [String: ChatPoll]] = [:]
    ) {
        self.version = version
        self.bundleId = bundleId
        self.exportedAt = exportedAt
        self.conversations = conversations
        self.threads = threads
        self.pollsByConversation = pollsByConversation
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        version = try container.decodeIfPresent(Int.self, forKey: .version) ?? 1
        bundleId = try container.decodeIfPresent(String.self, forKey: .bundleId) ?? UUID().uuidString
        exportedAt = try container.decodeIfPresent(String.self, forKey: .exportedAt) ?? HistorySyncBundle.isoNow()
        conversations = try container.decode([StoredConversation].self, forKey: .conversations)
        threads = try container.decode([String: [StoredThreadMessage]].self, forKey: .threads)
        pollsByConversation = try container.decodeIfPresent(
            [String: [String: ChatPoll]].self,
            forKey: .pollsByConversation
        ) ?? [:]
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

    private enum CodingKeys: String, CodingKey {
        case version
        case bundleId
        case exportedAt
        case conversations
        case threads
        case pollsByConversation
    }
}

// MARK: - Build + merge

@MainActor
enum HistorySyncBundleBuilder {
    /// Snapshot local conversations + thread history + poll state for sync/backup.
    static func build(from store: ConversationStore = .shared) -> HistorySyncBundle {
        let conversations = store.conversations
        var threads: [String: [StoredThreadMessage]] = [:]
        var pollsByConversation: [String: [String: ChatPoll]] = [:]

        let threadIds = Set(conversations.map(\.id)).union(Set(ConversationThreadStore.allStoredConversationIds()))
        for conversationId in threadIds {
            let messages = ConversationThreadStore.exportMessages(conversationId: conversationId)
            if !messages.isEmpty {
                threads[conversationId] = messages
            }
            let polls = ConversationPollStore.load(conversationId: conversationId)
            if !polls.isEmpty {
                pollsByConversation[conversationId] = polls
            }
        }

        return HistorySyncBundle(
            conversations: conversations,
            threads: threads,
            pollsByConversation: pollsByConversation
        )
    }
}

@MainActor
enum HistorySyncBundleMerger {
    /// Idempotent apply: upsert conversations, merge messages by id, union poll votes.
    static func apply(_ bundle: HistorySyncBundle, to store: ConversationStore = .shared) {
        for conversation in bundle.conversations {
            store.upsert(conversation)
        }
        for (conversationId, messages) in bundle.threads {
            ConversationThreadStore.mergeMessages(conversationId: conversationId, incoming: messages)
        }
        for (conversationId, polls) in bundle.pollsByConversation {
            PollMerger.merge(conversationId: conversationId, incoming: polls)
        }
    }
}

enum PollMerger {
    static func merge(conversationId: String, incoming: [String: ChatPoll]) {
        var existing = ConversationPollStore.load(conversationId: conversationId)
        for (pollId, remote) in incoming {
            if let local = existing[pollId] {
                existing[pollId] = union(local, remote)
            } else {
                existing[pollId] = remote
            }
        }
        ConversationPollStore.save(conversationId: conversationId, polls: existing)
    }

    static func union(_ local: ChatPoll, _ remote: ChatPoll) -> ChatPoll {
        var merged = local
        merged.isClosed = local.isClosed || remote.isClosed
        for index in merged.options.indices {
            guard index < remote.options.count else { continue }
            let remoteOption = remote.options[index]
            var voters = Set(merged.options[index].voters)
            voters.formUnion(remoteOption.voters)
            merged.options[index].voters = Array(voters)
            merged.options[index].voteCount = voters.count
        }
        return merged
    }
}
#endif
