#if os(iOS)
import Foundation

/// Last 100 local search queries (WO-16).
enum SearchHistoryStore {
    private static let key = "echo.search.history.v1"
    private static let maxCount = 100

    static func load() -> [String] {
        UserDefaults.standard.stringArray(forKey: key) ?? []
    }

    static func append(_ query: String) {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 2 else { return }
        var items = load().filter { $0.caseInsensitiveCompare(trimmed) != .orderedSame }
        items.insert(trimmed, at: 0)
        if items.count > maxCount {
            items = Array(items.prefix(maxCount))
        }
        UserDefaults.standard.set(items, forKey: key)
    }

    static func clear() {
        UserDefaults.standard.removeObject(forKey: key)
    }
}
#endif
