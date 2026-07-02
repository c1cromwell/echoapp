#if os(iOS)
import XCTest
@testable import Echo

final class GamificationDecodingTests: XCTestCase {
    func testQuestItemDecoding() throws {
        let json = """
        {
          "questId": "identity_builder",
          "title": "Identity Builder",
          "description": "Complete identity verification",
          "action": "identity_verification",
          "requiredCount": 1,
          "reward_echo": 2000000000,
          "badge": "Verified",
          "tier": "starter",
          "completedAt": "2026-01-15T00:00:00Z",
          "rewardClaimed": false,
          "progress": 1
        }
        """.data(using: .utf8)!
        let item = try JSONDecoder().decode(QuestItem.self, from: json)
        XCTAssertEqual(item.questId, "identity_builder")
        XCTAssertTrue(item.isCompleted)
        XCTAssertTrue(item.isClaimable)
        XCTAssertEqual(item.rewardAmount, Decimal(20))
    }

    func testQuestCatalogResponseDecoding() throws {
        let json = """
        {
          "starter": [],
          "advanced": []
        }
        """.data(using: .utf8)!
        let catalog = try JSONDecoder().decode(QuestCatalogResponse.self, from: json)
        XCTAssertTrue(catalog.starter.isEmpty)
    }

    func testFounderVestingProfileEquatable() {
        let vesting = VestingState(
            role: "CEO",
            totalAllocated: 100,
            vested: 25,
            locked: 75,
            withdrawable: 25,
            nextUnlockAmount: 2,
            nextUnlockDate: nil,
            cliffDate: Date(),
            cliffCompleted: true,
            vestingPercent: 25,
            explorerURL: URL(string: "https://dagexplorer.io/address/did:key:test")
        )
        let profile = FounderVestingProfile(
            vesting: vesting,
            revocationEvents: [],
            explorerURL: vesting.explorerURL,
            founderLockId: "lock-1"
        )
        XCTAssertEqual(profile.founderLockId, "lock-1")
    }
}
#endif
