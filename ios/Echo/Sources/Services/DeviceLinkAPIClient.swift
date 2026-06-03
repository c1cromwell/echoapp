#if os(iOS)
import Foundation

/// QR new-device registration (WO-288 / WO-273) via `/identity/devices/*`.
struct DeviceLinkAPIClient: Sendable {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    struct DeviceTokenResponse: Decodable, Sendable {
        let token: String
        let expiresIn: Int

        enum CodingKeys: String, CodingKey {
            case token
            case expiresIn = "expires_in"
        }
    }

    struct DeviceRegisterResponse: Decodable, Sendable {
        let subjectDid: String?
        let status: String?

        enum CodingKeys: String, CodingKey {
            case subjectDid = "subject_did"
            case status
        }
    }

    func issueLinkToken() async throws -> DeviceTokenResponse {
        struct Empty: Encodable {}
        return try await apiClient.post(endpoint: DeviceLinkEndpoint.issueToken, body: Empty())
    }

    func completeLink(token: String, newPublicKeyHex: String, deviceLabel: String) async throws -> DeviceRegisterResponse {
        struct Body: Encodable {
            let token: String
            let newPublicKeyHex: String
            let deviceLabel: String?

            enum CodingKeys: String, CodingKey {
                case token
                case newPublicKeyHex = "new_public_key_hex"
                case deviceLabel = "device_label"
            }
        }
        return try await apiClient.post(
            endpoint: DeviceLinkEndpoint.completeLink,
            body: Body(token: token, newPublicKeyHex: newPublicKeyHex, deviceLabel: deviceLabel)
        )
    }

    /// Builds `echo://link-device?token=` for QR display on the trusted device.
    static func linkQRURL(token: String) -> URL? {
        var components = URLComponents()
        components.scheme = "echo"
        components.host = "link-device"
        components.queryItems = [URLQueryItem(name: "token", value: token)]
        return components.url
    }

    /// Uncompressed P-256 SEC1 hex (`04` + 64 bytes) for device registration.
    static func publicKeyHex(fromBase64PublicKey base64: String) throws -> String {
        guard let data = Data(base64Encoded: base64) else {
            throw DeviceLinkError.invalidPublicKey
        }
        return data.map { String(format: "%02x", $0) }.joined()
    }
}

enum DeviceLinkEndpoint: APIEndpoint {
    case issueToken
    case completeLink

    var path: String {
        switch self {
        case .issueToken:
            return "/identity/devices/token"
        case .completeLink:
            return "/identity/devices"
        }
    }
}

enum DeviceLinkError: LocalizedError {
    case invalidPublicKey
    case notAuthenticated

    var errorDescription: String? {
        switch self {
        case .invalidPublicKey: return "Could not read this device's public key."
        case .notAuthenticated: return "Sign in on your existing device first."
        }
    }
}
#endif
