#if os(iOS)
import Foundation

/// Per-contact persona override for minimal disclosure (WO-SX6).
enum ContactPersonaStore {
    private static func key(contactDID: String) -> String {
        "echo.contact.persona.\(contactDID)"
    }

    static func personaId(for contactDID: String) -> String? {
        UserDefaults.standard.string(forKey: key(contactDID: contactDID))
    }

    static func setPersonaId(_ personaId: String?, for contactDID: String) {
        let k = key(contactDID: contactDID)
        if let personaId, !personaId.isEmpty {
            UserDefaults.standard.set(personaId, forKey: k)
        } else {
            UserDefaults.standard.removeObject(forKey: k)
        }
    }

    /// Fields visible when using a reduced-disclosure persona toward a contact.
    static func minimalDisplayName(base: String, personaId: String?) -> String {
        guard let personaId, personaId != "default" else { return base }
        if personaId == "hidden" { return "ECHO user" }
        return base
    }

    static func shouldHideTrustBadge(personaId: String?) -> Bool {
        personaId == "hidden"
    }

    static func shouldHideAvatar(personaId: String?) -> Bool {
        personaId == "hidden"
    }
}
#endif
