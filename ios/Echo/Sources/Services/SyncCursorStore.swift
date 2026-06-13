#if os(iOS)
import Foundation

/// Persists last consumed sync seq per device id (WO-CA3 pull cursor).
enum SyncCursorStore {
    private static let keyPrefix = "echo.sync.cursor."

    static func load(deviceId: String) -> Int64 {
        guard !deviceId.isEmpty else { return 0 }
        let value = UserDefaults.standard.object(forKey: keyPrefix + deviceId) as? Int64
        return value ?? Int64(UserDefaults.standard.integer(forKey: keyPrefix + deviceId))
    }

    static func save(deviceId: String, cursor: Int64) {
        guard !deviceId.isEmpty else { return }
        UserDefaults.standard.set(cursor, forKey: keyPrefix + deviceId)
    }

    #if DEBUG
    static func reset(deviceId: String) {
        UserDefaults.standard.removeObject(forKey: keyPrefix + deviceId)
    }
    #endif
}
#endif
