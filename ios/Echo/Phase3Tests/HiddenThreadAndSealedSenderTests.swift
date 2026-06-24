#if os(iOS)
import XCTest
@testable import Echo

@MainActor
final class HiddenThreadCryptoTests: XCTestCase {
    func testRoundTrip_encryptsHiddenThreadBlob() throws {
        let conversationId = "conv-hidden-test"
        let messages = [
            StoredThreadMessage(
                id: "m1",
                senderDID: "did:example:alice",
                content: "secret",
                timestamp: "Now"
            )
        ]
        let encrypted = try HiddenThreadCrypto.encrypt(messages: messages, conversationId: conversationId)
        let decoded = try HiddenThreadCrypto.decrypt(data: encrypted, conversationId: conversationId)
        XCTAssertEqual(decoded, messages)
    }
}

final class SealedSenderCodecTests: XCTestCase {
    func testEncodeDecode_sealedTextEnvelope() throws {
        let payload = SealedTextPayload(deliveryToken: "tok", ciphertext: Data([1, 2, 3]))
        let wire = try ConversationSignalCodec.encodeSealedTextMessage(
            to: "did:example:bob",
            conversationId: "conv-1",
            payload: payload
        )
        let event = try XCTUnwrap(try ConversationSignalCodec.decodeEvent(from: wire))
        guard case .textMessage(let textEvent) = event else {
            return XCTFail("expected textMessage event")
        }
        XCTAssertEqual(textEvent.sealedPayload, payload)
        XCTAssertTrue(textEvent.peerDID.isEmpty)
    }
}
#endif
