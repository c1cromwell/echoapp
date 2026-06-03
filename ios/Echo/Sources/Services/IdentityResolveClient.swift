#if os(iOS)
import Foundation

/// `GET /identity/resolve/{did}` — device public keys for E2E message encryption.
struct IdentityResolveClient: Sendable {
    struct Device: Decodable, Sendable {
        let publicKeyHex: String

        enum CodingKeys: String, CodingKey {
            case publicKeyHex = "public_key_hex"
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

    /// First registered device key (P-256 SEC1 hex) for the peer DID.
    func primaryPublicKeyHex(peerDID: String) async throws -> String {
        let encoded = peerDID.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? peerDID
        let response: Response = try await apiClient.get(
            endpoint: IdentityEndpoint.resolve(did: encoded)
        )
        guard let hex = response.devices.first?.publicKeyHex, !hex.isEmpty else {
            throw IdentityResolveError.noDevices
        }
        return hex
    }
}

enum IdentityEndpoint: APIEndpoint {
    case resolve(did: String)

    var path: String { "/identity/resolve/\(did)" }
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
