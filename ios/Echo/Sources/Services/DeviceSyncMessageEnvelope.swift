#if os(iOS)
import Foundation

/// Incremental WO-CA3 delta pushed to linked devices (content encrypted in transit).
struct DeviceSyncMessageEnvelope: Codable, Sendable, Equatable {
    enum Operation: String, Codable, Sendable {
        case upsert
        case tombstone
    }

    let operation: Operation
    let conversationId: String
    let message: StoredThreadMessage?
    let messageId: String?

    static func upsert(conversationId: String, message: StoredThreadMessage) -> DeviceSyncMessageEnvelope {
        DeviceSyncMessageEnvelope(
            operation: .upsert,
            conversationId: conversationId,
            message: message,
            messageId: message.id
        )
    }

    static func tombstone(conversationId: String, messageId: String) -> DeviceSyncMessageEnvelope {
        DeviceSyncMessageEnvelope(
            operation: .tombstone,
            conversationId: conversationId,
            message: nil,
            messageId: messageId
        )
    }
}
#endif
