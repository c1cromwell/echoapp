#if os(iOS)
import Foundation

/// `GET /identity/resolve/{did}` — device public keys for E2E message encryption.
struct IdentityResolveClient: Sendable {
    struct Device: Decodable, Sendable {
        let publicKeyHex: String
        let deviceLabel: String?

        enum CodingKeys: String, CodingKey {
            case publicKeyHex = "public_key_hex"
            case deviceLabel = "device_label"
        }
    }

    struct Response: Decodable, Sendable {
        let did: String
        let devices: [Device]
    }

    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    /// Public key (P-256 SEC1 hex) a peer should be encrypted to. Prefers the dedicated
    /// messaging key-agreement device (`MessagingAgreementKey.deviceLabel`); falls back
    /// to the first registered key for peers that predate Option B.
    func primaryPublicKeyHex(peerDID: String) async throws -> String {
        let encoded = peerDID.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? peerDID
        let response: Response = try await apiClient.get(
            endpoint: IdentityEndpoint.resolveDID(did: encoded)
        )
        if let agreement = response.devices.first(where: { $0.deviceLabel == MessagingAgreementKey.deviceLabel }),
           !agreement.publicKeyHex.isEmpty {
            return agreement.publicKeyHex
        }
        guard let hex = response.devices.first?.publicKeyHex, !hex.isEmpty else {
            throw IdentityResolveError.noDevices
        }
        return hex
    }
}

enum IdentityResolveError: LocalizedError {
    case noDevices
    case invalidKeyMaterial

    var errorDescription: String? {
        switch self {
        case .noDevices: return "Peer identity has no registered encryption keys."
        case .invalidKeyMaterial: return "Invalid public key from identity registry."
        }
    }
}
#endif
