// Features/Evidence/EvidenceService.swift
// Additional evidence types supplementing Core/Relay/DigitalEvidenceBridge

import Foundation

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

// MARK: - Mock for Testing

#if DEBUG
final class MockEvidenceAPI: EvidenceAPIProtocol {
    var shouldError = false

    func submitFingerprint(contentHash: String, sourceType: String, messageId: String) async throws -> EvidenceResult {
        if shouldError { throw NSError(domain: "MockEvidence", code: 1) }
        return EvidenceResult(
            eventId: "evt_\(messageId)",
            verificationURL: "https://digitalevidence.constellationnetwork.io/verify/evt_\(messageId)",
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
