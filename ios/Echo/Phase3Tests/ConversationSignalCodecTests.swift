import XCTest
@testable import Echo

final class ConversationSignalCodecTests: XCTestCase {

    func testTypingEnvelope_usesSnakeCaseKeys() throws {
        let json = try ConversationSignalCodec.encodeTyping(
            to: "did:key:peer",
            conversationId: "conv-1",
            state: .start
        )
        XCTAssertTrue(json.contains("\"conversation_id\""))
        XCTAssertTrue(json.contains("\"type\":\"typing\""))
        XCTAssertTrue(json.contains("\"to\":\"did:key:peer\""))
        XCTAssertTrue(json.contains("\"state\":\"start\""))
    }

    func testReadReceiptEnvelope_roundTrip() throws {
        let json = try ConversationSignalCodec.encodeReadReceipt(
            to: "did:key:alice",
            conversationId: "conv-2",
            messageIds: ["m1", "m2"],
            readAt: "2026-05-25T12:00:00Z"
        )
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .readReceipt(let e) = event else {
            return XCTFail("expected read receipt")
        }
        XCTAssertEqual(e.conversationId, "conv-2")
        XCTAssertEqual(e.messageIds, ["m1", "m2"])
        XCTAssertEqual(e.readAt, "2026-05-25T12:00:00Z")
    }

    func testReactionEnvelope_roundTrip() throws {
        let json = try ConversationSignalCodec.encodeReaction(
            to: "did:key:bob",
            conversationId: "conv-3",
            messageId: "msg-9",
            emoji: "👍"
        )
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .reaction(let e) = event else {
            return XCTFail("expected reaction")
        }
        XCTAssertEqual(e.messageId, "msg-9")
        XCTAssertEqual(e.emoji, "👍")
    }

    func testTextMessageEnvelope_roundTrip() throws {
        let json = try ConversationSignalCodec.encodeTextMessage(
            to: "did:key:bob",
            conversationId: "conv-1",
            payload: TextMessagePayload(messageId: "m1", text: "Hello", encrypted: nil)
        )
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .textMessage(let e) = event else {
            return XCTFail("expected text")
        }
        XCTAssertEqual(e.conversationId, "conv-1")
        XCTAssertEqual(e.messageId, "m1")
        XCTAssertEqual(e.text, "Hello")
    }

    func testDecodeEvent_unknownTypeReturnsNil() throws {
        let json = """
        {"type":"presence","to":"did:key:x","payload":{}}
        """
        XCTAssertNil(try ConversationSignalCodec.decodeEvent(from: json))
    }

    func testTypingDecode_fromFieldBecomesPeerDID() throws {
        let json = """
        {"type":"typing","from":"did:key:peer","to":"did:key:me","conversation_id":"c1","payload":{"conversation_id":"c1","state":"stop"}}
        """
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .typing(let e) = event else {
            return XCTFail("expected typing")
        }
        XCTAssertEqual(e.peerDID, "did:key:peer")
        XCTAssertEqual(e.state, .stop)
    }
}
