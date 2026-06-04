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
            if let decrypted = try? await crypto.decryptPayload(wire) {
                body = decrypted
            } else {
                body = TextMessagePayload.encryptedPlaceholder
            }
        }
        let preview = body == TextMessagePayload.encryptedPlaceholder ? "Encrypted message" : body
        return ResolvedBody(body: body, preview: preview)
    }
}
#endif
