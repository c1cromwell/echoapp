#if os(iOS)
import Foundation

/// Local persona list + active persona persistence for hub switcher (Phase B).
enum PersonaSessionStore {
    private static let activeKey = "echo.activePersonaId"

    static var activePersonaId: String {
        UserDefaults.standard.string(forKey: activeKey) ?? "default"
    }

    static func setActivePersonaId(_ id: String) {
        UserDefaults.standard.set(id, forKey: activeKey)
    }

    static func defaultPersonas(displayName: String) -> [PersonaSummary] {
        let initials = initials(from: displayName)
        let primary = displayName.isEmpty ? "Me" : displayName
        return [
            PersonaSummary(id: "default", name: primary, initials: initials, trustLevel: "Verified"),
            PersonaSummary(id: "work", name: "Work", initials: "WK", trustLevel: "Trusted", colorHex: 0x1F7A4C),
            PersonaSummary(id: "hidden", name: "Hidden", initials: "", isHidden: true, colorHex: 0x0B1220),
        ]
    }

    static func resolveActive(in personas: [PersonaSummary]) -> PersonaSummary {
        let id = activePersonaId
        return personas.first { $0.id == id } ?? personas.first!
    }

    private static func initials(from name: String) -> String {
        let parts = name.split(separator: " ")
        if parts.count >= 2 {
            return String(parts[0].prefix(1) + parts[1].prefix(1)).uppercased()
        }
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        return String(trimmed.prefix(2)).uppercased()
    }
}
#endif
