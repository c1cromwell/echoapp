#if os(iOS)
import Foundation

/// Local favorite contacts (WO-39); persisted per DID.
enum ContactFavoritesStore {
    private static let key = "echo.contactFavorites"

    static func isFavorite(did: String) -> Bool {
        load().contains(did)
    }

    static func toggle(did: String) {
        var set = load()
        if set.contains(did) {
            set.remove(did)
        } else {
            set.insert(did)
        }
        UserDefaults.standard.set(Array(set), forKey: key)
    }

    static func load() -> Set<String> {
        Set(UserDefaults.standard.stringArray(forKey: key) ?? [])
    }
}
#endif
