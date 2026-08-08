#if os(iOS)
import Foundation
import Observation

// MARK: - Model

/// A user-defined chat folder (Telegram-style). Membership is an ordered list of
/// conversation ids; `order` positions the folder in the hub selector.
struct CustomChatFolder: Identifiable, Codable, Hashable, Sendable {
    var id: String
    var name: String
    var order: Int
    var conversationIds: [String]

    init(id: String = UUID().uuidString, name: String, order: Int, conversationIds: [String] = []) {
        self.id = id
        self.name = name
        self.order = order
        self.conversationIds = conversationIds
    }
}

// MARK: - Store

/// Local + cross-device chat folders. Persists to UserDefaults and best-effort
/// syncs the whole list to `/v3/chat-folders` (last-write-wins). Mirrors the
/// `ConversationArchiveStore` / synced-preference pattern.
@MainActor
@Observable
final class ChatFolderStore {
    static let shared = ChatFolderStore()

    private(set) var folders: [CustomChatFolder] = []
    /// nil = "All chats" (no folder filter).
    var selectedFolderId: String?

    private let storageKey = "echo.chatfolders.v1"

    private init() { loadLocal() }

    // MARK: Queries

    var sortedFolders: [CustomChatFolder] {
        folders.sorted { $0.order < $1.order }
    }

    func folder(id: String) -> CustomChatFolder? { folders.first { $0.id == id } }

    func contains(_ conversationId: String, in folderId: String) -> Bool {
        folder(id: folderId)?.conversationIds.contains(conversationId) ?? false
    }

    /// Conversation ids for the selected folder, or nil when "All" is selected.
    var selectedConversationIds: Set<String>? {
        guard let id = selectedFolderId, let f = folder(id: id) else { return nil }
        return Set(f.conversationIds)
    }

    // MARK: Mutations

    @discardableResult
    func createFolder(name: String) -> CustomChatFolder {
        let clean = name.trimmingCharacters(in: .whitespacesAndNewlines)
        let folder = CustomChatFolder(name: clean, order: (folders.map(\.order).max() ?? -1) + 1)
        folders.append(folder)
        persist()
        return folder
    }

    func renameFolder(id: String, name: String) {
        let clean = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !clean.isEmpty, let idx = folders.firstIndex(where: { $0.id == id }) else { return }
        folders[idx].name = clean
        persist()
    }

    func deleteFolder(id: String) {
        folders.removeAll { $0.id == id }
        if selectedFolderId == id { selectedFolderId = nil }
        persist()
    }

    func setConversations(_ ids: [String], for folderId: String) {
        guard let idx = folders.firstIndex(where: { $0.id == folderId }) else { return }
        folders[idx].conversationIds = ids
        persist()
    }

    func toggleConversation(_ conversationId: String, in folderId: String) {
        guard let idx = folders.firstIndex(where: { $0.id == folderId }) else { return }
        if let pos = folders[idx].conversationIds.firstIndex(of: conversationId) {
            folders[idx].conversationIds.remove(at: pos)
        } else {
            folders[idx].conversationIds.append(conversationId)
        }
        persist()
    }

    // MARK: Persistence + sync

    private func loadLocal() {
        guard let data = UserDefaults.standard.data(forKey: storageKey),
              let list = try? JSONDecoder().decode([CustomChatFolder].self, from: data) else { return }
        folders = list
    }

    private func saveLocal() {
        if let data = try? JSONEncoder().encode(folders) {
            UserDefaults.standard.set(data, forKey: storageKey)
        }
    }

    /// Save locally and push the full list to the server (best effort).
    private func persist() {
        saveLocal()
        let snapshot = folders
        Task { await Self.pushToServer(snapshot) }
    }

    /// Pull the server copy and replace local state (call on hub appear / login).
    func hydrateFromServer() async {
        guard let api = DIContainer.shared.resolveAPIClient() else { return }
        let client = LiveChatFolderAPIClient(apiClient: api)
        guard let remote = try? await client.fetch() else { return }
        folders = remote
        saveLocal()
    }

    private static func pushToServer(_ folders: [CustomChatFolder]) async {
        guard let api = await DIContainer.shared.resolveAPIClient() else { return }
        let client = LiveChatFolderAPIClient(apiClient: api)
        try? await client.save(folders)
    }
}

// MARK: - Sync client

private struct ChatFolderWire: Codable, Sendable {
    let id: String
    let name: String
    let order: Int
    let conversationIds: [String]

    enum CodingKeys: String, CodingKey {
        case id, name, order
        case conversationIds = "conversation_ids"
    }
}

private struct ChatFoldersEnvelope: Codable, Sendable {
    let folders: [ChatFolderWire]
}

private enum ChatFolderEndpoint: APIEndpoint {
    case folders
    var path: String { "/v3/chat-folders" }
}

actor LiveChatFolderAPIClient {
    private let apiClient: APIClient
    init(apiClient: APIClient) { self.apiClient = apiClient }

    func fetch() async throws -> [CustomChatFolder] {
        let env: ChatFoldersEnvelope = try await apiClient.get(endpoint: ChatFolderEndpoint.folders)
        return env.folders.map { CustomChatFolder(id: $0.id, name: $0.name, order: $0.order, conversationIds: $0.conversationIds) }
    }

    func save(_ folders: [CustomChatFolder]) async throws {
        let body = ChatFoldersEnvelope(
            folders: folders.map { ChatFolderWire(id: $0.id, name: $0.name, order: $0.order, conversationIds: $0.conversationIds) }
        )
        let _: ChatFoldersEnvelope = try await apiClient.put(endpoint: ChatFolderEndpoint.folders, body: body)
    }
}
#endif
