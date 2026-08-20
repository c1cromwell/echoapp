import Foundation

/// Reward types
public enum RewardType: Int, Codable {
    case text = 0
    case voice = 1
    case video = 2
    case referral = 3
    case governance = 4
    case staking = 5
    case burn = 6
    case bridge = 7
}

/// Single reward earning
public struct RewardEarning: Identifiable {
    public let id: UUID
    public let userID: String
    public let rewardType: RewardType
    public let amount: Decimal
    public let multiplier: Double
    public let earnedAt: Date
    public var claimed: Bool = false
    public var claimedAt: Date? = nil
    
    public init(userID: String, rewardType: RewardType, amount: Decimal, multiplier: Double) {
        self.id = UUID()
        self.userID = userID
        self.rewardType = rewardType
        self.amount = amount
        self.multiplier = multiplier
        self.earnedAt = Date()
    }
}

/// Network activity tracking (auto-scaling model — no per-user daily caps)
/// Per PRD v2.5.1: auto-scaling model adopted, daily caps removed.
/// Rate = Daily Budget / Total Daily Activity Weight
/// Every message always earns; the per-message rate auto-scales based on network activity.
public struct NetworkActivityTracker {
    public let userID: String
    public let date: Date
    public private(set) var messagesRewarded: Int = 0
    public private(set) var echoEarned: Decimal = 0
    public private(set) var totalActions: Int = 0

    public init(userID: String) {
        self.userID = userID
        self.date = Date()
    }

    /// Every message always earns — no daily cap check needed
    public mutating func recordActivity() {
        messagesRewarded += 1
        totalActions += 1
    }

    /// Add earned amount
    public mutating func addEarnings(_ amount: Decimal) {
        echoEarned += amount
    }
}

/// Trust score with multiplier
public struct RewardsTrustScore {
    public let userID: String
    public private(set) var score: Int = 0 // 0-100
    public private(set) var level: String = "newcomer"
    public private(set) var updatedAt: Date = Date()
    public var components: [String: Int] = [:]
    
    public init(userID: String, score: Int = 0) {
        self.userID = userID
        self.score = max(0, min(100, score))
        self.updateLevel()
    }

    /// Update trust level based on score
    private mutating func updateLevel() {
        level = EchoScoreSnapshot.level(forScore: score)
    }

    public func getMultiplier() -> Double {
        EchoScoreSnapshot.multiplier(forScore: score)
    }
    
    /// Update score
    public mutating func updateScore(_ newScore: Int) {
        self.score = max(0, min(100, newScore))
        self.updatedAt = Date()
        self.updateLevel()
    }
}

/// Living Echo Score for the Rewards hub. Keep in lockstep with
/// `internal/tokenomics/models/echo_score.go`.
public struct EchoScoreSnapshot: Equatable, Sendable {
    public let score: Int
    public let tier: Int
    public let level: String
    public let multiplier: Double
    public let nextUnlock: UnlockFeature?

    public struct UnlockFeature: Equatable, Sendable {
        public let tier: Int
        public let minScore: Int
        public let feature: String
        public let pointsNeeded: Int
    }

    private static let ladder: [(tier: Int, minScore: Int, feature: String)] = [
        (2, 20, "Appear on the weekly leaderboard"),
        (3, 40, "Create broadcast channels"),
        (4, 60, "Governance voting"),
        (5, 80, "Highest earn multiplier"),
    ]

    public static func from(score raw: Int) -> EchoScoreSnapshot {
        let score = min(max(raw, 0), 100)
        var next: UnlockFeature?
        for step in ladder where score < step.minScore {
            next = UnlockFeature(
                tier: step.tier,
                minScore: step.minScore,
                feature: step.feature,
                pointsNeeded: step.minScore - score
            )
            break
        }
        return EchoScoreSnapshot(
            score: score,
            tier: tier(forScore: score),
            level: level(forScore: score),
            multiplier: multiplier(forScore: score),
            nextUnlock: next
        )
    }

    public static func from(tier: Int) -> EchoScoreSnapshot {
        from(score: midpoint(forTier: tier))
    }

    public static func multiplier(forScore score: Int) -> Double {
        switch score {
        case ..<20: return 1.0
        case ..<40: return 1.2
        case ..<60: return 1.5
        case ..<80: return 2.0
        default: return 3.0
        }
    }

    public static func tier(forScore score: Int) -> Int {
        switch score {
        case ..<20: return 1
        case ..<40: return 2
        case ..<60: return 3
        case ..<80: return 4
        default: return 5
        }
    }

    public static func level(forScore score: Int) -> String {
        switch score {
        case ..<20: return "unverified"
        case ..<40: return "newcomer"
        case ..<60: return "member"
        case ..<80: return "trusted"
        default: return "verified"
        }
    }

    public static func midpoint(forTier tier: Int) -> Int {
        switch tier {
        case ...1: return 10
        case 2: return 30
        case 3: return 50
        case 4: return 70
        default: return 90
        }
    }
}

/// Referral information
public struct ReferralInfo {
    public let referrerID: String
    public let refereeID: String
    public let signupBonus: Decimal
    public let verifyBonus: Decimal
    public let milestoneBonus: Decimal
    public let createdAt: Date
    
    public var totalBonus: Decimal {
        return signupBonus + verifyBonus + milestoneBonus
    }
    
    public init(referrerID: String, refereeID: String) {
        self.referrerID = referrerID
        self.refereeID = refereeID
        self.signupBonus = 5
        self.verifyBonus = 20
        self.milestoneBonus = 25
        self.createdAt = Date()
    }
}

// DailyRewards and RewardEarningEntry are defined in Core/Stargazer/WalletTypes.swift.
