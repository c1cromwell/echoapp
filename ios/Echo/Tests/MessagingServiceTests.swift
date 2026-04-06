import XCTest
@testable import Echo

final class MessagingServiceTests: XCTestCase {
    
    var service: MessagingService!
    
    override func setUp() {
        super.setUp()
        service = MessagingService()
    }
    
    @MainActor
    func testCreateConversation() async throws {
        let participants = ["user1", "user2"]
        let conversation = try await service.createConversation(participants: participants)
        
        XCTAssertFalse(conversation.id.isEmpty)
        XCTAssertEqual(conversation.type, "direct")
        XCTAssertEqual(conversation.participants.count, 2)
    }
    
    @MainActor
    func testCreateConversationInvalidParticipants() async throws {
        do {
            _ = try await service.createConversation(participants: ["user1"])
            XCTFail("Should throw invalidParticipants error")
        } catch MessagingError.invalidParticipants {
            // Expected
        }
    }
    
    @MainActor
    func testGetConversation() async throws {
        let participants = ["user1", "user2"]
        let created = try await service.createConversation(participants: participants)
        
        let retrieved = try await service.getConversation(created.id)
        XCTAssertEqual(retrieved.id, created.id)
        XCTAssertEqual(retrieved.type, "direct")
    }
    
    @MainActor
    func testGetConversationNotFound() async throws {
        do {
            _ = try await service.getConversation("nonexistent")
            XCTFail("Should throw conversationNotFound error")
        } catch MessagingError.conversationNotFound {
            // Expected
        }
    }
    
    @MainActor
    func testSendMessage() async throws {
        let participants = ["user1", "user2"]
        let conversation = try await service.createConversation(participants: participants)
        
        let testData = "test message".data(using: .utf8)!
        let message = try await service.sendMessage(
            senderID: "user1",
            conversationID: conversation.id,
            messageType: .text,
            encryptedContent: testData
        )
        
        XCTAssertFalse(message.id.isEmpty)
        XCTAssertEqual(message.senderID, "user1")
        XCTAssertEqual(message.messageType, .text)
        XCTAssertFalse(message.isRead)
    }
    
    @MainActor
    func testSendMessageInvalidSender() async throws {
        let participants = ["user1", "user2"]
        let conversation = try await service.createConversation(participants: participants)
        
        let testData = "test message".data(using: .utf8)!
        
        do {
            _ = try await service.sendMessage(
                senderID: "",
                conversationID: conversation.id,
                messageType: .text,
                encryptedContent: testData
            )
            XCTFail("Should throw invalidSender error")
        } catch MessagingError.invalidSender {
            // Expected
        }
    }
    
    @MainActor
    func testGetMessage() async throws {
        let participants = ["user1", "user2"]
        let conversation = try await service.createConversation(participants: participants)
        
        let testData = "test message".data(using: .utf8)!
        let sent = try await service.sendMessage(
            senderID: "user1",
            conversationID: conversation.id,
            messageType: .text,
            encryptedContent: testData
        )
        
        let retrieved = try await service.getMessage(sent.id)
        XCTAssertEqual(retrieved.id, sent.id)
        XCTAssertEqual(retrieved.senderID, "user1")
    }
    
    @MainActor
    func testMarkAsRead() async throws {
        let participants = ["user1", "user2"]
        let conversation = try await service.createConversation(participants: participants)
        
        let testData = "test message".data(using: .utf8)!
        let message = try await service.sendMessage(
            senderID: "user1",
            conversationID: conversation.id,
            messageType: .text,
            encryptedContent: testData
        )
        
        XCTAssertFalse(message.isRead)
        
        try await service.markAsRead(message.id)
        
        let updated = try await service.getMessage(message.id)
        XCTAssertTrue(updated.isRead)
        XCTAssertNotNil(updated.readAt)
    }
    
    @MainActor
    func testConversationLastMessage() async throws {
        let participants = ["user1", "user2"]
        let conversation = try await service.createConversation(participants: participants)
        
        let testData = "test message".data(using: .utf8)!
        let message = try await service.sendMessage(
            senderID: "user1",
            conversationID: conversation.id,
            messageType: .text,
            encryptedContent: testData
        )
        
        let updated = try await service.getConversation(conversation.id)
        XCTAssertNotNil(updated.lastMessage)
        XCTAssertEqual(updated.lastMessage?.id, message.id)
    }
}
