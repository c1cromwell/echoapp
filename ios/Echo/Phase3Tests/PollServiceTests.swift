import XCTest
@testable import Echo

#if os(iOS)
final class PollServiceTests: XCTestCase {
    func testPollLogic_createAndVote() {
        var polls: [String: ChatPoll] = [:]
        let create = PollWireBody(
            action: "create",
            question: "Lunch?",
            options: [
                PollOptionWire(id: "a", text: "Pizza"),
                PollOptionWire(id: "b", text: "Salad"),
            ],
            voterDID: nil,
            optionId: nil
        )
        let poll = PollLogic.apply(
            polls: &polls,
            pollId: "poll-1",
            conversationId: "c1",
            actorDID: "did:alice",
            body: create
        )
        XCTAssertEqual(poll?.question, "Lunch?")
        XCTAssertEqual(poll?.options.count, 2)

        let vote = PollWireBody(
            action: "vote",
            question: nil,
            options: nil,
            voterDID: "did:bob",
            optionId: "a"
        )
        let updated = PollLogic.apply(
            polls: &polls,
            pollId: "poll-1",
            conversationId: "c1",
            actorDID: "did:bob",
            body: vote
        )
        XCTAssertEqual(updated?.options.first?.voteCount, 1)
        XCTAssertTrue(updated?.options.first?.voters.contains("did:bob") == true)
    }

    func testPollLogic_close() {
        var polls: [String: ChatPoll] = [:]
        _ = PollLogic.apply(
            polls: &polls,
            pollId: "poll-2",
            conversationId: "c1",
            actorDID: "did:alice",
            body: PollWireBody(
                action: "create",
                question: "Done?",
                options: [PollOptionWire(id: "y", text: "Yes"), PollOptionWire(id: "n", text: "No")],
                voterDID: nil,
                optionId: nil
            )
        )
        let closed = PollLogic.apply(
            polls: &polls,
            pollId: "poll-2",
            conversationId: "c1",
            actorDID: "did:alice",
            body: PollWireBody(action: "close", question: nil, options: nil, voterDID: nil, optionId: nil)
        )
        XCTAssertTrue(closed?.isClosed == true)
    }
}
#endif
