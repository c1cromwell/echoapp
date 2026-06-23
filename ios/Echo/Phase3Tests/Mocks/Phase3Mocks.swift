import Foundation
@testable import Echo

/// In-memory transport for ConversationSignalService unit tests.
final class MockConversationSignalTransport: ConversationSignalTransport, @unchecked Sendable {
    var sentTexts: [String] = []
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

/// Mock message-ops client (WO-25/84/59). Records calls; `pinShouldFail` simulates
/// a server rejection (e.g. pin limit).
final class MockMessageOpsAPIClient: MessageOpsAPIClient, @unchecked Sendable {
    var editCalls: [(messageId: String, conversationId: String)] = []
    var deleteCalls: [String] = []
    var pinCalls: [String] = []
    var unpinCalls: [String] = []
    var disappearingCalls: [(conversationId: String, ttl: Int)] = []
    var pinShouldFail = false

    struct OpError: Error {}

    @discardableResult
    func editMessage(messageId: String, conversationId: String, ciphertext: Data) async throws -> MessageEditResult {
        editCalls.append((messageId, conversationId))
        return MessageEditResult(messageId: messageId, edited: true, retained: false, version: 0)
    }

    @discardableResult
    func deleteMessage(messageId: String, conversationId: String) async throws -> MessageDeleteResult {
        deleteCalls.append(messageId)
        return MessageDeleteResult(messageId: messageId, deleted: true, retained: false)
    }

    @discardableResult
    func pinMessage(messageId: String, conversationId: String) async throws -> MessagePinResult {
        if pinShouldFail { throw OpError() }
        pinCalls.append(messageId)
        return MessagePinResult(messageId: messageId, pinned: true)
    }

    @discardableResult
    func unpinMessage(messageId: String, conversationId: String) async throws -> MessagePinResult {
        unpinCalls.append(messageId)
        return MessagePinResult(messageId: messageId, pinned: false)
    }

    @discardableResult
    func setDisappearing(conversationId: String, ttlSeconds: Int, peerDID: String?) async throws -> DisappearingConfigResult {
        disappearingCalls.append((conversationId, ttlSeconds))
        return DisappearingConfigResult(conversationId: conversationId, ttlSeconds: ttlSeconds)
    }
}
