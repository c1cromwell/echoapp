import Foundation
import Observation

/// Persisted pin order for the Messages hub (spec §5.5, max 5).
@MainActor
@Observable
public final class PinnedConversationsStore {
    public static let shared = PinnedConversationsStore()

    public static let maxPins = 5

    private(set) var orderedIDs: [String] = []
    private let storageKey = "echo.pinned.conversations.v1"

    private init() {
        load()
    }

    public func isPinned(_ conversationId: String) -> Bool {
        orderedIDs.contains(conversationId)
    }

    public func pin(_ conversationId: String) {
        guard !conversationId.isEmpty else { return }
        orderedIDs.removeAll { $0 == conversationId }
        orderedIDs.insert(conversationId, at: 0)
        if orderedIDs.count > Self.maxPins {
            orderedIDs = Array(orderedIDs.prefix(Self.maxPins))
        }
        persist()
    }

    public func unpin(_ conversationId: String) {
        orderedIDs.removeAll { $0 == conversationId }
        persist()
    }

    public func toggle(_ conversationId: String) {
        if isPinned(conversationId) { unpin(conversationId) } else { pin(conversationId) }
    }

    public func move(from source: IndexSet, to destination: Int) {
        orderedIDs.move(fromOffsets: source, toOffset: destination)
        persist()
    }

    private func persist() {
        UserDefaults.standard.set(orderedIDs, forKey: storageKey)
    }

    private func load() {
        orderedIDs = UserDefaults.standard.stringArray(forKey: storageKey) ?? []
    }
}
