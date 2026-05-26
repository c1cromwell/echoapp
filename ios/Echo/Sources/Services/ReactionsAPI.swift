import Foundation

// MARK: - REST models (match internal/services/messaging/reactions.go)

struct ReactionCount: Codable, Sendable, Equatable, Identifiable {
    var id: String { emoji }
    let emoji: String
    let count: Int
    let reactors: [String]
}

struct MessageReactionsResponse: Codable, Sendable, Equatable {
    let messageId: String
    let reactions: [ReactionCount]

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case reactions
    }
}

struct MessageReactRequest: Codable, Sendable, Equatable {
    let messageId: String
    let emoji: String

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case emoji
    }
}

// MARK: - Client protocol (mock in tests)

protocol ReactionsAPIClient: Sendable {
    func react(messageId: String, emoji: String) async throws -> MessageReactionsResponse
    func removeReaction(messageId: String) async throws -> MessageReactionsResponse
    func reactions(messageId: String) async throws -> MessageReactionsResponse
}

extension ReactionsAPIClient {
    func removeReaction(messageId: String) async throws -> MessageReactionsResponse {
        try await react(messageId: messageId, emoji: "")
    }
}

#if os(iOS)

enum ReactionsEndpoint: APIEndpoint {
    case react
    case list(messageId: String)

    var path: String {
        switch self {
        case .react:
            return "/v3/messages/react"
        case .list(let messageId):
            return "/v3/messages/reactions?message_id=\(messageId.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? messageId)"
        }
    }
}

/// Durable reaction source of truth via signed REST (PasskeySigningInterceptor on APIClient).
actor ReactionsAPI: ReactionsAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func react(messageId: String, emoji: String) async throws -> MessageReactionsResponse {
        try await apiClient.post(
            endpoint: ReactionsEndpoint.react,
            body: MessageReactRequest(messageId: messageId, emoji: emoji)
        )
    }

    func reactions(messageId: String) async throws -> MessageReactionsResponse {
        try await apiClient.get(endpoint: ReactionsEndpoint.list(messageId: messageId))
    }
}

#endif
