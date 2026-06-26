#if os(iOS)
import Foundation

private struct SealedSenderInner: Codable, Sendable {
    let senderDID: String
    let message: TextMessagePayload

    enum CodingKeys: String, CodingKey {
        case senderDID = "sender_did"
        case message
    }
}

struct SealedDeliveryTokenResponse: Codable, Sendable {
    let deliveryToken: String
    let expiresIn: Int?

    enum CodingKeys: String, CodingKey {
        case deliveryToken = "delivery_token"
        case expiresIn = "expires_in"
    }
}

enum SealedSenderEndpoint: APIEndpoint {
    case deliveryToken

    var path: String { "/v3/messages/sealed-token" }
}

/// Wraps 1:1 chat payloads so relay metadata omits sender DID (WO-219).
actor SealedSenderService {
    private let identityResolve: IdentityResolveClient
    private let apiClient: APIClient

    init(identityResolve: IdentityResolveClient, apiClient: APIClient) {
        self.identityResolve = identityResolve
        self.apiClient = apiClient
    }

    func wrap(
        payload: TextMessagePayload,
        recipientDID: String,
        senderDID: String
    ) async throws -> SealedTextPayload {
        let inner = SealedSenderInner(senderDID: senderDID, message: payload)
        let innerData = try JSONEncoder().encode(inner)
        guard let innerText = String(data: innerData, encoding: .utf8) else {
            throw SealedSenderError.encodingFailed
        }
        let hex = try await identityResolve.primaryPublicKeyHex(peerDID: recipientDID)
        let pubData = try TextMessageCrypto.dataFromPublicKeyHex(hex)
        let kinnami = KinnamiEncryption()
        let encrypted = try await kinnami.encryptWithKeyAgreement(
            plaintext: innerText,
            recipientPublicKeyData: pubData
        )
        let ciphertext = try JSONEncoder().encode(encrypted)
        let token = try await fetchDeliveryToken()
        return SealedTextPayload(deliveryToken: token, ciphertext: ciphertext)
    }

    func unwrap(_ sealed: SealedTextPayload) async throws -> (senderDID: String, payload: TextMessagePayload) {
        let encrypted = try JSONDecoder().decode(EncryptedMessageWithPublicKey.self, from: sealed.ciphertext)
        let privateKey = try await TextMessageCrypto.loadAgreementPrivateKey()
        let kinnami = KinnamiEncryption()
        let plain = try await kinnami.decryptWithKeyAgreement(
            encryptedMessage: encrypted,
            ourPrivateKey: privateKey
        )
        guard let data = plain.data(using: .utf8) else {
            throw SealedSenderError.decodingFailed
        }
        let inner = try JSONDecoder().decode(SealedSenderInner.self, from: data)
        return (inner.senderDID, inner.message)
    }

    private func fetchDeliveryToken() async throws -> String {
        struct EmptyBody: Encodable {}
        let response: SealedDeliveryTokenResponse = try await apiClient.post(
            endpoint: SealedSenderEndpoint.deliveryToken,
            body: EmptyBody()
        )
        guard !response.deliveryToken.isEmpty else { throw SealedSenderError.missingToken }
        return response.deliveryToken
    }
}

enum SealedSenderError: LocalizedError {
    case encodingFailed
    case decodingFailed
    case missingToken

    var errorDescription: String? {
        switch self {
        case .encodingFailed: return "Could not encode sealed message."
        case .decodingFailed: return "Could not decode sealed message."
        case .missingToken: return "Sealed sender token unavailable."
        }
    }
}
#endif
