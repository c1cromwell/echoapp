#if os(iOS)
import Foundation

/// POST /v3/contacts/psi — OPRF-PSI contact discovery (WO-221 / WO-220).
struct ContactDiscoveryAPIClient: Sendable {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    struct PSIRequest: Encodable {
        let blinded: [String]
    }

    struct PSIResponse: Decodable, Sendable {
        let evaluated: [String]
        let index: [String: String]
        let request_id: String?
    }

    struct IdentityProfile: Decodable, Sendable {
        let did: String?
        let username: String?
        let display_name: String?
    }

    func evaluate(blinded: [String]) async throws -> PSIResponse {
        try await apiClient.post(endpoint: ContactsEndpoint.psiEvaluate, body: PSIRequest(blinded: blinded))
    }

    func resolveIdentity(did: String) async throws -> IdentityProfile {
        try await apiClient.get(endpoint: ContactsEndpoint.identityResolve(did: did))
    }
}

enum ContactDiscoveryError: LocalizedError, Sendable {
    case permissionDenied
    case oprfSessionExpired
    case discoveryUnavailable
    case noMatches

    var errorDescription: String? {
        switch self {
        case .permissionDenied:
            return "Contacts access is required to find friends on ECHO."
        case .oprfSessionExpired:
            return "Contact discovery session expired — try again."
        case .discoveryUnavailable:
            return "Contact discovery is temporarily unavailable."
        case .noMatches:
            return "No contacts on ECHO yet."
        }
    }
}
#endif
