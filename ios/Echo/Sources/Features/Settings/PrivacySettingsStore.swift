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

    /// Persists locally and best-effort syncs profile privacy to the server (WO-187).
    static func saveAndSync(_ settings: EnhancedPrivacySettings) async {
        save(settings)
        if let useCase = await DIContainer.shared.resolveSyncProfilePrivacyUseCase() {
            try? await useCase.sync(settings)
        }
    }

    /// WO-228: whether chat surfaces show the encrypted-thread bar.
    static var showsEncryptionIndicator: Bool {
        load().showEncryptionIndicator
    }
}

/// Per-persona messaging privacy (typing / read receipts) for Phase 3 signal merge.
enum PersonaPrivacySettingsStore {
    private static func key(personaId: String) -> String {
        "echo.personaPrivacy.\(personaId).v1"
    }

    static func load(personaId: String = PersonaSessionStore.activePersonaId) -> PersonaPrivacySettings {
        guard let data = UserDefaults.standard.data(forKey: key(personaId: personaId)),
              let decoded = try? JSONDecoder().decode(PersonaPrivacySettings.self, from: data) else {
            return PersonaPrivacySettings()
        }
        return decoded
    }

    static func save(_ settings: PersonaPrivacySettings, personaId: String = PersonaSessionStore.activePersonaId) {
        guard let data = try? JSONEncoder().encode(settings) else { return }
        UserDefaults.standard.set(data, forKey: key(personaId: personaId))
    }
}
#endif
