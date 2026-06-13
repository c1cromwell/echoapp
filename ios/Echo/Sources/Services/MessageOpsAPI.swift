import Foundation

// MARK: - REST models (match internal/api/v3_handlers.go message-ops handlers)

/// Shared body for edit/delete/pin/unpin. `ciphertext` (edit only) is opaque and
/// JSON-encoded as base64 (Go side: messageOpRequest.Ciphertext []byte).
struct MessageOpRequest: Codable, Sendable {
    let conversationId: String
    let ciphertext: Data?

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case ciphertext
    }

    init(conversationId: String, ciphertext: Data? = nil) {
        self.conversationId = conversationId
        self.ciphertext = ciphertext
    }
}

struct MessageEditResult: Codable, Sendable, Equatable {
    let messageId: String
    let edited: Bool
    let retained: Bool
    let version: Int

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case edited, retained, version
    }
}

struct MessageDeleteResult: Codable, Sendable, Equatable {
    let messageId: String
    let deleted: Bool
    let retained: Bool

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case deleted, retained
    }
}

struct MessagePinResult: Codable, Sendable, Equatable {
    let messageId: String
    let pinned: Bool

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case pinned
    }
}

struct DisappearingConfigRequest: Codable, Sendable {
    let ttlSeconds: Int
    let peerDID: String?

    enum CodingKeys: String, CodingKey {
        case ttlSeconds = "ttl_seconds"
        case peerDID = "peer_did"
    }
}

struct DisappearingConfigResult: Codable, Sendable, Equatable {
    let conversationId: String
    let ttlSeconds: Int

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case ttlSeconds = "ttl_seconds"
    }
}

// MARK: - Client protocol (mock in tests)

protocol MessageOpsAPIClient: Sendable {
    @discardableResult
    func editMessage(messageId: String, conversationId: String, ciphertext: Data) async throws -> MessageEditResult
    @discardableResult
    func deleteMessage(messageId: String, conversationId: String) async throws -> MessageDeleteResult
    @discardableResult
    func pinMessage(messageId: String, conversationId: String) async throws -> MessagePinResult
    @discardableResult
    func unpinMessage(messageId: String, conversationId: String) async throws -> MessagePinResult
    @discardableResult
    func setDisappearing(conversationId: String, ttlSeconds: Int, peerDID: String?) async throws -> DisappearingConfigResult
}

#if os(iOS)

enum MessageOpsEndpoint: APIEndpoint {
    case edit(messageId: String)
    case delete(messageId: String)
    case pin(messageId: String)
    case unpin(messageId: String)
    case disappearing(conversationId: String)

    private static func enc(_ s: String) -> String {
        s.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? s
    }

    var path: String {
        switch self {
        case .edit(let id):   return "/v3/messages/\(Self.enc(id))/edit"
        case .delete(let id): return "/v3/messages/\(Self.enc(id))/delete"
        case .pin(let id):    return "/v3/messages/\(Self.enc(id))/pin"
        case .unpin(let id):  return "/v3/messages/\(Self.enc(id))/unpin"
        case .disappearing(let id): return "/v3/conversations/\(Self.enc(id))/disappearing"
        }
    }
}

/// Durable message ops via signed REST. The server fans the change out to the peer
/// over WS; this client only issues the authoritative write.
actor MessageOpsAPI: MessageOpsAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    @discardableResult
    func editMessage(messageId: String, conversationId: String, ciphertext: Data) async throws -> MessageEditResult {
        try await apiClient.post(
            endpoint: MessageOpsEndpoint.edit(messageId: messageId),
            body: MessageOpRequest(conversationId: conversationId, ciphertext: ciphertext)
        )
    }

    @discardableResult
    func deleteMessage(messageId: String, conversationId: String) async throws -> MessageDeleteResult {
        try await apiClient.post(
            endpoint: MessageOpsEndpoint.delete(messageId: messageId),
            body: MessageOpRequest(conversationId: conversationId)
        )
    }

    @discardableResult
    func pinMessage(messageId: String, conversationId: String) async throws -> MessagePinResult {
        try await apiClient.post(
            endpoint: MessageOpsEndpoint.pin(messageId: messageId),
            body: MessageOpRequest(conversationId: conversationId)
        )
    }

    @discardableResult
    func unpinMessage(messageId: String, conversationId: String) async throws -> MessagePinResult {
        try await apiClient.post(
            endpoint: MessageOpsEndpoint.unpin(messageId: messageId),
            body: MessageOpRequest(conversationId: conversationId)
        )
    }

    @discardableResult
    func setDisappearing(conversationId: String, ttlSeconds: Int, peerDID: String?) async throws -> DisappearingConfigResult {
        try await apiClient.post(
            endpoint: MessageOpsEndpoint.disappearing(conversationId: conversationId),
            body: DisappearingConfigRequest(ttlSeconds: ttlSeconds, peerDID: peerDID)
        )
    }
}

#endif
