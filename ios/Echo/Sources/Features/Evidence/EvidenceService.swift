// Features/Evidence/DigitalEvidenceBridge.swift
// SHA-256 fingerprinting + Constellation Digital Evidence API submission

import CryptoKit
import Foundation

// MARK: - Evidence Result

struct EvidenceResult: Codable, Equatable {
    let eventId: String
    let verificationUrl: String
    let timestamp: Date

    enum CodingKeys: String, CodingKey {
        case eventId = "event_id"
        case verificationUrl = "verification_url"
        case timestamp
    }
}

// MARK: - Evidence API Protocol

protocol EvidenceAPIProtocol {
    func submitFingerprint(contentHash: String, sourceType: String, messageId: String) async throws -> EvidenceResult
    func verifyFingerprint(eventId: String) async throws -> EvidenceVerificationStatus
}

// MARK: - Verification Status

struct EvidenceVerificationStatus: Codable, Equatable {
    let eventId: String
    let status: String // "verified", "pending", "failed"
    let explorerUrl: String?

    enum CodingKeys: String, CodingKey {
        case eventId = "event_id"
        case status
        case explorerUrl = "explorer_url"
    }
}

// MARK: - Digital Evidence Bridge

actor DigitalEvidenceBridge {
    private let api: EvidenceAPIProtocol

    init(api: EvidenceAPIProtocol) {
        self.api = api
    }

    /// Fingerprint media before E2E encryption. Returns Event ID for message metadata.
    func fingerprintMedia(_ data: Data, messageId: String) async throws -> EvidenceResult {
        let hash = SHA256.hash(data: data)
        let hashHex = hash.map { String(format: "%02x", $0) }.joined()

        return try await api.submitFingerprint(
            contentHash: hashHex,
            sourceType: "media",
            messageId: messageId
        )
    }

    /// Fingerprint text message content.
    func fingerprintMessage(_ content: String, messageId: String) async throws -> EvidenceResult {
        let data = Data(content.utf8)
        let hash = SHA256.hash(data: data)
        let hashHex = hash.map { String(format: "%02x", $0) }.joined()

        return try await api.submitFingerprint(
            contentHash: hashHex,
            sourceType: "message",
            messageId: messageId
        )
    }

    /// Get verification URL for a verified message.
    func verificationURL(eventId: String) -> URL? {
        URL(string: "https://digitalevidence.constellationnetwork.io/verify/\(eventId)")
    }

    /// Check on-chain verification status.
    func checkVerification(eventId: String) async throws -> EvidenceVerificationStatus {
        return try await api.verifyFingerprint(eventId: eventId)
    }
}

// MARK: - Mock for Testing

#if DEBUG
final class MockEvidenceAPI: EvidenceAPIProtocol {
    var shouldError = false

    func submitFingerprint(contentHash: String, sourceType: String, messageId: String) async throws -> EvidenceResult {
        if shouldError { throw NSError(domain: "MockEvidence", code: 1) }
        return EvidenceResult(
            eventId: "evt_\(messageId)",
            verificationUrl: "https://digitalevidence.constellationnetwork.io/verify/evt_\(messageId)",
            timestamp: Date()
        )
    }

    func verifyFingerprint(eventId: String) async throws -> EvidenceVerificationStatus {
        return EvidenceVerificationStatus(
            eventId: eventId,
            status: "verified",
            explorerUrl: "https://digitalevidence.constellationnetwork.io/verify/\(eventId)"
        )
    }
}
#endif
