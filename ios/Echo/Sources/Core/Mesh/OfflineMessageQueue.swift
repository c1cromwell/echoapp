// Core/Mesh/OfflineMessageQueue.swift
//
// Store-and-forward for the mesh (P2): outbound frames that can't be delivered yet (no peer in
// range) are persisted and retried when a peer appears. Maps to DeliveryStatus — `.sending` while
// queued, dropped on delivery, `.failed` after maxAttempts.

import Foundation

public struct QueuedMeshMessage: Codable, Equatable, Sendable {
    public let id: String              // message id (dedup key)
    public let conversationId: String
    public let peerDID: String
    public let wire: String            // the encoded ConversationSignal envelope to (re)send
    public var attempts: Int

    public init(id: String, conversationId: String, peerDID: String, wire: String, attempts: Int = 0) {
        self.id = id
        self.conversationId = conversationId
        self.peerDID = peerDID
        self.wire = wire
        self.attempts = attempts
    }
}

public protocol OfflineQueueStore {
    func load() -> [QueuedMeshMessage]
    func save(_ items: [QueuedMeshMessage])
}

public final class UserDefaultsOfflineQueueStore: OfflineQueueStore {
    private let key = "echo.mesh.offlineQueue.v1"
    public init() {}
    public func load() -> [QueuedMeshMessage] {
        guard let data = UserDefaults.standard.data(forKey: key) else { return [] }
        return (try? JSONDecoder().decode([QueuedMeshMessage].self, from: data)) ?? []
    }
    public func save(_ items: [QueuedMeshMessage]) {
        UserDefaults.standard.set(try? JSONEncoder().encode(items), forKey: key)
    }
}

public final class OfflineMessageQueue {
    /// Attempt to transmit one queued frame; return true if it was accepted (delivered/relayed).
    public typealias Sender = (QueuedMeshMessage) async -> Bool

    private var items: [QueuedMeshMessage]
    private let store: OfflineQueueStore
    private let maxAttempts: Int

    public init(store: OfflineQueueStore = UserDefaultsOfflineQueueStore(), maxAttempts: Int = 8) {
        self.store = store
        self.maxAttempts = maxAttempts
        self.items = store.load()
    }

    public var pending: [QueuedMeshMessage] { items }

    /// Queue an outbound frame (idempotent on message id).
    @discardableResult
    public func enqueue(_ message: QueuedMeshMessage) -> Bool {
        guard !items.contains(where: { $0.id == message.id }) else { return false }
        items.append(message)
        store.save(items)
        return true
    }

    /// Drop a message once a peer acknowledges it.
    public func markDelivered(_ id: String) {
        items.removeAll { $0.id == id }
        store.save(items)
    }

    /// Try to flush all pending via `send`. Delivered frames are dropped; the rest have their
    /// attempt count bumped and are given up on (treated `.failed`) after `maxAttempts`.
    public func flush(using send: Sender) async {
        var remaining: [QueuedMeshMessage] = []
        for var message in items {
            if await send(message) { continue }
            message.attempts += 1
            if message.attempts < maxAttempts { remaining.append(message) }
        }
        items = remaining
        store.save(items)
    }
}
