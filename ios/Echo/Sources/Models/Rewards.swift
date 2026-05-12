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
    
    /// Get reward multiplier based on trust tier (per Tokenomics v2.0)
    /// Tier 1: 1.0x, Tier 2: 1.2x, Tier 3: 1.5x, Tier 4: 2.0x, Tier 5: 3.0x
    public func getMultiplier() -> Double {
        switch score {
        case 0..<20:
            return 1.0     // Tier 1 — Unverified
        case 20..<40:
            return 1.2     // Tier 2 — Newcomer
        case 40..<60:
            return 1.5     // Tier 3 — Member
        case 60..<80:
            return 2.0     // Tier 4 — Trusted
        default:
            return 3.0     // Tier 5 — Verified
        }
    }
    
    /// Update trust level based on score
    private mutating func updateLevel() {
        level = switch score {
        case 0..<20:
            "unverified"
        case 20..<40:
            "newcomer"
        case 40..<60:
            "member"
        case 60..<80:
            "trusted"
        default:
            "verified"
        }
    }
    
    /// Update score
    public mutating func updateScore(_ newScore: Int) {
        self.score = max(0, min(100, newScore))
        self.updatedAt = Date()
        self.updateLevel()
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
