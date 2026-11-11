import Foundation

/// Temporary permission for an agent to search the local message index.
/// The index key stays in `EncryptedIndexStore`; a hit never carries a store path.
struct AgentSearchGrant: Equatable, Sendable {
    var verb: String
    var expiresAt: Date

    var allowsSearch: Bool {
        verb == "search" && expiresAt > Date()
    }
}

struct AgentSearchHit: Equatable, Sendable {
    var messageId: String
    var matchedText: String
}

enum AgentEdgeSearch {
    /// Local-only search. An expired or non-search grant returns nothing.
    static func hits(
        query: String,
        grant: AgentSearchGrant,
        corpus: [(id: String, text: String)]
    ) -> [AgentSearchHit] {
        guard grant.allowsSearch else { return [] }
        let needle = query.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !needle.isEmpty else { return [] }
        return corpus.compactMap { item in
            guard item.text.lowercased().contains(needle) else { return nil }
            return AgentSearchHit(messageId: item.id, matchedText: item.text)
        }
    }
}
