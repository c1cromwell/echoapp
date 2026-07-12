#if os(iOS)
import Foundation

// MARK: - Quest models (WO-271)

struct QuestItem: Codable, Sendable, Equatable, Identifiable {
    var id: String { questId }
    let questId: String
    let title: String
    let description: String
    let action: String
    let requiredCount: Int
    let rewardEcho: Int64
    let badge: String
    let tier: String
    let minTrustTier: Int?
    let completedAt: String?
    let rewardClaimed: Bool
    let progress: Int

    enum CodingKeys: String, CodingKey {
        case questId, title, description, action, requiredCount, badge, tier
        case rewardEcho = "reward_echo"
        case minTrustTier
        case completedAt, rewardClaimed, progress
    }

    var rewardAmount: Decimal { EchoDatum.fromDatum(rewardEcho) }

    var isCompleted: Bool { completedAt != nil && !completedAt!.isEmpty }

    var isClaimable: Bool { isCompleted && !rewardClaimed }

    var progressFraction: Double {
        guard requiredCount > 0 else { return 0 }
        return min(Double(progress) / Double(requiredCount), 1.0)
    }
}

struct QuestCatalogResponse: Codable, Sendable {
    let starter: [QuestItem]
    let advanced: [QuestItem]
}

struct QuestClaimResponse: Codable, Sendable {
    let questId: String
    let txHash: String
}

// MARK: - Endpoints

enum GamificationEndpoint: APIEndpoint {
    case listQuests
    case claim(questId: String)
    case leaderboard(window: String)
    case streak
    case activityPing
    case status

    var path: String {
        switch self {
        case .listQuests:
            return "/v1/gamification/quests"
        case .claim(let questId):
            return "/v1/gamification/quests/\(questId)/claim"
        case .leaderboard(let window):
            return "/v1/gamification/leaderboard?window=\(window)"
        case .streak:
            return "/v1/gamification/streak"
        case .activityPing:
            return "/v1/gamification/activity/ping"
        case .status:
            return "/v1/gamification/status"
        }
    }
}

// MARK: - Leaderboard / streak / status models (R0 gamification)

struct LeaderboardEntry: Codable, Sendable, Identifiable {
    let did: String
    let trustTier: Int
    let score: Int64
    let rank: Int
    var id: String { did }

    enum CodingKeys: String, CodingKey {
        case did, score, rank
        case trustTier = "trust_tier"
    }

    /// ECHO earned in the window (display-only; no cash value at launch).
    var scoreAmount: Decimal { EchoDatum.fromDatum(score) }
}

struct LeaderboardSnapshot: Codable, Sendable {
    let window: String
    let bucketKey: String
    let entries: [LeaderboardEntry]
    let updatedAt: String?

    enum CodingKeys: String, CodingKey {
        case window, entries
        case bucketKey = "bucket_key"
        case updatedAt = "updated_at"
    }
}

struct LeaderboardResponse: Codable, Sendable {
    let snapshot: LeaderboardSnapshot
    let redeemable: Bool
}

struct StreakInfo: Codable, Sendable {
    let did: String
    let currentDays: Int
    let longestDays: Int
    let lastActive: String?
    let multiplier: Double
    let milestone: String?

    enum CodingKeys: String, CodingKey {
        case did, multiplier, milestone
        case currentDays = "current_days"
        case longestDays = "longest_days"
        case lastActive = "last_active"
    }
}

struct StreakResponse: Codable, Sendable {
    let streak: StreakInfo
}

/// Launch compliance posture — drives the persistent "no cash value" disclaimer.
/// While `custodyMode == "interim"`, `redeemable`/`transferable` are false.
struct GamificationStatus: Codable, Sendable {
    let custodyMode: String
    let redeemable: Bool
    let transferable: Bool
    let disclaimer: String?

    enum CodingKeys: String, CodingKey {
        case redeemable, transferable, disclaimer
        case custodyMode = "custody_mode"
    }
}

enum TokenomicsEndpoint: APIEndpoint {
    case emissionStatus
    case founderVesting

    var path: String {
        switch self {
        case .emissionStatus:
            return "/v1/tokens/emission/status"
        case .founderVesting:
            return "/v1/tokens/vesting"
        }
    }
}

// MARK: - Wire models for vesting + emission

struct EmissionStatusWire: Codable, Sendable {
    let currentYear: Int
    let annualCap: Int64
    let distributedToDate: Int64
    let remainingBudget: Int64
    let percentConsumed: Double

    enum CodingKeys: String, CodingKey {
        case currentYear = "current_year"
        case annualCap = "annual_cap"
        case distributedToDate = "distributed_to_date"
        case remainingBudget = "remaining_budget"
        case percentConsumed = "percent_consumed"
    }
}

struct TokenomicsVestingWire: Codable, Sendable {
    let role: String
    let totalAllocated: Int64
    let vested: Int64
    let locked: Int64
    let withdrawable: Int64
    let nextUnlockAmount: Int64
    let nextUnlockDate: Date?
    let cliffDate: Date
    let cliffCompleted: Bool
    let vestingPercent: Double
    let explorerUrl: String?
}

struct RevocationEventWire: Codable, Sendable {
    let targetFounderDID: String
    let revokedAmount: Int64
    let revokerDIDs: [String]
    let timestamp: String
    let txHash: String
    let destinationPool: String
}

struct FounderVestingWire: Codable, Sendable {
    let did: String
    let isFounder: Bool
    let vesting: TokenomicsVestingWire
    let revocationEvents: [RevocationEventWire]
    let explorerUrl: String

    enum CodingKeys: String, CodingKey {
        case did
        case isFounder = "is_founder"
        case vesting
        case revocationEvents = "revocation_events"
        case explorerUrl = "explorer_url"
    }
}

// MARK: - Protocol

protocol GamificationAPIClient: Sendable {
    func fetchQuestCatalog() async throws -> QuestCatalogResponse
    func claimQuest(_ questId: String) async throws -> QuestClaimResponse
    func fetchLeaderboard(window: String) async throws -> LeaderboardResponse
    func fetchStreak() async throws -> StreakResponse
    func pingActivity() async throws -> StreakResponse
    func fetchStatus() async throws -> GamificationStatus
}

// MARK: - Live client

actor GamificationAPI: GamificationAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func fetchQuestCatalog() async throws -> QuestCatalogResponse {
        try await apiClient.get(endpoint: GamificationEndpoint.listQuests)
    }

    func claimQuest(_ questId: String) async throws -> QuestClaimResponse {
        struct Empty: Encodable {}
        return try await apiClient.post(endpoint: GamificationEndpoint.claim(questId: questId), body: Empty())
    }

    func fetchLeaderboard(window: String) async throws -> LeaderboardResponse {
        try await apiClient.get(endpoint: GamificationEndpoint.leaderboard(window: window))
    }

    func fetchStreak() async throws -> StreakResponse {
        try await apiClient.get(endpoint: GamificationEndpoint.streak)
    }

    func pingActivity() async throws -> StreakResponse {
        struct Empty: Encodable {}
        return try await apiClient.post(endpoint: GamificationEndpoint.activityPing, body: Empty())
    }

    func fetchStatus() async throws -> GamificationStatus {
        try await apiClient.get(endpoint: GamificationEndpoint.status)
    }
}

// MARK: - Mapping helpers

enum TokenomicsMapping {
    static func mapEmission(_ wire: EmissionStatusWire) -> EmissionStatus {
        EmissionStatus(
            currentYear: wire.currentYear,
            annualCap: EchoDatum.fromDatum(wire.annualCap),
            distributedToDate: EchoDatum.fromDatum(wire.distributedToDate),
            remainingBudget: EchoDatum.fromDatum(wire.remainingBudget),
            percentConsumed: wire.percentConsumed
        )
    }

    static func mapVesting(_ wire: TokenomicsVestingWire, explorerFallback: String?) -> VestingState {
        let urlString = wire.explorerUrl ?? explorerFallback
        return VestingState(
            role: wire.role,
            totalAllocated: EchoDatum.fromDatum(wire.totalAllocated),
            vested: EchoDatum.fromDatum(wire.vested),
            locked: EchoDatum.fromDatum(wire.locked),
            withdrawable: EchoDatum.fromDatum(wire.withdrawable),
            nextUnlockAmount: EchoDatum.fromDatum(wire.nextUnlockAmount),
            nextUnlockDate: wire.nextUnlockDate,
            cliffDate: wire.cliffDate,
            cliffCompleted: wire.cliffCompleted,
            vestingPercent: wire.vestingPercent,
            explorerURL: urlString.flatMap { URL(string: $0) }
        )
    }

    static func mapFounderVesting(_ wire: FounderVestingWire, lockId: String?) -> FounderVestingProfile {
        let events = wire.revocationEvents.map {
            RevocationEvent(
                targetFounderDID: $0.targetFounderDID,
                revokedAmount: EchoDatum.fromDatum($0.revokedAmount),
                revokerDIDs: $0.revokerDIDs,
                timestamp: $0.timestamp,
                txHash: $0.txHash,
                destinationPool: $0.destinationPool
            )
        }
        return FounderVestingProfile(
            vesting: mapVesting(wire.vesting, explorerFallback: wire.explorerUrl),
            revocationEvents: events,
            explorerURL: URL(string: wire.explorerUrl),
            founderLockId: lockId
        )
    }
}
#endif
