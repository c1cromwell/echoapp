#if os(iOS)
import Foundation

// MARK: - Quest models (WO-271)

extension QuestItem {
    private enum PointsKey: String, CodingKey { case points }

    /// The server sends a quest's reward as whole `points`. (Older servers sent `reward_echo` in
    /// 8-decimal units; both decode to the same stored amount.)
    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        questId = try c.decode(String.self, forKey: .questId)
        title = try c.decode(String.self, forKey: .title)
        description = try c.decode(String.self, forKey: .description)
        action = try c.decode(String.self, forKey: .action)
        requiredCount = try c.decode(Int.self, forKey: .requiredCount)
        badge = try c.decode(String.self, forKey: .badge)
        tier = try c.decode(String.self, forKey: .tier)
        minTrustTier = try c.decodeIfPresent(Int.self, forKey: .minTrustTier)
        completedAt = try c.decodeIfPresent(String.self, forKey: .completedAt)
        rewardClaimed = try c.decodeIfPresent(Bool.self, forKey: .rewardClaimed) ?? false
        progress = try c.decodeIfPresent(Int.self, forKey: .progress) ?? 0
        points = try decoder.container(keyedBy: PointsKey.self).decodeIfPresent(Int.self, forKey: .points) ?? 0
    }
}

struct QuestItem: Codable, Sendable, Equatable, Identifiable {
    var id: String { questId }
    let questId: String
    let title: String
    let description: String
    let action: String
    let requiredCount: Int
    /// The quest's reward, in whole points.
    let points: Int
    let badge: String
    let tier: String
    let minTrustTier: Int?
    let completedAt: String?
    let rewardClaimed: Bool
    let progress: Int

    enum CodingKeys: String, CodingKey {
        case questId, title, description, action, requiredCount, badge, tier
        case points
        case minTrustTier
        case completedAt, rewardClaimed, progress
    }

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
    case weeklyPack
    case weeklyPackOpen

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
        case .weeklyPack, .weeklyPackOpen:
            return "/v1/gamification/weekly-pack"
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

    /// Points earned in the window.
    var scoreAmount: Decimal { Decimal(score) }
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
    let redeemable: Bool?
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

/// What the server says about rewards — drives the persistent "no cash value" note. Points are
/// a score; the custody and redeemability fields belong to the dormant token surface and are
/// absent unless it is switched on.
struct GamificationStatus: Codable, Sendable {
    let custodyMode: String?
    let redeemable: Bool?
    let transferable: Bool?
    let disclaimer: String?
    let echoScore: EchoScoreWire?

    enum CodingKeys: String, CodingKey {
        case redeemable, transferable, disclaimer
        case custodyMode = "custody_mode"
        case echoScore = "echo_score"
    }

    var scoreSnapshot: EchoScoreSnapshot? {
        echoScore?.snapshot
    }
}

struct EchoScoreWire: Codable, Sendable {
    let score: Int
    let tier: Int
    let level: String
    let multiplier: Double
    let nextUnlock: UnlockWire?

    enum CodingKeys: String, CodingKey {
        case score, tier, level, multiplier
        case nextUnlock = "next_unlock"
    }

    var snapshot: EchoScoreSnapshot {
        EchoScoreSnapshot(
            score: score,
            tier: tier,
            level: level,
            multiplier: multiplier,
            nextUnlock: nextUnlock.map {
                EchoScoreSnapshot.UnlockFeature(
                    tier: $0.tier,
                    minScore: $0.minScore,
                    feature: $0.feature,
                    pointsNeeded: $0.pointsNeeded
                )
            }
        )
    }
}

struct UnlockWire: Codable, Sendable {
    let tier: Int
    let minScore: Int
    let feature: String
    let pointsNeeded: Int

    enum CodingKeys: String, CodingKey {
        case tier, feature
        case minScore = "min_score"
        case pointsNeeded = "points_needed"
    }
}

struct WeeklyPackItem: Codable, Sendable, Identifiable, Equatable {
    let kind: String
    let title: String
    let detail: String
    var id: String { kind + title }
}

struct WeeklyPack: Codable, Sendable, Equatable {
    let weekKey: String
    let label: String
    let opened: Bool
    let items: [WeeklyPackItem]
    let openedAt: String?

    enum CodingKeys: String, CodingKey {
        case label, opened, items
        case weekKey = "week_key"
        case openedAt = "opened_at"
    }
}

struct WeeklyPackResponse: Codable, Sendable {
    let pack: WeeklyPack
}

// MARK: - Protocol

protocol GamificationAPIClient: Sendable {
    func fetchQuestCatalog() async throws -> QuestCatalogResponse
    func claimQuest(_ questId: String) async throws -> QuestClaimResponse
    func fetchLeaderboard(window: String) async throws -> LeaderboardResponse
    func fetchStreak() async throws -> StreakResponse
    func pingActivity() async throws -> StreakResponse
    func fetchStatus() async throws -> GamificationStatus
    func fetchWeeklyPack() async throws -> WeeklyPackResponse
    func openWeeklyPack() async throws -> WeeklyPackResponse
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

    func fetchWeeklyPack() async throws -> WeeklyPackResponse {
        try await apiClient.get(endpoint: GamificationEndpoint.weeklyPack)
    }

    func openWeeklyPack() async throws -> WeeklyPackResponse {
        struct Empty: Encodable {}
        return try await apiClient.post(endpoint: GamificationEndpoint.weeklyPackOpen, body: Empty())
    }
}

#endif
