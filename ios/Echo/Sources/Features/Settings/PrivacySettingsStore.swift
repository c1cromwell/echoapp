#if os(iOS)
import Foundation

/// Persists global messaging privacy toggles (typing, read receipts) for Phase 3 signals.
enum PrivacySettingsStore {
    private static let key = "echo.enhancedPrivacySettings.v1"

    static func load() -> EnhancedPrivacySettings {
        guard let data = UserDefaults.standard.data(forKey: key),
              let decoded = try? JSONDecoder().decode(EnhancedPrivacySettings.self, from: data) else {
            return EnhancedPrivacySettings()
        }
        return decoded
    }

    static func save(_ settings: EnhancedPrivacySettings) {
        guard let data = try? JSONEncoder().encode(settings) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}
#endif
