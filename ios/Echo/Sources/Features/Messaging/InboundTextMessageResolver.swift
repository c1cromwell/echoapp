#if os(iOS)
import Foundation

/// Decrypts inbound WS text payloads for inbox preview + thread persistence.
enum InboundTextMessageResolver {
    struct ResolvedBody: Sendable {
        let senderDID: String
        let body: String
        let preview: String
        let replyToMessageId: String?
        let replyPreview: String?
        let forwardedFromMessageId: String?
        let forwardedFromConversationId: String?
    }

    static func resolveBody(for event: TextMessageSignalEvent) async -> ResolvedBody {
        var peerDID = event.peerDID
        var wirePayload = event.wirePayload
        if let sealed = event.sealedPayload {
            let client = await DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
            let sealedService = SealedSenderService(
                identityResolve: IdentityResolveClient(apiClient: client),
                apiClient: client
            )
            if let unwrapped = try? await sealedService.unwrap(sealed) {
                peerDID = unwrapped.senderDID
                wirePayload = unwrapped.payload
            } else {
                return ResolvedBody(
                    senderDID: peerDID,
                    body: TextMessagePayload.encryptedPlaceholder,
                    preview: "Encrypted message",
                    replyToMessageId: nil,
                    replyPreview: nil,
                    forwardedFromMessageId: nil,
                    forwardedFromConversationId: nil
                )
            }
        }

        var body = event.text
        var replyToMessageId: String?
        var replyPreview: String?
        var forwardedFromMessageId: String?
        var forwardedFromConversationId: String?
        if let wire = wirePayload, wire.encrypted != nil {
            let client = await DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
            let crypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: client))
            if await crypto.senderVerification(for: wire, expectedSenderDID: peerDID) == .invalid {
                return ResolvedBody(
                    senderDID: peerDID,
                    body: TextMessagePayload.encryptedPlaceholder,
                    preview: "Unverified sender",
                    replyToMessageId: nil,
                    replyPreview: nil,
                    forwardedFromMessageId: nil,
                    forwardedFromConversationId: nil
                )
            }
            if let decrypted = try? await crypto.decryptPayload(wire) {
                let parsed = ChatMessageEnvelope.parseDecrypted(decrypted)
                replyToMessageId = parsed.envelope?.replyToMessageId
                replyPreview = parsed.envelope?.replyPreview
                forwardedFromMessageId = parsed.envelope?.forwardedFromMessageId
                forwardedFromConversationId = parsed.envelope?.forwardedFromConversationId
                if let mediaWire = MediaMessageService.parseWire(from: parsed.body) {
                    body = TextMessagePayload.mediaPlaceholder(for: mediaWire.media)
                } else {
                    body = parsed.body
                }
            } else {
                body = TextMessagePayload.encryptedPlaceholder
            }
        } else if let media = wirePayload?.media {
            body = TextMessagePayload.mediaPlaceholder(for: media)
        }
        let preview = body == TextMessagePayload.encryptedPlaceholder ? "Encrypted message" : body
        return ResolvedBody(
            senderDID: peerDID,
            body: body,
            preview: preview,
            replyToMessageId: replyToMessageId,
            replyPreview: replyPreview,
            forwardedFromMessageId: forwardedFromMessageId,
            forwardedFromConversationId: forwardedFromConversationId
        )
    }
}
#endif
