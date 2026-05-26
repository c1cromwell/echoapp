import XCTest
@testable import Echo

final class ConversationSignalServiceTests: XCTestCase {

    private var transport: MockConversationSignalTransport!
    private var service: ConversationSignalService!

    override func setUp() {
        super.setUp()
        transport = MockConversationSignalTransport()
        service = ConversationSignalService(transport: transport)
    }

    func testSendTyping_buildsCorrectJSON() async throws {
        try await service.sendTyping(conversationId: "conv-a", peerDID: "did:key:peer", state: .start)
        XCTAssertEqual(transport.sentTexts.count, 1)
        let json = try XCTUnwrap(transport.sentTexts.first)
        XCTAssertTrue(json.contains("\"type\":\"typing\""))
        XCTAssertTrue(json.contains("\"to\":\"did:key:peer\""))
    }

    func testReceive_dispatchesByType() {
        let exp = expectation(description: "typing event")
        service.setEventHandler { event in
            if case .typing(let e) = event, e.state == .start {
                exp.fulfill()
            }
        }
        transport.simulateIncoming("""
        {"type":"typing","from":"did:key:peer","to":"did:key:me","conversation_id":"c1","payload":{"conversation_id":"c1","state":"start"}}
        """)
        wait(for: [exp], timeout: 1.0)
    }

    func testReceive_ignoresUnknownType() {
        let exp = expectation(description: "no event")
        exp.isInverted = true
        service.setEventHandler { _ in exp.fulfill() }
        transport.simulateIncoming("{\"type\":\"unknown\",\"to\":\"x\",\"payload\":{}}")
        wait(for: [exp], timeout: 0.2)
    }

    func testSendReadReceipt_skipsEmptyBatch() async throws {
        try await service.sendReadReceipt(conversationId: "c", peerDID: "did:key:p", messageIds: [])
        XCTAssertTrue(transport.sentTexts.isEmpty)
    }
}
