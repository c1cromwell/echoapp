#if os(iOS)
import Foundation

/// Local persona list + active persona persistence for hub switcher (Phase B).
/// The persona list is persisted (add/edit/delete). `"default"` (primary) and
/// `"hidden"` (drives the hidden-vault gate) are system personas and cannot be
/// deleted; the primary can be renamed.
enum PersonaSessionStore {
    private static let activeKey = "echo.activePersonaId"
    private static let listKey = "echo.personas.v1"

    static let primaryPersonaId = "default"
    static let hiddenPersonaId = "hidden"

    // MARK: - Active persona

    static var activePersonaId: String {
        UserDefaults.standard.string(forKey: activeKey) ?? primaryPersonaId
    }

    static func setActivePersonaId(_ id: String) {
        UserDefaults.standard.set(id, forKey: activeKey)
    }

    static func resolveActive(in personas: [PersonaSummary]) -> PersonaSummary {
        let id = activePersonaId
        return personas.first { $0.id == id } ?? personas.first!
    }

    // MARK: - Persona list (persisted)

    /// The persisted persona list, seeded from `defaultPersonas` on first use.
    static func personas(displayName: String) -> [PersonaSummary] {
        if let stored = loadStored(), !stored.isEmpty {
            return stored
        }
        let seed = defaultPersonas(displayName: displayName)
        save(seed)
        return seed
    }

    static func save(_ personas: [PersonaSummary]) {
        guard let data = try? JSONEncoder().encode(personas) else { return }
        UserDefaults.standard.set(data, forKey: listKey)
    }

    private static func loadStored() -> [PersonaSummary]? {
        guard let data = UserDefaults.standard.data(forKey: listKey),
              let list = try? JSONDecoder().decode([PersonaSummary].self, from: data) else {
            return nil
        }
        return list
    }

    // MARK: - CRUD

    static func isSystemPersona(_ id: String) -> Bool {
        id == primaryPersonaId || id == hiddenPersonaId
    }

    /// Appends a new custom persona and returns the updated list.
    @discardableResult
    static func add(_ persona: PersonaSummary, to personas: [PersonaSummary]) -> [PersonaSummary] {
        var list = personas
        list.append(persona)
        save(list)
        return list
    }

    /// Replaces the persona with the same id (by value) and returns the updated list.
    @discardableResult
    static func update(_ persona: PersonaSummary, in personas: [PersonaSummary]) -> [PersonaSummary] {
        var list = personas
        if let idx = list.firstIndex(where: { $0.id == persona.id }) {
            list[idx] = persona
            save(list)
        }
        return list
    }

    /// Deletes a custom persona (system personas are ignored) and returns the updated list.
    @discardableResult
    static func delete(id: String, from personas: [PersonaSummary]) -> [PersonaSummary] {
        guard !isSystemPersona(id) else { return personas }
        var list = personas
        list.removeAll { $0.id == id }
        save(list)
        return list
    }

    /// Renames the primary persona (recomputing its initials) and returns the updated list.
    @discardableResult
    static func renamePrimary(name: String, in personas: [PersonaSummary]) -> [PersonaSummary] {
        var list = personas
        let clean = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !clean.isEmpty, let idx = list.firstIndex(where: { $0.id == primaryPersonaId }) else {
            return list
        }
        let current = list[idx]
        list[idx] = PersonaSummary(
            id: current.id,
            name: clean,
            initials: initials(from: clean),
            trustLevel: current.trustLevel,
            isHidden: current.isHidden,
            colorHex: current.colorHex
        )
        save(list)
        return list
    }

    /// A fresh id for a user-created persona (avoids clashing with system ids).
    static func newPersonaId() -> String {
        "persona-\(UUID().uuidString.prefix(8).lowercased())"
    }

    // MARK: - Defaults / helpers

    static func defaultPersonas(displayName: String) -> [PersonaSummary] {
        let primaryName = displayName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? "You" : displayName
        return [
            PersonaSummary(id: primaryPersonaId, name: primaryName, initials: initials(from: primaryName), trustLevel: "Verified"),
            PersonaSummary(id: "work", name: "Work", initials: "WK", trustLevel: "Trusted", colorHex: 0x1F7A4C),
            PersonaSummary(id: hiddenPersonaId, name: "Hidden", initials: "", isHidden: true, colorHex: 0x0B1220),
        ]
    }

    static func initials(from name: String) -> String {
        let parts = name.split(separator: " ")
        if parts.count >= 2 {
            return String(parts[0].prefix(1) + parts[1].prefix(1)).uppercased()
        }
        let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
        return String(trimmed.prefix(2)).uppercased()
    }
}
#endif
