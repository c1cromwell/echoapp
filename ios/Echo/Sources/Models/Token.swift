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

/// Allocation breakdown for token distribution
public struct AllocationBreakdown {
    public let userRewards: Decimal      // 40% - 400M
    public let validatorRewards: Decimal // 25% - 250M
    public let ecosystem: Decimal        // 20% - 200M
    public let team: Decimal             // 8% - 80M
    public let treasury: Decimal         // 5% - 50M
    public let liquidity: Decimal        // 2% - 20M
    
    public init() {
        let total = Decimal(1_000_000_000)
        self.userRewards = total * 0.40
        self.validatorRewards = total * 0.25
        self.ecosystem = total * 0.20
        self.team = total * 0.08
        self.treasury = total * 0.05
        self.liquidity = total * 0.02
    }
    
    /// Calculate total allocation to verify correctness
    public func totalAllocation() -> Decimal {
        return userRewards + validatorRewards + ecosystem + team + treasury + liquidity
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
