import Foundation

// MARK: - Signal type constants (match internal/api/ws.go)

enum ConversationSignalType {
    static let typing = "typing"
    static let readReceipt = "read_receipt"
    static let reaction = "reaction"
    /// E2E chat body relay (`internal/api/ws.go` — routed to `to` DID when set).
    static let text = "text"
    // M1 message ops (server fans these out after a REST write).
    static let edit = "edit"
    static let delete = "delete"
    static let pin = "pin"
    static let disappearingConfig = "disappearing_config"
    static let groupKey = "group_key"
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

struct EditPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let messageId: String
    let ciphertext: Data       // opaque; JSON of a TextMessagePayload
    let version: Int?

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case messageId = "message_id"
        case ciphertext
        case version
    }
}

struct DeletePayload: Codable, Sendable, Equatable {
    let conversationId: String
    let messageId: String

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case messageId = "message_id"
    }
}

struct PinPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let messageId: String
    let pinned: Bool

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case messageId = "message_id"
        case pinned
    }
}

struct DisappearingPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let ttlSeconds: Int

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case ttlSeconds = "ttl_seconds"
    }
}

/// Per-member sealed group AES key package (opaque on the wire).
struct GroupKeyPayload: Codable, Sendable, Equatable {
    let groupId: String
    let version: Int
    let encryptedKey: Data
    let distributedBy: String

    enum CodingKeys: String, CodingKey {
        case groupId = "group_id"
        case version
        case encryptedKey = "encrypted_key"
        case distributedBy = "distributed_by"
    }
}

/// Chat payload for relay `type: text` — plaintext (legacy), Kinnami 1:1 envelope, or group AES-GCM blob.
struct TextMessagePayload: Codable, Sendable, Equatable {
    let messageId: String
    let text: String?
    let encrypted: EncryptedMessageWithPublicKey?
    /// AES-256-GCM ciphertext sealed with the group symmetric key (M2).
    let groupCiphertext: Data?
    let groupKeyVersion: Int?

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case text
        case encrypted
        case groupCiphertext = "group_ciphertext"
        case groupKeyVersion = "group_key_version"
    }

    init(
        messageId: String,
        text: String? = nil,
        encrypted: EncryptedMessageWithPublicKey? = nil,
        groupCiphertext: Data? = nil,
        groupKeyVersion: Int? = nil
    ) {
        self.messageId = messageId
        self.text = text
        self.encrypted = encrypted
        self.groupCiphertext = groupCiphertext
        self.groupKeyVersion = groupKeyVersion
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
    static let groupEncryptedPlaceholder = "🔒 Encrypted group message"
}

struct EditSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let messageId: String
    let ciphertext: Data
    let version: Int?
}

struct DeleteSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let messageId: String
}

struct PinSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let messageId: String
    let pinned: Bool
}

struct DisappearingSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let ttlSeconds: Int
}

struct GroupKeySignalEvent: Sendable, Equatable {
    let groupId: String
    let version: Int
    let encryptedKey: Data
    let distributedBy: String
}

enum ConversationSignalEvent: Sendable, Equatable {
    case typing(TypingSignalEvent)
    case readReceipt(ReadReceiptSignalEvent)
    case reaction(ReactionSignalEvent)
    case textMessage(TextMessageSignalEvent)
    case edit(EditSignalEvent)
    case delete(DeleteSignalEvent)
    case pin(PinSignalEvent)
    case disappearing(DisappearingSignalEvent)
    case groupKey(GroupKeySignalEvent)
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

    static func encodeGroupTextMessage(
        conversationId: String,
        payload: TextMessagePayload
    ) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.text,
            to: "",
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
        case ConversationSignalType.edit:
            let envelope = try decoder.decode(WSEnvelope<EditPayload>.self, from: data)
            return .edit(EditSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                messageId: envelope.payload.messageId,
                ciphertext: envelope.payload.ciphertext,
                version: envelope.payload.version
            ))
        case ConversationSignalType.delete:
            let envelope = try decoder.decode(WSEnvelope<DeletePayload>.self, from: data)
            return .delete(DeleteSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                messageId: envelope.payload.messageId
            ))
        case ConversationSignalType.pin:
            let envelope = try decoder.decode(WSEnvelope<PinPayload>.self, from: data)
            return .pin(PinSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                messageId: envelope.payload.messageId,
                pinned: envelope.payload.pinned
            ))
        case ConversationSignalType.disappearingConfig:
            let envelope = try decoder.decode(WSEnvelope<DisappearingPayload>.self, from: data)
            return .disappearing(DisappearingSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                ttlSeconds: envelope.payload.ttlSeconds
            ))
        case ConversationSignalType.groupKey:
            let envelope = try decoder.decode(WSEnvelope<GroupKeyPayload>.self, from: data)
            return .groupKey(GroupKeySignalEvent(
                groupId: envelope.payload.groupId,
                version: envelope.payload.version,
                encryptedKey: envelope.payload.encryptedKey,
                distributedBy: envelope.payload.distributedBy
            ))
        case ConversationSignalType.text:
            if let envelope = try? decoder.decode(WSEnvelope<TextMessagePayload>.self, from: data) {
                let convId = envelope.conversationId ?? ""
                guard !convId.isEmpty else { return nil }
                let msgId = envelope.payload.messageId.isEmpty ? UUID().uuidString : envelope.payload.messageId
                let preview: String
                if let plain = envelope.payload.text, !plain.isEmpty {
                    preview = plain
                } else if envelope.payload.groupCiphertext != nil {
                    preview = TextMessagePayload.groupEncryptedPlaceholder
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
