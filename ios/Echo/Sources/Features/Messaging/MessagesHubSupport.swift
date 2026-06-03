import Foundation

/// Pure helpers for Messages hub pinned rows and folder unread badges (Phase B).
enum MessagesHubSupport {
    static func pinnedItems(
        conversations: [StoredConversation],
        orderedPinIDs: [String]
    ) -> [PinnedItem] {
        let byID = Dictionary(uniqueKeysWithValues: conversations.map { ($0.id, $0) })
        return orderedPinIDs.compactMap { id -> PinnedItem? in
            guard let conv = byID[id] else { return nil }
            let initials = initialsFromName(conv.contactName)
            return PinnedItem(
                id: conv.id,
                type: .contact,
                name: conv.contactName,
                initials: initials,
                gradientIndex: abs(conv.id.hashValue) % 6,
                isOnline: conv.isOnline,
                unreadCount: conv.unreadCount
            )
        }
    }

    static func folderUnreadCounts(
        conversations: [StoredConversation],
        trustTier: (String) -> Int
    ) -> [ChatFolder: Int] {
        var counts: [ChatFolder: Int] = [:]
        for folder in ChatFolder.allCases {
            let total = conversations
                .filter { folder.includes(tier: trustTier($0.id)) }
                .reduce(0) { $0 + max(0, $1.unreadCount) }
            if total > 0 { counts[folder] = total }
        }
        return counts
    }

    private static func initialsFromName(_ name: String) -> String {
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.hasPrefix("@") {
            return String(trimmed.dropFirst().prefix(2)).uppercased()
        }
        let parts = trimmed.split(separator: " ")
        if parts.count >= 2 {
            return String(parts[0].prefix(1) + parts[1].prefix(1)).uppercased()
        }
        return String(trimmed.prefix(2)).uppercased()
    }
}
