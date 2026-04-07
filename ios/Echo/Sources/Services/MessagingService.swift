import Foundation

/// Messaging service for message operations
public class MessagingService: BaseEchoService {
    private var messages: [String: Message] = [:]
    private var conversations: [String: Conversation] = [:]
    
    public override init(name: String = "messaging-service", version: String = "v1") {
        super.init(name: name, version: version)
    }
    
    // MARK: - Models
    
    public enum MessageType {
        case text
        case voice
        case video
        case image
    }
    
    public struct Message: Identifiable {
        public let id: String
        public let senderID: String
        public let recipientID: String?
        public let conversationID: String
        public let messageType: MessageType
        public let encryptedContent: Data
        public let timestamp: Date
        public var isRead: Bool = false
        public var readAt: Date? = nil
    }
    
    public struct Conversation: Identifiable {
        public let id: String
        public let type: String // "direct" or "group"
        public let participants: [String]
        public let createdAt: Date
        public var updatedAt: Date
        public var lastMessage: Message? = nil
    }
    
    // MARK: - Conversation Management
    
    /// Create a new conversation
    public func createConversation(participants: [String]) async throws -> Conversation {
        guard participants.count >= 2 else {
            throw MessagingError.invalidParticipants
        }
        
        let convID = generateConversationID()
        let conversation = Conversation(
            id: convID,
            type: "direct",
            participants: participants,
            createdAt: Date(),
            updatedAt: Date()
        )
        
        conversations[convID] = conversation
        return conversation
    }
    
    /// Get conversation by ID
    public func getConversation(_ conversationID: String) async throws -> Conversation {
        guard let conversation = conversations[conversationID] else {
            throw MessagingError.conversationNotFound
        }
        return conversation
    }
    
    // MARK: - Message Management
    
    /// Send a message
    public func sendMessage(
        senderID: String,
        conversationID: String,
        messageType: MessageType,
        encryptedContent: Data
    ) async throws -> Message {
        guard !senderID.isEmpty else {
            throw MessagingError.invalidSender
        }
        
        guard let _ = conversations[conversationID] else {
            throw MessagingError.conversationNotFound
        }
        
        let message = Message(
            id: generateMessageID(),
            senderID: senderID,
            recipientID: nil,
            conversationID: conversationID,
            messageType: messageType,
            encryptedContent: encryptedContent,
            timestamp: Date()
        )
        
        messages[message.id] = message
        
        // Update conversation
        if var conversation = conversations[conversationID] {
            conversation.lastMessage = message
            conversation.updatedAt = Date()
            conversations[conversationID] = conversation
        }
        
        return message
    }
    
    /// Get message by ID
    public func getMessage(_ messageID: String) async throws -> Message {
        guard let message = messages[messageID] else {
            throw MessagingError.messageNotFound
        }
        return message
    }
    
    /// Mark message as read
    public func markAsRead(_ messageID: String) async throws {
        guard var message = messages[messageID] else {
            throw MessagingError.messageNotFound
        }
        
        message.isRead = true
        message.readAt = Date()
        messages[messageID] = message
    }
    
    // MARK: - Helper Methods
    
    private func generateConversationID() -> String {
        return "conv-" + ISO8601DateFormatter().string(from: Date())
    }
    
    private func generateMessageID() -> String {
        return "msg-" + UUID().uuidString.prefix(12).lowercased()
    }
}

/// Messaging service errors
public enum MessagingError: LocalizedError {
    case invalidParticipants
    case invalidSender
    case conversationNotFound
    case messageNotFound
    
    public var errorDescription: String? {
        switch self {
        case .invalidParticipants:
            return "At least 2 participants are required"
        case .invalidSender:
            return "Sender ID is required"
        case .conversationNotFound:
            return "Conversation not found"
        case .messageNotFound:
            return "Message not found"
        }
    }
}
