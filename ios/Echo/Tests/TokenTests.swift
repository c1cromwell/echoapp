import XCTest
@testable import Echo

final class TokenTests: XCTestCase {
    
    func testTokenConfiguration() {
        let config = TokenConfig()
        
        XCTAssertEqual(config.name, "ECHO")
        XCTAssertEqual(config.symbol, "ECHO")
        XCTAssertEqual(config.totalSupply, 1_000_000_000)
        XCTAssertEqual(config.decimals, 8)
        XCTAssertTrue(config.hardCapped)
    }
    
    func testAllocationBreakdown() {
        let allocation = AllocationBreakdown()
        
        // Verify percentages
        XCTAssertEqual(allocation.userRewards, 400_000_000)
        XCTAssertEqual(allocation.validatorRewards, 250_000_000)
        XCTAssertEqual(allocation.ecosystem, 200_000_000)
        XCTAssertEqual(allocation.team, 80_000_000)
        XCTAssertEqual(allocation.treasury, 50_000_000)
        XCTAssertEqual(allocation.liquidity, 20_000_000)
    }
    
    func testAllocationTotal() {
        let allocation = AllocationBreakdown()
        let total = allocation.totalAllocation()
        
        XCTAssertEqual(total, 1_000_000_000)
    }
    
    func testTokenBalance() {
        let balance = TokenBalance(address: "test-address", availableBalance: 1000)
        
        XCTAssertEqual(balance.address, "test-address")
        XCTAssertEqual(balance.availableBalance, 1000)
    }
    
    func testVestingScheduleBeforeCliff() {
        let schedule = VestingSchedule(
            totalAmount: 1000,
            releasedAt: Date(),
            cliffMonths: 12,
            vestMonths: 24
        )
        
        let releasable = schedule.calculateReleasable()
        XCTAssertEqual(releasable, 0)
    }
    
    func testVestingScheduleAfterCliff() {
        let releaseDate = Calendar.current.date(byAdding: .month, value: -13, to: Date())!
        let schedule = VestingSchedule(
            totalAmount: 1200,
            releasedAt: releaseDate,
            cliffMonths: 12,
            vestMonths: 24
        )
        
        let releasable = schedule.calculateReleasable()
        XCTAssertGreaterThan(releasable, 0)
    }
}
