import Foundation

// MARK: - Signal type constants (match internal/api/ws.go)

enum ConversationSignalType {
    static let typing = "typing"
    static let readReceipt = "read_receipt"
    static let reaction = "reaction"
    /// E2E chat body relay (`internal/api/ws.go` — routed to `to` DID when set).
    static let text = "text"
    /// Sealed 1:1 body relay — sender DID hidden from recipient wire metadata (WO-219).
    static let sealedText = "sealed_text"
    // M1 message ops (server fans these out after a REST write).
    static let edit = "edit"
    static let delete = "delete"
    static let pin = "pin"
    static let disappearingConfig = "disappearing_config"
    static let groupKey = "group_key"
    static let poll = "poll"
    static let screenshotAlert = "screenshot_alert"
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

/// Outer sealed-sender body (WO-219). Inner sender DID is encrypted in `ciphertext`.
struct SealedTextPayload: Codable, Sendable, Equatable {
    let deliveryToken: String
    let ciphertext: Data

    enum CodingKeys: String, CodingKey {
        case deliveryToken = "delivery_token"
        case ciphertext
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

struct PollPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let pollId: String
    let action: String
    let optionId: String?
    let ciphertext: Data?

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case pollId = "poll_id"
        case action
        case optionId = "option_id"
        case ciphertext
    }
}

struct ScreenshotAlertPayload: Codable, Sendable, Equatable {
    let conversationId: String
    let alertedAt: String

    enum CodingKeys: String, CodingKey {
        case conversationId = "conversation_id"
        case alertedAt = "alerted_at"
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
    /// Client-encrypted media relay reference (M5). Cleartext only in legacy/dev paths.
    let media: MediaAttachmentRef?
    /// Sender authentication (1:1): the sender's DID and a P-256 signature over
    /// `(messageId, ciphertext)`. Optional for backward compatibility; see `MessageSenderAuth`.
    let senderDID: String?
    let signature: Data?

    enum CodingKeys: String, CodingKey {
        case messageId = "message_id"
        case text
        case encrypted
        case groupCiphertext = "group_ciphertext"
        case groupKeyVersion = "group_key_version"
        case media
        case senderDID = "sender_did"
        case signature
    }

    init(
        messageId: String,
        text: String? = nil,
        encrypted: EncryptedMessageWithPublicKey? = nil,
        groupCiphertext: Data? = nil,
        groupKeyVersion: Int? = nil,
        media: MediaAttachmentRef? = nil,
        senderDID: String? = nil,
        signature: Data? = nil
    ) {
        self.messageId = messageId
        self.text = text
        self.encrypted = encrypted
        self.groupCiphertext = groupCiphertext
        self.groupKeyVersion = groupKeyVersion
        self.media = media
        self.senderDID = senderDID
        self.signature = signature
    }
}

/// Opaque media blob reference carried inside encrypted chat payloads (M5).
struct MediaAttachmentRef: Codable, Sendable, Equatable {
    let fileId: String
    let mimeType: String
    let mediaKind: String
    let byteSize: Int
    let chunkCount: Int
    let caption: String?
    /// Precomputed playback bars (amplitude only, not speech content).
    let waveformBars: [Float]?

    enum CodingKeys: String, CodingKey {
        case fileId = "file_id"
        case mimeType = "mime_type"
        case mediaKind = "media_kind"
        case byteSize = "byte_size"
        case chunkCount = "chunk_count"
        case caption
        case waveformBars = "waveform_bars"
    }

    init(
        fileId: String,
        mimeType: String,
        mediaKind: String,
        byteSize: Int,
        chunkCount: Int,
        caption: String?,
        waveformBars: [Float]? = nil
    ) {
        self.fileId = fileId
        self.mimeType = mimeType
        self.mediaKind = mediaKind
        self.byteSize = byteSize
        self.chunkCount = chunkCount
        self.caption = caption
        self.waveformBars = waveformBars
    }
}

enum MediaKind: String, Sendable {
    case image
    case video
    case audio
    case file
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
    /// Sealed-sender outer payload before unwrap (WO-219).
    let sealedPayload: SealedTextPayload?

    init(
        conversationId: String,
        peerDID: String,
        messageId: String,
        text: String,
        wirePayload: TextMessagePayload? = nil,
        sealedPayload: SealedTextPayload? = nil
    ) {
        self.conversationId = conversationId
        self.peerDID = peerDID
        self.messageId = messageId
        self.text = text
        self.wirePayload = wirePayload
        self.sealedPayload = sealedPayload
    }
}

extension TextMessagePayload {
    static let encryptedPlaceholder = "🔒 Encrypted message"
    static let groupEncryptedPlaceholder = "🔒 Encrypted group message"

    static func mediaPlaceholder(for ref: MediaAttachmentRef) -> String {
        if let caption = ref.caption?.trimmingCharacters(in: .whitespacesAndNewlines), !caption.isEmpty {
            return caption
        }
        switch ref.mediaKind {
        case MediaKind.audio.rawValue: return "🎤 Voice note"
        case MediaKind.image.rawValue: return "📷 Photo"
        case MediaKind.video.rawValue: return "🎬 Video"
        default: return "📎 Attachment"
        }
    }
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

struct PollSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let pollId: String
    let action: String
    let optionId: String?
    let ciphertext: Data?
}

struct ScreenshotAlertSignalEvent: Sendable, Equatable {
    let conversationId: String
    let peerDID: String
    let alertedAt: String
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
    case poll(PollSignalEvent)
    case screenshotAlert(ScreenshotAlertSignalEvent)
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

    static func encodePoll(
        to peerDID: String,
        conversationId: String,
        payload: PollPayload
    ) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.poll,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: payload,
            timestamp: isoTimestamp()
        )
        let data = try encoder.encode(envelope)
        guard let text = String(data: data, encoding: .utf8) else {
            throw ConversationSignalError.encodingFailed
        }
        return text
    }

    static func encodeScreenshotAlert(to peerDID: String, conversationId: String) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.screenshotAlert,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: ScreenshotAlertPayload(
                conversationId: conversationId,
                alertedAt: isoTimestamp()
            ),
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

    static func encodeSealedTextMessage(
        to peerDID: String,
        conversationId: String,
        payload: SealedTextPayload
    ) throws -> String {
        let envelope = WSEnvelope(
            type: ConversationSignalType.sealedText,
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
        case ConversationSignalType.poll:
            let envelope = try decoder.decode(WSEnvelope<PollPayload>.self, from: data)
            return .poll(PollSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                pollId: envelope.payload.pollId,
                action: envelope.payload.action,
                optionId: envelope.payload.optionId,
                ciphertext: envelope.payload.ciphertext
            ))
        case ConversationSignalType.screenshotAlert:
            let envelope = try decoder.decode(WSEnvelope<ScreenshotAlertPayload>.self, from: data)
            return .screenshotAlert(ScreenshotAlertSignalEvent(
                conversationId: envelope.payload.conversationId,
                peerDID: peerDID,
                alertedAt: envelope.payload.alertedAt
            ))
        case ConversationSignalType.text:
            if let envelope = try? decoder.decode(WSEnvelope<TextMessagePayload>.self, from: data) {
                let convId = envelope.conversationId ?? ""
                guard !convId.isEmpty else { return nil }
                let msgId = envelope.payload.messageId.isEmpty ? UUID().uuidString : envelope.payload.messageId
                let preview: String
                if let plain = envelope.payload.text, !plain.isEmpty {
                    preview = plain
                } else if let media = envelope.payload.media {
                    preview = TextMessagePayload.mediaPlaceholder(for: media)
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
        case ConversationSignalType.sealedText:
            if let envelope = try? decoder.decode(WSEnvelope<SealedTextPayload>.self, from: data) {
                let convId = envelope.conversationId ?? ""
                guard !convId.isEmpty else { return nil }
                return .textMessage(TextMessageSignalEvent(
                    conversationId: convId,
                    peerDID: peerDID,
                    messageId: UUID().uuidString,
                    text: TextMessagePayload.encryptedPlaceholder,
                    wirePayload: nil,
                    sealedPayload: envelope.payload
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
