#if os(iOS)
import Foundation

struct DisappearingRestrictionPolicy: Codable, Sendable, Equatable {
    let minTTLSeconds: Int
    let reason: String
    let blockedTTLs: [Int]?

    enum CodingKeys: String, CodingKey {
        case minTTLSeconds = "min_ttl_seconds"
        case reason
        case blockedTTLs = "blocked_ttls"
    }
}

struct DisappearingRestrictionsResponse: Codable, Sendable {
    let trustTier: Int
    let policy: DisappearingRestrictionPolicy
    let violationCount: Int?

    enum CodingKeys: String, CodingKey {
        case trustTier = "trust_tier"
        case policy
        case violationCount = "violation_count"
    }
}

/// Trust-tier disappearing TTL policy client (WO-115).
enum DisappearingRestrictionsAPI {
    static func fetchPolicy() async -> DisappearingRestrictionsResponse? {
        guard let token = try? await KeychainManager.shared.getAuthToken() else { return nil }
        guard let url = URL(string: APIConfiguration.default.baseURL.absoluteString + "/v3/disappearing/restrictions") else { return nil }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        guard let (data, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse, http.statusCode == 200 else { return nil }
        return try? JSONDecoder().decode(DisappearingRestrictionsResponse.self, from: data)
    }

    static func isTimerAllowed(_ timer: DisappearingTimer, policy: DisappearingRestrictionPolicy) -> Bool {
        let seconds = timer.seconds
        if seconds == 0 { return true }
        if let blocked = policy.blockedTTLs, blocked.contains(seconds) { return false }
        return seconds >= policy.minTTLSeconds
    }

    static func allowedTimers(policy: DisappearingRestrictionPolicy) -> [DisappearingTimer] {
        DisappearingTimer.allCases.filter { isTimerAllowed($0, policy: policy) }
    }
}
#endif
