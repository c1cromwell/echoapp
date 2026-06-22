#if os(iOS)
import Foundation

/// Decrypts inbound WS text payloads for inbox preview + thread persistence.
enum InboundTextMessageResolver {
    struct ResolvedBody: Sendable {
        let body: String
        let preview: String
    }

    static func resolveBody(for event: TextMessageSignalEvent) async -> ResolvedBody {
        var body = event.text
        if let wire = event.wirePayload, wire.encrypted != nil {
            let client = await DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
            let crypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: client))
            // Sender authentication: reject a message whose signature is present but does
            // not verify against the claimed sender — a tampered or spoofed relay payload.
            // (`.unsigned` legacy messages still flow through until signing is universal.)
            if await crypto.senderVerification(for: wire, expectedSenderDID: event.peerDID) == .invalid {
                return ResolvedBody(body: TextMessagePayload.encryptedPlaceholder, preview: "Unverified sender")
            }
            if let decrypted = try? await crypto.decryptPayload(wire) {
                if let mediaWire = MediaMessageService.parseWire(from: decrypted) {
                    body = TextMessagePayload.mediaPlaceholder(for: mediaWire.media)
                } else {
                    body = decrypted
                }
            } else {
                body = TextMessagePayload.encryptedPlaceholder
            }
        } else if let media = event.wirePayload?.media {
            body = TextMessagePayload.mediaPlaceholder(for: media)
        }
        let preview = body == TextMessagePayload.encryptedPlaceholder ? "Encrypted message" : body
        return ResolvedBody(body: body, preview: preview)
    }
}
#endif
