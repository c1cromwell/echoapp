#if os(iOS)
import Foundation

// MARK: - Gamification API

/// Client for quests, badges, and gamification features (WO-271).
/// In production, delegates to the Go backend's gamification endpoints.
final class GamificationAPI: @unchecked Sendable {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    struct Quest: Identifiable, Sendable {
        let id: String
        let title: String
        let description: String
        let progress: Double // 0.0–1.0
        let rewardAmount: Decimal
        let isCompleted: Bool
    }

    struct Badge: Identifiable, Sendable {
        let id: String
        let name: String
        let icon: String
        let earnedDate: Date?
        var isEarned: Bool { earnedDate != nil }
    }

    func fetchQuests() async throws -> [Quest] {
        // Stub — real implementation calls /api/v3/gamification/quests
        []
    }

    func fetchBadges() async throws -> [Badge] {
        // Stub — real implementation calls /api/v3/gamification/badges
        []
    }
}

#endif
