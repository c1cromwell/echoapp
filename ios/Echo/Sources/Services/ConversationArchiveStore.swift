#if os(iOS)
import Foundation

/// Local archive flags per conversation (WO-198); synced via `/v3/conversations/{id}/archive`.
enum ConversationArchiveStore {
    private static let key = "echo.conversation.archive.v1"

    static func isArchived(conversationId: String) -> Bool {
        load()[conversationId] == true
    }

    static func setArchived(_ archived: Bool, conversationId: String) {
        var map = load()
        if archived {
            map[conversationId] = true
        } else {
            map.removeValue(forKey: conversationId)
        }
        persist(map)
    }

    static func archivedIds() -> Set<String> {
        Set(load().filter { $0.value }.map(\.key))
    }

    private static func load() -> [String: Bool] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let decoded = try? JSONDecoder().decode([String: Bool].self, from: data) else {
            return [:]
        }
        return decoded
    }

    private static func persist(_ map: [String: Bool]) {
        guard let data = try? JSONEncoder().encode(map) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}
#endif
