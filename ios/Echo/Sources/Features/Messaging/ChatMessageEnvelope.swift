import Foundation

/// JSON envelope encrypted inside 1:1 Kinnami ciphertext (M1 reply/forward).
/// Backward compatible: legacy messages used a plain string body.
struct ChatMessageEnvelope: Codable, Sendable, Equatable {
    static let formatVersion = "v1"

    var format: String
    var body: String
    var replyToMessageId: String?
    var replyPreview: String?
    var forwardedFromMessageId: String?
    var forwardedFromConversationId: String?

    enum CodingKeys: String, CodingKey {
        case format
        case body
        case replyToMessageId = "reply_to_message_id"
        case replyPreview = "reply_preview"
        case forwardedFromMessageId = "forwarded_from_message_id"
        case forwardedFromConversationId = "forwarded_from_conversation_id"
    }

    init(
        body: String,
        replyToMessageId: String? = nil,
        replyPreview: String? = nil,
        forwardedFromMessageId: String? = nil,
        forwardedFromConversationId: String? = nil
    ) {
        self.format = Self.formatVersion
        self.body = body
        self.replyToMessageId = replyToMessageId
        self.replyPreview = replyPreview
        self.forwardedFromMessageId = forwardedFromMessageId
        self.forwardedFromConversationId = forwardedFromConversationId
    }

    /// Plain string or JSON envelope → display body + optional thread metadata.
    static func parseDecrypted(_ decrypted: String) -> (body: String, envelope: ChatMessageEnvelope?) {
        let trimmed = decrypted.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.hasPrefix("{"),
              let data = trimmed.data(using: .utf8),
              let env = try? JSONDecoder().decode(ChatMessageEnvelope.self, from: data),
              env.format == formatVersion else {
            return (decrypted, nil)
        }
        return (env.body, env)
    }

    func serialized() throws -> String {
        let data = try JSONEncoder().encode(self)
        guard let s = String(data: data, encoding: .utf8) else {
            throw ChatMessageEnvelopeError.encodingFailed
        }
        return s
    }
}

enum ChatMessageEnvelopeError: Error {
    case encodingFailed
}
