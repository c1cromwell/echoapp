import XCTest
@testable import Echo

final class AgentSearchGrantTests: XCTestCase {
    func testAgentSearchGrant_expiredReturnsNoHits() {
        let grant = AgentSearchGrant(verb: "search", expiresAt: Date().addingTimeInterval(-60))
        let hits = AgentEdgeSearch.hits(
            query: "lighthouse",
            grant: grant,
            corpus: [("m1", "meet at the lighthouse")]
        )
        XCTAssertTrue(hits.isEmpty)
    }

    func testAgentSearchGrant_validReturnsFixtureHitWithoutStorePath() {
        let grant = AgentSearchGrant(verb: "search", expiresAt: Date().addingTimeInterval(3600))
        let hits = AgentEdgeSearch.hits(
            query: "lighthouse",
            grant: grant,
            corpus: [("m1", "meet at the lighthouse")]
        )
        XCTAssertEqual(hits.count, 1)
        XCTAssertEqual(hits[0].messageId, "m1")
        let mirror = Mirror(reflecting: hits[0])
        let labels = mirror.children.compactMap(\.label)
        XCTAssertFalse(labels.contains("storePath"))
        XCTAssertFalse(labels.contains("indexKey"))
        for child in mirror.children {
            guard let text = child.value as? String else { continue }
            XCTAssertFalse(text.contains("EncryptedIndex"))
            XCTAssertFalse(text.contains("/"))
        }
    }

    func testAgentSearchGrant_sendVerbReturnsNoHits() {
        let grant = AgentSearchGrant(verb: "send", expiresAt: Date().addingTimeInterval(3600))
        let hits = AgentEdgeSearch.hits(
            query: "lighthouse",
            grant: grant,
            corpus: [("m1", "meet at the lighthouse")]
        )
        XCTAssertTrue(hits.isEmpty)
    }

    func testAgentSearchGrant_reviewRevokeAndConfirmOnce() {
        var review = GrantReviewState(
            grants: [
                AgentGrantReviewItem(
                    id: "g1",
                    agentName: "Desk",
                    verb: "send",
                    threadID: "dm:a:b",
                    expiresAt: Date().addingTimeInterval(3600)
                )
            ],
            pending: AgentChallenge(id: "c1", verb: "pay", threadID: "dm:a:b")
        )
        review.revoke(id: "g1")
        XCTAssertTrue(review.grants.isEmpty)
        let challenge = AgentChallenge(id: "c1", verb: "pay", threadID: "dm:a:b")
        XCTAssertTrue(review.confirm(challenge))
        XCTAssertNil(review.pending)
        XCTAssertFalse(review.confirm(challenge))
    }
}
