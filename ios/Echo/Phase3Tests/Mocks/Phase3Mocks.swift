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
