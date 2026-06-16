import Foundation

// MARK: - Models (WO-313 / WO-314 — zero-PII aggregates only)

struct ComplyDashboardSummary: Codable, Sendable, Equatable {
    let deCoverageRate: String
    let activeRetentionPolicies: Int
    let litigationHolds: Int
    let pendingExports: Int
    let anchorHealth: String
}

struct ComplyRetentionPolicy: Codable, Sendable, Equatable, Identifiable {
    let id: String
    let orgDid: String
    let policyType: String
    let conversationId: String?
    let scopeLabel: String?
    let dataL1Ref: String?
    let active: Bool
}

struct ComplyRetentionPolicyList: Codable, Sendable, Equatable {
    let policies: [ComplyRetentionPolicy]
}

struct ComplyAuditReport: Codable, Sendable, Equatable {
    let orgDid: String
    let generatedAt: Date
    let activeRetentionPolicies: Int
    let activeLitigationHolds: Int
    let pendingExports: Int
    let anchorHealth: String
    let verificationNotice: String
}

// MARK: - Client

protocol ComplyAPIClient: Sendable {
    func dashboard(orgDID: String) async throws -> ComplyDashboardSummary
    func listPolicies(orgDID: String) async throws -> [ComplyRetentionPolicy]
    func auditReport(orgDID: String) async throws -> ComplyAuditReport
    func auditReportPDF(orgDID: String) async throws -> Data
}

#if os(iOS)

enum ComplyEndpoint: APIEndpoint {
    case dashboard(orgDID: String)
    case policies(orgDID: String)
    case audit(orgDID: String)
    case auditPDF(orgDID: String)

    var path: String {
        switch self {
        case .dashboard:
            return "/comply/dashboard"
        case .policies:
            return "/comply/retention/policy"
        case .audit:
            return "/comply/audit/report?format=json"
        case .auditPDF:
            return "/comply/audit/report?format=pdf"
        }
    }

    var headers: [String: String] {
        switch self {
        case .dashboard(let org), .policies(let org), .audit(let org), .auditPDF(let org):
            return ["X-Org-DID": org]
        }
    }
}

/// Read-mostly Comply companion client (WO-314). Uses passkey-signed REST via `APIClient`.
actor ComplyAPI: ComplyAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func dashboard(orgDID: String) async throws -> ComplyDashboardSummary {
        try await apiClient.get(endpoint: ComplyEndpoint.dashboard(orgDID: orgDID))
    }

    func listPolicies(orgDID: String) async throws -> [ComplyRetentionPolicy] {
        let resp: ComplyRetentionPolicyList = try await apiClient.get(
            endpoint: ComplyEndpoint.policies(orgDID: orgDID)
        )
        return resp.policies
    }

    func auditReport(orgDID: String) async throws -> ComplyAuditReport {
        try await apiClient.get(endpoint: ComplyEndpoint.audit(orgDID: orgDID))
    }

    func auditReportPDF(orgDID: String) async throws -> Data {
        try await apiClient.getRaw(endpoint: ComplyEndpoint.auditPDF(orgDID: orgDID))
    }
}

#endif
