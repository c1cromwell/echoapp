import XCTest
@testable import Echo

final class RewardsTests: XCTestCase {
    
    func testRewardTypes() {
        XCTAssertEqual(RewardType.text.rawValue, 0)
        XCTAssertEqual(RewardType.voice.rawValue, 1)
        XCTAssertEqual(RewardType.video.rawValue, 2)
        XCTAssertEqual(RewardType.referral.rawValue, 3)
    }
    
    func testRewardEarning() {
        let earning = RewardEarning(
            userID: "user-123",
            rewardType: .text,
            amount: 10,
            multiplier: 1.5
        )
        
        XCTAssertEqual(earning.userID, "user-123")
        XCTAssertEqual(earning.rewardType, .text)
        XCTAssertEqual(earning.amount, 10)
        XCTAssertEqual(earning.multiplier, 1.5)
        XCTAssertFalse(earning.claimed)
    }
    
    func testDailyRewardTrackerInit() {
        let tracker = DailyRewardTracker(userID: "user-123")
        
        XCTAssertEqual(tracker.userID, "user-123")
        XCTAssertEqual(tracker.messagesRewarded, 0)
        XCTAssertEqual(tracker.echoEarned, 0)
        XCTAssertFalse(tracker.isLimitReached())
    }
    
    func testDailyRewardTrackerLimitNotReached() {
        var tracker = DailyRewardTracker(userID: "user-123")
        
        for _ in 0..<250 {
            tracker.incrementMessages()
        }
        
        XCTAssertEqual(tracker.messagesRewarded, 250)
        XCTAssertFalse(tracker.isLimitReached())
    }
    
    func testDailyRewardTrackerLimitReached() {
        var tracker = DailyRewardTracker(userID: "user-123")
        
        for _ in 0..<500 {
            tracker.incrementMessages()
        }
        
        XCTAssertEqual(tracker.messagesRewarded, 500)
        XCTAssertTrue(tracker.isLimitReached())
    }
    
    func testDailyRewardTrackerEarnings() {
        var tracker = DailyRewardTracker(userID: "user-123")
        
        tracker.addEarnings(10.5)
        XCTAssertEqual(tracker.echoEarned, 10.5)
        
        tracker.addEarnings(5.25)
        XCTAssertEqual(tracker.echoEarned, 15.75)
    }
    
    func testTrustScoreMultipliers() {
        let tests: [(Int, Double)] = [
            (10, 0.5),
            (30, 1.0),
            (50, 1.5),
            (70, 2.5),
            (90, 5.0),
        ]
        
        for (score, expectedMultiplier) in tests {
            let trustScore = RewardsTrustScore(userID: "user-123", score: score)
            XCTAssertEqual(trustScore.getMultiplier(), expectedMultiplier, "Score \(score) failed")
        }
    }
    
    func testTrustScoreLevels() {
        let tests: [(Int, String)] = [
            (10, "unverified"),
            (30, "newcomer"),
            (50, "member"),
            (70, "trusted"),
            (90, "verified"),
        ]
        
        for (score, expectedLevel) in tests {
            let trustScore = RewardsTrustScore(userID: "user-123", score: score)
            XCTAssertEqual(trustScore.level, expectedLevel, "Score \(score) level failed")
        }
    }
    
    func testTrustScoreUpdate() {
        var trustScore = RewardsTrustScore(userID: "user-123", score: 0)
        XCTAssertEqual(trustScore.level, "unverified")
        
        trustScore.updateScore(50)
        XCTAssertEqual(trustScore.score, 50)
        XCTAssertEqual(trustScore.level, "member")
    }
    
    func testReferralInfo() {
        let referral = ReferralInfo(referrerID: "ref-1", refereeID: "ref-2")
        
        XCTAssertEqual(referral.referrerID, "ref-1")
        XCTAssertEqual(referral.refereeID, "ref-2")
        XCTAssertEqual(referral.signupBonus, 5)
        XCTAssertEqual(referral.verifyBonus, 20)
        XCTAssertEqual(referral.milestoneBonus, 25)
        XCTAssertEqual(referral.totalBonus, 50)
    }
}
