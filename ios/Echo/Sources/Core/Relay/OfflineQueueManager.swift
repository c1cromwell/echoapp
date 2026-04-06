import Foundation

/// Manages local outbox for offline message sends.
/// Queues encrypted relay requests when WebSocket is unavailable
/// and drains them on reconnection.
actor OfflineQueueManager {

    private var queue: [RelayRequest] = []
    private let maxQueueSize = 1000

    /// Number of queued messages
    var count: Int {
        queue.count
    }

    /// Enqueue a relay request for later send
    func enqueue(_ request: RelayRequest) {
        if queue.count >= maxQueueSize {
            // Evict oldest to stay within limit
            queue.removeFirst()
        }
        queue.append(request)
    }

    /// Dequeue all pending requests for drain
    func dequeueAll() -> [RelayRequest] {
        let items = queue
        queue.removeAll()
        return items
    }

    /// Peek at queued requests without removing
    func peek() -> [RelayRequest] {
        queue
    }

    /// Remove a specific request by message ID
    func remove(messageId: String) {
        queue.removeAll { $0.messageId == messageId }
    }

    /// Clear all queued requests
    func clear() {
        queue.removeAll()
    }
}
