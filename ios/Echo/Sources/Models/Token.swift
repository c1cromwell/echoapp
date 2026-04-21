import Foundation

/// Token model representing ECHO token specifications
public struct TokenConfig {
    public let name: String = "ECHO"
    public let symbol: String = "ECHO"
    public let totalSupply: Decimal = 1_000_000_000
    public let decimals: Int = 8
    public let hardCapped: Bool = true
    
    public init() {}
}

/// Allocation breakdown for token distribution (per ECHO Tokenomics v2.0)
public struct AllocationBreakdown {
    public let communityRewards: Decimal  // 40% - 400M (auto-scaling model, no daily caps)
    public let treasury: Decimal          // 22% - 220M
    public let founders: Decimal          // 18% - 180M (CEO 100M + 4 co-founders 20M each)
    public let futureTeam: Decimal        // 10% - 100M
    public let validatorRewards: Decimal  // 10% - 100M

    public init() {
        let total = Decimal(1_000_000_000)
        self.communityRewards = total * Decimal(string: "0.40")!
        self.treasury = total * Decimal(string: "0.22")!
        self.founders = total * Decimal(string: "0.18")!
        self.futureTeam = total * Decimal(string: "0.10")!
        self.validatorRewards = total * Decimal(string: "0.10")!
    }

    /// Calculate total allocation to verify correctness
    public func totalAllocation() -> Decimal {
        return communityRewards + treasury + founders + futureTeam + validatorRewards
    }

    /// Founder allocation detail (per Tokenomics v2.0)
    public struct FounderAllocation {
        public static let ceoAllocation: Decimal = 100_000_000      // 10% - CEO
        public static let coFounderAllocation: Decimal = 20_000_000  // 2% each - 4 co-founders
        public static let multiSigThreshold = 3                      // 3-of-5 founder multi-sig
        public static let multiSigTotal = 5
        public static let cliffMonths = 12
        public static let vestingMonths = 36                         // 1/36th monthly after cliff
        public static let withdrawCooldownDays = 14
    }
}

/// Token balance tracking
public struct TokenBalance {
    public let address: String
    public var availableBalance: Decimal
    public var vestingSchedule: VestingSchedule?
    
    public init(address: String, availableBalance: Decimal = 0) {
        self.address = address
        self.availableBalance = availableBalance
    }
}

/// Vesting schedule for locked tokens
public struct VestingSchedule {
    public let totalAmount: Decimal
    public let releasedAt: Date
    public let cliffMonths: Int
    public let vestMonths: Int
    public private(set) var releasedSoFar: Decimal = 0
    
    public init(totalAmount: Decimal, releasedAt: Date, cliffMonths: Int, vestMonths: Int) {
        self.totalAmount = totalAmount
        self.releasedAt = releasedAt
        self.cliffMonths = cliffMonths
        self.vestMonths = vestMonths
    }
    
    /// Calculate releasable amount
    public func calculateReleasable() -> Decimal {
        let calendar = Calendar.current
        let now = Date()
        
        let releaseComponents = calendar.dateComponents([.year, .month], from: releasedAt, to: now)
        
        let monthsElapsed = (releaseComponents.year ?? 0) * 12 + (releaseComponents.month ?? 0)
        
        // Before cliff period
        if monthsElapsed < cliffMonths {
            return 0
        }
        
        // After full vesting period
        if monthsElapsed >= cliffMonths + vestMonths {
            return totalAmount - releasedSoFar
        }
        
        // During vesting
        let vestingMonths = monthsElapsed - cliffMonths
        let monthlyRelease = totalAmount / Decimal(vestMonths)
        let totalReleasable = monthlyRelease * Decimal(vestingMonths)
        
        return totalReleasable - releasedSoFar
    }
}
