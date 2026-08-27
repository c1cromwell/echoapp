import Foundation

/// Opaque body the gateway stores. Encrypt before POST so T0 plaintext never hits `/v3/messages/schedule`.
struct ScheduledRemoteEnvelope: Codable, Equatable, Sendable {
    let peerDID: String
    let body: String

    enum CodingKeys: String, CodingKey {
        case peerDID = "peer_did"
        case body
    }
}

struct ScheduledCreateRequest: Encodable, Equatable, Sendable {
    let conversationId: String
    let content: String
    let contentType: String
    let scheduledAt: String
    let timezone: String
    let silent: Bool

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case content
        case contentType = "content_type"
        case scheduledAt = "scheduled_at"
        case timezone
        case silent
    }
}

struct ScheduledMessageDTO: Codable, Equatable, Sendable {
    let id: String
    let conversationId: String
    let scheduledAt: String
    let timezone: String?
    let status: String
    let silent: Bool
    let contentType: String?
    let content: Data?

    enum CodingKeys: String, CodingKey {
        case id
        case conversationId = "conversation_id"
        case scheduledAt = "scheduled_at"
        case timezone
        case status
        case silent
        case contentType = "content_type"
        case content
    }
}

struct ScheduledListResponse: Codable, Equatable, Sendable {
    let messages: [ScheduledMessageDTO]
}

struct ScheduledCancelResponse: Codable, Equatable, Sendable {
    let id: String
    let status: String
}

enum ScheduledMessageRemoteCodec {
    static let collectionPath = "/v3/messages/schedule"

    static func itemPath(id: String) -> String {
        let enc = id.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? id
        return "/v3/messages/schedule/\(enc)"
    }

    static func rfc3339(_ date: Date) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter.string(from: date)
    }

    static func envelopeJSON(peerDID: String, plaintext: String) throws -> Data {
        try JSONEncoder().encode(ScheduledRemoteEnvelope(peerDID: peerDID, body: plaintext))
    }

    static func parseEnvelope(_ data: Data) throws -> ScheduledRemoteEnvelope {
        try JSONDecoder().decode(ScheduledRemoteEnvelope.self, from: data)
    }

    static func createRequest(
        conversationId: String,
        ciphertext: String,
        fireAt: Date,
        timezone: String,
        silent: Bool
    ) -> ScheduledCreateRequest {
        ScheduledCreateRequest(
            conversationId: conversationId,
            content: ciphertext,
            contentType: "text",
            scheduledAt: rfc3339(fireAt),
            timezone: timezone,
            silent: silent
        )
    }
}

#if os(iOS)
enum ScheduledMessageEndpoint: APIEndpoint {
    case collection
    case item(id: String)

    var path: String {
        switch self {
        case .collection:
            return ScheduledMessageRemoteCodec.collectionPath
        case .item(let id):
            return ScheduledMessageRemoteCodec.itemPath(id: id)
        }
    }
}

/// Passkey-signed scheduled-message client (WO-338). Gateway stores opaque content only.
actor ScheduledMessageAPI {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func create(
        conversationId: String,
        peerDID: String,
        plaintext: String,
        fireAt: Date,
        timezone: String = TimeZone.current.identifier,
        silent: Bool = false
    ) async throws -> ScheduledMessageDTO {
        let envelope = try ScheduledMessageRemoteCodec.envelopeJSON(peerDID: peerDID, plaintext: plaintext)
        let sealed = try ScheduledMessageCrypto.encryptBlob(envelope)
        let request = ScheduledMessageRemoteCodec.createRequest(
            conversationId: conversationId,
            ciphertext: sealed.base64EncodedString(),
            fireAt: fireAt,
            timezone: timezone,
            silent: silent
        )
        return try await apiClient.post(endpoint: ScheduledMessageEndpoint.collection, body: request)
    }

    func list() async throws -> [ScheduledMessageDTO] {
        let response: ScheduledListResponse = try await apiClient.get(endpoint: ScheduledMessageEndpoint.collection)
        return response.messages
    }

    func get(id: String) async throws -> ScheduledMessageDTO {
        try await apiClient.get(endpoint: ScheduledMessageEndpoint.item(id: id))
    }

    func cancel(id: String) async throws {
        let _: ScheduledCancelResponse = try await apiClient.delete(endpoint: ScheduledMessageEndpoint.item(id: id))
    }

    func decryptEnvelope(from dto: ScheduledMessageDTO) throws -> ScheduledRemoteEnvelope? {
        guard let raw = dto.content, !raw.isEmpty else { return nil }
        let base64 = String(data: raw, encoding: .utf8) ?? raw.base64EncodedString()
        guard let sealed = Data(base64Encoded: base64) else { return nil }
        let envelope = try ScheduledMessageCrypto.decryptBlob(sealed)
        return try ScheduledMessageRemoteCodec.parseEnvelope(envelope)
    }
}
#endif
