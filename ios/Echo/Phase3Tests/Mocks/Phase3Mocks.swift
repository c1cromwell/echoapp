import Foundation
@testable import Echo

/// In-memory transport for ConversationSignalService unit tests.
final class MockConversationSignalTransport: ConversationSignalTransport, @unchecked Sendable {
    private(set) var sentTexts: [String] = []
    private(set) var lastAccessToken: String?
    var onTextMessage: (@Sendable (String) -> Void)?

    func connect(accessToken: String) async throws {
        lastAccessToken = accessToken
    }

    func disconnect() async {}

    func send(text: String) async throws {
        sentTexts.append(text)
    }

    func simulateIncoming(_ text: String) {
        onTextMessage?(text)
    }
}

/// Mock reactions REST client.
final class MockReactionsAPIClient: ReactionsAPIClient, @unchecked Sendable {
    var responses: [String: MessageReactionsResponse] = [:]
    var reactCalls: [(messageId: String, emoji: String)] = []

    func react(messageId: String, emoji: String) async throws -> MessageReactionsResponse {
        reactCalls.append((messageId, emoji))
        if emoji.isEmpty {
            responses[messageId] = MessageReactionsResponse(messageId: messageId, reactions: [])
            return responses[messageId]!
        }
        let existing = responses[messageId]?.reactions ?? []
        let updated = ReactionCount(emoji: emoji, count: 1, reactors: ["did:key:me"])
        responses[messageId] = MessageReactionsResponse(messageId: messageId, reactions: [updated])
        return responses[messageId]!
    }

    func reactions(messageId: String) async throws -> MessageReactionsResponse {
        responses[messageId] ?? MessageReactionsResponse(messageId: messageId, reactions: [])
    }
}

/// Mock durable receipts client (WO-192/48). Records markRead calls and serves
/// canned status responses for reconnect-sync tests.
final class MockMessageReceiptsAPIClient: MessageReceiptsAPIClient, @unchecked Sendable {
    var markReadCalls: [String] = []
    var markDeliveredCalls: [String] = []
    var statuses: [String: MessageStatusResponse] = [:]

    @discardableResult
    func markRead(messageId: String) async throws -> MessageReceiptResponse {
        markReadCalls.append(messageId)
        return MessageReceiptResponse(messageId: messageId, receiptType: "read", timestamp: "")
    }

    @discardableResult
    func markDelivered(messageId: String) async throws -> MessageReceiptResponse {
        markDeliveredCalls.append(messageId)
        return MessageReceiptResponse(messageId: messageId, receiptType: "delivered", timestamp: "")
    }

    func status(messageId: String) async throws -> MessageStatusResponse {
        statuses[messageId]
            ?? MessageStatusResponse(messageId: messageId, conversationId: "conv-1", status: "queued", deliveredAt: nil, readAt: nil)
    }
}
