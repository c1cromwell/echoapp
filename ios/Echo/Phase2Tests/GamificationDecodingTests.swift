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

    func testGamificationStatusDecodesEchoScore() throws {
        let json = """
        {
          "custody_mode": "interim",
          "redeemable": false,
          "transferable": false,
          "disclaimer": "ECHO earned in-app has no cash value and cannot be transferred or redeemed.",
          "echo_score": {
            "score": 34,
            "tier": 2,
            "level": "newcomer",
            "multiplier": 1.2,
            "next_unlock": {
              "tier": 3,
              "min_score": 40,
              "feature": "Create broadcast channels",
              "points_needed": 6
            }
          }
        }
        """.data(using: .utf8)!
        let status = try JSONDecoder().decode(GamificationStatus.self, from: json)
        XCTAssertEqual(status.scoreSnapshot?.score, 34)
        XCTAssertEqual(status.scoreSnapshot?.nextUnlock?.tier, 3)
        XCTAssertEqual(status.scoreSnapshot?.nextUnlock?.pointsNeeded, 6)
    }

    func testWeeklyPackDecoding() throws {
        let json = """
        {
          "pack": {
            "week_key": "weekly:2026-W34",
            "label": "Week One pack",
            "opened": false,
            "items": [
              {"kind": "badge", "title": "Week One", "detail": "Seven-day streak standing."}
            ]
          }
        }
        """.data(using: .utf8)!
        let resp = try JSONDecoder().decode(WeeklyPackResponse.self, from: json)
        XCTAssertEqual(resp.pack.label, "Week One pack")
        XCTAssertEqual(resp.pack.items.count, 1)
        XCTAssertFalse(resp.pack.opened)
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
