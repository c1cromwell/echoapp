import XCTest
@testable import Echo

final class EchoScoreLadderTests: XCTestCase {
    func testUnlockLadderMatchesTokenomicsV2() {
        let low = EchoScoreSnapshot.from(score: 10)
        XCTAssertEqual(low.tier, 1)
        XCTAssertEqual(low.level, "unverified")
        XCTAssertEqual(low.multiplier, 1.0)
        XCTAssertEqual(low.nextUnlock?.tier, 2)
        XCTAssertEqual(low.nextUnlock?.pointsNeeded, 10)
        XCTAssertEqual(low.nextUnlock?.feature, "Appear on the weekly leaderboard")

        let member = EchoScoreSnapshot.from(score: 50)
        XCTAssertEqual(member.tier, 3)
        XCTAssertEqual(member.multiplier, 1.5)
        XCTAssertEqual(member.nextUnlock?.tier, 4)

        let top = EchoScoreSnapshot.from(score: 90)
        XCTAssertEqual(top.tier, 5)
        XCTAssertEqual(top.multiplier, 3.0)
        XCTAssertNil(top.nextUnlock)
    }

    func testTierMidpoints() {
        XCTAssertEqual(EchoScoreSnapshot.from(tier: 3).score, 50)
        XCTAssertEqual(EchoScoreSnapshot.from(score: -3).score, 0)
        XCTAssertEqual(EchoScoreSnapshot.from(score: 200).score, 100)
    }
}
