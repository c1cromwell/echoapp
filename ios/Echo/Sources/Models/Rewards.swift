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

/// Daily reward tracking
public struct DailyRewardTracker {
    public let userID: String
    public let date: Date
    public private(set) var messagesRewarded: Int = 0
    public private(set) var echoEarned: Decimal = 0
    public private(set) var totalActions: Int = 0
    
    public init(userID: String) {
        self.userID = userID
        self.date = Date()
    }
    
    /// Check if daily limits are reached
    public func isLimitReached() -> Bool {
        return messagesRewarded >= 500
    }
    
    /// Increment message counter
    public mutating func incrementMessages() {
        if !isLimitReached() {
            messagesRewarded += 1
            totalActions += 1
        }
    }
    
    /// Add earned amount
    public mutating func addEarnings(_ amount: Decimal) {
        echoEarned += amount
    }
}

/// Trust score with multiplier
public struct TrustScore {
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
    
    /// Get reward multiplier based on trust score
    public func getMultiplier() -> Double {
        switch score {
        case 0..<20:
            return 0.5     // Unverified
        case 20..<40:
            return 1.0     // Newcomer
        case 40..<60:
            return 1.5     // Member
        case 60..<80:
            return 2.5     // Trusted
        default:
            return 5.0     // Verified
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
