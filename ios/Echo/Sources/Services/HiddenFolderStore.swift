#if os(iOS)
import Foundation

/// Lightweight hidden-folder metadata (WO-30 subset).
struct HiddenFolder: Codable, Identifiable, Equatable, Sendable {
    let id: String
    var name: String
    var createdAt: Date
    var conversationIds: [String]
}

/// Device-local hidden folder registry (max 20).
@MainActor
enum HiddenFolderStore {
    static let maxFolders = 20
    private static let key = "echo.hidden.folders.v1"
    private static let defaultFolderIdentifier = "default"

    static func all() -> [HiddenFolder] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let rows = try? JSONDecoder().decode([HiddenFolder].self, from: data) else {
            return [defaultFolder()]
        }
        return rows.isEmpty ? [defaultFolder()] : rows
    }

    static func folder(id: String) -> HiddenFolder? {
        all().first { $0.id == id }
    }

    static func defaultFolderId() -> String { defaultFolderIdentifier }

    @discardableResult
    static func create(name: String) throws -> HiddenFolder {
        var rows = all()
        guard rows.count < maxFolders else {
            throw HiddenFolderStoreError.limitReached
        }
        let trimmed = String(name.prefix(50)).trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { throw HiddenFolderStoreError.invalidName }
        let folder = HiddenFolder(
            id: UUID().uuidString,
            name: trimmed,
            createdAt: Date(),
            conversationIds: []
        )
        rows.append(folder)
        persist(rows)
        return folder
    }

    static func rename(id: String, name: String) {
        var rows = all()
        guard let idx = rows.firstIndex(where: { $0.id == id }) else { return }
        rows[idx].name = String(name.prefix(50))
        persist(rows)
    }

    static func assign(conversationId: String, folderId: String) {
        var rows = all()
        for i in rows.indices {
            rows[i].conversationIds.removeAll { $0 == conversationId }
        }
        guard let idx = rows.firstIndex(where: { $0.id == folderId }) else { return }
        if !rows[idx].conversationIds.contains(conversationId) {
            rows[idx].conversationIds.append(conversationId)
        }
        persist(rows)
    }

    static func folderId(for conversationId: String) -> String {
        all().first { $0.conversationIds.contains(conversationId) }?.id ?? defaultFolderIdentifier
    }

    static func delete(id: String) {
        guard id != defaultFolderIdentifier else { return }
        var rows = all().filter { $0.id != id }
        if rows.isEmpty { rows = [defaultFolder()] }
        persist(rows)
    }

    private static func defaultFolder() -> HiddenFolder {
        HiddenFolder(id: defaultFolderIdentifier, name: "Hidden", createdAt: Date(), conversationIds: [])
    }

    private static func persist(_ rows: [HiddenFolder]) {
        guard let data = try? JSONEncoder().encode(rows) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }
}

enum HiddenFolderStoreError: LocalizedError {
    case limitReached
    case invalidName

    var errorDescription: String? {
        switch self {
        case .limitReached: return "You can create up to 20 hidden folders."
        case .invalidName: return "Folder name is required."
        }
    }
}
#endif
