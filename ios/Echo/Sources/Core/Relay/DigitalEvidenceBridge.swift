import Foundation
import CryptoKit

/// Handles client-side Digital Evidence interactions:
/// - Media fingerprinting before E2E encryption (VIP+ users, optional)
/// - Smart Checkmark rendering for Org-tier messages
/// - Verification URL management for Digital Evidence Explorer
actor DigitalEvidenceBridge {

    private let apiClient: APIClient

    /// Base URL for Digital Evidence Explorer verification pages
    static let verificationBaseURL = "https://digitalevidence.constellationnetwork.io/verify/"

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    // MARK: - Media Fingerprinting

    /// Fingerprint media before E2E encryption (VIP+ users, optional).
    /// Computes SHA-256 of raw media data, submits to backend which
    /// forwards to Constellation's Digital Evidence API.
    func fingerprintMedia(_ mediaData: Data, messageId: String) async throws -> EvidenceResult {
        let hash = SHA256.hash(data: mediaData)
        let hashHex = hash.compactMap { String(format: "%02x", $0) }.joined()

        let request = EvidenceFingerprintRequest(
            contentHash: hashHex,
            messageId: messageId,
            metadata: EvidenceMetadata(type: "media", source: "echo_ios")
        )

        let response: EvidenceFingerprintResponse = try await apiClient.post(
            endpoint: EvidenceEndpoint.submitFingerprint,
            body: request
        )

        return EvidenceResult(
            eventId: response.eventId,
            verificationURL: response.verificationUrl,
            timestamp: response.timestamp
        )
    }

    /// Compute SHA-256 fingerprint of data without submitting
    func computeFingerprint(_ data: Data) -> String {
        let hash = SHA256.hash(data: data)
        return hash.compactMap { String(format: "%02x", $0) }.joined()
    }

    // MARK: - Smart Checkmark

    /// Get the Digital Evidence verification URL for a message.
    /// Returns nil if the message has no evidence event ID.
    func verificationURL(for evidenceEventId: String) -> URL? {
        URL(string: Self.verificationBaseURL + evidenceEventId)
    }

    /// Check if a message qualifies for Smart Checkmark display
    func isSmartCheckmarkEligible(deliveryStatus: DeliveryStatus, evidenceEventId: String?) -> Bool {
        deliveryStatus == .verified && evidenceEventId != nil
    }
}

// MARK: - API Types

struct EvidenceFingerprintRequest: Codable {
    let contentHash: String
    let messageId: String
    let metadata: EvidenceMetadata
}

struct EvidenceMetadata: Codable {
    let type: String
    let source: String
}

struct EvidenceFingerprintResponse: Codable {
    let eventId: String
    let verificationUrl: String
    let timestamp: Date
}

struct EvidenceResult: Codable {
    let eventId: String
    let verificationURL: String
    let timestamp: Date
}

// MARK: - Evidence API Endpoint

enum EvidenceEndpoint: APIEndpoint {
    case submitFingerprint
    case getStatus(eventId: String)

    var path: String {
        switch self {
        case .submitFingerprint:
            return "/api/v1/evidence/fingerprint"
        case .getStatus(let eventId):
            return "/api/v1/evidence/\(eventId)/status"
        }
    }

    var method: HTTPMethod {
        switch self {
        case .submitFingerprint: return .post
        case .getStatus: return .get
        }
    }

    var requiresAuth: Bool { true }
}
