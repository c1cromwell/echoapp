import Foundation

// MARK: - Signal type constants (match internal/api/ws.go)

enum ConversationSignalType {
    static let typing = "typing"
    static let readReceipt = "read_receipt"
    static let reaction = "reaction"
    /// E2E chat body relay (`internal/api/ws.go` — routed to `to` DID when set).
    static let text = "text"
}

enum TypingState: String, Codable, Sendable {
    case start
    case stop
}

// MARK: - Wire envelope (server WSMessage shape)

struct WSEnvelope<Payload: Codable>: Codable, Sendable where Payload: Sendable {
    var type: String
    var to: String
    var from: String?
    var conversationId: String?
    var payload: Payload
    var timestamp: String?

    enum CodingKeys: String, CodingKey {
        case type, to, from
        case conversationId = "conversation_id"
        case payload, timestamp
    }
}

struct TypingPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let state: String

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case state
    }

    init(conversationId: String, state: TypingState) {
        self.conversationId = conversationId
        self.state = state.rawValue
    }
}

struct ReadReceiptPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let messageIds: [String]
    let readAt: String

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case messageIds = "message_ids"
        case readAt = "read_at"
    }
}

struct ReactionPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let messageId: String
    let emoji: String

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case messageId = "message_id"
        case emoji
    }
}

/// Chat payload for relay `type: text` — plaintext (legacy) or Kinnami-encrypted envelope.
struct TextMessagePayload: Codable, Sendable, Equatable {
    let messageId: String
    let text: String?
    let encrypted: EncryptedMessageWithPublicKey?

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case text
        case encrypted
    }

    init(messageId: String, text: String? = nil, encrypted: EncryptedMessageWithPublicKey? = nil) {
        self.messageId = messageId
        self.text = text
        self.encrypted = encrypted
    }
}

// MARK: - Inbound events (decoded from envelope)

struct TypingSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let state: TypingState
}

struct ReadReceiptSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let messageIds: [String]
    let readAt: String
}

struct ReactionSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let messageId: String
    let emoji: String
}

struct TextMessageSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let messageId: String
    /// Preview or decrypted body for UI.
    let text: String
    /// Raw wire payload when decryption should run off the main decode path.
    let wirePayload: TextMessagePayload?

    init(
        conversationId: String,
        peerDID: String,
        messageId: String,
        text: String,
        wirePayload: TextMessagePayload? = nil
    ) {
        self.conversationId = conversationId
        self.peerDID = peerDID
        self.messageId = messageId
        self.text = text
        self.wirePayload = wirePayload
    }
}

extension TextMessagePayload {
    static let encryptedPlaceholder = "🔒 Encrypted message"
}

enum ConversationSignalEvent: Sendable, Equatable {
    case typing(TypingSignalEvent)
    case readReceipt(ReadReceiptSignalEvent)
    case reaction(ReactionSignalEvent)
    case textMessage(TextMessageSignalEvent)
}

// MARK: - Partial header for routing inbound JSON

struct WSEnvelopeHeader: Codable {
    let type: String
    let to: String?
    let from: String?
    let conversationId: String?

    enum CodingKeys: String, CodingKey {
        case type, to, from
        case conversationId = "conversation_id"
    }
}

// MARK: - Encode / decode helpers

enum ConversationSignalCodec {
    private static let encoder: JSONEncoder = {
        let e = JSONEncoder()
        return e
    }()

    private static let decoder: JSONDecoder = {
        let d = JSONDecoder()
        return d
    }()

    static func encodeTyping(to peerDID: String, conversationId: String, state: TypingState) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.typing,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: TypingPayload(conversationId: conversationId, state: state),
            timestamp: isoTimestamp()
        )
        let data = try encoder.encode(envelope)
        guard let text = String(data: data, encoding: .utf8) else {
            throw ConversationSignalError.encodingFailed
        }
        return text
    }

    static func encodeReadReceipt(
        to peerDID: String,
        conversationId: String,
        messageIds: [String],
        readAt: String
    ) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.readReceipt,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: ReadReceiptPayload(conversationId: conversationId, messageIds: messageIds, readAt: readAt),
            timestamp: isoTimestamp()
        )
        let data = try encoder.encode(envelope)
        guard let text = String(data: data, encoding: .utf8) else {
            throw ConversationSignalError.encodingFailed
        }
        return text
    }

    static func encodeTextMessage(
        to peerDID: String,
        conversationId: String,
        payload: TextMessagePayload
    ) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.text,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: payload,
            timestamp: isoTimestamp()
        )
        let data = try encoder.encode(envelope)
        guard let json = String(data: data, encoding: .utf8) else {
            throw ConversationSignalError.encodingFailed
        }
        return json
    }

    static func encodeReaction(
        to peerDID: String,
        conversationId: String,
        messageId: String,
        emoji: String
    ) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.reaction,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: ReactionPayload(conversationId: conversationId, messageId: messageId, emoji: emoji),
            timestamp: isoTimestamp()
        )
        let data = try encoder.encode(envelope)
        guard let text = String(data: data, encoding: .utf8) else {
            throw ConversationSignalError.encodingFailed
        }
        return text
    }

    /// Parses inbound WebSocket text; returns nil for unknown types (forward-compat).
    static func decodeEvent(from text: String) throws -> ConversationSignalEvent? {
        let data = Data(text.utf8)
        let header = try decoder.decode(WSEnvelopeHeader.self, from: data)
        let peerDID = header.from ?? ""

        switch header.type {
        case ConversationSignalType.typing:
            let envelope = try decoder.decode(WSEnvelope<TypingPayload>.self, from: data)
            guard let state = TypingState(rawValue: envelope.payload.state) else { return nil }
            return .typing(TypingSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                state: state
            ))
        case ConversationSignalType.readReceipt:
            let envelope = try decoder.decode(WSEnvelope<ReadReceiptPayload>.self, from: data)
            return .readReceipt(ReadReceiptSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                messageIds: envelope.payload.messageIds,
                readAt: envelope.payload.readAt
            ))
        case ConversationSignalType.reaction:
            let envelope = try decoder.decode(WSEnvelope<ReactionPayload>.self, from: data)
            return .reaction(ReactionSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                messageId: envelope.payload.messageId,
                emoji: envelope.payload.emoji
            ))
        case ConversationSignalType.text:
            if let envelope = try? decoder.decode(WSEnvelope<TextMessagePayload>.self, from: data) {
                let convId = envelope.conversationId ?? ""
                guard !convId.isEmpty else { return nil }
                let msgId = envelope.payload.messageId.isEmpty ? UUID().uuidString : envelope.payload.messageId
                let preview: String
                if let plain = envelope.payload.text, !plain.isEmpty {
                    preview = plain
                } else if envelope.payload.encrypted != nil {
                    preview = TextMessagePayload.encryptedPlaceholder
                } else {
                    return nil
                }
                return .textMessage(TextMessageSignalEvent(
                    conversationId: convId,
                    peerDID: peerDID,
                    messageId: msgId,
                    text: preview,
                    wirePayload: envelope.payload
                ))
            }
            if let envelope = try? decoder.decode(WSEnvelope<String>.self, from: data) {
                let convId = envelope.conversationId ?? ""
                guard !convId.isEmpty else { return nil }
                return .textMessage(TextMessageSignalEvent(
                    conversationId: convId,
                    peerDID: peerDID,
                    messageId: UUID().uuidString,
                    text: envelope.payload
                ))
            }
            return nil
        default:
            return nil
        }
    }

    static func isoTimestamp(date: Date = Date()) -> String {
        ISO8601DateFormatter().string(from: date)
    }
}

enum ConversationSignalError: Error, Equatable {
    case encodingFailed
    case notConnected
}
